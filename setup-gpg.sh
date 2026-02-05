#!/bin/bash
# Quick GPG Setup for Terraform Registry
# Run this script to generate and configure GPG keys for signing releases

set -e

echo "🔐 Terraform Registry GPG Setup"
echo "================================"
echo ""

# Check if GPG is installed
if ! command -v gpg &> /dev/null; then
    echo "❌ GPG not found. Please install GPG first:"
    echo "   Ubuntu/Debian: sudo apt install gnupg"
    echo "   macOS: brew install gnupg"
    exit 1
fi

# Get user info
echo "📝 Enter your information:"
read -p "Your name: " NAME
read -p "Your GitHub email: " EMAIL

# Generate key
echo ""
echo "🔑 Generating GPG key (this may take a while)..."
echo ""

gpg --batch --generate-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: $NAME
Name-Email: $EMAIL
Name-Comment: Terraform Provider Signing Key
Expire-Date: 0
EOF

# Get key ID
KEY_ID=$(gpg --list-secret-keys --keyid-format=long "$EMAIL" | grep sec | awk '{print $2}' | cut -d'/' -f2)

echo ""
echo "✅ GPG key generated!"
echo "   Key ID: $KEY_ID"
echo ""

# Export keys
echo "📤 Exporting keys..."
gpg --armor --export-secret-keys "$KEY_ID" > terraform-gpg-private.asc
gpg --armor --export "$KEY_ID" > terraform-gpg-public.asc

# Get fingerprint
FINGERPRINT=$(gpg --fingerprint "$KEY_ID" | grep -A 1 "Key fingerprint" | tail -1 | tr -d ' ')

echo ""
echo "✅ Keys exported!"
echo "   Private key: terraform-gpg-private.asc"
echo "   Public key:  terraform-gpg-public.asc"
echo ""

# Upload to keyserver
echo "📤 Uploading public key to keyserver..."
gpg --keyserver keyserver.ubuntu.com --send-keys "$KEY_ID" || {
    echo "⚠️  Keyserver upload failed. Try manually:"
    echo "   gpg --keyserver keyserver.ubuntu.com --send-keys $KEY_ID"
}

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Setup Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 Next Steps:"
echo ""
echo "1️⃣  Add GitHub Secrets (Settings → Secrets → Actions):"
echo ""
echo "    GPG_PRIVATE_KEY:"
echo "    ----------------------------------------"
cat terraform-gpg-private.asc
echo "    ----------------------------------------"
echo ""
echo "    GPG_PASSPHRASE:"
echo "    (Leave empty - key was generated without passphrase)"
echo ""
echo "2️⃣  Add to Terraform Registry:"
echo ""
echo "    Key ID: $KEY_ID"
echo "    Fingerprint: $FINGERPRINT"
echo ""
echo "    Public Key:"
echo "    ----------------------------------------"
cat terraform-gpg-public.asc
echo "    ----------------------------------------"
echo ""
echo "3️⃣  Verify keyserver upload:"
echo ""
echo "    gpg --keyserver keyserver.ubuntu.com --recv-keys $KEY_ID"
echo ""
echo "4️⃣  Secure cleanup (after adding to GitHub):"
echo ""
echo "    shred -vfz -n 10 terraform-gpg-private.asc"
echo "    rm terraform-gpg-public.asc"
echo ""
echo "5️⃣  Create signed release:"
echo ""
echo "    git commit -m 'BUILD: v0.1.0 - Initial release'"
echo "    git push origin main"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
