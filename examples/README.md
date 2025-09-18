# HAProxy Terraform Provider Examples

This directory contains comprehensive examples for using the HAProxy Terraform provider with the HAProxy Data Plane API.

## Examples Overview

### 1. `resources-example.tf` - Complete Working Configuration
**Purpose**: Demonstrates a comprehensive HAProxy stack with all supported features
**What it includes**:
- Complete backend configuration with SSL, health checks, and rules
- Frontend with multiple binds (HTTP, HTTPS, admin)
- ACLs, HTTP request/response rules, TCP request/response rules
- HTTP checks and TCP checks
- Working example based on real user configuration
- Note about v3 TLS field issues

### 2. `data-sources.tf` - Data Discovery Examples
**Purpose**: Demonstrates how to query existing HAProxy configurations
**What it includes**:
- Data sources for all backends and frontends
- Single resource data sources (backend, frontend, bind, ACL, etc.)
- Rule and check data sources with correct parent types
- Outputs showing how to use discovered data
- Based on actual resources from `resources-example.tf`

## Getting Started

### Prerequisites

1. **HAProxy with Data Plane API**: Ensure HAProxy is running with the Data Plane API enabled
2. **Terraform**: Version 1.0 or later
3. **Provider**: The HAProxy Terraform provider

### Step 1: Choose an Example

Start with `resources-example.tf` to create a complete HAProxy configuration, or `data-sources.tf` to query existing configurations.

### Step 2: Configure Provider

Update the provider configuration in your chosen example:

```hcl
provider "haproxy" {
  url         = "https://haproxy.example.com:5555"  # Update with your HAProxy URL
  username    = "admin"                             # Update with your username
  password    = "admin"                             # Update with your password
  api_version = "v3"                               # Use v3 for full features
  insecure    = false                              # Set to true to skip SSL verification
}
```

### Step 3: Initialize and Apply

```bash
cd examples
terraform init
terraform plan
terraform apply
```

**Expected**: Resources should be created successfully

### Step 4: Verify Configuration

```bash
terraform show
```

**Expected**: All resources should be in the desired state

## Example Details

### Resources Example (`resources-example.tf`)
- **Use case**: Complete HAProxy configuration with all features
- **Features**: SSL/TLS, ACLs, rules, health checks, multiple binds
- **Complexity**: Comprehensive - shows all supported features
- **Based on**: Real working configuration from user testing

### Data Sources (`data-sources.tf`)
- **Use case**: Querying existing HAProxy configurations
- **Features**: All data sources with correct parameters and parent types
- **Complexity**: Intermediate - shows data discovery patterns
- **Based on**: Resources created by `resources-example.tf`

## Key Features Demonstrated

### Backend Configuration
- SSL/TLS settings with certificates
- Health checks (HTTP and TCP)
- Load balancing algorithms
- ACLs and rules
- Timeout configurations

### Frontend Configuration
- Multiple binds (HTTP, HTTPS, admin)
- SSL/TLS configuration
- ACLs and rules
- Monitor fail settings

### Data Sources
- Query all backends and frontends
- Get specific resources by name
- Access rules and checks by index
- Cross-reference data between resources

## API Version Considerations

### HAProxy Data Plane API v3 (Recommended)
- ✅ Full support for all features
- ✅ All examples work as written
- ⚠️ Note: `tlsv*` and `sslv*` fields have issues in v3 API
- ✅ Use `force_tlsv*` and `force_sslv*` fields instead

### HAProxy Data Plane API v2 (Limited)
- ⚠️ Some advanced features may not work
- ✅ Basic examples work fine
- ✅ `force_tlsv*` and `force_sslv*` fields work in both versions

## Customizing Examples

### Update Provider Configuration
All examples use placeholder values. Update these for your environment:

```hcl
provider "haproxy" {
  url         = "https://your-haproxy:5555"     # Your HAProxy Data Plane API URL
  username    = "your-username"                 # Your API username
  password    = "your-password"                 # Your API password
  api_version = "v3"                           # Use v3 for full features
  insecure    = false                          # Set to true to skip SSL verification
}
```

### Update Server Addresses
Replace placeholder IP addresses with your actual server addresses:

```hcl
servers = {
  "web_server_1" = {
    address = "192.168.1.10"  # Update with your server IP
    port    = 8080            # Update with your server port
    check   = "enabled"
    weight  = 100
  }
}
```

### Update SSL Certificates
For SSL examples, update certificate paths:

```hcl
binds = {
  https_bind = {
    address         = "0.0.0.0"
    port            = 443
    ssl             = true
    ssl_certificate = "/path/to/your/certificate.crt"  # Update path
    ssl_cafile      = "/path/to/your/ca.crt"           # Update path
  }
}
```

## Data Source Usage

Data sources can be referenced in other resources:

```hcl
# Use data source in a new resource
resource "haproxy_stack" "new_stack" {
  name = "data_driven_stack"

  backend {
    name = "copy_of_${data.haproxy_backend_single.existing_backend.name}"
    mode = "http"
  }
}

# Use data source in outputs
output "existing_backend_mode" {
  value = data.haproxy_backend_single.existing_backend.mode
}

# Use data source in locals
locals {
  backend_names = [for backend in data.haproxy_backends.all_backends.backends : backend.name]
}
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check HAProxy Data Plane API is running
   - Verify URL and port are correct
   - Check firewall settings

2. **Authentication Failed**
   - Verify username and password
   - Check API user permissions

3. **SSL Certificate Errors**
   - Set `insecure = true` for testing
   - Verify certificate paths and validity

4. **API Version Errors**
   - Use `api_version = "v3"` for full features
   - Some features require v3

5. **Data Source "Not Found" Errors**
   - Verify the resource exists in HAProxy
   - Check parent_type and parent_name parameters
   - Verify index numbers for rules and checks

### Getting Help

- Check the [main documentation](../docs/)
- Review [data sources](../docs/data-sources/)
- See [resource documentation](../docs/resources/)
- Open an [issue](https://github.com/cepitacio/terraform-provider-haproxy/issues)