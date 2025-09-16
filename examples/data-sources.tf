# Data Sources Example
# This example demonstrates how to use data sources to discover and reference existing HAProxy configurations

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

# List data sources - get all items of a type
data "haproxy_backends" "all" {}
data "haproxy_frontends" "all" {}

# Single item data sources - get specific items
data "haproxy_backend_single" "web_backend" {
  name = "web_backend"
}

data "haproxy_frontend_single" "web_frontend" {
  name = "web_frontend"
}

# Get servers from a specific backend
data "haproxy_server" "backend_servers" {
  backend = "web_backend"
}

# Get ACLs from a frontend
data "haproxy_acl" "frontend_acls" {
  parent_type = "frontend"
  parent_name = "web_frontend"
}

# Get HTTP request rules from a backend
data "haproxy_http_request_rule" "backend_rules" {
  parent_type = "backend"
  parent_name = "web_backend"
}

# Get bind configurations from a frontend
data "haproxy_bind" "frontend_binds" {
  parent_type = "frontend"
  parent_name = "web_frontend"
}

# Single item data sources for specific rules
data "haproxy_acl_single" "admin_acl" {
  index       = 0
  parent_type = "frontend"
  parent_name = "web_frontend"
}

data "haproxy_http_request_rule_single" "allow_rule" {
  index       = 0
  parent_type = "backend"
  parent_name = "web_backend"
}

# Outputs demonstrating data source usage
output "all_backends" {
  description = "All backends from HAProxy"
  value       = jsondecode(data.haproxy_backends.all.backends)
}

output "all_frontends" {
  description = "All frontends from HAProxy"
  value       = jsondecode(data.haproxy_frontends.all.frontends)
}

output "web_backend_config" {
  description = "Specific backend configuration"
  value       = jsondecode(data.haproxy_backend_single.web_backend.backend)
}

output "backend_servers" {
  description = "Servers in the web_backend"
  value       = jsondecode(data.haproxy_server.backend_servers.servers)
}

output "frontend_acls" {
  description = "ACLs in the web_frontend"
  value       = jsondecode(data.haproxy_acl.frontend_acls.acls)
}

output "backend_rules" {
  description = "HTTP request rules in the web_backend"
  value       = jsondecode(data.haproxy_http_request_rule.backend_rules.http_request_rules)
}

output "frontend_binds" {
  description = "Bind configurations in the web_frontend"
  value       = jsondecode(data.haproxy_bind.frontend_binds.binds)
}

# Example of using data source outputs in local values
locals {
  # Extract backend names
  backend_names = [for backend in jsondecode(data.haproxy_backends.all.backends) : backend.name]

  # Extract server addresses
  server_addresses = [for server in jsondecode(data.haproxy_server.backend_servers.servers) : server.address]

  # Extract ACL names
  acl_names = [for acl in jsondecode(data.haproxy_acl.frontend_acls.acls) : acl.acl_name]

  # Extract rule types
  rule_types = [for rule in jsondecode(data.haproxy_http_request_rule.backend_rules.http_request_rules) : rule.type]
}

output "extracted_data" {
  description = "Extracted data from data sources"
  value = {
    backend_names    = local.backend_names
    server_addresses = local.server_addresses
    acl_names        = local.acl_names
    rule_types       = local.rule_types
  }
}
