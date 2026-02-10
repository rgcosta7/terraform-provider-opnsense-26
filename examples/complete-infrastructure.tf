# Complete OPNsense Infrastructure Example

terraform {
  required_version = ">= 1.0"
  required_providers {
    opnsense = {
      source  = "yourusername/opnsense"
      version = "~> 1.0"
    }
  }
}

provider "opnsense" {
  host       = var.opnsense_host
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
  insecure   = var.opnsense_insecure
}

# ========================================
# Variables
# ========================================

variable "opnsense_host" {
  description = "OPNsense host URL"
  type        = string
  default     = "https://192.168.1.1"
}

variable "opnsense_api_key" {
  description = "OPNsense API key"
  type        = string
  sensitive   = true
}

variable "opnsense_api_secret" {
  description = "OPNsense API secret"
  type        = string
  sensitive   = true
}

variable "opnsense_insecure" {
  description = "Skip TLS verification"
  type        = bool
  default     = true
}

variable "lan_network" {
  description = "LAN network CIDR"
  type        = string
  default     = "192.168.1.0/24"
}

variable "dmz_network" {
  description = "DMZ network CIDR"
  type        = string
  default     = "192.168.100.0/24"
}

variable "vpn_network" {
  description = "VPN tunnel network CIDR"
  type        = string
  default     = "10.20.30.0/24"
}

# ========================================
# Firewall Categories
# ========================================

resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"
}

resource "opnsense_firewall_category" "block" {
  name  = "Block"
  color = "#FF0000"
}

resource "opnsense_firewall_category" "public_services" {
  name  = "Public Services"
  color = "#0000FF"
}

resource "opnsense_firewall_category" "vpn" {
  name  = "VPN"
  color = "#800080"
}

resource "opnsense_firewall_category" "management" {
  name  = "Management"
  color = "#FFA500"
}

resource "opnsense_firewall_category" "application" {
  name  = "Application"
  color = "#00FFFF"
}

# ========================================
# Firewall Aliases
# ========================================

# Internal networks
resource "opnsense_firewall_alias" "internal_networks" {
  name    = "INTERNAL_NETWORKS"
  type    = "network"
  content = [var.lan_network, var.dmz_network]
  description = "All internal networks"
  enabled = true
}

# Web servers in DMZ
resource "opnsense_firewall_alias" "dmz_web_servers" {
  name    = "DMZ_WEB_SERVERS"
  type    = "host"
  content = ["192.168.100.10", "192.168.100.11", "192.168.100.12"]
  description = "Web servers in DMZ"
  enabled = true
}

# Database servers
resource "opnsense_firewall_alias" "database_servers" {
  name    = "DATABASE_SERVERS"
  type    = "host"
  content = ["192.168.1.20", "192.168.1.21"]
  description = "Database servers"
  enabled = true
}

# Management hosts
resource "opnsense_firewall_alias" "management_hosts" {
  name    = "MANAGEMENT_HOSTS"
  type    = "host"
  content = ["192.168.1.5", "192.168.1.6"]
  description = "Management workstations"
  enabled = true
}

# Common service ports
resource "opnsense_firewall_alias" "web_ports" {
  name    = "WEB_PORTS"
  type    = "port"
  content = ["80", "443", "8080", "8443"]
  description = "Common web service ports"
  enabled = true
}

# ========================================
# Firewall Rules - WAN Interface
# ========================================

# Allow HTTPS to DMZ web servers
resource "opnsense_firewall_rule" "wan_to_dmz_https" {
  enabled     = true
  sequence    = 100
  description = "WAN: Allow HTTPS to DMZ web servers"
  
  interface   = "wan"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "tcp"
  
  source_net       = "any"
  destination_net  = "_DMZ_WEB_SERVERS"  # Use underscore prefix!
  destination_port = "443"
  
  action = "pass"
  log    = true
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.public_services.id,
  ]
}

# Allow WireGuard VPN
resource "opnsense_firewall_rule" "wan_wireguard" {
  enabled     = true
  sequence    = 110
  description = "WAN: Allow WireGuard VPN"
  
  interface   = "wan"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "udp"
  
  source_net       = "any"
  destination_net  = "wanip"  # WAN address
  destination_port = "51820"
  
  action = "pass"
  log    = true
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.vpn.id,
  ]
}

