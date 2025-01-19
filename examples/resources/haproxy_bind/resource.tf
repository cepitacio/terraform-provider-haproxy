resource "haproxy_bind" "bind" {
  name            = "bind_test"
  port            = 80
  address         = "192.168.1.5"
  parent_name     = "frontend_test"
  parent_type     = "frontend"

  depends_on         = [
    haproxy_frontend.front_end
  ]
}
