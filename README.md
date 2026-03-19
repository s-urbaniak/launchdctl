# launchdctl

`launchdctl` is a small macOS-focused tool for turning "a few files plus a process command" into a repeatable local app bundle and a real `launchd` job.

The project is now centered on a single `Launchdfile`. One file describes:

- where the app root lives
- which directories and files should exist inside it
- which command `launchd` should run
- how the plist should be configured
- which install actions should happen after the plist is written

The goal is not to replace `launchd`. The goal is to make working with `launchd` feel more declarative, repeatable, and reusable across personal macOS projects.

## Why This Exists

macOS already has a strong answer for background services: Apple recommends `launchd` for both daemons and per-user agents. The problem is not that `launchd` is missing. The problem is the gap between:

- a normal app or CLI project
- a real, maintainable LaunchAgent install

That gap often gets filled in awkward ways:

- ad-hoc shell scripts
- handwritten plist XML
- one-off `launchctl bootstrap` commands
- installer logic embedded deep inside an otherwise unrelated binary

`launchdctl` tries to be the small missing layer in that space:

- generic enough to reuse across unrelated projects
- explicit enough that the manifest mirrors what lands on disk and in the plist
- narrow enough that it does not turn into a packaging framework

## Where It Fits

`launchdctl` is useful when:

- you want a self-contained bundle under `~/Library/Application Support/...`
- you want the install logic outside your app runtime
- you want to reuse the same install pattern across very different personal projects
- you want examples that model both supervised services and scheduled one-shot jobs

It is especially handy for the "small but real" class of macOS projects:

- local gateways
- backup jobs
- helper daemons
- personal automation services
- node/python/go CLIs that need a stable launchd home

It is not trying to be:

- a full package manager
- a replacement for `brew services`
- a GUI app installer
- a general build pipeline

## Launchdfile

Example:

```text
ROOT "~/Library/Application Support/example-app"

MKDIR bin MODE 0755
MKDIR config MODE 0755
MKDIR logs MODE 0755

COPY "./dist/example-app" "bin/example-app" MODE 0755
COPY "./config.example.json" "config/config.json" MODE 0644

LABEL com.example.app
DOMAIN user

CMD ["~/Library/Application Support/example-app/bin/example-app","serve"]
WORKDIR "~/Library/Application Support/example-app"

STDOUT "~/Library/Application Support/example-app/logs/stdout.log"
STDERR "~/Library/Application Support/example-app/logs/stderr.log"

ENV APP_ROOT=~/Library/Application Support/example-app
ENVFROM HOME
ENVFROM TMPDIR

RUNATLOAD true
KEEPALIVE true
THROTTLE 1
UMASK 63

INSTALL validate=true bootout_existing=true bootstrap=true kickstart=false
```

## Commands

```bash
go run ./cmd/launchdctl apply --file examples/backup-agent/Launchdfile
```

`apply` performs the full flow in order:

1. creates the bundle directories
2. copies managed files into the app root
3. writes the plist
4. performs optional install actions such as validation, bootout, bootstrap, and kickstart

## Design Principles

- one file owns the managed app layout and `launchd` job definition
- the manifest stays generic and does not absorb app-specific runtime logic
- mutable user data remains outside the manifest contract
- the documented contract should match the implementation closely

## Docs

- [Launchdfile guide](docs/launchdfile.md)
- [instruction reference](docs/instruction-reference.md)
- [plist mapping](docs/plist-mapping.md)
- [examples guide](docs/examples.md)

## Examples

- `examples/openclaw`: mirrors OpenClaw's gateway install model
- `examples/backup-agent`: a hypothetical scheduled backup-style app

## License

Apache-2.0. See [LICENSE](LICENSE).
