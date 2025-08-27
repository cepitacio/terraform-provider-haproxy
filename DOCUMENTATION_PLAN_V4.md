# HAProxy Terraform Provider Documentation Plan (V4)

This document outlines the plan for creating comprehensive documentation for the HAProxy Terraform provider. This version incorporates support for Data Plane API v2 and v3, adds new resources, and provides a clear structure for data sources.

## 1. High-Level Overview

This will be an introductory section in the main `README.md` that explains:

- The purpose of the provider.
- The benefits of using it to manage HAProxy with Terraform.
- The relationship between the provider and the HAProxy Data Plane API.
- A simple diagram showing the architecture: Terraform -> Provider -> Data Plane API -> HAProxy.

```mermaid
graph TD
    A[Terraform] -- Manages --> B(HAProxy Provider);
    B -- Communicates with --> C(HAProxy Data Plane API);
    C -- Configures --> D(HAProxy);
```

**Note:** The documentation will be structured to integrate seamlessly with the official [Terraform Registry](https://registry.terraform.io/). It will also serve as a companion to the official [HAProxy Data Plane API Documentation](https://www.haproxy.com/documentation/dataplaneapi/), which is the source of truth for all API interactions.

## 2. Provider Configuration

This section will detail how to configure the provider itself. It will cover:

- **Authentication:** How to provide the URL, username, and password for the Data Plane API.
- **Environment Variables:** Show how to use `HAPROXY_ENDPOINT`, `HAPROXY_USER`, and `HAPROXY_PASSWORD`.
- **SSL Verification:** Explain the `insecure` option for disabling SSL certificate verification.
- **API Versioning:** A new field to allow users to specify the Data Plane API version (`v2` or `v3`).

- **Example `provider` block:**

```terraform
provider "haproxy" {
  url          = "http://haproxy.example.com:5555"
  username     = "admin"
  password     = "mypassword"
  insecure     = true
  api_version  = "v3" # (Optional) Defaults to v2 for backward compatibility
}
```

## 3. API Versioning and Compatibility

This new section will explain how the provider handles different versions of the HAProxy Data Plane API.

- **Supported Versions:** Clearly state that the provider supports both `v2` and `v3` of the Data Plane API.
- **Feature Parity:** Explain that some resources or arguments may only be available in `v3`.
- **Provider Behavior:** Describe how the provider adapts its behavior based on the selected `api_version`.
- **Migration Guide:** Provide guidance for users who want to migrate their configurations from `v2` to `v3`.

## 4. API Transaction Handling

This section will explain how the provider interacts with the HAProxy Data Plane API's transaction-based system. It will cover:

- A brief, high-level explanation of how the API requires changes to be wrapped in a transaction.
- How the provider automates the creation, application, and deletion of transactions.
- The importance of this for ensuring atomic configuration updates.

## 5. Resource Relationships

This section will clarify the hierarchy and dependencies between the main resources.

```mermaid
graph TD
    subgraph Global
        G[haproxy_global]
    end

    subgraph Frontend
        F[haproxy_frontend]
        B[haproxy_bind]
    end

    subgraph Backend
        BK[haproxy_backend]
        S[haproxy_server]
    end

    subgraph Peers
        P[haproxy_peers]
    end

    subgraph Resolvers
        R[haproxy_resolver]
    end

    F --> B;
    F -- Uses default_backend --> BK;
    BK -- Contains --> S;
```

## 6. Resource Documentation

Each resource will have its own dedicated documentation page. The structure for each page will be:

- **Resource Overview:** A brief description of the HAProxy object the resource manages.
- **Data Plane API Endpoint:** The corresponding Data Plane API endpoint (e.g., `/v2/services/haproxy/configuration/frontends`).
- **Example Usage:** A basic, complete example.
- **Argument Reference:** A detailed list of all arguments.
- **Attribute Reference:** A list of all exported attributes.
- **Import:** Instructions on how to import existing resources.

### 6.1. Core Resources
- `haproxy_global`
- `haproxy_frontend`
- `haproxy_backend`
- `haproxy_server`
- `haproxy_bind`

### 6.2. Service Discovery
- `haproxy_resolver`
- `haproxy_nameserver`

### 6.3. High Availability & Stickiness
- `haproxy_peers`
- `haproxy_peer_entry`
- `haproxy_stick_rule`
- `haproxy_stick_table`

### 6.4. Rules and Checks
- `haproxy_acl`
- `haproxy_tcp_check`
- `haproxy_http_check`
- `haproxy_http_request_rule`
- `haproxy_http_response_rule`
- `haproxy_tcp_request_rule`
- `haproxy_tcp_response_rule`

### 6.5. Logging
- `haproxy_log_forward`

## 7. Data Source Documentation

This section will document all available data sources.

### 7.1. `haproxy_backends`
- **Overview:** Fetches a list of all configured HAProxy backends.
- **Example Usage:**
  ```terraform
  data "haproxy_backends" "all" {}

  output "backend_names" {
    value = data.haproxy_backends.all.names
  }
  ```
- **Argument Reference:** None.
- **Attribute Reference:** `names` (a list of all backend names).

### 7.2. `haproxy_frontends`
- **Overview:** Fetches a list of all configured HAProxy frontends.
- **Example Usage:**
  ```terraform
  data "haproxy_frontends" "all" {}

  output "frontend_names" {
    value = data.haproxy_frontends.all.names
  }
  ```
- **Argument Reference:** None.
- **Attribute Reference:** `names` (a list of all frontend names).

## 8. Comprehensive Usage Examples

This section will provide complete, real-world examples that demonstrate how to combine multiple resources to build a functional HAProxy configuration. Examples will include:

- **Basic HTTP/HTTPS Load Balancer**
- **Path-Based Routing with ACLs**
- **Blue/Green Deployments**
- **DNS-Based Service Discovery with Resolvers**
- **Session Persistence with Stick Tables**

## 9. Contributing Guide

To encourage community contributions, a `CONTRIBUTING.md` file will be created with:

- Instructions on how to set up the development environment.
- The process for submitting bug reports and feature requests.
- Coding standards and best practices.
- How to run the test suite.

## 10. Changelog and Versioning

A `CHANGELOG.md` file will be maintained to track all changes to the provider. It will follow the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format and will include sections for:

- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be-removed features.
- `Removed` for now-removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## 11. Troubleshooting Section

A section in the main documentation dedicated to common issues and their solutions, such as:

- "404 Not Found" errors when communicating with the Data Plane API.
- Authentication failures.
- Issues with transaction versions.
- Common configuration errors and how to resolve them.