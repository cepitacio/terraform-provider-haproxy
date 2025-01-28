resource "haproxy_frontend" "frontend" {
  name            = "frontend_test"
  default_backend = "backend_test"
  mode            = "http"
  monitor_uri     = "/fe-status-check"

  monitor_fail {
      cond        = "if"
      cond_test   = "acl_test"
  }

  acl {
    acl_name    = "acl_test"
    index       = 0
    criterion   = "nbsrv(backend_test)"
    value       = "lt 1"
  }

  acl {
    acl_name    = "denied_ips"
    index       = 0
    criterion   = src
    value       = ""
  }

  httprequestrule {
    index       = 0
    type        = "deny"
    cond        = "if"
    cond_test   = "denied_ips"
  }

  httpresponserule {
    index       = 0
    type        = "set-header"
    hdr_name    = "Strict-Transport-Security"
    hdr_format   = "foo"
  }

  depends_on = [
    haproxy_backend.backend
  ]
}
