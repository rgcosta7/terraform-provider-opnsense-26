# Firewall Category Example
#
# This example demonstrates creating and using firewall categories
# to organize firewall rules in OPNsense 26.1

terraform {
  required_providers {
    opnsense = {
      source  = "rgcosta7/opnsense"
      version = "~> 0.1"
    }
  }
}

provider "opnsense" {
  host       = var.opnsense_host
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
  insecure   = true
}

# Variables
variable "opnsense_host" {
  type        = string
  description = "OPNsense host URL"
}

variable "opnsense_api_key" {
  type        = string
  description = "OPNsense API key"
  sensitive   = true
}

variable "opnsense_api_secret" {
  type        = string
  description = "OPNsense API secret"
  sensitive   = true
}

# Create categories by action type
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"  # Green
}

resource "opnsense_firewall_category" "block" {
  name  = "Block"
  color = "#FF0000"  # Red
}

resource "opnsense_firewall_category" "reject" {
  name  = "Reject"
  color = "#FFA500"  # Orange
}

# Create categories by service type
resource "opnsense_firewall_category" "web" {
  name  = "Web Services"
  color = "#0000FF"  # Blue
}

resource "opnsense_firewall_category" "database" {
  name  = "Database"
  color = "#800080"  # Purple
}

resource "opnsense_firewall_category" "vpn" {
  name  = "VPN"
  color = "#008000"  # Dark Green
}

resource "opnsense_firewall_category" "management" {
  name  = "Management"
  color = "#FFD700"  # Gold
}

# Create categories by environment
resource "opnsense_firewall_category" "production" {
  name  = "Production"
  color = "#DC143C"  # Crimson
}

resource "opnsense_firewall_category" "staging" {
  name  = "Staging"
  color = "#FF8C00"  # Dark Orange
}

resource "opnsense_firewall_category" "development" {
  name  = "Development"
  color = "#32CD32"  # Lime Green
}

# Temporary category with auto-cleanup
resource "opnsense_firewall_category" "temporary" {
  name  = "Temporary Rules"
  color = "#808080"  # Gray
  auto  = true       # Auto-delete when unused
}

# Output category IDs for use in firewall rules
output "category_ids" {
  value = {
    allow       = opnsense_firewall_category.allow.id
    block       = opnsense_firewall_category.block.id
    reject      = opnsense_firewall_category.reject.id
    web         = opnsense_firewall_category.web.id
    database    = opnsense_firewall_category.database.id
    vpn         = opnsense_firewall_category.vpn.id
    management  = opnsense_firewall_category.management.id
    production  = opnsense_firewall_category.production.id
    staging     = opnsense_firewall_category.staging.id
    development = opnsense_firewall_category.development.id
    temporary   = opnsense_firewall_category.temporary.id
  }
  description = "UUIDs of created firewall categories"
}

# Example: Create firewall rule (category assignment would be added when supported)
resource "opnsense_firewall_rule" "ssh_allow" {
  enabled     = true
  description = "Allow SSH from management network"
  action      = "pass"
  interface   = "lan"
  protocol    = "tcp"
  
  source_net       = "192.168.10.0/24"
  destination_port = "22"
  
  log = true
  
  # Categories will be supported in future version:
  # categories = [
  #   opnsense_firewall_category.allow.id,
  #   opnsense_firewall_category.management.id
  # ]
}
