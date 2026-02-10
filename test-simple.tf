# Test Configuration - Simple Firewall Alias
# This is a minimal test to verify the provider works

terraform {
  required_providers {
    opnsense = {
      source  = "rgcosta7/opnsense"
      version = "0.1.0"
    }
  }
}

# Configure the provider
provider "opnsense" {
  # Replace these with your OPNsense details
  host       = "https://192.168.1.1"  # Your OPNsense IP/hostname
  api_key    = "your-api-key-here"
  api_secret = "your-api-secret-here"
  insecure   = true  # Set to false if you have valid SSL certificates
}

# Create a simple firewall alias for testing
resource "opnsense_firewall_alias" "test_alias" {
  name        = "terraform_test"
  type        = "host"
  content     = ["8.8.8.8", "8.8.4.4"]
  description = "Test alias created by Terraform"
  enabled     = true
}

# Output the created alias ID
output "alias_id" {
  value       = opnsense_firewall_alias.test_alias.id
  description = "The UUID of the created alias"
}

output "alias_name" {
  value       = opnsense_firewall_alias.test_alias.name
  description = "The name of the created alias"
}


