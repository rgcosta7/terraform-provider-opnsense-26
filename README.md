# Terraform Provider for OPNsense 26.1

A Terraform provider for managing OPNsense 26.1 firewall configuration via the API. This provider supports managing firewall rules, aliases, Kea DHCP, and WireGuard VPN configurations.

## Features

- **Firewall Management**
  - Firewall rules (filter rules)
  - Firewall aliases (host, network, port, etc.)
  
- **Kea DHCP Server**
  - DHCP subnets
  - DHCP reservations
  
- **WireGuard VPN**
  - WireGuard server instances
  - WireGuard peers/clients

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22 (for building from source)
- OPNsense 26.1 or later
- API access enabled on OPNsense

## Building the Provider

Clone the repository and build:

```bash
git clone https://github.com/rgcosta7/terraform-provider-opnsense-26
cd terraform-provider-opnsense
go build -o terraform-provider-opnsense
```

## Installation

### Local Development

For local development, you can use the provider by placing it in the appropriate directory:

```bash
# Create the plugins directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/rgcosta7/opnsense/0.1.0/linux_amd64/

# Copy the built binary
cp terraform-provider-opnsense ~/.terraform.d/plugins/registry.terraform.io/rgcosta7/opnsense/0.1.0/linux_amd64/
```

Update your Terraform configuration to use the local provider:

```hcl
terraform {
  required_providers {
    opnsense = {
      source  = "rgcosta7/opnsense"
      version = "0.1.0"
    }
  }
}
```

## OPNsense API Setup

1. Log into your OPNsense web interface
2. Navigate to **System > Access > Users**
3. Select or create a user
4. Click the **+** icon in the API keys section
5. Download the generated API key file (contains key and secret)
6. Note the key and secret for use in the provider configuration

## Usage

### Provider Configuration

```hcl
provider "opnsense" {
  host       = "https://192.168.1.1"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
  insecure   = true  # Set to false with valid certificates
}
```

Or use environment variables:

```bash
export OPNSENSE_HOST="https://192.168.1.1"
export OPNSENSE_API_KEY="your-api-key"
export OPNSENSE_API_SECRET="your-api-secret"
```

### Provider Arguments

- `host` (Required) - OPNsense host URL (e.g., `https://192.168.1.1`)
- `api_key` (Required) - API key from OPNsense
- `api_secret` (Required) - API secret from OPNsense
- `insecure` (Optional) - Skip TLS certificate verification. Default: `false`
- `timeout_seconds` (Optional) - HTTP timeout in seconds. Default: `30`

## Resources

### opnsense_firewall_rule

Manages firewall filter rules.

```hcl
resource "opnsense_firewall_rule" "allow_http" {
  description      = "Allow HTTP from LAN"
  interface        = "lan"
  direction        = "in"
  ip_protocol      = "inet"
  protocol         = "tcp"
  source_net       = "192.168.1.0/24"
  destination_net  = "any"
  destination_port = "80"
  action           = "pass"
  enabled          = true
  log              = true
}
```

**Arguments:**
- `description` (Required) - Description of the rule
- `interface` (Optional) - Interface name (wan, lan, opt1, etc.)
- `direction` (Optional) - Traffic direction: `in` or `out`. Default: `in`
- `ip_protocol` (Optional) - IP version: `inet` (IPv4) or `inet6` (IPv6). Default: `inet`
- `protocol` (Required) - Protocol: tcp, udp, icmp, any, etc.
- `source_net` (Required) - Source network/IP (CIDR or 'any')
- `source_port` (Optional) - Source port or port range
- `destination_net` (Required) - Destination network/IP
- `destination_port` (Optional) - Destination port or port range
- `action` (Optional) - Action: `pass`, `block`, or `reject`. Default: `pass`
- `enabled` (Optional) - Enable the rule. Default: `true`
- `log` (Optional) - Log matching packets
- `category` (Optional) - Rule category for organization

**Attributes:**
- `id` - Rule UUID

### opnsense_firewall_alias

Manages firewall aliases.

