package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadLaunchdfileResolvesPaths(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "Launchdfile")
	if err := os.WriteFile(manifestPath, []byte(`
RUN ["./tools/prepare","--out","./dist/output"]

ROOT "~/Library/Application Support/example"

MKDIR bin MODE 0755
COPY "./input/file.txt" "config/file.txt" MODE 0644

LABEL com.example.service
CMD ["./bin/app"]
STDOUT "./logs/stdout.log"
STDERR "./logs/stderr.log"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadLaunchdfile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Root, filepath.Join("Library", "Application Support", "example")) {
		t.Fatalf("unexpected root %s", got.Root)
	}
	if got.Files[0].Source != filepath.Join(dir, "input", "file.txt") {
		t.Fatalf("unexpected source %s", got.Files[0].Source)
	}
	if got.Prepare[0].Argv[0] != filepath.Join(dir, "tools", "prepare") {
		t.Fatalf("unexpected prepare argv[0] %s", got.Prepare[0].Argv[0])
	}
	if got.Prepare[0].Argv[2] != filepath.Join(dir, "dist", "output") {
		t.Fatalf("unexpected prepare argv[2] %s", got.Prepare[0].Argv[2])
	}
	if got.Program.Argv[0] != filepath.Join(dir, "bin", "app") {
		t.Fatalf("unexpected argv[0] %s", got.Program.Argv[0])
	}
	if filepath.Base(got.Agent.PlistPath) != "com.example.service.plist" {
		t.Fatalf("unexpected plist path %s", got.Agent.PlistPath)
	}
}

func TestLoadLaunchdfileRejectsKickstartWithoutBootstrap(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "Launchdfile")
	if err := os.WriteFile(manifestPath, []byte(`
ROOT ./app
LABEL com.example.service
CMD ["/bin/echo","hi"]
STDOUT ./logs/stdout.log
STDERR ./logs/stderr.log
INSTALL kickstart=true bootstrap=false
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadLaunchdfile(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "kickstart=true requires bootstrap=true") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadLaunchdfileRejectsRunAfterOtherDirectives(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "Launchdfile")
	if err := os.WriteFile(manifestPath, []byte(`
ROOT ./app
LABEL com.example.service
CMD ["/bin/echo","hi"]
STDOUT ./logs/stdout.log
STDERR ./logs/stderr.log
RUN ["/bin/echo","nope"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadLaunchdfile(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "RUN must appear before all other directives") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadLaunchdfileRejectsEmptyRunArray(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "Launchdfile")
	if err := os.WriteFile(manifestPath, []byte(`
RUN []
ROOT ./app
LABEL com.example.service
CMD ["/bin/echo","hi"]
STDOUT ./logs/stdout.log
STDERR ./logs/stderr.log
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadLaunchdfile(manifestPath)
	if err == nil || !strings.Contains(err.Error(), "RUN expects a non-empty JSON array") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadLaunchdfileResolvesRelativeManifestPathAgainstWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.WriteFile("Launchdfile", []byte(`
RUN ["go","build","-o","./dist/output","./cmd/example"]
ROOT ./app
LABEL com.example.service
CMD ["./bin/app"]
STDOUT ./logs/stdout.log
STDERR ./logs/stderr.log
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadLaunchdfile("Launchdfile")
	if err != nil {
		t.Fatal(err)
	}

	if got.ManifestDir != dir {
		t.Fatalf("unexpected manifest dir %s", got.ManifestDir)
	}
	if got.Prepare[0].Argv[3] != filepath.Join(dir, "dist", "output") {
		t.Fatalf("unexpected prepare output path %s", got.Prepare[0].Argv[3])
	}
	if got.Prepare[0].Argv[4] != filepath.Join(dir, "cmd", "example") {
		t.Fatalf("unexpected prepare package path %s", got.Prepare[0].Argv[4])
	}
}
