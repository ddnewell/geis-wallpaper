#!/usr/bin/env bash
#
# Installs the GEIS wallpaper renderer (and fetcher, if built) as per-user
# launchd LaunchAgents. Safe to re-run: it rebuilds, refreshes the install dir,
# preserves an existing config.json, and reloads the agents.
#
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DIR="$HOME/Library/Application Support/GEIS"
LOG_DIR="$HOME/Library/Logs/GEIS"
AGENTS_DIR="$HOME/Library/LaunchAgents"
RENDER_LABEL="at.newell.geis.changewallpaper"
FETCH_LABEL="at.newell.geis.fetch"
UID_NUM="$(id -u)"

echo "==> Building binaries"
cd "$REPO_DIR"
go build -o "$REPO_DIR/geis" ./cmd/geis
HAVE_FETCH=0
if [ -d "$REPO_DIR/cmd/geis-fetch" ]; then
  go build -o "$REPO_DIR/geis-fetch" ./cmd/geis-fetch && HAVE_FETCH=1
fi

echo "==> Installing to $INSTALL_DIR"
mkdir -p "$INSTALL_DIR" "$LOG_DIR" "$INSTALL_DIR/cache" "$INSTALL_DIR/frames" "$AGENTS_DIR"
mkdir -p "$INSTALL_DIR/img" "$INSTALL_DIR/ico" "$INSTALL_DIR/json"
cp "$REPO_DIR/geis" "$INSTALL_DIR/geis"
[ "$HAVE_FETCH" = "1" ] && cp "$REPO_DIR/geis-fetch" "$INSTALL_DIR/geis-fetch"
cp "$REPO_DIR/img/NaturalEarth_Mac13Retina.png" "$INSTALL_DIR/img/"
cp -R "$REPO_DIR/ico/wx" "$INSTALL_DIR/ico/"
cp "$REPO_DIR/json/clocks.json" "$INSTALL_DIR/json/"
if [ ! -f "$INSTALL_DIR/config.json" ]; then
  cp "$REPO_DIR/json/config.example.json" "$INSTALL_DIR/config.json"
  echo "    wrote default config.json"
else
  echo "    kept existing config.json"
fi

install_agent() {
  local label="$1" template="$2"
  sed -e "s|{{INSTALL_DIR}}|$INSTALL_DIR|g" -e "s|{{HOME}}|$HOME|g" \
    "$template" > "$AGENTS_DIR/$label.plist"
  launchctl bootout "gui/$UID_NUM/$label" 2>/dev/null || true
  launchctl bootstrap "gui/$UID_NUM" "$AGENTS_DIR/$label.plist"
  launchctl enable "gui/$UID_NUM/$label"
  echo "    loaded $label"
}

echo "==> Installing LaunchAgents"
install_agent "$RENDER_LABEL" "$REPO_DIR/deploy/geis.plist.template"
if [ "$HAVE_FETCH" = "1" ]; then
  install_agent "$FETCH_LABEL" "$REPO_DIR/deploy/geis-fetch.plist.template"
fi

echo "==> Registering the wallpaper path (may prompt once for Automation access — click OK)"
"$INSTALL_DIR/geis" --register --config "$INSTALL_DIR/config.json" || true

echo "==> Done. Renderer runs every 1 min on every Space; fetcher every 15 min (if installed)."
echo "    Uninstall with: deploy/uninstall.sh"
