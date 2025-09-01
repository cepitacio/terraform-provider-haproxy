package haproxy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetFrontendSchema returns the schema for the frontend block
func GetFrontendSchema(schemaBuilder *VersionAwareSchemaBuilder) schema.SingleNestedBlock {
	// If no schema builder is provided, include all fields for backward compatibility
	if schemaBuilder == nil {
		schemaBuilder = NewVersionAwareSchemaBuilder("v2") // Default to v2
	}
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
		},
		Blocks: map[string]schema.Block{
			"bind": schema.ListNestedBlock{
				Description: "Bind configuration for the frontend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the bind.",
						},
						"address": schema.StringAttribute{
							Required:    true,
							Description: "The bind address.",
						},
						"port": schema.Int64Attribute{
							Required:    true,
							Description: "The bind port.",
						},
						"ssl": schema.StringAttribute{
							Optional:    true,
							Description: "SSL configuration for the bind.",
						},
						"ssl_cafile": schema.StringAttribute{
							Optional:    true,
							Description: "SSL CA file for the bind.",
						},
						"ssl_certificate": schema.StringAttribute{
							Optional:    true,
							Description: "SSL certificate for the bind.",
						},
						"ssl_max_ver": schema.StringAttribute{
							Optional:    true,
							Description: "SSL maximum version for the bind.",
						},
						"ssl_min_ver": schema.StringAttribute{
							Optional:    true,
							Description: "SSL minimum version for the bind.",
						},
						"ssl_reuse": schema.StringAttribute{
							Optional:    true,
							Description: "SSL reuse configuration for the bind.",
						},
						"ciphers": schema.StringAttribute{
							Optional:    true,
							Description: "Ciphers for the bind.",
						},
						"ciphersuites": schema.StringAttribute{
							Optional:    true,
							Description: "Cipher suites for the bind.",
						},
						"verify": schema.StringAttribute{
							Optional:    true,
							Description: "SSL verification for the bind.",
						},
						"transparent": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether the bind is transparent.",
						},
						// v3 fields
						"sslv3": schema.BoolAttribute{
							Optional:    true,
							Description: "SSLv3 support for the bind (Data Plane API v3 only).",
						},
						"tlsv10": schema.BoolAttribute{
							Optional:    true,
							Description: "TLSv1.0 support for the bind (Data Plane API v3 only).",
						},
						"tlsv11": schema.BoolAttribute{
							Optional:    true,
							Description: "TLSv1.1 support for the bind (Data Plane API v3 only).",
						},
						"tlsv12": schema.BoolAttribute{
							Optional:    true,
							Description: "TLSv1.2 support for the bind (Data Plane API v3 only).",
						},
						"tlsv13": schema.BoolAttribute{
							Optional:    true,
							Description: "TLSv1.3 support for the bind (Data Plane API v3 only).",
						},
						// v2 fields (deprecated in v3)
						"force_sslv3": schema.BoolAttribute{
							Optional:    true,
							Description: "Force SSLv3 for the bind (Data Plane API v2 only, deprecated in v3).",
						},
						"force_tlsv10": schema.BoolAttribute{
							Optional:    true,
							Description: "Force TLSv1.0 for the bind (Data Plane API v2 only, deprecated in v3).",
						},
						"force_tlsv11": schema.BoolAttribute{
							Optional:    true,
							Description: "Force TLSv1.1 for the bind (Data Plane API v2 only, deprecated in v3).",
						},
						"force_tlsv12": schema.BoolAttribute{
							Optional:    true,
							Description: "Force TLSv1.2 for the bind (Data Plane API v2 only, deprecated in v3).",
						},
						"force_tlsv13": schema.BoolAttribute{
							Optional:    true,
							Description: "Force TLSv1.3 for the bind (Data Plane API v2 only, deprecated in v3).",
						},
						"force_strict_sni": schema.StringAttribute{
							Optional:    true,
							Description: "Force strict SNI for the bind (Data Plane API v2 only, deprecated in v3).",
						},
					},
				},
			},
			"acls": schema.ListNestedBlock{
				Description: "Access Control List (ACL) configuration blocks for content switching and decision making in the frontend.",
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
				Description: "HTTP request rule configuration for the frontend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index/order of the HTTP request rule.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the HTTP request rule.",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the HTTP request rule (if, unless).",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the HTTP request rule.",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name for the HTTP request rule.",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format for the HTTP request rule.",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match pattern for the HTTP request rule.",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type (location, prefix, scheme).",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value.",
						},
						"redir_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The redirection HTTP status code (301, 302, 303, 307, 308).",
						},
						"redir_option": schema.StringAttribute{
							Optional:    true,
							Description: "Additional redirection options.",
						},
						"path_match": schema.StringAttribute{
							Optional:    true,
							Description: "The path match pattern for the HTTP request rule.",
						},
						"path_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The path format for the HTTP request rule.",
						},
						"uri_match": schema.StringAttribute{
							Optional:    true,
							Description: "The URI match pattern for the HTTP request rule.",
						},
						"uri_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The URI format for the HTTP request rule.",
						},
						"query_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The query format for the HTTP request rule.",
						},
						"method_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The method format for the HTTP request rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the HTTP request rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the HTTP request rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the HTTP request rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the HTTP request rule.",
						},
						"capture_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture ID for the HTTP request rule.",
						},
						"capture_len": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture length for the HTTP request rule.",
						},
						"capture_sample": schema.StringAttribute{
							Optional:    true,
							Description: "The capture sample for the HTTP request rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level for the HTTP request rule.",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout for the HTTP request rule.",
						},
						"timeout_type": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout type (server, tunnel, client).",
						},
						"strict_mode": schema.StringAttribute{
							Optional:    true,
							Description: "The strict mode setting (on, off).",
						},
						"normalizer": schema.StringAttribute{
							Optional:    true,
							Description: "The URI normalizer setting.",
						},
						"normalizer_full": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether to use full normalizer.",
						},
						"normalizer_strict": schema.BoolAttribute{
							Optional:    true,
							Description: "Whether to use strict normalizer.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value for the HTTP request rule (-1024 to 1024).",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value for the HTTP request rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value for the HTTP request rule.",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter key for tracking.",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter table for tracking.",
						},
						"track_sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The stick counter ID for tracking.",
						},
						"track_sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The stick counter index for tracking.",
						},
						"track_sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The stick counter integer for tracking.",
						},
						"return_status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The return status code (200-599).",
						},
						"return_content": schema.StringAttribute{
							Optional:    true,
							Description: "The return content for the HTTP request rule.",
						},
						"return_content_type": schema.StringAttribute{
							Optional:    true,
							Description: "The return content type for the HTTP request rule.",
						},
						"return_content_format": schema.StringAttribute{
							Optional:    true,
							Description: "The return content format for the HTTP request rule.",
						},
						"deny_status": schema.Int64Attribute{
							Optional:    true,
							Description: "The deny status code (200-599).",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time for the HTTP request rule.",
						},
						"wait_at_least": schema.Int64Attribute{
							Optional:    true,
							Description: "The minimum wait time for the HTTP request rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression for the HTTP request rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action for the HTTP request rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters for the HTTP request rule.",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE engine for the HTTP request rule.",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE group for the HTTP request rule.",
						},
						"service_name": schema.StringAttribute{
							Optional:    true,
							Description: "The service name for the HTTP request rule.",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name for the HTTP request rule.",
						},
						"resolvers": schema.StringAttribute{
							Optional:    true,
							Description: "The resolvers for the HTTP request rule.",
						},
						"protocol": schema.StringAttribute{
							Optional:    true,
							Description: "The protocol for the HTTP request rule (ipv4, ipv6).",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name for the HTTP request rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit value for the HTTP request rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period for the HTTP request rule.",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file for the HTTP request rule.",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map key format for the HTTP request rule.",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map value format for the HTTP request rule.",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL file for the HTTP request rule.",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL key format for the HTTP request rule.",
						},
						"auth_realm": schema.StringAttribute{
							Optional:    true,
							Description: "The authentication realm for the HTTP request rule.",
						},
						"hint_name": schema.StringAttribute{
							Optional:    true,
							Description: "The hint name for the HTTP request rule.",
						},
						"hint_format": schema.StringAttribute{
							Optional:    true,
							Description: "The hint format for the HTTP request rule.",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter expression for the HTTP request rule.",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The stick counter ID for the HTTP request rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The stick counter index for the HTTP request rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The stick counter integer for the HTTP request rule.",
						},
						"sc_add_gpc": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter add GPC for the HTTP request rule.",
						},
						"sc_inc_gpc": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter increment GPC for the HTTP request rule.",
						},
						"sc_inc_gpc0": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter increment GPC0 for the HTTP request rule.",
						},
						"sc_inc_gpc1": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter increment GPC1 for the HTTP request rule.",
						},
						"sc_set_gpt": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter set GPT for the HTTP request rule.",
						},
						"sc_set_gpt0": schema.StringAttribute{
							Optional:    true,
							Description: "The stick counter set GPT0 for the HTTP request rule.",
						},
						"set_priority_class": schema.StringAttribute{
							Optional:    true,
							Description: "The priority class for the HTTP request rule.",
						},
						"set_priority_offset": schema.StringAttribute{
							Optional:    true,
							Description: "The priority offset for the HTTP request rule.",
						},
						"set_retries": schema.StringAttribute{
							Optional:    true,
							Description: "The retries setting for the HTTP request rule.",
						},
						"set_bc_mark": schema.StringAttribute{
							Optional:    true,
							Description: "The backend connection mark for the HTTP request rule.",
						},
						"set_bc_tos": schema.StringAttribute{
							Optional:    true,
							Description: "The backend connection TOS for the HTTP request rule.",
						},
						"set_fc_mark": schema.StringAttribute{
							Optional:    true,
							Description: "The frontend connection mark for the HTTP request rule.",
						},
						"set_fc_tos": schema.StringAttribute{
							Optional:    true,
							Description: "The frontend connection TOS for the HTTP request rule.",
						},
						"set_dst": schema.StringAttribute{
							Optional:    true,
							Description: "The destination setting for the HTTP request rule.",
						},
						"set_dst_port": schema.StringAttribute{
							Optional:    true,
							Description: "The destination port setting for the HTTP request rule.",
						},
						"set_src": schema.StringAttribute{
							Optional:    true,
							Description: "The source setting for the HTTP request rule.",
						},
						"set_src_port": schema.StringAttribute{
							Optional:    true,
							Description: "The source port setting for the HTTP request rule.",
						},
						"set_timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout setting for the HTTP request rule.",
						},
						"set_tos": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS setting for the HTTP request rule.",
						},
						"set_mark": schema.StringAttribute{
							Optional:    true,
							Description: "The mark setting for the HTTP request rule.",
						},
						"set_var": schema.StringAttribute{
							Optional:    true,
							Description: "The variable setting for the HTTP request rule.",
						},
						"set_var_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format setting for the HTTP request rule.",
						},
						"unset_var": schema.StringAttribute{
							Optional:    true,
							Description: "The variable unsetting for the HTTP request rule.",
						},
						"early_hint": schema.StringAttribute{
							Optional:    true,
							Description: "The early hint for the HTTP request rule.",
						},
						"use_service": schema.StringAttribute{
							Optional:    true,
							Description: "The service to use for the HTTP request rule.",
						},
						"wait_for_body": schema.StringAttribute{
							Optional:    true,
							Description: "The wait for body setting for the HTTP request rule.",
						},
						"wait_for_handshake": schema.StringAttribute{
							Optional:    true,
							Description: "The wait for handshake setting for the HTTP request rule.",
						},
						"silent_drop": schema.StringAttribute{
							Optional:    true,
							Description: "The silent drop setting for the HTTP request rule.",
						},
						"tarpit": schema.StringAttribute{
							Optional:    true,
							Description: "The tarpit setting for the HTTP request rule.",
						},
						"disable_l7_retry": schema.StringAttribute{
							Optional:    true,
							Description: "The disable L7 retry setting for the HTTP request rule.",
						},
						"do_resolve": schema.StringAttribute{
							Optional:    true,
							Description: "The do resolve setting for the HTTP request rule.",
						},
						"send_spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The send SPOE group setting for the HTTP request rule.",
						},
						"replace_header": schema.StringAttribute{
							Optional:    true,
							Description: "The replace header setting for the HTTP request rule.",
						},
						"replace_path": schema.StringAttribute{
							Optional:    true,
							Description: "The replace path setting for the HTTP request rule.",
						},
						"replace_pathq": schema.StringAttribute{
							Optional:    true,
							Description: "The replace path query setting for the HTTP request rule.",
						},
						"replace_uri": schema.StringAttribute{
							Optional:    true,
							Description: "The replace URI setting for the HTTP request rule.",
						},
						"replace_value": schema.StringAttribute{
							Optional:    true,
							Description: "The replace value setting for the HTTP request rule.",
						},
						"add_header": schema.StringAttribute{
							Optional:    true,
							Description: "The add header setting for the HTTP request rule.",
						},
						"del_header": schema.StringAttribute{
							Optional:    true,
							Description: "The delete header setting for the HTTP request rule.",
						},
						"add_acl": schema.StringAttribute{
							Optional:    true,
							Description: "The add ACL setting for the HTTP request rule.",
						},
						"del_acl": schema.StringAttribute{
							Optional:    true,
							Description: "The delete ACL setting for the HTTP request rule.",
						},
						"set_map": schema.StringAttribute{
							Optional:    true,
							Description: "The set map setting for the HTTP request rule.",
						},
						"del_map": schema.StringAttribute{
							Optional:    true,
							Description: "The delete map setting for the HTTP request rule.",
						},
						"cache_use": schema.StringAttribute{
							Optional:    true,
							Description: "The cache use setting for the HTTP request rule.",
						},
						"capture": schema.StringAttribute{
							Optional:    true,
							Description: "The capture setting for the HTTP request rule.",
						},
						"auth": schema.StringAttribute{
							Optional:    true,
							Description: "The authentication setting for the HTTP request rule.",
						},
						"allow": schema.StringAttribute{
							Optional:    true,
							Description: "The allow setting for the HTTP request rule.",
						},
						"deny": schema.StringAttribute{
							Optional:    true,
							Description: "The deny setting for the HTTP request rule.",
						},
						"return": schema.StringAttribute{
							Optional:    true,
							Description: "The return setting for the HTTP request rule.",
						},
						"reject": schema.StringAttribute{
							Optional:    true,
							Description: "The reject setting for the HTTP request rule.",
						},
						"pause": schema.StringAttribute{
							Optional:    true,
							Description: "The pause setting for the HTTP request rule.",
						},
						"normalize_uri": schema.StringAttribute{
							Optional:    true,
							Description: "The normalize URI setting for the HTTP request rule.",
						},
						"set_method": schema.StringAttribute{
							Optional:    true,
							Description: "The set method setting for the HTTP request rule.",
						},
						"set_query": schema.StringAttribute{
							Optional:    true,
							Description: "The set query setting for the HTTP request rule.",
						},
						"set_uri": schema.StringAttribute{
							Optional:    true,
							Description: "The set URI setting for the HTTP request rule.",
						},
						"set_log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The set log level setting for the HTTP request rule.",
						},
						"set_bandwidth_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The set bandwidth limit setting for the HTTP request rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL setting for the HTTP request rule.",
						},
					},
					Blocks: map[string]schema.Block{
						"return_hdrs": schema.ListNestedBlock{
							Description: "Return headers configuration for the HTTP request rule.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: "The header name.",
									},
									"fmt": schema.StringAttribute{
										Required:    true,
										Description: "The header format.",
									},
								},
							},
						},
					},
				},
			},
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
		},
	}
}

