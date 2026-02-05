# Publishing to Terraform Registry

This guide covers everything you need to publish your provider to the official Terraform Registry.

## 🔐 Step 1: Generate GPG Key

The Terraform Registry requires GPG-signed releases for security.

### Generate a New GPG Key

```bash
# Generate key (use your GitHub email)
gpg --full-generate-key

# Follow prompts:
# - Key type: (1) RSA and RSA (default)
# - Key size: 4096
# - Expiration: 0 (does not expire) or set expiration
# - Real name: Your Name
# - Email: your-github-email@example.com
# - Comment: Terraform Provider Signing Key
```

### Export Your Keys

```bash
# List your keys
gpg --list-secret-keys --keyid-format=long

# Output will look like:
# sec   rsa4096/ABCD1234EFGH5678 2024-02-05 [SC]
#       Full fingerprint here
# uid                 [ultimate] Your Name <email@example.com>

# Export private key (for GitHub Secrets)
gpg --armor --export-secret-keys ABCD1234EFGH5678 > private-key.asc

# Export public key (for Terraform Registry)
gpg --armor --export ABCD1234EFGH5678 > public-key.asc

# Get key fingerprint (for registry verification)
gpg --fingerprint ABCD1234EFGH5678
```

### Upload Public Key to Registry

The public key needs to be uploaded to a keyserver that Terraform Registry can access:

```bash
# Upload to Ubuntu keyserver (recommended)
gpg --keyserver keyserver.ubuntu.com --send-keys ABCD1234EFGH5678

# Or upload to keys.openpgp.org
gpg --keyserver keys.openpgp.org --send-keys ABCD1234EFGH5678

# Verify it's uploaded
gpg --keyserver keyserver.ubuntu.com --recv-keys ABCD1234EFGH5678
```

## 🔑 Step 2: Add Secrets to GitHub

Add these secrets to your GitHub repository (Settings → Secrets and variables → Actions):

### 1. GPG_PRIVATE_KEY

```bash
# Copy the entire content of private-key.asc
cat private-key.asc
```

Copy the output (including `-----BEGIN PGP PRIVATE KEY BLOCK-----` and `-----END PGP PRIVATE KEY BLOCK-----`) and add as `GPG_PRIVATE_KEY`

### 2. GPG_PASSPHRASE

Add the passphrase you used when creating the GPG key as `GPG_PASSPHRASE`

**⚠️ Important:** Keep `private-key.asc` secure and delete it after adding to GitHub!

```bash
# After adding to GitHub, securely delete
shred -vfz -n 10 private-key.asc
```

## 📋 Step 3: Prepare Repository

### Required Files

Your repository needs these files (already included):

```
terraform-provider-opnsense/
├── .github/
│   └── workflows/
│       └── release.yml          # ✅ Updated with GPG signing
├── docs/
│   ├── index.md                 # Provider documentation
│   ├── resources/               # Resource docs
│   │   ├── firewall_rule.md
│   │   ├── firewall_alias.md
│   │   ├── firewall_category.md
│   │   ├── nat_destination.md
│   │   ├── kea_subnet.md
│   │   ├── kea_reservation.md
│   │   ├── wireguard_server.md
│   │   └── wireguard_peer.md
│   └── data-sources/
│       └── firewall_rule.md
├── examples/                    # Example configurations
├── internal/
│   └── provider/
├── go.mod
├── go.sum
├── main.go
└── README.md
```

### Create Provider Documentation Index

Create `docs/index.md`:

```markdown
---
page_title: "Provider: OPNsense"
description: |-
  The OPNsense provider allows Terraform to manage OPNsense 26.1 firewall configurations.
---

# OPNsense Provider

The OPNsense provider allows you to manage OPNsense 26.1 firewall configurations using Terraform.

## Features

- **Firewall Rules:** Complete firewall rule management with multi-category support
- **Firewall Aliases:** Network, host, and port aliases
- **Firewall Categories:** Organize rules with categories
- **NAT:** Destination NAT (port forwarding)
- **DHCP:** Kea DHCP subnets and reservations
- **VPN:** WireGuard servers and peers

## Example Usage

```terraform
terraform {
  required_providers {
    opnsense = {
      source  = "rgcosta7/opnsense"
      version = "~> 0.1"
    }
  }
}

provider "opnsense" {
  host       = "https://192.168.1.1"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
  insecure   = true  # Only for self-signed certificates
}

# Create a firewall rule
resource "opnsense_firewall_rule" "allow_https" {
  enabled     = true
  description = "Allow HTTPS"
  
  action      = "pass"
  interface   = "lan"
  protocol    = "tcp"
  
  source_net       = "192.168.1.0/24"
  destination_net  = "any"
  destination_port = "443"
}
```

## Authentication

The provider requires API credentials from your OPNsense firewall.

### Creating API Credentials

1. Log into OPNsense web interface
2. Go to **System → Access → Users**
3. Select a user (or create a new one)
4. Scroll to **API Keys** section
5. Click **+** to generate a new key
6. Download the credentials file (only chance!)
7. Use the key and secret in provider configuration

## Schema

- `host` (String, Required) - OPNsense host URL (e.g., `https://192.168.1.1`)
- `api_key` (String, Required, Sensitive) - API key
- `api_secret` (String, Required, Sensitive) - API secret
- `insecure` (Boolean, Optional) - Skip TLS verification (default: false)
- `timeout_seconds` (Number, Optional) - HTTP timeout in seconds (default: 30)