# ========================================
# Firewall Rules - LAN Interface
# ========================================

# Allow LAN to internet
resource "opnsense_firewall_rule" "lan_to_internet" {
  enabled     = true
  sequence    = 200
  description = "LAN: Allow to internet"
  
  interface   = "lan"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "any"
  
  source_net      = "lan"
  destination_net = "any"
  
  action = "pass"
  log    = false
  
  categories = [
    opnsense_firewall_category.allow.id,
  ]
}

# Allow management to SSH OPNsense
resource "opnsense_firewall_rule" "mgmt_to_ssh" {
  enabled     = true
  sequence    = 210
  description = "LAN: Allow management SSH to OPNsense"
  
  interface   = "lan"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "tcp"
  
  source_net       = "_MANAGEMENT_HOSTS"  # Use underscore prefix!
  destination_net  = "lanip"  # Firewall IP
  destination_port = "22"
  
  action = "pass"
  log    = true
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.management.id,
  ]
}

# Allow management to web interface
resource "opnsense_firewall_rule" "mgmt_to_webgui" {
  enabled     = true
  sequence    = 220
  description = "LAN: Allow management to web interface"
  
  interface   = "lan"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "tcp"
  
  source_net       = "_MANAGEMENT_HOSTS"  # Use underscore prefix!
  destination_net  = "lanip"  # Firewall IP
  destination_port = "443"
  
  action = "pass"
  log    = true
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.management.id,
  ]
}

# ========================================
# Firewall Rules - DMZ Interface
# ========================================

# Allow DMZ web servers to database
resource "opnsense_firewall_rule" "dmz_to_database" {
  enabled     = true
  sequence    = 300
  description = "DMZ: Allow web to database"
  
  interface   = "dmz"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "tcp"
  
  source_net       = "_DMZ_WEB_SERVERS"  # Use underscore prefix!
  destination_net  = "_DATABASE_SERVERS"  # Use underscore prefix!
  destination_port = "3306"
  
  action = "pass"
  log    = true
  
  categories = [
    opnsense_firewall_category.allow.id,
    opnsense_firewall_category.application.id,
  ]
}

# Block DMZ to LAN (except database) - This rule should come AFTER specific allows
resource "opnsense_firewall_rule" "block_dmz_to_lan" {
  enabled     = true
  sequence    = 999  # Low priority - after specific allows
  description = "DMZ: Block to LAN"
  
  interface   = "dmz"
  direction   = "in"
  ip_protocol = "inet"
  protocol    = "any"
  
  source_net      = "dmz"
  destination_net = "lan"
  
  action = "block"
  log    = true
  
  categories = [
    opnsense_firewall_category.block.id,
  ]
}

# ========================================
# Kea DHCP Configuration
# ========================================

# LAN DHCP subnet
resource "opnsense_kea_subnet" "lan_dhcp" {
  subnet       = var.lan_network
  pools        = "192.168.1.100-192.168.1.200"
  description  = "LAN DHCP subnet"
  auto_collect = false
  
  option_data = {
    routers             = "192.168.1.1"
    domain-name-servers = "192.168.1.1,8.8.8.8"
    domain-name         = "lan.local"
    ntp-servers         = "192.168.1.1"
  }
}

# DMZ DHCP subnet
resource "opnsense_kea_subnet" "dmz_dhcp" {
  subnet       = var.dmz_network
  pools        = "192.168.100.50-192.168.100.100"
  description  = "DMZ DHCP subnet"
  auto_collect = false
  
  option_data = {
    routers             = "192.168.100.1"
    domain-name-servers = "192.168.1.1,8.8.8.8"
    domain-name         = "dmz.local"
  }
}

# ========================================
# DHCP Reservations
# ========================================

# Static reservations for management hosts
resource "opnsense_kea_reservation" "mgmt_host1" {
  subnet      = opnsense_kea_subnet.lan_dhcp.id
  ip_address  = "192.168.1.5"
  hw_address  = "00:11:22:33:44:55"
  hostname    = "mgmt-workstation-1"
  description = "Management workstation 1"
}

resource "opnsense_kea_reservation" "mgmt_host2" {
  subnet      = opnsense_kea_subnet.lan_dhcp.id
  ip_address  = "192.168.1.6"
  hw_address  = "00:11:22:33:44:66"
  hostname    = "mgmt-workstation-2"
  description = "Management workstation 2"
}

