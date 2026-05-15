#!/bin/sh
REPO="neutron420/StackAudit"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
BIN_DIR="$HOME/.stack/bin"
SHELL_RC=""

case "$ARCH" in
	x86_64) ARCH="amd64" ;;
	aarch64|arm64) ARCH="arm64" ;;
esac

URL=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "browser_download_url" | grep "${OS}" | grep "${ARCH}" | grep "tar.gz" | head -n 1 | cut -d '"' -f 4)

if [ -z "$URL" ]; then
	echo "Error: could not resolve a Linux/macOS release asset for ${OS}/${ARCH}." >&2
	exit 1
fi

echo "Downloading stack from $URL..."
mkdir -p "$BIN_DIR"
tmpdir=$(mktemp -d)
curl -L "$URL" | tar -xz -C "$tmpdir"

if [ -f "$tmpdir/stack" ]; then
	install -m 755 "$tmpdir/stack" "$BIN_DIR/stack"
elif [ -f "$tmpdir/bin/stack" ]; then
	install -m 755 "$tmpdir/bin/stack" "$BIN_DIR/stack"
else
	echo "Error: could not find stack binary in the release archive." >&2
	exit 1
fi

rm -rf "$tmpdir"

case "$SHELL" in
	*/zsh) SHELL_RC="$HOME/.zshrc" ;;
	*/bash) SHELL_RC="$HOME/.bashrc" ;;
	*) SHELL_RC="$HOME/.profile" ;;
esac

if ! grep -q "\.stack/bin" "$SHELL_RC" 2>/dev/null; then
	echo 'export PATH="$HOME/.stack/bin:$PATH"' >> "$SHELL_RC"
fi

export PATH="$HOME/.stack/bin:$PATH"

echo "$BIN_DIR added to PATH. Restart your terminal or run: source $SHELL_RC"
echo "stack installed successfully! Type 'stack' to begin."
