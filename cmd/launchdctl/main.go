package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"launchdctl/internal/bundle"
	"launchdctl/internal/launchd"
	"launchdctl/internal/prepare"
	"launchdctl/internal/spec"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "apply":
		if err := runApply(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	path := fs.String("file", "Launchdfile", "path to the Launchdfile")
	if err := fs.Parse(args); err != nil {
		return err
	}

	manifest, err := spec.LoadLaunchdfile(*path)
	if err != nil {
		return err
	}
	if err := prepare.Apply(context.Background(), manifest, nil); err != nil {
		return err
	}
	if err := bundle.Apply(manifest); err != nil {
		return err
	}
	return launchd.Apply(context.Background(), manifest, nil)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n  %s apply --file Launchdfile\n", os.Args[0])
}
