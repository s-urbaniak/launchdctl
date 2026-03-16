# backup-agent Example

This example mirrors the current backup-agent layout with generic bundle rules:

```bash
launchdctl bundle --file examples/backup-agent/bundle.yaml
launchdctl install --file examples/backup-agent/install.yaml
```

Provide local inputs under `examples/backup-agent/inputs/`:

- `restic`
- `config.yaml`
- `signal-cli/`
- `jre/`