// FrontendManager handles all frontend-related operations
type FrontendManager struct {
	client *HAProxyClient
}

// NewFrontendManager creates a new FrontendManager instance
func NewFrontendManager(client *HAProxyClient) *FrontendManager {
	return &FrontendManager{
		client: client,
	}
}

// CreateFrontend creates a frontend with all its components
func (r *FrontendManager) CreateFrontend(ctx context.Context, plan *haproxyFrontendModel) (*FrontendPayload, error) {
	// Create the frontend payload
	frontendPayload := &FrontendPayload{
		Name:           plan.Name.ValueString(),
		Mode:           plan.Mode.ValueString(),
		DefaultBackend: plan.DefaultBackend.ValueString(),
		MaxConn:        plan.Maxconn.ValueInt64(),
		Backlog:        plan.Backlog.ValueInt64(),
	}

	// Create frontend in HAProxy
	err := r.client.CreateFrontend(ctx, frontendPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to create frontend: %w", err)
	}

	return frontendPayload, nil
}

// CreateFrontendInTransaction creates a frontend using an existing transaction ID
func (r *FrontendManager) CreateFrontendInTransaction(ctx context.Context, transactionID string, plan *haproxyFrontendModel) error {
	// Create the frontend payload
	frontendPayload := &FrontendPayload{
		Name:           plan.Name.ValueString(),
		Mode:           plan.Mode.ValueString(),
		DefaultBackend: plan.DefaultBackend.ValueString(),
		MaxConn:        plan.Maxconn.ValueInt64(),
		Backlog:        plan.Backlog.ValueInt64(),
	}

	// Create frontend in HAProxy using the existing transaction
	if err := r.client.CreateFrontendInTransaction(ctx, transactionID, frontendPayload); err != nil {
		return fmt.Errorf("failed to create frontend: %w", err)
	}

	// Create ACLs AFTER frontend exists
	if plan.Acls != nil && len(plan.Acls) > 0 {
		aclManager := NewACLManager(r.client)
		if err := aclManager.CreateACLsInTransaction(ctx, transactionID, "frontend", plan.Name.ValueString(), plan.Acls); err != nil {
			return fmt.Errorf("failed to create frontend ACLs: %w", err)
		}
	}

	// Note: HTTP request rules are now handled at the stack level, not here
	return nil
}

