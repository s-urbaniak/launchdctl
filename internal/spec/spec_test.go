package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadBundleFileResolvesPaths(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "bundle.yaml")
	if err := os.WriteFile(manifestPath, []byte(`
bundle:
  root: ~/Library/Application Support/example
directories:
  - path: bin
    mode: "0755"
files:
  - source: ./input/file.txt
    destination: config/file.txt
    mode: "0644"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadBundleFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Bundle.Root, filepath.Join("Library", "Application Support", "example")) {
		t.Fatalf("unexpected root %s", got.Bundle.Root)
	}
	if got.Files[0].Source != filepath.Join(dir, "input", "file.txt") {
		t.Fatalf("unexpected source %s", got.Files[0].Source)
	}
}

func TestLoadInstallFileDefaultsUserPlistPath(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "install.yaml")
	if err := os.WriteFile(manifestPath, []byte(`
agent:
  label: com.example.service
  domain: user
program:
  argv:
    - ./bin/app
logging:
  stdout_path: ./logs/stdout.log
  stderr_path: ./logs/stderr.log
service:
  run_at_load: true
install:
  validate_plist: true
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadInstallFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got.Agent.PlistPath) != "com.example.service.plist" {
		t.Fatalf("unexpected plist path %s", got.Agent.PlistPath)
	}
}
