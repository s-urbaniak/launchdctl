# Schema Reference

This reference documents the manifest contract as it exists today in `internal/spec/spec.go`.

## `bundle.yaml`

### `BundleManifest`

| Field | Type | Required | Default | Notes |
| --- | --- | --- | --- | --- |
| `bundle` | object | yes | none | Contains `root` |
| `directories` | array | no | empty | Directory creation entries |
| `files` | array | no | empty | File or directory copy entries |

### `bundle.root`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `root` | string | yes | none | must be non-empty |

### `directories[]`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `path` | string | yes | none | must be non-empty |
| `mode` | string | no | `"0755"` | must parse as octal |

### `files[]`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `source` | string | yes | none | must be non-empty |
| `destination` | string | yes | none | must be non-empty |
| `mode` | string | no | `"0644"` | must parse as octal for file copies |
| `copy_directory` | bool | no | `false` | when `true`, source is copied recursively |

## `install.yaml`

### `InstallManifest`

| Field | Type | Required | Default | Notes |
| --- | --- | --- | --- | --- |
| `agent` | object | yes | none | Launchd identity and plist location |
| `program` | object | yes | none | Executable and argv |
| `logging` | object | yes | none | stdout/stderr files |
| `environment` | map | no | empty | explicit environment |
| `env_from_host` | array | no | empty | host env allowlist |
| `service` | object | no | zero values | launchd runtime behavior |
| `install` | object | no | zero values | install-time command behavior |

### `agent`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `label` | string | yes | none | must be non-empty |
| `domain` | string | no | `user` | must be `user` or `system` |
| `plist_path` | string | no | derived from domain and label | expanded if provided |

### `program`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `argv` | array of string | yes | none | must contain at least one entry |
| `working_directory` | string | no | empty | path-expanded |

### `logging`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `stdout_path` | string | yes | none | must be non-empty |
| `stderr_path` | string | yes | none | must be non-empty |

### `environment`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| arbitrary keys | string map | no | empty | keys must be non-empty |

Implementation note:

- values are path-expanded only when they look like paths

### `env_from_host`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| list entries | array of string | no | empty | entries must be non-empty |

### `service`

| Field | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `run_at_load` | bool | no | `false` | none |
| `keep_alive` | bool | no | `false` | none |
| `throttle_interval` | int | no | `0` | must be `>= 0` |
| `umask` | int | no | `0` | must be `>= 0` |
| `start_calendar_interval` | array | no | empty | each field validated individually |

### `start_calendar_interval[]`

| Field | Type | Required | Validation |
| --- | --- | --- | --- |
| `minute` | int | no | `0-59` |
| `hour` | int | no | `0-23` |
| `weekday` | int | no | `0-7` |
| `day` | int | no | `1-31` |
| `month` | int | no | `1-12` |

### `install`

| Field | Type | Required | Default | Notes |
| --- | --- | --- | --- | --- |
| `validate_plist` | bool | no | `false` | runs `plutil -lint` |
| `bootout_existing` | bool | no | `false` | runs `launchctl bootout` first |
| `bootstrap` | bool | no | `false` | runs `launchctl bootstrap` |
| `kickstart_after_bootstrap` | bool | no | `false` | runs `launchctl kickstart -k` |
