# haproxy_frontend

Manages a frontend in HAProxy.

## Example Usage

```hcl
resource "haproxy_frontend" "example" {
  name            = "example"
  default_backend = "example"

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
```

## Argument Reference

- `name` - (Required) The name of the frontend.
- `default_backend` - (Required) The name of the default backend.
- `http_connection_mode` - (Optional) The HTTP connection mode.
- `accept_invalid_http_request` - (Optional) Whether to accept invalid HTTP requests.
- `maxconn` - (Optional) The maximum number of concurrent connections.
- `mode` - (Optional) The mode of the frontend.
- `backlog` - (Optional) The backlog size.
- `http_keep_alive_timeout` - (Optional) The HTTP keep-alive timeout.
- `http_request_timeout` - (Optional) The HTTP request timeout.
- `http_use_proxy_header` - (Optional) Whether to use the Proxy Protocol header.
- `httplog` - (Optional) Whether to enable HTTP logging.
- `httpslog` - (Optional) Whether to enable HTTPS logging.
- `error_log_format` - (Optional) The error log format.
- `log_format` - (Optional) The log format.
- `log_format_sd` - (Optional) The log format for structured data.
- `monitor_uri` - (Optional) The monitor URI.
- `tcplog` - (Optional) Whether to enable TCP logging.
- `from` - (Optional) The source address for the frontend.
- `client_timeout` - (Optional) The client timeout.
- `http_use_htx` - (Optional) Whether to use HTX.
- `http_ignore_probes` - (Optional) Whether to ignore probes.
- `log_tag` - (Optional) The log tag.
- `clflog` - (Optional) Whether to enable CLF logging.
- `contstats` - (Optional) Whether to enable continuous statistics.
- `dontlognull` - (Optional) Whether to prevent logging of null connections.
- `log_separate_errors` - (Optional) Whether to log errors separately.
- `option_http_server_close` - (Optional) Whether to enable the `http-server-close` option.
- `option_httpclose` - (Optional) Whether to enable the `httpclose` option.
- `option_http_keep_alive` - (Optional) Whether to enable the `http-keep-alive` option.
- `option_dontlog_normal` - (Optional) Whether to prevent logging of normal connections.
- `option_logasap` - (Optional) Whether to enable the `logasap` option.
- `option_tcplog` - (Optional) Whether to enable the `tcplog` option.
- `option_socket_stats` - (Optional) Whether to enable socket statistics.
- `option_forwardfor` - (Optional) Whether to enable the `forwardfor` option.
- `timeout_client` - (Optional) The client timeout.
- `timeout_http_keep_alive` - (Optional) The HTTP keep-alive timeout.
- `timeout_http_request` - (Optional) The HTTP request timeout.
- `timeout_cont` - (Optional) The continuous timeout.
- `timeout_tarpit` - (Optional) The tarpit timeout.
- `stats_options` - (Optional) A block to configure statistics options.
  - `stats_enable` - (Optional) Whether to enable statistics.
  - `stats_hide_version` - (Optional) Whether to hide the HAProxy version.
  - `stats_show_legends` - (Optional) Whether to show legends.
  - `stats_show_node` - (Optional) Whether to show the node name.
  - `stats_uri` - (Optional) The statistics URI.
  - `stats_realm` - (Optional) The statistics realm.
  - `stats_auth` - (Optional) The statistics authentication.
  - `stats_refresh` - (Optional) The statistics refresh interval.
- `monitor_fail` - (Optional) A block to configure monitor fail conditions.
  - `cond` - (Required) The condition.
  - `cond_test` - (Required) The condition test.
- `bind` - (Optional) A block to configure a bind.
  - `name` - (Required) The name of the bind.
  - `address` - (Required) The address of the bind.
  - `port` - (Optional) The port of the bind.
  - `transparent` - (Optional) Whether to enable transparent binding.
  - `mode` - (Optional) The mode of the bind.
  - `maxconn` - (Optional) The maximum number of concurrent connections for the bind.
  - `user` - (Optional) The user for the bind.
  - `group` - (Optional) The group for the bind.
  - `force_sslv3` - (Optional) Whether to force SSLv3.
  - `force_tlsv10` - (Optional) Whether to force TLSv1.0.
  - `force_tlsv11` - (Optional) Whether to force TLSv1.1.
  - `force_tlsv12` - (Optional) Whether to force TLSv1.2.
  - `force_tlsv13` - (Optional) Whether to force TLSv1.3.
  - `ssl` - (Optional) Whether to enable SSL.
  - `ssl_cafile` - (Optional) The SSL CA file.
  - `ssl_max_ver` - (Optional) The maximum SSL version to support.
  - `ssl_min_ver` - (Optional) The minimum SSL version to support.
  - `ssl_certificate` - (Optional) The SSL certificate.
  - `ciphers` - (Optional) The supported ciphers.
  - `ciphersuites` - (Optional) The supported ciphersuites.
- `acl` - (Optional) A block to configure an ACL.
  - `index` - (Required) The index of the ACL.
  - `acl_name` - (Required) The name of the ACL.
  - `criterion` - (Required) The criterion of the ACL.
  - `value` - (Required) The value of the ACL.
- `httprequestrule` - (Optional) A block to configure an HTTP request rule.
  - `index` - (Required) The index of the HTTP request rule.
  - `type` - (Required) The type of the HTTP request rule.
  - `cond` - (Optional) The condition of the HTTP request rule.
  - `cond_test` - (Optional) The condition test of the HTTP request rule.
  - `hdr_name` - (Optional) The header name of the HTTP request rule.
  - `hdr_format` - (Optional) The header format of the HTTP request rule.
  - `redir_type` - (Optional) The redirection type of the HTTP request rule.
  - `redir_value` - (Optional) The redirection value of the HTTP request rule.
- `httpresponserule` - (Optional) A block to configure an HTTP response rule.
  - `index` - (Required) The index of the HTTP response rule.
  - `type` - (Required) The type of the HTTP response rule.
  - `cond` - (Optional) The condition of the HTTP response rule.
  - `cond_test` - (Optional) The condition test of the HTTP response rule.
  - `hdr_name` - (Optional) The header name of the HTTP response rule.
  - `hdr_format` - (Optional) The header format of the HTTP response rule.
  - `redir_type` - (Optional) The redirection type of the HTTP response rule.
  - `redir_value` - (Optional) The redirection value of the HTTP response rule.
- `tcprequestrule` - (Optional) A block to configure a TCP request rule.
  - `index` - (Required) The index of the TCP request rule.
  - `type` - (Required) The type of the TCP request rule.
  - `action` - (Optional) The action of the TCP request rule.
  - `cond` - (Optional) The condition of the TCP request rule.
  - `cond_test` - (Optional) The condition test of the TCP request rule.
- `tcpresponserule` - (Optional) A block to configure a TCP response rule.
  - `index` - (Required) The index of the TCP response rule.
  - `action` - (Required) The action of the TCP response rule.
  - `cond` - (Optional) The condition of the TCP response rule.
  - `cond_test` - (Optional) The condition test of the TCP response rule.