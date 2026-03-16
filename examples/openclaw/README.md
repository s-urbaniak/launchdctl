# OpenClaw Example

This example mirrors the launchd behavior behind `openclaw onboard --install-daemon` by modeling the underlying gateway install path:

```bash
launchdctl bundle --file examples/openclaw/bundle.yaml
launchdctl install --file examples/openclaw/install.yaml
```

Provide local inputs under `examples/openclaw/inputs/`:

- `node`
- `openclaw-dist/`
- `config.json`

See also:

- [bundle.yaml guide](../../docs/bundle.md)
- [install.yaml guide](../../docs/install.md)
- [examples guide](../../docs/examples.md)
