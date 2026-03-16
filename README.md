# launchdctl

`launchdctl` is a small macOS-focused tool for turning "a few files plus a process command" into a repeatable local app bundle and a real `launchd` job.

If you have ever ended up with some mix of:

- ad-hoc shell scripts
- handwritten plist XML
- one-off `launchctl bootstrap` commands
- app-specific install logic embedded deep inside an otherwise unrelated binary

then `launchdctl` is meant to give that work a cleaner home.

It does two things:

- bundle an app into an isolated on-disk root
- install and manage a LaunchAgent from an explicit manifest

The repo uses two YAML manifests:

- `bundle.yaml` describes filesystem layout and file copies
- `install.yaml` describes launchd registration and plist generation

The goal is not to replace `launchd`. The goal is to make working with `launchd` feel more declarative, repeatable, and reusable across personal macOS projects.

## Why This Exists

macOS already has a strong answer for background services: Apple recommends `launchd` for both daemons and per-user agents. The problem is not that `launchd` is missing. The problem is the gap between:

- a normal app or CLI project
- a real, maintainable LaunchAgent install

That gap often gets filled in awkward ways:

- plist XML embedded in shell scripts
- installer logic mixed into the runtime binary
- per-project copy-and-paste wrappers
- package-manager-specific service definitions that are useful only in one distribution path

`launchdctl` tries to be the small missing layer in that space:

- generic enough to reuse across unrelated projects
- explicit enough that the YAML mirrors what actually lands on disk
- narrow enough that it does not turn into a packaging framework

## Prior Art

There is good prior art in the ecosystem, but each option sits at a different layer:

- Apple `launchd` documentation and plist conventions are the source of truth for how jobs work:
  https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html
- Homebrew has `service do` blocks, which are excellent when your distribution model is a Homebrew formula:
  https://docs.brew.sh/Formula-Cookbook#service-blocks
- Some projects carry their own reusable code for launchd management, such as Keybase's internal Go package:
  https://github.com/keybase/client/tree/master/go/launchd

What felt missing for this repo was a project-local tool that is:

- file-manifest driven
- centered on app bundling plus LaunchAgent installation
- independent of Homebrew formulas
- not tied to one application's runtime logic

That is the niche `launchdctl` is trying to fill.

## Where `launchdctl` Fits

`launchdctl` is useful when:

- you want a self-contained bundle under `~/Library/Application Support/...`
- you want the install logic outside your app runtime
- you want to reuse the same install pattern across very different personal projects
- you want examples that can model both supervised services and scheduled one-shot jobs

It is especially handy for the "small but real" class of macOS projects:

- local gateways
- backup jobs
- helper daemons
- personal automation services
- node/python/go CLIs that need a stable launchd home

This is mostly a good fit for personal projects and project-local deployment workflows.

It is a weaker fit when a packaging ecosystem already owns installation and service wiring for you. For example:

- Homebrew formulas already have `service do`
- MacPorts ports usually have their own packaging and service conventions
- Nix-based packaging often wants the package manager to own paths, closures, and service setup declaratively

So the best way to think about `launchdctl` is:

- strong fit for project-local macOS bundling and LaunchAgent installation
- weaker fit as a replacement for package-manager-native service definitions

It is not trying to be:

- a full package manager
- a replacement for `brew services`
- a replacement for MacPorts or Nix packaging
- a universal launchd schema generator
- a GUI app installer

## CLI First, Embeddable Second

`launchdctl` is primarily a standalone CLI.

That is the main supported interface today:

- `bundle.yaml` for bundle creation
- `install.yaml` for launchd installation

At the same time, the project is intentionally structured so the same logic can be a useful foundation for Go projects that want to embed self-contained bundling or LaunchAgent installation behavior.

That matters for projects that want:

- install logic outside the runtime path, but still in Go
- a project-specific installer command built on top of the same primitives
- vendored or copied implementation rather than handwritten plist/install code

The important caveat is that the CLI is the primary contract today. The repo is not yet presented as a polished public Go SDK with a stable supported API boundary.

## Design Principles

- `bundle.yaml` owns filesystem layout only.
- `install.yaml` owns launchd registration only.
- The tool stays generic and does not absorb app-specific runtime logic.
- The manifest contract should match the implementation closely enough that users do not have to read Go code to understand it.

The top-level docs live under `docs/`:

- [bundle.yaml guide](docs/bundle.md)
- [install.yaml guide](docs/install.md)
- [plist mapping](docs/plist-mapping.md)
- [schema reference](docs/schema-reference.md)
- [examples guide](docs/examples.md)

## Commands

```bash
launchdctl bundle --file examples/backup-agent/bundle.yaml
launchdctl install --file examples/backup-agent/install.yaml
```

## Examples

- `examples/openclaw`: mirrors OpenClaw's launchd gateway install model
- `examples/backup-agent`: a hypothetical scheduled backup-style app used as a sample manifest set

These examples are intentionally different from each other:

- `openclaw` is a supervised long-running service
- `backup-agent` is a hypothetical scheduled one-shot backup job

That contrast is part of the point. `launchdctl` should be useful for both patterns without baking either one into the tool itself.

## License

Apache-2.0. See [LICENSE](LICENSE).
