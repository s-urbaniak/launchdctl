package bundle

import (
	"os"
	"path/filepath"
	"testing"

	"launchdctl/internal/spec"
)

func TestApplyCopiesFilesAndDirectories(t *testing.T) {
	src := t.TempDir()
	root := t.TempDir()

	filePath := filepath.Join(src, "input.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirPath := filepath.Join(src, "tree")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dirPath, "nested.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest := &spec.Manifest{
		Root: root,
		Directories: []spec.BundleDir{
			{Path: "bin", Mode: "0755"},
		},
		Files: []spec.BundleFile{
			{Source: filePath, Destination: "bin/app", Mode: "0755"},
			{Source: dirPath, Destination: "vendor/tree", CopyDirectory: true},
		},
	}
	if err := Apply(manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "app")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "vendor", "tree", "nested.txt")); err != nil {
		t.Fatal(err)
	}
}
