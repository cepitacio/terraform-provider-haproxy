# HAProxy Terraform Provider

This is a Terraform provider for managing HAProxy configuration using the HAProxy Data Plane API.

## Architecture

![HAProxy Terraform Provider Architecture](assets/haproxy.png)

## Usage

```hcl
provider "haproxy" {
  url      = "http://localhost:5555"
  username = "admin"
  password = "password"
}
```

## Resources

- `haproxy_frontend`
- `haproxy_backend`
- `haproxy_server`
- `haproxy_bind`
- `haproxy_acl`
- `haproxy_http_request_rule`
- `haproxy_http_response_rule`
- `haproxy_tcp_request_rule`
- `haproxy_tcp_response_rule`
- `haproxy_resolver`
- `haproxy_nameserver`
- `haproxy_peers`
- `haproxy_peer_entry`
- `haproxy_stick_rule`
- `haproxy_stick_table`
- `haproxy_log_forward`
- `haproxy_global`

## Data Sources

- `haproxy_frontends`
- `haproxy_backends`
