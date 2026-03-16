package bundle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"launchdctl/internal/spec"
)

func Apply(manifest *spec.BundleManifest) error {
	for _, dir := range manifest.Directories {
		mode, err := spec.ModeOrDefault(dir.Mode, 0o755)
		if err != nil {
			return err
		}
		target := filepath.Join(manifest.Bundle.Root, dir.Path)
		if err := os.MkdirAll(target, mode); err != nil {
			return fmt.Errorf("create directory %s: %w", target, err)
		}
	}
	for _, file := range manifest.Files {
		target := filepath.Join(manifest.Bundle.Root, file.Destination)
		if file.CopyDirectory {
			if err := copyDirectory(file.Source, target); err != nil {
				return err
			}
			continue
		}
		mode, err := spec.ModeOrDefault(file.Mode, 0o644)
		if err != nil {
			return err
		}
		if err := copyFile(file.Source, target, mode); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", destination, err)
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", destination, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s to %s: %w", source, destination, err)
	}
	return nil
}

func copyDirectory(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if info.IsDir() {
			if err := os.MkdirAll(target, info.Mode().Perm()|0o700); err != nil {
				return fmt.Errorf("create dir %s: %w", target, err)
			}
			return nil
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}
