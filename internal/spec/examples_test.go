package spec

import (
	"path/filepath"
	"testing"
)

func TestExamplesParse(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	for _, path := range []string{
		filepath.Join(root, "examples", "openclaw", "Launchdfile"),
		filepath.Join(root, "examples", "backup-agent", "Launchdfile"),
	} {
		_, err := LoadLaunchdfile(path)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
	}
}
