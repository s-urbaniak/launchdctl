package bundle

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"launchdctl/internal/spec"
)

var nixWrappedExecutablePattern = regexp.MustCompile(`(/[^[:space:]'"]+/bin/\.[^[:space:]'"]+-wrapped)`)

func Apply(manifest *spec.Manifest) error {
	for _, dir := range manifest.Directories {
		mode, err := spec.ModeOrDefault(dir.Mode, 0o755)
		if err != nil {
			return err
		}
		target := filepath.Join(manifest.Root, dir.Path)
		if err := os.MkdirAll(target, mode); err != nil {
			return fmt.Errorf("create directory %s: %w", target, err)
		}
	}
	for _, file := range manifest.Files {
		target := filepath.Join(manifest.Root, file.Destination)
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
	resolvedSource, err := resolveCopySource(source)
	if err != nil {
		return err
	}

	in, err := os.Open(resolvedSource)
	if err != nil {
		return fmt.Errorf("open %s: %w", resolvedSource, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", destination, err)
	}
	if err := removeExistingDestination(destination); err != nil {
		return err
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create %s: %w", destination, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s to %s: %w", resolvedSource, destination, err)
	}
	return nil
}

func resolveCopySource(source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", source, err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return source, nil
	}

	content, err := os.ReadFile(source)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", source, err)
	}
	match := nixWrappedExecutablePattern.FindSubmatch(content)
	if len(match) < 2 {
		return source, nil
	}

	wrapped := string(bytes.TrimSpace(match[1]))
	if _, err := os.Stat(wrapped); err != nil {
		if os.IsNotExist(err) {
			return source, nil
		}
		return "", fmt.Errorf("stat wrapped executable %s: %w", wrapped, err)
	}
	return wrapped, nil
}

func removeExistingDestination(destination string) error {
	info, err := os.Lstat(destination)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", destination, err)
	}

	if info.IsDir() {
		return fmt.Errorf("destination %s exists as a directory", destination)
	}
	if err := os.Chmod(destination, info.Mode().Perm()|0o200); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("make %s writable: %w", destination, err)
	}
	if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing %s: %w", destination, err)
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
