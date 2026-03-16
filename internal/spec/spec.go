package spec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type BundleManifest struct {
	SourcePath  string
	ManifestDir string
	Bundle      BundleRoot   `yaml:"bundle"`
	Directories []BundleDir  `yaml:"directories"`
	Files       []BundleFile `yaml:"files"`
}

type BundleRoot struct {
	Root string `yaml:"root"`
}

type BundleDir struct {
	Path string `yaml:"path"`
	Mode string `yaml:"mode"`
}

type BundleFile struct {
	Source        string `yaml:"source"`
	Destination   string `yaml:"destination"`
	Mode          string `yaml:"mode"`
	CopyDirectory bool   `yaml:"copy_directory"`
}

type InstallManifest struct {
	SourcePath  string
	ManifestDir string
	Agent       AgentSpec         `yaml:"agent"`
	Program     ProgramSpec       `yaml:"program"`
	Logging     LoggingSpec       `yaml:"logging"`
	Environment map[string]string `yaml:"environment"`
	EnvFromHost []string          `yaml:"env_from_host"`
	Service     ServiceSpec       `yaml:"service"`
	Install     InstallSpec       `yaml:"install"`
}

type AgentSpec struct {
	Label     string `yaml:"label"`
	Domain    string `yaml:"domain"`
	PlistPath string `yaml:"plist_path"`
}

type ProgramSpec struct {
	Argv             []string `yaml:"argv"`
	WorkingDirectory string   `yaml:"working_directory"`
}

type LoggingSpec struct {
	StdoutPath string `yaml:"stdout_path"`
	StderrPath string `yaml:"stderr_path"`
}

type ServiceSpec struct {
	RunAtLoad             bool               `yaml:"run_at_load"`
	KeepAlive             bool               `yaml:"keep_alive"`
	ThrottleInterval      int                `yaml:"throttle_interval"`
	Umask                 int                `yaml:"umask"`
	StartCalendarInterval []CalendarInterval `yaml:"start_calendar_interval"`
}

type CalendarInterval struct {
	Minute  *int `yaml:"minute"`
	Hour    *int `yaml:"hour"`
	Weekday *int `yaml:"weekday"`
	Day     *int `yaml:"day"`
	Month   *int `yaml:"month"`
}

type InstallSpec struct {
	ValidatePlist           bool `yaml:"validate_plist"`
	BootoutExisting         bool `yaml:"bootout_existing"`
	Bootstrap               bool `yaml:"bootstrap"`
	KickstartAfterBootstrap bool `yaml:"kickstart_after_bootstrap"`
}

func LoadBundleFile(path string) (*BundleManifest, error) {
	var manifest BundleManifest
	if err := loadFile(path, &manifest); err != nil {
		return nil, err
	}
	manifest.SourcePath = path
	manifest.ManifestDir = filepath.Dir(path)
	if err := manifest.normalizeAndValidate(); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func LoadInstallFile(path string) (*InstallManifest, error) {
	var manifest InstallManifest
	if err := loadFile(path, &manifest); err != nil {
		return nil, err
	}
	manifest.SourcePath = path
	manifest.ManifestDir = filepath.Dir(path)
	if err := manifest.normalizeAndValidate(); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func loadFile(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read manifest %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse manifest %s: %w", path, err)
	}
	return nil
}

func (m *BundleManifest) normalizeAndValidate() error {
	if strings.TrimSpace(m.Bundle.Root) == "" {
		return errors.New("bundle.root is required")
	}
	root, err := expandPath(m.Bundle.Root, m.ManifestDir)
	if err != nil {
		return err
	}
	m.Bundle.Root = root

	destinations := map[string]struct{}{}
	for i := range m.Directories {
		entry := &m.Directories[i]
		if strings.TrimSpace(entry.Path) == "" {
			return errors.New("directories[].path is required")
		}
		if _, err := parseMode(entry.Mode, 0o755); err != nil {
			return fmt.Errorf("directories[%d].mode: %w", i, err)
		}
		dest := filepath.Clean(filepath.Join(m.Bundle.Root, entry.Path))
		if _, ok := destinations[dest]; ok {
			return fmt.Errorf("duplicate destination %s", dest)
		}
		destinations[dest] = struct{}{}
	}

	for i := range m.Files {
		entry := &m.Files[i]
		if strings.TrimSpace(entry.Source) == "" {
			return errors.New("files[].source is required")
		}
		if strings.TrimSpace(entry.Destination) == "" {
			return errors.New("files[].destination is required")
		}
		source, err := expandPath(entry.Source, m.ManifestDir)
		if err != nil {
			return err
		}
		entry.Source = source
		if !entry.CopyDirectory {
			if _, err := parseMode(entry.Mode, 0o644); err != nil {
				return fmt.Errorf("files[%d].mode: %w", i, err)
			}
		}
		dest := filepath.Clean(filepath.Join(m.Bundle.Root, entry.Destination))
		if _, ok := destinations[dest]; ok {
			return fmt.Errorf("duplicate destination %s", dest)
		}
		destinations[dest] = struct{}{}
	}

	return nil
}

func (m *InstallManifest) normalizeAndValidate() error {
	if strings.TrimSpace(m.Agent.Label) == "" {
		return errors.New("agent.label is required")
	}
	if strings.TrimSpace(m.Agent.Domain) == "" {
		m.Agent.Domain = "user"
	}
	if !slices.Contains([]string{"user", "system"}, m.Agent.Domain) {
		return fmt.Errorf("unsupported agent.domain %q", m.Agent.Domain)
	}
	if len(m.Program.Argv) == 0 {
		return errors.New("program.argv is required")
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
		return errors.New("logging.stdout_path and logging.stderr_path are required")
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
			return errors.New("environment keys must not be empty")
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
			return fmt.Errorf("env_from_host[%d] must not be empty", i)
		}
		m.EnvFromHost[i] = trimmed
	}
	if m.Service.ThrottleInterval < 0 {
		return errors.New("service.throttle_interval must be >= 0")
	}
	if m.Service.Umask < 0 {
		return errors.New("service.umask must be >= 0")
	}
	for i, interval := range m.Service.StartCalendarInterval {
		if interval.Minute != nil && (*interval.Minute < 0 || *interval.Minute > 59) {
			return fmt.Errorf("service.start_calendar_interval[%d].minute must be 0-59", i)
		}
		if interval.Hour != nil && (*interval.Hour < 0 || *interval.Hour > 23) {
			return fmt.Errorf("service.start_calendar_interval[%d].hour must be 0-23", i)
		}
		if interval.Weekday != nil && (*interval.Weekday < 0 || *interval.Weekday > 7) {
			return fmt.Errorf("service.start_calendar_interval[%d].weekday must be 0-7", i)
		}
		if interval.Day != nil && (*interval.Day < 1 || *interval.Day > 31) {
			return fmt.Errorf("service.start_calendar_interval[%d].day must be 1-31", i)
		}
		if interval.Month != nil && (*interval.Month < 1 || *interval.Month > 12) {
			return fmt.Errorf("service.start_calendar_interval[%d].month must be 1-12", i)
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
