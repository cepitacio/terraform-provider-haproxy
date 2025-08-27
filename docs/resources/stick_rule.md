# haproxy_stick_rule

Manages a stick rule in HAProxy.

## Example Usage

```hcl
resource "haproxy_stick_rule" "example" {
  index   = 0
  type    = "match"
  pattern = "src"
  backend = "example"
}
```

## Argument Reference

- `index` - (Required) The index of the stick rule.
- `type` - (Required) The type of the stick rule.
- `cond` - (Optional) The condition of the stick rule.
- `cond_test` - (Optional) The condition test of the stick rule.
- `pattern` - (Optional) The pattern of the stick rule.
- `table` - (Optional) The table of the stick rule.
- `backend` - (Required) The backend to which the stick rule belongs.