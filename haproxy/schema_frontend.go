package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// GetFrontendSchema returns the schema for the frontend block
func GetFrontendSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Frontend configuration.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the frontend.",
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "The mode of the frontend (http, tcp).",
			},
			"default_backend": schema.StringAttribute{
				Required:    true,
				Description: "The default backend for the frontend.",
			},
			"maxconn": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of connections for the frontend.",
			},
			"backlog": schema.Int64Attribute{
				Optional:    true,
				Description: "Backlog setting for the frontend.",
			},
			"ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether SSL is enabled for the frontend.",
			},
			"ssl_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "SSL certificate for the frontend.",
			},
			"ssl_cafile": schema.StringAttribute{
				Optional:    true,
				Description: "SSL CA file for the frontend.",
			},
			"ssl_max_ver": schema.StringAttribute{
				Optional:    true,
				Description: "SSL maximum version for the frontend.",
			},
			"ssl_min_ver": schema.StringAttribute{
				Optional:    true,
				Description: "SSL minimum version for the frontend.",
			},
			"ciphers": schema.StringAttribute{
				Optional:    true,
				Description: "Ciphers for the frontend.",
			},
			"ciphersuites": schema.StringAttribute{
				Optional:    true,
				Description: "Cipher suites for the frontend.",
			},
			"verify": schema.StringAttribute{
				Optional:    true,
				Description: "SSL verification for the frontend.",
			},
			"accept_proxy": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to accept proxy protocol.",
			},
			"defer_accept": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to defer accept.",
			},
			"tcp_user_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "TCP user timeout for the frontend.",
			},
			"tfo": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether TCP Fast Open is enabled.",
			},
			"v4v6": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to use both IPv4 and IPv6.",
			},
			"v6only": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to use IPv6 only.",
			},
			"monitor_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The URI to use for health monitoring of the frontend.",
			},
			"binds": GetBindSchema(),
		},
		Blocks: map[string]schema.Block{
			"stats_options": schema.ListNestedBlock{
				Description: "Stats options configuration for the frontend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"stats_enable": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether to enable stats for the frontend.",
						},
						"stats_uri": schema.StringAttribute{
							Optional:    true,
							Description: "The stats URI for the frontend.",
						},
						"stats_realm": schema.StringAttribute{
							Optional:    true,
							Description: "The stats realm for the frontend.",
						},
						"stats_auth": schema.StringAttribute{
							Optional:    true,
							Description: "The stats authentication for the frontend.",
						},
					},
				},
			},
			"monitor_fail": schema.ListNestedBlock{
				Description: "Monitor fail configuration for the frontend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cond": schema.StringAttribute{
							Required:    true,
							Description: "The condition for monitor fail (if, unless).",
							Validators: []validator.String{
								stringvalidator.OneOf("if", "unless"),
							},
						},
						"cond_test": schema.StringAttribute{
							Required:    true,
							Description: "The condition test for monitor fail.",
						},
					},
				},
			},
			"acls":                GetACLSchema(),
			"http_request_rules":  GetHttpRequestRuleSchema(),
			"http_response_rules": GetHttpResponseRuleSchema(),
			"tcp_request_rules":   GetTcpRequestRuleSchema(),
		},
	}
}
