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

# Define multiple applications using locals
locals {
  applications = {
    backend_test = {
      backend_name  = "test_backend"
      frontend_name = "test_frontend"
    }
    
    api_test = {
      backend_name  = "api_backend"
      frontend_name = "api_frontend"
    }



    backend_test1 = {
      backend_name  = "test_backend1"
      frontend_name = "test_frontend1"
    }
    
    api_test2 = {
      backend_name  = "api_backend3"
      frontend_name = "api_frontend3"
    }


    backend_test4 = {
      backend_name  = "test_backend4"
      frontend_name = "test_frontend4"
    }
    
    api_test5 = {
      backend_name  = "api_backend5"
      frontend_name = "api_frontend5"
    }


    backend_test6 = {
      backend_name  = "test_backend6"
      frontend_name = "test_frontend6"
    }
    
    api_test7 = {
      backend_name  = "api_backend7"
      frontend_name = "api_frontend7"
    }

    backend_test8 = {
      backend_name  = "test_backend8"
      frontend_name = "test_frontend8"
    }
    
    api_test9 = {
      backend_name  = "api_backend9"
      frontend_name = "api_frontend9"
    }
  }
}

# Create multiple HAProxy stacks using for_each
resource "haproxy_stack" "apps" {
  for_each = local.applications
  
  name = each.key

  backend {
    name = each.value.backend_name
    mode = "http"
    # adv_check = "ssl-hello-chk"
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
      ssl = "enabled"                    # Use "enabled" instead of true
      ssl_cafile = "/etc/haproxy/ssl/cert.pem"
      ssl_certificate = "/etc/haproxy/ssl/cert.pem"
      ssl_max_ver = "TLSv1.3"
      ssl_min_ver = "TLSv1.2"
      ssl_reuse = "enabled"
      ciphers = "ECDHE-RSA-AES256-GCM-SHA384"
      ciphersuites = "TLS_AES_256_GCM_SHA384"
      verify = "required"
      force_sslv3 = "enabled"               # Use "enabled" instead of true
      # force_tlsv10 = "enabled"              # Use "enabled" instead of true
      # force_tlsv11 = "enabled"              # Use "enabled" instead of true
      # force_tlsv12 = "enabled"             # Use "disabled" instead of false
      # force_tlsv13 = "enabled"             # Use "disabled" instead of false
      # # Protocol control (v3) - Use strings
      # sslv3 = "disabled"                 # Use "disabled" instead of false
      # tlsv10 = "enabled"                # Use "disabled" instead of false
      # tlsv11 = "disabled"                # Use "disabled" instead of false
      # tlsv12 = "enabled"                 # Use "enabled" instead of true
      # tlsv13 = "disabled"                 # Use "enabled" instead of true
      
      #force_strict_sni = "enabled"

      # Deprecated fields (v2) - Use strings
      # no_sslv3 = "enabled"               # Use "enabled" instead of true
      # no_tlsv10 = "enabled"              # Use "enabled" instead of true
      # no_tlsv11 = "enabled"              # Use "enabled" instead of true
      # no_tlsv12 = "enabled"             # Use "disabled" instead of false
      # no_tlsv13 = "enabled"             # Use "disabled" instead of false
      
    }
    servers = {
      test_server2 = {
        address = "127.0.0.1"
        port = 8081
    }
      test_server = {
        address = "127.0.0.1"
        port = 8080
        check = "enabled"
        backup = "disabled"
        maxconn = 2000
        weight = 200
        rise = 2
        fall = 3
        inter = 5000
        fastinter = 1000
        downinter = 5000
        ssl = "enabled"
        verify = "none"
        cookie = "test_cookie"
      }
    }
    #These balance, httpchk_params and forwardfor commented out not working
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

    http_request_rules {
      type = "allow"
      cond = "if"
      cond_test = "is_admin_back"
    }

    http_request_rules {
      type = "deny"
      cond = "if"
      cond_test = "is_api_back"
    }

    http_request_rules {
      type = "set-header"
      cond = "if"
      cond_test = "is_api_back"
    }


    # Let's test with just the working actions first
    tcp_request_rules {
      type = "content"
      action = "set-var"
      var_name = "backend_var"
      var_scope = "sess"
      expr = "req.hdr(host)"
    }

    tcp_request_rules {
      type = "content"
      action = "set-mark"
      mark_value = "0x100"
    }

    tcp_request_rules {
      type = "content"
      action = "set-nice"
      nice_value = 100
    }




    # Example 1: Set response header (always)
    http_response_rules {
      type = "set-header"
      hdr_name = "X-Response-Time"
      hdr_format = "100ms"
      cond = "if"
      cond_test = "TRUE"
    }

    # Example 2: Set cache control header (always)
    http_response_rules {
      type = "set-header"
      hdr_name = "Cache-Control"
      hdr_format = "max-age=360"
      cond = "if"
      cond_test = "TRUE"
    }

    # Example 4: Set log level (always)
    http_response_rules {
      type = "set-log-level"
      log_level = "info"
      cond = "if"
      cond_test = "TRUE"
    }

    # Example 3: Set custom status (always)
    http_response_rules {
      type = "return"
      return_status_code = 200
      return_content = "OK"
      cond = "if"
      cond_test = "TRUE"
    }



    # Index 0: Set log level (VALID action)
    tcp_response_rules {
      type = "content"
      action = "set-log-level"
      log_level = "info"
      cond = "if"
      cond_test = "TRUE"
    }

    # Index 1: Set connection mark (VALID action)
    tcp_response_rules {
      type = "content"
      action = "set-mark"
      mark_value = "0x1"
      cond = "if"
      cond_test = "TRUE"
    }

    # Index 2: Set connection priority (VALID action)
    tcp_response_rules {
      type = "content"
      action = "set-nice"
      nice_value = 0
      cond = "if"
      cond_test = "TRUE"
    }


    # Index 3: Set TOS value (VALID action)
    tcp_response_rules {
      type = "content"
      action = "set-tos"
      tos_value = "0x100"
      cond = "if"
      cond_test = "TRUE"
    }

    http_checks {
      type = "connect"
      addr = "127.0.0.1"
      port = 80
    }

    http_checks {
      type = "comment"
      check_comment = "Basic HTTP health check2"
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

    # TCP Checks (simplified for testing):
    tcp_checks {
      action = "connect"
      addr = "127.0.0.1"
      port = 80
    }
    tcp_checks {
      action = "send"
      data = "ping"
    }
    tcp_checks {
      action = "expect"
      pattern = "pong"
      match = "string"
    }
    

  
  }


  frontend {
    name = each.value.frontend_name
    mode = "http"
    default_backend = each.value.backend_name
    maxconn = 10000
    backlog = 100
    binds = {
      complex_ssl_bind = {
        address             = "10.0.10.2"
        port                = 8443
        ssl                 = true
        ssl_certificate     = "/etc/haproxy/ssl/cert.pem"
        # ssl_cafile          = "/etc/haproxy/ssl/cert.pem"
        ssl_min_ver         = "TLSv1.2"
        ssl_max_ver         = "TLSv1.3"
        ciphers             = "ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256"
        ciphersuites        = "TLS_AES_256_GCM_SHA384"#:TLS_AES_128_GCM_SHA256"
        # verify              = "required"
        strict_sni          = true
        prefer_client_ciphers = true
        alpn                = "h2,http/1.1"
        npn                 = "h2,http/1.1"
        allow_0rtt          = true
        # tls_ticket_keys     = "/etc/ssl/ticket-keys"
        transparent         = true
        accept_proxy        = true
        defer_accept        = true
        tfo                 = true
        v4v6                = true
        maxconn             = 2000
        backlog             = "100"
        tcp_user_timeout    = 30000
        # generate_certificates = true
        gid                 = 1000
        group               = "haproxy"
        # id                  = "ssl_bind_1"
        interface           = "eth0"
        level               = "admin"
        # namespace           = "default"
        nice                = 0
        no_ca_names         = false
        # process             = "thread"
        # proto               = "tcp"
        severity_output     = "string"
        uid                 = "haproxy"
        user                = "haproxy"
        v6only              = false
        
        # v3 fields (if using v3 API)
        # sslv3               = false
        # tlsv10              = false
        # tlsv11              = false
        # tlsv12              = true
        # tlsv13              = true
        # tls_tickets         = "enabled"
        # force_strict_sni    = "enabled"
        # no_strict_sni       = false
        # # # guid_prefix         = "haproxy"
        # idle_ping           = 30
        # quic_cc_algo        = "cubic"
        # quic_force_retry    = true
        # quic_cc_algo_burst_size = 10
        # quic_cc_algo_max_window = 1000000
        # metadata            = "complex-ssl-example"
        
        # v2 fields (if using v2 API)
        force_sslv3         = false
        force_tlsv10        = false
        force_tlsv11        = false
        force_tlsv12        = true
        force_tlsv13        = true
        # no_sslv3            = true
        # no_tlsv10           = true
        # no_tlsv11           = true
        no_tlsv12           = true
        # no_tlsv13           = false
        # no_tls_tickets      = false
      }
      # HTTPS bind
      https_bind = {
        address        = "0.0.0.1"
        port           = 443
      }
      # Admin bind (localhost only)
      admin_bind = {
        address     = "127.0.0.2"
        port        = 80
        maxconn     = 100
        level       = "admin"
      }
    }
    monitor_fail {
      cond      = "unless"
      cond_test = "{ nbsrv(${each.value.backend_name}) gt 1 }"
    }
    acls {
      acl_name = "is_admin"
      criterion = "path"
      value = "/admin"
    }
    acls {
      acl_name = "is_public2"
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

    http_request_rules {
      type = "allow"
      cond = "if"
      cond_test = "is_admin"
    }
    http_request_rules {
      type = "deny"
      cond = "if"
      cond_test = "is_api"
    }
    http_request_rules {
      type = "set-header"
      cond = "if"
      cond_test = "is_admin"
    }
    
    # Frontend HTTP Response Rules
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
    tcp_request_rules {
      type = "content"
      action = "set-nice"
      nice_value = 70
    }

   tcp_request_rules {
      type = "content"
      action = "set-mark"
      mark_value = "0x200"
    }

   tcp_request_rules {
      type = "content"
      action = "set-var"
      var_name = "frontend_var"
      var_scope = "sess"
      expr = "req.hdr(host)"
    }



  }
}