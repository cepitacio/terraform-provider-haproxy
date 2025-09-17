package haproxy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetBackendSchema returns the schema for the backend block
func GetBackendSchema(schemaBuilder *VersionAwareSchemaBuilder) schema.SingleNestedBlock {
	// If no schema builder is provided, include all fields for backward compatibility
	if schemaBuilder == nil {
		schemaBuilder = CreateVersionAwareSchemaBuilder("v3") // Default to v3
	}
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
			"servers": schema.MapNestedAttribute{
				Optional:    true,
				Description: "Server configuration blocks for backend servers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Required:    true,
							Description: "The address of the server.",
						},
						"port": schema.Int64Attribute{
							Required:    true,
							Description: "The port of the server.",
						},
						"check": schema.StringAttribute{
							Optional:    true,
							Description: "Whether to enable health checks for the server.",
						},
						"backup": schema.StringAttribute{
							Optional:    true,
							Description: "Whether the server is a backup server.",
						},
						"maxconn": schema.Int64Attribute{
							Optional:    true,
							Description: "Maximum number of connections for the server.",
						},
						"weight": schema.Int64Attribute{
							Optional:    true,
							Description: "Load balancing weight for the server.",
						},
						"rise": schema.Int64Attribute{
							Optional:    true,
							Description: "Number of successful health checks to mark server as up.",
						},
						"fall": schema.Int64Attribute{
							Optional:    true,
							Description: "Number of failed health checks to mark server as down.",
						},
						"inter": schema.Int64Attribute{
							Optional:    true,
							Description: "Interval between health checks in milliseconds.",
						},
						"fastinter": schema.Int64Attribute{
							Optional:    true,
							Description: "Fast interval between health checks in milliseconds.",
						},
						"downinter": schema.Int64Attribute{
							Optional:    true,
							Description: "Down interval between health checks in milliseconds.",
						},
						"ssl": schema.StringAttribute{
							Optional:    true,
							Description: "SSL configuration for the server.",
						},
						"verify": schema.StringAttribute{
							Optional:    true,
							Description: "SSL verification for the server.",
						},
						"cookie": schema.StringAttribute{
							Optional:    true,
							Description: "Cookie for the server.",
						},
						// SSL/TLS Protocol Control (v3 fields)
						"sslv3": schema.StringAttribute{
							Optional:    true,
							Description: "SSLv3 support for the server (v3 only).",
						},
						"tlsv10": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.0 support for the server (v3 only).",
						},
						"tlsv11": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.1 support for the server (v3 only).",
						},
						"tlsv12": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.2 support for the server (v3 only).",
						},
						"tlsv13": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.3 support for the server (v3 only).",
						},
						// SSL/TLS Protocol Control (deprecated v2 fields)
						"no_sslv3": schema.StringAttribute{
							Optional:    true,
							Description: "Disable SSLv3 for the server (v2 only, deprecated in v3).",
						},
						"no_tlsv10": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.0 for the server (v2 only, deprecated in v3).",
						},
						"no_tlsv11": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.1 for the server (v2 only, deprecated in v3).",
						},
						"no_tlsv12": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.2 for the server (v2 only, deprecated in v3).",
						},
						"no_tlsv13": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.3 for the server (v2 only, deprecated in v3).",
						},
						"force_sslv3": schema.StringAttribute{
							Optional:    true,
							Description: "Force SSLv3 for the server (v2 only, deprecated in v3).",
						},
						"force_tlsv10": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.0 for the server (v2 only, deprecated in v3).",
						},
						"force_tlsv11": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.1 for the server (v2 only, deprecated in v3).",
						},
						"force_tlsv12": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.2 for the server (v2 only, deprecated in v3).",
						},
						"force_tlsv13": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.3 for the server (v2 only, deprecated in v3).",
						},
						"force_strict_sni": schema.StringAttribute{
							Optional:    true,
							Description: "Force strict SNI for the server (v2 only, deprecated in v3).",
						},
					},
				},
			},
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
			"http_checks": schema.ListNestedBlock{
				Description: "HTTP health check configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the httpcheck.",
						},
						"addr": schema.StringAttribute{
							Optional:    true,
							Description: "The address for the HTTP check.",
						},
						"alpn": schema.StringAttribute{
							Optional:    true,
							Description: "The ALPN for the HTTP check.",
						},
						"body": schema.StringAttribute{
							Optional:    true,
							Description: "The body for the HTTP check.",
						},
						"body_log_format": schema.StringAttribute{
							Optional:    true,
							Description: "The body log format for the HTTP check.",
						},
						"check_comment": schema.StringAttribute{
							Optional:    true,
							Description: "The check comment for the HTTP check.",
						},
						"default": schema.BoolAttribute{
							Optional:    true,
							Description: "The default flag for the HTTP check.",
						},
						"error_status": schema.StringAttribute{
							Optional:    true,
							Description: "The error status for the HTTP check.",
						},
						"exclamation_mark": schema.BoolAttribute{
							Optional:    true,
							Description: "The exclamation mark flag for the HTTP check.",
						},
						"headers": schema.ListAttribute{
							Optional:    true,
							Description: "The headers for the HTTP check.",
							ElementType: types.StringType,
						},
						"linger": schema.BoolAttribute{
							Optional:    true,
							Description: "The linger flag for the HTTP check.",
						},
						"match": schema.StringAttribute{
							Optional:    true,
							Description: "The match condition for the health check.",
						},
						"method": schema.StringAttribute{
							Optional:    true,
							Description: "The HTTP method for the health check.",
						},
						"min_recv": schema.Int64Attribute{
							Optional:    true,
							Description: "The minimum receive for the HTTP check.",
						},
						"ok_status": schema.StringAttribute{
							Optional:    true,
							Description: "The OK status for the HTTP check.",
						},
						"on_error": schema.StringAttribute{
							Optional:    true,
							Description: "The on error action for the HTTP check.",
						},
						"on_success": schema.StringAttribute{
							Optional:    true,
							Description: "The on success action for the HTTP check.",
						},
						"pattern": schema.StringAttribute{
							Optional:    true,
							Description: "The pattern to match for the health check.",
						},
						"port": schema.Int64Attribute{
							Optional:    true,
							Description: "The port for the HTTP check.",
						},
						"port_string": schema.StringAttribute{
							Optional:    true,
							Description: "The port string for the HTTP check.",
						},
						"proto": schema.StringAttribute{
							Optional:    true,
							Description: "The protocol for the HTTP check.",
						},
						"send_proxy": schema.BoolAttribute{
							Optional:    true,
							Description: "The send proxy flag for the HTTP check.",
						},
						"sni": schema.StringAttribute{
							Optional:    true,
							Description: "The SNI for the HTTP check.",
						},
						"ssl": schema.BoolAttribute{
							Optional:    true,
							Description: "The SSL flag for the HTTP check.",
						},
						"status_code": schema.StringAttribute{
							Optional:    true,
							Description: "The status code for the HTTP check.",
						},
						"tout_status": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout status for the HTTP check.",
						},
						"uri": schema.StringAttribute{
							Optional:    true,
							Description: "The URI for the HTTP check.",
						},
						"uri_log_format": schema.StringAttribute{
							Optional:    true,
							Description: "The URI log format for the HTTP check.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the HTTP check.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the HTTP check.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the HTTP check.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the HTTP check.",
						},
						"version": schema.StringAttribute{
							Optional:    true,
							Description: "The HTTP version for the HTTP check.",
						},
					},
				},
			},
			"tcp_checks": schema.ListNestedBlock{
				Description: "TCP health check configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The action of the tcp_check.",
						},
						"addr": schema.StringAttribute{
							Optional:    true,
							Description: "The address for the TCP check.",
						},
						"alpn": schema.StringAttribute{
							Optional:    true,
							Description: "The ALPN for the TCP check.",
						},
						"check_comment": schema.StringAttribute{
							Optional:    true,
							Description: "The check comment for the TCP check.",
						},
						"data": schema.StringAttribute{
							Optional:    true,
							Description: "The data for the tcp_check.",
						},
						"default": schema.BoolAttribute{
							Optional:    true,
							Description: "The default flag for the TCP check.",
						},
						"error_status": schema.StringAttribute{
							Optional:    true,
							Description: "The error status for the TCP check.",
						},
						"exclamation_mark": schema.BoolAttribute{
							Optional:    true,
							Description: "The exclamation mark flag for the TCP check.",
						},
						"fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The format for the TCP check.",
						},
						"hex_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The hex format for the TCP check.",
						},
						"hex_string": schema.StringAttribute{
							Optional:    true,
							Description: "The hex string for binary tcp_check.",
						},
						"linger": schema.BoolAttribute{
							Optional:    true,
							Description: "The linger flag for the TCP check.",
						},
						"match": schema.StringAttribute{
							Optional:    true,
							Description: "The match condition for the tcp_check.",
						},
						"min_recv": schema.Int64Attribute{
							Optional:    true,
							Description: "The minimum receive for the TCP check.",
						},
						"ok_status": schema.StringAttribute{
							Optional:    true,
							Description: "The OK status for the TCP check.",
						},
						"on_error": schema.StringAttribute{
							Optional:    true,
							Description: "The on error action for the TCP check.",
						},
						"on_success": schema.StringAttribute{
							Optional:    true,
							Description: "The on success action for the TCP check.",
						},
						"pattern": schema.StringAttribute{
							Optional:    true,
							Description: "The pattern to match for the tcp_check.",
						},
						"port": schema.Int64Attribute{
							Optional:    true,
							Description: "The port for the tcp_check.",
						},
						"port_string": schema.StringAttribute{
							Optional:    true,
							Description: "The port string for the TCP check.",
						},
						"proto": schema.StringAttribute{
							Optional:    true,
							Description: "The protocol for the TCP check.",
						},
						"send_proxy": schema.BoolAttribute{
							Optional:    true,
							Description: "The send proxy flag for the TCP check.",
						},
						"sni": schema.StringAttribute{
							Optional:    true,
							Description: "The SNI for the TCP check.",
						},
						"ssl": schema.BoolAttribute{
							Optional:    true,
							Description: "The SSL flag for the TCP check.",
						},
						"status_code": schema.StringAttribute{
							Optional:    true,
							Description: "The status code for the TCP check.",
						},
						"tout_status": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout status for the TCP check.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the tcp_check.",
						},
						"var_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the TCP check.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the tcp_check.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the TCP check.",
						},
						"via_socks4": schema.BoolAttribute{
							Optional:    true,
							Description: "The via SOCKS4 flag for the TCP check.",
						},
					},
				},
			},
			"acls": schema.ListNestedBlock{
				Description: "Access Control List (ACL) configuration blocks for content switching and decision making in the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"acl_name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the ACL rule.",
						},
						"criterion": schema.StringAttribute{
							Required:    true,
							Description: "The criterion for the ACL rule (e.g., 'path', 'hdr', 'src').",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The value for the ACL rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index/order of the ACL rule. If not specified, will be auto-assigned.",
						},
					},
				},
			},
			"http_request_rules": schema.ListNestedBlock{
				Description: "HTTP request rule configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the http-request rule.",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the http-request rule.",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the http-request rule.",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name of the http-request rule.",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format of the http-request rule.",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match of the http-request rule.",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method of the http-request rule.",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type of the http-request rule.",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value of the http-request rule.",
						},
						"redir_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The redirection code of the http-request rule.",
						},
						"redir_option": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection option of the http-request rule.",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name of the http-request rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit of the http-request rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period of the http-request rule.",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL file of the http-request rule.",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL key format of the http-request rule.",
						},
						"auth_realm": schema.StringAttribute{
							Optional:    true,
							Description: "The authentication realm of the http-request rule.",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name of the http-request rule.",
						},
						"capture_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture ID of the http-request rule.",
						},
						"capture_len": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture length of the http-request rule.",
						},
						"capture_sample": schema.StringAttribute{
							Optional:    true,
							Description: "The capture sample of the http-request rule.",
						},
						"deny_status": schema.Int64Attribute{
							Optional:    true,
							Description: "The deny status of the http-request rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression of the http-request rule.",
						},
						"hint_format": schema.StringAttribute{
							Optional:    true,
							Description: "The hint format of the http-request rule.",
						},
						"hint_name": schema.StringAttribute{
							Optional:    true,
							Description: "The hint name of the http-request rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level of the http-request rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action of the http-request rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters of the http-request rule.",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file of the http-request rule.",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map key format of the http-request rule.",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map value format of the http-request rule.",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value of the http-request rule.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value of the http-request rule.",
						},
						"return_content": schema.StringAttribute{
							Optional:    true,
							Description: "The return content of the http-request rule.",
						},
						"return_content_format": schema.StringAttribute{
							Optional:    true,
							Description: "The return content format of the http-request rule.",
						},
						"return_content_type": schema.StringAttribute{
							Optional:    true,
							Description: "The return content type of the http-request rule.",
						},
						"return_status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The return status code of the http-request rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL of the http-request rule.",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The SC expression of the http-request rule.",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC ID of the http-request rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC index of the http-request rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC integer of the http-request rule.",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE engine of the http-request rule.",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE group of the http-request rule.",
						},
						"status": schema.Int64Attribute{
							Optional:    true,
							Description: "The status of the http-request rule.",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-request rule.",
						},
						"strict_mode": schema.StringAttribute{
							Optional:    true,
							Description: "The strict mode of the http-request rule.",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout of the http-request rule.",
						},
						"timeout_type": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout type of the http-request rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value of the http-request rule.",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC key of the http-request rule.",
						},
						"track_sc_stick_counter": schema.Int64Attribute{
							Optional:    true,
							Description: "The track SC stick counter of the http-request rule.",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC table of the http-request rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression of the http-request rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format of the http-request rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name of the http-request rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope of the http-request rule.",
						},
						"wait_at_least": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait at least of the http-request rule.",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time of the http-request rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index of the http-request rule (for backward compatibility).",
						},
					},
				},
			},
			"http_response_rules": schema.ListNestedBlock{
				Description: "HTTP response rule configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the http-response rule.",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the http-response rule.",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the http-response rule.",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name of the http-response rule.",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format of the http-response rule.",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match of the http-response rule.",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method of the http-response rule.",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type of the http-response rule.",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value of the http-response rule.",
						},
						"redir_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The redirection code of the http-response rule.",
						},
						"redir_option": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection option of the http-response rule.",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name of the http-response rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit of the http-response rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period of the http-response rule.",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL file of the http-response rule.",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL key format of the http-response rule.",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name of the http-response rule.",
						},
						"capture_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture ID of the http-response rule.",
						},
						"capture_len": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture length of the http-response rule.",
						},
						"capture_sample": schema.StringAttribute{
							Optional:    true,
							Description: "The capture sample of the http-response rule.",
						},
						"deny_status": schema.Int64Attribute{
							Optional:    true,
							Description: "The deny status of the http-response rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression of the http-response rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level of the http-response rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action of the http-response rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters of the http-response rule.",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file of the http-response rule.",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map key format of the http-response rule.",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map value format of the http-response rule.",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value of the http-response rule.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value of the http-response rule.",
						},
						"return_content": schema.StringAttribute{
							Optional:    true,
							Description: "The return content of the http-response rule.",
						},
						"return_content_format": schema.StringAttribute{
							Optional:    true,
							Description: "The return content format of the http-response rule.",
						},
						"return_content_type": schema.StringAttribute{
							Optional:    true,
							Description: "The return content type of the http-response rule.",
						},
						"return_status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The return status code of the http-response rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL of the http-response rule.",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The SC expression of the http-response rule.",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC ID of the http-response rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC index of the http-response rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC integer of the http-response rule.",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE engine of the http-response rule.",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE group of the http-response rule.",
						},
						"status": schema.Int64Attribute{
							Optional:    true,
							Description: "The status of the http-response rule.",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-response rule.",
						},
						"strict_mode": schema.StringAttribute{
							Optional:    true,
							Description: "The strict mode of the http-response rule.",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout of the http-response rule.",
						},
						"timeout_type": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout type of the http-response rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value of the http-response rule.",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC key of the http-response rule.",
						},
						"track_sc_stick_counter": schema.Int64Attribute{
							Optional:    true,
							Description: "The track SC stick counter of the http-response rule.",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC table of the http-response rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression of the http-response rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format of the http-response rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name of the http-response rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope of the http-response rule.",
						},
						"wait_at_least": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait at least of the http-response rule.",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time of the http-response rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index of the http-response rule (for backward compatibility).",
						},
					},
				},
			},
			"tcp_request_rules": schema.ListNestedBlock{
				Description: "TCP request rule configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the tcp-request rule.",
						},
						"action": schema.StringAttribute{
							Optional:    true,
							Description: "The action of the tcp-request rule.",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the tcp-request rule.",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the tcp-request rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression for the tcp-request rule.",
						},
						"timeout": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout for the tcp-request rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action for the tcp-request rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters for the tcp-request rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level for the tcp-request rule.",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value for the tcp-request rule.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value for the tcp-request rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value for the tcp-request rule.",
						},
						"capture_len": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture length for the tcp-request rule.",
						},
						"capture_sample": schema.StringAttribute{
							Optional:    true,
							Description: "The capture sample for the tcp-request rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit for the tcp-request rule.",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name for the tcp-request rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period for the tcp-request rule.",
						},
						"resolve_protocol": schema.StringAttribute{
							Optional:    true,
							Description: "The resolve protocol for the tcp-request rule.",
						},
						"resolve_resolvers": schema.StringAttribute{
							Optional:    true,
							Description: "The resolve resolvers for the tcp-request rule.",
						},
						"resolve_var": schema.StringAttribute{
							Optional:    true,
							Description: "The resolve variable for the tcp-request rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL for the tcp-request rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC index for the tcp-request rule.",
						},
						"sc_inc_id": schema.StringAttribute{
							Optional:    true,
							Description: "The SC increment ID for the tcp-request rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC integer for the tcp-request rule.",
						},
						"server_name": schema.StringAttribute{
							Optional:    true,
							Description: "The server name for the tcp-request rule.",
						},
						"service_name": schema.StringAttribute{
							Optional:    true,
							Description: "The service name for the tcp-request rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the tcp-request rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the tcp-request rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the tcp-request rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the tcp-request rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index/order of the tcp-request rule (for backward compatibility).",
						},
					},
				},
			},
			"tcp_response_rules": schema.ListNestedBlock{
				Description: "TCP response rule configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the tcp-response rule.",
						},
						"action": schema.StringAttribute{
							Optional:    true,
							Description: "The action of the tcp-response rule.",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the tcp-response rule.",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the tcp-response rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression for the tcp-response rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level for the tcp-response rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action for the tcp-response rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters for the tcp-response rule.",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value for the tcp-response rule.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value for the tcp-response rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL for the tcp-response rule.",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The SC expression for the tcp-response rule.",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC ID for the tcp-response rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC index for the tcp-response rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC integer for the tcp-response rule.",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE engine for the tcp-response rule.",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE group for the tcp-response rule.",
						},
						"timeout": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout for the tcp-response rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value for the tcp-response rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the tcp-response rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the tcp-response rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the tcp-response rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit for the tcp-response rule.",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name for the tcp-response rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period for the tcp-response rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the tcp-response rule.",
						},
						"capture_len": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture length for the tcp-response rule.",
						},
						"capture_sample": schema.StringAttribute{
							Optional:    true,
							Description: "The capture sample for the tcp-response rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index/order of the tcp-response rule (for backward compatibility).",
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
			"stick_rule": schema.ListNestedBlock{
				Description: "Stick rule configuration for the backend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the stick rule.",
						},
						"type": schema.StringAttribute{
							Optional:    true,
							Description: "The type of the stick rule.",
						},
						"table": schema.StringAttribute{
							Optional:    true,
							Description: "The table for the stick rule.",
						},
						"pattern": schema.StringAttribute{
							Optional:    true,
							Description: "The pattern for the stick rule.",
						},
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
		},
	}
}

