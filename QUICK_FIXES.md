# Quick Fixes for Terraform Errors

## ✅ All Issues Fixed!

### 1. WireGuard DNS Type Error ✅

**Error:**
```
Inappropriate value for attribute "dns": string required, but have list of string.
```

**Fix:**
```hcl
# OLD (wrong - list)
dns = var.wg_config.dns

# NEW (correct - convert list to comma-separated string)
dns = join(",", var.wg_config.dns)
```

Or update your variable to be a string:
```hcl
variable "wg_config" {
  type = object({
    dns = string  # Change from list(string) to string
    # ...
  })
  
  default = {
    dns = "10.0.20.11,10.0.20.22"  # Comma-separated string
    # ...
  }
}
```

### 2. Firewall Rules - Wrong Resource Name ✅

**Error:**
```
The provider localhost/local/opnsense does not support resource type "opnsense_firewall_filter".
```

**Fix:**
```bash
# Find and replace in your files
sed -i 's/opnsense_firewall_filter/opnsense_firewall_rule/g' opnsense/firewall/rules/main.tf
```

Or manually change:
```hcl
# OLD (wrong)
resource "opnsense_firewall_filter" "wan_to_wg" {

# NEW (correct)
resource "opnsense_firewall_rule" "wan_to_wg" {
```

### 3. NAT Resource - Now Implemented! ✅

**Error:**
```
The provider localhost/local/opnsense does not support resource type "opnsense_firewall_nat".
```

**Fix:** Use the correct resource name:
```hcl
# OLD (wrong)
resource "opnsense_firewall_nat" "traefik_https" {

# NEW (correct)
resource "opnsense_nat_destination" "traefik_https" {
  enabled          = true
  interface        = "wan"
  protocol         = "tcp"
  destination_port = "443"
  target_ip        = "192.168.1.100"
  target_port      = "443"
  description      = "HTTPS to Traefik"
}
```

## 📋 Complete Example NAT Rule

```hcl
# Port forward HTTPS to Traefik
resource "opnsense_nat_destination" "traefik_https" {
  enabled          = true
  interface        = "wan"
  protocol         = "tcp"
  
  # Source (optional, defaults to any)
  source_net       = "any"
  
  # Destination
  destination_port = "443"
  
  # Target (where to forward)
  target_ip        = "192.168.1.100"
  target_port      = "443"
  
  description      = "HTTPS to Traefik"
  log              = false
}

# Port forward SSH
resource "opnsense_nat_destination" "ssh" {
  enabled          = true
  interface        = "wan"
  protocol         = "tcp"
  destination_port = "22"
  target_ip        = "192.168.1.10"
  target_port      = "22"
  description      = "SSH to server"
}

# Port forward with different port
resource "opnsense_nat_destination" "webserver" {
  enabled          = true
  interface        = "wan"
  protocol         = "tcp"
  destination_port = "8080"   # External port
  target_ip        = "192.168.1.50"
  target_port      = "80"     # Internal port
  description      = "Web server"
}
```

## 🔧 Summary of Changes Needed

### In your WireGuard config:
```hcl
resource "opnsense_wireguard_server" "wg_server" {
  # ...
  dns = join(",", var.wg_config.dns)  # Add join()
  # ...
}
```

### In your firewall rules:
```bash
# Run this command
find opnsense/firewall/rules -name "*.tf" -exec sed -i 's/opnsense_firewall_filter/opnsense_firewall_rule/g' {} \;
```

### In your NAT config:
```hcl
# Change resource type
resource "opnsense_nat_destination" "traefik_https" {  # Changed from opnsense_firewall_nat
  # ...
}
```

## 🚀 Rebuild & Test

```bash
# 1. Extract updated provider
tar -xzf terraform-provider-opnsense.tar.gz
cd terraform-provider-opnsense

# 2. Build
./clean-build.sh

# 3. Copy to runner
sudo cp terraform-provider-opnsense /opt/terraform-providers/

# 4. Test
terraform init -upgrade
terraform plan
```

## ✅ After These Changes

All 8 resources will work:
1. ✅ `opnsense_firewall_rule` (fixed name)
2. ✅ `opnsense_firewall_alias`
3. ✅ `opnsense_firewall_category`
4. ✅ `opnsense_nat_destination` (newly implemented!)
5. ✅ `opnsense_kea_subnet`
6. ✅ `opnsense_kea_reservation`
7. ✅ `opnsense_wireguard_server` (dns fixed)
8. ✅ `opnsense_wireguard_peer`

Good night! 🌙
