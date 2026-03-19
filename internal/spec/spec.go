package spec

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Manifest struct {
	SourcePath  string
	ManifestDir string

	Root        string
	Directories []BundleDir
	Files       []BundleFile

	Agent       AgentSpec
	Program     ProgramSpec
	Logging     LoggingSpec
	Environment map[string]string
	EnvFromHost []string
	Service     ServiceSpec
	Install     InstallSpec
}

type BundleDir struct {
	Path string
	Mode string
}

type BundleFile struct {
	Source        string
	Destination   string
	Mode          string
	CopyDirectory bool
}

type AgentSpec struct {
	Label     string
	Domain    string
	PlistPath string
}

type ProgramSpec struct {
	Argv             []string
	WorkingDirectory string
}

type LoggingSpec struct {
	StdoutPath string
	StderrPath string
}

type ServiceSpec struct {
	RunAtLoad             bool
	KeepAlive             bool
	ThrottleInterval      int
	Umask                 int
	StartCalendarInterval []CalendarInterval
	ProcessType           string
	Disabled              *bool
}

type CalendarInterval struct {
	Minute  *int
	Hour    *int
	Weekday *int
	Day     *int
	Month   *int
}

type InstallSpec struct {
	ValidatePlist           bool
	BootoutExisting         bool
	Bootstrap               bool
	KickstartAfterBootstrap bool
}

func LoadLaunchdfile(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Launchdfile %s: %w", path, err)
	}

	manifest := &Manifest{
		SourcePath:  path,
		ManifestDir: filepath.Dir(path),
		Environment: map[string]string{},
	}
	if err := manifest.parse(string(data)); err != nil {
		return nil, err
	}
	if err := manifest.normalizeAndValidate(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *Manifest) parse(body string) error {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(stripComment(scanner.Text()))
		if line == "" {
			continue
		}

		tokens, err := splitFields(line)
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
		if len(tokens) == 0 {
			continue
		}

		directive := strings.ToUpper(tokens[0])
		rest := strings.TrimSpace(line[len(tokens[0]):])
		switch directive {
		case "ROOT":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: ROOT expects exactly 1 argument", lineNo)
			}
			m.Root = tokens[1]
		case "MKDIR":
			dir, err := parseMkdir(tokens)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Directories = append(m.Directories, dir)
		case "COPY":
			file, err := parseCopy(tokens, false)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Files = append(m.Files, file)
		case "COPYDIR":
			file, err := parseCopy(tokens, true)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Files = append(m.Files, file)
		case "LABEL":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: LABEL expects exactly 1 argument", lineNo)
			}
			m.Agent.Label = tokens[1]
		case "DOMAIN":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: DOMAIN expects exactly 1 argument", lineNo)
			}
			m.Agent.Domain = tokens[1]
		case "PLIST":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: PLIST expects exactly 1 argument", lineNo)
			}
			m.Agent.PlistPath = tokens[1]
		case "WORKDIR":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: WORKDIR expects exactly 1 argument", lineNo)
			}
			m.Program.WorkingDirectory = tokens[1]
		case "CMD":
			if err := parseJSONArgv(rest, &m.Program.Argv); err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
		case "STDOUT":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: STDOUT expects exactly 1 argument", lineNo)
			}
			m.Logging.StdoutPath = tokens[1]
		case "STDERR":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: STDERR expects exactly 1 argument", lineNo)
			}
			m.Logging.StderrPath = tokens[1]
		case "ENV":
			raw := strings.TrimSpace(rest)
			if raw == "" {
				return fmt.Errorf("line %d: ENV expects KEY=value", lineNo)
			}
			key, value, ok := strings.Cut(raw, "=")
			if !ok {
				return fmt.Errorf("line %d: ENV expects KEY=value", lineNo)
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if m.Environment == nil {
				m.Environment = map[string]string{}
			}
			m.Environment[key] = value
		case "ENVFROM":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: ENVFROM expects exactly 1 argument", lineNo)
			}
			m.EnvFromHost = append(m.EnvFromHost, tokens[1])
		case "RUNATLOAD":
			value, err := parseDirectiveBool(tokens, "RUNATLOAD")
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Service.RunAtLoad = value
		case "KEEPALIVE":
			value, err := parseDirectiveBool(tokens, "KEEPALIVE")
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Service.KeepAlive = value
		case "THROTTLE":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: THROTTLE expects exactly 1 argument", lineNo)
			}
			value, err := strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: invalid THROTTLE value %q", lineNo, tokens[1])
			}
			m.Service.ThrottleInterval = value
		case "UMASK":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: UMASK expects exactly 1 argument", lineNo)
			}
			value, err := strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: invalid UMASK value %q", lineNo, tokens[1])
			}
			m.Service.Umask = value
		case "SCHEDULE":
			interval, err := parseSchedule(tokens)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Service.StartCalendarInterval = append(m.Service.StartCalendarInterval, interval)
		case "PROCESSTYPE":
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: PROCESSTYPE expects exactly 1 argument", lineNo)
			}
			m.Service.ProcessType = tokens[1]
		case "DISABLED":
			value, err := parseDirectiveBool(tokens, "DISABLED")
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Service.Disabled = &value
		case "INSTALL":
			install, err := parseInstall(tokens)
			if err != nil {
				return fmt.Errorf("line %d: %w", lineNo, err)
			}
			m.Install = install
		default:
			return fmt.Errorf("line %d: unknown directive %s", lineNo, directive)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan Launchdfile: %w", err)
	}
	return nil
}

