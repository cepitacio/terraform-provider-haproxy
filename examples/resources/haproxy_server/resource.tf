resource "haproxy_server" "server" {
  name               = "example-server"
  address            = "192.168.1.10"
  port               = 8080
  parent_name        = "example-backend"
  check              = "enabled"
  check_ssl          = "enabled"
  parent_type        = "backend"
  health_check_port  = 8081
  fall               = 3
  rise               = 2
  inter              = 5000
  ssl                = "enabled"
  verify             = "none"
  ssl_certificate    = "/etc/haproxy/certs/example-server.pem"
  ssl_cafile         = "/etc/haproxy/ca.crt"
  ciphersuites       = "TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256"
  force_sslv3        = "disabled"
  force_tlsv10       = "disabled"
  force_tlsv11       = "enabled"
  force_tlsv12       = "enabled"
  force_tlsv13       = "enabled"

  depends_on         = [
    haproxy_backend.backend
  ]
}

