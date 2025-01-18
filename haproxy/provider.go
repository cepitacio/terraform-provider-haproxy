package haproxy

import (
	backend "terraform-provider-haproxy/internal/backend"
	bind "terraform-provider-haproxy/internal/bind"
	acl "terraform-provider-haproxy/internal/common_services/acl"
	httpcheck "terraform-provider-haproxy/internal/common_services/httpcheck"
	httprequestrule "terraform-provider-haproxy/internal/common_services/httprequestrule"
	httpresponserule "terraform-provider-haproxy/internal/common_services/httpresponserule"
	frontend "terraform-provider-haproxy/internal/frontend"
	server "terraform-provider-haproxy/internal/server"
	transaction "terraform-provider-haproxy/internal/transaction"

	"terraform-provider-haproxy/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Haproxy Host and Port",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"HAPROXY_ENDPOINT",
				}, nil),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Haproxy User",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"HAPROXY_USER",
				}, nil),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Haproxy Password",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"HAPROXY_PASSWORD",
				}, nil),
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable SSL certificate verification (default: false)",
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"HAPROXY_INSECURE",
				}, nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"haproxy_frontend": frontend.ResourceHaproxyFrontend(),
			"haproxy_backend":  backend.ResourceHaproxyBackend(),
			"haproxy_server":   server.ResourceHaproxyServer(),
			"haproxy_bind":     bind.ResourceHaproxyBind(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {
	commonConfig := utils.Configuration{
		Username: data.Get("username").(string),
		Password: data.Get("password").(string),
		BaseURL:  data.Get("url").(string),
	}

	// Create backend config for backend
	backendConfig := &backend.ConfigBackend{}
	utils.SetConfigValues(backendConfig, commonConfig)

	// Create frontend config for frontend
	frontendConfig := &frontend.ConfigFrontend{}
	utils.SetConfigValues(frontendConfig, commonConfig)

	// Create server config for server
	serverConfig := &server.ConfigServer{}
	utils.SetConfigValues(serverConfig, commonConfig)

	// Create transaction config for transaction
	transactionConfig := &transaction.ConfigTransaction{}
	utils.SetConfigValues(transactionConfig, commonConfig)

	bindConfig := &bind.ConfigBind{}
	utils.SetConfigValues(bindConfig, commonConfig)

	aclConfig := &acl.ConfigAcl{}
	utils.SetConfigValues(aclConfig, commonConfig)

	httprequestruleConfig := &httprequestrule.ConfigHttpRequestRule{}
	utils.SetConfigValues(httprequestruleConfig, commonConfig)

	httpresponseruleConfig := &httpresponserule.ConfigHttpResponseRule{}
	utils.SetConfigValues(httpresponseruleConfig, commonConfig)

	httpcheckConfig := &httpcheck.ConfigHttpCheck{}
	utils.SetConfigValues(httpcheckConfig, commonConfig)

	return map[string]interface{}{
		"backend":          backendConfig,
		"frontend":         frontendConfig,
		"server":           serverConfig,
		"transaction":      transactionConfig,
		"acl":              aclConfig,
		"bind":             bindConfig,
		"httprequestrule":  httprequestruleConfig,
		"httpresponserule": httpresponseruleConfig,
		"httpcheck":        httpcheckConfig,
	}, nil
}
