# `install.yaml`

`install.yaml` describes how `launchdctl install` should write and register a launchd job.

It is intentionally limited to launchd-facing concerns:

- agent label and plist location
- program arguments
- working directory
- stdout and stderr paths
- environment variables
- launchd service policy such as `RunAtLoad`, `KeepAlive`, `ThrottleInterval`, `Umask`, and `StartCalendarInterval`
- `launchctl` actions such as bootout, bootstrap, and kickstart

It does **not** copy files into place. That belongs in `bundle.yaml`.

## Mental Model

Use `install.yaml` when you already know:

- which executable should run
- with which arguments
- where logs should go
- whether the job is a supervised service or a scheduled one-shot task

## Schema

```yaml
agent:
  label: ai.openclaw.gateway
  domain: user

program:
  argv:
    - ~/Library/Application Support/openclaw-local/bin/node
    - ~/Library/Application Support/openclaw-local/app/dist/entry.js
    - gateway
    - --port
    - "18789"

logging:
  stdout_path: ~/Library/Application Support/openclaw-local/state/logs/gateway.log
  stderr_path: ~/Library/Application Support/openclaw-local/state/logs/gateway.err.log

environment:
  OPENCLAW_STATE_DIR: ~/Library/Application Support/openclaw-local/state

env_from_host:
  - HOME
  - TMPDIR

service:
  run_at_load: true
  keep_alive: true
  throttle_interval: 1
  umask: 63

install:
  validate_plist: true
  bootout_existing: true
  bootstrap: true
  kickstart_after_bootstrap: false
```

## Sections

### `agent`

Fields:

- `label`: required launchd label
- `domain`: optional, `user` or `system`, defaults to `user`
- `plist_path`: optional explicit plist destination

Current default plist paths:

- `domain: user`
  - `~/Library/LaunchAgents/<label>.plist`
- `domain: system`
  - `/Library/LaunchDaemons/<label>.plist`

### `program`

Fields:

- `argv`: required program arguments array
- `working_directory`: optional working directory

Current path behavior:

- values in `argv` are expanded only when they look like paths
- current path-like prefixes are:
  - `~`
  - `./`
  - `../`
  - `/`

That means:

- `~/bin/app` is expanded
- `./dist/entry.js` is resolved
- `--port` stays as `--port`
- `"18789"` stays as `"18789"`

### `logging`

Fields:

- `stdout_path`: required
- `stderr_path`: required

These become the launchd-managed stdout and stderr files.

### `environment`

Optional explicit environment map.

Current behavior matches the implementation exactly:

- keys must be non-empty
- values are expanded only when they look like paths
- non-path strings are kept as-is

### `env_from_host`

Optional list of environment variable names to copy from the current process environment.

Current behavior:

- only the named keys are considered
- missing keys are ignored
- copied values are written into the plist `EnvironmentVariables` dictionary

### `service`

Launchd-facing runtime settings.

Fields:

- `run_at_load`
- `keep_alive`
- `throttle_interval`
- `umask`
- `start_calendar_interval`

Use patterns:

- supervised service
  - `run_at_load: true`
  - `keep_alive: true`
- scheduled one-shot job
  - `run_at_load: false`
  - `keep_alive: false`
  - `start_calendar_interval: [...]`

### `install`

Controls `launchdctl install` behavior after the plist is generated.

Fields:

- `validate_plist`
- `bootout_existing`
- `bootstrap`
- `kickstart_after_bootstrap`

Important: these fields do **not** become plist keys. They control follow-up `launchctl` operations.

## Domain Behavior

Current implementation supports:

- `user`
- `system`

The implementation is primarily oriented around LaunchAgents on macOS, and the examples in this repo use `domain: user`.

## `launchctl` Semantics

Current behavior:

- `validate_plist: true`
  - runs `plutil -lint`
- `bootstrap: true`
  - runs `launchctl bootstrap`
- `bootout_existing: true`
  - runs `launchctl bootout` first
- `kickstart_after_bootstrap: true`
  - runs `launchctl kickstart -k` after bootstrap

## Not Yet Modeled

The current schema does not yet expose every launchd plist feature. For example:

- `UserName`
- `GroupName`
- `Disabled`
- `ProcessType`
- `MachServices`
- `Sockets`
- `SoftResourceLimits`
- `HardResourceLimits`
- `KeepAlive` sub-dictionaries

Those are intentionally out of scope for the current implementation and should not be assumed available.
