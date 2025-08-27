# haproxy_stick_table

Manages a stick table in HAProxy.

## Example Usage

```hcl
resource "haproxy_stick_table" "example" {
  name  = "example"
  type  = "ip"
  size  = "1m"
  store = "gpc0"
}
```

## Argument Reference

- `name` - (Required) The name of the stick table.
- `type` - (Optional) The type of the stick table.
- `size` - (Optional) The size of the stick table.
- `store` - (Optional) The store of the stick table.
- `peers` - (Optional) The peers of the stick table.
- `no_purge` - (Optional) Whether to disable purging of the stick table.