# `Launchdfile`

`Launchdfile` is the single manifest format for `launchdctl`.

It describes:

- optional preparation commands that produce managed inputs
- the managed application root
- directories and files that should exist inside that root
- the `launchd` command and plist settings
- optional install-time actions after plist generation

It does not model mutable user data.

## Mental Model

Use `Launchdfile` when you want one explicit recipe for:

- preparing build or staging outputs
- creating a predictable app directory tree
- copying the files your service depends on
- writing a plist
- registering that plist with `launchd`

`launchdctl apply --file Launchdfile` executes that recipe from top to bottom.
It first runs any `RUN` preparation steps, then applies the managed filesystem
layout, then writes and installs the plist.

## Example

```text
RUN ["go","build","-o","./dist/example-app","./cmd/example-app"]

ROOT "~/Library/Application Support/example-app"

MKDIR bin MODE 0755
MKDIR logs MODE 0755

COPY "./dist/example-app" "bin/example-app" MODE 0755

LABEL com.example.app
CMD ["~/Library/Application Support/example-app/bin/example-app","serve"]
WORKDIR "~/Library/Application Support/example-app"

STDOUT "~/Library/Application Support/example-app/logs/stdout.log"
STDERR "~/Library/Application Support/example-app/logs/stderr.log"

RUNATLOAD true
KEEPALIVE true

INSTALL validate=true bootout_existing=true bootstrap=true kickstart=false
```

## Path Rules

Current path behavior matches the implementation:

- `RUN` and `CMD`
  - path-like argv entries are expanded when they begin with `~`, `./`, `../`, or `/`
- `ROOT`
  - `~` expands to the current user home
  - relative paths resolve from the `Launchdfile` directory
- `COPY` and `COPYDIR` sources
  - `~` expands to the current user home
  - relative paths resolve from the `Launchdfile` directory
- `MKDIR`, `COPY`, and `COPYDIR` destinations
  - always resolve relative to `ROOT`
- `WORKDIR`, `STDOUT`, `STDERR`, `PLIST`
  - `~` expands to the current user home
  - relative paths resolve from the `Launchdfile` directory

## Validation

`launchdctl` currently requires:

- `ROOT`
- `LABEL`
- `CMD`
- `STDOUT`
- `STDERR`

It also validates:

- `RUN` must appear before all other directives
- destination paths must not be duplicated
- `MODE` values must parse as octal
- `DOMAIN` must be `user` or `system`
- `PROCESSTYPE` must be one of the supported values
- `SCHEDULE` fields must be in range
- `INSTALL kickstart=true` requires `bootstrap=true`
