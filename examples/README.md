# HAProxy Terraform Provider Examples

This directory contains comprehensive examples for using the HAProxy Terraform provider with the HAProxy Data Plane API.

## Examples Overview

### 1. `example.tf` - Multi-Application Configuration
**Purpose**: Demonstrates using `for_each` to create multiple HAProxy stacks
**What it includes**:
- Multiple applications (web, api) using local variables
- Dynamic server configuration
- Basic health checks and load balancing
- Data source usage for discovery

### 2. `basic-stack.tf` - Simple Web Application
**Purpose**: Shows a complete, simple HAProxy configuration
**What it includes**:
- Single stack with backend and frontend
- Servers nested under backend (new structure)
- ACLs and HTTP request rules
- Health checks and load balancing
- Multiple backend servers
- Data source discovery and outputs

### 3. `data-sources.tf` - Data Discovery Examples
**Purpose**: Demonstrates all available data sources
**What it includes**:
- List data sources (all backends, frontends, etc.)
- Single item data sources (specific backend, ACL, etc.)
- Cross-referencing data between sources
- Local values and outputs using discovered data

### 4. `advanced-configuration.tf` - Enterprise Features
**Purpose**: Shows advanced HAProxy features and configurations
**What it includes**:
- SSL/TLS configuration with certificates
- Multiple ACLs and complex rule chains
- HTTP and TCP request/response rules
- Advanced health checks
- Session persistence with stick tables
- Multiple environments using for_each

## Getting Started

### Prerequisites

1. **HAProxy with Data Plane API**: Ensure HAProxy is running with the Data Plane API enabled
2. **Terraform**: Version 1.0 or later
3. **Provider**: The HAProxy Terraform provider

### Step 1: Choose an Example

Start with `basic-stack.tf` for a simple configuration, or `example.tf` for multiple applications.

### Step 2: Configure Provider

Update the provider configuration in your chosen example:

```hcl
provider "haproxy" {
  url         = "http://your-haproxy:5555"  # Update with your HAProxy URL
  username    = "admin"                     # Update with your username
  password    = "admin"                     # Update with your password
  api_version = "v3"                       # Use v3 for full features
  insecure    = false                      # Set to true to skip SSL verification
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

### Basic Stack (`basic-stack.tf`)
- **Use case**: Simple web application with load balancing
- **Features**: ACLs, HTTP rules, health checks, multiple servers
- **Complexity**: Beginner-friendly

### Data Sources (`data-sources.tf`)
- **Use case**: Discovering and referencing existing HAProxy configurations
- **Features**: All 22 data sources with examples
- **Complexity**: Intermediate - shows data discovery patterns

### Advanced Configuration (`advanced-configuration.tf`)
- **Use case**: Enterprise-grade HAProxy with SSL, complex rules
- **Features**: SSL/TLS, stick tables, multiple environments
- **Complexity**: Advanced - production-ready configurations

### Multi-Application (`example.tf`)
- **Use case**: Managing multiple applications with for_each
- **Features**: Dynamic server configuration, environment separation
- **Complexity**: Intermediate - shows scaling patterns

## Customizing Examples

### Update Provider Configuration
All examples use placeholder values. Update these for your environment:

```hcl
provider "haproxy" {
  url         = "http://your-haproxy:5555"  # Your HAProxy Data Plane API URL
  username    = "your-username"             # Your API username
  password    = "your-password"             # Your API password
  api_version = "v3"                       # Use v3 for full features
  insecure    = false                      # Set to true to skip SSL verification
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
bind {
  name            = "https_bind"
  address         = "0.0.0.0"
  port            = 443
  ssl             = true
  ssl_certificate = "/path/to/your/certificate.crt"  # Update path
  ssl_cafile      = "/path/to/your/ca.crt"           # Update path
}
```

## API Version Considerations

### HAProxy Data Plane API v3 (Recommended)
- ✅ Full support for all features
- ✅ TCP rules work with both frontends and backends
- ✅ HTTP checks work with both frontends and backends
- ✅ All examples work as written

### HAProxy Data Plane API v2 (Limited)
- ⚠️ TCP rules only work with backends
- ⚠️ HTTP checks only work with backends
- ⚠️ Some advanced examples may not work
- ✅ Basic examples work fine

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

### Getting Help

- Check the [main documentation](../docs/)
- Review [data sources](../docs/data-sources/)
- See [resource documentation](../docs/resources/)
- Open an [issue](https://github.com/cepitacio/terraform-provider-haproxy/issues)