```hcl
resource "opnsense_firewall_alias" "dns_servers" {
  name        = "public_dns"
  type        = "host"
  content     = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
  description = "Public DNS servers"
  enabled     = true
}
```

**Arguments:**
- `name` (Required) - Alias name
- `type` (Required) - Alias type: host, network, port, url, urltable, geoip, mac, etc.
- `content` (Required) - List of alias entries
- `description` (Optional) - Description
- `enabled` (Optional) - Enable the alias. Default: `true`

**Attributes:**
- `id` - Alias UUID

### opnsense_kea_subnet

Manages Kea DHCP subnets.

```hcl
resource "opnsense_kea_subnet" "lan" {
  subnet      = "192.168.1.0/24"
  pools       = "192.168.1.100-192.168.1.200"
  description = "LAN DHCP Subnet"
}
```

**Arguments:**
- `subnet` (Required) - Subnet in CIDR notation
- `pools` (Optional) - IP address pool ranges (comma-separated)
- `option_data` (Optional) - DHCP options
- `description` (Optional) - Description

**Attributes:**
- `id` - Subnet UUID

### opnsense_kea_reservation

Manages Kea DHCP reservations.

```hcl
resource "opnsense_kea_reservation" "server" {
  subnet      = opnsense_kea_subnet.lan.id
  ip_address  = "192.168.1.10"
  hw_address  = "00:11:22:33:44:55"
  hostname    = "server1"
  description = "Main server"
}
```

**Arguments:**
- `subnet` (Required) - Subnet UUID
- `ip_address` (Required) - Reserved IP address
- `hw_address` (Required) - MAC address
- `hostname` (Optional) - Hostname
- `description` (Optional) - Description

**Attributes:**
- `id` - Reservation UUID

### opnsense_wireguard_server

Manages WireGuard server instances.

```hcl
resource "opnsense_wireguard_server" "vpn" {
  name            = "wg0"
  enabled         = true
  listen_port     = 51820
  tunnel_address  = "10.20.30.1/24"
  peers           = [opnsense_wireguard_peer.client1.id]
}
```

**Arguments:**
- `name` (Required) - Server instance name
- `enabled` (Optional) - Enable the server. Default: `true`
- `listen_port` (Required) - UDP listen port
- `tunnel_address` (Required) - Tunnel IP in CIDR notation
- `private_key` (Optional) - Private key (auto-generated if not provided)
- `peers` (Optional) - List of peer UUIDs
- `disable_routes` (Optional) - Disable automatic routes

**Attributes:**
- `id` - Server UUID
- `public_key` - Server public key

### opnsense_wireguard_peer

Manages WireGuard peers.

```hcl
resource "opnsense_wireguard_peer" "laptop" {
  name          = "laptop"
  enabled       = true
  public_key    = "your-public-key"
  allowed_ips   = "10.20.30.10/32"
  keepalive     = 25
}
```

**Arguments:**
- `name` (Required) - Peer name
- `enabled` (Optional) - Enable the peer. Default: `true`
- `public_key` (Required) - Peer's public key
- `allowed_ips` (Required) - Allowed IP addresses (comma-separated)
- `endpoint` (Optional) - Endpoint hostname/IP
- `endpoint_port` (Optional) - Endpoint port
- `preshared_key` (Optional) - Pre-shared key
- `keepalive` (Optional) - Persistent keepalive interval (seconds)

**Attributes:**
- `id` - Peer UUID

## Data Sources

### opnsense_firewall_rule

Fetches information about an existing firewall rule.

```hcl
data "opnsense_firewall_rule" "existing" {
  id = "rule-uuid-here"
}
```

## Complete Example

