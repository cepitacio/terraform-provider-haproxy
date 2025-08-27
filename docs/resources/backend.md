# haproxy_backend

Manages a backend in HAProxy.

## Example Usage

```hcl
resource "haproxy_backend" "example" {
  name = "example"
  mode = "http"

  server {
    name    = "server1"
    address = "127.0.0.1"
    port    = 8080
  }
}
```

## Argument Reference

- `name` - (Required) The name of the backend.
- `mode` - (Optional) The mode of the backend.
- `adv_check` - (Optional) The advanced check for the backend.
- `http_connection_mode` - (Optional) The HTTP connection mode.
- `server_timeout` - (Optional) The server timeout.
- `check_timeout` - (Optional) The check timeout.
- `connect_timeout` - (Optional) The connect timeout.
- `queue_timeout` - (Optional) The queue timeout.
- `tunnel_timeout` - (Optional) The tunnel timeout.
- `tarpit_timeout` - (Optional) The tarpit timeout.
- `check_cache` - (Optional) Whether to cache checks.
- `retries` - (Optional) The number of retries.
- `balance` - (Optional) A block to configure load balancing.
  - `algorithm` - (Optional) The load balancing algorithm.
  - `url_param` - (Optional) The URL parameter to use for the `url_param` algorithm.
- `httpchk_params` - (Optional) A block to configure HTTP check parameters.
  - `method` - (Optional) The HTTP method to use for the check.
  - `uri` - (Optional) The URI to use for the check.
  - `version` - (Optional) The HTTP version to use for the check.
- `forwardfor` - (Optional) A block to configure the `forwardfor` option.
  - `enabled` - (Optional) Whether to enable the `forwardfor` option.
- `server` - (Optional) A block to configure a server.
  - `name` - (Required) The name of the server.
  - `address` - (Required) The address of the server.
  - `port` - (Required) The port of the server.
  - `agent_addr` - (Optional) The agent address of the server.
  - `agent_check` - (Optional) The agent check of the server.
  - `agent_inter` - (Optional) The agent inter of the server.
  - `agent_port` - (Optional) The agent port of the server.
  - `agent_send` - (Optional) The agent send of the server.
  - `allow_0rtt` - (Optional) The allow 0rtt of the server.
  - `alpn` - (Optional) The alpn of the server.
  - `backup` - (Optional) The backup of the server.
  - `check` - (Optional) Whether to enable health checks.
  - `check_alpn` - (Optional) The check alpn of the server.
  - `check_sni` - (Optional) The check sni of the server.
  - `check_ssl` - (Optional) Whether to enable SSL health checks.
  - `check_via_socks4` - (Optional) The check via socks4 of the server.
  - `ciphers` - (Optional) The supported ciphers.
  - `ciphersuites` - (Optional) The supported ciphersuites.
  - `cookie` - (Optional) The cookie of the server.
  - `crt` - (Optional) The crt of the server.
  - `downinter` - (Optional) The downinter of the server.
  - `error_limit` - (Optional) The error limit of the server.
  - `fall` - (Optional) The number of unsuccessful health checks required to consider the server as failed.
  - `fastinter` - (Optional) The fastinter of the server.
  - `force_sslv3` - (Optional) Whether to force SSLv3.
  - `force_tlsv10` - (Optional) Whether to force TLSv1.0.
  - `force_tlsv11` - (Optional) Whether to force TLSv1.1.
  - `force_tlsv12` - (Optional) Whether to force TLSv1.2.
  - `force_tlsv13` - (Optional) Whether to force TLSv1.3.
  - `health_check_port` - (Optional) The health check port.
  - `init_addr` - (Optional) The init addr of the server.
  - `inter` - (Optional) The interval between health checks.
  - `maintenance` - (Optional) The maintenance of the server.
  - `maxconn` - (Optional) The maxconn of the server.
  - `maxqueue` - (Optional) The maxqueue of the server.
  - `minconn` - (Optional) The minconn of the server.
  - `no_sslv3` - (Optional) The no sslv3 of the server.
  - `no_tlsv10` - (Optional) The no tlsv10 of the server.
  - `no_tlsv11` - (Optional) The no tlsv11 of the server.
  - `no_tlsv12` - (Optional) The no tlsv12 of the server.
  - `no_tlsv13` - (Optional) The no tlsv13 of the server.
  - `on_error` - (Optional) The on error of the server.
  - `on_marked_down` - (Optional) The on marked down of the server.
  - `on_marked_up` - (Optional) The on marked up of the server.
  - `pool_low_conn` - (Optional) The pool low conn of the server.
  - `pool_max_conn` - (Optional) The pool max conn of the server.
  - `pool_purge_delay` - (Optional) The pool purge delay of the server.
  - `proto` - (Optional) The proto of the server.
  - `proxy_v2_options` - (Optional) The proxy v2 options of the server.
  - `rise` - (Optional) The number of successful health checks required to consider the server as operational.
  - `send_proxy` - (Optional) Whether to send the Proxy Protocol header.
  - `send_proxy_v2` - (Optional) The send proxy v2 of the server.
  - `send_proxy_v2_ssl` - (Optional) The send proxy v2 ssl of the server.
  - `send_proxy_v2_ssl_cn` - (Optional) The send proxy v2 ssl cn of the server.
  - `slowstart` - (Optional) The slowstart of the server.
  - `sni` - (Optional) The sni of the server.
  - `source` - (Optional) The source of the server.
  - `ssl` - (Optional) Whether to enable SSL.
  - `ssl_cafile` - (Optional) The SSL CA file.
  - `ssl_certificate` - (Optional) The SSL certificate.
  - `ssl_max_ver` - (Optional) The maximum SSL version to support.
  - `ssl_min_ver` - (Optional) The minimum SSL version to support.
  - `ssl_reuse` - (Optional) Whether to reuse SSL connections.
  - `stick` - (Optional) The stick of the server.
  - `tfo` - (Optional) The tfo of the server.
  - `tls_tickets` - (Optional) The tls tickets of the server.
  - `track` - (Optional) The track of the server.
  - `verify` - (Optional) The certificate verification mode.
  - `weight` - (Optional) The weight of the server.
- `httpcheck` - (Optional) A block to configure an HTTP check.
  - `index` - (Required) The index of the HTTP check.
  - `addr` - (Optional) The address to check.
  - `match` - (Optional) The match type.
  - `pattern` - (Optional) The pattern to match.
  - `type` - (Optional) The check type.
  - `method` - (Optional) The HTTP method to use for the check.
  - `port` - (Optional) The port to check.
- `acl` - (Optional) A block to configure an ACL.
  - `index` - (Required) The index of the ACL.
  - `acl_name` - (Required) The name of the ACL.
  - `criterion` - (Required) The criterion of the ACL.
  - `value` - (Required) The value of the ACL.
- `tcp_check` - (Optional) A block to configure a TCP check.
  - `index` - (Required) The index of the TCP check.
  - `action` - (Required) The action of the TCP check.
  - `comment` - (Optional) A comment for the TCP check.
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