func parseMkdir(tokens []string) (BundleDir, error) {
	if len(tokens) != 2 && len(tokens) != 4 {
		return BundleDir{}, errors.New("MKDIR expects <path> [MODE <octal>]")
	}
	dir := BundleDir{Path: tokens[1]}
	if len(tokens) == 4 {
		if strings.ToUpper(tokens[2]) != "MODE" {
			return BundleDir{}, errors.New("MKDIR expects MODE before the optional mode value")
		}
		dir.Mode = tokens[3]
	}
	return dir, nil
}

func parseCopy(tokens []string, copyDirectory bool) (BundleFile, error) {
	want := "COPY expects <source> <destination> [MODE <octal>]"
	if copyDirectory {
		want = "COPYDIR expects <source> <destination>"
	}
	if (!copyDirectory && len(tokens) != 3 && len(tokens) != 5) || (copyDirectory && len(tokens) != 3) {
		return BundleFile{}, errors.New(want)
	}
	file := BundleFile{
		Source:        tokens[1],
		Destination:   tokens[2],
		CopyDirectory: copyDirectory,
	}
	if !copyDirectory && len(tokens) == 5 {
		if strings.ToUpper(tokens[3]) != "MODE" {
			return BundleFile{}, errors.New("COPY expects MODE before the optional mode value")
		}
		file.Mode = tokens[4]
	}
	return file, nil
}

func parseJSONArgv(rest string, out *[]string) error {
	payload := strings.TrimSpace(rest)
	if payload == "" {
		return errors.New("CMD expects a JSON array")
	}
	var argv []string
	if err := json.Unmarshal([]byte(payload), &argv); err != nil {
		return fmt.Errorf("parse CMD JSON array: %w", err)
	}
	*out = argv
	return nil
}

func parseDirectiveBool(tokens []string, name string) (bool, error) {
	if len(tokens) != 2 {
		return false, fmt.Errorf("%s expects exactly 1 argument", name)
	}
	value, err := strconv.ParseBool(tokens[1])
	if err != nil {
		return false, fmt.Errorf("%s expects true or false", name)
	}
	return value, nil
}

func parseSchedule(tokens []string) (CalendarInterval, error) {
	if len(tokens) < 2 {
		return CalendarInterval{}, errors.New("SCHEDULE expects at least one key=value pair")
	}
	var interval CalendarInterval
	for _, token := range tokens[1:] {
		key, raw, ok := strings.Cut(token, "=")
		if !ok {
			return CalendarInterval{}, fmt.Errorf("invalid SCHEDULE token %q", token)
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return CalendarInterval{}, fmt.Errorf("invalid SCHEDULE value %q", token)
		}
		switch strings.ToLower(key) {
		case "minute":
			interval.Minute = intPtr(value)
		case "hour":
			interval.Hour = intPtr(value)
		case "weekday":
			interval.Weekday = intPtr(value)
		case "day":
			interval.Day = intPtr(value)
		case "month":
			interval.Month = intPtr(value)
		default:
			return CalendarInterval{}, fmt.Errorf("unknown SCHEDULE field %q", key)
		}
	}
	return interval, nil
}

