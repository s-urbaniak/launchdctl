package launchd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"howett.net/plist"

	"launchdctl/internal/spec"
)

type Dependencies struct {
	RunCommand func(context.Context, string, ...string) error
	GetUID     func() int
	HomeDir    func() (string, error)
}

func Apply(ctx context.Context, manifest *spec.Manifest, deps *Dependencies) error {
	deps = withDefaults(deps)

	env := map[string]string{}
	for key, value := range manifest.Environment {
		env[key] = value
	}
	for _, key := range manifest.EnvFromHost {
		if value := os.Getenv(key); value != "" {
			env[key] = value
		}
	}

	content, err := BuildPlist(manifest, env)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(manifest.Agent.PlistPath), 0o755); err != nil {
		return fmt.Errorf("create plist dir: %w", err)
	}
	if err := os.WriteFile(manifest.Agent.PlistPath, content, 0o644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(manifest.Logging.StdoutPath), 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(manifest.Logging.StderrPath), 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	if manifest.Install.ValidatePlist {
		if err := deps.RunCommand(ctx, "/usr/bin/plutil", "-lint", manifest.Agent.PlistPath); err != nil {
			return fmt.Errorf("validate plist: %w", err)
		}
	}

	if !manifest.Install.Bootstrap {
		return nil
	}

	target := domainTarget(manifest.Agent.Domain, deps.GetUID(), manifest.Agent.Label)
	domain := bootstrapDomain(manifest.Agent.Domain, deps.GetUID())
	if manifest.Install.BootoutExisting {
		_ = deps.RunCommand(ctx, "/bin/launchctl", "bootout", target)
		_ = deps.RunCommand(ctx, "/bin/launchctl", "bootout", domain, manifest.Agent.PlistPath)
	}
	if err := deps.RunCommand(ctx, "/bin/launchctl", "bootstrap", domain, manifest.Agent.PlistPath); err != nil {
		return fmt.Errorf("bootstrap launch agent: %w", err)
	}
	if manifest.Install.KickstartAfterBootstrap {
		if err := deps.RunCommand(ctx, "/bin/launchctl", "kickstart", "-k", target); err != nil {
			return fmt.Errorf("kickstart launch agent: %w", err)
		}
	}
	return nil
}

func BuildPlist(manifest *spec.Manifest, environment map[string]string) ([]byte, error) {
	doc := map[string]any{
		"Label":             manifest.Agent.Label,
		"ProgramArguments":  manifest.Program.Argv,
		"StandardOutPath":   manifest.Logging.StdoutPath,
		"StandardErrorPath": manifest.Logging.StderrPath,
	}
	if manifest.Program.WorkingDirectory != "" {
		doc["WorkingDirectory"] = manifest.Program.WorkingDirectory
	}
	if len(environment) > 0 {
		doc["EnvironmentVariables"] = environment
	}
	doc["RunAtLoad"] = manifest.Service.RunAtLoad
	doc["KeepAlive"] = manifest.Service.KeepAlive
	if manifest.Service.ProcessType != "" {
		doc["ProcessType"] = manifest.Service.ProcessType
	}
	if manifest.Service.Disabled != nil {
		doc["Disabled"] = *manifest.Service.Disabled
	}
	if manifest.Service.ThrottleInterval > 0 {
		doc["ThrottleInterval"] = manifest.Service.ThrottleInterval
	}
	if manifest.Service.Umask > 0 {
		doc["Umask"] = manifest.Service.Umask
	}
	if len(manifest.Service.StartCalendarInterval) > 0 {
		intervals := make([]map[string]int, 0, len(manifest.Service.StartCalendarInterval))
		for _, interval := range manifest.Service.StartCalendarInterval {
			item := map[string]int{}
			if interval.Minute != nil {
				item["Minute"] = *interval.Minute
			}
			if interval.Hour != nil {
				item["Hour"] = *interval.Hour
			}
			if interval.Weekday != nil {
				item["Weekday"] = *interval.Weekday
			}
			if interval.Day != nil {
				item["Day"] = *interval.Day
			}
			if interval.Month != nil {
				item["Month"] = *interval.Month
			}
			intervals = append(intervals, item)
		}
		if len(intervals) == 1 {
			doc["StartCalendarInterval"] = intervals[0]
		} else {
			doc["StartCalendarInterval"] = intervals
		}
	}
	var buf bytes.Buffer
	enc := plist.NewEncoderForFormat(&buf, plist.XMLFormat)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encode plist: %w", err)
	}
	return buf.Bytes(), nil
}

func withDefaults(deps *Dependencies) *Dependencies {
	if deps == nil {
		deps = &Dependencies{}
	}
	if deps.RunCommand == nil {
		deps.RunCommand = func(ctx context.Context, name string, args ...string) error {
			cmd := exec.CommandContext(ctx, name, args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s %v: %w: %s", name, args, err, string(output))
			}
			return nil
		}
	}
	if deps.GetUID == nil {
		deps.GetUID = os.Getuid
	}
	if deps.HomeDir == nil {
		deps.HomeDir = os.UserHomeDir
	}
	return deps
}

func bootstrapDomain(domain string, uid int) string {
	if domain == "system" {
		return "system"
	}
	return "gui/" + strconv.Itoa(uid)
}

func domainTarget(domain string, uid int, label string) string {
	return bootstrapDomain(domain, uid) + "/" + label
}
