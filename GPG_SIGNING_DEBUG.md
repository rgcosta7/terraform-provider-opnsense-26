# GPG Signing Debug Guide

## 🐛 Current Issue

GPG key imports successfully but signing fails:
```
❌ Error: Signature file not created
```

## 🔍 Possible Causes

### 1. Passphrase Protected Key

**Problem:** Your GPG key was created WITH a passphrase

**Check:**
```bash
# Check if key has passphrase
gpg --list-secret-keys --with-colons | grep -A 1 "^sec"
```

**Fix:** You need to either:

**Option A: Remove passphrase from key (RECOMMENDED for CI)**
```bash
# Export key without passphrase
gpg --export-secret-keys YOUR_KEY_ID > private.gpg
gpg --batch --yes --pinentry-mode=loopback --passphrase="YOUR_PASSPHRASE" \
  --import private.gpg
gpg --batch --yes --pinentry-mode=loopback --passphrase="YOUR_PASSPHRASE" \
  --change-passphrase YOUR_KEY_ID
# When prompted, leave new passphrase EMPTY
# Then re-export
gpg --armor --export-secret-keys YOUR_KEY_ID > new-private-key.asc
# Update GitHub secret with this new key
```

**Option B: Use the setup-gpg.sh script**

The script creates a key WITHOUT passphrase specifically for CI:
```bash
./setup-gpg.sh
```

This key has `%no-protection` so it works in CI without a passphrase.

### 2. Key Trust Issue

**Problem:** Key not trusted for signing

**Fix in workflow:** Already added `--batch --yes --pinentry-mode loopback`

### 3. GPG Version Compatibility

**Problem:** Different GPG versions have different defaults

**Check workflow logs for:**
```
GnuPG version: 2.x.x
```

## 🎯 Recommended Solution

**Create a NEW key specifically for CI (no passphrase):**

```bash
# 1. Create batch configuration
cat > gpg-ci-key.conf <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: Raul Costa
Name-Email: rgcosta7@gmail.com
Name-Comment: Terraform Provider CI Signing
Expire-Date: 0
EOF

# 2. Generate key
gpg --batch --generate-key gpg-ci-key.conf

# 3. Get the new key ID
NEW_KEY_ID=$(gpg --list-secret-keys --keyid-format=long rgcosta7@gmail.com | grep sec | tail -1 | awk '{print $2}' | cut -d'/' -f2)
echo "New Key ID: $NEW_KEY_ID"

# 4. Export for GitHub (no passphrase!)
gpg --armor --export-secret-keys $NEW_KEY_ID > ci-private-key.asc

# 5. Export public key
gpg --armor --export $NEW_KEY_ID > ci-public-key.asc

# 6. Upload to keyserver
gpg --keyserver keyserver.ubuntu.com --send-keys $NEW_KEY_ID

# 7. Verify it works
echo "test" > test.txt
gpg --batch --yes --pinentry-mode loopback --detach-sign --armor test.txt
ls -la test.txt.sig  # Should exist
gpg --verify test.txt.sig test.txt  # Should succeed

# 8. Update GitHub Secrets
# - Replace GPG_PRIVATE_KEY with content of ci-private-key.asc
# - Leave GPG_PASSPHRASE empty (or delete it)

# 9. Update Terraform Registry
# - Add the NEW public key (ci-public-key.asc)
# - Or re-upload to keyserver

# 10. Cleanup
shred -vfz -n 10 ci-private-key.asc gpg-ci-key.conf
```

## 🔍 Debug Checklist

When the workflow fails, check these in the logs:

1. **Key imported?**
   ```
   ✅ gpg: key XXXXXXXX: secret key imported
   ```

2. **Key listed?**
   ```
   ✅ sec   rsa4096/XXXXXXXX
   ```

3. **File exists before signing?**
   ```
   ✅ terraform-provider-opnsense_0.1.0_SHA256SUMS
   ```

4. **Signing command shows error?**
   Look for GPG error messages

5. **Signature file created?**
   ```
   ✅ terraform-provider-opnsense_0.1.0_SHA256SUMS.sig
   ```

## 🧪 Test Locally

Test your key works for signing:

```bash
# 1. Import your private key
gpg --import terraform-gpg-private.asc

# 2. Create test file
echo "test" > test.txt

# 3. Try signing (CI mode)
gpg --batch --yes --pinentry-mode loopback --detach-sign --armor test.txt

# 4. Check if signature created
ls -la test.txt.sig

# 5. Verify
gpg --verify test.txt.sig test.txt
```

If this fails locally, your key has issues. Generate a new one.

## ✅ Working Example

This is what should happen:

```bash
$ gpg --batch --yes --pinentry-mode loopback --detach-sign --armor test.txt
$ ls test.txt.sig
test.txt.sig  ← Created successfully
$ gpg --verify test.txt.sig test.txt
gpg: Signature made Wed 05 Feb 2026 08:00:00 PM UTC
gpg:                using RSA key XXXXXXXX
gpg: Good signature from "Raul Costa <rgcosta7@gmail.com>"
```

## 🚨 Most Common Issue

**90% of the time it's: Key has a passphrase but workflow doesn't provide it**

**Solution:** Generate new key WITHOUT passphrase using the script:
```bash
./setup-gpg.sh
```

Then replace the GitHub secret with the new key.

## 📞 Still Not Working?

If you've tried everything:

1. Share the workflow log output (from signing step)
2. Run the local test and share results
3. Check if `GPG_PASSPHRASE` secret is set (might be wrong)
