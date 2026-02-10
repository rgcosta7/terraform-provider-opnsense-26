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
  insecure   = true  # Set to false in production with valid certs
}

# Alternatively, use environment variables:
# export OPNSENSE_HOST="https://192.168.1.1"
# export OPNSENSE_API_KEY="your-key"
# export OPNSENSE_API_SECRET="your-secret"