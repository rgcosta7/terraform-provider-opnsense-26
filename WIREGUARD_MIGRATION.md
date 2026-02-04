# WireGuard Migration Guide: OPNsense 25.10 → 26.1

## ✅ Updated Provider - DNS, MTU, Gateway Added!

The provider now supports **all** the fields you need for WireGuard configuration.

## 🔄 Resource Name Changes

| Old (25.10) | New (26.1) |
|-------------|------------|
| `opnsense_wireguard_client` | `opnsense_wireguard_peer` |
| `opnsense_wireguard_server` | `opnsense_wireguard_server` (same) |

## 📋 Complete Field Mapping

### Peer (Client) Resource

| Old Field | New Field | Type | Notes |
|-----------|-----------|------|-------|
| `tunnel_address` | `allowed_ips` | list → string | Changed type |
| `server_address` | `endpoint` | string | Renamed |
| `server_port` | `endpoint_port` | int | Renamed |
| `public_key` | `public_key` | string | Same |
| `enabled` | `enabled` | bool | Same |
| `name` | `name` | string | Same |

### Server Resource

| Old Field | New Field | Type | Notes |
|-----------|-----------|------|-------|
| `port` | `listen_port` | int | Renamed |
| `tunnel_address` | `tunnel_address` | list → string | Changed type |
| `public_key` | `public_key` | computed | Now auto-generated |
| `private_key` | `private_key` | string | Same |
| `dns` | `dns` | string | ✅ **NOW SUPPORTED** |
| `mtu` | `mtu` | int | ✅ **NOW SUPPORTED** |
| `gateway` | `gateway` | string | ✅ **NOW SUPPORTED** |
| `disable_routes` | `disable_routes` | bool | Same |
| `peers` | `peers` | list | Same |

## ✅ Complete Converted Example

### Before (25.10)

```terraform
resource "opnsense_wireguard_client" "peers" {
  for_each = nonsensitive(var.wg_config.peers)

  enabled    = true
  name       = each.key
  public_key = each.value.key
  tunnel_address = [each.value.ip]

  server_address = each.value.host
  server_port    = each.value.port
}

resource "opnsense_wireguard_server" "wg_server" {
  enabled        = true
  name           = var.wg_config.server_name
  private_key    = var.wg_config.server_priv
  public_key     = var.wg_config.server_pub
  dns            = var.wg_config.dns
  tunnel_address = [var.wg_config.server_ip]
  port           = var.wg_config.port
  disable_routes = true
  mtu            = var.wg_config.mtu
  gateway        = "10.1.1.1"
  
  peers = [for p in opnsense_wireguard_client.peers : p.id]
}
```

### After (26.1) ✅

```terraform
resource "opnsense_wireguard_peer" "peers" {
  for_each = nonsensitive(var.wg_config.peers)

  enabled       = true
  name          = each.key
  public_key    = each.value.key
  allowed_ips   = each.value.ip  # String: "10.1.1.2/32"
  
  endpoint      = each.value.host
  endpoint_port = each.value.port
  
  # Optional fields:
  # keepalive     = 25
  # preshared_key = "..."
}

resource "opnsense_wireguard_server" "wg_server" {
  enabled        = true
  name           = var.wg_config.server_name
  private_key    = var.wg_config.server_priv
  # public_key is auto-generated - don't specify it
  
  listen_port    = var.wg_config.port
  tunnel_address = var.wg_config.server_ip  # String: "10.1.1.1/24"
  disable_routes = true
  
  # ✅ Now supported!
  dns     = var.wg_config.dns
  mtu     = var.wg_config.mtu
  gateway = "10.1.1.1"
  
  peers = [for p in opnsense_wireguard_peer.peers : p.id]
}
```

## 🔧 Step-by-Step Conversion

### 1. Rename Resources

```bash
# Change all occurrences
sed -i 's/opnsense_wireguard_client/opnsense_wireguard_peer/g' *.tf
```

### 2. Update Peer Fields

```terraform
# Old
tunnel_address = [each.value.ip]
server_address = each.value.host
server_port    = each.value.port

# New
allowed_ips   = each.value.ip  # No brackets, string not list
endpoint      = each.value.host
endpoint_port = each.value.port
```

### 3. Update Server Fields

```terraform
# Old
port           = var.wg_config.port
tunnel_address = [var.wg_config.server_ip]
public_key     = var.wg_config.server_pub

# New
listen_port    = var.wg_config.port
tunnel_address = var.wg_config.server_ip  # No brackets
# Remove public_key - it's auto-generated
```

### 4. Keep DNS, MTU, Gateway

```terraform
# These are now supported! Keep them as-is:
dns     = var.wg_config.dns     # ✅ Supported
mtu     = var.wg_config.mtu     # ✅ Supported
gateway = "10.1.1.1"            # ✅ Supported
```

## 📝 Example Variable Structure

```terraform
variable "wg_config" {
  type = object({
    server_name = string
    server_priv = string
    server_ip   = string
    port        = number
    dns         = string
    mtu         = number
    peers = map(object({
      key  = string
      ip   = string
      host = string
      port = number
    }))
  })
  
  default = {
    server_name = "wg0"
    server_priv = "your-private-key"
    server_ip   = "10.1.1.1/24"
    port        = 51820
    dns         = "10.0.20.11,10.0.20.22"
    mtu         = 1420
    peers = {
      "mobile" = {
        key  = "peer-public-key"
        ip   = "10.1.1.2/32"
        host = "vpn.example.com"
        port = 51820
      }
    }
  }
}
```

## 🎯 Summary of Changes

### What You Need to Change:
1. ✅ `opnsense_wireguard_client` → `opnsense_wireguard_peer`
2. ✅ `tunnel_address` → `allowed_ips` (and remove list brackets)
3. ✅ `server_address` → `endpoint`
4. ✅ `server_port` → `endpoint_port`
5. ✅ `port` → `listen_port` (in server)
6. ✅ Remove `public_key` from server (auto-generated)
7. ✅ Remove list brackets from `tunnel_address`

### What You DON'T Need to Change:
- ✅ `dns` - Now works!
- ✅ `mtu` - Now works!
- ✅ `gateway` - Now works!
- ✅ `disable_routes` - Still works
- ✅ `peers` - Still works
- ✅ `enabled` - Still works
- ✅ `name` - Still works
- ✅ `private_key` - Still works

## 🚀 Rebuild & Test

```bash
# Rebuild the provider with new fields
cd terraform-provider-opnsense-26
./clean-build.sh
./install.sh

# Test
terraform init -upgrade
terraform plan
```

That's it! Your WireGuard configuration should now work with all fields supported.
