# Advanced HAProxy Configuration Example
# This example shows advanced features like SSL, multiple rules, and complex configurations

terraform {
  required_providers {
    haproxy = {
      source  = "cepitacio/haproxy"
      version = "~> 1.0"
    }
  }
}

provider "haproxy" {
  url         = "https://haproxy.example.com:5555"
  username    = "admin"
  password    = "admin"
  api_version = "v3"
  insecure    = false
}

# Advanced backend with SSL, multiple rules, and health checks
resource "haproxy_stack" "secure_app" {
  name = "secure_application"

  backend {
    name = "secure_backend"
    mode = "http"
    
    # Load balancing with stickiness
    balance {
      algorithm = "leastconn"
    }
    
    # SSL configuration
    default_server {
      ssl                = "enabled"
      ssl_certificate    = "/etc/ssl/certs/server.crt"
      ssl_cafile         = "/etc/ssl/certs/ca.crt"
      ssl_min_ver        = "TLSv1.2"
      ssl_max_ver        = "TLSv1.3"
      verify             = "required"
    }
    
    # Multiple ACLs
    acls {
      acl_name = "is_api"
      criterion = "path"
      value     = "/api"
    }
    
    acls {
      acl_name = "is_admin"
      criterion = "path"
      value     = "/admin"
    }
    
    acls {
      acl_name = "is_secure"
      criterion = "req.ssl_ver"
      value     = "TLSv1.2"
    }
    
    # HTTP request rules
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_secure"
    }
    
    http_request_rules {
      type      = "redirect"
      redir_type = "location"
      redir_code = 301
      redir_value = "https://secure.example.com"
      cond      = "unless"
      cond_test = "is_secure"
    }
    
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_api"
    }
    
    # HTTP response rules
    http_response_rules {
      type      = "add-header"
      hdr_name  = "X-Backend"
      hdr_format = "secure_backend"
    }
    
    # TCP request rules (API v3 only)
    tcp_request_rules {
      type      = "inspect-delay"
      timeout   = 5000
    }
    
    # Health checks
    http_checks {
      type = "connect"
      addr = "127.0.0.1"
      port = 443
      ssl  = true
    }
    
    http_checks {
      type    = "send"
      method  = "GET"
      uri     = "/health"
      version = "HTTP/1.1"
    }
    
    http_checks {
      type    = "expect"
      match   = "status"
      pattern = "200"
    }
    
    # Stick table for session persistence
    stick_table {
      type   = "ip"
      size   = "100k"
      expire = "30m"
    }
    
    stick_rule {
      type    = "match"
      pattern = "src"
      table   = "secure_backend"
    }
  }

  # Advanced frontend with SSL and multiple binds
  frontend {
    name           = "secure_frontend"
    mode           = "http"
    default_backend = "secure_backend"
    
    # HTTP bind
    bind {
      name    = "http_bind"
      address = "0.0.0.0"
      port    = 80
    }
    
    # HTTPS bind with SSL
    bind {
      name            = "https_bind"
      address         = "0.0.0.0"
      port            = 443
      ssl             = true
      ssl_certificate = "/etc/ssl/certs/frontend.crt"
      ssl_cafile      = "/etc/ssl/certs/ca.crt"
      ssl_min_ver     = "TLSv1.2"
      ssl_max_ver     = "TLSv1.3"
    }
    
    # Multiple ACLs
    acls {
      acl_name = "is_admin"
      criterion = "path"
      value     = "/admin"
    }
    
    acls {
      acl_name = "is_api"
      criterion = "path"
      value     = "/api"
    }
    
    acls {
      acl_name = "is_secure"
      criterion = "req.ssl_ver"
      value     = "TLSv1.2"
    }
    
    acls {
      acl_name = "is_local"
      criterion = "src"
      value     = "192.168.0.0/16"
    }
    
    # HTTP request rules
    http_request_rules {
      type      = "redirect"
      redir_type = "location"
      redir_code = 301
      redir_value = "https://secure.example.com"
      cond      = "if"
      cond_test = "!is_secure"
    }
    
    http_request_rules {
      type      = "allow"
      cond      = "if"
      cond_test = "is_local"
    }
    
    http_request_rules {
      type      = "deny"
      cond      = "if"
      cond_test = "is_admin"
    }
    
    # HTTP response rules
    http_response_rules {
      type      = "add-header"
      hdr_name  = "X-Frontend"
      hdr_format = "secure_frontend"
    }
    
    http_response_rules {
      type      = "add-header"
      hdr_name  = "Strict-Transport-Security"
      hdr_format = "max-age=31536000; includeSubDomains"
      cond      = "if"
      cond_test = "is_secure"
    }
    
    # TCP request rules (API v3 only)
    tcp_request_rules {
      type      = "inspect-delay"
      timeout   = 5000
    }
    
    tcp_request_rules {
      type      = "capture"
      capture_len = 64
      capture_sample = "req.cook(JSESSIONID)"
    }

    # Multiple servers with different configurations
    servers = {
      "secure_server_1" = {
        address = "192.168.1.10"
        port    = 8443
        check   = "enabled"
        weight  = 100
        ssl     = "enabled"
        verify  = "required"
      }
      
      "secure_server_2" = {
        address = "192.168.1.11"
        port    = 8443
        check   = "enabled"
        weight  = 100
        ssl     = "enabled"
        verify  = "required"
      }
      
      "secure_server_3" = {
        address = "192.168.1.12"
        port    = 8443
        check   = "enabled"
        weight  = 50
        ssl     = "enabled"
        verify  = "required"
        backup  = "enabled"
      }
    }
  }
}

# Multiple stacks using for_each
locals {
  environments = {
    production = {
      port = 443
      servers = {
        "prod1" = { address = "10.0.1.10", port = 8443, weight = 100 }
        "prod2" = { address = "10.0.1.11", port = 8443, weight = 100 }
      }
    }
    staging = {
      port = 8443
      servers = {
        "stage1" = { address = "10.0.2.10", port = 8443, weight = 50 }
      }
    }
  }
}

resource "haproxy_stack" "environments" {
  for_each = local.environments
  
  name = each.key

  backend {
    name = "${each.key}_backend"
    mode = "http"
    
    balance {
      algorithm = "roundrobin"
    }

    # Dynamic servers
    servers = {
      for name, server in each.value.servers : name => merge(server, {
        check = "enabled"
      })
    }
  }

  frontend {
    name           = "${each.key}_frontend"
    mode           = "http"
    default_backend = "${each.key}_backend"
    
    bind {
      name    = "${each.key}_bind"
      address = "0.0.0.0"
      port    = each.value.port
    }
  }
}

# Outputs
output "secure_app_name" {
  description = "Name of the secure application stack"
  value       = haproxy_stack.secure_app.name
}

output "environment_names" {
  description = "Names of all environment stacks"
  value       = [for env in haproxy_stack.environments : env.name]
}

output "total_servers" {
  description = "Total number of servers across all stacks"
  value       = length(haproxy_stack.secure_app.servers) + sum([for env in haproxy_stack.environments : length(env.servers)])
}
