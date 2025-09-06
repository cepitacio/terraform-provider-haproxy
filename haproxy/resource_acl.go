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
					Description: "The index/order of the ACL rule (for backward compatibility).",
				},
			},
		},
	}
}

// haproxyAclModel maps the acl block schema data.
type haproxyAclModel struct {
	AclName   types.String `tfsdk:"acl_name"`
	Index     types.Int64  `tfsdk:"index"` // For backward compatibility with existing state
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

	// Create ACLs in order using array position
	for i, acl := range sortedAcls {
		aclPayload := ACLPayload{
			AclName:   acl.AclName.ValueString(),
			Criterion: acl.Criterion.ValueString(),
			Value:     acl.Value.ValueString(),
			Index:     int64(i), // Use array position instead of index field
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

	// For consistency with HTTP request rules, use the same "create all at once" approach
	// This ensures consistent formatting from HAProxy API
	if r.client.apiVersion == "v3" {
		// Convert all ACLs to payloads
		var allPayloads []ACLPayload
		for i, acl := range sortedAcls {
			aclPayload := ACLPayload{
				AclName:   acl.AclName.ValueString(),
				Criterion: acl.Criterion.ValueString(),
				Value:     acl.Value.ValueString(),
				Index:     int64(i), // Use array position instead of index field
			}
			allPayloads = append(allPayloads, aclPayload)
		}

		// Send all ACLs in one request (same as HTTP request rules)
		if err := r.client.CreateAllACLsInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
			return fmt.Errorf("failed to create all ACLs for %s %s: %w", parentType, parentName, err)
		}

		log.Printf("Created all %d ACLs for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
		return nil
	}

	// Fallback to individual operations for v2
	// Create ACLs in order using the existing transaction and array position
	for i, acl := range sortedAcls {
		aclPayload := ACLPayload{
			AclName:   acl.AclName.ValueString(),
			Criterion: acl.Criterion.ValueString(),
			Value:     acl.Value.ValueString(),
			Index:     int64(i), // Use array position instead of index field
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

// UpdateACLsInTransaction updates ACLs using an existing transaction ID with smart comparison
func (r *ACLManager) UpdateACLsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, acls []haproxyAclModel) error {
	return r.updateAclsWithIndexingInTransaction(ctx, transactionID, parentType, parentName, acls)
}

// updateAclsWithIndexingInTransaction performs smart ACL updates by comparing content rather than just deleting/recreating
func (r *ACLManager) updateAclsWithIndexingInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, desiredAcls []haproxyAclModel) error {
	// For consistency with HTTP request rules, use the same "create all at once" approach
	// This ensures consistent formatting from HAProxy API
	if r.client.apiVersion == "v3" {
		// Process ACLs with proper indexing and deduplication
		sortedAcls := r.processAclsBlock(desiredAcls)

		// Convert all ACLs to payloads
		var allPayloads []ACLPayload
		for i, acl := range sortedAcls {
			aclPayload := ACLPayload{
				AclName:   acl.AclName.ValueString(),
				Criterion: acl.Criterion.ValueString(),
				Value:     acl.Value.ValueString(),
				Index:     int64(i), // Use array position instead of index field
			}
			allPayloads = append(allPayloads, aclPayload)
		}

		// Send all ACLs in one request (same as HTTP request rules)
		if err := r.client.CreateAllACLsInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
			return fmt.Errorf("failed to update all ACLs for %s %s: %w", parentType, parentName, err)
		}

		log.Printf("Updated all %d ACLs for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
		return nil
	}

	// Fallback to individual operations for v2
	// Read existing ACLs from HAProxy
	existingAcls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing ACLs: %w", err)
	}

	// Convert desired ACLs to map for easier comparison
	desiredMap := make(map[string]ACLPayload)
	for i, acl := range desiredAcls {
		aclName := acl.AclName.ValueString()
		log.Printf("DEBUG: Desired ACL: %s (array position: %d)", aclName, i)
		desiredMap[aclName] = ACLPayload{
			AclName:   aclName,
			Criterion: acl.Criterion.ValueString(),
			Value:     acl.Value.ValueString(),
			Index:     int64(i), // Use array position instead of index field
		}
	}

	// Convert existing ACLs to map for easier comparison
	existingMap := make(map[string]ACLPayload)
	for i, acl := range existingAcls {
		// Use array position instead of API index since HAProxy API returns wrong indices
		acl.Index = int64(i)
		log.Printf("DEBUG: Found existing ACL: %s (corrected index: %d)", acl.AclName, acl.Index)
		existingMap[acl.AclName] = acl
	}

	// Find ACLs to delete (exist in HAProxy but not in desired state)
	var aclsToDelete []ACLPayload
	for name, existingAcl := range existingMap {
		if _, exists := desiredMap[name]; !exists {
			aclsToDelete = append(aclsToDelete, existingAcl)
		}
	}

	// Find ACLs to create (exist in desired state but not in HAProxy)
	var aclsToCreate []ACLPayload
	for name, desiredAcl := range desiredMap {
		if _, exists := existingMap[name]; !exists {
			aclsToCreate = append(aclsToCreate, desiredAcl)
		}
	}

	// Find ACLs to update (exist in both but have different content or position)
	var aclsToUpdate []ACLPayload
	for name, desiredAcl := range desiredMap {
		if existingAcl, exists := existingMap[name]; exists {
			if hasAclChanged(existingAcl, desiredAcl) {
				log.Printf("DEBUG: ACL '%s' content changed, will update", name)
				aclsToUpdate = append(aclsToUpdate, desiredAcl)
			} else if existingAcl.Index != desiredAcl.Index {
				log.Printf("DEBUG: ACL '%s' position changed from %d to %d, will reorder", name, existingAcl.Index, desiredAcl.Index)
				aclsToUpdate = append(aclsToUpdate, desiredAcl)
			}
		}
	}

	// Delete ACLs that are no longer needed
	for _, acl := range aclsToDelete {
		log.Printf("Deleting ACL '%s' at index %d in transaction %s", acl.AclName, acl.Index, transactionID)
		if err := r.client.DeleteACLInTransaction(ctx, transactionID, parentType, parentName, acl.Index); err != nil {
			return fmt.Errorf("failed to delete ACL %s: %w", acl.AclName, err)
		}
	}

	// Update ACLs that have changed
	for _, acl := range aclsToUpdate {
		log.Printf("Updating ACL '%s' at index %d in transaction %s", acl.AclName, acl.Index, transactionID)
		if err := r.client.UpdateACLInTransaction(ctx, transactionID, parentType, parentName, acl.Index, &acl); err != nil {
			return fmt.Errorf("failed to update ACL %s: %w", acl.AclName, err)
		}
	}

	// Create new ACLs
	for _, acl := range aclsToCreate {
		log.Printf("Creating ACL '%s' at index %d for %s '%s' in transaction %s", acl.AclName, acl.Index, parentType, parentName, transactionID)
		if err := r.client.CreateACLInTransaction(ctx, transactionID, parentType, parentName, &acl); err != nil {
			return fmt.Errorf("failed to create ACL %s: %w", acl.AclName, err)
		}
	}

	// Note: Reordering is not necessary for simple ACL updates
	// The ACLs will maintain their existing order unless explicitly changed

	return nil
}

// hasAclChanged compares two ACLs to determine if they have different content
func hasAclChanged(existing, desired ACLPayload) bool {
	return existing.Criterion != desired.Criterion || existing.Value != desired.Value
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

// processAclsBlock processes the ACLs block configuration using array position for ordering
func (r *ACLManager) processAclsBlock(acls []haproxyAclModel) []haproxyAclModel {
	if len(acls) == 0 {
		return nil
	}

	// Return ACLs as-is, using array position for ordering
	// The order in the configuration determines the order in HAProxy
	log.Printf("ACL order based on array position: %s", r.formatAclOrder(acls))

	return acls
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *ACLManager) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return "none"
	}

	var order []string
	for _, acl := range acls {
		order = append(order, acl.AclName.ValueString())
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
	for i, newAcl := range sortedNewAcls {
		newIndex := int64(i)
		existingAcl, exists := existingAclMap[newIndex]

		if !exists || r.hasAclChanged(existingAcl, &newAcl) {
			aclsToRecreate = append(aclsToRecreate, newAcl)
		}
	}

	// Delete ACLs that are no longer needed
	for _, existingAcl := range existingAcls {
		found := false
		for i := range sortedNewAcls {
			if existingAcl.Index == int64(i) {
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
	for i, aclToRecreate := range aclsToRecreate {
		// Delete existing ACL if it exists
		if existingAcl, exists := existingAclMap[int64(i)]; exists {
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
			Index:     int64(i),
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
