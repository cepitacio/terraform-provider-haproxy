# haproxy_global

Manages the global configuration in HAProxy.

## Example Usage

```hcl
resource "haproxy_global" "example" {
  name    = "example"
  maxconn = 20000
}
```

## Argument Reference

- `name` - (Required) The name of the global configuration.
- `maxconn` - (Optional) The maximum number of concurrent connections.
- `daemon` - (Optional) The daemon mode.
- `stats_timeout` - (Optional) The stats timeout.
- `tune_ssl_default_dh_param` - (Optional) The default DH parameter size for SSL.
- `ssl_default_bind_ciphers` - (Optional) The default bind ciphers for SSL.
- `ssl_default_bind_options` - (Optional) The default bind options for SSL.
- `ssl_default_server_ciphers` - (Optional) The default server ciphers for SSL.
- `ssl_default_server_options` - (Optional) The default server options for SSL.