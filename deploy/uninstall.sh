#!/usr/bin/env bash
#
# Removes the GEIS LaunchAgents. Pass --purge to also delete the install dir,
# cache, frames, and logs.
#
set -euo pipefail

AGENTS_DIR="$HOME/Library/LaunchAgents"
INSTALL_DIR="$HOME/Library/Application Support/GEIS"
LOG_DIR="$HOME/Library/Logs/GEIS"
UID_NUM="$(id -u)"

for label in at.newell.geis.changewallpaper at.newell.geis.fetch; do
  launchctl bootout "gui/$UID_NUM/$label" 2>/dev/null || true
  rm -f "$AGENTS_DIR/$label.plist"
  echo "removed agent $label"
done

if [ "${1:-}" = "--purge" ]; then
  rm -rf "$INSTALL_DIR" "$LOG_DIR"
  echo "purged $INSTALL_DIR and $LOG_DIR"
else
  echo "left $INSTALL_DIR in place (run with --purge to delete it)"
fi

echo "Done. Your desktop wallpaper is unchanged; set a new one in System Settings."
