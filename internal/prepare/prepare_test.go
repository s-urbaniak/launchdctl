package prepare

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"launchdctl/internal/spec"
)

func TestApplyRunsStepsInOrderWithManifestDir(t *testing.T) {
	dir := t.TempDir()
	manifest := &spec.Manifest{
		ManifestDir: dir,
		Prepare: []spec.ExecStep{
			{Argv: []string{"/bin/echo", "first"}},
			{Argv: []string{"/bin/echo", "second"}},
		},
	}

	type call struct {
		name string
		args []string
		dir  string
	}
	var calls []call
	deps := &Dependencies{
		RunCommand: func(_ context.Context, name string, args []string, workdir string, env []string) error {
			calls = append(calls, call{name: name, args: append([]string(nil), args...), dir: workdir})
			if len(env) == 0 {
				t.Fatal("expected inherited environment")
			}
			return nil
		},
	}

	if err := Apply(context.Background(), manifest, deps); err != nil {
		t.Fatal(err)
	}

	want := []call{
		{name: "/bin/echo", args: []string{"first"}, dir: dir},
		{name: "/bin/echo", args: []string{"second"}, dir: dir},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func TestApplyStopsOnFirstError(t *testing.T) {
	manifest := &spec.Manifest{
		ManifestDir: t.TempDir(),
		Prepare: []spec.ExecStep{
			{Argv: []string{"/bin/false"}},
			{Argv: []string{"/bin/echo", "never"}},
		},
	}

	var calls int
	deps := &Dependencies{
		RunCommand: func(_ context.Context, name string, args []string, workdir string, env []string) error {
			calls++
			return errors.New("boom")
		},
	}

	err := Apply(context.Background(), manifest, deps)
	if err == nil || err.Error() != `run "/bin/false": boom` {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestApplyUsesExpandedPathLikeArgs(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "Launchdfile")
	sourcePath := filepath.Join(dir, "inputs", "tool")
	if err := os.WriteFile(manifestPath, []byte(`
RUN ["./inputs/tool","--flag"]
ROOT ./app
LABEL com.example.service
CMD ["/bin/echo","hi"]
STDOUT ./logs/stdout.log
STDERR ./logs/stderr.log
`), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := spec.LoadLaunchdfile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := manifest.Prepare[0].Argv[0]; got != sourcePath {
		t.Fatalf("unexpected expanded path %s", got)
	}
}
