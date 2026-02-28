#!/bin/bash

# Regenerate the app icon from icon_1024.png
# Run this after modifying icon_1024.png

set -e

echo "Updating app icon..."

if [ ! -f "icon_1024.png" ]; then
    echo "Error: icon_1024.png not found!"
    echo "Generate it first with: go run generate_icon.go"
    exit 1
fi

# Create iconset
mkdir -p icon.iconset

# Generate all required sizes
sips -z 16 16 icon_1024.png --out icon.iconset/icon_16x16.png
sips -z 32 32 icon_1024.png --out icon.iconset/icon_16x16@2x.png
sips -z 32 32 icon_1024.png --out icon.iconset/icon_32x32.png
sips -z 64 64 icon_1024.png --out icon.iconset/icon_32x32@2x.png
sips -z 128 128 icon_1024.png --out icon.iconset/icon_128x128.png
sips -z 256 256 icon_1024.png --out icon.iconset/icon_128x128@2x.png
sips -z 256 256 icon_1024.png --out icon.iconset/icon_256x256.png
sips -z 512 512 icon_1024.png --out icon.iconset/icon_256x256@2x.png
sips -z 512 512 icon_1024.png --out icon.iconset/icon_512x512.png
cp icon_1024.png icon.iconset/icon_512x512@2x.png

# Convert to icns
iconutil -c icns icon.iconset -o build/darwin/icons.icns

# Cleanup
rm -rf icon.iconset

echo "✅ Icon updated in build/darwin/icons.icns"
echo "Run 'wails3 build && ./package.sh' to apply to the app"