// UpdateFrontendInTransaction updates a frontend using an existing transaction ID
func (r *FrontendManager) UpdateFrontendInTransaction(ctx context.Context, transactionID string, plan *haproxyFrontendModel) error {
	// Update frontend payload
	frontendPayload := &FrontendPayload{
		Name:           plan.Name.ValueString(),
		Mode:           plan.Mode.ValueString(),
		DefaultBackend: plan.DefaultBackend.ValueString(),
		MaxConn:        plan.Maxconn.ValueInt64(),
		Backlog:        plan.Backlog.ValueInt64(),
	}

	// Update frontend in HAProxy using the existing transaction
	err := r.client.UpdateFrontendInTransaction(ctx, transactionID, frontendPayload)
			if err != nil {
		return fmt.Errorf("failed to update frontend: %w", err)
	}

	// Update ACLs if specified
	if plan.Acls != nil && len(plan.Acls) > 0 {
		aclManager := NewACLManager(r.client)
		if err := aclManager.UpdateACLsInTransaction(ctx, transactionID, "frontend", plan.Name.ValueString(), plan.Acls); err != nil {
			return fmt.Errorf("failed to update frontend ACLs: %w", err)
		}
	}

	// Note: HTTP request rules are now handled at the stack level, not here
	return nil
}

