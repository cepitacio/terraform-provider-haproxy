package haproxy

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// Config defines variable for haproxy configuration
type APIClient struct {
	Username string
	Password string
	BaseURL  string
	Insecure bool
}

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider
