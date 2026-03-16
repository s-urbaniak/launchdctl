# `bundle.yaml`

`bundle.yaml` describes how `launchdctl bundle` should materialize an application root on disk.

It is intentionally limited to filesystem concerns:

- where the bundle root lives
- which directories should exist
- which files should be copied
- which source directories should be copied recursively

It does **not** contain launchd configuration. That belongs in `install.yaml`.

## Mental Model

Use `bundle.yaml` when you want to create a predictable, self-contained directory tree such as:

- `bin/`
- `config/`
- `logs/`
- `state/`
- `vendor/`

The manifest says what should exist inside that tree. It does not care what the app does at runtime.

## Schema

```yaml
bundle:
  root: ~/Library/Application Support/example-app

directories:
  - path: bin
    mode: "0755"

files:
  - source: ./inputs/app
    destination: bin/app
    mode: "0755"

  - source: ./inputs/vendor-tree
    destination: vendor/vendor-tree
    copy_directory: true
```

## Fields

### `bundle.root`

Required.

The absolute or home-relative destination root to build.

Examples:

- `~/Library/Application Support/openclaw-local`
- `/opt/example-app`

Behavior:

- `~` is expanded to the current user home
- relative paths are resolved from the manifest directory

### `directories`

Optional.

Each entry creates a directory relative to `bundle.root`.

Fields:

- `path`: required relative path under the bundle root
- `mode`: optional octal string, defaults to `0755`

Example:

```yaml
directories:
  - path: state/logs
    mode: "0755"
```

### `files`

Optional.

Each entry copies a source file or source directory into the bundle root.

Fields:

- `source`: required source path
- `destination`: required destination path relative to `bundle.root`
- `mode`: optional octal string for file copies, defaults to `0644`
- `copy_directory`: optional boolean, when `true` copies the source tree recursively

Examples:

Single file:

```yaml
files:
  - source: ./inputs/restic
    destination: bin/restic
    mode: "0755"
```

Recursive directory copy:

```yaml
files:
  - source: ./inputs/signal-cli
    destination: vendor/signal-cli
    copy_directory: true
```

## Path Resolution Rules

The current implementation resolves paths exactly as follows:

- `bundle.root`
  - `~` expands to the current user home
  - relative paths resolve from the manifest directory
- `files[].source`
  - `~` expands to the current user home
  - relative paths resolve from the manifest directory
- `directories[].path` and `files[].destination`
  - always resolve relative to `bundle.root`

## Validation Rules

`launchdctl` currently validates:

- `bundle.root` must be present
- each `directories[].path` must be present
- each `files[].source` must be present
- each `files[].destination` must be present
- `mode` values must parse as octal strings
- destination paths must not be duplicated within one manifest

## Common Mistakes

- Putting launchd settings in `bundle.yaml`
  - launchd registration belongs in `install.yaml`
- Using an absolute destination path in `files[].destination`
  - destination is interpreted relative to `bundle.root`
- Forgetting `copy_directory: true` for directory trees
  - without it, the source is treated like a single file
