package haproxy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetFrontendSchema returns the schema for the frontend block
func GetFrontendSchema(schemaBuilder *VersionAwareSchemaBuilder) schema.SingleNestedBlock {
	// If no schema builder is provided, include all fields for backward compatibility
	if schemaBuilder == nil {
		schemaBuilder = CreateVersionAwareSchemaBuilder("v3") // Default to v3
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
			"monitor_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The URI to use for health monitoring of the frontend.",
			},
			"binds": GetBindSchema(),
		},
		Blocks: map[string]schema.Block{
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
							Description: "The header match for the HTTP request rule.",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method for the HTTP request rule.",
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
							Description: "The redirection code for the HTTP request rule.",
						},
						"redir_option": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection option for the HTTP request rule.",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name for the HTTP request rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit for the HTTP request rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period for the HTTP request rule.",
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
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name for the HTTP request rule.",
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
						"deny_status": schema.Int64Attribute{
							Optional:    true,
							Description: "The deny status for the HTTP request rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression for the HTTP request rule.",
						},
						"hint_format": schema.StringAttribute{
							Optional:    true,
							Description: "The hint format for the HTTP request rule.",
						},
						"hint_name": schema.StringAttribute{
							Optional:    true,
							Description: "The hint name for the HTTP request rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level for the HTTP request rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action for the HTTP request rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters for the HTTP request rule.",
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
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value for the HTTP request rule.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value for the HTTP request rule.",
						},
						"return_content": schema.StringAttribute{
							Optional:    true,
							Description: "The return content for the HTTP request rule.",
						},
						"return_content_format": schema.StringAttribute{
							Optional:    true,
							Description: "The return content format for the HTTP request rule.",
						},
						"return_content_type": schema.StringAttribute{
							Optional:    true,
							Description: "The return content type for the HTTP request rule.",
						},
						"return_status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The return status code for the HTTP request rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL for the HTTP request rule.",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The SC expression for the HTTP request rule.",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC ID for the HTTP request rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC index for the HTTP request rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC integer for the HTTP request rule.",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE engine for the HTTP request rule.",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE group for the HTTP request rule.",
						},
						"status": schema.Int64Attribute{
							Optional:    true,
							Description: "The status for the HTTP request rule.",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason for the HTTP request rule.",
						},
						"strict_mode": schema.StringAttribute{
							Optional:    true,
							Description: "The strict mode for the HTTP request rule.",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout for the HTTP request rule.",
						},
						"timeout_type": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout type for the HTTP request rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value for the HTTP request rule.",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC key for the HTTP request rule.",
						},
						"track_sc_stick_counter": schema.Int64Attribute{
							Optional:    true,
							Description: "The track SC stick counter for the HTTP request rule.",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC table for the HTTP request rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the HTTP request rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the HTTP request rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the HTTP request rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the HTTP request rule.",
						},
						"wait_at_least": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait at least for the HTTP request rule.",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time for the HTTP request rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index/order of the HTTP request rule (for backward compatibility).",
						},
					},
				},
			},
			"http_response_rules": schema.ListNestedBlock{
				Description: "HTTP response rule configuration for the frontend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the HTTP response rule.",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the HTTP response rule (if, unless).",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the HTTP response rule.",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name for the HTTP response rule.",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format for the HTTP response rule.",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match for the HTTP response rule.",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method for the HTTP response rule.",
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
							Description: "The redirection code for the HTTP response rule.",
						},
						"redir_option": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection option for the HTTP response rule.",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name for the HTTP response rule.",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit for the HTTP response rule.",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period for the HTTP response rule.",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL file for the HTTP response rule.",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The ACL key format for the HTTP response rule.",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name for the HTTP response rule.",
						},
						"capture_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture ID for the HTTP response rule.",
						},
						"capture_len": schema.Int64Attribute{
							Optional:    true,
							Description: "The capture length for the HTTP response rule.",
						},
						"capture_sample": schema.StringAttribute{
							Optional:    true,
							Description: "The capture sample for the HTTP response rule.",
						},
						"deny_status": schema.Int64Attribute{
							Optional:    true,
							Description: "The deny status for the HTTP response rule.",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression for the HTTP response rule.",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level for the HTTP response rule.",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua action for the HTTP response rule.",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The Lua parameters for the HTTP response rule.",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file for the HTTP response rule.",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map key format for the HTTP response rule.",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map value format for the HTTP response rule.",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value for the HTTP response rule.",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value for the HTTP response rule.",
						},
						"return_content": schema.StringAttribute{
							Optional:    true,
							Description: "The return content for the HTTP response rule.",
						},
						"return_content_format": schema.StringAttribute{
							Optional:    true,
							Description: "The return content format for the HTTP response rule.",
						},
						"return_content_type": schema.StringAttribute{
							Optional:    true,
							Description: "The return content type for the HTTP response rule.",
						},
						"return_status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The return status code for the HTTP response rule.",
						},
						"rst_ttl": schema.Int64Attribute{
							Optional:    true,
							Description: "The RST TTL for the HTTP response rule.",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The SC expression for the HTTP response rule.",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC ID for the HTTP response rule.",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC index for the HTTP response rule.",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The SC integer for the HTTP response rule.",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE engine for the HTTP response rule.",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The SPOE group for the HTTP response rule.",
						},
						"status": schema.Int64Attribute{
							Optional:    true,
							Description: "The status for the HTTP response rule.",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason for the HTTP response rule.",
						},
						"strict_mode": schema.StringAttribute{
							Optional:    true,
							Description: "The strict mode for the HTTP response rule.",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout for the HTTP response rule.",
						},
						"timeout_type": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout type for the HTTP response rule.",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The TOS value for the HTTP response rule.",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC key for the HTTP response rule.",
						},
						"track_sc_stick_counter": schema.Int64Attribute{
							Optional:    true,
							Description: "The track SC stick counter for the HTTP response rule.",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track SC table for the HTTP response rule.",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression for the HTTP response rule.",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format for the HTTP response rule.",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name for the HTTP response rule.",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope for the HTTP response rule.",
						},
						"wait_at_least": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait at least for the HTTP response rule.",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time for the HTTP response rule.",
						},
						"index": schema.Int64Attribute{
							Optional:    true,
							Description: "The index/order of the HTTP response rule (for backward compatibility).",
						},
					},
				},
			},
			"tcp_request_rules": schema.ListNestedBlock{
				Description: "TCP request rule configuration for the frontend.",
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
		},
	}
}

