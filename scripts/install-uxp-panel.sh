#!/usr/bin/env bash
# Install UXP panel for PremierPro MCP transcript bridge (macOS).
set -euo pipefail

PLUGIN_ID="com.premierpro.mcp.uxpbridge"
SOURCE_DIR="$(cd "$(dirname "$0")/../uxp-panel" && pwd)"
TARGET_DIR="${HOME}/Library/Application Support/Adobe/UXP/Plugins/External/${PLUGIN_ID}"

echo "Installing UXP transcript bridge panel..."
echo "Source: ${SOURCE_DIR}"
echo "Target: ${TARGET_DIR}"

if [[ ! -f "${SOURCE_DIR}/manifest.json" ]]; then
  echo "ERROR: uxp-panel source not found" >&2
  exit 1
fi

rm -rf "${TARGET_DIR}"
mkdir -p "${TARGET_DIR}"
cp -R "${SOURCE_DIR}/." "${TARGET_DIR}/"

echo ""
echo "UXP panel installed."
echo "1. Enable UXP developer mode in Premiere Pro Preferences"
echo "2. Restart Premiere Pro"
echo "3. Open Window > UXP Plugins > PremierPro MCP UXP Bridge"
echo "4. Confirm Connected to ws://127.0.0.1:9802"
