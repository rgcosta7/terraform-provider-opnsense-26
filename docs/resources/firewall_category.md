# opnsense_firewall_category

Manages firewall categories in OPNsense 26.1.

Categories allow you to organize and filter firewall rules. Each rule can be assigned one or more categories for easier management of large rulesets.

## Example Usage

### Basic Category

```hcl
resource "opnsense_firewall_category" "allow" {
  name = "Allow"
}
```

### Category with Color

```hcl
resource "opnsense_firewall_category" "general" {
  name  = "General"
  color = "#FF0000"  # Red
}
```

### Auto-cleanup Category

```hcl
resource "opnsense_firewall_category" "temporary" {
  name = "Temporary Rules"
  auto = true  # Automatically delete when no rules use it
}
```

### Complete Example with Firewall Rules

```hcl
# Create categories
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"
}

resource "opnsense_firewall_category" "block" {
  name  = "Block"
  color = "#FF0000"
}

resource "opnsense_firewall_category" "management" {
  name  = "Management"
  color = "#0000FF"
}

# Use categories in firewall rules
resource "opnsense_firewall_rule" "ssh_allow" {
  enabled     = true
  description = "Allow SSH from management network"
  action      = "pass"
  interface   = "lan"
  protocol    = "tcp"
  
  source_net       = "192.168.10.0/24"
  destination_port = "22"
  
  # Reference categories (when category support is added to rules)
  # categories = [
  #   opnsense_firewall_category.allow.id,
  #   opnsense_firewall_category.management.id
  # ]
}
```

## Argument Reference

* `name` - (Required) The name of the category.
* `color` - (Optional) Category color in hex format (e.g., `#FF0000` for red).
* `auto` - (Optional) When set to `true`, the category will be automatically deleted when no rules reference it. Defaults to `false`.

## Attribute Reference

* `id` - The UUID of the category.

## Import

Firewall categories can be imported using their UUID:

```bash
terraform import opnsense_firewall_category.allow 12345678-1234-1234-1234-123456789012
```

## Notes

### Color Format

Colors must be in hexadecimal format with a leading `#`:
- Valid: `#FF0000`, `#00FF00`, `#0000FF`
- Invalid: `FF0000`, `red`, `#F00`

### Auto-cleanup

When `auto = true`, the category will be automatically removed by OPNsense when:
- No firewall rules reference it
- No other resources use it

This is useful for temporary categorization or dynamic rule management.

### Using with Firewall Rules

Once categories are created, you can reference them in firewall rules to organize your ruleset. Categories appear in the OPNsense GUI and can be used to filter the rule list.

## Common Use Cases

### Organizing by Action

```hcl
resource "opnsense_firewall_category" "allow" {
  name  = "Allow"
  color = "#00FF00"
}

resource "opnsense_firewall_category" "block" {
  name  = "Block"
  color = "#FF0000"
}

resource "opnsense_firewall_category" "reject" {
  name  = "Reject"
  color = "#FFA500"
}
```

### Organizing by Purpose

```hcl
resource "opnsense_firewall_category" "web" {
  name  = "Web Services"
  color = "#0000FF"
}

resource "opnsense_firewall_category" "database" {
  name  = "Database Access"
  color = "#800080"
}

resource "opnsense_firewall_category" "vpn" {
  name  = "VPN"
  color = "#008000"
}
```

### Organizing by Environment

```hcl
resource "opnsense_firewall_category" "production" {
  name  = "Production"
  color = "#FF0000"
}

resource "opnsense_firewall_category" "staging" {
  name  = "Staging"
  color = "#FFA500"
}

resource "opnsense_firewall_category" "development" {
  name  = "Development"
  color = "#00FF00"
}
```