```hcl
terraform {
  required_providers {
    opnsense = {
      source = "yourusername/opnsense"
    }
  }
}

provider "opnsense" {
  host       = "https://192.168.1.1"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
  insecure   = true
}

# Create firewall alias
resource "opnsense_firewall_alias" "web_servers" {
  name        = "web_servers"
  type        = "host"
  content     = ["192.168.1.10", "192.168.1.11"]
  description = "Web server pool"
}

# Create firewall rule
resource "opnsense_firewall_rule" "allow_web" {
  description      = "Allow HTTPS to web servers"
  interface        = "wan"
  protocol         = "tcp"
  source_net       = "any"
  destination_net  = opnsense_firewall_alias.web_servers.name
  destination_port = "443"
  action           = "pass"
  enabled          = true
  log              = true
}

# Create DHCP subnet
resource "opnsense_kea_subnet" "lan" {
  subnet      = "192.168.1.0/24"
  pools       = "192.168.1.100-192.168.1.200"
  description = "Main LAN"
}

# Create DHCP reservation
resource "opnsense_kea_reservation" "server" {
  subnet      = opnsense_kea_subnet.lan.id
  ip_address  = "192.168.1.10"
  hw_address  = "00:11:22:33:44:55"
  hostname    = "webserver1"
}

# Create WireGuard VPN
resource "opnsense_wireguard_server" "vpn" {
  name           = "wg0"
  enabled        = true
  listen_port    = 51820
  tunnel_address = "10.20.30.1/24"
}

resource "opnsense_wireguard_peer" "remote_user" {
  name        = "remote-user"
  enabled     = true
  public_key  = "your-public-key-here"
  allowed_ips = "10.20.30.10/32"
  keepalive   = 25
}
```

## API Endpoints Reference

This provider uses the following OPNsense API endpoints:

### Firewall
- `/api/firewall/filter/add_rule` - Create rule
- `/api/firewall/filter/get-rule/{uuid}` - Get rule
- `/api/firewall/filter/set_rule/{uuid}` - Update rule
- `/api/firewall/filter/del_rule/{uuid}` - Delete rule
- `/api/firewall/filter/apply` - Apply changes
- `/api/firewall/alias/add_item` - Create alias
- `/api/firewall/alias/set_item/{uuid}` - Update alias
- `/api/firewall/alias/del_item/{uuid}` - Delete alias
- `/api/firewall/alias/reconfigure` - Apply alias changes

### Kea DHCP
- `/api/kea/dhcpv4/add_subnet` - Create subnet
- `/api/kea/dhcpv4/set_subnet/{uuid}` - Update subnet
- `/api/kea/dhcpv4/del_subnet/{uuid}` - Delete subnet
- `/api/kea/dhcpv4/add_reservation` - Create reservation
- `/api/kea/dhcpv4/set_reservation/{uuid}` - Update reservation
- `/api/kea/dhcpv4/del_reservation/{uuid}` - Delete reservation
- `/api/kea/service/reconfigure` - Apply Kea changes

### WireGuard
- `/api/wireguard/server/add_server` - Create server
- `/api/wireguard/server/set_server/{uuid}` - Update server
- `/api/wireguard/server/del_server/{uuid}` - Delete server
- `/api/wireguard/client/add_client` - Create peer
- `/api/wireguard/client/set_client/{uuid}` - Update peer
- `/api/wireguard/client/del_client/{uuid}` - Delete peer
- `/api/wireguard/service/reconfigure` - Apply WireGuard changes

## Testing

Run the acceptance tests:

```bash
TF_ACC=1 go test ./... -v -timeout 120m
```

## Known Limitations

1. The provider currently supports OPNsense 26.1 API endpoints
2. Some advanced firewall rule options may not be implemented yet
3. IPv6 support is included but not extensively tested
4. NAT rules are not yet implemented (planned for future release)

## Troubleshooting

### Certificate Errors

If you're using self-signed certificates and getting SSL errors, set `insecure = true` in the provider configuration:

```hcl
provider "opnsense" {
  # ...
  insecure = true
}
```

### API Authentication Errors

Ensure your API key has the proper permissions. Check the user's "Effective Privileges" in the OPNsense web interface.

### Firewall Rule Not Applied

The provider automatically calls the `apply` endpoint after creating/updating/deleting rules. If changes don't appear, check the OPNsense logs.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This provider is released under the Mozilla Public License 2.0. See LICENSE for details.

## Authors

Created for OPNsense 26.1 API compatibility.

## Acknowledgments

- HashiCorp for the Terraform Plugin Framework
- OPNsense team for the excellent firewall platform and API documentation
