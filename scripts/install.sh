#!/bin/bash
REPO="neutron420/StackAudit"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" == "x86_64" ]; then ARCH="x86_64"; fi
if [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then ARCH="arm64"; fi

URL=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "browser_download_url" | grep "${OS}_${ARCH}.tar.gz" | cut -d '"' -f 4)

echo "Downloading stack from $URL..."
curl -L $URL | tar -xz
sudo mv stack /usr/local/bin/

echo "stack installed successfully! Type 'stack' to begin."