// FrontendManager handles all frontend-related operations
type FrontendManager struct {
	client *HAProxyClient
}

// NewFrontendManager creates a new FrontendManager instance
func CreateFrontendManager(client *HAProxyClient) *FrontendManager {
	return &FrontendManager{
		client: client,
	}
}

// CreateFrontend creates a frontend with all its components
func (r *FrontendManager) CreateFrontend(ctx context.Context, plan *haproxyFrontendModel) (*FrontendPayload, error) {
	// Create the frontend payload
	frontendPayload := r.processFrontendBlock(plan)

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
	frontendPayload := r.processFrontendBlock(plan)

	// Create frontend in HAProxy using the existing transaction
	if err := r.client.CreateFrontendInTransaction(ctx, transactionID, frontendPayload); err != nil {
		return fmt.Errorf("failed to create frontend: %w", err)
	}

	// ACLs handled at stack level for coordinated operations

	// HTTP request rules handled at stack level for coordinated operations
	return nil
}

// UpdateFrontendInTransaction updates a frontend using an existing transaction ID
func (r *FrontendManager) UpdateFrontendInTransaction(ctx context.Context, transactionID string, plan *haproxyFrontendModel) error {
	// Update frontend payload
	frontendPayload := r.processFrontendBlock(plan)

	// Update frontend in HAProxy using the existing transaction
	err := r.client.UpdateFrontendInTransaction(ctx, transactionID, frontendPayload)
	if err != nil {
		return fmt.Errorf("failed to update frontend: %w", err)
	}

	// ACLs handled at stack level for coordinated operations

	// HTTP request rules handled at stack level for coordinated operations
	return nil
}

// DeleteFrontendInTransaction deletes a frontend using an existing transaction ID
func (r *FrontendManager) DeleteFrontendInTransaction(ctx context.Context, transactionID string, frontendName string) error {
	// Delete ACLs first (if any)
	aclManager := CreateACLManager(r.client)
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
		aclManager := CreateACLManager(r.client)
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
		MonitorFail:    r.convertMonitorFailFromPayload(frontend.MonitorFail),
		MonitorUri:     types.StringValue(frontend.MonitorUri),
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
		httpRequestRuleManager := CreateHttpRequestRuleManager(r.client)
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
	frontendPayload := r.processFrontendBlock(plan)

	// Update frontend in HAProxy
	err := r.client.UpdateFrontend(ctx, plan.Name.ValueString(), frontendPayload)
	if err != nil {
		return fmt.Errorf("failed to update frontend: %w", err)
	}

	// Update ACLs if specified
	if plan.Acls != nil && len(plan.Acls) > 0 {
		aclManager := CreateACLManager(r.client)
		if err := aclManager.UpdateACLs(ctx, "frontend", plan.Name.ValueString(), plan.Acls); err != nil {
			return fmt.Errorf("failed to update frontend ACLs: %w", err)
		}
	}

	// Update HTTP Request Rules if specified
	if plan.HttpRequestRules != nil && len(plan.HttpRequestRules) > 0 {
		httpRequestRuleManager := CreateHttpRequestRuleManager(r.client)
		if err := httpRequestRuleManager.UpdateHttpRequestRules(ctx, "frontend", plan.Name.ValueString(), plan.HttpRequestRules); err != nil {
			return fmt.Errorf("failed to update frontend HTTP request rules: %w", err)
		}
	}

	return nil
}

