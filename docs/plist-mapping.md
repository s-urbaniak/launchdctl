# `install.yaml` To Plist Mapping

This document maps `install.yaml` fields to the plist entries emitted by `internal/launchd/launchd.go`.

## Direct Mappings

| `install.yaml` field | Plist key | Notes |
| --- | --- | --- |
| `agent.label` | `Label` | Required |
| `program.argv` | `ProgramArguments` | Required |
| `program.working_directory` | `WorkingDirectory` | Only emitted when non-empty |
| `logging.stdout_path` | `StandardOutPath` | Required |
| `logging.stderr_path` | `StandardErrorPath` | Required |
| `environment` + `env_from_host` | `EnvironmentVariables` | Only emitted when non-empty |
| `service.run_at_load` | `RunAtLoad` | Always emitted |
| `service.keep_alive` | `KeepAlive` | Always emitted |
| `service.throttle_interval` | `ThrottleInterval` | Only emitted when `> 0` |
| `service.umask` | `Umask` | Only emitted when `> 0` |
| `service.start_calendar_interval` | `StartCalendarInterval` | Only emitted when non-empty |

## `StartCalendarInterval`

Each entry in `service.start_calendar_interval` maps to a launchd calendar dictionary with some combination of:

- `Minute`
- `Hour`
- `Weekday`
- `Day`
- `Month`

Current emission behavior:

- one interval entry becomes a single dictionary
- multiple interval entries become an array of dictionaries

## Not Mapped To Plist

The following `install.yaml` fields drive install behavior only and do **not** become plist keys:

- `agent.domain`
- `agent.plist_path`
- `install.validate_plist`
- `install.bootout_existing`
- `install.bootstrap`
- `install.kickstart_after_bootstrap`

## Operational Mapping

These fields currently map to command behavior:

| `install.yaml` field | Operational effect |
| --- | --- |
| `agent.domain` | selects `gui/<uid>` or `system` launchd domain |
| `agent.plist_path` | determines where the plist is written |
| `install.validate_plist` | runs `plutil -lint` |
| `install.bootout_existing` | runs `launchctl bootout` before bootstrap |
| `install.bootstrap` | runs `launchctl bootstrap` |
| `install.kickstart_after_bootstrap` | runs `launchctl kickstart -k` |

## Caveats

- `EnvironmentVariables` is the merged result of explicit `environment` plus any present keys from `env_from_host`.
- `RunAtLoad` and `KeepAlive` are always emitted, even when `false`.
- The current implementation does not model advanced launchd keys such as `ProcessType`, `MachServices`, or `KeepAlive` sub-keys.
