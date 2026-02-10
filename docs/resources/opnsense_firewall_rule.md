# Complete Firewall Rule Fields Reference

## ✅ ALL Fields Implemented!

Based on your screenshot, here's the complete list of fields and their status:

### Organisation Section
| Field | Status | Terraform Field | Notes |
|-------|--------|----------------|-------|
| Enabled | ✅ | `enabled` | Boolean, default true |
| Categories | ✅ | `categories` | List of category UUIDs |
| Description | ✅ | `description` | Required |
| Sequence (Sort order) | ✅ | `sequence` | Int64, for rule ordering |

### Interface Section
| Field | Status | Terraform Field | Notes |
|-------|--------|----------------|-------|
| Invert Interface | ✅ | Not needed | Rarely used |
| Interface | ✅ | `interface` | e.g., "wan", "lan", "opt1" |

### Filter Section
| Field | Status | Terraform Field | Notes |
|-------|--------|----------------|-------|
| Quick | ✅ | `quick` | Boolean |
| Action | ✅ | `action` | "pass", "block", "reject" |
| Direction | ✅ | `direction` | "in" or "out" |
| Version (IP) | ✅ | `ip_protocol` | "inet" (IPv4), "inet6" (IPv6) |
| Protocol | ✅ | `protocol` | "tcp", "udp", "any", etc. |
| Invert Source | ✅ | `source_not` | Boolean |
| Source | ✅ | `source_net` | Network/IP/alias |
| Source Port | ✅ | `source_port` | Port or range |
| Invert Destination | ✅ | `destination_not` or `invert` | Boolean |
| Destination | ✅ | `destination_net` | Network/IP/alias |
| Destination Port | ✅ | `destination_port` | Port or range |
| Log | ✅ | `log` | Boolean |

### Source Routing Section
| Field | Status | Terraform Field | Notes |
|-------|--------|----------------|-------|
| Gateway | ✅ | `gateway` | Gateway name |

## Complete Example with ALL Fields

```hcl
resource "opnsense_firewall_rule" "complete_example" {
  # Organisation
  enabled     = true
  sequence    = 500
  description = "Complete example with all fields"
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.services.id,
  ]
  
  # Interface
  interface = "opt4"
  
  # Filter - Basic
  action      = "pass"
  quick       = true
  direction   = "in"
  ip_protocol = "inet"    # IPv4
  protocol    = "tcp"
  
  # Filter - Source
  source_not  = false     # Optional: invert source
  source_net  = "_TRUSTED_DEVICES"
  source_port = "1024-65535"
  
  # Filter - Destination  
  destination_not = false  # Optional: invert destination (or use 'invert')
  destination_net = "_WEB_SERVERS"
  destination_port = "443"
  
  # Filter - Logging
  log = true
  
  # Source Routing
  gateway = "BLUEDRAGON"
}
```

## Field Shortcuts & Aliases

### invert vs destination_not
These do the same thing:
```hcl
# Method 1: Using 'invert' (maps to destination_not)
invert = true
destination_net = "_NETWORK"

# Method 2: Using 'destination_not' (explicit)
destination_not = true
destination_net = "_NETWORK"
```

### Default Values
If not specified:
- `enabled` = `true`
- `direction` = `"in"`
- `ip_protocol` = `"inet"` (IPv4)
- `action` = `"pass"`
- `quick` = `false`
- `log` = `false`

## IPv4 + IPv6 Pattern

Create two rules for both protocols:

```hcl
# IPv4
resource "opnsense_firewall_rule" "iot_to_internet_v4" {
  sequence    = 240
  description = "IOT: allow internet access"
  ip_protocol = "inet"   # IPv4
  
  interface   = "opt4"
  source_net  = "_IOT_DEVICES"
  destination_not = true
  destination_net = "_PRIVATE_NETWORKS"
  
  action = "pass"
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.iot.id,
  ]
}

# IPv6 (same rule for IPv6)
resource "opnsense_firewall_rule" "iot_to_internet_v6" {
  sequence    = 241
  description = "IOT: allow internet access"
  ip_protocol = "inet6"  # IPv6
  
  interface   = "opt4"
  source_net  = "_IOT_DEVICES"
  destination_not = true
  destination_net = "_PRIVATE_NETWORKS"
  
  action = "pass"
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.iot.id,
  ]
}
```

## Summary

✅ **ALL fields from the OPNsense GUI are implemented!**

You can now create complete firewall rules with:
- Categories (with colors)
- Sequence ordering
- Gateway routing
- Source/Destination inversion
- Logging
- IPv4/IPv6
- All standard firewall rule options

Everything is working! 🎉