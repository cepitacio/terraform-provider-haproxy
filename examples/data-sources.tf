# HAProxy Data Sources Example
# This example shows how to use data sources to query existing HAProxy configurations
# Based on the resources created in resources-example.tf

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

# Data source to get all backends (no parameters required)
data "haproxy_backends" "all_backends" {}

# Data source to get the specific backend from resources-example.tf
data "haproxy_backend_single" "test_backend" {
  name = "test_backend2"
}

# Data source to get all frontends (no parameters required)
data "haproxy_frontends" "all_frontends" {}

# Data source to get the specific frontend from resources-example.tf
data "haproxy_frontend_single" "test_frontend" {
  name = "test_frontend"
}

# Data source to get all servers from test_backend
data "haproxy_server" "all_servers" {
  backend = "test_backend"
}


# Data source to get the https_bind from the frontend
data "haproxy_bind_single" "https_bind" {
  parent_type = "frontend"
  parent_name = "test_frontend"
  name        = "https_bind"
}

# Data source to get the is_admin ACL from the frontend (index 0)
data "haproxy_acl_single" "is_admin_acl" {
  parent_type = "frontend"
  parent_name = "test_frontend"
  index       = 0
}

# Data source to get the is_api ACL from the frontend (index 3)
data "haproxy_acl_single" "is_api_acl" {
  parent_type = "frontend"
  parent_name = "test_frontend"
  index       = 3
}

# Data source to get the first HTTP request rule from the frontend
data "haproxy_http_request_rule_single" "allow_rule" {
  parent_type = "frontend"
  parent_name = "test_frontend"
  index       = 0
}

# Data source to get the first HTTP response rule from the frontend
data "haproxy_http_response_rule_single" "header_rule" {
  parent_type = "frontend"
  parent_name = "test_frontend"
  index       = 0
}

# Data source to get the first TCP request rule from the frontend
data "haproxy_tcp_request_rule_single" "nice_rule" {
  parent_type = "frontend"
  parent_name = "test_frontend"
  index       = 0
}

# Note: TCP response rules are only in the backend, not frontend
# Data source to get the first TCP response rule from the backend
data "haproxy_tcp_response_rule_single" "log_rule" {
  parent_type = "backend"
  parent_name = "test_backend"
  index       = 0
}

# Data source to get the first HTTP check from the backend
data "haproxy_httpcheck_single" "health_check" {
  parent_type = "backend"
  parent_name = "test_backend"
  index       = 0
}

# Data source to get the first TCP check from the backend
data "haproxy_tcp_check_single" "connect_check" {
  parent_type = "backend"
  parent_name = "test_backend"
  index       = 0
}

# Outputs showing how to use the data sources
output "backend_count" {
  description = "Number of backends"
  value       = length(jsondecode(data.haproxy_backends.all_backends.backends))
}

output "test_backend_name" {
  description = "Name of the test backend"
  value       = data.haproxy_backend_single.test_backend.name
}

output "frontend_count" {
  description = "Number of frontends"
  value       = length(jsondecode(data.haproxy_frontends.all_frontends.frontends))
}

output "test_frontend_name" {
  description = "Name of the test frontend"
  value       = data.haproxy_frontend_single.test_frontend.name
}

output "server_count" {
  description = "Number of servers"
  value       = length(jsondecode(data.haproxy_server.all_servers.servers))
}

# Note: Individual server outputs removed because servers web1 and web2 
# don't exist in test_backend. To get server information, parse the JSON from:
# data.haproxy_server.all_servers.servers

output "https_bind_name" {
  description = "Name of the https bind"
  value       = data.haproxy_bind_single.https_bind.name
}

output "is_admin_acl_data" {
  description = "Raw ACL data for is_admin"
  value       = data.haproxy_acl_single.is_admin_acl.acl
}

output "is_api_acl_data" {
  description = "Raw ACL data for is_api"
  value       = data.haproxy_acl_single.is_api_acl.acl
}

output "http_request_rule_data" {
  description = "Raw HTTP request rule data"
  value       = data.haproxy_http_request_rule_single.allow_rule.http_request_rule
}

output "http_response_rule_data" {
  description = "Raw HTTP response rule data"
  value       = data.haproxy_http_response_rule_single.header_rule.http_response_rule
}

output "tcp_request_rule_data" {
  description = "Raw TCP request rule data"
  value       = data.haproxy_tcp_request_rule_single.nice_rule.tcp_request_rule
}

output "tcp_response_rule_data" {
  description = "Raw TCP response rule data"
  value       = data.haproxy_tcp_response_rule_single.log_rule.tcp_response_rule
}

output "httpcheck_data" {
  description = "Raw HTTP check data"
  value       = data.haproxy_httpcheck_single.health_check.httpcheck
}

output "tcp_check_data" {
  description = "Raw TCP check data"
  value       = data.haproxy_tcp_check_single.connect_check.tcp_check
}