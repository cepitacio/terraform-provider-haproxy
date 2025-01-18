resource "haproxy_frontend" "frontend" {
  name            = "frontend_test"
  backend         = "backend_test"
  mode            = "http"
  monitor_uri     = "/fe-status-check"

  monitor_fail {
      cond        = "if"
      cond_test   = "nbsrv(frontend_test)"
  }

  acl {
    acl_name    = "acl_test"
    index       = 1
    criterion   = "hdr(host)"
    value       = "example.com"
  }

  httprequestrule {
    index       = 1
    type        = "allow"
  }

  httpresponserule {
    index       = 1
    type        = "set-header"
  }

  depends_on = [
    haproxy_backend.backend
  ]
}
