#!/bin/bash

# Ensure the script is run with root privileges
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this installer as root (e.g., sudo ./install.sh)"
  exit 1
fi

echo "⚙️ Step 1: Stopping service..."
systemctl stop nvr.service

echo "⚙️ Step 2: Updating..."
cp nvr_service /opt/nvr/bin/
cp nvr_worker /opt/nvr/bin/
cp -r web /opt/nvr/bin/web

echo "🔄 Step 3: Starting systemd background service..."
systemctl start nvr.service

echo "✅ NVR update complete!"
echo "Use 'systemctl status nvr.service' to check its status."