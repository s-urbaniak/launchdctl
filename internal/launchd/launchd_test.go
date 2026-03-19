package launchd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"launchdctl/internal/spec"
)

func intptr(v int) *int { return &v }

func TestBuildPlistIncludesServiceFields(t *testing.T) {
	disabled := false
	manifest := &spec.Manifest{
		Agent:   spec.AgentSpec{Label: "ai.openclaw.gateway"},
		Program: spec.ProgramSpec{Argv: []string{"/bin/node", "app.js"}},
		Logging: spec.LoggingSpec{StdoutPath: "/tmp/stdout.log", StderrPath: "/tmp/stderr.log"},
		Service: spec.ServiceSpec{
			RunAtLoad:        true,
			KeepAlive:        true,
			ThrottleInterval: 1,
			Umask:            63,
			ProcessType:      "background",
			Disabled:         &disabled,
		},
	}
	data, err := BuildPlist(manifest, map[string]string{"FOO": "bar"})
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, needle := range []string{"Label", "ThrottleInterval", "Umask", "EnvironmentVariables", "KeepAlive", "ProcessType", "Disabled"} {
		if !strings.Contains(body, needle) {
			t.Fatalf("plist missing %s: %s", needle, body)
		}
	}
}

func TestApplyRunsValidationAndBootstrap(t *testing.T) {
	dir := t.TempDir()
	manifest := &spec.Manifest{
		Agent: spec.AgentSpec{
			Label:     "com.example.app",
			Domain:    "user",
			PlistPath: filepath.Join(dir, "LaunchAgents", "com.example.app.plist"),
		},
		Program: spec.ProgramSpec{Argv: []string{"/bin/echo", "hi"}},
		Logging: spec.LoggingSpec{
			StdoutPath: filepath.Join(dir, "logs", "stdout.log"),
			StderrPath: filepath.Join(dir, "logs", "stderr.log"),
		},
		Service: spec.ServiceSpec{
			RunAtLoad: false,
			KeepAlive: false,
			StartCalendarInterval: []spec.CalendarInterval{
				{Hour: intptr(2), Minute: intptr(0)},
			},
		},
		Install: spec.InstallSpec{
			ValidatePlist:   true,
			BootoutExisting: true,
			Bootstrap:       true,
		},
	}
	var calls []string
	deps := &Dependencies{
		RunCommand: func(_ context.Context, name string, args ...string) error {
			calls = append(calls, name+" "+strings.Join(args, " "))
			return nil
		},
		GetUID: func() int { return 123 },
	}
	if err := Apply(context.Background(), manifest, deps); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(manifest.Agent.PlistPath); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 4 {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}
