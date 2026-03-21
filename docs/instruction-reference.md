# Instruction Reference

This reference documents the `Launchdfile` contract as implemented in `internal/spec/spec.go`.

## Filesystem Instructions

### `RUN ["arg0","arg1",...]`

Optional and repeatable.

Executes a preparation command before bundle directories, copies, plist writing,
or `launchd` install actions.

Rules:

- must use JSON-array form
- must appear before all other directives
- runs with the `Launchdfile` directory as the working directory
- path-like argv entries are expanded using the same rules as `CMD`

### `ROOT <path>`

Required.

Sets the bundle root on disk.

### `MKDIR <path> [MODE <octal>]`

Optional and repeatable.

Creates a directory relative to `ROOT`.

Defaults:

- `MODE` defaults to `0755`

### `COPY <source> <destination> [MODE <octal>]`

Optional and repeatable.

Copies one file into the bundle root.

Defaults:

- `MODE` defaults to `0644`

### `COPYDIR <source> <destination>`

Optional and repeatable.

Copies a directory tree recursively into the bundle root.

## Launchd Identity Instructions

### `LABEL <launchd-label>`

Required.

Sets the plist `Label`.

### `DOMAIN user|system`

Optional.

Defaults to `user`.

### `PLIST <path>`

Optional.

Overrides the default plist destination.

Current defaults:

- `user`: `~/Library/LaunchAgents/<label>.plist`
- `system`: `/Library/LaunchDaemons/<label>.plist`

## Program Instructions

### `CMD ["arg0","arg1",...]`

Required.

Must use JSON-array form.

### `WORKDIR <path>`

Optional.

Sets `WorkingDirectory`.

### `STDOUT <path>`

Required.

Sets `StandardOutPath`.

### `STDERR <path>`

Required.

Sets `StandardErrorPath`.

## Environment Instructions

### `ENV <key>=<value>`

Optional and repeatable.

Adds an explicit environment variable.

### `ENVFROM <key>`

Optional and repeatable.

Copies a host environment variable into `EnvironmentVariables` when present.

## Service Instructions

### `RUNATLOAD true|false`

Optional.

### `KEEPALIVE true|false`

Optional.

### `THROTTLE <seconds>`

Optional.

### `UMASK <int>`

Optional.

### `SCHEDULE [minute=..] [hour=..] [weekday=..] [day=..] [month=..]`

Optional and repeatable.

Each instruction adds one `StartCalendarInterval` entry.

### `PROCESSTYPE standard|background|adaptive|interactive`

Optional.

Sets plist `ProcessType`.

### `DISABLED true|false`

Optional.

Sets plist `Disabled`.

## Install Instruction

### `INSTALL [validate=<bool>] [bootout_existing=<bool>] [bootstrap=<bool>] [kickstart=<bool>]`

Optional.

Controls follow-up operations after plist generation.

Supported options:

- `validate`
  - runs `plutil -lint`
- `bootout_existing`
  - runs `launchctl bootout` before bootstrap
- `bootstrap`
  - runs `launchctl bootstrap`
- `kickstart`
  - runs `launchctl kickstart -k` after bootstrap
