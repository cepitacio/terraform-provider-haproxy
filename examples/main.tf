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
  }

  acl {
    index     = 0
    acl_name  = "acl1"
    criterion = "hdr_beg(host)"
    value     = "test.com"
  }

  httprequestrule {
    index = 0
    type  = "use_backend"
    cond  = "if"
    cond_test = "acl1"
  }
}

resource "haproxy_backend" "backend" {
  name = "test_backend"
  mode = "http"

  server {
    name    = "server1"
    address = "127.0.0.1"
    port    = 8080
  }
}