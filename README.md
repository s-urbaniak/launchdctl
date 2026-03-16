# launchdctl

`launchdctl` is a small macOS-focused tool for:

- bundling an app into an isolated on-disk root
- installing and managing a LaunchAgent from an explicit manifest

The repo uses two YAML manifests:

- `bundle.yaml` describes filesystem layout and file copies
- `install.yaml` describes launchd registration and plist generation

The top-level docs live under `docs/`:

- [bundle.yaml guide](docs/bundle.md)
- [install.yaml guide](docs/install.md)
- [plist mapping](docs/plist-mapping.md)
- [schema reference](docs/schema-reference.md)
- [examples guide](docs/examples.md)

## Commands

```bash
launchdctl bundle --file examples/backup-agent/bundle.yaml
launchdctl install --file examples/backup-agent/install.yaml
```

## Examples

- `examples/openclaw`: mirrors OpenClaw's launchd gateway install model
- `examples/backup-agent`: mirrors this repo's bundled backup layout

## License

Apache-2.0. See [LICENSE](LICENSE).
