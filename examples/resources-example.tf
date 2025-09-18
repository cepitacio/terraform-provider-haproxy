# Comprehensive HAProxy Configuration Example
# This is a working example with all supported features

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

resource "haproxy_stack" "test" {
  name = "backend_test_stack"

  backend {
    name = "test_backend"
    mode = "http"
    http_connection_mode = "httpclose"
    server_timeout = 3000
    check_timeout = 2000
    connect_timeout = 5000
    queue_timeout = 10000
    tunnel_timeout = 60000
    tarpit_timeout = 1000
    checkcache = "enabled"
    retries = 3

    default_server {
      ssl = "enabled"
      ssl_cafile = "/etc/haproxy/ssl/test15.awesome.8x8testa.com.pem"
      ssl_certificate = "/etc/haproxy/ssl/test15.awesome.8x8testa.com.pem"
      ssl_max_ver = "TLSv1.3"
      ssl_min_ver = "TLSv1.2"
      ssl_reuse = "enabled"
      ciphers = "ECDHE-RSA-AES256-GCM-SHA384"
      ciphersuites = "TLS_AES_256_GCM_SHA384"
      verify = "required"
      
      # v2 fields (work in both v2 and v3)
      force_sslv3 = "enabled"
      force_tlsv10 = "enabled"
      force_tlsv11 = "enabled"
      force_tlsv12 = "enabled"
      force_tlsv13 = "enabled"
      
      # NOTE: v3 TLS fields (sslv3, tlsv10, tlsv11, tlsv12, tlsv13) have issues in v3 API
      # These fields are not working properly in HAProxy Data Plane API v3
      # Use force_sslv3, force_tlsv10, etc. instead for v3 compatibility
      # sslv3 = "disabled"
      # tlsv10 = "disabled"
      # tlsv11 = "disabled"
      # tlsv12 = "enabled"
      # tlsv13 = "enabled"
    }

    # Individual servers
    servers = {
      web1 = {
        address = "192.168.1.10"
        port    = 8080
        check   = "enabled"
        weight  = 100
        ssl     = "enabled"
        ssl_certificate = "/etc/haproxy/ssl/test15.awesome.8x8testa.com.pem"
        ssl_max_ver = "TLSv1.3"
        ssl_min_ver = "TLSv1.2"
        # v2 fields (work in both v2 and v3)
        force_sslv3 = "disabled"
        force_tlsv10 = "disabled"
        force_tlsv11 = "disabled"
        force_tlsv12 = "enabled"
        force_tlsv13 = "enabled"
      }
      web2 = {
        address = "192.168.1.11"
        port    = 8080
        check   = "enabled"
        weight  = 100
        ssl     = "enabled"
        ssl_certificate = "/etc/haproxy/ssl/test15.awesome.8x8testa.com.pem"
        ssl_max_ver = "TLSv1.3"
        ssl_min_ver = "TLSv1.2"
        # v2 fields (work in both v2 and v3)
        force_sslv3 = "disabled"
        force_tlsv10 = "disabled"
        force_tlsv11 = "disabled"
        force_tlsv12 = "enabled"
        force_tlsv13 = "enabled"
      }
    }

    balance {
      algorithm = "source"
    }

    httpchk_params {
      method = "GET"
      uri = "/health2"
      version = "HTTP/1.1"
    }

    forwardfor {
      enabled = "enabled"
    }

    # ACLs
    acls {
      acl_name = "is_admin_back"
      criterion = "path"
      value = "/admin"
    }
    acls {
      acl_name = "is_static_back"
      criterion = "path"
      value = "/static"
    }
    acls {
      acl_name = "is_public_back"
      criterion = "path"
      value = "/public"
    }
    acls {
      acl_name = "is_api_back"
      criterion = "path"
      value = "/api"
    }

    # HTTP request rules
    http_request_rules {
      type = "allow"
      cond = "if"
      cond_test = "is_admin_back"
    }
    
    http_request_rules {
      type = "set-header"
      cond = "if"
      cond_test = "is_api_back"
    }

    http_request_rules {
      type = "deny"
      cond = "if"
      cond_test = "is_api_back"
    }

    # TCP request rules
    tcp_request_rules {
      type = "content"
      action = "set-var"
      var_name = "backend_var"
      var_scope = "sess"
      expr = "req.hdr(host)"
    }

    tcp_request_rules {
      type = "content"
      action = "set-nice"
      nice_value = 100
    }

    tcp_request_rules {
      type = "content"
      action = "set-mark"
      mark_value = "0x100"
    }

    # HTTP response rules
    http_response_rules {
      type = "set-header"
      hdr_name = "X-Response-Time"
      hdr_format = "100ms"
      cond = "if"
      cond_test = "TRUE"
    }

    http_response_rules {
      type = "set-header"
      hdr_name = "Cache-Control"
      hdr_format = "max-age=360"
      cond = "if"
      cond_test = "TRUE"
    }

    http_response_rules {
      type = "set-log-level"
      log_level = "info"
      cond = "if"
      cond_test = "TRUE"
    }

    http_response_rules {
      type = "return"
      return_status_code = 200
      return_content = "OK"
      cond = "if"
      cond_test = "TRUE"
    }

    # TCP response rules
    tcp_response_rules {
      type = "content"
      action = "set-log-level"
      log_level = "info"
      cond = "if"
      cond_test = "TRUE"
    }

    tcp_response_rules {
      type = "content"
      action = "set-mark"
      mark_value = "0x1"
      cond = "if"
      cond_test = "TRUE"
    }

    tcp_response_rules {
      type = "content"
      action = "set-tos"
      tos_value = "0x100"
      cond = "if"
      cond_test = "TRUE"
    }
    
    tcp_response_rules {
      type = "content"
      action = "set-nice"
      nice_value = 0
      cond = "if"
      cond_test = "TRUE"
    }

    # HTTP checks
    http_checks {
      type = "connect"
      addr = "127.0.0.1"
      port = 80
    }

    http_checks {
      type = "comment"
      check_comment = "Basic HTTP health check"
    }

    http_checks {
      type = "send"
      method = "POST"
      uri = "/api/health"
      version = "HTTP/1.1"
      headers = ["Content-Type: application/json"]
    }

    http_checks {
      type = "expect"
      match = "status"
      pattern = "300"
    }

    # TCP checks
    tcp_checks {
      action = "connect"
      addr = "127.0.0.1"
      port = 80
    }

    tcp_checks {
      action = "expect"
      pattern = "pong"
      match = "string"
    }

    tcp_checks {
      action = "send"
      data = "ping"
    }
  }

  frontend {
    name = "test_frontend"
    mode = "http"
    default_backend = "test_backend"
    maxconn = 10000
    backlog = 100

    # Multiple binds
    binds = {
      complex_ssl_bind = {
        address             = "10.0.10.2"
        port                = 8443
        ssl                 = true
        ssl_certificate     = "/etc/haproxy/ssl/test15.awesome.8x8testa.com.pem"
        ssl_min_ver         = "TLSv1.2"
        ssl_max_ver         = "TLSv1.3"
        ciphers             = "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256"
        ciphersuites        = "TLS_AES_256_GCM_SHA384:TLS_AES_128_GCM_SHA256"
        strict_sni          = true
        prefer_client_ciphers = true
        alpn                = "h2,http/1.1"
        npn                 = "h2,http/1.1"
        allow_0rtt          = true
        transparent         = true
        accept_proxy        = true
        defer_accept        = true
        tfo                 = true
        v4v6                = true
        maxconn             = 2000
        backlog             = "100"
        tcp_user_timeout    = 30000
        gid                 = 1000
        group               = "haproxy"
        interface           = "eth0"
        level               = "admin"
        nice                = 0
        no_ca_names         = false
        severity_output     = "string"
        uid                 = "haproxy"
        user                = "haproxy"
        v6only              = false
        
        # NOTE: v3 TLS fields (sslv3, tlsv10, tlsv11, tlsv12, tlsv13) have issues in v3 API
        # These fields are not working properly in HAProxy Data Plane API v3
        # Use force_sslv3, force_tlsv10, etc. instead for v3 compatibility
        # sslv3               = false
        # tlsv10              = false
        # tlsv11              = false
        # tlsv12              = true
        # tlsv13              = true

        # v2 fields (work in both v2 and v3)
        force_sslv3         = false
        force_tlsv10        = false
        force_tlsv11        = false
        force_tlsv12        = true
        force_tlsv13        = true
      }
      
      https_bind = {
        address = "0.0.0.1"
        port    = 443
      }
      
      admin_bind = {
        address = "127.0.0.2"
        port    = 80
        maxconn = 100
        level   = "admin"
      }
    }

    monitor_fail {
      cond      = "unless"
      cond_test = "{ nbsrv(test_backend) gt 1 }"
    }

    # ACLs
    acls {
      acl_name = "is_admin"
      criterion = "path"
      value = "/admin"
    }
    acls {
      acl_name = "is_public"
      criterion = "path"
      value = "/public"
    }
    acls {
      acl_name = "is_static"
      criterion = "path"
      value = "/static"
    }
    acls {
      acl_name = "is_api"
      criterion = "path"
      value = "/api"
    }

    # HTTP request rules
    http_request_rules {
      type = "allow"
      cond = "if"
      cond_test = "is_admin"
    }
    
    http_request_rules {
      type = "set-header"
      cond = "if"
      cond_test = "is_admin"
    }
    
    http_request_rules {
      type = "deny"
      cond = "if"
      cond_test = "is_api"
    }
    
    # HTTP response rules
    http_response_rules {
      type      = "set-header"
      hdr_name  = "X-Frontend-Response-Time"
      hdr_format = "%T"
    }

    http_response_rules {
      type      = "set-header"
      hdr_name  = "X-Frontend-Server"
      hdr_format = "HAProxy-Frontend"
    }
    
    http_response_rules {
      type      = "set-header"
      hdr_name  = "X-Forwarded-For"
      hdr_format = "%[src]"
    }
    
    http_response_rules {
      type      = "set-header"
      hdr_name  = "X-Client-IP"
      hdr_format = "%[src]"
    }

    # TCP request rules
    tcp_request_rules {
      type = "content"
      action = "set-nice"
      nice_value = 70
    }

    tcp_request_rules {
      type = "content"
      action = "set-var"
      var_name = "frontend_var"
      var_scope = "sess"
      expr = "req.hdr(host)"
    }

    tcp_request_rules {
      type = "content"
      action = "set-mark"
      mark_value = "0x200"
    }
  }
}

# Outputs
output "stack_name" {
  description = "Name of the HAProxy stack"
  value       = haproxy_stack.test.name
}

output "backend_name" {
  description = "Name of the backend"
  value       = haproxy_stack.test.backend.name
}

output "frontend_name" {
  description = "Name of the frontend"
  value       = haproxy_stack.test.frontend.name
}
