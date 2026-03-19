# OpenClaw Example

This example mirrors the launchd behavior behind `openclaw onboard --install-daemon` by modeling the underlying gateway install path in one `Launchdfile`.

```bash
go run ./cmd/launchdctl apply --file examples/openclaw/Launchdfile
```

Provide local inputs under `examples/openclaw/inputs/`:

- `node`
- `openclaw-dist/`
- `config.json`

See also:

- [Launchdfile guide](../../docs/launchdfile.md)
- [instruction reference](../../docs/instruction-reference.md)
- [examples guide](../../docs/examples.md)
