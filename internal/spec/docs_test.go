package spec_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"launchdctl/internal/launchd"
	"launchdctl/internal/spec"
)

func TestDocsReferenceExistingFiles(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	docsPath := filepath.Join(root, "docs", "examples.md")
	data, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, ref := range []string{
		"examples/openclaw/inputs/node",
		"examples/openclaw/inputs/openclaw-dist/",
		"examples/openclaw/inputs/config.json",
		"examples/backup-agent/inputs/restic",
		"examples/backup-agent/inputs/config.yaml",
		"examples/backup-agent/inputs/signal-cli/",
		"examples/backup-agent/inputs/jre/",
	} {
		if !strings.Contains(body, ref) {
			t.Fatalf("docs/examples.md missing reference to %s", ref)
		}
	}
}

func TestPlistMappingDocMatchesEmittedKeys(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	data, err := os.ReadFile(filepath.Join(root, "docs", "plist-mapping.md"))
	if err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	for _, key := range []string{
		"Label",
		"ProgramArguments",
		"WorkingDirectory",
		"EnvironmentVariables",
		"ProcessType",
		"Disabled",
		"RunAtLoad",
		"KeepAlive",
		"ThrottleInterval",
		"Umask",
		"StartCalendarInterval",
		"StandardOutPath",
		"StandardErrorPath",
	} {
		if !strings.Contains(doc, key) {
			t.Fatalf("plist mapping doc missing %s", key)
		}
	}

	disabled := false
	manifest := &spec.Manifest{
		Agent: spec.AgentSpec{Label: "com.example.test"},
		Program: spec.ProgramSpec{
			Argv:             []string{"/bin/echo", "hello"},
			WorkingDirectory: "/tmp",
		},
		Logging: spec.LoggingSpec{
			StdoutPath: "/tmp/stdout.log",
			StderrPath: "/tmp/stderr.log",
		},
		Service: spec.ServiceSpec{
			RunAtLoad:        true,
			KeepAlive:        true,
			ThrottleInterval: 1,
			Umask:            63,
			ProcessType:      "background",
			Disabled:         &disabled,
			StartCalendarInterval: []spec.CalendarInterval{
				{Hour: intPtr(2), Minute: intPtr(0)},
			},
		},
	}
	plistBytes, err := launchd.BuildPlist(manifest, map[string]string{"FOO": "bar"})
	if err != nil {
		t.Fatal(err)
	}
	plistBody := string(plistBytes)
	for _, key := range []string{
		"Label",
		"ProgramArguments",
		"WorkingDirectory",
		"EnvironmentVariables",
		"ProcessType",
		"Disabled",
		"RunAtLoad",
		"KeepAlive",
		"ThrottleInterval",
		"Umask",
		"StartCalendarInterval",
		"StandardOutPath",
		"StandardErrorPath",
	} {
		if !strings.Contains(plistBody, key) {
			t.Fatalf("emitted plist missing %s", key)
		}
	}
}

func intPtr(v int) *int {
	return &v
}
