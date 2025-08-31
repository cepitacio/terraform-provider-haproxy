# Example demonstrating ACL index normalization
# The provider will automatically normalize non-sequential indices to sequential order
# while preserving the user's intended ACL sequence

resource "haproxy_stack" "acl_normalization_demo" {
  name = "acl_normalization_demo"

  backend {
    name = "demo_backend"
    mode = "http"
    
    # Example 1: Non-sequential indices that will be normalized
    acls {
      acl_name = "first_acl"
      criterion = "path"
      value = "/first"
      index = 5  # Will be normalized to 0
    }
    
    acls {
      acl_name = "second_acl"
      criterion = "path"
      value = "/second"
      index = 2  # Will be normalized to 1
    }
    
    acls {
      acl_name = "third_acl"
      criterion = "path"
      value = "/third"
      index = 8  # Will be normalized to 2
    }
    
    # Example 2: Gaps in indices that will be filled
    acls {
      acl_name = "fourth_acl"
      criterion = "path"
      value = "/fourth"
      index = 0  # Will be normalized to 3
    }
    
    acls {
      acl_name = "fifth_acl"
      criterion = "path"
      value = "/fifth"
      index = 10 # Will be normalized to 4
    }

    server_timeout = 30000
  }

  server {
    name    = "demo_server"
    address = "192.168.1.100"
    port    = 8080
    check   = "enabled"
  }

  frontend {
    name = "demo_frontend"
    mode = "http"
    default_backend = "demo_backend"
    
    # Example 3: Mixed sequential and non-sequential indices
    acls {
      acl_name = "frontend_first"
      criterion = "path"
      value = "/frontend"
      index = 1  # Will be normalized to 0
    }
    
    acls {
      acl_name = "frontend_second"
      criterion = "path"
      value = "/api"
      index = 0  # Will be normalized to 1
    }
    
    acls {
      acl_name = "frontend_third"
      criterion = "path"
      value = "/admin"
      index = 3  # Will be normalized to 2
    }

    bind {
      name    = "demo_bind"
      address = "0.0.0.0"
      port    = 80
    }
  }
}

# Example showing the expected normalized result:
# 
# Backend ACLs will be normalized to:
# - first_acl (index: 0) - was index 5
# - second_acl (index: 1) - was index 2  
# - third_acl (index: 2) - was index 8
# - fourth_acl (index: 3) - was index 0
# - fifth_acl (index: 4) - was index 10
#
# Frontend ACLs will be normalized to:
# - frontend_first (index: 0) - was index 1
# - frontend_second (index: 1) - was index 0
# - frontend_third (index: 2) - was index 3
#
# The provider will log:
# "ACL indices normalized to sequential order for HAProxy compatibility"
# "Original order: first_acl(index:5) → second_acl(index:2) → third_acl(index:8) → fourth_acl(index:0) → fifth_acl(index:10)"
# "Normalized order: first_acl(index:0) → second_acl(index:1) → third_acl(index:2) → fourth_acl(index:3) → fifth_acl(index:4)"


