resource "haproxy_backend" "backend" {
  name           = "backend_test"
  mode           = "http"
  adv_check      = "enabled"
  server_timeout = "30000"

  balance {
    algorithm = "round-robin"
  }

  httpchk_params {
    method = "GET"
    uri    = "/health"
  }

  httpcheck {
    index       = 1
  }
}
