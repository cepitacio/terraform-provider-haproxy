# HAProxy Terraform Provider

A comprehensive Terraform provider for managing HAProxy configuration using the HAProxy Data Plane API (v2/v3).

## Features

- **Complete HAProxy Management**: Manage frontends, backends, servers, ACLs, rules, and more
- **Data Discovery**: Powerful data sources for discovering existing HAProxy configurations
- **Atomic Operations**: Stack resource for creating complete HAProxy configurations in single transactions
- **Multi-Version Support**: Compatible with HAProxy Data Plane API v2 and v3 (v3 recommended for full feature support)
- **Comprehensive Coverage**: Support for HTTP/TCP rules, health checks, binds, and advanced features
- **Concurrent Operations**: Support for multiple `haproxy_stack` resources with `for_each`
- **Transaction Retry Logic**: Robust error handling with automatic retry for transaction conflicts
- **Version-Aware Operations**: Different handling for API v2 (individual operations) vs v3 (bulk operations)

## Quick Start

### Provider Configuration

```hcl
terraform {
  required_providers {
    haproxy = {
      source  = "your-org/haproxy"
      version = "~> 1.0"
    }
  }
}

provider "haproxy" {
  host        = "localhost"
  port        = 5555
  username    = "admin"
  password    = "admin"
  api_version = "v3"  # Default is v3, use v2 if needed (v2 has limitations)
}
```

### Basic Example

```hcl
# Create a complete HAProxy stack
resource "haproxy_stack" "web_app" {
  name = "web_application"

  backend {
    name = "web_backend"
    mode = "http"
    
    # Backend ACLs
    acls {
      acl_name = "is_api"
      criterion = "path"
      value     = "/api"
    }
    
    # HTTP request rules
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_api"
    }

    # Health checks
    http_checks {
      type = "connect"
      addr = "127.0.0.1"
      port = 80
    }
  }

  frontend {
    name           = "web_frontend"
    mode           = "http"
    default_backend = "web_backend"
    
    # Frontend ACLs
    acls {
      acl_name = "is_admin"
      criterion = "path"
      value     = "/admin"
    }
    
    # HTTP request rules
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_admin"
    }

    # Bind configuration
    bind {
      name    = "http_bind"
      address = "0.0.0.0"
      port    = 80
    }
  }

  # Backend servers
  server {
    name    = "web_server_1"
    address = "192.168.1.10"
    port    = 8080
    check   = "enabled"
    weight  = 100
  }
}
```

## Data Sources (Discovery Tools)

The provider includes powerful data sources for discovering existing HAProxy configurations:

### List Data Sources

```hcl
# Get all backends
data "haproxy_backends" "all" {}

# Get all frontends
data "haproxy_frontends" "all" {}

# Get all servers in a backend
data "haproxy_server" "backend_servers" {
  parent_type = "backend"
  parent_name = "my_backend"
}

# Get all ACLs for a backend
data "haproxy_acl" "backend_acls" {
  parent_type = "backend"
  parent_name = "my_backend"
}

# Get all HTTP request rules for a frontend
data "haproxy_http_request_rule" "frontend_rules" {
  parent_type = "frontend"
  parent_name = "my_frontend"
}

# Get all health checks for a backend
data "haproxy_httpcheck" "backend_checks" {
  parent_type = "backend"
  parent_name = "my_backend"
}
```

### Single Item Data Sources

```hcl
# Get a specific ACL
data "haproxy_acl_single" "admin_acl" {
  index       = 0
  parent_type = "backend"
  parent_name = "my_backend"
}

# Get a specific HTTP request rule
data "haproxy_http_request_rule_single" "allow_rule" {
  index       = 0
  parent_type = "frontend"
  parent_name = "my_frontend"
}

# Get a specific backend
data "haproxy_backend_single" "web_backend" {
  name = "web_backend"
}

# Get a specific server
data "haproxy_server_single" "web_server" {
  name    = "web_server_1"
  backend = "web_backend"
}
```

### Using Data Source Outputs

Data sources return complete JSON data from the HAProxy API:

```hcl
# Reference ACL data in resources
resource "haproxy_http_request_rule" "conditional_rule" {
  parent_type = "backend"
  parent_name = "my_backend"
  type        = "allow"
  cond        = "if"
  cond_test   = jsondecode(data.haproxy_acl_single.admin_acl.acl).acl_name
}

# Use in outputs
output "backend_config" {
  value = {
    name = jsondecode(data.haproxy_backend_single.web_backend.backend).name
    mode = jsondecode(data.haproxy_backend_single.web_backend.backend).mode
  }
}

# Loop through discovered resources
resource "haproxy_http_request_rule" "discovered_rules" {
  for_each = { for rule in jsondecode(data.haproxy_http_request_rule.frontend_rules.http_request_rules) : rule.index => rule }
  
  parent_type = "frontend"
  parent_name = "my_frontend"
  type        = each.value.type
  cond        = each.value.cond
  cond_test   = each.value.cond_test
  index       = each.value.index
}
```

## Resources

### Stack Resource

The `haproxy_stack` resource creates complete HAProxy configurations in a single atomic transaction:

```hcl
resource "haproxy_stack" "complete_app" {
  name = "my_application"

  # Backend configuration
  backend {
    name = "app_backend"
    mode = "http"
    
    # Multiple ACLs
    acls {
      acl_name = "is_api"
      criterion = "path"
      value     = "/api"
    }
    
    # HTTP rules
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_api"
    }
    
    # Health checks
    http_checks {
      type = "connect"
      addr = "127.0.0.1"
      port = 80
    }
  }

  # Frontend configuration
  frontend {
    name           = "app_frontend"
    mode           = "http"
    default_backend = "app_backend"
    
    # Frontend ACLs
    acls {
      acl_name = "is_admin"
      criterion = "path"
      value     = "/admin"
    }
    
    # Bind configuration
    bind {
      name    = "http_bind"
      address = "0.0.0.0"
      port    = 80
    }
  }

  # Multiple servers
  server {
    name    = "server_1"
    address = "192.168.1.10"
    port    = 8080
    weight  = 100
  }
  
  server {
    name    = "server_2"
    address = "192.168.1.11"
    port    = 8080
    weight  = 100
  }
}
```

## Examples

See the `/examples` directory for comprehensive examples:

- `01-basic-stack.tf` - Basic HAProxy configuration
- `02-data-sources.tf` - Data source usage examples
- `03-advanced-configuration.tf` - Advanced backend configuration

**Note**: Index fields for ACLs, rules, and health checks are automatically managed by the provider - you don't need to specify them in your configuration.

## API Version Differences

### HAProxy Data Plane API v3 (Recommended)
- ✅ **Full feature support** for all resource types
- ✅ **TCP rules** work with both frontends and backends
- ✅ **HTTP checks** work with both frontends and backends
- ✅ **All ACLs and rules** supported everywhere

### HAProxy Data Plane API v2 (Limited)
- ⚠️ **TCP request rules**: Only work with backends, not frontends
- ⚠️ **TCP response rules**: Only work with backends, not frontends  
- ⚠️ **HTTP checks**: Only work with backends, not frontends
- ✅ **HTTP request/response rules**: Work with both frontends and backends
- ✅ **ACLs**: Work with both frontends and backends

**Recommendation**: Use v3 for full feature support, or v2 only if you have specific compatibility requirements.

## Requirements

- Terraform >= 1.0
- HAProxy with Data Plane API enabled
- Go 1.23+ (for building from source)

## Installation

### From Terraform Registry (when published)

```hcl
terraform {
  required_providers {
    haproxy = {
      source  = "your-org/haproxy"
      version = "~> 1.0"
    }
  }
}
```

### From Source

```bash
git clone https://github.com/your-org/terraform-provider-haproxy
cd terraform-provider-haproxy
go build -o terraform-provider-haproxy
```

## Configuration

### Provider Arguments

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| host | HAProxy Data Plane API host | string | localhost | yes |
| port | HAProxy Data Plane API port | number | 5555 | yes |
| username | API username | string | - | yes |
| password | API password | string | - | yes |
| api_version | API version (v2 or v3) | string | v3 | no* |

*Required when using v2, optional when using v3 (default). **Note**: v2 has limitations - TCP rules and HTTP checks only work with backends, not frontends.
| insecure | Skip TLS verification | bool | false | no |

### Environment Variables

- `HAPROXY_HOST` - Override host
- `HAPROXY_PORT` - Override port
- `HAPROXY_USERNAME` - Override username
- `HAPROXY_PASSWORD` - Override password
- `HAPROXY_API_VERSION` - Override API version (default: v3)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [Issues](https://github.com/your-org/terraform-provider-haproxy/issues)
- [Documentation](https://github.com/your-org/terraform-provider-haproxy/tree/main/docs)
- [Examples](https://github.com/your-org/terraform-provider-haproxy/tree/main/examples)