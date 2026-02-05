# GitHub Actions Workflow Guide

## 🚀 Automated Release Workflow

The provider includes a GitHub Actions workflow that automatically builds and releases the provider for multiple platforms.

## 📋 How It Works

### Trigger Conditions

The workflow triggers on push to `main` branch when:
1. Commit message starts with `BUILD:`
2. Changes are not just documentation files

### Version Detection

The workflow extracts version from the commit message:

```bash
# Commit message format:
BUILD: v1.2.3 - Description of changes

# Or without 'v':
BUILD: 1.2.3 - Description of changes

# Or default to 0.1.0:
BUILD: Initial release
```

## 🎯 Usage Examples

### Example 1: Release Version 0.1.0

```bash
git add .
git commit -m "BUILD: v0.1.0 - Initial release with firewall rules, NAT, Kea DHCP, and WireGuard support"
git push origin main
```

**Result:** Creates release `v0.1.0` with binaries for all platforms

### Example 2: Release Version 0.2.0

```bash
git add .
git commit -m "BUILD: v0.2.0 - Added multi-category support for firewall rules"
git push origin main
```

**Result:** Creates release `v0.2.0`

### Example 3: Bug Fix Release

```bash
git add .
git commit -m "BUILD: v0.1.1 - Fixed WireGuard DNS field type issue"
git push origin main
```

**Result:** Creates release `v0.1.1`

### Example 4: Skip Build (Regular Commit)

```bash
git add .
git commit -m "docs: Update README with examples"
git push origin main
```

**Result:** No build or release (commit doesn't start with `BUILD:`)

## 📦 What Gets Built

The workflow builds binaries for:

| Platform | Architecture | File Name |
|----------|-------------|-----------|
| Linux | amd64 | `terraform-provider-opnsense_VERSION_linux_amd64.tar.gz` |
| Linux | arm64 | `terraform-provider-opnsense_VERSION_linux_arm64.tar.gz` |
| macOS | amd64 | `terraform-provider-opnsense_VERSION_darwin_amd64.tar.gz` |
| macOS | arm64 (M1/M2) | `terraform-provider-opnsense_VERSION_darwin_arm64.tar.gz` |
| Windows | amd64 | `terraform-provider-opnsense_VERSION_windows_amd64.zip` |

Plus: `terraform-provider-opnsense_VERSION_SHA256SUMS` for verification

## 🔐 Security

All binaries are:
- ✅ Built with `CGO_ENABLED=0` (statically linked)
- ✅ Stripped and optimized (`-ldflags="-s -w"`)
- ✅ Include SHA256 checksums
- ✅ Signed by GitHub Actions

## 📥 Download from Release

Users can download from the GitHub Releases page:

```bash
# Download specific version
wget https://github.com/rgcosta7/terraform-provider-opnsense-26/releases/download/v0.1.0/terraform-provider-opnsense_0.1.0_linux_amd64.tar.gz

# Verify checksum
wget https://github.com/rgcosta7/terraform-provider-opnsense-26/releases/download/v0.1.0/terraform-provider-opnsense_0.1.0_SHA256SUMS
sha256sum -c terraform-provider-opnsense_0.1.0_SHA256SUMS

# Extract
tar -xzf terraform-provider-opnsense_0.1.0_linux_amd64.tar.gz

# Install
mkdir -p ~/.terraform.d/plugins/localhost/local/opnsense/0.1.0/linux_amd64
mv terraform-provider-opnsense ~/.terraform.d/plugins/localhost/local/opnsense/0.1.0/linux_amd64/
chmod +x ~/.terraform.d/plugins/localhost/local/opnsense/0.1.0/linux_amd64/terraform-provider-opnsense
```

## 🔄 Release Process

### Step 1: Make Changes

```bash
# Make your code changes
vim internal/provider/resource_firewall_rule.go

# Test locally
go test ./...
```

### Step 2: Commit with BUILD Message

```bash
git add .
git commit -m "BUILD: v0.1.0 - Initial release

Features:
- Firewall rules with multi-category support
- Firewall aliases
- Firewall categories
- Destination NAT (port forwarding)
- Kea DHCP subnets and reservations
- WireGuard servers and peers (with DNS, MTU, Gateway)

Fixes:
- WireGuard DNS field now accepts comma-separated string
- All Kea endpoints corrected to snake_case
- Categories support in firewall rules
"
```

### Step 3: Push to Main

```bash
git push origin main
```

### Step 4: Monitor Workflow

1. Go to GitHub Actions tab
2. Watch the "Build and Release Provider" workflow
3. Once complete, check Releases page

## 📊 Workflow Status

Check workflow status:
- ✅ Green checkmark = Build successful
- ❌ Red X = Build failed
- 🟡 Yellow dot = Build in progress

## 🛠️ Local Testing Before Release

Test the build locally before pushing:

```bash
# Test build for your platform
CGO_ENABLED=0 go build -o terraform-provider-opnsense -ldflags="-s -w"

# Test installation
mkdir -p ~/.terraform.d/plugins/localhost/local/opnsense/0.1.0/linux_amd64
cp terraform-provider-opnsense ~/.terraform.d/plugins/localhost/local/opnsense/0.1.0/linux_amd64/
chmod +x ~/.terraform.d/plugins/localhost/local/opnsense/0.1.0/linux_amd64/terraform-provider-opnsense

# Test with Terraform
cd examples/firewall
terraform init
terraform plan
```

## 📝 Best Practices

### Semantic Versioning

Follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR** (1.0.0): Breaking changes
  - Example: `BUILD: v1.0.0 - Changed resource schema (breaking)`
  
- **MINOR** (0.2.0): New features, backwards compatible
  - Example: `BUILD: v0.2.0 - Added Source NAT support`
  
- **PATCH** (0.1.1): Bug fixes, backwards compatible
  - Example: `BUILD: v0.1.1 - Fixed category field validation`

### Commit Message Format

```
BUILD: vX.Y.Z - Brief summary

Detailed description of changes:
- Feature 1
- Feature 2
- Bug fix 1

Breaking changes (if any):
- Change 1

Migration guide (if needed):
- Step 1
- Step 2
```

## 🚨 Troubleshooting

### Build Fails

1. Check GitHub Actions logs
2. Verify Go syntax: `go build ./...`
3. Run tests: `go test ./...`
4. Check imports and dependencies

### Release Not Created

1. Verify commit message starts with `BUILD:`
2. Check workflow ran successfully
3. Verify GITHUB_TOKEN permissions
4. Check for duplicate version tags

### Wrong Version Number

1. Version is extracted from commit message
2. Format: `BUILD: v0.1.0` or `BUILD: 0.1.0`
3. Default is `0.1.0` if not specified

## 💡 Tips

- Use draft releases for testing: Modify workflow `draft: true`
- Tag pre-releases: `BUILD: v0.1.0-beta1`
- Skip CI for docs: Regular commits without `BUILD:`
- Test locally first: Build and test before pushing

## 📞 Support

Issues with the workflow? Open an issue on GitHub with:
- Commit message used
- Workflow run link
- Error logs
