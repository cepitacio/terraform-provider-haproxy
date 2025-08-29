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

# Comprehensive HAProxy Stack with Full Backend Configuration
resource "haproxy_stack" "production_stack" {
  name = "production_stack"
  
  # Comprehensive Backend Configuration
  backend {
    name = "app_backend"
    mode = "http"
    
    # Timeouts
    server_timeout = 30000      # 30 seconds
    check_timeout = 2000        # 2 seconds
    connect_timeout = 5000      # 5 seconds
    queue_timeout = 10000       # 10 seconds
    tunnel_timeout = 60000      # 1 minute
    tarpit_timeout = 1000      # 1 second
    
    # Health Check Configuration
    adv_check = "ssl-hello-chk"
    checkcache = "enabled"
    retries = 3
    
    # HTTP Connection Mode
    http_connection_mode = "http-keep-alive"
    
    # SSL/TLS Configuration
    ssl = true
    ssl_cafile = "/etc/ssl/certs/ca-bundle.crt"
    ssl_certificate = "/etc/ssl/certs/app.crt"
    ssl_max_ver = "TLSv1.3"
    ssl_min_ver = "TLSv1.2"
    ssl_reuse = "enabled"
    ciphers = "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256"
    ciphersuites = "TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256"
    verify = "required"
    
    # SSL/TLS Protocol Support (v3 fields - non-deprecated)
    sslv3 = false   # Disable SSLv3
    tlsv10 = false  # Disable TLSv1.0
    tlsv11 = false  # Disable TLSv1.1
    tlsv12 = true   # Enable TLSv1.2
    tlsv13 = true   # Enable TLSv1.3
    
    # SSL/TLS Protocol Support (v2 fields - deprecated but supported)
    no_sslv3 = true     # Disable SSLv3 (deprecated)
    no_tlsv10 = true    # Disable TLSv1.0 (deprecated)
    no_tlsv11 = true    # Disable TLSv1.1 (deprecated)
    no_tlsv12 = false   # Enable TLSv1.2 (deprecated)
    no_tlsv13 = false   # Enable TLSv1.3 (deprecated)
    force_strict_sni = "enabled"
    
    # Load Balancing Configuration
    balance {
      algorithm = "roundrobin"
      url_param = "id"
    }
    
    # HTTP Health Check Parameters
    httpchk_params {
      method = "GET"
      uri = "/health"
      version = "HTTP/1.1"
    }
    
    # Forward For Configuration
    forwardfor {
      enabled = "enabled"
    }
    
    # Multiple HTTP Health Checks
    httpcheck {
      index = 1
      type = "connect"
      port = 8080
      timeout = 5000
      match = "status"
      pattern = "200"
      addr = "127.0.0.1"
      log_level = "info"
      send_proxy = "enabled"
      check_comment = "Main health check"
    }
    
    httpcheck {
      index = 2
      type = "request"
      method = "GET"
      uri = "/api/health"
      version = "HTTP/1.1"
      timeout = 3000
      match = "rstatus"
      pattern = "2[0-9][0-9]"
      log_level = "info"
      check_comment = "API health check"
    }
    
    # TCP Health Checks
    tcp_check {
      index = 1
      type = "connect"
      action = "accept"
      cond = "if"
      cond_test = "is_ssl"
    }
    
    tcp_check {
      index = 2
      type = "send"
      action = "reject"
      cond = "unless"
      cond_test = "is_ssl"
    }
    
    # Access Control Lists
    acl {
      acl_name = "local_network"
      index = 1
      criterion = "src"
      value = "192.168.1.0/24"
    }
    
    acl {
      acl_name = "admin_users"
      index = 2
      criterion = "req.hdr"
      value = "X-User-Id"
    }
    
    # HTTP Request Rules
    http_request_rule {
      index = 1
      type = "allow"
      cond = "if"
      cond_test = "local_network"
    }
    
    http_request_rule {
      index = 2
      type = "redirect"
      redir_type = "prefix"
      redir_value = "/api"
      cond = "if"
      cond_test = "path_beg /v1"
    }
    
    http_request_rule {
      index = 3
      type = "set-header"
      hdr_name = "X-Proxy-By"
      hdr_format = "HAProxy"
    }
    
    # HTTP Response Rules
    http_response_rule {
      index = 1
      type = "set-header"
      hdr_name = "X-Cache-Control"
      hdr_format = "no-cache"
      cond = "if"
      cond_test = "status 500"
    }
    
    # TCP Request Rules
    tcp_request_rule {
      index = 1
      type = "content"
      action = "accept"
      cond = "if"
      cond_test = "req_len gt 0"
    }
    
    # TCP Response Rules
    tcp_response_rule {
      index = 1
      type = "content"
      action = "accept"
      cond = "if"
      cond_test = "res_len gt 0"
    }
    
    # Sticky Session Table
    stick_table {
      type = "ip"
      size = 10000
      expire = 300000  # 5 minutes
      nopurge = false
      peers = "haproxy_peers"
    }
    
    # Sticky Session Rules
    stick_rule {
      index = 1
      type = "match"
      table = "haproxy_peers"
      pattern = "src"
    }
    
    stick_rule {
      index = 2
      type = "store-request"
      table = "haproxy_peers"
      pattern = "src"
    }
    
    # Statistics Options
    stats_options {
      stats_enable = true
      stats_uri = "/stats"
      stats_realm = "HAProxy Statistics"
      stats_auth = "admin:admin123"
    }
  }
  
  # Simple Server Configuration
  server {
    name = "app_server_1"
    address = "192.168.1.10"
    port = 8080
    check = "enabled"
    weight = 100
    rise = 2
    fall = 3
    inter = 2000
    fastinter = 1000
    downinter = 5000
    maxconn = 1000
    ssl = "enabled"
    verify = "required"
    cookie = "server1"
    disabled = false
  }
  
  # Simple Frontend Configuration
  frontend {
    name = "app_frontend"
    mode = "http"
    default_backend = "app_backend"
    maxconn = 5000
    ssl = true
    ssl_certificate = "/etc/ssl/certs/app.crt"
    ssl_min_ver = "TLSv1.2"
    tlsv12 = true
    tlsv13 = true
    
    bind {
      name = "http_bind"
      address = "0.0.0.0"
      port = 80
    }
    
    bind {
      name = "https_bind"
      address = "0.0.0.0"
      port = 443
      ssl = true
    }
  }
}

# Outputs to verify the configuration
output "stack_name" {
  value = haproxy_stack.production_stack.name
}

output "backend_name" {
  value = haproxy_stack.production_stack.backend.name
}

output "backend_mode" {
  value = haproxy_stack.production_stack.backend.mode
}

output "backend_ssl" {
  value = haproxy_stack.production_stack.backend.ssl
}

output "server_name" {
  value = haproxy_stack.production_stack.server.name
}

output "frontend_name" {
  value = haproxy_stack.production_stack.frontend.name
}
