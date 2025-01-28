terraform {
  required_providers {
    haproxy = {
      source  = "cepitacio/haproxy"
      version = "0.0.6"
    }
  }
}

provider "haproxy" {
  url      = "http://haproxy.example.com:8080"
  username = "username"
  password = "password"
}
