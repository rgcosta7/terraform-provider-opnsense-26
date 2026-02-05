# Quick Release Guide

## 🚀 How to Create a Release

### Quick Steps

```bash
# 1. Make your changes
git add .

# 2. Commit with BUILD: prefix and version
git commit -m "BUILD: v0.1.0 - Description of changes"

# 3. Push to main
git push origin main

# 4. Done! GitHub Actions builds and releases automatically
```

## 📋 Commit Message Format

```
BUILD: vX.Y.Z - Brief summary

Optional detailed description:
- Feature 1
- Feature 2
- Bug fix 1
```

## 🎯 Examples

### Initial Release
```bash
git commit -m "BUILD: v0.1.0 - Initial release with complete OPNsense 26.1 support"
```

### Feature Release
```bash
git commit -m "BUILD: v0.2.0 - Added Source NAT support"
```

### Bug Fix
```bash
git commit -m "BUILD: v0.1.1 - Fixed WireGuard DNS field type"
```

### Beta Release
```bash
git commit -m "BUILD: v0.2.0-beta1 - Testing new features"
```

## ✅ What Happens

1. GitHub Actions detects `BUILD:` prefix
2. Extracts version from commit message
3. Builds for 5 platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
4. Creates SHA256 checksums
5. Publishes GitHub Release
6. Uploads all binaries

## 🔍 Check Status

Go to: https://github.com/YOUR_USERNAME/terraform-provider-opnsense-26/actions

## 📦 Find Release

Go to: https://github.com/YOUR_USERNAME/terraform-provider-opnsense-26/releases

## ⏭️ Skip Build

For regular commits (no release):

```bash
git commit -m "docs: Updated README"  # No BUILD: prefix
git push origin main  # No release created
```

## 🎉 That's It!

No manual builds, no manual releases - just commit and push!
