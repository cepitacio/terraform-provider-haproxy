provider "haproxy" {
  url      = "http://localhost:5555"
  username = "admin"
  password = "password"
}

resource "haproxy_frontend" "frontend" {
  name            = "test_frontend"
  default_backend = haproxy_backend.backend.name

  bind {
    name    = "bind1"
    address = "*"
    port    = 80
    ssl     = true
    
    # Deprecated fields (API v2) - will be removed in future
    no_sslv3  = true
    no_tlsv10 = true
    no_tlsv11 = true
    
    # New v3 fields (recommended for API v3)
    # sslv3  = false  # Enable SSLv3 (opposite of no_sslv3)
    # tlsv10 = false  # Enable TLSv1.0 (opposite of no_tlsv10)
    # tlsv11 = false  # Enable TLSv1.1 (opposite of no_tlsv11)
    
    # Force specific versions (API v2)
    force_tlsv12 = true
    force_tlsv13 = true
    force_strict_sni = "enabled"
    
    # New v3 fields for version control
    # tlsv12 = true   # Enable TLSv1.2
    # tlsv13 = true   # Enable TLSv1.3
    
    strict_sni = true
  }

  acl {
    index     = 0
    acl_name  = "acl1"
    criterion = "hdr_beg(host)"
    value     = "test.com"
  }

  httprequestrule {
    index     = 0
    type      = "use_backend"
    cond      = "if"
    cond_test = "acl1"
  }

  httpresponserule {
    index     = 0
    type      = "set-header"
    hdr_name  = "X-Forwarded-For"
    hdr_format = "%[src]"
  }

  tcprequestrule {
    index     = 0
    type      = "accept"
    cond      = "if"
    cond_test = "src 127.0.0.1"
  }
}

resource "haproxy_backend" "backend" {
  name = "test_backend"
  mode = "http"
  
  # SSL/TLS Configuration for backend
  ssl = true
  ssl_cafile = "/etc/ssl/certs/ca-certificates.crt"
  ssl_certificate = "/etc/ssl/certs/server.crt"
  ssl_max_ver = "TLSv1.3"
  ssl_min_ver = "TLSv1.2"
  ssl_reuse = "enabled"
  ciphers = "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384"
  ciphersuites = "TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384"
  verify = "required"
  
  # Deprecated fields (API v2) - will be removed in future
  no_sslv3  = true
  no_tlsv10 = true
  no_tlsv11 = true
  
  # New v3 fields (API v3)
  # sslv3  = false  # Enable SSLv3 (opposite of no_sslv3)
  # tlsv10 = false  # Enable TLSv1.0 (opposite of no_tlsv10)
  # tlsv11 = false  # Enable TLSv1.1 (opposite of no_tlsv11)
  
  # Force specific versions (API v2)
  force_tlsv12 = true
  force_tlsv13 = true
  force_strict_sni = "enabled"
  
  # New v3 fields for version control
  # tlsv12 = true   # Enable TLSv1.2
  # tlsv13 = true   # Enable TLSv1.3

  # Note: Servers are managed as separate haproxy_server resources, not nested blocks

  httpcheck {
    index  = 0
    type   = "httpchk"
    uri    = "/health"
    method = "GET"
  }

  stickrule {
    index     = 0
    type      = "match"
    cond      = "if"
    cond_test = "src"
    pattern   = "src"
    table     = "stick_table_1"
  }
}