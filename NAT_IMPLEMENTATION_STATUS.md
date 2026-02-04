# Destination NAT - Implementation Status

## 🎯 Your Request

You want to forward port 443 → Traefik IP using Terraform.

## ⏳ Current Status

The `opnsense_nat_destination` resource **is not yet implemented**. We need to create it first before you can use it in Terraform.

## 🔧 What's Needed

Based on your API documentation screenshot, we need to create:

1. **Resource file**: `resource_nat_destination.go`
2. **Register it** in `provider.go`
3. **Test it** with your OPNsense

## ⚡ Immediate Options

While I implement the NAT resource, you have 3 options:

### Option 1: Use OPNsense GUI (Quickest)
- Go to **Firewall → NAT → Destination NAT**
- Add rule manually:
  - Interface: WAN
  - Protocol: TCP
  - Destination Port: 443
  - Target IP: 192.168.1.100 (Traefik)
  - Target Port: 443

### Option 2: Use curl/API directly
```bash
curl -k -u "$API_KEY:$API_SECRET" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "rule": {
      "interface": "wan",
      "protocol": "tcp",
      "destination_port": "443",
      "target": "192.168.1.100",
      "target_port": "443",
      "description": "HTTPS to Traefik",
      "enabled": "1"
    }
  }' \
  "https://10.0.10.10/api/firewall/d_nat/add_rule"
```

### Option 3: Wait for Implementation
I can implement the NAT resource, but it will take:
- Creating the resource file (~300 lines)
- Testing with actual API
- Debugging any issues
- Estimated time: 30-60 minutes

## 🧪 What You CAN Test Now

The following resources ARE implemented and ready to test:

1. **Firewall Aliases** ✅ - Already working!
2. **Firewall Rules** ✅ - Ready to test
3. **Kea DHCP** ✅ - Fixed, ready to test
4. **WireGuard VPN** ✅ - Fixed, ready to test

## 📋 Implementation Plan for NAT

When I implement it, the resource will look like:

```hcl
resource "opnsense_nat_destination" "traefik_https" {
  interface        = "wan"
  protocol         = "tcp"
  source_net       = "any"  # optional
  destination      = "wan_address"
  destination_port = "443"
  target_ip        = "192.168.1.100"
  target_port      = "443"
  description      = "HTTPS to Traefik"
  enabled          = true
  log              = false  # optional
}
```

## 🎯 Next Steps

**Choose one:**

**A. Test existing resources now**
- Try firewall rules, Kea DHCP, or WireGuard
- Report back what works
- I can fix any issues

**B. Wait for NAT implementation**
- I implement `opnsense_nat_destination`
- You test it
- We iterate until it works

**C. Use GUI/API for now**
- Create NAT rules manually in OPNsense
- Use Terraform for everything else
- Add NAT support later

## 💡 My Recommendation

1. **Test the 4 existing resources** - They're ready!
2. **Use OPNsense GUI for NAT** - It's quick
3. **Let me implement NAT resource** - For future automation

This way you can start using Terraform NOW for firewall rules, aliases, DHCP, and VPN, while I build out the NAT support.

**What would you prefer?**
