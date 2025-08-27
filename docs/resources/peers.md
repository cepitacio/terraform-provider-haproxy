# haproxy_peers

Manages a peers section in HAProxy.

## Example Usage

```hcl
resource "haproxy_peers" "example" {
  name = "example"
}
```

## Argument Reference

- `name` - (Required) The name of the peers section.