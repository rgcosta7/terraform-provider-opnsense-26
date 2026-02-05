# Firewall Rule Complete Migration Guide

## ✅ Enhanced Provider Features

The provider now supports:
- ✅ **Multiple categories** per rule (as a list)
- ✅ **Quick field** for immediate action
- ✅ All standard firewall rule options

## 🔄 Complete Schema Conversion

### OLD (25.10) - Nested Schema
```hcl
resource "opnsense_firewall_filter" "wan_to_wg" {
  enabled     = true
  sequence    = 1
  description = "WAN: allow access to WIREGUARD services"

  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.services.id,
  ]

  interface = {
    interface = ["wan"]
  }

  filter = {
    quick     = true
    action    = "pass"
    direction = "in"
    protocol  = "UDP"
    ip_protocol = "inet"

    source   = { net = "any" }
    destination = {
      net  = "wanip"
      port = "51220"
    }
  }
}
```

### NEW (26.1) - Flat Schema
```hcl
resource "opnsense_firewall_rule" "wan_to_wg" {
  enabled     = true
  description = "WAN: allow access to WIREGUARD services"

  # Categories - now supported as list!
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.services.id,
  ]

  # Interface - flat field (string)
  interface = "wan"

  # Action & behavior - flat fields
  action      = "pass"
  quick       = true
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "udp"

  # Source - flat fields
  source_net = "any"

  # Destination - flat fields
  destination_net  = "wanip"
  destination_port = "51220"

  log = false
}
```

## 📋 Field-by-Field Mapping

| Old (25.10) | New (26.1) | Notes |
|-------------|------------|-------|
| `sequence` | ❌ Removed | Not supported in API |
| `categories` | `categories` | ✅ **Now a list!** |
| `interface.interface` | `interface` | String, not object |
| `filter.quick` | `quick` | Moved to root level |
| `filter.action` | `action` | Moved to root level |
| `filter.direction` | `direction` | Moved to root level |
| `filter.protocol` | `protocol` | Moved to root level |
| `filter.ip_protocol` | `ip_protocol` | Moved to root level |
| `filter.source.net` | `source_net` | Flattened |
| `filter.source.port` | `source_port` | Flattened |
| `filter.destination.net` | `destination_net` | Flattened |
| `filter.destination.port` | `destination_port` | Flattened |

## ✅ Complete Examples

### Example 1: WAN to WireGuard
```hcl
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"
}

resource "opnsense_firewall_category" "services" {
  name  = "Services"
  color = "#0000FF"
}

resource "opnsense_firewall_rule" "wan_to_wg" {
  enabled     = true
  description = "WAN: allow access to WIREGUARD services"

  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.services.id,
  ]

  interface   = "wan"
  action      = "pass"
  quick       = true
  direction   = "in"
  protocol    = "udp"
  ip_protocol = "inet"

  source_net       = "any"
  destination_net  = "wanip"
  destination_port = "51220"

  log = false
}
```

### Example 2: IoT to NTP
```hcl
resource "opnsense_firewall_rule" "iot_to_ntp" {
  enabled     = true
  description = "IOT: allow access to NTP services"

  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.general.id,
  ]

  interface = "opt4"
  action    = "pass"
  quick     = true
  direction = "in"
  protocol  = "udp"

  source_net       = "opt4"
  destination_net  = "opt4ip"
  destination_port = "123"

  log = false
}
```

### Example 3: Block Rule
```hcl
resource "opnsense_firewall_rule" "iot_block_all" {
  enabled     = true
  description = "IOT: block all other traffic"

  categories = [
    opnsense_firewall_category.block.id,
  ]

  interface = "opt4"
  action    = "block"
  quick     = true
  direction = "in"
  protocol  = "any"

  source_net      = "any"
  destination_net = "any"

  log = true
}
```

### Example 4: TCP Rule with Port Range
```hcl
resource "opnsense_firewall_rule" "wg_to_services" {
  enabled     = true
  description = "WG: allow access to services"

  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.internal.id,
  ]

  interface = "wg0"
  action    = "pass"
  quick     = true
  direction = "in"
  protocol  = "tcp"

  source_net       = "10.1.1.0/24"
  destination_net  = "192.168.10.0/24"
  destination_port = "80,443,8080-8090"

  log = false
}
```

## 🔧 Conversion Script

Use this pattern to convert your rules:

```bash
# Step 1: Change resource type
sed -i 's/opnsense_firewall_filter/opnsense_firewall_rule/g' *.tf

# Step 2: Your rules are already using the right structure!
# Just need to flatten the nested objects
```

## 📝 Conversion Checklist

For each rule, make these changes:

### 1. Resource Type
```hcl
# OLD
resource "opnsense_firewall_filter" "name" {

# NEW
resource "opnsense_firewall_rule" "name" {
```

### 2. Remove sequence
```hcl
# OLD
sequence = 1

# NEW
# Remove this line - not supported
```

### 3. Keep categories as-is!
```hcl
# Already correct! ✅
categories = [
  opnsense_firewall_category.allow.id,
  opnsense_firewall_category.services.id,
]
```

### 4. Flatten interface
```hcl
# OLD
interface = {
  interface = ["wan"]
}

# NEW
interface = "wan"
```

### 5. Move filter.* to root level
```hcl
# OLD
filter = {
  quick     = true
  action    = "pass"
  protocol  = "UDP"
}

# NEW
quick     = true
action    = "pass"
protocol  = "udp"  # lowercase
```

### 6. Flatten source and destination
```hcl
# OLD
source   = { net = "any" }
destination = {
  net  = "wanip"
  port = "51220"
}

# NEW
source_net       = "any"
destination_net  = "wanip"
destination_port = "51220"
```

## 🎯 Quick Reference

### Required Fields
- `description` - Rule description
- `protocol` - Protocol (tcp, udp, icmp, any)
- `source_net` - Source network
- `destination_net` - Destination network

### Optional Fields
- `categories` - List of category UUIDs
- `interface` - Interface name
- `action` - pass, block, reject (default: pass)
- `quick` - Apply immediately (default: false)
- `direction` - in, out (default: in)
- `ip_protocol` - inet, inet6 (default: inet)
- `source_port` - Source port(s)
- `destination_port` - Destination port(s)
- `enabled` - Enable rule (default: true)
- `log` - Log matches (default: false)

## ✅ Result

After conversion, your rules will:
- ✅ Support multiple categories
- ✅ Work with the flat schema
- ✅ Be fully manageable via Terraform
- ✅ Be future-proof

No more manual GUI configuration needed!
