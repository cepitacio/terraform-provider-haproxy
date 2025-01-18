resource "haproxy_frontend" "frontend" {
  name            = "frontend_test"
  backend         = "backend_test"
  mode            = "http"
  monitor_uri     = "/fe-status-check"

  dynamic "monitor_fail" {
    content {
      cond        = "if"
      cond_test   = "nbsrv(frontend_test)"
    }
  }

  dynamic "acl" {
    content {
      acl_name    = "acl_test"
      index       = 1
      parent_name = "frontend_test"
      parent_type = "frontend"
      criterion   = "hdr(host)"
      value       = "example.com"
    }
  }

  dynamic "httprequestrule" {
    content {
      index       = 1
      type        = "allow"
      parent_name = "frontend_test"
      parent_type = "frontend"
    }
  }

  dynamic "httpresponserule" {
    content {
      index       = 1
      type        = "set-header"
      parent_name = "frontend_test"
      parent_type = "frontend"
    }
  }

  depends_on = [
    haproxy_backend.backend
  ]
}