// BackendManager handles all backend-related operations
type BackendManager struct {
	client *HAProxyClient
}

// NewBackendManager creates a new BackendManager instance
func CreateBackendManager(client *HAProxyClient) *BackendManager {
	return &BackendManager{
		client: client,
	}
}

// CreateBackend creates a backend with all its components
func (r *BackendManager) CreateBackend(ctx context.Context, plan *haproxyBackendModel) (*BackendPayload, error) {
	// Create the backend payload
	backendPayload := &BackendPayload{
		Name:               plan.Name.ValueString(),
		Mode:               plan.Mode.ValueString(),
		AdvCheck:           r.determineAdvCheckForAPI(plan.AdvCheck, plan.HttpchkParams),
		HttpConnectionMode: plan.HttpConnectionMode.ValueString(),
		ServerTimeout:      plan.ServerTimeout.ValueInt64(),
		CheckTimeout:       plan.CheckTimeout.ValueInt64(),
		ConnectTimeout:     plan.ConnectTimeout.ValueInt64(),
		QueueTimeout:       plan.QueueTimeout.ValueInt64(),
		TunnelTimeout:      plan.TunnelTimeout.ValueInt64(),
		TarpitTimeout:      plan.TarpitTimeout.ValueInt64(),
		CheckCache:         plan.Checkcache.ValueString(),
		Retries:            plan.Retries.ValueInt64(),

		// Process nested blocks (only those supported by BackendPayload)
		Balance:       r.processBalanceBlock(plan.Balance),
		HttpchkParams: r.processHttpchkParamsBlock(plan.HttpchkParams),
		Forwardfor:    r.processForwardforBlock(plan.Forwardfor),
		DefaultServer: r.processDefaultServerBlock(plan.DefaultServer),
		StatsOptions:  r.processStatsOptionsBlock(plan.StatsOptions),
	}

	// Create backend in HAProxy
	err := r.client.CreateBackend(ctx, backendPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	// ACLs handled at stack level for coordinated operations

	return backendPayload, nil
}

// CreateBackendInTransaction creates a backend using an existing transaction ID
func (r *BackendManager) CreateBackendInTransaction(ctx context.Context, transactionID string, plan *haproxyBackendModel) error {
	// Create the backend payload
	backendPayload := &BackendPayload{
		Name:               plan.Name.ValueString(),
		Mode:               plan.Mode.ValueString(),
		AdvCheck:           r.determineAdvCheckForAPI(plan.AdvCheck, plan.HttpchkParams),
		HttpConnectionMode: plan.HttpConnectionMode.ValueString(),
		ServerTimeout:      plan.ServerTimeout.ValueInt64(),
		CheckTimeout:       plan.CheckTimeout.ValueInt64(),
		ConnectTimeout:     plan.ConnectTimeout.ValueInt64(),
		QueueTimeout:       plan.QueueTimeout.ValueInt64(),
		TarpitTimeout:      plan.TarpitTimeout.ValueInt64(),
		CheckCache:         plan.Checkcache.ValueString(),
		Retries:            plan.Retries.ValueInt64(),

		// Process nested blocks (only those supported by BackendPayload)
		Balance:       r.processBalanceBlock(plan.Balance),
		HttpchkParams: r.processHttpchkParamsBlock(plan.HttpchkParams),
		Forwardfor:    r.processForwardforBlock(plan.Forwardfor),
		DefaultServer: r.processDefaultServerBlock(plan.DefaultServer),
		StatsOptions:  r.processStatsOptionsBlock(plan.StatsOptions),
	}

	// Create backend in HAProxy using the existing transaction
	if err := r.client.CreateBackendInTransaction(ctx, transactionID, backendPayload); err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	// ACLs handled at stack level for coordinated operations

	return nil
}

// UpdateBackendInTransaction updates a backend in HAProxy using an existing transaction ID
func (r *BackendManager) UpdateBackendInTransaction(ctx context.Context, transactionID string, plan *haproxyBackendModel) error {
	// Update backend payload
	backendPayload := &BackendPayload{
		Name:               plan.Name.ValueString(),
		Mode:               plan.Mode.ValueString(),
		AdvCheck:           r.determineAdvCheckForAPI(plan.AdvCheck, plan.HttpchkParams),
		HttpConnectionMode: plan.HttpConnectionMode.ValueString(),
		ServerTimeout:      plan.ServerTimeout.ValueInt64(),
		CheckTimeout:       plan.CheckTimeout.ValueInt64(),
		ConnectTimeout:     plan.ConnectTimeout.ValueInt64(),
		QueueTimeout:       plan.QueueTimeout.ValueInt64(),
		TarpitTimeout:      plan.TarpitTimeout.ValueInt64(),
		CheckCache:         plan.Checkcache.ValueString(),
		Retries:            plan.Retries.ValueInt64(),

		// Process nested blocks (only those supported by BackendPayload)
		Balance:       r.processBalanceBlock(plan.Balance),
		HttpchkParams: r.processHttpchkParamsBlock(plan.HttpchkParams),
		Forwardfor:    r.processForwardforBlock(plan.Forwardfor),
		DefaultServer: r.processDefaultServerBlock(plan.DefaultServer),
		StatsOptions:  r.processStatsOptionsBlock(plan.StatsOptions),
	}

	// Update backend in HAProxy using the existing transaction
	err := r.client.UpdateBackendInTransaction(ctx, transactionID, backendPayload)
	if err != nil {
		return fmt.Errorf("failed to update backend: %w", err)
	}

	// ACLs handled at stack level for coordinated operations

	return nil
}

// ReadBackend reads a backend and its components from HAProxy
func (r *BackendManager) ReadBackend(ctx context.Context, backendName string, existingBackend *haproxyBackendModel) (*haproxyBackendModel, error) {
	// Read backend from HAProxy
	backend, err := r.client.ReadBackend(ctx, backendName)
	if err != nil {
		return nil, fmt.Errorf("failed to read backend: %w", err)
	}

	// Check if backend is nil
	if backend == nil {
		return nil, fmt.Errorf("backend %s not found", backendName)
	}

	// Read ACLs for the backend
	var backendAcls []ACLPayload
	aclManager := CreateACLManager(r.client)
	backendAcls, err = aclManager.ReadACLs(ctx, "backend", backendName)
	if err != nil {
		log.Printf("Warning: Failed to read ACLs for backend %s: %v", backendName, err)
		// Continue without ACLs if reading fails
	}

	// Build the backend model
	backendModel := &haproxyBackendModel{
		Name: types.StringValue(backendName),
	}

	// Set basic fields if HAProxy returned them
	if backend.Mode != "" {
		backendModel.Mode = types.StringValue(backend.Mode)
	}
	if backend.HttpConnectionMode != "" {
		backendModel.HttpConnectionMode = types.StringValue(backend.HttpConnectionMode)
	}
	if backend.ServerTimeout != 0 {
		backendModel.ServerTimeout = types.Int64Value(backend.ServerTimeout)
	}
	if backend.CheckTimeout != 0 {
		backendModel.CheckTimeout = types.Int64Value(backend.CheckTimeout)
	}
	if backend.ConnectTimeout != 0 {
		backendModel.ConnectTimeout = types.Int64Value(backend.ConnectTimeout)
	}
	if backend.QueueTimeout != 0 {
		backendModel.QueueTimeout = types.Int64Value(backend.QueueTimeout)
	}
	if backend.TunnelTimeout != 0 {
		backendModel.TunnelTimeout = types.Int64Value(backend.TunnelTimeout)
	}
	if backend.TarpitTimeout != 0 {
		backendModel.TarpitTimeout = types.Int64Value(backend.TarpitTimeout)
	}
	if backend.CheckCache != "" {
		backendModel.Checkcache = types.StringValue(backend.CheckCache)
	}
	if backend.Retries != 0 {
		backendModel.Retries = types.Int64Value(backend.Retries)
	}

	// Handle adv_check based on whether httpchk_params is present
	if existingBackend != nil && len(existingBackend.HttpchkParams) > 0 && existingBackend.AdvCheck.IsNull() {
		// If httpchk_params is configured and adv_check was not explicitly set,
		// adv_check should be "httpchk" but we don't store it in state since it's auto-managed
		backendModel.AdvCheck = types.StringNull()
	} else if existingBackend != nil && !existingBackend.AdvCheck.IsNull() && !existingBackend.AdvCheck.IsUnknown() {
		// Preserve the explicitly configured adv_check value
		backendModel.AdvCheck = existingBackend.AdvCheck
	} else if backend.AdvCheck != "" {
		// Only set adv_check if HAProxy returned it and no explicit configuration
		backendModel.AdvCheck = types.StringValue(backend.AdvCheck)
	} else {
		backendModel.AdvCheck = types.StringNull()
	}

	// Handle default_server configuration
	if backend.DefaultServer != nil {
		// Initialize DefaultServer only if we have data to set
		backendModel.DefaultServer = &haproxyDefaultServerModel{}

		// Only set fields that HAProxy actually returned (non-empty)
		if backend.DefaultServer.Ssl != "" {
			backendModel.DefaultServer.Ssl = types.StringValue(backend.DefaultServer.Ssl)
		}
		if backend.DefaultServer.SslCafile != "" {
			backendModel.DefaultServer.SslCafile = types.StringValue(backend.DefaultServer.SslCafile)
		}
		if backend.DefaultServer.SslCertificate != "" {
			backendModel.DefaultServer.SslCertificate = types.StringValue(backend.DefaultServer.SslCertificate)
		}
		if backend.DefaultServer.SslMaxVer != "" {
			backendModel.DefaultServer.SslMaxVer = types.StringValue(backend.DefaultServer.SslMaxVer)
		}
		if backend.DefaultServer.SslMinVer != "" {
			backendModel.DefaultServer.SslMinVer = types.StringValue(backend.DefaultServer.SslMinVer)
		}
		if backend.DefaultServer.SslReuse != "" {
			backendModel.DefaultServer.SslReuse = types.StringValue(backend.DefaultServer.SslReuse)
		}
		if backend.DefaultServer.Ciphers != "" {
			backendModel.DefaultServer.Ciphers = types.StringValue(backend.DefaultServer.Ciphers)
		}
		if backend.DefaultServer.Ciphersuites != "" {
			backendModel.DefaultServer.Ciphersuites = types.StringValue(backend.DefaultServer.Ciphersuites)
		}
		if backend.DefaultServer.Verify != "" {
			backendModel.DefaultServer.Verify = types.StringValue(backend.DefaultServer.Verify)
		}

		// Protocol control fields (v3 only)
		if backend.DefaultServer.Sslv3 != "" {
			backendModel.DefaultServer.Sslv3 = types.StringValue(backend.DefaultServer.Sslv3)
		}
		if backend.DefaultServer.Tlsv10 != "" {
			backendModel.DefaultServer.Tlsv10 = types.StringValue(backend.DefaultServer.Tlsv10)
		}
		if backend.DefaultServer.Tlsv11 != "" {
			backendModel.DefaultServer.Tlsv11 = types.StringValue(backend.DefaultServer.Tlsv11)
		}
		if backend.DefaultServer.Tlsv12 != "" {
			backendModel.DefaultServer.Tlsv12 = types.StringValue(backend.DefaultServer.Tlsv12)
		}
		if backend.DefaultServer.Tlsv13 != "" {
			backendModel.DefaultServer.Tlsv13 = types.StringValue(backend.DefaultServer.Tlsv13)
		}

		// Deprecated fields (v2 only) - translate from force fields
		if backend.DefaultServer.NoSslv3 != "" {
			backendModel.DefaultServer.NoSslv3 = types.StringValue(backend.DefaultServer.NoSslv3)
		}
		if backend.DefaultServer.NoTlsv10 != "" {
			backendModel.DefaultServer.NoTlsv10 = types.StringValue(backend.DefaultServer.NoTlsv10)
		}
		if backend.DefaultServer.NoTlsv11 != "" {
			backendModel.DefaultServer.NoTlsv11 = types.StringValue(backend.DefaultServer.NoTlsv11)
		}
		if backend.DefaultServer.NoTlsv12 != "" {
			backendModel.DefaultServer.NoTlsv12 = types.StringValue(backend.DefaultServer.NoTlsv12)
		}
		if backend.DefaultServer.NoTlsv13 != "" {
			backendModel.DefaultServer.NoTlsv13 = types.StringValue(backend.DefaultServer.NoTlsv13)
		}

		// Force fields (v3 only) - only set when explicitly "enabled"
		if backend.DefaultServer.ForceSslv3 == "enabled" {
			backendModel.DefaultServer.ForceSslv3 = types.StringValue("enabled")
		}
		if backend.DefaultServer.ForceTlsv10 == "enabled" {
			backendModel.DefaultServer.ForceTlsv10 = types.StringValue("enabled")
		}
		if backend.DefaultServer.ForceTlsv11 == "enabled" {
			backendModel.DefaultServer.ForceTlsv11 = types.StringValue("enabled")
		}
		if backend.DefaultServer.ForceTlsv12 == "enabled" {
			backendModel.DefaultServer.ForceTlsv12 = types.StringValue("enabled")
		}
		if backend.DefaultServer.ForceTlsv13 == "enabled" {
			backendModel.DefaultServer.ForceTlsv13 = types.StringValue("enabled")
		}
		if backend.DefaultServer.ForceStrictSni != "" {
			backendModel.DefaultServer.ForceStrictSni = types.StringValue(backend.DefaultServer.ForceStrictSni)
		}
	}

	// Handle ACLs - prioritize existing state to preserve user's exact order
	if existingBackend != nil && existingBackend.Acls != nil && len(existingBackend.Acls) > 0 {
		// ALWAYS use the existing ACLs from state to preserve user's exact order
		log.Printf("DEBUG: Using existing backend ACLs from state to preserve user's exact order: %s", r.formatAclOrder(existingBackend.Acls))
		backendModel.Acls = existingBackend.Acls
	} else if len(backendAcls) > 0 {
		log.Printf("DEBUG: No existing ACLs in state, creating from HAProxy response")
		var aclModels []haproxyAclModel
		for _, acl := range backendAcls {
			aclModels = append(aclModels, haproxyAclModel{
				AclName:   types.StringValue(acl.AclName),
				Criterion: types.StringValue(acl.Criterion),
				Value:     types.StringValue(acl.Value),
			})
		}
		backendModel.Acls = aclModels
		log.Printf("Backend ACLs created from HAProxy: %s", r.formatAclOrder(aclModels))
	} else if existingBackend != nil {
		// No HAProxy ACLs returned, preserve existing ACLs from state
		log.Printf("No HAProxy ACLs returned, preserving existing backend ACLs")
		backendModel.Acls = existingBackend.Acls
		log.Printf("Existing backend ACLs preserved: %s", r.formatAclOrder(existingBackend.Acls))
	}

	return backendModel, nil
}

// UpdateBackend updates a backend and its components
func (r *BackendManager) UpdateBackend(ctx context.Context, plan *haproxyBackendModel) error {
	// Update backend payload
	backendPayload := &BackendPayload{
		Name:               plan.Name.ValueString(),
		Mode:               plan.Mode.ValueString(),
		AdvCheck:           r.determineAdvCheckForAPI(plan.AdvCheck, plan.HttpchkParams),
		HttpConnectionMode: plan.HttpConnectionMode.ValueString(),
		ServerTimeout:      plan.ServerTimeout.ValueInt64(),
		CheckTimeout:       plan.CheckTimeout.ValueInt64(),
		ConnectTimeout:     plan.ConnectTimeout.ValueInt64(),
		QueueTimeout:       plan.QueueTimeout.ValueInt64(),
		TunnelTimeout:      plan.TunnelTimeout.ValueInt64(),
		TarpitTimeout:      plan.TarpitTimeout.ValueInt64(),
		CheckCache:         plan.Checkcache.ValueString(),
		Retries:            plan.Retries.ValueInt64(),

		// Process nested blocks (only those supported by BackendPayload)
		Balance:       r.processBalanceBlock(plan.Balance),
		HttpchkParams: r.processHttpchkParamsBlock(plan.HttpchkParams),
		Forwardfor:    r.processForwardforBlock(plan.Forwardfor),
		DefaultServer: r.processDefaultServerBlock(plan.DefaultServer),
		StatsOptions:  r.processStatsOptionsBlock(plan.StatsOptions),
	}

	// Update backend in HAProxy
	err := r.client.UpdateBackend(ctx, plan.Name.ValueString(), backendPayload)
	if err != nil {
		return fmt.Errorf("failed to update backend: %w", err)
	}

	// Update ACLs if specified
	if len(plan.Acls) > 0 {
		aclManager := CreateACLManager(r.client)
		if err := aclManager.UpdateACLs(ctx, "backend", plan.Name.ValueString(), plan.Acls); err != nil {
			return fmt.Errorf("failed to update backend ACLs: %w", err)
		}
	}

	return nil
}

// DeleteBackend deletes a backend and its components
func (r *BackendManager) DeleteBackend(ctx context.Context, backendName string) error {
	// Delete ACLs first
	aclManager := CreateACLManager(r.client)
	if err := aclManager.DeleteACLs(ctx, "backend", backendName); err != nil {
		log.Printf("Warning: Failed to delete backend ACLs: %v", err)
		// Continue with backend deletion even if ACL deletion fails
	}

	// Delete backend
	err := r.client.DeleteBackend(ctx, backendName)
	if err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	return nil
}

// DeleteBackendInTransaction deletes a backend using an existing transaction ID
func (r *BackendManager) DeleteBackendInTransaction(ctx context.Context, transactionID, backendName string) error {
	// Delete ACLs first (if any)
	aclManager := CreateACLManager(r.client)
	if err := aclManager.DeleteACLsInTransaction(ctx, transactionID, "backend", backendName); err != nil {
		log.Printf("Warning: Failed to delete backend ACLs: %v", err)
		// Continue with backend deletion even if ACL deletion fails
	}

	// Delete backend in HAProxy using the existing transaction
	err := r.client.DeleteBackendInTransaction(ctx, transactionID, backendName)
	if err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	return nil
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *BackendManager) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return "none"
	}

	var order []string
	for _, acl := range acls {
		order = append(order, acl.AclName.ValueString())
	}
	return strings.Join(order, " → ")
}

