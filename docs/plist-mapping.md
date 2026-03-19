# `Launchdfile` To Plist Mapping

This document maps `Launchdfile` instructions to the plist entries emitted by `internal/launchd/launchd.go`.

| `Launchdfile` instruction | Plist key | Notes |
| --- | --- | --- |
| `LABEL` | `Label` | required |
| `CMD` | `ProgramArguments` | JSON array preserved as argv |
| `WORKDIR` | `WorkingDirectory` | omitted when empty |
| `STDOUT` | `StandardOutPath` | required |
| `STDERR` | `StandardErrorPath` | required |
| `ENV`, `ENVFROM` | `EnvironmentVariables` | omitted when empty |
| `RUNATLOAD` | `RunAtLoad` | emitted as a boolean |
| `KEEPALIVE` | `KeepAlive` | emitted as a boolean |
| `THROTTLE` | `ThrottleInterval` | omitted when `0` |
| `UMASK` | `Umask` | omitted when `0` |
| `SCHEDULE` | `StartCalendarInterval` | one entry becomes one dictionary, many become an array |
| `PROCESSTYPE` | `ProcessType` | omitted when empty |
| `DISABLED` | `Disabled` | emitted only when specified |

Install-time instructions do not become plist keys:

- `DOMAIN`
- `PLIST`
- `INSTALL validate=...`
- `INSTALL bootout_existing=...`
- `INSTALL bootstrap=...`
- `INSTALL kickstart=...`

Operational effects:

- `DOMAIN`
  - selects `gui/<uid>` or `system`
- `PLIST`
  - overrides the plist destination on disk
- `INSTALL validate=true`
  - runs `plutil -lint`
- `INSTALL bootout_existing=true`
  - runs `launchctl bootout` before bootstrap
- `INSTALL bootstrap=true`
  - runs `launchctl bootstrap`
- `INSTALL kickstart=true`
  - runs `launchctl kickstart -k`