# Database server reservations
resource "opnsense_kea_reservation" "db_primary" {
  subnet      = opnsense_kea_subnet.lan_dhcp.id
  ip_address  = "192.168.1.20"
  hw_address  = "aa:bb:cc:dd:ee:01"
  hostname    = "db-primary"
  description = "Primary database server"
}

resource "opnsense_kea_reservation" "db_secondary" {
  subnet      = opnsense_kea_subnet.lan_dhcp.id
  ip_address  = "192.168.1.21"
  hw_address  = "aa:bb:cc:dd:ee:02"
  hostname    = "db-secondary"
  description = "Secondary database server"
}

# DMZ web server reservations
resource "opnsense_kea_reservation" "web1" {
  subnet      = opnsense_kea_subnet.dmz_dhcp.id
  ip_address  = "192.168.100.10"
  hw_address  = "bb:cc:dd:ee:ff:01"
  hostname    = "web-server-1"
  description = "Web server 1"
}

resource "opnsense_kea_reservation" "web2" {
  subnet      = opnsense_kea_subnet.dmz_dhcp.id
  ip_address  = "192.168.100.11"
  hw_address  = "bb:cc:dd:ee:ff:02"
  hostname    = "web-server-2"
  description = "Web server 2"
}

resource "opnsense_kea_reservation" "web3" {
  subnet      = opnsense_kea_subnet.dmz_dhcp.id
  ip_address  = "192.168.100.12"
  hw_address  = "bb:cc:dd:ee:ff:03"
  hostname    = "web-server-3"
  description = "Web server 3"
}

# ========================================
# WireGuard VPN Configuration
# ========================================

# Note: WireGuard resources not implemented in your provider yet
# These are placeholder examples

# WireGuard server (if implemented)
# resource "opnsense_wireguard_server" "main_vpn" {
#   name           = "wg0"
#   enabled        = true
#   listen_port    = 51820
#   tunnel_address = "10.20.30.1/24"
# }

# Remote worker peers (if implemented)
# resource "opnsense_wireguard_peer" "remote_worker_1" {
#   name        = "remote-worker-1-laptop"
#   enabled     = true
#   public_key  = "worker1-public-key-replace-me"
#   allowed_ips = "10.20.30.10/32"
#   keepalive   = 25
# }

# ========================================
# NAT Rules (Port Forwarding)
# ========================================

# Port forward HTTPS to DMZ web server 1
# resource "opnsense_nat_destination" "web1_https" {
#   interface        = "wan"
#   protocol         = "tcp"
#   source           = "any"
#   destination      = "wanip"
#   destination_port = "443"
#   target           = "192.168.100.10"
#   local_port       = "443"
#   description      = "HTTPS to web server 1"
# }

# ========================================
# Outputs
# ========================================

output "firewall_categories" {
  description = "Created firewall categories"
  value = {
    allow           = opnsense_firewall_category.allow.id
    block           = opnsense_firewall_category.block.id
    public_services = opnsense_firewall_category.public_services.id
    vpn             = opnsense_firewall_category.vpn.id
    management      = opnsense_firewall_category.management.id
    application     = opnsense_firewall_category.application.id
  }
}

output "firewall_aliases" {
  description = "Created firewall aliases"
  value = {
    internal_networks = opnsense_firewall_alias.internal_networks.id
    dmz_web_servers   = opnsense_firewall_alias.dmz_web_servers.id
    database_servers  = opnsense_firewall_alias.database_servers.id
    management_hosts  = opnsense_firewall_alias.management_hosts.id
    web_ports         = opnsense_firewall_alias.web_ports.id
  }
}

output "dhcp_subnets" {
  description = "DHCP subnet IDs"
  value = {
    lan = opnsense_kea_subnet.lan_dhcp.id
    dmz = opnsense_kea_subnet.dmz_dhcp.id
  }
}

output "summary" {
  description = "Infrastructure summary"
  value = {
    categories_created   = 6
    aliases_created      = 5
    firewall_rules       = 7
    dhcp_subnets         = 2
    dhcp_reservations    = 7
    management_ips       = ["192.168.1.5", "192.168.1.6"]
    database_ips         = ["192.168.1.20", "192.168.1.21"]
    web_server_ips       = ["192.168.100.10", "192.168.100.11", "192.168.100.12"]
  }
}