// Helper methods for processing nested blocks
func (r *BackendManager) determineAdvCheckForAPI(advCheck types.String, httpchkParams []haproxyHttpchkParamsModel) string {
	if !advCheck.IsNull() && !advCheck.IsUnknown() && advCheck.ValueString() != "" {
		return advCheck.ValueString()
	}
	if len(httpchkParams) > 0 {
		return "httpchk"
	}
	return ""
}

func (r *BackendManager) processBalanceBlock(balance []haproxyBalanceModel) *Balance {
	if len(balance) == 0 {
		return nil
	}
	// For now, just use the first balance if available
	b := balance[0]
	return &Balance{
		Algorithm: b.Algorithm.ValueString(),
	}
}

func (r *BackendManager) processHttpchkParamsBlock(httpchkParams []haproxyHttpchkParamsModel) *HttpchkParams {
	if len(httpchkParams) == 0 {
		return nil
	}
	// For now, just use the first httpchk_params if available
	h := httpchkParams[0]
	return &HttpchkParams{
		Method:  h.Method.ValueString(),
		Uri:     h.Uri.ValueString(),
		Version: h.Version.ValueString(),
	}
}

func (r *BackendManager) processForwardforBlock(forwardfor []haproxyForwardforModel) *ForwardFor {
	if len(forwardfor) == 0 {
		return nil
	}
	// For now, just use the first forwardfor if available
	f := forwardfor[0]
	return &ForwardFor{
		Enabled: f.Enabled.ValueString(),
	}
}

