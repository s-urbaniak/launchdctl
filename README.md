# launchdctl

`launchdctl` is a small macOS-focused tool for:

- bundling an app into an isolated on-disk root
- installing and managing a LaunchAgent from an explicit manifest

## Commands

```bash
launchdctl bundle --file examples/backup-agent/bundle.yaml
launchdctl install --file examples/backup-agent/install.yaml
```

## Examples

- `examples/openclaw`: mirrors OpenClaw's launchd gateway install model
- `examples/backup-agent`: mirrors this repo's bundled backup layout
