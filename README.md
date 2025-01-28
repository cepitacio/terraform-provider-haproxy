<p align="center">
  <a href="https://github.com/cepitacio/terraform-provider-haproxy">
    <img src="./assets/haproxy.png" alt="minio-provider-terraform" width="200">
  </a>
  <h1 align="center" style="font-weight: bold">Terraform Provider for HAProxy</h1>

</p>

## Table Of Contents
- [Table Of Contents](#table-of-contents)
  - [About This Project](#about-this-project)
  - [Data Plane API Installation](#data-plane-api-installation)
  - [License](#license)

### About This Project

A [Terraform](https://www.terraform.io) provider to manage [HAProxy](https://www.haproxy.com/). It uses [HAProxy Data Plane API](https://github.com/haproxytech/dataplaneapi) to manage HAProxy. This provider is tested with HAProxy version 2.x, up to 2.9.8. Compatibility with HAProxy 3.x is not currently supported due to significant changes, but I plan to add support for version 3 in the future.

### Data Plane API Installation

To use this provider, you need to install [HAProxy Data Plane API](https://www.haproxy.com/documentation/hapee/2-0r1/api/data-plane-api/installation/haproxy-community/) on your HAProxy server or use entrprise version of HAProxy.


### License

Distributed under the Apache License. See [LICENSE](./LICENSE) for more information.

Made with <span style="color: #e25555;">&#9829;</span> using [Go](https://golang.org/).