// Helper functions for other nested blocks are not needed for BackendPayload

func (r *BackendManager) processDefaultServerBlock(defaultServer *haproxyDefaultServerModel) *DefaultServerPayload {
	if defaultServer == nil {
		return nil
	}

	payload := &DefaultServerPayload{}

	// Core SSL fields (supported in both v2 and v3) - only set if not null/unknown
	if !defaultServer.Ssl.IsNull() && !defaultServer.Ssl.IsUnknown() {
		payload.Ssl = defaultServer.Ssl.ValueString()
	}
	if !defaultServer.SslCafile.IsNull() && !defaultServer.SslCafile.IsUnknown() {
		payload.SslCafile = defaultServer.SslCafile.ValueString()
	}
	if !defaultServer.SslCertificate.IsNull() && !defaultServer.SslCertificate.IsUnknown() {
		payload.SslCertificate = defaultServer.SslCertificate.ValueString()
	}
	if !defaultServer.SslMaxVer.IsNull() && !defaultServer.SslMaxVer.IsUnknown() {
		payload.SslMaxVer = defaultServer.SslMaxVer.ValueString()
	}
	if !defaultServer.SslMinVer.IsNull() && !defaultServer.SslMinVer.IsUnknown() {
		payload.SslMinVer = defaultServer.SslMinVer.ValueString()
	}
	if !defaultServer.SslReuse.IsNull() && !defaultServer.SslReuse.IsUnknown() {
		payload.SslReuse = defaultServer.SslReuse.ValueString()
	}
	if !defaultServer.Ciphers.IsNull() && !defaultServer.Ciphers.IsUnknown() {
		payload.Ciphers = defaultServer.Ciphers.ValueString()
	}
	if !defaultServer.Ciphersuites.IsNull() && !defaultServer.Ciphersuites.IsUnknown() {
		payload.Ciphersuites = defaultServer.Ciphersuites.ValueString()
	}
	if !defaultServer.Verify.IsNull() && !defaultServer.Verify.IsUnknown() {
		payload.Verify = defaultServer.Verify.ValueString()
	}

	// Protocol control fields (v3 only) - only set if not null/unknown and API v3
	apiVersion := r.client.GetAPIVersion()
	if apiVersion == "v3" {
		if !defaultServer.Sslv3.IsNull() && !defaultServer.Sslv3.IsUnknown() {
			payload.Sslv3 = defaultServer.Sslv3.ValueString()
		}
		if !defaultServer.Tlsv10.IsNull() && !defaultServer.Tlsv10.IsUnknown() {
			payload.Tlsv10 = defaultServer.Tlsv10.ValueString()
		}
		if !defaultServer.Tlsv11.IsNull() && !defaultServer.Tlsv11.IsUnknown() {
			payload.Tlsv11 = defaultServer.Tlsv11.ValueString()
		}
		if !defaultServer.Tlsv12.IsNull() && !defaultServer.Tlsv12.IsUnknown() {
			payload.Tlsv12 = defaultServer.Tlsv12.ValueString()
		}
		if !defaultServer.Tlsv13.IsNull() && !defaultServer.Tlsv13.IsUnknown() {
			payload.Tlsv13 = defaultServer.Tlsv13.ValueString()
		}
	}

	// Deprecated fields (v2 only) - translate to force fields - only set if not null/unknown and API v2
	if apiVersion == "v2" {
		if !defaultServer.NoSslv3.IsNull() && !defaultServer.NoSslv3.IsUnknown() {
			payload.NoSslv3 = r.translateNoTlsToForceTls(defaultServer.NoSslv3.ValueString())
		}
		if !defaultServer.NoTlsv10.IsNull() && !defaultServer.NoTlsv10.IsUnknown() {
			payload.NoTlsv10 = r.translateNoTlsToForceTls(defaultServer.NoTlsv10.ValueString())
		}
		if !defaultServer.NoTlsv11.IsNull() && !defaultServer.NoTlsv11.IsUnknown() {
			payload.NoTlsv11 = r.translateNoTlsToForceTls(defaultServer.NoTlsv11.ValueString())
		}
		if !defaultServer.NoTlsv12.IsNull() && !defaultServer.NoTlsv12.IsUnknown() {
			payload.NoTlsv12 = r.translateNoTlsToForceTls(defaultServer.NoTlsv12.ValueString())
		}
		if !defaultServer.NoTlsv13.IsNull() && !defaultServer.NoTlsv13.IsUnknown() {
			payload.NoTlsv13 = r.translateNoTlsToForceTls(defaultServer.NoTlsv13.ValueString())
		}
	}

	// Force fields (v3 only) - only set when explicitly "enabled" and API v3
	if apiVersion == "v3" {
		if !defaultServer.ForceSslv3.IsNull() && !defaultServer.ForceSslv3.IsUnknown() && defaultServer.ForceSslv3.ValueString() == "enabled" {
			payload.ForceSslv3 = "enabled"
		}
		if !defaultServer.ForceTlsv10.IsNull() && !defaultServer.ForceTlsv10.IsUnknown() && defaultServer.ForceTlsv10.ValueString() == "enabled" {
			payload.ForceTlsv10 = "enabled"
		}
		if !defaultServer.ForceTlsv11.IsNull() && !defaultServer.ForceTlsv11.IsUnknown() && defaultServer.ForceTlsv11.ValueString() == "enabled" {
			payload.ForceTlsv11 = "enabled"
		}
		if !defaultServer.ForceTlsv12.IsNull() && !defaultServer.ForceTlsv12.IsUnknown() && defaultServer.ForceTlsv12.ValueString() == "enabled" {
			payload.ForceTlsv12 = "enabled"
		}
		if !defaultServer.ForceTlsv13.IsNull() && !defaultServer.ForceTlsv13.IsUnknown() && defaultServer.ForceTlsv13.ValueString() == "enabled" {
			payload.ForceTlsv13 = "enabled"
		}
		if !defaultServer.ForceStrictSni.IsNull() && !defaultServer.ForceStrictSni.IsUnknown() {
			payload.ForceStrictSni = defaultServer.ForceStrictSni.ValueString()
		}
	}

	return payload
}