// DeleteFrontendInTransaction deletes a frontend using an existing transaction ID
func (r *FrontendManager) DeleteFrontendInTransaction(ctx context.Context, transactionID string, frontendName string) error {
	// Delete ACLs first (if any)
	aclManager := NewACLManager(r.client)
	if err := aclManager.DeleteACLsInTransaction(ctx, transactionID, "frontend", frontendName); err != nil {
		log.Printf("Warning: Failed to delete frontend ACLs: %v", err)
		// Continue with frontend deletion even if ACL deletion fails
	}

	// Delete frontend in HAProxy using the existing transaction
	err := r.client.DeleteFrontendInTransaction(ctx, transactionID, frontendName)
			if err != nil {
		return fmt.Errorf("failed to delete frontend: %w", err)
	}

	return nil
}

// ReadFrontend reads a frontend and its components from HAProxy
func (r *FrontendManager) ReadFrontend(ctx context.Context, frontendName string, existingFrontend *haproxyFrontendModel) (*haproxyFrontendModel, error) {
	// Read frontend from HAProxy
	frontend, err := r.client.ReadFrontend(ctx, frontendName)
			if err != nil {
		return nil, fmt.Errorf("failed to read frontend: %w", err)
	}

	// Read ACLs for the frontend
	var frontendAcls []ACLPayload
	if frontend != nil {
		aclManager := NewACLManager(r.client)
		frontendAcls, err = aclManager.ReadACLs(ctx, "frontend", frontendName)
			if err != nil {
			log.Printf("Warning: Failed to read ACLs for frontend %s: %v", frontendName, err)
			// Continue without ACLs if reading fails
		}
	}

	// Build the frontend model
	frontendModel := &haproxyFrontendModel{
		Name:           types.StringValue(frontendName),
		Mode:           types.StringValue(frontend.Mode),
		DefaultBackend: types.StringValue(frontend.DefaultBackend),
		Maxconn:        types.Int64Value(frontend.MaxConn),
		Backlog:        types.Int64Value(frontend.Backlog),
	}

	// Handle ACLs - prioritize existing state to preserve user's exact order
	if existingFrontend != nil && existingFrontend.Acls != nil && len(existingFrontend.Acls) > 0 {
		// ALWAYS use the existing ACLs from state to preserve user's exact order
		log.Printf("DEBUG: Using existing frontend ACLs from state to preserve user's exact order: %s", r.formatAclOrder(existingFrontend.Acls))
		frontendModel.Acls = existingFrontend.Acls
	} else if len(frontendAcls) > 0 {
		log.Printf("DEBUG: No existing ACLs in state, creating from HAProxy response")
		var aclModels []haproxyAclModel
		for _, acl := range frontendAcls {
			aclModels = append(aclModels, haproxyAclModel{
				AclName:   types.StringValue(acl.AclName),
				Criterion: types.StringValue(acl.Criterion),
				Value:     types.StringValue(acl.Value),
				Index:     types.Int64Value(acl.Index),
			})
		}
		frontendModel.Acls = aclModels
		log.Printf("Frontend ACLs created from HAProxy: %s", r.formatAclOrder(aclModels))
	} else if existingFrontend != nil {
		// No HAProxy ACLs returned, preserve existing ACLs from state
		log.Printf("No HAProxy ACLs returned, preserving existing frontend ACLs")
		frontendModel.Acls = existingFrontend.Acls
		log.Printf("Existing frontend ACLs preserved: %s", r.formatAclOrder(existingFrontend.Acls))
	}

	// Handle HTTP Request Rules - prioritize existing state to preserve user's exact order
	if existingFrontend != nil && existingFrontend.HttpRequestRules != nil && len(existingFrontend.HttpRequestRules) > 0 {
		// ALWAYS use the existing HTTP request rules from state to preserve user's exact order
		log.Printf("DEBUG: Using existing frontend HTTP request rules from state to preserve user's exact order: %s", r.formatHttpRequestRuleOrder(existingFrontend.HttpRequestRules))
		frontendModel.HttpRequestRules = existingFrontend.HttpRequestRules
	} else {
		// Read HTTP request rules from HAProxy
		httpRequestRuleManager := NewHttpRequestRuleManager(r.client)
		httpRequestRules, err := httpRequestRuleManager.ReadHttpRequestRules(ctx, "frontend", frontendName)
			if err != nil {
			log.Printf("Warning: Failed to read HTTP request rules for frontend %s: %v", frontendName, err)
			// Continue without HTTP request rules if reading fails
		} else if len(httpRequestRules) > 0 {
			log.Printf("DEBUG: Creating frontend HTTP request rules from HAProxy response")
			var ruleModels []haproxyHttpRequestRuleModel
			for _, rule := range httpRequestRules {
				ruleModels = append(ruleModels, r.convertHttpRequestRulePayloadToModel(&rule))
			}
			frontendModel.HttpRequestRules = ruleModels
			log.Printf("Frontend HTTP request rules created from HAProxy: %s", r.formatHttpRequestRuleOrder(ruleModels))
		}
	}

	return frontendModel, nil
}

