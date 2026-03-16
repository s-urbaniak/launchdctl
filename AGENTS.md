# Repository Instructions

## Git Commit Messages

- Use `package: imperative summary`.
- Keep the subject concise and specific.
- Add a body only when extra context is useful.

Examples:

- `bundle: copy directories recursively`
- `launchd: write user agent plists`

## Git Workflow

- Commit completed changes before finishing a task.
- Push committed changes when an `origin` remote exists.
- If no remote exists yet, mention that push is blocked by missing remote.

## Manifest Constraints

- `bundle.yaml` owns filesystem layout and file copies only.
- `install.yaml` owns launchd registration only.
- Do not add app-specific defaults to the core tool.
