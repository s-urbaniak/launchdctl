# Examples

This repo ships two worked examples that represent two different launchd usage patterns.

## `examples/openclaw`

This example models OpenClaw as a supervised gateway service.

Why it is a service:

- it should start when loaded
- it should stay running
- launchd should supervise and restart it
- it should not wait for a scheduled clock time

That is why the example uses:

- `run_at_load: true`
- `keep_alive: true`
- `throttle_interval: 1`
- `umask: 63`
- no `start_calendar_interval`

It also models the OpenClaw service environment explicitly:

- `OPENCLAW_STATE_DIR`
- `OPENCLAW_CONFIG_PATH`
- `OPENCLAW_GATEWAY_PORT`
- `OPENCLAW_LAUNCHD_LABEL`
- `OPENCLAW_SERVICE_MARKER`
- `OPENCLAW_SERVICE_KIND`

Expected local inputs:

- `examples/openclaw/inputs/node`
- `examples/openclaw/inputs/openclaw-dist/`
- `examples/openclaw/inputs/config.json`

## `examples/backup-agent`

This example models a hypothetical backup-style application called `backup-agent` as a scheduled one-shot job.

Why it is a scheduled job:

- it performs a backup run and exits
- launchd is used for periodic scheduling rather than long-running supervision

That is why the example uses:

- `run_at_load: false`
- `keep_alive: false`
- `start_calendar_interval`

Its bundle also demonstrates recursive vendor-tree copies:

- `signal-cli/`
- `jre/`

Expected local inputs:

- `examples/backup-agent/inputs/restic`
- `examples/backup-agent/inputs/config.yaml`
- `examples/backup-agent/inputs/signal-cli/`
- `examples/backup-agent/inputs/jre/`

## Which One Should You Copy?

Start from:

- `openclaw` if your process should stay alive under launchd supervision
- `backup-agent` if your process should run on a schedule and exit

Important: the `backup-agent` example is just a sample project shape. It is not part of `launchdctl` itself and should be read as a template for "backup-like scheduled job" rather than a known application.

Use the example manifests as templates, then change:

- bundle root
- executable path
- config path
- logs
- environment
- scheduling or supervision settings