// UpdateFrontend updates a frontend and its components
func (r *FrontendManager) UpdateFrontend(ctx context.Context, plan *haproxyFrontendModel) error {
	// Update frontend payload
	frontendPayload := &FrontendPayload{
		Name:           plan.Name.ValueString(),
		Mode:           plan.Mode.ValueString(),
		DefaultBackend: plan.DefaultBackend.ValueString(),
		MaxConn:        plan.Maxconn.ValueInt64(),
		Backlog:        plan.Backlog.ValueInt64(),
	}

	// Update frontend in HAProxy
	err := r.client.UpdateFrontend(ctx, plan.Name.ValueString(), frontendPayload)
	if err != nil {
		return fmt.Errorf("failed to update frontend: %w", err)
	}

	// Update ACLs if specified
	if plan.Acls != nil && len(plan.Acls) > 0 {
		aclManager := NewACLManager(r.client)
		if err := aclManager.UpdateACLs(ctx, "frontend", plan.Name.ValueString(), plan.Acls); err != nil {
			return fmt.Errorf("failed to update frontend ACLs: %w", err)
		}
	}

	// Update HTTP Request Rules if specified
	if plan.HttpRequestRules != nil && len(plan.HttpRequestRules) > 0 {
		httpRequestRuleManager := NewHttpRequestRuleManager(r.client)
		if err := httpRequestRuleManager.UpdateHttpRequestRules(ctx, "frontend", plan.Name.ValueString(), plan.HttpRequestRules); err != nil {
			return fmt.Errorf("failed to update frontend HTTP request rules: %w", err)
		}
	}

	return nil
}

// DeleteFrontend deletes a frontend and its components
func (r *FrontendManager) DeleteFrontend(ctx context.Context, frontendName string) error {
	// Delete ACLs first
	aclManager := NewACLManager(r.client)
	if err := aclManager.DeleteACLs(ctx, "frontend", frontendName); err != nil {
		log.Printf("Warning: Failed to delete frontend ACLs: %v", err)
		// Continue with frontend deletion even if ACL deletion fails
	}

	// Delete HTTP Request Rules first
	httpRequestRuleManager := NewHttpRequestRuleManager(r.client)
	if err := httpRequestRuleManager.DeleteHttpRequestRules(ctx, "frontend", frontendName); err != nil {
		log.Printf("Warning: Failed to delete frontend HTTP request rules: %v", err)
		// Continue with frontend deletion even if HTTP request rules deletion fails
	}

	// Delete frontend
	err := r.client.DeleteFrontend(ctx, frontendName)
	if err != nil {
		return fmt.Errorf("failed to delete frontend: %w", err)
	}

	return nil
}

