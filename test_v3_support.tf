terraform {
  required_providers {
    haproxy = {
      source = "local/haproxy/haproxy"
      version = "~> 1.0"
    }
  }
}

provider "haproxy" {
  url          = "http://10.0.10.29:5555"
  username     = "admin"
  password     = "admin"
  api_version  = "v3"  # This will enable v3-specific fields
}

# Test v3-specific fields
resource "haproxy_stack" "v3_test" {
  name = "v3_test_stack"

  backend {
    name = "v3_backend"
    mode = "http"
    
    default_server {
      # v3-specific fields (these will only be available with api_version = "v3")
      sslv3 = "disabled"   # v3 field
      tlsv10 = "disabled"  # v3 field
      tlsv11 = "disabled"  # v3 field
      tlsv12 = "enabled"   # v3 field
      tlsv13 = "enabled"   # v3 field
      
      # Common fields (available in both v2 and v3)
      ssl = "enabled"
      ssl_certificate = "/etc/ssl/certs/server.crt"
      ssl_cafile = "/etc/ssl/certs/ca.crt"
    }
  }

  server {
    name    = "v3_server"
    address = "192.168.1.10"
    port    = 8080
    check   = "enabled"
    weight  = 100
    
    # v3-specific fields
    sslv3 = "disabled"   # v3 field
    tlsv10 = "disabled"  # v3 field
    tlsv11 = "disabled"  # v3 field
    tlsv12 = "enabled"   # v3 field
    tlsv13 = "enabled"   # v3 field
  }

  frontend {
    name = "v3_frontend"
    mode = "http"
    default_backend = "v3_backend"
    
    bind {
      name    = "v3_bind"
      address = "0.0.0.0"
      port    = 80
      
      # v3-specific fields
      sslv3 = false   # v3 field
      tlsv10 = false  # v3 field
      tlsv11 = false  # v3 field
      tlsv12 = true   # v3 field
      tlsv13 = true   # v3 field
    }
  }
}