// Helper functions for stick table and stick rule are not needed for BackendPayload

func (r *BackendManager) processStatsOptionsBlock(statsOptions []haproxyStatsOptionsModel) *StatsOptionsPayload {
	if len(statsOptions) == 0 {
		return nil
	}
	// For now, just use the first stats option if available
	s := statsOptions[0]
	return &StatsOptionsPayload{
		StatsEnable:      s.StatsEnable.ValueBool(),
		StatsHideVersion: false, // Default values
		StatsShowLegends: true,
		StatsShowNode:    true,
		StatsUri:         s.StatsUri.ValueString(),
		StatsRealm:       s.StatsRealm.ValueString(),
		StatsAuth:        s.StatsAuth.ValueString(),
		StatsRefresh:     "2s", // Default value
	}
}

const (
	enabledValue  = "enabled"
	disabledValue = "disabled"
)

// translateNoTlsToForceTls translates no_tlsv* fields to force_tlsv* fields
func (r *BackendManager) translateNoTlsToForceTls(noTlsValue string) string {
	if noTlsValue == enabledValue {
		return disabledValue // "Don't allow TLSv1.x" → "Force disabled"
	} else if noTlsValue == disabledValue {
		return enabledValue // "Allow TLSv1.x" → "Force enabled"
	}
	return noTlsValue // Return as-is for other values
}
