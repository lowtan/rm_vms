#!/bin/bash

# Ensure the script is run with root privileges
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this installer as root (e.g., sudo ./install.sh)"
  exit 1
fi

echo "📦 Step 1: Installing FFmpeg runtime dependencies..."
apt-get update
apt-get install -y libavformat58 libavcodec58 libavutil56 libswscale5 ca-certificates

echo "📂 Step 2: Creating system directories..."
mkdir -p /opt/nvr/bin
mkdir -p /opt/nvr/recordings

echo "⚙️ Step 3: Installing binaries..."
# Copy binaries to the secure /opt directory
cp nvr_service /opt/nvr/bin/
cp nvr_worker /opt/nvr/bin/
chmod +x /opt/nvr/bin/nvr_service
chmod +x /opt/nvr/bin/nvr_worker

echo "🔄 Step 4: Setting up systemd background service..."
cp nvr.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable nvr.service
systemctl restart nvr.service

echo "✅ NVR installation complete! The service is now running in the background."
echo "Use 'systemctl status nvr.service' to check its status."
