package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"launchdctl/internal/bundle"
	"launchdctl/internal/launchd"
	"launchdctl/internal/spec"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return errors.New("missing subcommand")
	}

	switch args[0] {
	case "bundle":
		return runBundle(args[1:])
	case "install":
		return runInstall(args[1:])
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `launchdctl manages app bundles and launchd installation.

Usage:
  launchdctl <subcommand> [flags]

Subcommands:
  bundle    Materialize a bundle root from bundle.yaml
  install   Write and register a LaunchAgent from install.yaml
`)
}

func runBundle(args []string) error {
	fs := flag.NewFlagSet("bundle", flag.ContinueOnError)
	path := fs.String("file", "", "Path to bundle.yaml")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return errors.New("--file is required")
	}

	manifest, err := spec.LoadBundleFile(*path)
	if err != nil {
		return err
	}
	return bundle.Apply(manifest)
}

func runInstall(args []string) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	path := fs.String("file", "", "Path to install.yaml")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return errors.New("--file is required")
	}

	manifest, err := spec.LoadInstallFile(*path)
	if err != nil {
		return err
	}
	return launchd.Apply(context.Background(), manifest, nil)
}
