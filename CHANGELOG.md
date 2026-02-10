# Changelog

All notable changes to the OPNsense Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1]

### Added
- **Firewall Rules**: Complete firewall rule management with all OPNsense GUI fields
  - Rule sequencing/ordering support
  - Category assignment (multiple categories per rule)
  - Gateway routing (policy-based routing, split tunneling)
  - Source and destination NOT/invert functionality
  - Full IPv4 and IPv6 support
  - Logging configuration
  - All protocol types (TCP, UDP, ICMP, any)
  - Port ranges and aliases support

- **Firewall Categories**: Rule organization with visual categories
  - Color coding support (hex colors)
  - Auto-delete when unused option
  - Full CRUD operations

- **Firewall Aliases**: IP/network/port grouping
  - Multiple alias types (host, network, port, url, urltable, geoip, mac)
  - Multi-value content support
  - Reference in firewall rules

- **Kea DHCP Subnets**: Complete DHCP subnet management
  - DHCP option data support (DNS, NTP, gateway, domain, etc.)
  - IP pool configuration
  - Auto-collect lease management
  - Direct string format for option_data (no nested objects)

- **Kea DHCP Reservations**: Static IP assignments
  - MAC to IP mapping
  - Hostname configuration
  - Subnet association
  - for_each pattern support for bulk management

### Fixed
- **Kea DHCP**: Fixed option_data format based on OPNsense XML model
  - Changed from nested object format to direct string values
  - Automatic comma-space normalization
  - Support for comma-separated lists (DNS servers, etc.)
  
- **Kea DHCP**: Fixed reservation API key from "reservation4" to "reservation"
  
- **Kea DHCP**: Fixed Read function Content-Type header
  - Only set Content-Type when request has body
  - Prevents 400 errors on GET requests

- **Firewall Rules**: Fixed destination_not and source_not functionality
  - Changed from sending "invert" field to correct "destination_not"/"source_not" fields
  - Both explicit fields and "invert" alias now work correctly

- **Firewall Rules**: Removed unnecessary API response warnings
  - Changed API response logging from warning to debug level
  - Removed deprecation warnings from destination_not/source_not fields

### Changed
- **Firewall Aliases**: Content format changed from comma-separated to newline-separated
  - Aligns with OPNsense API expectations
  - Improves multi-line alias readability

### Documentation
- Added complete field reference for all resources
- Added comprehensive examples for common use cases
- Added for_each patterns for scaling to hundreds of resources
- Added troubleshooting guides
- Added API testing scripts

## [0.1.0] - Initial Release

### Added
- Initial provider structure
- Basic firewall rule support
- Basic Kea DHCP support
- WireGuard support
- NAT destination rule support

---

## Upgrade Notes

### Upgrading to Unreleased

#### Firewall Rules
- **destination_not/source_not**: No longer deprecated. You can use either:
  - `destination_not = true` (explicit, recommended)
  - `invert = true` (alias for destination_not)
  - `source_not = true` (for source inversion)

- **Categories**: Now fully functional. Add colors to existing categories:
  ```hcl
  resource "opnsense_firewall_category" "allow" {
    name  = "Allow"
    color = "#00FF00"  # Add this
  }
  ```
  Run `terraform apply` to update existing categories.

- **Gateway**: New field for policy-based routing:
  ```hcl
  resource "opnsense_firewall_rule" "vpn_route" {
    gateway = "BLUEDRAGON"  # Route via specific gateway
    # ...
  }
  ```

- **Sequence**: New field for rule ordering:
  ```hcl
  resource "opnsense_firewall_rule" "critical" {
    sequence = 100  # Processed first
    # ...
  }
  ```

#### Kea DHCP Subnets
- **option_data format change**: If you have existing subnets with option_data, the format has changed from nested objects to direct strings:
  
  **Old (broken):**
  ```hcl
  option_data = {
    routers = {
      value = "10.0.10.1"
    }
  }
  ```
  
  **New (working):**
  ```hcl
  option_data = {
    routers = "10.0.10.1"
    domain-name-servers = "10.0.20.11, 10.0.20.22"
  }
  ```
  
  Update your configurations and re-apply.

#### Kea DHCP Reservations
- No breaking changes. Existing reservations continue to work.
- Consider migrating to for_each pattern for easier management:
  ```hcl
  resource "opnsense_kea_reservation" "reservations" {
    for_each = var.kea_reservations
    # ...
  }
  ```

### Migration Guide

#### From Individual Resources to for_each

**Before:**
```hcl
resource "opnsense_kea_reservation" "server1" { ... }
resource "opnsense_kea_reservation" "server2" { ... }
# ... 100 more ...
```

**After:**
```hcl
# In main.tf
resource "opnsense_kea_reservation" "reservations" {
  for_each = var.kea_reservations
  subnet      = each.value.subnet_id
  ip_address  = each.value.ip_address
  hw_address  = each.value.hw_address
  hostname    = each.value.hostname
  description = try(each.value.description, "")
}

# In terraform.tfvars
kea_reservations = {
  "server1" = { subnet_id = "...", ip_address = "...", hw_address = "...", hostname = "..." }
  "server2" = { subnet_id = "...", ip_address = "...", hw_address = "...", hostname = "..." }
}
```

---

## Support

- **Issues**: [GitHub Issues](https://github.com/your-repo/terraform-provider-opnsense/issues)
- **Documentation**: [Complete Field Reference](docs/RESOURCES.md)
- **Examples**: [examples/](examples/)

---

## Contributors

Thanks to all contributors who helped make this provider production-ready!