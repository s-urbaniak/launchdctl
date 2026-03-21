# backup-agent Example

This example is a hypothetical backup-style sample project.

It is intentionally not a built-in or otherwise special application inside `launchdctl`. The name is just a concrete stand-in for a scheduled backup job with a bundled binary, config, logs, state, and vendor directories.

```bash
go run ./cmd/launchdctl apply --file examples/backup-agent/Launchdfile
```

Any leading `RUN` steps in the manifest execute first from the example
directory, then the managed files are copied into the target app root and the
LaunchAgent is refreshed.

Provide local inputs under `examples/backup-agent/inputs/`:

- `restic`
- `config.yaml`
- `signal-cli/`
- `jre/`

See also:

- [Launchdfile guide](../../docs/launchdfile.md)
- [instruction reference](../../docs/instruction-reference.md)
- [examples guide](../../docs/examples.md)
