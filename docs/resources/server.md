# haproxy_server

Manages a server in HAProxy.

## Example Usage

```hcl
resource "haproxy_server" "example" {
  name        = "example"
  address     = "127.0.0.1"
  port        = 8080
  parent_name = "example"
  parent_type = "backend"
}
```

## Argument Reference

- `name` - (Required) The name of the server.
- `address` - (Required) The address of the server.
- `port` - (Required) The port of the server.
- `parent_name` - (Required) The name of the parent object.
- `parent_type` - (Required) The type of the parent object.
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