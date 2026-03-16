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
