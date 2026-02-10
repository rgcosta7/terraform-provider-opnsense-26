# Terraform Provider for OPNsense

Manage OPNsense firewall, DHCP, VPN, and NAT configurations with Terraform.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Terraform](https://img.shields.io/badge/terraform-%3E%3D1.0-blue)](https://www.terraform.io)

## Features

- 🔥 **Firewall Management**
  - Complete rule configuration with all GUI fields
  - Categories with color coding
  - Aliases (IP/network/port groups)
  - Rule sequencing and ordering
  - Gateway routing (policy-based routing)
  - Source/destination inversion (NOT logic)
  - IPv4 and IPv6 support

- 🌐 **DHCP (Kea)**
  - Subnet management with DHCP options
  - Static reservations (MAC → IP mapping)
  - Bulk management with for_each pattern

- 🔐 **VPN (WireGuard)**
  - Server configuration
  - Peer management

- 🔀 **NAT**
  - Destination NAT (port forwarding)

## Quick Start

### Installation

```hcl
terraform {
  required_providers {
    opnsense = {
      source  = "your-org/opnsense"
      version = "~> 1.0"
    }
  }
}

provider "opnsense" {
  host       = "https://10.0.10.10"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
}
```

### Basic Example

```hcl
# Create a category
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"
}

# Create an alias
resource "opnsense_firewall_alias" "dns_servers" {
  name    = "DNS_SERVERS"
  type    = "host"
  content = ["10.0.20.11", "10.0.20.22"]
}

# Create a firewall rule
resource "opnsense_firewall_rule" "allow_dns" {
  enabled     = true
  sequence    = 100
  description = "Allow DNS queries"
  
  interface   = "lan"
  protocol    = "udp"
  source_net  = "lan"
  destination_net  = "_DNS_SERVERS"
  destination_port = "53"
  
  action = "pass"
  log    = false
  
  categories = [opnsense_firewall_category.allow.id]
}

# Create DHCP subnet
resource "opnsense_kea_subnet" "mgmt" {
  subnet      = "10.0.10.0/24"
  pools       = "10.0.10.100-10.0.10.200"
  description = "Management VLAN"
  
  option_data = {
    routers             = "10.0.10.1"
    domain-name-servers = "10.0.20.11, 10.0.20.22"
    ntp-servers         = "10.0.10.1"
  }
}

# Create DHCP reservation
resource "opnsense_kea_reservation" "server1" {
  subnet      = opnsense_kea_subnet.mgmt.id
  ip_address  = "10.0.10.20"
  hw_address  = "aa:bb:cc:dd:ee:ff"
  hostname    = "server1"
  description = "Web server"
}
```

## Resources

### Firewall

#### opnsense_firewall_rule

Complete firewall rule with all OPNsense GUI fields.

```hcl
resource "opnsense_firewall_rule" "example" {
  # Organization
  enabled     = true
  sequence    = 500
  description = "Example rule"
  categories  = [opnsense_firewall_category.allow.id]
  
  # Interface
  interface = "lan"
  
  # Filter
  action      = "pass"
  quick       = true
  direction   = "in"
  ip_protocol = "inet"  # or "inet6"
  protocol    = "tcp"
  
  # Source
  source_not  = false
  source_net  = "lan"
  source_port = "any"
  
  # Destination
  destination_not  = false
  destination_net  = "any"
  destination_port = "443"
  
  # Routing
  gateway = "BLUEDRAGON"  # Optional: route via specific gateway
  
  # Logging
  log = true
}
```

**Key Fields:**
- `sequence` - Rule order (lower = processed first)
- `categories` - List of category UUIDs for organization
- `gateway` - Route traffic via specific gateway (VPN, multi-WAN)
- `destination_not` / `source_not` - Invert match (NOT logic)
- `ip_protocol` - "inet" (IPv4) or "inet6" (IPv6)

**Common Patterns:**

```hcl
# Allow to internet only (NOT private networks)
resource "opnsense_firewall_rule" "internet_only" {
  destination_not = true
  destination_net = "_PRIVATE_NETWORKS"
  # Matches anything EXCEPT private networks
}

# Route specific traffic via VPN
resource "opnsense_firewall_rule" "vpn_route" {
  gateway = "WIREGUARD_GW"
  # All matching traffic routes via VPN
}

# IPv4 and IPv6 rules
resource "opnsense_firewall_rule" "web_v4" {
  ip_protocol = "inet"
  # ... IPv4 rule
}

resource "opnsense_firewall_rule" "web_v6" {
  ip_protocol = "inet6"
  # ... IPv6 rule
}
```

[→ Complete field reference](docs/resources/firewall_rule.md)

#### opnsense_firewall_category

Organize rules with visual categories.

```hcl
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"  # Green
  auto  = false      # Don't auto-delete when unused
}
```

**Common Colors:**
- Allow: `#00FF00` (Green)
- Block: `#FF0000` (Red)
- IoT: `#800080` (Purple)
- VPN: `#0000FF` (Blue)

[→ Complete field reference](docs/resources/firewall_category.md)

#### opnsense_firewall_alias

Create groups of IPs, networks, or ports.

```hcl
resource "opnsense_firewall_alias" "dns_servers" {
  name    = "DNS_SERVERS"
  type    = "host"
  content = ["10.0.20.11", "10.0.20.22", "8.8.8.8"]
  description = "DNS server list"
  enabled = true
}
```

**Alias Types:**
- `host` - IP addresses
- `network` - Network ranges (CIDR)
- `port` - Port numbers/ranges
- `url` - URLs for dynamic lists
- `mac` - MAC addresses
- `geoip` - Geographic IP blocks

[→ Complete field reference](docs/resources/firewall_alias.md)

### DHCP (Kea)

#### opnsense_kea_subnet

DHCP subnet with options.

```hcl
resource "opnsense_kea_subnet" "vlan10" {
  subnet       = "10.0.10.0/24"
  pools        = "10.0.10.100-10.0.10.200"
  description  = "Management VLAN"
  auto_collect = false
  
  option_data = {
    routers             = "10.0.10.1"
    domain-name-servers = "10.0.20.11, 10.0.20.22"
    domain-name         = "mgmt.local"
    ntp-servers         = "10.0.10.1"
  }
}
```

**Supported DHCP Options:**
- `routers` - Default gateway
- `domain-name-servers` - DNS servers (comma-separated)
- `domain-name` - DNS domain
- `domain-search` - DNS search domains
- `ntp-servers` - NTP servers
- `time-servers` - Time servers
- `tftp-server-name` - TFTP server
- `boot-file-name` - Boot filename (PXE)

[→ Complete field reference](docs/resources/kea_subnet.md)

#### opnsense_kea_reservation

Static IP assignments.

```hcl
resource "opnsense_kea_reservation" "server1" {
  subnet      = opnsense_kea_subnet.vlan10.id
  ip_address  = "10.0.10.20"
  hw_address  = "aa:bb:cc:dd:ee:ff"
  hostname    = "server1"
  description = "Web server"
}
```

**for_each Pattern** (recommended for many reservations):

```hcl
resource "opnsense_kea_reservation" "reservations" {
  for_each = var.kea_reservations

  subnet      = each.value.subnet_id
  ip_address  = each.value.ip_address
  hw_address  = each.value.hw_address
  hostname    = each.value.hostname
  description = try(each.value.description, "")
}
```

[→ Complete field reference](docs/resources/kea_reservation.md)

### VPN (WireGuard)

#### opnsense_wireguard_server

WireGuard VPN server.

```hcl
resource "opnsense_wireguard_server" "main" {
  name            = "wg0"
  enabled         = true
  port            = 51820
  tunnel_address  = "10.255.0.1/24"
}
```

[→ Complete field reference](docs/resources/wireguard_server.md)

#### opnsense_wireguard_peer

WireGuard VPN clients.

```hcl
resource "opnsense_wireguard_peer" "laptop" {
  name        = "laptop"
  enabled     = true
  public_key  = "base64-key-here"
  allowed_ips = "10.255.0.2/32"
}
```

[→ Complete field reference](docs/resources/wireguard_peer.md)

### NAT

#### opnsense_nat_destination

Port forwarding / destination NAT.

```hcl
resource "opnsense_nat_destination" "web_https" {
  interface        = "wan"
  protocol         = "tcp"
  destination      = "wanip"
  destination_port = "443"
  target           = "10.0.20.80"
  local_port       = "443"
  description      = "HTTPS to web server"
}
```

[→ Complete field reference](docs/resources/nat_destination.md)

## Advanced Usage

### Policy-Based Routing

Route different traffic via different gateways:

```hcl
# Work traffic via VPN
resource "opnsense_firewall_rule" "work_vpn" {
  description = "Work traffic via VPN"
  source_net  = "_WORK_DEVICES"
  gateway     = "WIREGUARD_GW"
  action      = "pass"
}

# Streaming via WAN1
resource "opnsense_firewall_rule" "streaming_wan1" {
  description = "Streaming via WAN1"
  source_net  = "_STREAMING_DEVICES"
  gateway     = "WAN1_DHCP"
  action      = "pass"
}

# Everything else via default
resource "opnsense_firewall_rule" "default" {
  sequence    = 999
  description = "Default routing"
  source_net  = "any"
  # No gateway = default route
  action      = "pass"
}
```

### Split Tunneling

Route specific destinations via VPN:

```hcl
# Corporate networks via VPN
resource "opnsense_firewall_rule" "corporate_vpn" {
  description     = "Corporate networks via VPN"
  destination_net = "_CORPORATE_NETWORKS"
  gateway         = "WIREGUARD_GW"
  action          = "pass"
}

# Everything else via normal gateway
resource "opnsense_firewall_rule" "internet_normal" {
  description     = "Internet via normal gateway"
  destination_not = true
  destination_net = "_CORPORATE_NETWORKS"
  # No gateway specified = default
  action          = "pass"
}
```

### Managing Hundreds of Rules/Reservations

Use for_each with variables:

**variables.tf:**
```hcl
variable "firewall_rules" {
  type = map(object({
    description = string
    interface   = string
    protocol    = string
    source_net  = string
    dest_net    = string
    action      = string
  }))
}
```

**main.tf:**
```hcl
resource "opnsense_firewall_rule" "rules" {
  for_each = var.firewall_rules

  description     = each.value.description
  interface       = each.value.interface
  protocol        = each.value.protocol
  source_net      = each.value.source_net
  destination_net = each.value.dest_net
  action          = each.value.action
}
```

**terraform.tfvars:**
```hcl
firewall_rules = {
  "allow_dns" = {
    description = "Allow DNS"
    interface   = "lan"
    protocol    = "udp"
    source_net  = "lan"
    dest_net    = "_DNS_SERVERS"
    action      = "pass"
  }
  # ... 100 more rules ...
}
```

### Modular Organization

```
terraform/
├── main.tf
├── variables.tf
├── terraform.tfvars  # Encrypted!
├── modules/
│   ├── firewall/
│   │   ├── categories.tf
│   │   ├── aliases.tf
│   │   └── rules.tf
│   ├── dhcp/
│   │   ├── subnets.tf
│   │   └── reservations.tf
│   └── vpn/
│       ├── server.tf
│       └── peers.tf
```

## Best Practices

### 1. Use Categories for Organization

```hcl
# Define categories first
resource "opnsense_firewall_category" "allow" { ... }
resource "opnsense_firewall_category" "block" { ... }
resource "opnsense_firewall_category" "iot" { ... }

# Tag all rules
resource "opnsense_firewall_rule" "iot_dns" {
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.iot.id,
  ]
}
```

### 2. Use Aliases for Reusability

```hcl
# Define once
resource "opnsense_firewall_alias" "dns_servers" {
  name = "DNS_SERVERS"
  type = "host"
  content = ["10.0.2.11", "10.0.2.22"]
}

# Use everywhere
resource "opnsense_firewall_rule" "allow_dns" {
  destination_net = "_DNS_SERVERS"
}
```

### 3. Use Sequence for Ordering

```hcl
# Critical rules first
resource "opnsense_firewall_rule" "critical" {
  sequence = 100
}

# Normal rules
resource "opnsense_firewall_rule" "normal" {
  sequence = 500
}

# Catch-all last
resource "opnsense_firewall_rule" "default" {
  sequence = 999
}
```

### 4. Use for_each for Scale

```hcl
# Bad: 100 individual resources
resource "opnsense_kea_reservation" "server1" { ... }
resource "opnsense_kea_reservation" "server2" { ... }
# ... 98 more ...

# Good: One resource with for_each
resource "opnsense_kea_reservation" "reservations" {
  for_each = var.kea_reservations
  # ...
}
```

### 5. Version Control

```bash
# Initialize Git
git init

# Add .gitignore
echo "*.tfstate*" >> .gitignore
echo ".terraform/" >> .gitignore
echo "terraform.tfvars" >> .gitignore  # Contains secrets

# Commit
git add .
git commit -m "Initial OPNsense configuration"
```

### 6. Encrypt Sensitive Data

```bash
# Encrypt tfvars with git-crypt, sops, or Vault
git-crypt init
echo "terraform.tfvars filter=git-crypt diff=git-crypt" >> .gitattributes
```

## Testing

### Test in Non-Production First

```hcl
# Create test firewall in separate config
resource "opnsense_firewall_rule" "test" {
  description = "TEST - Delete me"
  # ... test configuration
}
```

### Verify Changes Before Apply

```bash
# Always review plan
terraform plan

# Apply only specific resources
terraform apply -target=opnsense_firewall_rule.test

# Review in OPNsense GUI before full apply
```

### Backup First

```bash
# OPNsense: System → Configuration → Backups
# Download configuration before terraform apply
```

## Troubleshooting

### Enable Debug Logging

```bash
export TF_LOG=DEBUG
terraform apply 2>&1 | tee terraform-debug.log
```

### Common Issues

**Categories not showing:**
- Run `terraform apply` to update existing categories with colors
- Check category UUIDs are valid

**Rules not in correct order:**
- Use `sequence` field to control order
- Lower numbers = processed first

**Gateway not working:**
- Use gateway **name** (e.g., "BLUEDRAGON"), not interface (e.g., "opt7")
- Check: System → Gateways → Single for gateway names

**Alias not found:**
- Prefix with underscore: `"_ALIAS_NAME"`
- Or use alias name without underscore

## Migration

### From Manual Configuration

1. **Export current config** from OPNsense GUI
2. **Create Terraform resources** matching current state
3. **Import existing resources:**
   ```bash
   terraform import opnsense_firewall_rule.example <UUID>
   ```
4. **Verify with plan:**
   ```bash
   terraform plan  # Should show no changes
   ```

### From Other Tools

Import existing UUIDs:
```bash
# Get UUID from OPNsense
terraform import opnsense_firewall_rule.web_allow <uuid-from-opnsense>
```

## Examples

See [examples/](examples/) directory for complete examples:
- [Complete network setup](examples/complete/)
- [Multi-VLAN with DHCP](examples/multi-vlan/)
- [VPN split tunneling](examples/vpn-split/)
- [Policy-based routing](examples/policy-routing/)
- [IoT network isolation](examples/iot-isolation/)

## Documentation

- [Complete Resources Reference](docs/ALL_RESOURCES_COMPLETE_REFERENCE.md)
- [CHANGELOG](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)

## Requirements

- Terraform >= 1.0
- OPNsense >= 24.x
- OPNsense API key with appropriate permissions

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/your-org/terraform-provider-opnsense/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/terraform-provider-opnsense/discussions)

## Acknowledgments

Built with ❤️ for the OPNsense community.

Special thanks to all contributors who helped make this provider production-ready!