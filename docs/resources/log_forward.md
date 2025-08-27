# haproxy_log_forward

Manages a log forward in HAProxy.

## Example Usage

```hcl
resource "haproxy_log_forward" "example" {
  name = "example"
}
```

## Argument Reference

- `name` - (Required) The name of the log forward.
- `backlog` - (Optional) The backlog size.
- `maxconn` - (Optional) The maximum number of concurrent connections.
- `timeout` - (Optional) The timeout.
- `loglevel` - (Optional) The log level.