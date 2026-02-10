# Complete OPNsense Terraform Provider - All Resources Fields Reference

## Table of Contents
1. [opnsense_firewall_alias](#opnsense_firewall_alias)
2. [opnsense_firewall_category](#opnsense_firewall_category)
3. [opnsense_firewall_rule](#opnsense_firewall_rule)
4. [opnsense_kea_subnet](#opnsense_kea_subnet)
5. [opnsense_kea_reservation](#opnsense_kea_reservation)
6. [opnsense_nat_destination](#opnsense_nat_destination)
7. [opnsense_wireguard_server](#opnsense_wireguard_server)
8. [opnsense_wireguard_peer](#opnsense_wireguard_peer)

---

## opnsense_firewall_alias

Create aliases (groups) of IPs, networks, ports, URLs, etc.

### Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Alias UUID | Auto-generated |
| `name` | string | ✅ Required | Alias name | `"DNS_SERVERS"` |
| `type` | string | ✅ Required | Alias type | `"host"`, `"network"`, `"port"`, `"url"`, `"urltable"`, `"geoip"`, `"mac"` |
| `content` | list(string) | ✅ Required | List of entries | `["10.0.20.11", "10.0.20.22"]` |
| `description` | string | Optional | Description | `"Primary DNS servers"` |
| `enabled` | bool | Optional | Enable alias | `true` (default) |

### Alias Types

- **host** - Single IP addresses
- **network** - Network ranges (CIDR)
- **port** - Port numbers or ranges
- **url** - URLs for dynamic lists
- **urltable** - URL tables
- **geoip** - Geographic IP blocks
- **mac** - MAC addresses
- **networkgroup** - Group of networks
- **external** - External sources

### Complete Example

```hcl
# Host alias (IP addresses)
resource "opnsense_firewall_alias" "dns_servers" {
  name        = "DNS_SERVERS"
  type        = "host"
  content     = [
    "10.0.20.11",
    "10.0.20.22",
    "8.8.8.8"
  ]
  description = "DNS server list"
  enabled     = true
}

# Network alias (CIDR blocks)
resource "opnsense_firewall_alias" "private_networks" {
  name    = "PRIVATE_NETWORKS"
  type    = "network"
  content = [
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16"
  ]
  description = "RFC1918 private networks"
}

# Port alias
resource "opnsense_firewall_alias" "web_ports" {
  name    = "WEB_PORTS"
  type    = "port"
  content = [
    "80",
    "443",
    "8080",
    "8443"
  ]
  description = "Common web ports"
}

# MAC address alias
resource "opnsense_firewall_alias" "iot_devices" {
  name    = "IOT_DEVICES"
  type    = "mac"
  content = [
    "aa:bb:cc:dd:ee:01",
    "aa:bb:cc:dd:ee:02"
  ]
  description = "IoT device MAC addresses"
}
```

### Usage in Firewall Rules

```hcl
resource "opnsense_firewall_rule" "allow_dns" {
  description     = "Allow DNS queries"
  source_net      = "lan"
  destination_net = opnsense_firewall_alias.dns_servers.name  # Use alias name
  destination_port = "53"
  protocol        = "udp"
}

# Or reference with underscore prefix (OPNsense convention)
resource "opnsense_firewall_rule" "block_private" {
  description     = "Block private networks"
  destination_net = "_PRIVATE_NETWORKS"  # Underscore prefix
}
```

---

## opnsense_firewall_category

Create categories for organizing firewall rules.

### Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Category UUID | Auto-generated |
| `name` | string | ✅ Required | Category name | `"Allow"` |
| `color` | string | Optional | Hex color code | `"#00FF00"` |
| `auto` | bool | Optional | Auto-delete when unused | `false` (default) |

### Complete Example

```hcl
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"  # Green
  auto  = false
}

resource "opnsense_firewall_category" "block" {
  name  = "Block"
  color = "#FF0000"  # Red
}

resource "opnsense_firewall_category" "iot" {
  name  = "IoT"
  color = "#800080"  # Purple
}

resource "opnsense_firewall_category" "services" {
  name  = "Services"
  color = "#FFA500"  # Orange
}

resource "opnsense_firewall_category" "wireguard" {
  name  = "WireGuard"
  color = "#0000FF"  # Blue
}
```

### Common Color Schemes

```hcl
# Traffic Control
Allow     = "#00FF00"  # Green
Block     = "#FF0000"  # Red
Reject    = "#FFA500"  # Orange

# Network Types
IoT       = "#800080"  # Purple
Guest     = "#FFFF00"  # Yellow
Admin     = "#00FFFF"  # Cyan
Home      = "#0000FF"  # Blue

# Services
DNS       = "#008000"  # Dark Green
Web       = "#4169E1"  # Royal Blue
VPN       = "#8B4513"  # Saddle Brown
```

### Usage in Rules

```hcl
resource "opnsense_firewall_rule" "iot_dns" {
  description = "IoT DNS access"
  # ... rule config ...
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.iot.id,
  ]
}
```

---

## opnsense_firewall_rule

Create firewall rules to control network traffic.

### Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Rule UUID | Auto-generated |
| `enabled` | bool | Optional | Enable rule | `true` (default) |
| `sequence` | int | Optional | Rule order (lower = first) | `500` |
| `description` | string | ✅ Required | Rule description | `"Allow DNS"` |
| `interface` | string | Optional | Interface name | `"lan"`, `"wan"`, `"opt1"` |
| `direction` | string | Optional | Traffic direction | `"in"` (default), `"out"` |
| `ip_protocol` | string | Optional | IP version | `"inet"` (IPv4, default), `"inet6"` (IPv6) |
| `protocol` | string | ✅ Required | Protocol | `"tcp"`, `"udp"`, `"icmp"`, `"any"` |
| `source_net` | string | ✅ Required | Source network/IP | `"lan"`, `"10.0.1.0/24"`, `"_ALIAS"` |
| `source_port` | string | Optional | Source port | `"any"`, `"80"`, `"1024-65535"` |
| `source_not` | bool | Optional | Invert source match | `false` (default) |
| `destination_net` | string | ✅ Required | Destination network/IP | `"any"`, `"192.168.1.0/24"` |
| `destination_port` | string | Optional | Destination port | `"443"`, `"8080-8090"` |
| `destination_not` | bool | Optional | Invert destination | `false` (default) |
| `invert` | bool | Optional | Alias for destination_not | `false` (default) |
| `action` | string | Optional | Rule action | `"pass"` (default), `"block"`, `"reject"` |
| `quick` | bool | Optional | Stop processing on match | `false` (default) |
| `log` | bool | Optional | Log matching packets | `false` (default) |
| `gateway` | string | Optional | Route via gateway | `"WAN_DHCP"`, `"BLUEDRAGON"` |
| `categories` | list(string) | Optional | Category UUIDs | `[category.allow.id]` |

### Complete Example

```hcl
resource "opnsense_firewall_rule" "complete_example" {
  # Organisation
  enabled     = true
  sequence    = 500
  description = "IoT internet access via VPN"
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.iot.id,
  ]
  
  # Interface
  interface = "opt4"
  
  # Filter - Basic
  action      = "pass"
  quick       = true
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "any"
  
  # Filter - Source
  source_not  = false
  source_net  = "_IOT_DEVICES"
  source_port = "any"
  
  # Filter - Destination
  destination_not = true  # NOT destination (allow internet, block private)
  destination_net = "_PRIVATE_NETWORKS"
  
  # Logging
  log = true
  
  # Routing
  gateway = "BLUEDRAGON"  # Route via VPN
}
```

### Common Patterns

```hcl
# Allow LAN to DNS
resource "opnsense_firewall_rule" "lan_dns" {
  description      = "Allow DNS queries"
  interface        = "lan"
  protocol         = "udp"
  source_net       = "lan"
  destination_net  = "_DNS_SERVERS"
  destination_port = "53"
  action           = "pass"
}

# Block IoT to LAN
resource "opnsense_firewall_rule" "iot_block_lan" {
  description     = "Block IoT to LAN"
  interface       = "opt4"
  protocol        = "any"
  source_net      = "opt4"
  destination_net = "lan"
  action          = "block"
  log             = true
}

# Allow to internet only (NOT private networks)
resource "opnsense_firewall_rule" "guest_internet_only" {
  description     = "Guest: internet only"
  interface       = "opt5"
  protocol        = "any"
  source_net      = "opt5"
  destination_not = true
  destination_net = "_PRIVATE_NETWORKS"
  action          = "pass"
}

# IPv4 + IPv6 pair
resource "opnsense_firewall_rule" "allow_web_v4" {
  description     = "Allow web traffic IPv4"
  ip_protocol     = "inet"
  protocol        = "tcp"
  destination_port = "80,443"
  # ...
}

resource "opnsense_firewall_rule" "allow_web_v6" {
  description     = "Allow web traffic IPv6"
  ip_protocol     = "inet6"
  protocol        = "tcp"
  destination_port = "80,443"
  # ...
}
```

---

## opnsense_kea_subnet

Create DHCP subnets with Kea DHCP server.

### Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Subnet UUID | Auto-generated |
| `subnet` | string | ✅ Required | Network CIDR | `"10.0.10.0/26"` |
| `pools` | string | Optional | IP address pools | `"10.0.10.1-10.0.10.5"` |
| `description` | string | Optional | Subnet description | `"Management VLAN"` |
| `auto_collect` | bool | Optional | Auto-collect leases | `false` (default) |
| `option_data` | map(string) | Optional | DHCP options | See below |

### DHCP Option Data Fields

| Option | Description | Example |
|--------|-------------|---------|
| `routers` | Default gateway | `"10.0.10.1"` |
| `domain-name-servers` | DNS servers (comma-separated) | `"10.0.20.11,10.0.20.22"` |
| `domain-name` | DNS domain | `"example.local"` |
| `domain-search` | DNS search domains | `"example.local,corp.local"` |
| `ntp-servers` | NTP servers | `"10.0.10.1"` |
| `time-servers` | Time servers | `"10.0.10.1"` |
| `tftp-server-name` | TFTP server | `"tftp.example.com"` |
| `boot-file-name` | Boot filename | `"pxelinux.0"` |

### Complete Example

```hcl
resource "opnsense_kea_subnet" "vlan10_mgmt" {
  subnet       = "10.0.10.0/26"
  pools        = "10.0.10.1-10.0.10.5"
  description  = "VLAN10 - Management"
  auto_collect = false
  
  option_data = {
    routers             = "10.0.10.10"
    domain-name-servers = "10.0.20.11, 10.0.20.22"
    domain-name         = "mgmt.local"
    ntp-servers         = "10.0.10.10"
  }
}

resource "opnsense_kea_subnet" "vlan20_services" {
  subnet      = "10.0.20.0/24"
  pools       = "10.0.20.100-10.0.20.200"
  description = "VLAN20 - Services"
  
  option_data = {
    routers             = "10.0.20.1"
    domain-name-servers = "10.0.20.11, 10.0.20.22"
    ntp-servers         = "10.0.20.11"
  }
}

# Multiple pools
resource "opnsense_kea_subnet" "vlan30_cluster" {
  subnet      = "10.0.30.0/24"
  pools       = "10.0.30.10-10.0.30.50,10.0.30.100-10.0.30.150"
  description = "VLAN30 - Kubernetes Cluster"
  
  option_data = {
    routers             = "10.0.30.1"
    domain-name-servers = "10.0.30.11"
    domain-name         = "cluster.local"
  }
}

# PXE boot configuration
resource "opnsense_kea_subnet" "vlan_pxe" {
  subnet      = "10.0.50.0/24"
  pools       = "10.0.50.100-10.0.50.200"
  description = "PXE Boot Network"
  
  option_data = {
    routers          = "10.0.50.1"
    tftp-server-name = "10.0.50.10"
    boot-file-name   = "pxelinux.0"
  }
}
```

### Usage with Reservations

```hcl
# First create subnet
resource "opnsense_kea_subnet" "vlan10" {
  subnet = "10.0.10.0/26"
  # ...
}

# Then create reservations referencing the subnet
resource "opnsense_kea_reservation" "server1" {
  subnet     = opnsense_kea_subnet.vlan10.id
  ip_address = "10.0.10.20"
  hw_address = "aa:bb:cc:dd:ee:ff"
  hostname   = "server1"
}
```

---

## opnsense_kea_reservation

Create DHCP reservations (static IP assignments).

### Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Reservation UUID | Auto-generated |
| `subnet` | string | ✅ Required | Subnet UUID | `opnsense_kea_subnet.vlan10.id` |
| `ip_address` | string | ✅ Required | Reserved IP address | `"10.0.10.20"` |
| `hw_address` | string | ✅ Required | MAC address | `"aa:bb:cc:dd:ee:ff"` |
| `hostname` | string | Optional | Hostname | `"server1"` |
| `description` | string | Optional | Description | `"Web server"` |

### Complete Example

```hcl
# Individual reservations
resource "opnsense_kea_reservation" "unifi_controller" {
  subnet      = opnsense_kea_subnet.vlan10.id
  ip_address  = "10.0.10.20"
  hw_address  = "bc:24:11:c5:c2:3b"
  hostname    = "unifi"
  description = "UniFi Network Controller"
}

resource "opnsense_kea_reservation" "dns_primary" {
  subnet      = opnsense_kea_subnet.vlan20.id
  ip_address  = "10.0.20.11"
  hw_address  = "aa:bb:cc:dd:ee:01"
  hostname    = "dns-primary"
  description = "Primary DNS Server"
}

resource "opnsense_kea_reservation" "k8s_master1" {
  subnet      = opnsense_kea_subnet.vlan30.id
  ip_address  = "10.0.30.11"
  hw_address  = "aa:bb:cc:dd:ee:11"
  hostname    = "k8s-master-1"
  description = "Kubernetes master node 1"
}
```

### Using for_each Pattern (Better!)

```hcl
# Define all reservations in a variable
variable "kea_reservations" {
  type = map(object({
    subnet_id   = string
    ip_address  = string
    hw_address  = string
    hostname    = string
    description = optional(string)
  }))
}

# Single resource block for all reservations
resource "opnsense_kea_reservation" "reservations" {
  for_each = var.kea_reservations

  subnet      = each.value.subnet_id
  ip_address  = each.value.ip_address
  hw_address  = each.value.hw_address
  hostname    = each.value.hostname
  description = try(each.value.description, "")
}
```

In terraform.tfvars:
```hcl
kea_reservations = {
  "unifi" = {
    subnet_id   = "subnet-uuid-here"
    ip_address  = "10.0.10.20"
    hw_address  = "bc:24:11:c5:c2:3b"
    hostname    = "unifi"
    description = "UniFi Controller"
  }
  
  "dns-primary" = {
    subnet_id   = "subnet-uuid-here"
    ip_address  = "10.0.20.11"
    hw_address  = "aa:bb:cc:dd:ee:01"
    hostname    = "dns-primary"
  }
  
  # ... hundreds more ...
}
```

---

## opnsense_nat_destination

Create destination NAT rules (port forwarding).

### Fields

*Note: I don't have the schema file for this resource. Here's the typical structure:*

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | NAT rule UUID | Auto-generated |
| `interface` | string | Required | Interface | `"wan"` |
| `protocol` | string | Required | Protocol | `"tcp"`, `"udp"`, `"tcp/udp"` |
| `source` | string | Optional | Source address | `"any"` |
| `destination` | string | Required | Destination (WAN IP) | `"wanip"` |
| `destination_port` | string | Required | External port | `"443"` |
| `target` | string | Required | Internal IP | `"10.0.10.20"` |
| `local_port` | string | Required | Internal port | `"443"` |
| `description` | string | Optional | Description | `"Web server HTTPS"` |

### Example Structure

```hcl
# Port forward for web server
resource "opnsense_nat_destination" "web_https" {
  interface        = "wan"
  protocol         = "tcp"
  source           = "any"
  destination      = "wanip"
  destination_port = "443"
  target           = "10.0.20.80"
  local_port       = "443"
  description      = "HTTPS to web server"
}

# Port forward with different internal/external ports
resource "opnsense_nat_destination" "ssh_alt" {
  interface        = "wan"
  protocol         = "tcp"
  source           = "any"
  destination      = "wanip"
  destination_port = "2222"  # External
  target           = "10.0.10.50"
  local_port       = "22"    # Internal
  description      = "SSH on alternate port"
}
```

---

## opnsense_wireguard_server

Create WireGuard VPN server configuration.

### Fields

*Note: I don't have the schema file. Typical structure:*

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Server UUID | Auto-generated |
| `name` | string | Required | Server name | `"wg0"` |
| `enabled` | bool | Optional | Enable server | `true` |
| `port` | int | Required | Listen port | `51820` |
| `tunnel_address` | string | Required | Tunnel IP/CIDR | `"10.255.0.1/24"` |
| `peers` | list(string) | Optional | Peer UUIDs | `[peer.laptop.id]` |
| `dns` | string | Optional | DNS servers | `"10.0.20.11"` |
| `interface` | string | Optional | Network interface | `"opt7"` |

### Example Structure

```hcl
resource "opnsense_wireguard_server" "main" {
  name            = "wg0"
  enabled         = true
  port            = 51820
  tunnel_address  = "10.255.0.1/24"
  dns             = "10.0.20.11,10.0.20.22"
  interface       = "opt7"
}
```

---

## opnsense_wireguard_peer

Create WireGuard VPN peer (client) configuration.

### Fields

*Note: I don't have the schema file. Typical structure:*

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `id` | string | Computed | Peer UUID | Auto-generated |
| `name` | string | Required | Peer name | `"laptop"` |
| `enabled` | bool | Optional | Enable peer | `true` |
| `public_key` | string | Required | Peer public key | WireGuard public key |
| `allowed_ips` | string | Required | Allowed IPs | `"10.255.0.2/32"` |
| `endpoint` | string | Optional | Peer endpoint | `"peer.example.com:51820"` |
| `preshared_key` | string | Optional | Pre-shared key | WireGuard PSK |
| `keepalive` | int | Optional | Persistent keepalive | `25` |

### Example Structure

```hcl
resource "opnsense_wireguard_peer" "laptop" {
  name        = "laptop"
  enabled     = true
  public_key  = "base64-encoded-public-key-here"
  allowed_ips = "10.255.0.2/32"
  keepalive   = 25
}

resource "opnsense_wireguard_peer" "phone" {
  name        = "phone"
  enabled     = true
  public_key  = "base64-encoded-public-key-here"
  allowed_ips = "10.255.0.3/32"
}

resource "opnsense_wireguard_peer" "remote_site" {
  name        = "branch_office"
  enabled     = true
  public_key  = "base64-encoded-public-key-here"
  allowed_ips = "10.255.0.4/32,192.168.100.0/24"
  endpoint    = "branch.example.com:51820"
  keepalive   = 25
}
```

---

## Summary Table - All Resources

| Resource | Purpose | Key Fields |
|----------|---------|------------|
| `opnsense_firewall_alias` | IP/network/port groups | name, type, content |
| `opnsense_firewall_category` | Rule organization | name, color |
| `opnsense_firewall_rule` | Traffic control | source, destination, action, gateway |
| `opnsense_kea_subnet` | DHCP subnets | subnet, pools, option_data |
| `opnsense_kea_reservation` | Static DHCP | ip_address, hw_address |
| `opnsense_nat_destination` | Port forwarding | destination_port, target, local_port |
| `opnsense_wireguard_server` | VPN server | port, tunnel_address |
| `opnsense_wireguard_peer` | VPN clients | public_key, allowed_ips |

---

## Common Patterns

### 1. Complete Network Setup

```hcl
# Categories
resource "opnsense_firewall_category" "allow" {
  name = "Allow"
  color = "#00FF00"
}

# Aliases
resource "opnsense_firewall_alias" "dns_servers" {
  name = "DNS_SERVERS"
  type = "host"
  content = ["10.0.20.11", "10.0.20.22"]
}

# DHCP Subnet
resource "opnsense_kea_subnet" "mgmt" {
  subnet = "10.0.10.0/24"
  pools  = "10.0.10.100-10.0.10.200"
  option_data = {
    routers = "10.0.10.1"
    domain-name-servers = "10.0.20.11,10.0.20.22"
  }
}

# DHCP Reservation
resource "opnsense_kea_reservation" "server1" {
  subnet     = opnsense_kea_subnet.mgmt.id
  ip_address = "10.0.10.20"
  hw_address = "aa:bb:cc:dd:ee:ff"
  hostname   = "server1"
}

# Firewall Rule
resource "opnsense_firewall_rule" "allow_dns" {
  description      = "Allow DNS"
  interface        = "lan"
  protocol         = "udp"
  source_net       = "lan"
  destination_net  = "_DNS_SERVERS"
  destination_port = "53"
  action           = "pass"
  categories       = [opnsense_firewall_category.allow.id]
}
```

### 2. Modular Organization

```
terraform/
├── categories.tf       # All categories
├── aliases.tf          # All aliases
├── dhcp_subnets.tf     # DHCP subnets
├── dhcp_reservations.tf # DHCP reservations
├── firewall_rules.tf   # Firewall rules
├── nat.tf              # NAT rules
└── wireguard.tf        # VPN config
```

### 3. Using for_each for Scale

Best for managing hundreds of items:

```hcl
# Define in variables
variable "aliases" { ... }
variable "reservations" { ... }
variable "rules" { ... }

# Create with for_each
resource "opnsense_firewall_alias" "aliases" {
  for_each = var.aliases
  # ...
}

resource "opnsense_kea_reservation" "reservations" {
  for_each = var.reservations
  # ...
}
```

---

🎉 **All resources documented with complete field references and examples!**