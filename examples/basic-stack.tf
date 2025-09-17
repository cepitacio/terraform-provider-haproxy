# Basic HAProxy Stack Example
# This example shows a simple web application with load balancing

terraform {
  required_providers {
    haproxy = {
      source  = "cepitacio/haproxy"
      version = "~> 1.0"
    }
  }
}

provider "haproxy" {
  url         = "http://localhost:5555"
  username    = "admin"
  password    = "admin"
  api_version = "v3"
  insecure    = false
}

# Create a complete HAProxy stack
resource "haproxy_stack" "web_app" {
  name = "web_application"

  # Backend configuration
  backend {
    name = "web_backend"
    mode = "http"

    # Load balancing
    balance {
      algorithm = "roundrobin"
    }

    # Health checks
    http_checks {
      type = "connect"
      addr = "127.0.0.1"
      port = 80
    }

    # ACL for API requests
    acls {
      acl_name  = "is_api"
      criterion = "path"
      value     = "/api"
    }

    # HTTP request rules
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_api"
    }

    # Backend servers
    servers = {
      "web_server_1" = {
        address = "192.168.1.10"
        port    = 8080
        check   = "enabled"
        weight  = 100
      }

      "web_server_2" = {
        address = "192.168.1.11"
        port    = 8080
        check   = "enabled"
        weight  = 100
      }
    }
  }

  # Frontend configuration
  frontend {
    name            = "web_frontend"
    mode            = "http"
    default_backend = "web_backend"

    # Bind to port 80
    binds = {
      http_bind = {
        address = "0.0.0.0"
        port    = 80
      }
    }

    # ACL for admin requests
    acls {
      acl_name  = "is_admin"
      criterion = "path"
      value     = "/admin"
    }

    # HTTP request rules
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_admin"
    }
  }

}

# Data sources to discover existing configurations
data "haproxy_backends" "all" {}
data "haproxy_frontends" "all" {}

# Outputs
output "backend_names" {
  description = "Names of all discovered backends"
  value       = [for backend in jsondecode(data.haproxy_backends.all.backends) : backend.name]
}

output "frontend_names" {
  description = "Names of all discovered frontends"
  value       = [for frontend in jsondecode(data.haproxy_frontends.all.frontends) : frontend.name]
}

output "stack_name" {
  description = "Name of the created stack"
  value       = haproxy_stack.web_app.name
}
