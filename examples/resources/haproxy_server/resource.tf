resource "haproxy_server" "server" {
  name               = "example-server"
  address            = "192.168.1.10"
  port               = 8080
  parent_name        = "example-backend"
  parent_type        = "backend"

  depends_on         = [
    haproxy_backend.backend
  ]
}

