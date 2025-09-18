package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// GetBackendSchema returns the schema for the backend block
func GetBackendSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Backend configuration.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the backend.",
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "The mode of the backend (http, tcp).",
			},
			"adv_check": schema.StringAttribute{
				Optional:    true,
				Description: "Advanced health check configuration.",
			},
			"http_connection_mode": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP connection mode for the backend.",
			},
			"server_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Server timeout in milliseconds.",
			},
			"check_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Health check timeout in milliseconds.",
			},
			"connect_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Connection timeout in milliseconds.",
			},
			"queue_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Queue timeout in milliseconds.",
			},
			"tunnel_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Tunnel timeout in milliseconds.",
			},
			"tarpit_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Tarpit timeout in milliseconds.",
			},
			"checkcache": schema.StringAttribute{
				Optional:    true,
				Description: "Health check cache configuration.",
			},
			"retries": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of retries for failed operations.",
			},
			"servers": GetServersSchema(),
		},
		Blocks: map[string]schema.Block{
			"balance": schema.ListNestedBlock{
				Description: "Load balancing configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"algorithm": schema.StringAttribute{
							Required:    true,
							Description: "The load balancing algorithm.",
						},
						"url_param": schema.StringAttribute{
							Optional:    true,
							Description: "The URL parameter for load balancing.",
						},
					},
				},
			},
			"httpchk_params": schema.ListNestedBlock{
				Description: "HTTP health check parameters for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Required:    true,
							Description: "The HTTP method for health checks.",
						},
						"uri": schema.StringAttribute{
							Required:    true,
							Description: "The URI for health checks.",
						},
						"version": schema.StringAttribute{
							Optional:    true,
							Description: "The HTTP version for health checks.",
						},
					},
				},
			},
			"forwardfor": schema.ListNestedBlock{
				Description: "Forward for configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.StringAttribute{
							Required:    true,
							Description: "Whether forward for is enabled.",
						},
					},
				},
			},
			"default_server": schema.SingleNestedBlock{
				Description: "Default server configuration for SSL/TLS settings.",
				Attributes: map[string]schema.Attribute{
					"ssl": schema.StringAttribute{
						Optional:    true,
						Description: "SSL configuration for the default server.",
					},
					"ssl_cafile": schema.StringAttribute{
						Optional:    true,
						Description: "SSL CA file for the default server.",
					},
					"ssl_certificate": schema.StringAttribute{
						Optional:    true,
						Description: "SSL certificate for the default server.",
					},
					"ssl_max_ver": schema.StringAttribute{
						Optional:    true,
						Description: "SSL maximum version for the default server.",
					},
					"ssl_min_ver": schema.StringAttribute{
						Optional:    true,
						Description: "SSL minimum version for the default server.",
					},
					"ssl_reuse": schema.StringAttribute{
						Optional:    true,
						Description: "SSL reuse configuration for the default server.",
					},
					"ciphers": schema.StringAttribute{
						Optional:    true,
						Description: "Ciphers for the default server.",
					},
					"ciphersuites": schema.StringAttribute{
						Optional:    true,
						Description: "Cipher suites for the default server.",
					},
					"verify": schema.StringAttribute{
						Optional:    true,
						Description: "SSL verification for the default server.",
					},
					// v3 fields
					"sslv3": schema.StringAttribute{
						Optional:    true,
						Description: "SSLv3 support for the default server (Data Plane API v3 only).",
					},
					"tlsv10": schema.StringAttribute{
						Optional:    true,
						Description: "TLSv1.0 support for the default server (Data Plane API v3 only).",
					},
					"tlsv11": schema.StringAttribute{
						Optional:    true,
						Description: "TLSv1.1 support for the default server (Data Plane API v3 only).",
					},
					"tlsv12": schema.StringAttribute{
						Optional:    true,
						Description: "TLSv1.2 support for the default server (Data Plane API v3 only).",
					},
					"tlsv13": schema.StringAttribute{
						Optional:    true,
						Description: "TLSv1.3 support for the default server (Data Plane API v3 only).",
					},
					// v2 fields (deprecated in v3)
					"no_sslv3": schema.StringAttribute{
						Optional:    true,
						Description: "Disable SSLv3 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"no_tlsv10": schema.StringAttribute{
						Optional:    true,
						Description: "Disable TLSv1.0 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"no_tlsv11": schema.StringAttribute{
						Optional:    true,
						Description: "Disable TLSv1.1 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"no_tlsv12": schema.StringAttribute{
						Optional:    true,
						Description: "Disable TLSv1.2 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"no_tlsv13": schema.StringAttribute{
						Optional:    true,
						Description: "Disable TLSv1.3 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"force_sslv3": schema.StringAttribute{
						Optional:    true,
						Description: "Force SSLv3 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"force_tlsv10": schema.StringAttribute{
						Optional:    true,
						Description: "Force TLSv1.0 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"force_tlsv11": schema.StringAttribute{
						Optional:    true,
						Description: "Force TLSv1.1 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"force_tlsv12": schema.StringAttribute{
						Optional:    true,
						Description: "Force TLSv1.2 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"force_tlsv13": schema.StringAttribute{
						Optional:    true,
						Description: "Force TLSv1.3 for the default server (Data Plane API v2 only, deprecated in v3).",
					},
					"force_strict_sni": schema.StringAttribute{
						Optional:    true,
						Description: "Force strict SNI for the default server (Data Plane API v2 only, deprecated in v3).",
					},
				},
			},
			"http_request_rules":  GetHttpRequestRuleSchema(),
			"http_response_rules": GetHttpResponseRuleSchema(),
			"tcp_request_rules":   GetTcpRequestRuleSchema(),
			"tcp_response_rules":  GetTcpResponseRuleSchema(),
			"stick_table": schema.SingleNestedBlock{
				Description: "Stick table configuration for the backend.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Description: "The type of the stick table.",
					},
					"size": schema.StringAttribute{
						Optional:    true,
						Description: "The size of the stick table.",
					},
					"expire": schema.StringAttribute{
						Optional:    true,
						Description: "The expiration time for the stick table.",
					},
					"nopurge": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to disable purging for the stick table.",
					},
					"peers": schema.StringAttribute{
						Optional:    true,
						Description: "The peers for the stick table.",
					},
				},
			},
			"stats_options": schema.ListNestedBlock{
				Description: "Stats options configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"stats_enable": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether to enable stats for the backend.",
						},
						"stats_uri": schema.StringAttribute{
							Optional:    true,
							Description: "The stats URI for the backend.",
						},
						"stats_realm": schema.StringAttribute{
							Optional:    true,
							Description: "The stats realm for the backend.",
						},
						"stats_auth": schema.StringAttribute{
							Optional:    true,
							Description: "The stats authentication for the backend.",
						},
					},
				},
			},
			"acls":        GetACLSchema(),
			"http_checks": GetHttpcheckSchema(),
			"tcp_checks":  GetTcpCheckSchema(),
		},
	}
}