// DeleteFrontend deletes a frontend and its components
func (r *FrontendManager) DeleteFrontend(ctx context.Context, frontendName string) error {
	// Delete ACLs first
	aclManager := CreateACLManager(r.client)
	if err := aclManager.DeleteACLs(ctx, "frontend", frontendName); err != nil {
		log.Printf("Warning: Failed to delete frontend ACLs: %v", err)
		// Continue with frontend deletion even if ACL deletion fails
	}

	// Delete HTTP Request Rules first
	httpRequestRuleManager := CreateHttpRequestRuleManager(r.client)
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

	log.Printf("DEBUG: processFrontendBlock - Frontend model MonitorFail field: %+v (length: %d)", frontend.MonitorFail, len(frontend.MonitorFail))
	monitorFail := r.processMonitorFailBlock(frontend.MonitorFail)
	log.Printf("DEBUG: processFrontendBlock - MonitorFail result: %+v", monitorFail)

	payload := &FrontendPayload{
		Name:           frontend.Name.ValueString(),
		Mode:           frontend.Mode.ValueString(),
		DefaultBackend: frontend.DefaultBackend.ValueString(),
		MaxConn:        frontend.Maxconn.ValueInt64(),
		Backlog:        frontend.Backlog.ValueInt64(),
		MonitorFail:    monitorFail,
		MonitorUri:     frontend.MonitorUri.ValueString(),
	}

	log.Printf("DEBUG: processFrontendBlock - Final payload MonitorFail: %+v", payload.MonitorFail)
	return payload
}

// processMonitorFailBlock processes the monitor_fail block configuration
func (r *FrontendManager) processMonitorFailBlock(monitorFail []haproxyMonitorFailModel) *MonitorFailPayload {
	log.Printf("DEBUG: processMonitorFailBlock called with %d monitor_fail blocks", len(monitorFail))
	if len(monitorFail) == 0 {
		log.Printf("DEBUG: No monitor_fail blocks, returning nil")
		return nil
	}
	// Use the first monitor_fail block (should only be one due to SizeAtMost(1) validator)
	mf := monitorFail[0]
	payload := &MonitorFailPayload{
		Cond:     mf.Cond.ValueString(),
		CondTest: mf.CondTest.ValueString(),
	}
	log.Printf("DEBUG: Created MonitorFailPayload: %+v", payload)
	return payload
}

// convertMonitorFailFromPayload converts MonitorFailPayload to haproxyMonitorFailModel
func (r *FrontendManager) convertMonitorFailFromPayload(monitorFail *MonitorFailPayload) []haproxyMonitorFailModel {
	if monitorFail == nil {
		return nil
	}
	return []haproxyMonitorFailModel{
		{
			Cond:     types.StringValue(monitorFail.Cond),
			CondTest: types.StringValue(monitorFail.CondTest),
		},
	}
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *FrontendManager) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return "none"
	}

	var order []string
	for _, acl := range acls {
		order = append(order, acl.AclName.ValueString())
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
		order = append(order, rule.Type.ValueString())
	}
	return fmt.Sprintf("[%s]", strings.Join(order, ", "))
}

// convertHttpRequestRulePayloadToModel converts HAProxy API payload to Terraform model
func (r *FrontendManager) convertHttpRequestRulePayloadToModel(payload *HttpRequestRulePayload) haproxyHttpRequestRuleModel {
	model := haproxyHttpRequestRuleModel{
		Type: types.StringValue(payload.Type),
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
	if payload.RedirType != "" {
		model.RedirType = types.StringValue(payload.RedirType)
	}
	if payload.RedirValue != "" {
		model.RedirValue = types.StringValue(payload.RedirValue)
	}

	return model
}
