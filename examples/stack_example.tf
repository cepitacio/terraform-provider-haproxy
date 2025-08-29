terraform {
  required_providers {
    haproxy = {
      source = "hashicorp.com/local/haproxy"
      version = "~> 0.1"
    }
  }
}

provider "haproxy" {
  host     = "localhost"
  port     = 5555
  username = "admin"
  password = "admin"
}

# Create a complete HAProxy stack in a single transaction
resource "haproxy_stack" "simple_stack" {
  name = "simple_stack"
  
  backend {
    name = "simple_backend"
    mode = "http"
  }
  
  server {
    name    = "simple_server"
    address = "192.168.1.10"
    port    = 8080
  }
  
  frontend {
    name            = "simple_frontend"
    mode            = "http"
    default_backend = "simple_backend"
  }
}

# Outputs to verify the configuration
output "stack_name" {
  value = haproxy_stack.simple_stack.name
}

output "backend_name" {
  value = haproxy_stack.simple_stack.backend.name
}

output "server_name" {
  value = haproxy_stack.simple_stack.server.name
}

output "frontend_name" {
  value = haproxy_stack.simple_stack.frontend.name
}
