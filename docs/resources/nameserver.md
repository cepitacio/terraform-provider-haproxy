# haproxy_nameserver

Manages a nameserver in HAProxy.

## Example Usage

```hcl
resource "haproxy_nameserver" "example" {
  name     = "example"
  address  = "127.0.0.1"
  resolver = "example"
}
```

## Argument Reference

- `name` - (Required) The name of the nameserver.
- `address` - (Required) The address of the nameserver.
- `port` - (Optional) The port of the nameserver.
- `resolver` - (Required) The resolver to which the nameserver belongs.