# HAProxy Terraform Provider Example
# This example shows basic usage of the haproxy_stack resource

terraform {
  required_providers {
    haproxy = {
      source  = "cepitacio/haproxy"
      version = "~> 1.0"
    }
  }
}

provider "haproxy" {
  host        = "localhost"
  port        = 5555
  username    = "admin"
  password    = "admin"
  api_version = "v3"
}

# Multiple applications using for_each
locals {
  applications = {
    web = {
      backend_name  = "web_backend"
      frontend_name = "web_frontend"
      servers = {
        "web1" = { address = "192.168.1.10", port = 8080 }
        "web2" = { address = "192.168.1.11", port = 8080 }
      }
    }
    api = {
      backend_name  = "api_backend"
      frontend_name = "api_frontend"
      servers = {
        "api1" = { address = "192.168.1.20", port = 8081 }
        "api2" = { address = "192.168.1.21", port = 8081 }
      }
    }
  }
}

# Create multiple HAProxy stacks using for_each
resource "haproxy_stack" "apps" {
  for_each = local.applications
  
  name = each.key

  # Servers for this application
  servers = {
    for name, server in each.value.servers : name => merge(server, {
      check = "enabled"
    })
  }

  # Backend configuration
  backend {
    name = each.value.backend_name
    mode = "http"
    
    # Health check
    http_checks {
      type = "connect"
      port = each.value.servers[keys(each.value.servers)[0]].port
    }
  }

  # Frontend configuration
  frontend {
    name = each.value.frontend_name
    mode = "http"
    default_backend = each.value.backend_name
    
    binds = {
      "http" = {
        address = "*:80"
      }
    }
  }
}

# Data sources example
data "haproxy_backends" "all" {}
data "haproxy_frontends" "all" {}

# Output discovered configurations
output "all_backends" {
  value = data.haproxy_backends.all
}

output "all_frontends" {
  value = data.haproxy_frontends.all
}