// processFrontendBlock processes the frontend block configuration
func (r *FrontendManager) processFrontendBlock(frontend *haproxyFrontendModel) *FrontendPayload {
	if frontend == nil {
		return nil
	}

	return &FrontendPayload{
		Name:           frontend.Name.ValueString(),
		Mode:           frontend.Mode.ValueString(),
		DefaultBackend: frontend.DefaultBackend.ValueString(),
		MaxConn:        frontend.Maxconn.ValueInt64(),
		Backlog:        frontend.Backlog.ValueInt64(),
	}
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *FrontendManager) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return "none"
	}

	var order []string
	for _, acl := range acls {
		order = append(order, fmt.Sprintf("%s(index:%d)", acl.AclName.ValueString(), acl.Index.ValueInt64()))
	}
	return strings.Join(order, " → ")
}

// formatHttpRequestRuleOrder formats the order of HTTP request rules for logging
func (r *FrontendManager) formatHttpRequestRuleOrder(rules []haproxyHttpRequestRuleModel) string {
	if len(rules) == 0 {
		return "[]"
	}

	var order []string
	for _, rule := range rules {
		order = append(order, fmt.Sprintf("%s(%d)", rule.Type.ValueString(), rule.Index.ValueInt64()))
	}
	return fmt.Sprintf("[%s]", strings.Join(order, ", "))
}

