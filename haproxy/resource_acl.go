package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetACLSchema returns the schema for the ACL block
func GetACLSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Access Control List (ACL) configuration blocks for content switching and decision making.",
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
	}
}

// haproxyAclModel maps the acl block schema data.
type haproxyAclModel struct {
	AclName   types.String `tfsdk:"acl_name"`
	Index     types.Int64  `tfsdk:"index"`
	Criterion types.String `tfsdk:"criterion"`
	Value     types.String `tfsdk:"value"`
}

// ACLManager handles ACL operations for both frontend and backend
type ACLManager struct {
	client *HAProxyClient
}

// NewACLManager creates a new ACL manager
func NewACLManager(client *HAProxyClient) *ACLManager {
	return &ACLManager{
		client: client,
	}
}

// CreateACLs creates ACLs for a given parent (frontend/backend)
func (r *ACLManager) CreateACLs(ctx context.Context, parentType string, parentName string, acls []haproxyAclModel) error {
	if len(acls) == 0 {
		return nil
	}

	// Process ACLs with proper indexing
	sortedAcls := r.processAclsBlock(acls)

	// Create ACLs in order
	for _, acl := range sortedAcls {
		aclPayload := ACLPayload{
			AclName:   acl.AclName.ValueString(),
			Criterion: acl.Criterion.ValueString(),
			Value:     acl.Value.ValueString(),
			Index:     acl.Index.ValueInt64(),
		}

		log.Printf("Creating ACL '%s' at index %d for %s '%s'", aclPayload.AclName, aclPayload.Index, parentType, parentName)
		err := r.client.CreateAcl(ctx, parentType, parentName, &aclPayload)
		if err != nil {
			return fmt.Errorf("failed to create ACL '%s': %w", aclPayload.AclName, err)
		}
	}

	return nil
}

// CreateACLsInTransaction creates ACLs using an existing transaction ID
func (r *ACLManager) CreateACLsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, acls []haproxyAclModel) error {
	if len(acls) == 0 {
		return nil
	}

	// Process ACLs with proper indexing
	sortedAcls := r.processAclsBlock(acls)

	// Create ACLs in order using the existing transaction
	for _, acl := range sortedAcls {
		aclPayload := ACLPayload{
			AclName:   acl.AclName.ValueString(),
			Criterion: acl.Criterion.ValueString(),
			Value:     acl.Value.ValueString(),
			Index:     acl.Index.ValueInt64(),
		}

		log.Printf("Creating ACL '%s' at index %d for %s '%s' in transaction %s", aclPayload.AclName, aclPayload.Index, parentType, parentName, transactionID)
		err := r.client.CreateACLInTransaction(ctx, transactionID, parentType, parentName, &aclPayload)
		if err != nil {
			return fmt.Errorf("failed to create ACL '%s': %w", aclPayload.AclName, err)
		}
	}

	return nil
}

// ReadACLs reads ACLs for a given parent (frontend/backend)
func (r *ACLManager) ReadACLs(ctx context.Context, parentType string, parentName string) ([]ACLPayload, error) {
	acls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACLs for %s %s: %w", parentType, parentName, err)
	}
	return acls, nil
}

// UpdateACLs updates ACLs for a given parent (frontend/backend)
func (r *ACLManager) UpdateACLs(ctx context.Context, parentType string, parentName string, newAcls []haproxyAclModel) error {
	if len(newAcls) == 0 {
		// Delete all existing ACLs
		return r.DeleteACLs(ctx, parentType, parentName)
	}

	// Read existing ACLs
	existingAcls, err := r.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing ACLs: %w", err)
	}

	// Process updates with proper indexing
	return r.updateAclsWithIndexing(ctx, parentType, parentName, existingAcls, newAcls)
}

// UpdateACLsInTransaction updates ACLs using an existing transaction ID
func (r *ACLManager) UpdateACLsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, acls []haproxyAclModel) error {
	// For now, we'll use the existing UpdateACLs logic but with transaction support
	// This is a simplified version that creates a new transaction for ACL updates
	// In a more sophisticated implementation, we could reuse the existing transaction

	// Delete existing ACLs first
	if err := r.DeleteACLsInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing ACLs: %w", err)
	}

	// Create new ACLs with the transaction
	if err := r.CreateACLsInTransaction(ctx, transactionID, parentType, parentName, acls); err != nil {
		return fmt.Errorf("failed to create new ACLs: %w", err)
	}

	return nil
}

// DeleteACLsInTransaction deletes all ACLs for a given parent using an existing transaction ID
func (r *ACLManager) DeleteACLsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string) error {
	acls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read ACLs for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(acls, func(i, j int) bool {
		return acls[i].Index > acls[j].Index
	})

	for _, acl := range acls {
		log.Printf("Deleting ACL '%s' at index %d in transaction %s", acl.AclName, acl.Index, transactionID)
		err := r.client.DeleteACLInTransaction(ctx, transactionID, parentType, parentName, acl.Index)
		if err != nil {
			return fmt.Errorf("failed to delete ACL '%s': %w", acl.AclName, err)
		}
	}

	return nil
}

