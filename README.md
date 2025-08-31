# HAProxy Terraform Provider

This is a Terraform provider for managing HAProxy configuration using the HAProxy Data Plane API.

## Architecture

![HAProxy Terraform Provider Architecture](assets/haproxy.png)

## Usage

### Basic Provider Configuration

```hcl
provider "haproxy" {
  host         = "localhost"
  port         = 5555
  username     = "admin"
  password     = "admin"
  api_version  = "v2"  # or "v3"
}
```

### Complete Stack with Frontend and Backend ACLs

```hcl
resource "haproxy_stack" "example" {
  name = "web_application"

  backend {
    name = "web_backend"
    mode = "http"
    
    # Backend ACLs for content switching
    acls {
      acl_name = "is_api_request"
      criterion = "path"
      value = "/api"
      index = 0
    }
    
    acls {
      acl_name = "is_admin_user"
      criterion = "hdr"
      value = "X-User-Role admin"
      index = 1
    }

    server_timeout = 30000
    check_timeout = 2000
  }

  server {
    name    = "web_server_1"
    address = "192.168.1.10"
    port    = 8080
    check   = "enabled"
    weight  = 100
  }

  frontend {
    name = "web_frontend"
    mode = "http"
    default_backend = "web_backend"
    
    # Frontend ACLs for routing
    acls {
      acl_name = "is_static_content"
      criterion = "path"
      value = "/static"
      index = 0
    }
    
    acls {
      acl_name = "is_internal_network"
      criterion = "src"
      value = "192.168.0.0/16"
      index = 1
    }

    bind {
      name    = "http_bind"
      address = "0.0.0.0"
      port    = 80
    }
  }
}
```

### Key Features

- **Atomic Operations**: All resources (ACLs, frontend, backend, servers) are created, updated, and deleted in single transactions
- **ACL Management**: Both frontend and backend ACLs with automatic index management and deduplication
- **Robust Error Handling**: Graceful handling of missing resources and automatic retry logic
- **State Consistency**: Maintains configuration consistency even with HAProxy state mismatches

## Resources

### Individual Resources
- `haproxy_frontend`
- `haproxy_backend`
- `haproxy_server`
- `haproxy_bind`
- `haproxy_acl`
- `haproxy_http_request_rule`
- `haproxy_http_response_rule`
- `haproxy_tcp_request_rule`
- `haproxy_tcp_response_rule`
- `haproxy_resolver`
- `haproxy_nameserver`
- `haproxy_peers`
- `haproxy_peer_entry`
- `haproxy_stick_rule`
- `haproxy_stick_table`
- `haproxy_log_forward`
- `haproxy_global`

### Stack Resource (Recommended)
- `haproxy_stack` - Creates a complete HAProxy stack (backend, frontend, server) in a single transaction with support for:
  - **Frontend ACLs**: Access control lists for routing and decision making
  - **Backend ACLs**: Access control lists for content switching and backend-specific logic
  - **Atomic operations**: All resources created, updated, and deleted in single transactions
  - **Robust error handling**: Graceful handling of missing resources and retry logic

## Data Sources

- `haproxy_frontends`
- `haproxy_backends`
