# Exposing oc-sync Folder to Syncthing on Android

## Problem

Termux lives in `/data/data/com.termux/` — Android's private app storage. Other apps (Syncthing) cannot access it. The `~/Sync/oc-sync` folder was invisible to the Syncthing Android app.

## Solution

Move the sync data to Android's shared storage (`/storage/emulated/0/`) and symlink back so Termux's oc-sync still finds it.

### Steps

```bash
# 1. Move the sync directory to shared storage
mv ~/Sync/oc-sync ~/storage/shared/oc-sync

# 2. Symlink so Termux sees it at the original path
ln -sf ~/storage/shared/oc-sync ~/Sync/oc-sync
```

### Final layout

| Path | Accessible by |
|---|---|
| `~/storage/shared/oc-sync` | Syncthing Android app |
| `/storage/emulated/0/oc-sync` | Same location (Android filesystem path) |
| `~/Sync/oc-sync` → symlink to above | oc-sync CLI via Termux |
| `~/Sync/oc-sync/pc-desktop/` | PC exports |
| `~/Sync/oc-sync/phone-android/` | Phone exports |

### Syncthing app setup

1. Install Syncthing from F-Droid or Google Play
2. Add folder → pick `/storage/emulated/0/oc-sync`
3. Share with the PC device ID
4. Accept on PC web UI (http://127.0.0.1:8384)

Both subdirectories sync automatically.
