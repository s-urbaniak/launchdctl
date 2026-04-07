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

func TestApplyReplacesReadOnlyManagedFiles(t *testing.T) {
	src := t.TempDir()
	root := t.TempDir()

	filePath := filepath.Join(src, "input.txt")
	if err := os.WriteFile(filePath, []byte("updated"), 0o444); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(root, "bin", "app")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), 0o444); err != nil {
		t.Fatal(err)
	}

	manifest := &spec.Manifest{
		Root: root,
		Files: []spec.BundleFile{
			{Source: filePath, Destination: "bin/app", Mode: "0755"},
		},
	}

	if err := Apply(manifest); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "updated" {
		t.Fatalf("unexpected file content %q", string(body))
	}
}

func TestApplyCopiesWrappedExecutablePayload(t *testing.T) {
	src := t.TempDir()
	root := t.TempDir()

	wrappedDir := filepath.Join(src, "store", "restic", "bin")
	if err := os.MkdirAll(wrappedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	wrappedPath := filepath.Join(wrappedDir, ".restic-wrapped")
	if err := os.WriteFile(wrappedPath, []byte("real restic payload"), 0o755); err != nil {
		t.Fatal(err)
	}

	wrapperPath := filepath.Join(src, "restic")
	wrapperContent := "#!/bin/sh\nmakeCWrapper '" + wrappedPath + "'\n"
	if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := &spec.Manifest{
		Root: root,
		Files: []spec.BundleFile{
			{Source: wrapperPath, Destination: "bin/restic", Mode: "0755"},
		},
	}

	if err := Apply(manifest); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(root, "bin", "restic"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "real restic payload" {
		t.Fatalf("unexpected file content %q", string(body))
	}
}
