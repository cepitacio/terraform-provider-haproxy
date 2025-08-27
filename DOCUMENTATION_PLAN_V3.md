# HAProxy Terraform Provider Documentation Plan (V3)

This document outlines the plan for creating comprehensive documentation for the HAProxy Terraform provider. This version builds upon V2, adding more detail to resource interactions and complex configurations.

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

**Note:** The documentation will be structured to integrate seamlessly with the official [Terraform Registry](https://registry.terraform.io/), which automatically parses `*.md` files to generate documentation pages.

## 2. Provider Configuration

This section will detail how to configure the provider itself. It will cover:

- **Authentication:** How to provide the URL, username, and password for the Data Plane API.
- **Environment Variables:** Show how to use `HAPROXY_ENDPOINT`, `HAPROXY_USER`, and `HAPROXY_PASSWORD`.
- **SSL Verification:** Explain the `insecure` option for disabling SSL certificate verification.
- **Example `provider` block:**

```terraform
provider "haproxy" {
  url      = "http://haproxy.example.com:5555"
  username = "admin"
  password = "mypassword"
  insecure = true
}
```

## 3. API Transaction Handling

This section will explain how the provider interacts with the HAProxy Data Plane API's transaction-based system. It will cover:

- A brief, high-level explanation of how the API requires changes to be wrapped in a transaction.
- How the provider automates the creation, application, and deletion of transactions.
- The importance of this for ensuring atomic configuration updates.
- How this relates to potential troubleshooting scenarios (e.g., stale transaction files).

## 4. Resource Relationships

This new section will clarify the hierarchy and dependencies between the main resources to help users model their configurations correctly.

```mermaid
graph TD
    subgraph Frontend Configuration
        F[haproxy_frontend]
        B[haproxy_bind]
    end

    subgraph Backend Configuration
        BK[haproxy_backend]
        S[haproxy_server]
    end

    F --> B;
    F -- Uses default_backend --> BK;
    BK -- Contains --> S;
```

- **`haproxy_frontend`**: The entry point for traffic. It defines how HAProxy listens for and routes incoming requests.
- **`haproxy_bind`**: A child of a frontend. It specifies the IP address and port that the frontend listens on.
- **`haproxy_backend`**: A pool of servers that will handle the requests forwarded by a frontend.
- **`haproxy_server`**: A child of a backend. It represents an individual application server that will receive traffic.

## 5. Resource Documentation

Each resource will have its own dedicated documentation page. The structure for each page will be:

- **Resource Overview:** A brief description of the HAProxy object the resource manages.
- **Example Usage:** A basic, complete example of how to use the resource.
- **Argument Reference:** A detailed list of all the arguments supported by the resource, including:
    - `(Required)` or `(Optional)` designation.
    - A clear description of what the argument does.
    - Any default values.
    - Any constraints or allowed values.
- **Nested Blocks:** Detailed documentation for any nested blocks (like `balance` or `httpchk_params`), following the same format as the top-level arguments.
- **Attribute Reference:** A list of all the attributes exported by the resource.
- **Import:** Instructions on how to import existing HAProxy resources into Terraform state.

### 5.1. `haproxy_frontend`

- Document all top-level arguments (`name`, `default_backend`, `mode`, etc.).
- Document the `monitor_fail` nested block.
- Document the `acl`, `httprequestrule`, and `httpresponserule` nested blocks, explaining how they are used to build complex frontend logic.

### 5.2. `haproxy_backend`

- Document all top-level arguments (`name`, `mode`, `adv_check`, etc.).
- Document the `balance`, `httpchk_params`, and `forwardfor` nested blocks.
- Document the `httpcheck` nested block and how it relates to the top-level `adv_check`.

### 5.3. `haproxy_server`

- Document all top-level arguments (`name`, `address`, `port`, `parent_name`, `parent_type`, etc.).
- Explain the relationship between a server and its parent backend.
- Detail all health check parameters (`check`, `inter`, `rise`, `fall`).
- Document all SSL-related parameters.

### 5.4. `haproxy_bind`

- Document all top-level arguments (`name`, `address`, `port`, `parent_name`, `parent_type`, etc.).
- Explain the relationship between a bind and its parent frontend.
- Detail all SSL-related parameters.

## 6. Data Source Documentation

This section will document all available data sources. While none are currently implemented, this section will serve as a placeholder for future additions. The structure for each data source page will be:

- **Data Source Overview:** A brief description of the information the data source provides.
- **Example Usage:** A basic example of how to use the data source.
- **Argument Reference:** A list of arguments to filter or specify the data to be fetched.
- **Attribute Reference:** A list of all the attributes exported by the data source.

## 7. Common Services and Advanced Configuration

This section will provide more in-depth explanations of the common services and advanced features that can be configured within other resources.

### 7.1. ACLs and Rules

- **ACLs:** How to define and use ACLs for conditional logic. Provide examples for common use cases like path-based routing and host-based routing.
- **HTTP Request/Response Rules:** How to manipulate HTTP traffic with rules. Explain the difference between common actions like `add-header`, `del-header`, `replace-header`, and `redirect`.
- **Example:** Show how to use an `acl` with an `httprequestrule` to redirect traffic.

### 7.2. Health Checks

- A deeper dive into the different types of health checks available (`httpchk`, `ssl-hello-chk`, etc.).
- Explain the relationship between the `adv_check` argument in the `haproxy_backend` resource and the `httpcheck` nested block.
- Provide a detailed example of configuring a custom HTTP health check with specific headers and status code expectations.

### 7.3. Timeouts

- Explain the different timeout settings available across the provider (`server_timeout`, `check_timeout`, `connect_timeout`, etc.).
- Provide guidance on how to tune these values for different application profiles (e.g., long-polling vs. quick API requests).

## 8. Comprehensive Usage Examples

This section will provide complete, real-world examples that demonstrate how to combine multiple resources to build a functional HAProxy configuration. Examples will include:

- **Basic HTTP Load Balancer:** A simple frontend and backend with two servers.
- **HTTPS Load Balancer:** An example showing how to configure SSL termination.
- **Path-Based Routing:** An example using ACLs and HTTP request rules to route traffic based on the URL path.
- **Blue/Green Deployments:** A more advanced example showing how to manage two backends for blue/green deployments.

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