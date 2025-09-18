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
func (r *FrontendManager) DeleteFrontendInTransaction(ctx context.Context, transactionID, frontendName string) error {
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
		Name: types.StringValue(frontendName),
	}

	// Set fields only if they exist
	if frontend != nil && frontend.Mode != "" {
		frontendModel.Mode = types.StringValue(frontend.Mode)
	}
	if frontend != nil && frontend.DefaultBackend != "" {
		frontendModel.DefaultBackend = types.StringValue(frontend.DefaultBackend)
	}
	if frontend != nil && frontend.MaxConn != 0 {
		frontendModel.Maxconn = types.Int64Value(frontend.MaxConn)
	}
	if frontend != nil && frontend.Backlog != 0 {
		frontendModel.Backlog = types.Int64Value(frontend.Backlog)
	}
	if frontend != nil && frontend.MonitorUri != "" {
		frontendModel.MonitorUri = types.StringValue(frontend.MonitorUri)
	}
	if frontend != nil {
		frontendModel.MonitorFail = r.convertMonitorFailFromPayload(frontend.MonitorFail)
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
	if len(plan.Acls) > 0 {
		aclManager := CreateACLManager(r.client)
		if err := aclManager.UpdateACLs(ctx, "frontend", plan.Name.ValueString(), plan.Acls); err != nil {
			return fmt.Errorf("failed to update frontend ACLs: %w", err)
		}
	}

	// Update HTTP Request Rules if specified
	if len(plan.HttpRequestRules) > 0 {
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
