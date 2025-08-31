package haproxy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
