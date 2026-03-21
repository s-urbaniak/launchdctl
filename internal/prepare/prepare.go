package prepare

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"launchdctl/internal/spec"
)

type Dependencies struct {
	RunCommand func(context.Context, string, []string, string, []string) error
}

func Apply(ctx context.Context, manifest *spec.Manifest, deps *Dependencies) error {
	deps = withDefaults(deps)

	for _, step := range manifest.Prepare {
		if err := deps.RunCommand(ctx, step.Argv[0], step.Argv[1:], manifest.ManifestDir, os.Environ()); err != nil {
			return fmt.Errorf("run %q: %w", step.Argv[0], err)
		}
	}

	return nil
}

func withDefaults(deps *Dependencies) *Dependencies {
	if deps == nil {
		deps = &Dependencies{}
	}
	if deps.RunCommand == nil {
		deps.RunCommand = func(ctx context.Context, name string, args []string, dir string, env []string) error {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = dir
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("%s %v: %w", name, args, err)
			}
			return nil
		}
	}
	return deps
}