func parseInstall(tokens []string) (InstallSpec, error) {
	install := InstallSpec{}
	for _, token := range tokens[1:] {
		key, raw, ok := strings.Cut(token, "=")
		if !ok {
			return InstallSpec{}, fmt.Errorf("invalid INSTALL token %q", token)
		}
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return InstallSpec{}, fmt.Errorf("invalid INSTALL boolean %q", token)
		}
		switch strings.ToLower(key) {
		case "validate":
			install.ValidatePlist = value
		case "bootout_existing":
			install.BootoutExisting = value
		case "bootstrap":
			install.Bootstrap = value
		case "kickstart":
			install.KickstartAfterBootstrap = value
		default:
			return InstallSpec{}, fmt.Errorf("unknown INSTALL option %q", key)
		}
	}
	return install, nil
}

func (m *Manifest) normalizeAndValidate() error {
	if strings.TrimSpace(m.Root) == "" {
		return errors.New("ROOT is required")
	}
	root, err := expandPath(m.Root, m.ManifestDir)
	if err != nil {
		return err
	}
	m.Root = root

	destinations := map[string]struct{}{}
	for i := range m.Directories {
		entry := &m.Directories[i]
		if strings.TrimSpace(entry.Path) == "" {
			return errors.New("MKDIR path is required")
		}
		if _, err := parseMode(entry.Mode, 0o755); err != nil {
			return fmt.Errorf("MKDIR[%d] mode: %w", i, err)
		}
		dest := filepath.Clean(filepath.Join(m.Root, entry.Path))
		if _, ok := destinations[dest]; ok {
			return fmt.Errorf("duplicate destination %s", dest)
		}
		destinations[dest] = struct{}{}
	}

	for i := range m.Files {
		entry := &m.Files[i]
		if strings.TrimSpace(entry.Source) == "" {
			return errors.New("COPY source is required")
		}
		if strings.TrimSpace(entry.Destination) == "" {
			return errors.New("COPY destination is required")
		}
		source, err := expandPath(entry.Source, m.ManifestDir)
		if err != nil {
			return err
		}
		entry.Source = source
		if !entry.CopyDirectory {
			if _, err := parseMode(entry.Mode, 0o644); err != nil {
				return fmt.Errorf("COPY[%d] mode: %w", i, err)
			}
		}
		dest := filepath.Clean(filepath.Join(m.Root, entry.Destination))
		if _, ok := destinations[dest]; ok {
			return fmt.Errorf("duplicate destination %s", dest)
		}
		destinations[dest] = struct{}{}
	}

	if strings.TrimSpace(m.Agent.Label) == "" {
		return errors.New("LABEL is required")
	}
	if strings.TrimSpace(m.Agent.Domain) == "" {
		m.Agent.Domain = "user"
	}
	if !slices.Contains([]string{"user", "system"}, m.Agent.Domain) {
		return fmt.Errorf("unsupported DOMAIN %q", m.Agent.Domain)
	}

	if len(m.Program.Argv) == 0 {
		return errors.New("CMD is required")
	}
	for i, arg := range m.Program.Argv {
		if looksLikePath(arg) {
			resolved, err := expandPath(arg, m.ManifestDir)
			if err != nil {
				return err
			}
			m.Program.Argv[i] = resolved
		}
	}
	if strings.TrimSpace(m.Program.WorkingDirectory) != "" {
		resolved, err := expandPath(m.Program.WorkingDirectory, m.ManifestDir)
		if err != nil {
			return err
		}
		m.Program.WorkingDirectory = resolved
	}

	if strings.TrimSpace(m.Logging.StdoutPath) == "" || strings.TrimSpace(m.Logging.StderrPath) == "" {
		return errors.New("STDOUT and STDERR are required")
	}
	stdout, err := expandPath(m.Logging.StdoutPath, m.ManifestDir)
	if err != nil {
		return err
	}
	stderr, err := expandPath(m.Logging.StderrPath, m.ManifestDir)
	if err != nil {
		return err
	}
	m.Logging.StdoutPath = stdout
	m.Logging.StderrPath = stderr

	for key, value := range m.Environment {
		if strings.TrimSpace(key) == "" {
			return errors.New("ENV keys must not be empty")
		}
		if looksLikePath(value) {
			resolved, err := expandPath(value, m.ManifestDir)
			if err != nil {
				return err
			}
			m.Environment[key] = resolved
		}
	}
	for i, key := range m.EnvFromHost {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			return fmt.Errorf("ENVFROM[%d] must not be empty", i)
		}
		m.EnvFromHost[i] = trimmed
	}

	if m.Service.ThrottleInterval < 0 {
		return errors.New("THROTTLE must be >= 0")
	}
	if m.Service.Umask < 0 {
		return errors.New("UMASK must be >= 0")
	}
	if m.Service.ProcessType != "" && !slices.Contains([]string{"standard", "background", "adaptive", "interactive"}, m.Service.ProcessType) {
		return fmt.Errorf("unsupported PROCESSTYPE %q", m.Service.ProcessType)
	}
	for i, interval := range m.Service.StartCalendarInterval {
		if interval.Minute != nil && (*interval.Minute < 0 || *interval.Minute > 59) {
			return fmt.Errorf("SCHEDULE[%d].minute must be 0-59", i)
		}
		if interval.Hour != nil && (*interval.Hour < 0 || *interval.Hour > 23) {
			return fmt.Errorf("SCHEDULE[%d].hour must be 0-23", i)
		}
		if interval.Weekday != nil && (*interval.Weekday < 0 || *interval.Weekday > 7) {
			return fmt.Errorf("SCHEDULE[%d].weekday must be 0-7", i)
		}
		if interval.Day != nil && (*interval.Day < 1 || *interval.Day > 31) {
			return fmt.Errorf("SCHEDULE[%d].day must be 1-31", i)
		}
		if interval.Month != nil && (*interval.Month < 1 || *interval.Month > 12) {
			return fmt.Errorf("SCHEDULE[%d].month must be 1-12", i)
		}
	}

	if strings.TrimSpace(m.Agent.PlistPath) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("resolve user home: %w", err)
		}
		switch m.Agent.Domain {
		case "user":
			m.Agent.PlistPath = filepath.Join(home, "Library", "LaunchAgents", m.Agent.Label+".plist")
		case "system":
			m.Agent.PlistPath = filepath.Join(string(filepath.Separator), "Library", "LaunchDaemons", m.Agent.Label+".plist")
		}
	} else {
		resolved, err := expandPath(m.Agent.PlistPath, m.ManifestDir)
		if err != nil {
			return err
		}
		m.Agent.PlistPath = resolved
	}

	if m.Install.KickstartAfterBootstrap && !m.Install.Bootstrap {
		return errors.New("INSTALL kickstart=true requires bootstrap=true")
	}
	return nil
}