// convertHttpRequestRulePayloadToModel converts HAProxy API payload to Terraform model
func (r *FrontendManager) convertHttpRequestRulePayloadToModel(payload *HttpRequestRulePayload) haproxyHttpRequestRuleModel {
	model := haproxyHttpRequestRuleModel{
		Index: types.Int64Value(payload.Index),
		Type:  types.StringValue(payload.Type),
	}

	// Set optional fields only if they have values
	if payload.Cond != "" {
		model.Cond = types.StringValue(payload.Cond)
	}
	if payload.CondTest != "" {
		model.CondTest = types.StringValue(payload.CondTest)
	}
	if payload.HdrName != "" {
		model.HdrName = types.StringValue(payload.HdrName)
	}
	if payload.HdrFormat != "" {
		model.HdrFormat = types.StringValue(payload.HdrFormat)
	}
	if payload.HdrMatch != "" {
		model.HdrMatch = types.StringValue(payload.HdrMatch)
	}
	if payload.RedirType != "" {
		model.RedirType = types.StringValue(payload.RedirType)
	}
	if payload.RedirValue != "" {
		model.RedirValue = types.StringValue(payload.RedirValue)
	}
	if payload.RedirCode != 0 {
		model.RedirCode = types.Int64Value(payload.RedirCode)
	}
	if payload.RedirOption != "" {
		model.RedirOption = types.StringValue(payload.RedirOption)
	}
	if payload.PathMatch != "" {
		model.PathMatch = types.StringValue(payload.PathMatch)
	}
	if payload.PathFmt != "" {
		model.PathFmt = types.StringValue(payload.PathFmt)
	}
	if payload.UriMatch != "" {
		model.UriMatch = types.StringValue(payload.UriMatch)
	}
	if payload.UriFmt != "" {
		model.UriFmt = types.StringValue(payload.UriFmt)
	}
	if payload.QueryFmt != "" {
		model.QueryFmt = types.StringValue(payload.QueryFmt)
	}
	if payload.MethodFmt != "" {
		model.MethodFmt = types.StringValue(payload.MethodFmt)
	}
	if payload.VarName != "" {
		model.VarName = types.StringValue(payload.VarName)
	}
	if payload.VarFormat != "" {
		model.VarFormat = types.StringValue(payload.VarFormat)
	}
	if payload.VarExpr != "" {
		model.VarExpr = types.StringValue(payload.VarExpr)
	}
	if payload.VarScope != "" {
		model.VarScope = types.StringValue(payload.VarScope)
	}
	if payload.CaptureID != 0 {
		model.CaptureID = types.Int64Value(payload.CaptureID)
	}
	if payload.CaptureLen != 0 {
		model.CaptureLen = types.Int64Value(payload.CaptureLen)
	}
	if payload.CaptureSample != "" {
		model.CaptureSample = types.StringValue(payload.CaptureSample)
	}
	if payload.LogLevel != "" {
		model.LogLevel = types.StringValue(payload.LogLevel)
	}
	if payload.Timeout != "" {
		model.Timeout = types.StringValue(payload.Timeout)
	}
	if payload.TimeoutType != "" {
		model.TimeoutType = types.StringValue(payload.TimeoutType)
	}
	if payload.StrictMode != "" {
		model.StrictMode = types.StringValue(payload.StrictMode)
	}
	if payload.Normalizer != "" {
		model.Normalizer = types.StringValue(payload.Normalizer)
	}
	if payload.NormalizerFull {
		model.NormalizerFull = types.BoolValue(payload.NormalizerFull)
	}
	if payload.NormalizerStrict {
		model.NormalizerStrict = types.BoolValue(payload.NormalizerStrict)
	}
	if payload.NiceValue != 0 {
		model.NiceValue = types.Int64Value(payload.NiceValue)
	}
	if payload.MarkValue != "" {
		model.MarkValue = types.StringValue(payload.MarkValue)
	}
	if payload.TosValue != "" {
		model.TosValue = types.StringValue(payload.TosValue)
	}
	if payload.TrackScKey != "" {
		model.TrackScKey = types.StringValue(payload.TrackScKey)
	}
	if payload.TrackScTable != "" {
		model.TrackScTable = types.StringValue(payload.TrackScTable)
	}
	if payload.TrackScID != 0 {
		model.TrackScID = types.Int64Value(payload.TrackScID)
	}
	if payload.TrackScIdx != 0 {
		model.TrackScIdx = types.Int64Value(payload.TrackScIdx)
	}
	if payload.TrackScInt != 0 {
		model.TrackScInt = types.Int64Value(payload.TrackScInt)
	}
	if payload.ReturnStatusCode != 0 {
		model.ReturnStatusCode = types.Int64Value(payload.ReturnStatusCode)
	}
	if payload.ReturnContent != "" {
		model.ReturnContent = types.StringValue(payload.ReturnContent)
	}
	if payload.ReturnContentType != "" {
		model.ReturnContentType = types.StringValue(payload.ReturnContentType)
	}
	if payload.ReturnContentFormat != "" {
		model.ReturnContentFormat = types.StringValue(payload.ReturnContentFormat)
	}
	if payload.DenyStatus != 0 {
		model.DenyStatus = types.Int64Value(payload.DenyStatus)
	}
	if payload.WaitTime != 0 {
		model.WaitTime = types.Int64Value(payload.WaitTime)
	}
	if payload.WaitAtLeast != 0 {
		model.WaitAtLeast = types.Int64Value(payload.WaitAtLeast)
	}
	if payload.Expr != "" {
		model.Expr = types.StringValue(payload.Expr)
	}
	if payload.LuaAction != "" {
		model.LuaAction = types.StringValue(payload.LuaAction)
	}
	if payload.LuaParams != "" {
		model.LuaParams = types.StringValue(payload.LuaParams)
	}
	if payload.SpoeEngine != "" {
		model.SpoeEngine = types.StringValue(payload.SpoeEngine)
	}
	if payload.SpoeGroup != "" {
		model.SpoeGroup = types.StringValue(payload.SpoeGroup)
	}
	if payload.ServiceName != "" {
		model.ServiceName = types.StringValue(payload.ServiceName)
	}
	if payload.CacheName != "" {
		model.CacheName = types.StringValue(payload.CacheName)
	}
	if payload.Resolvers != "" {
		model.Resolvers = types.StringValue(payload.Resolvers)
	}
	if payload.Protocol != "" {
		model.Protocol = types.StringValue(payload.Protocol)
	}
	if payload.BandwidthLimitName != "" {
		model.BandwidthLimitName = types.StringValue(payload.BandwidthLimitName)
	}
	if payload.BandwidthLimitLimit != "" {
		model.BandwidthLimitLimit = types.StringValue(payload.BandwidthLimitLimit)
	}
	if payload.BandwidthLimitPeriod != "" {
		model.BandwidthLimitPeriod = types.StringValue(payload.BandwidthLimitPeriod)
	}
	if payload.MapFile != "" {
		model.MapFile = types.StringValue(payload.MapFile)
	}
	if payload.MapKeyfmt != "" {
		model.MapKeyfmt = types.StringValue(payload.MapKeyfmt)
	}
	if payload.MapValuefmt != "" {
		model.MapValuefmt = types.StringValue(payload.MapValuefmt)
	}
	if payload.AclFile != "" {
		model.AclFile = types.StringValue(payload.AclFile)
	}
	if payload.AclKeyfmt != "" {
		model.AclKeyfmt = types.StringValue(payload.AclKeyfmt)
	}
	if payload.AuthRealm != "" {
		model.AuthRealm = types.StringValue(payload.AuthRealm)
	}
	if payload.HintName != "" {
		model.HintName = types.StringValue(payload.HintName)
	}
	if payload.HintFormat != "" {
		model.HintFormat = types.StringValue(payload.HintFormat)
	}
	if payload.ScExpr != "" {
		model.ScExpr = types.StringValue(payload.ScExpr)
	}
	if payload.ScID != 0 {
		model.ScID = types.Int64Value(payload.ScID)
	}
	if payload.ScIdx != 0 {
		model.ScIdx = types.Int64Value(payload.ScIdx)
	}
	if payload.ScInt != 0 {
		model.ScInt = types.Int64Value(payload.ScInt)
	}
	if payload.ScAddGpc != "" {
		model.ScAddGpc = types.StringValue(payload.ScAddGpc)
	}
	if payload.ScIncGpc != "" {
		model.ScIncGpc = types.StringValue(payload.ScIncGpc)
	}
	if payload.ScIncGpc0 != "" {
		model.ScIncGpc0 = types.StringValue(payload.ScIncGpc0)
	}
	if payload.ScIncGpc1 != "" {
		model.ScIncGpc1 = types.StringValue(payload.ScIncGpc1)
	}
	if payload.ScSetGpt != "" {
		model.ScSetGpt = types.StringValue(payload.ScSetGpt)
	}
	if payload.ScSetGpt0 != "" {
		model.ScSetGpt0 = types.StringValue(payload.ScSetGpt0)
	}
	if payload.SetPriorityClass != "" {
		model.SetPriorityClass = types.StringValue(payload.SetPriorityClass)
	}
	if payload.SetPriorityOffset != "" {
		model.SetPriorityOffset = types.StringValue(payload.SetPriorityOffset)
	}
	if payload.SetRetries != "" {
		model.SetRetries = types.StringValue(payload.SetRetries)
	}
	if payload.SetBcMark != "" {
		model.SetBcMark = types.StringValue(payload.SetBcMark)
	}
	if payload.SetBcTos != "" {
		model.SetBcTos = types.StringValue(payload.SetBcTos)
	}
	if payload.SetFcMark != "" {
		model.SetFcMark = types.StringValue(payload.SetFcMark)
	}
	if payload.SetFcTos != "" {
		model.SetFcTos = types.StringValue(payload.SetFcTos)
	}
	if payload.SetDst != "" {
		model.SetDst = types.StringValue(payload.SetDst)
	}
	if payload.SetDstPort != 0 {
		model.SetDstPort = types.StringValue(fmt.Sprintf("%d", payload.SetDstPort))
	}
	if payload.SetSrc != "" {
		model.SetSrc = types.StringValue(payload.SetSrc)
	}
	if payload.SetSrcPort != 0 {
		model.SetSrcPort = types.StringValue(fmt.Sprintf("%d", payload.SetSrcPort))
	}
	if payload.SetTimeout != "" {
		model.SetTimeout = types.StringValue(payload.SetTimeout)
	}
	if payload.SetTos != "" {
		model.SetTos = types.StringValue(payload.SetTos)
	}
	if payload.SetMark != "" {
		model.SetMark = types.StringValue(payload.SetMark)
	}
	if payload.SetVar != "" {
		model.SetVar = types.StringValue(payload.SetVar)
	}
	if payload.SetVarFmt != "" {
		model.SetVarFmt = types.StringValue(payload.SetVarFmt)
	}
	if payload.UnsetVar != "" {
		model.UnsetVar = types.StringValue(payload.UnsetVar)
	}
	if payload.EarlyHint != "" {
		model.EarlyHint = types.StringValue(payload.EarlyHint)
	}
	if payload.UseService != "" {
		model.UseService = types.StringValue(payload.UseService)
	}
	if payload.WaitForBody != "" {
		model.WaitForBody = types.StringValue(payload.WaitForBody)
	}
	if payload.WaitForHandshake != "" {
		model.WaitForHandshake = types.StringValue(payload.WaitForHandshake)
	}
	if payload.SilentDrop != "" {
		model.SilentDrop = types.StringValue(payload.SilentDrop)
	}
	if payload.Tarpit != "" {
		model.Tarpit = types.StringValue(payload.Tarpit)
	}
	if payload.DisableL7Retry != "" {
		model.DisableL7Retry = types.StringValue(payload.DisableL7Retry)
	}
	if payload.DoResolve != "" {
		model.DoResolve = types.StringValue(payload.DoResolve)
	}
	if payload.SendSpoeGroup != "" {
		model.SendSpoeGroup = types.StringValue(payload.SendSpoeGroup)
	}
	if payload.ReplaceHeader != "" {
		model.ReplaceHeader = types.StringValue(payload.ReplaceHeader)
	}
	if payload.ReplacePath != "" {
		model.ReplacePath = types.StringValue(payload.ReplacePath)
	}
	if payload.ReplacePathq != "" {
		model.ReplacePathq = types.StringValue(payload.ReplacePathq)
	}
	if payload.ReplaceUri != "" {
		model.ReplaceUri = types.StringValue(payload.ReplaceUri)
	}
	if payload.ReplaceValue != "" {
		model.ReplaceValue = types.StringValue(payload.ReplaceValue)
	}
	if payload.AddHeader != "" {
		model.AddHeader = types.StringValue(payload.AddHeader)
	}
	if payload.DelHeader != "" {
		model.DelHeader = types.StringValue(payload.DelHeader)
	}
	if payload.AddAcl != "" {
		model.AddAcl = types.StringValue(payload.AddAcl)
	}
	if payload.DelAcl != "" {
		model.DelAcl = types.StringValue(payload.DelAcl)
	}
	if payload.SetMap != "" {
		model.SetMap = types.StringValue(payload.SetMap)
	}
	if payload.DelMap != "" {
		model.DelMap = types.StringValue(payload.DelMap)
	}
	if payload.CacheUse != "" {
		model.CacheUse = types.StringValue(payload.CacheUse)
	}
	if payload.Capture != "" {
		model.Capture = types.StringValue(payload.Capture)
	}
	if payload.Auth != "" {
		model.Auth = types.StringValue(payload.Auth)
	}
	if payload.Allow != "" {
		model.Allow = types.StringValue(payload.Allow)
	}
	if payload.Deny != "" {
		model.Deny = types.StringValue(payload.Deny)
	}
	if payload.Return != "" {
		model.Return = types.StringValue(payload.Return)
	}
	if payload.Reject != "" {
		model.Reject = types.StringValue(payload.Reject)
	}
	if payload.Pause != "" {
		model.Pause = types.StringValue(payload.Pause)
	}
	if payload.NormalizeUri != "" {
		model.NormalizeUri = types.StringValue(payload.NormalizeUri)
	}
	if payload.SetMethod != "" {
		model.SetMethod = types.StringValue(payload.SetMethod)
	}
	if payload.SetQuery != "" {
		model.SetQuery = types.StringValue(payload.SetQuery)
	}
	if payload.SetUri != "" {
		model.SetUri = types.StringValue(payload.SetUri)
	}
	if payload.SetLogLevel != "" {
		model.SetLogLevel = types.StringValue(payload.SetLogLevel)
	}
	if payload.SetBandwidthLimit != "" {
		model.SetBandwidthLimit = types.StringValue(payload.SetBandwidthLimit)
	}
	if payload.RstTtl != 0 {
		model.RstTtl = types.Int64Value(payload.RstTtl)
	}

	// Handle return headers if present
	if payload.ReturnHdrs != nil && len(payload.ReturnHdrs) > 0 {
		var returnHdrs []haproxyReturnHdrModel
		for _, hdr := range payload.ReturnHdrs {
			returnHdrs = append(returnHdrs, haproxyReturnHdrModel{
				Name: types.StringValue(hdr.Name),
				Fmt:  types.StringValue(hdr.Fmt),
			})
		}
		model.ReturnHdrs = returnHdrs
	}

	return model
}
