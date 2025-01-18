resource "haproxy_backend" "backend" {
  name           = "backend_test"
  mode           = "http"
  adv_check      = "enabled"
  server_timeout = "30000"

  balance {
    algorithm = "round-robin"
  }

  dynamic "httpchk_params" {
    content {
      method = "GET"
      uri    = "/health"
    }
  }

  dynamic "httpcheck" {
    content {
      parent_name = "backend_test"
      parent_type = "backend"
      index       = 1
    }
  }
}
