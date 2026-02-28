#!/bin/bash

# Package clipboard-island as a macOS .app bundle

set -e

APP_NAME="Clipboard Island"
BUNDLE_NAME="Clipboard Island.app"
BINARY_NAME="clipboard-island"
BUNDLE_ID="com.example.clipboardisland"
VERSION="0.1.0"

echo "Packaging ${APP_NAME}..."

# Create .app bundle structure
mkdir -p "bin/${BUNDLE_NAME}/Contents/MacOS"
mkdir -p "bin/${BUNDLE_NAME}/Contents/Resources"

# Copy binary
cp "bin/${BINARY_NAME}" "bin/${BUNDLE_NAME}/Contents/MacOS/"

# Create Info.plist
cat > "bin/${BUNDLE_NAME}/Contents/Info.plist" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleName</key>
    <string>Clipboard Island</string>
    <key>CFBundleExecutable</key>
    <string>clipboard-island</string>
    <key>CFBundleIdentifier</key>
    <string>com.example.clipboardisland</string>
    <key>CFBundleVersion</key>
    <string>0.1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>0.1.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
    <key>LSBackgroundOnly</key>
    <false/>
</dict>
</plist>
PLIST

# Copy icon
if [ -f "build/darwin/icons.icns" ]; then
    cp "build/darwin/icons.icns" "bin/${BUNDLE_NAME}/Contents/Resources/icons.icns"
fi

# Also copy icon PNG for reference
if [ -f "icon_1024.png" ]; then
    cp "icon_1024.png" "bin/${BUNDLE_NAME}/Contents/Resources/icon.png"
fi

# Make binary executable
chmod +x "bin/${BUNDLE_NAME}/Contents/MacOS/${BINARY_NAME}"

echo "✅ Created bin/${BUNDLE_NAME}"
echo ""
echo "To run without terminal:"
echo "  1. Double-click bin/${BUNDLE_NAME} in Finder"
echo "  2. Or run: open 'bin/${BUNDLE_NAME}'"
echo ""
echo "To install:"
echo "  cp -r 'bin/${BUNDLE_NAME}' /Applications/"
