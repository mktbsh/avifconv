#!/bin/bash
set -e

BINARY_NAME="avifconv"
SOURCE_PATH="./bin/$BINARY_NAME"
INSTALL_DIR="/usr/local/avifconv"
BIN_DIR="/usr/local/bin"
SYMLINK_PATH="$BIN_DIR/$BINARY_NAME"

if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run with sudo privileges"
    echo "Please run: sudo $0"
    exit 1
fi

if [ ! -f "$SOURCE_PATH" ]; then
    echo "Error: Binary '$SOURCE_PATH' not found"
    echo "Please run build.sh first to create the binary"
    exit 1
fi

echo "Installing avifconv..."

mkdir -p "$INSTALL_DIR"
mkdir -p "$BIN_DIR"

echo "Copying binary to $INSTALL_DIR/$BINARY_NAME"
cp "$SOURCE_PATH" "$INSTALL_DIR/$BINARY_NAME"
chmod 755 "$INSTALL_DIR/$BINARY_NAME"

echo "Creating symlink at $SYMLINK_PATH"
if [ -L "$SYMLINK_PATH" ]; then
    echo "Updating existing symlink"
    ln -sf "$INSTALL_DIR/$BINARY_NAME" "$SYMLINK_PATH"
elif [ -e "$SYMLINK_PATH" ]; then
    echo "Warning: $SYMLINK_PATH exists but is not a symlink"
    echo "Creating backup at ${SYMLINK_PATH}.bak"
    mv "$SYMLINK_PATH" "${SYMLINK_PATH}.bak"
    ln -sf "$INSTALL_DIR/$BINARY_NAME" "$SYMLINK_PATH"
else
    ln -sf "$INSTALL_DIR/$BINARY_NAME" "$SYMLINK_PATH"
fi

echo "Verifying installation..."
if [ -x "$INSTALL_DIR/$BINARY_NAME" ] && [ -L "$SYMLINK_PATH" ]; then
    echo "Successfully installed avifconv to $INSTALL_DIR"
    echo "Symlink created at $SYMLINK_PATH"

    echo
    echo "You can now run 'avifconv' from any directory"
else
    echo "Installation verification failed"
    exit 1
fi

cat >"$INSTALL_DIR/uninstall.sh" <<EOF
#!/bin/bash
set -e

if [ "\$EUID" -ne 0 ]; then
  echo "Error: This script must be run with sudo privileges"
  echo "Please run: sudo \$0"
  exit 1
fi

echo "Uninstalling avifconv..."
if [ -L "$SYMLINK_PATH" ]; then
  echo "Removing symlink $SYMLINK_PATH"
  rm "$SYMLINK_PATH"
fi

echo "Removing installation directory $INSTALL_DIR"
rm -rf "$INSTALL_DIR"

echo "avifconv has been uninstalled"
EOF

chmod 755 "$INSTALL_DIR/uninstall.sh"

echo
echo "To uninstall, run: sudo $INSTALL_DIR/uninstall.sh"
