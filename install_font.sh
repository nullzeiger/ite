#!/bin/bash

// Copyright 2025 Ivan Guerreschi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

set -e

FONT_DIR="$HOME/.local/share/fonts"
FONT_URLS=(
    "https://go.googlesource.com/image/+/refs/tags/v0.34.0/font/gofont/ttfs/Go-Mono-Bold.ttf?format=TEXT"
    "https://go.googlesource.com/image/+/refs/tags/v0.34.0/font/gofont/ttfs/Go-Mono.ttf?format=TEXT"
)
FONT_NAMES=(
    "Go-Mono-Bold.ttf"
    "Go-Mono.ttf"
)

echo "Installing Go Mono fonts..."

# Create fonts directory if it doesn't exist
if [ ! -d "$FONT_DIR" ]; then
    echo "Creating directory: $FONT_DIR"
    mkdir -p "$FONT_DIR"
fi

# Download and install fonts
for i in "${!FONT_URLS[@]}"; do
    url="${FONT_URLS[$i]}"
    filename="${FONT_NAMES[$i]}"
    filepath="$FONT_DIR/$filename"

    echo "Downloading $filename..."

    # Download base64 encoded font and decode it
    if command -v curl &> /dev/null; then
        curl -s "$url" | base64 -d > "$filepath"
    elif command -v wget &> /dev/null; then
        wget -q -O - "$url" | base64 -d > "$filepath"
    else
        echo "Error: Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [ -f "$filepath" ]; then
        echo "\u2713 Installed: $filename"
    else
        echo "\u2717 Failed to install: $filename"
        exit 1
    fi
done

# Update font cache
echo "Updating font cache..."
if command -v fc-cache &> /dev/null; then
    fc-cache -f "$FONT_DIR"
    echo "\u2713 Font cache updated"
else
    echo "Warning: fc-cache not found. You may need to restart your system for fonts to take effect."
fi

echo ""
echo "Installation complete!"
echo "Fonts installed in: $FONT_DIR"
echo ""
echo "Verify installation with: fc-list | grep 'Go Mono'"