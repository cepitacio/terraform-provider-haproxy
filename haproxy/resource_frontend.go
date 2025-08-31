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
		},
		Blocks: map[string]schema.Block{
			"bind": schema.ListNestedBlock{
				Description: "Bind configuration for the frontend.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
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
						"sslv3": schema.StringAttribute{
							Optional:    true,
							Description: "SSLv3 support for the bind.",
						},
						"tlsv10": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.0 support for the bind.",
						},
						"tlsv11": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.1 support for the bind.",
						},
						"tlsv12": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.2 support for the bind.",
						},
						"tlsv13": schema.StringAttribute{
							Optional:    true,
							Description: "TLSv1.3 support for the bind.",
						},
						"no_sslv3": schema.StringAttribute{
							Optional:    true,
							Description: "Disable SSLv3 for the bind.",
						},
						"no_tlsv10": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.0 for the bind.",
						},
						"no_tlsv11": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.1 for the bind.",
						},
						"no_tlsv12": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.2 for the bind.",
						},
						"no_tlsv13": schema.StringAttribute{
							Optional:    true,
							Description: "Disable TLSv1.3 for the bind.",
						},
						"force_sslv3": schema.StringAttribute{
							Optional:    true,
							Description: "Force SSLv3 for the bind.",
						},
						"force_tlsv10": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.0 for the bind.",
						},
						"force_tlsv11": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.1 for the bind.",
						},
						"force_tlsv12": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.2 for the bind.",
						},
						"force_tlsv13": schema.StringAttribute{
							Optional:    true,
							Description: "Force TLSv1.3 for the bind.",
						},
						"force_strict_sni": schema.StringAttribute{
							Optional:    true,
							Description: "Force strict SNI for the bind.",
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

	// Create ACLs if specified
	if plan.Acls != nil && len(plan.Acls) > 0 {
		// Use ACLManager to create ACLs
		aclManager := NewACLManager(r.client)
		if err := aclManager.CreateACLs(ctx, "frontend", plan.Name.ValueString(), plan.Acls); err != nil {
			return nil, fmt.Errorf("failed to create frontend ACLs: %w", err)
		}
	}

	return frontendPayload, nil
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
