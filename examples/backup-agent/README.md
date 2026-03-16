# backup-agent Example

This example is a hypothetical backup-style sample project.

It is intentionally not a built-in or otherwise special application inside `launchdctl`. The name is just a concrete stand-in for a scheduled backup job with a bundled binary, config, logs, state, and vendor directories.

```bash
launchdctl bundle --file examples/backup-agent/bundle.yaml
launchdctl install --file examples/backup-agent/install.yaml
```

Provide local inputs under `examples/backup-agent/inputs/`:

- `restic`
- `config.yaml`
- `signal-cli/`
- `jre/`

See also:

- [bundle.yaml guide](../../docs/bundle.md)
- [install.yaml guide](../../docs/install.md)
- [examples guide](../../docs/examples.md)
