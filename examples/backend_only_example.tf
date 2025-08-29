resource "haproxy_stack" "backend_test" {
  backend {
    name = "test_backend"
    mode = "http"
    adv_check = "ssl-hello-chk"
    http_connection_mode = "http-keep-alive"
    server_timeout = 30000
    check_timeout = 2000
    connect_timeout = 5000
    queue_timeout = 10000
    tunnel_timeout = 60000
    tarpit_timeout = 1000
    checkcache = "enabled"
    retries = 3

    default_server {
      ssl = true
      ssl_cafile = "/etc/ssl/certs/ca-bundle.crt"
      ssl_certificate = "/etc/ssl/certs/app.crt"
      ssl_max_ver = "TLSv1.3"
      ssl_min_ver = "TLSv1.2"
      ssl_reuse = "enabled"
      ciphers = "ECDHE-RSA-AES256-GCM-SHA384"
      ciphersuites = "TLS_AES_256_GCM_SHA384"
      verify = "required"
      
      # Protocol control (v3)
      sslv3 = false
      tlsv10 = false
      tlsv11 = false
      tlsv12 = true
      tlsv13 = true
      
      # Deprecated fields (v2)
      no_sslv3 = true
      no_tlsv10 = true
      no_tlsv11 = true
      no_tlsv12 = false
      no_tlsv13 = false
      
      force_strict_sni = "enabled"
    }

    balance {
      algorithm = "roundrobin"
    }

    httpchk_params {
      method = "GET"
      uri = "/health"
      version = "HTTP/1.1"
    }

    forwardfor {
      enabled = "enabled"
    }
  }

  server {
    name = "test_server"
    address = "127.0.0.1"
    port = 8080
    check = "enabled"
    backup = "disabled"
    maxconn = 1000
    weight = 100
    rise = 2
    fall = 3
    inter = 2000
    fastinter = 1000
    downinter = 5000
    ssl = "enabled"
    verify = "required"
    cookie = "test_cookie"
    disabled = false
  }

  frontend {
    name = "test_frontend"
    mode = "http"
    default_backend = "test_backend"
    maxconn = 10000
    backlog = 100
  }
}