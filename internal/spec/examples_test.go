package spec

import (
	"path/filepath"
	"testing"
)

func TestExamplesParse(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	for _, path := range []string{
		filepath.Join(root, "examples", "openclaw", "bundle.yaml"),
		filepath.Join(root, "examples", "openclaw", "install.yaml"),
		filepath.Join(root, "examples", "backup-agent", "bundle.yaml"),
		filepath.Join(root, "examples", "backup-agent", "install.yaml"),
	} {
		var err error
		if filepath.Base(path) == "bundle.yaml" {
			_, err = LoadBundleFile(path)
		} else {
			_, err = LoadInstallFile(path)
		}
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
	}
}
