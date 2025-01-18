resource "haproxy_bind" "bind" {
  name            = "bind_test"
  port            = 80
  address         = "192.168.1.5"

  depends_on         = [
    haproxy_frontend.front_end
  ]
}
