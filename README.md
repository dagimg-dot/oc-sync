# oc-sync

Offline OpenCode session sync between machines.

Export sessions as JSON to a shared directory (Syncthing, SSHFS, USB), import
on another machine. Only sessions you choose are transferred — per-session
JSON files with INSERT OR IGNORE import for safe merging.

## Install

```bash
go install github.com/dagimg-dot/oc-sync/cmd/oc-sync@latest
```

Or build from source:

```bash
git clone https://github.com/dagimg-dot/oc-sync
cd oc-sync
make build
```

## Config

Config file: `~/.config/oc-sync/config.yaml`

```yaml
db_path: ~/.local/share/opencode/opencode.db
sync_dir: ~/Sync/oc-sync
hostname: my-laptop
mappings:
  - remote_hostname: desktop
    remote_project_id: abc123
    remote_worktree: /home/user/project
    local_worktree: /home/user/project
    local_project_id: xyz789
```

All fields are optional. Defaults work with Syncthing's default folder.

## Usage

```bash
# List sessions in the local database
oc-sync list

# Export specific sessions to the sync directory
oc-sync export ses_abc123 ses_def456

# Import sessions from peer machines
oc-sync import

# Bidirectional sync — export new + import foreign
oc-sync sync

# Show config and pending state
oc-sync status

# Print version
oc-sync version
```

## How it works

Each machine has its own OpenCode SQLite database. oc-sync reads from
the local DB (read-only) and writes per-session JSON files to a shared
directory. Peer machines read those JSON files and insert the sessions
into their own local DB.

- Export: `local DB → JSON file in <sync_dir>/<hostname>/<session-id>.json`
- Import: `peer JSON file → local DB` (INSERT OR IGNORE, idempotent)
- Project mapping: configure `mappings` in config to translate project IDs
  between machines with different worktree paths
