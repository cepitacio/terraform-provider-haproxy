terraform {
  required_providers {
    haproxy = {
      source = "haproxy/haproxy"
      version = "~> 0.1"
    }
  }
}

provider "haproxy" {
  host     = "localhost"
  port     = 5555
  username = "admin"
  password = "admin"
  api_version = "v2"
}

# Example with both frontend and backend ACLs
resource "haproxy_stack" "example_with_acls" {
  name = "example_stack_with_acls"

  backend {
    name = "example_backend"
    mode = "http"
    
    # Backend ACLs for content switching and decision making
    acls {
      acl_name = "is_api_request"
      criterion = "path"
      value = "/api"
      index = 0
    }
    
    acls {
      acl_name = "is_admin_user"
      criterion = "hdr"
      value = "X-User-Role admin"
      index = 1
    }
    
    acls {
      acl_name = "is_mobile_client"
      criterion = "hdr"
      value = "User-Agent mobile"
      index = 2
    }
    
    acls {
      acl_name = "is_secure_connection"
      criterion = "ssl_fc"
      value = ""
      index = 3
    }

    # Other backend configurations
    server_timeout = 30000
    check_timeout = 2000
    connect_timeout = 5000
    
    default_server {
      ssl = "enabled"
      ssl_certificate = "/etc/ssl/certs/server.crt"
      ssl_cafile = "/etc/ssl/certs/ca.crt"
    }
  }

  server {
    name    = "web_server_1"
    address = "192.168.1.10"
    port    = 8080
    check   = "enabled"
    weight  = 100
    rise    = 2
    fall    = 3
  }

  server {
    name    = "web_server_2"
    address = "192.168.1.11"
    port    = 8080
    check   = "enabled"
    weight  = 100
    rise    = 2
    fall    = 3
  }

  frontend {
    name = "example_frontend"
    mode = "http"
    default_backend = "example_backend"
    maxconn = 10000
    backlog = 100
    
    # Frontend ACLs for routing and access control
    acls {
      acl_name = "is_static_content"
      criterion = "path"
      value = "/static"
      index = 0
    }
    
    acls {
      acl_name = "is_admin_path"
      criterion = "path"
      value = "/admin"
      index = 1
    }
    
    acls {
      acl_name = "is_public_api"
      criterion = "path"
      value = "/public"
      index = 2
    }
    
    acls {
      acl_name = "is_internal_network"
      criterion = "src"
      value = "192.168.0.0/16"
      index = 3
    }

    bind {
      name    = "http_bind"
      address = "0.0.0.0"
      port    = 80
    }
    
    bind {
      name    = "https_bind"
      address = "0.0.0.0"
      port    = 443
      ssl     = true
    }
  }
}

# Example with only backend ACLs (no frontend ACLs)
resource "haproxy_stack" "backend_only_acls" {
  name = "backend_only_stack"

  backend {
    name = "api_backend"
    mode = "http"
    
    # Backend ACLs for API routing
    acls {
      acl_name = "is_v1_api"
      criterion = "path"
      value = "/api/v1"
      index = 0
    }
    
    acls {
      acl_name = "is_v2_api"
      criterion = "path"
      value = "/api/v2"
      index = 1
    }
    
    acls {
      acl_name = "is_authenticated"
      criterion = "hdr"
      value = "Authorization Bearer"
      index = 2
    }
    
    acls {
      acl_name = "is_rate_limited"
      criterion = "src"
      value = "10.0.0.0/8"
      index = 3
    }

    server_timeout = 60000
    check_timeout = 5000
    connect_timeout = 10000
  }

  server {
    name    = "api_server_1"
    address = "10.0.1.10"
    port    = 9000
    check   = "enabled"
    weight  = 50
    rise    = 3
    fall    = 2
  }

  server {
    name    = "api_server_2"
    address = "10.0.1.11"
    port    = 9000
    check   = "enabled"
    weight  = 50
    rise    = 3
    fall    = 2
  }

  frontend {
    name = "api_frontend"
    mode = "http"
    default_backend = "api_backend"
    maxconn = 5000
    
    bind {
      name    = "api_bind"
      address = "0.0.0.0"
      port    = 8080
    }
  }
}

# Example showing ACLs with different criteria types
resource "haproxy_stack" "advanced_acls_example" {
  name = "advanced_acls_stack"

  backend {
    name = "advanced_backend"
    mode = "http"
    
    # Advanced backend ACLs with various criteria
    acls {
      acl_name = "is_get_request"
      criterion = "method"
      value = "GET"
      index = 0
    }
    
    acls {
      acl_name = "is_post_request"
      criterion = "method"
      value = "POST"
      index = 1
    }
    
    acls {
      acl_name = "has_json_content"
      criterion = "hdr"
      value = "Content-Type application/json"
      index = 2
    }
    
    acls {
      acl_name = "is_high_priority"
      criterion = "hdr"
      value = "X-Priority high"
      index = 3
    }
    
    acls {
      acl_name = "is_localhost"
      criterion = "src"
      value = "127.0.0.1"
      index = 4
    }
    
    acls {
      acl_name = "is_secure_protocol"
      criterion = "ssl_fc_protocol"
      value = "TLSv1.3"
      index = 5
    }

    server_timeout = 45000
    check_timeout = 3000
    connect_timeout = 8000
  }

  server {
    name    = "advanced_server"
    address = "172.16.1.10"
    port    = 7000
    check   = "enabled"
    weight  = 100
    rise    = 2
    fall    = 3
  }

  frontend {
    name = "advanced_frontend"
    mode = "http"
    default_backend = "advanced_backend"
    maxconn = 8000
    
    bind {
      name    = "advanced_bind"
      address = "0.0.0.0"
      port    = 9090
    }
  }
}