// DeleteACLs deletes all ACLs for a given parent
func (r *ACLManager) DeleteACLs(ctx context.Context, parentType string, parentName string) error {
	acls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read ACLs for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(acls, func(i, j int) bool {
		return acls[i].Index > acls[j].Index
	})

	for _, acl := range acls {
		log.Printf("Deleting ACL '%s' at index %d", acl.AclName, acl.Index)
		err := r.client.DeleteAcl(ctx, acl.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete ACL '%s': %w", acl.AclName, err)
		}
	}

	return nil
}

// processAclsBlock processes the ACLs block configuration while preserving the user's intended ACL sequence
func (r *ACLManager) processAclsBlock(acls []haproxyAclModel) []haproxyAclModel {
	if len(acls) == 0 {
		return nil
	}

	// Create a copy to avoid modifying the original
	normalizedAcls := make([]haproxyAclModel, len(acls))
	copy(normalizedAcls, acls)

	// Sort ACLs by user-specified index to determine the intended order
	sort.Slice(normalizedAcls, func(i, j int) bool {
		indexI := normalizedAcls[i].Index.ValueInt64()
		indexJ := normalizedAcls[j].Index.ValueInt64()
		return indexI < indexJ
	})

	// DO NOT normalize indices - preserve user's exact configuration
	// HAProxy can handle non-sequential indices, and normalization causes state drift
	log.Printf("ACL order preserved as configured: %s", r.formatAclOrder(normalizedAcls))

	return normalizedAcls
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *ACLManager) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return "none"
	}

	var order []string
	for _, acl := range acls {
		order = append(order, fmt.Sprintf("%s(index:%d)", acl.AclName.ValueString(), acl.Index.ValueInt64()))
	}
	return strings.Join(order, " → ")
}

// updateAclsWithIndexing handles the complex logic of updating ACLs while maintaining order
func (r *ACLManager) updateAclsWithIndexing(ctx context.Context, parentType string, parentName string, existingAcls []ACLPayload, newAcls []haproxyAclModel) error {
	// Sort new ACLs by index to ensure proper order
	sortedNewAcls := r.processAclsBlock(newAcls)

	// Create a map of existing ACLs by index for quick lookup
	existingAclMap := make(map[int64]*ACLPayload)
	for i := range existingAcls {
		existingAclMap[existingAcls[i].Index] = &existingAcls[i]
	}

	// Process ACLs that need to be recreated (different content or new)
	var aclsToRecreate []haproxyAclModel
	for _, newAcl := range sortedNewAcls {
		newIndex := newAcl.Index.ValueInt64()
		existingAcl, exists := existingAclMap[newIndex]

		if !exists || r.hasAclChanged(existingAcl, &newAcl) {
			aclsToRecreate = append(aclsToRecreate, newAcl)
		}
	}

	// Delete ACLs that are no longer needed
	for _, existingAcl := range existingAcls {
		found := false
		for _, newAcl := range sortedNewAcls {
			if existingAcl.Index == newAcl.Index.ValueInt64() {
				found = true
				break
			}
		}
		if !found {
			log.Printf("Deleting ACL '%s' at index %d (no longer needed)", existingAcl.AclName, existingAcl.Index)
			err := r.client.DeleteAcl(ctx, existingAcl.Index, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete ACL '%s': %w", existingAcl.AclName, err)
			}
		}
	}

	// Recreate ACLs that need updating
	for _, aclToRecreate := range aclsToRecreate {
		// Delete existing ACL if it exists
		if existingAcl, exists := existingAclMap[aclToRecreate.Index.ValueInt64()]; exists {
			log.Printf("Deleting ACL '%s' at index %d for recreation", existingAcl.AclName, existingAcl.Index)
			err := r.client.DeleteAcl(ctx, existingAcl.Index, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete ACL '%s' for recreation: %w", existingAcl.AclName, err)
			}
		}

		// Create new ACL
		aclPayload := ACLPayload{
			AclName:   aclToRecreate.AclName.ValueString(),
			Criterion: aclToRecreate.Criterion.ValueString(),
			Value:     aclToRecreate.Value.ValueString(),
			Index:     aclToRecreate.Index.ValueInt64(),
		}

		log.Printf("Recreating ACL '%s' at index %d", aclPayload.AclName, aclPayload.Index)
		err := r.client.CreateAcl(ctx, parentType, parentName, &aclPayload)
		if err != nil {
			return fmt.Errorf("failed to recreate ACL '%s': %w", aclPayload.AclName, err)
		}
	}

	return nil
}

// hasAclChanged checks if an ACL has changed between existing and new configurations
func (r *ACLManager) hasAclChanged(existing *ACLPayload, new *haproxyAclModel) bool {
	return existing.AclName != new.AclName.ValueString() ||
		existing.Criterion != new.Criterion.ValueString() ||
		existing.Value != new.Value.ValueString()
}
