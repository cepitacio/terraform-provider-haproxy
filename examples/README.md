# HAProxy Terraform Provider Examples

This directory contains comprehensive examples for testing the HAProxy Terraform provider.

## Examples Overview

### 1. `test_fix.tf` - Simple Test Configuration
**Purpose**: Test the Read method fix we implemented
**What it tests**: 
- Minimal configuration with no optional fields set
- Verifies that `terraform plan` shows no changes after initial apply
- Tests that fields remain `null` in state when not configured

### 2. `comprehensive_example.tf` - Full Feature Demo
**Purpose**: Demonstrate all provider capabilities
**What it includes**:
- All resource types with nested blocks
- SSL/TLS configuration (API v2/v3)
- ACL rules and HTTP/TCP rules
- Health checks and monitoring
- Load balancing and stick tables
- Logging and statistics

## Testing the Read Method Fix

### Step 1: Use the Simple Test
```bash
cd examples
terraform init
terraform plan
```

**Expected**: Should show resources to be created

### Step 2: Apply the Configuration
```bash
terraform apply
```

**Expected**: Resources should be created successfully

### Step 3: Test the Fix
```bash
terraform plan
```

**Expected**: Should show "No changes. No objects need to be modified."

**Before the fix**: This would show constant updates to fields like:
- `ssl = false -> null`
- `timeout_client = 0 -> null`
- `http_log = false -> null`

**After the fix**: No changes should be detected

### Step 4: Verify State
```bash
terraform show
```

**Expected**: Fields that weren't set should show as `null` in the state

## Testing the Comprehensive Example

### Step 1: Review the Configuration
The comprehensive example includes:
- **Global**: Basic HAProxy settings
- **Resolver**: DNS resolution configuration
- **Peers**: HAProxy cluster configuration
- **Backend**: Full-featured backend with SSL, rules, and health checks
- **Frontend**: Complete frontend with bindings, ACLs, and rules
- **Servers**: Multiple servers with health checks and SSL
- **Log Forward**: External logging configuration
- **Stick Tables**: Session persistence configuration

### Step 2: Customize for Your Environment
Before running, update:
- Provider credentials (host, port, username, password)
- IP addresses and ports
- SSL certificate paths
- Network ranges

### Step 3: Test Incrementally
```bash
# Start with just the global configuration
terraform apply -target=haproxy_global.main

# Add resolver and peers
terraform apply -target=haproxy_resolver.main -target=haproxy_peers.main

# Add backend
terraform apply -target=haproxy_backend.web_backend

# Add frontend
terraform apply -target=haproxy_frontend.web_frontend

# Add servers
terraform apply -target=haproxy_server.web_servers
```

## What the Fix Accomplishes

### Before the Fix
- Read methods set ALL fields to their zero values (`""`, `false`, `0`)
- Terraform constantly tried to "update" fields to `null`
- `terraform plan` showed constant changes
- Unnecessary API calls and state updates

### After the Fix
- Read methods only set fields when they have meaningful values
- Fields remain `null` when not configured
- `terraform plan` shows "No changes" after initial apply
- Clean, predictable state management

## Testing Specific Scenarios

### Test 1: Empty String Fields
```hcl
resource "haproxy_backend" "test" {
  name = "test"
  mode = "http"
  # ciphers not set - should remain null
}
```

### Test 2: Boolean Fields
```hcl
resource "haproxy_frontend" "test" {
  name = "test"
  mode = "http"
  # http_log not set - should remain null
}
```

### Test 3: Numeric Fields
```hcl
resource "haproxy_server" "test" {
  name = "test"
  parent_type = "backend"
  parent_name = "test_backend"
  address = "192.168.1.10"
  port = 8080
  # weight not set - should remain null
}
```

## Troubleshooting

### If You Still See Updates
1. Check that you're using the updated provider
2. Verify the provider was rebuilt after changes
3. Check that all Read methods were updated
4. Look for any fields that might have been missed

### Common Issues
- **Provider not rebuilt**: Run `go build` again
- **Wrong provider version**: Check `terraform init` output
- **Cached state**: Try `terraform refresh`

## Expected Results

After applying the fix:
- ✅ `terraform plan` shows no changes
- ✅ Fields remain `null` when not configured
- ✅ No unnecessary state updates
- ✅ Clean, predictable behavior
- ✅ All nested blocks work correctly

## Next Steps

Once you've verified the fix works:
1. Test with your actual HAProxy configuration
2. Add more complex nested blocks
3. Test SSL/TLS configurations
4. Verify health checks and monitoring
5. Test load balancing scenarios

