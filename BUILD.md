# Building and Installing the Provider

## Quick Method (Recommended)

### Step 1: Build the Provider

```bash
# Make the build script executable (if not already)
chmod +x build.sh

# Build
./build.sh
```

This will create the `terraform-provider-opnsense` binary.

### Step 2: Install the Provider

```bash
# Make the install script executable (if not already)
chmod +x install.sh

# Install
./install.sh
```

This will copy the provider to your Terraform plugins directory.

---

## Alternative: Using Make

If you prefer using Make:

```bash
make build
make install
```

If you get "No rule to make target 'build'" error, it means Make is having issues with the Makefile format. Use the scripts instead.

---

## Alternative: Manual Build and Install

### Build manually:

```bash
go build -o terraform-provider-opnsense
```

### Install manually:

```bash
# For Linux AMD64
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/linux_amd64/
cp terraform-provider-opnsense ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/linux_amd64/

# For Linux ARM64
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/linux_arm64/
cp terraform-provider-opnsense ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/linux_arm64/

# For macOS AMD64
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/darwin_amd64/
cp terraform-provider-opnsense ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/darwin_amd64/

# For macOS ARM64 (Apple Silicon)
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/darwin_arm64/
cp terraform-provider-opnsense ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/darwin_arm64/
```

---

## Verify Installation

```bash
# Check the binary was created
ls -lh terraform-provider-opnsense

# Check it was installed
ls -lh ~/.terraform.d/plugins/registry.terraform.io/yourusername/opnsense/0.1.0/
```

---

## First Use

Create a test configuration file `test.tf`:

```hcl
terraform {
  required_providers {
    opnsense = {
      source  = "yourusername/opnsense"
      version = "0.1.0"
    }
  }
}

provider "opnsense" {
  host       = "https://192.168.1.1"
  api_key    = "your-api-key"
  api_secret = "your-api-secret"
  insecure   = true
}

resource "opnsense_firewall_rule" "test" {
  description      = "Test rule"
  interface        = "lan"
  protocol         = "tcp"
  source_net       = "192.168.1.0/24"
  destination_net  = "any"
  destination_port = "443"
  action           = "pass"
  enabled          = true
}
```

Initialize and test:

```bash
terraform init
terraform plan
```

If you see the plan output, the provider is working!

---

## Troubleshooting

### "No rule to make target 'build'"

This is a Makefile tab/space issue. Use the shell scripts instead:
```bash
./build.sh
./install.sh
```

### "go: command not found"

Install Go:
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# CentOS/RHEL/Rocky
sudo yum install golang

# macOS
brew install go
```

### "terraform: command not found"

Install Terraform:
```bash
# Download from https://www.terraform.io/downloads
# Or use package manager

# Ubuntu/Debian
wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
sudo apt update && sudo apt install terraform
```

### Provider not found after installation

1. Check the path matches your OS/architecture
2. Verify the directory structure is correct
3. Run `terraform init -upgrade`
4. Clear the lock file: `rm .terraform.lock.hcl`

### "module not found" errors when building

Run:
```bash
go mod download
go mod tidy
```

Then try building again.

---



## Next Steps

Once installed, check out:
- `QUICKSTART.md` - 10-minute quick start
- `README.md` - Full documentation
- `examples/` - Example configurations
- `IMPLEMENTATION_GUIDE.md` - Technical details