Environment variables can be used:
- `OPNSENSE_HOST`
- `OPNSENSE_API_KEY`
- `OPNSENSE_API_SECRET`
```

## 🚀 Step 4: Create Signed Release

### Push with BUILD: prefix

```bash
git add .
git commit -m "BUILD: v0.1.0 - Initial release

Complete OPNsense 26.1 provider with:
- Firewall rules (with multi-category support)
- Firewall aliases and categories
- Destination NAT
- Kea DHCP
- WireGuard VPN
"
git push origin main
```

### Verify Release Files

After GitHub Actions completes, verify your release has these files:

```
terraform-provider-opnsense_0.1.0_linux_amd64.tar.gz
terraform-provider-opnsense_0.1.0_linux_arm64.tar.gz
terraform-provider-opnsense_0.1.0_darwin_amd64.tar.gz
terraform-provider-opnsense_0.1.0_darwin_arm64.tar.gz
terraform-provider-opnsense_0.1.0_windows_amd64.zip
terraform-provider-opnsense_0.1.0_SHA256SUMS
terraform-provider-opnsense_0.1.0_SHA256SUMS.sig    # ✅ CRITICAL for registry!
```

## 📝 Step 5: Register on Terraform Registry

### 1. Sign In

Go to https://registry.terraform.io and sign in with your GitHub account.

### 2. Publish Provider

1. Click **Publish** → **Provider**
2. Select your GitHub repository: `rgcosta7/terraform-provider-opnsense-26`
3. Click **Publish Provider**

### 3. Add GPG Public Key

1. Go to your provider settings
2. Add your GPG public key:
   - Key ID: `ABCD1234EFGH5678` (from earlier)
   - ASCII Armor Public Key: Content of `public-key.asc`
3. Or provide keyserver URL: `keyserver.ubuntu.com`

### 4. Verify Provider

The registry will:
- ✅ Verify your GPG signature
- ✅ Check release files
- ✅ Parse documentation
- ✅ Validate examples

## 🎯 Step 6: Repository Naming Convention

**⚠️ IMPORTANT:** Terraform Registry requires specific naming:

### Current Repository Name
```
terraform-provider-opnsense-26  ❌ Wrong for registry
```

### Required Name for Registry
```
terraform-provider-opnsense     ✅ Correct
```

### How to Fix

**Option 1: Rename Repository (Recommended)**
1. Go to repository Settings
2. Rename to: `terraform-provider-opnsense`
3. Update your git remote:
   ```bash
   git remote set-url origin https://github.com/rgcosta7/terraform-provider-opnsense.git
   ```

**Option 2: Create New Repository**
1. Create new repo: `terraform-provider-opnsense`
2. Push your code there
3. Register that repo with Terraform Registry

### Namespace on Registry

After publishing, your provider will be available as:

```hcl
terraform {
  required_providers {
    opnsense = {
      source  = "rgcosta7/opnsense"
      version = "~> 0.1.0"
    }
  }
}
```

## 🔍 Troubleshooting

### "Missing SHASUMS signature file"

**Cause:** No `.sig` file in release

**Fix:** 
1. Verify `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE` secrets are set
2. Re-run the workflow or create new release
3. Check release has `terraform-provider-opnsense_VERSION_SHA256SUMS.sig`

### "Invalid signature"

**Cause:** Public key not uploaded or doesn't match

**Fix:**
1. Verify public key on keyserver:
   ```bash
   gpg --keyserver keyserver.ubuntu.com --recv-keys YOUR_KEY_ID
   ```
2. Re-upload public key:
   ```bash
   gpg --keyserver keyserver.ubuntu.com --send-keys YOUR_KEY_ID
   ```
3. Add correct public key in registry settings

### "Repository name must match pattern"

**Cause:** Repository name doesn't follow `terraform-provider-{NAME}` format

**Fix:** Rename repository to `terraform-provider-opnsense`

### "Documentation not found"

**Cause:** Missing `docs/` directory or `docs/index.md`

**Fix:** Ensure proper documentation structure (see Step 3)

## 📚 Additional Resources

- [Terraform Registry Publishing Guide](https://www.terraform.io/docs/registry/providers/publishing.html)
- [Provider Documentation Guide](https://www.terraform.io/docs/registry/providers/docs.html)
- [GPG Key Management](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key)

## ✅ Checklist

Before publishing:

- [ ] Repository name: `terraform-provider-opnsense`
- [ ] GPG key generated
- [ ] Public key uploaded to keyserver
- [ ] `GPG_PRIVATE_KEY` secret added to GitHub
- [ ] `GPG_PASSPHRASE` secret added to GitHub
- [ ] `docs/index.md` exists
- [ ] All resource docs in `docs/resources/`
- [ ] Examples in `examples/`
- [ ] Release created with `BUILD:` commit
- [ ] Release has `.sig` file
- [ ] Signed in to Terraform Registry
- [ ] Provider registered
- [ ] GPG public key added to registry settings

## 🎉 Success!

Once published, users can use your provider:

```hcl
terraform {
  required_providers {
    opnsense = {
      source  = "rgcosta7/opnsense"
      version = "~> 0.1.0"
    }
  }
}
```

No manual installation needed! 🚀