func parseMode(raw string, fallback os.FileMode) (os.FileMode, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	value, err := strconv.ParseUint(trimmed, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid mode %q", raw)
	}
	return os.FileMode(value), nil
}

func ModeOrDefault(raw string, fallback os.FileMode) (os.FileMode, error) {
	return parseMode(raw, fallback)
}

func expandPath(value string, baseDir string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed, nil
	}
	if strings.HasPrefix(trimmed, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home: %w", err)
		}
		if trimmed == "~" {
			return home, nil
		}
		if strings.HasPrefix(trimmed, "~/") {
			return filepath.Join(home, trimmed[2:]), nil
		}
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed), nil
	}
	return filepath.Clean(filepath.Join(baseDir, trimmed)), nil
}

func looksLikePath(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	return strings.HasPrefix(trimmed, "~") ||
		strings.HasPrefix(trimmed, "./") ||
		strings.HasPrefix(trimmed, "../") ||
		strings.HasPrefix(trimmed, "/")
}

func stripComment(line string) string {
	var out strings.Builder
	var quote rune
	escaped := false
	for _, r := range line {
		if escaped {
			out.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			out.WriteRune(r)
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			out.WriteRune(r)
			continue
		}
		if r == '"' || r == '\'' {
			quote = r
			out.WriteRune(r)
			continue
		}
		if r == '#' {
			break
		}
		out.WriteRune(r)
	}
	return out.String()
}

func splitFields(line string) ([]string, error) {
	var fields []string
	var current strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}

	for _, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		switch {
		case quote != 0:
			if r == '\\' && quote == '"' {
				escaped = true
				continue
			}
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
		case r == '\\':
			escaped = true
		case r == ' ' || r == '\t':
			flush()
		default:
			current.WriteRune(r)
		}
	}
	if escaped || quote != 0 {
		return nil, errors.New("unterminated quoted field")
	}
	flush()
	return fields, nil
}

func intPtr(v int) *int {
	return &v
}
