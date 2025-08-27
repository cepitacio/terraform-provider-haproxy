# haproxy_peer_entry

Manages a peer entry in HAProxy.

## Example Usage

```hcl
resource "haproxy_peer_entry" "example" {
  name    = "example"
  address = "127.0.0.1"
  peers   = "example"
}
```

## Argument Reference

- `name` - (Required) The name of the peer entry.
- `address` - (Required) The address of the peer entry.
- `port` - (Optional) The port of the peer entry.
- `peers` - (Required) The peers to which the peer entry belongs.