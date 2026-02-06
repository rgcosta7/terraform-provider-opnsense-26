#!/bin/bash
# Test with the correct top-level key: subnet4 instead of subnet

HOST="${1}"
API_KEY="${2}"
API_SECRET="${3}"

echo "=== Testing with subnet4 key ==="
echo ""

# Test 1: Basic subnet4 structure
echo "Test 1: Basic subnet4"
curl -s -k -u "${API_KEY}:${API_SECRET}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"subnet4":{"subnet":"192.168.210.0/24"}}' \
  "${HOST}/api/kea/dhcpv4/add_subnet"
echo ""
echo "---"

# Test 2: With pools as string
echo "Test 2: With pools as STRING"
curl -s -k -u "${API_KEY}:${API_SECRET}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"subnet4":{"subnet":"192.168.211.0/24","pools":"192.168.211.10-192.168.211.100"}}' \
  "${HOST}/api/kea/dhcpv4/add_subnet"
echo ""
echo "---"

# Test 3: With description
echo "Test 3: With description"
curl -s -k -u "${API_KEY}:${API_SECRET}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"subnet4":{"subnet":"192.168.212.0/24","pools":"192.168.212.10-192.168.212.100","description":"Test subnet"}}' \
  "${HOST}/api/kea/dhcpv4/add_subnet"
echo ""
echo "---"

# Test 4: With option_data as empty object
echo "Test 4: With option_data"
curl -s -k -u "${API_KEY}:${API_SECRET}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"subnet4":{"subnet":"192.168.213.0/24","pools":"192.168.213.10-192.168.213.100","option_data":{}}}' \
  "${HOST}/api/kea/dhcpv4/add_subnet"
echo ""
echo "---"

# Check if any were created
echo "Checking existing subnets..."
curl -s -k -u "${API_KEY}:${API_SECRET}" \
  -X GET \
  "${HOST}/api/kea/dhcpv4/search_subnet" | python3 -m json.tool 2>/dev/null || cat
echo ""