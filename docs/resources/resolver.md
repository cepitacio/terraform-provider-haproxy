# haproxy_resolver

Manages a resolver in HAProxy.

## Example Usage

```hcl
resource "haproxy_resolver" "example" {
  name = "example"
}
```

## Argument Reference

- `name` - (Required) The name of the resolver.