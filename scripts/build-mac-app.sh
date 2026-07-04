#!/usr/bin/env bash
# Build a double-clickable ProdCal.app bundle around cmd/prodcal-app.
#
# Personal single-user use: no code signing, no notarization. The bundle is
# written to ./ProdCal.app (gitignored) and bakes two things into Info.plist's
# LSEnvironment so a Finder launch (cwd=/, minimal PATH) still works:
#   - JDBBS_TYPESETTING_DIR → this repo's typesetting/ (fonts + conversion scripts)
#   - PATH including Homebrew, so pandoc/typst/python3 are found
#
# Usage: ./scripts/build-mac-app.sh      (from anywhere; repo root is derived)
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
app="$repo_root/ProdCal.app"
typesetting_dir="$repo_root/typesetting"

if [ "$(uname -s)" != "Darwin" ]; then
  echo "build-mac-app: macOS only (cmd/prodcal-app is //go:build darwin)" >&2
  exit 1
fi

echo "-- building prodcal-app --"
(cd "$repo_root" && go build -o /tmp/prodcal-app-build ./cmd/prodcal-app)

echo "-- assembling $app --"
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
mv /tmp/prodcal-app-build "$app/Contents/MacOS/prodcal-app"

cat > "$app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>ProdCal</string>
	<key>CFBundleDisplayName</key>
	<string>ProdCal</string>
	<key>CFBundleIdentifier</key>
	<string>com.jdbb.prodcal</string>
	<key>CFBundleVersion</key>
	<string>1.0</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleExecutable</key>
	<string>prodcal-app</string>
	<key>LSMinimumSystemVersion</key>
	<string>12.0</string>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>LSEnvironment</key>
	<dict>
		<key>JDBBS_TYPESETTING_DIR</key>
		<string>${typesetting_dir}</string>
		<key>PATH</key>
		<string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
	</dict>
</dict>
</plist>
PLIST

# Finder caches LSEnvironment per bundle path; touching the bundle re-registers it.
touch "$app"

echo "done: $app"
echo "  launch: open \"$app\"   (or double-click in Finder)"
echo "  data dir: ~/Library/Application Support/ProdCal"
echo "  typesetting: $typesetting_dir (baked into Info.plist LSEnvironment)"
