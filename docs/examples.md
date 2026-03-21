# Examples

This repo ships two worked `Launchdfile` examples that represent two different `launchd` usage patterns.

## `openclaw`

`examples/openclaw/Launchdfile` models a supervised long-running gateway:

- the app root is `~/Library/Application Support/openclaw-local`
- `node`, the app distribution, and `config.json` are copied into place
- the service is configured with `RUNATLOAD true` and `KEEPALIVE true`
- host proxy-related environment variables are forwarded through `ENVFROM`

Provide local inputs under `examples/openclaw/inputs/`:

- `examples/openclaw/inputs/node`
- `examples/openclaw/inputs/openclaw-dist/`
- `examples/openclaw/inputs/config.json`

Run it with:

```bash
go run ./cmd/launchdctl apply --file examples/openclaw/Launchdfile
```

## `backup-agent`

`examples/backup-agent/Launchdfile` models a scheduled one-shot job:

- preparation commands can build or stage inputs before the bundle copy phase
- the app root is `~/Library/Application Support/backup-agent`
- the main binary, config, and vendor trees are copied into place
- the job is driven by `SCHEDULE hour=2 minute=0`
- `KEEPALIVE` remains false so the process is not treated as a supervised daemon

Provide local inputs under `examples/backup-agent/inputs/`:

- `examples/backup-agent/inputs/restic`
- `examples/backup-agent/inputs/config.yaml`
- `examples/backup-agent/inputs/signal-cli/`
- `examples/backup-agent/inputs/jre/`

Run it with:

```bash
go run ./cmd/launchdctl apply --file examples/backup-agent/Launchdfile
```

## Choosing A Shape

Use:

- `openclaw` if your process should stay alive under `launchd` supervision
- `backup-agent` if your process should run on a schedule and then exit

Important: the `backup-agent` example is just a sample project shape. It is not part of `launchdctl` itself and should be read as a template for "backup-like scheduled job" rather than a known application.
