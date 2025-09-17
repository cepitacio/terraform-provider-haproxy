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

const (
	aclCriterionNoneValue = "none"
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
func CreateACLManager(client *HAProxyClient) *ACLManager {
	return &ACLManager{
		client: client,
	}
}

// CreateACLs creates ACLs for a given parent (frontend/backend)
func (r *ACLManager) CreateACLs(ctx context.Context, parentType, parentName string, acls []haproxyAclModel) error {
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
func (r *ACLManager) CreateACLsInTransaction(ctx context.Context, transactionID, parentType, parentName string, acls []haproxyAclModel) error {
	if len(acls) == 0 {
		return nil
	}

	// Process ACLs with proper indexing
	sortedAcls := r.processAclsBlock(acls)

	// Use the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
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

	// Send all ACLs in one request (same for both v2 and v3)
	if err := r.client.CreateAllACLsInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		return fmt.Errorf("failed to create all ACLs for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Created all %d ACLs for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)

	return nil
}

// ReadACLs reads ACLs for a given parent (frontend/backend)
func (r *ACLManager) ReadACLs(ctx context.Context, parentType, parentName string) ([]ACLPayload, error) {
	acls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACLs for %s %s: %w", parentType, parentName, err)
	}
	return acls, nil
}

// UpdateACLs updates ACLs for a given parent (frontend/backend)
func (r *ACLManager) UpdateACLs(ctx context.Context, parentType, parentName string, newAcls []haproxyAclModel) error {
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
func (r *ACLManager) UpdateACLsInTransaction(ctx context.Context, transactionID, parentType, parentName string, acls []haproxyAclModel) error {
	return r.updateAclsWithIndexingInTransaction(ctx, transactionID, parentType, parentName, acls)
}

// updateAclsWithIndexingInTransaction performs smart ACL updates by comparing existing vs desired
func (r *ACLManager) updateAclsWithIndexingInTransaction(ctx context.Context, transactionID, parentType, parentName string, desiredAcls []haproxyAclModel) error {
	// Read existing ACLs to compare with desired ones
	existingAcls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing ACLs for %s %s: %w", parentType, parentName, err)
	}

	// Use the smart comparison logic to only update what changed
	return r.updateAclsWithIndexingInTransactionSmart(ctx, transactionID, parentType, parentName, existingAcls, desiredAcls)
}

// updateAclsWithIndexingInTransactionSmart performs smart ACL updates by comparing existing vs desired
func (r *ACLManager) updateAclsWithIndexingInTransactionSmart(ctx context.Context, transactionID, parentType, parentName string, existingAcls []ACLPayload, desiredAcls []haproxyAclModel) error {
	// Process desired ACLs with proper indexing and deduplication
	sortedDesiredAcls := r.processAclsBlock(desiredAcls)

	// Create a map of existing ACLs by name for quick lookup (since HAProxy API v3 returns all with index 0)
	existingAclMap := make(map[string]*ACLPayload)
	for i := range existingAcls {
		existingAclMap[existingAcls[i].AclName] = &existingAcls[i]
	}

	// Check if any ACLs actually changed (content OR order)
	hasChanges := false
	var aclsToRecreate []haproxyAclModel

	log.Printf("DEBUG: Comparing %d existing ACLs with %d desired ACLs for %s %s", len(existingAcls), len(sortedDesiredAcls), parentType, parentName)

	// First, check if the number of ACLs changed
	if len(existingAcls) != len(sortedDesiredAcls) {
		log.Printf("DEBUG: ACL count changed from %d to %d, marking for recreation", len(existingAcls), len(sortedDesiredAcls))
		hasChanges = true
		aclsToRecreate = sortedDesiredAcls
	} else {
		// Check if ACLs have changed content OR order
		for i, desiredAcl := range sortedDesiredAcls {
			desiredName := desiredAcl.AclName.ValueString()
			existingAcl, exists := existingAclMap[desiredName]

			log.Printf("DEBUG: ACL %d - desired: name='%s', criterion='%s', value='%s'", i, desiredName, desiredAcl.Criterion.ValueString(), desiredAcl.Value.ValueString())

			if !exists {
				log.Printf("DEBUG: ACL %d - no existing ACL with name '%s', marking for recreation", i, desiredName)
				hasChanges = true
				aclsToRecreate = sortedDesiredAcls
				break
			} else {
				log.Printf("DEBUG: ACL %d - existing: name='%s', criterion='%s', value='%s'", i, existingAcl.AclName, existingAcl.Criterion, existingAcl.Value)
				changed := r.hasAclChanged(existingAcl, &desiredAcl)
				log.Printf("DEBUG: ACL %d - hasAclChanged returned: %t", i, changed)
				if changed {
					log.Printf("DEBUG: ACL %d - marked for recreation due to content changes", i)
					hasChanges = true
					aclsToRecreate = sortedDesiredAcls
					break
				}
			}
		}

		// If no content changes detected, check for order changes by comparing the sequence
		if !hasChanges {
			log.Printf("DEBUG: No content changes detected, checking for order changes...")
			for i, desiredAcl := range sortedDesiredAcls {
				desiredName := desiredAcl.AclName.ValueString()
				// Check if the ACL at position i has the same name as the existing ACL at position i
				if i < len(existingAcls) {
					existingName := existingAcls[i].AclName
					if desiredName != existingName {
						log.Printf("DEBUG: Order change detected at position %d - desired: '%s', existing: '%s'", i, desiredName, existingName)
						hasChanges = true
						aclsToRecreate = sortedDesiredAcls
						break
					}
				}
			}
		}
	}

	// Also check if any existing ACLs need to be removed (not in desired list)
	if !hasChanges {
		for existingName := range existingAclMap {
			found := false
			for _, desiredAcl := range sortedDesiredAcls {
				if desiredAcl.AclName.ValueString() == existingName {
					found = true
					break
				}
			}
			if !found {
				log.Printf("DEBUG: Existing ACL '%s' not in desired list, marking for removal", existingName)
				hasChanges = true
				aclsToRecreate = sortedDesiredAcls
				break
			}
		}
	}

	log.Printf("DEBUG: Final hasChanges result: %t, ACLs to recreate: %d", hasChanges, len(aclsToRecreate))

	// If no changes detected, skip the update
	if !hasChanges {
		log.Printf("No ACL changes detected for %s %s, skipping update", parentType, parentName)
		return nil
	}

	// First, delete all existing ACLs to avoid duplicates
	if err := r.DeleteACLsInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing ACLs for %s %s: %w", parentType, parentName, err)
	}

	// Then create all desired ACLs using the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
	// Process ACLs with proper indexing and deduplication
	sortedRules := r.processAclsBlock(desiredAcls)

	// Convert all ACLs to payloads
	var allPayloads []ACLPayload
	for i, acl := range sortedRules {
		aclPayload := ACLPayload{
			AclName:   acl.AclName.ValueString(),
			Criterion: acl.Criterion.ValueString(),
			Value:     acl.Value.ValueString(),
			Index:     int64(i), // Use array position instead of index field
		}
		allPayloads = append(allPayloads, aclPayload)
	}

	// Send all ACLs in one request (same for both v2 and v3)
	if err := r.client.CreateAllACLsInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		return fmt.Errorf("failed to create new ACLs for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Updated %d ACLs for %s %s in transaction %s (delete-then-create)", len(allPayloads), parentType, parentName, transactionID)
	return nil
}

// DeleteACLsInTransaction deletes all ACLs for a given parent using an existing transaction ID
func (r *ACLManager) DeleteACLsInTransaction(ctx context.Context, transactionID, parentType, parentName string) error {
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
func (r *ACLManager) DeleteACLs(ctx context.Context, parentType, parentName string) error {
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

// processAclsBlock processes the ACLs block configuration using array position for ordering and deduplication
func (r *ACLManager) processAclsBlock(acls []haproxyAclModel) []haproxyAclModel {
	if len(acls) == 0 {
		return acls
	}

	// Deduplicate ACLs by creating a map with unique keys based on ACL name
	// This ensures consistent behavior like HTTP request rules
	aclMap := make(map[string]haproxyAclModel)

	for i, acl := range acls {
		// Use ACL name as the unique key (like HTTP request rules use content-based keys)
		key := acl.AclName.ValueString()

		// If an ACL with the same name already exists, keep the last one
		aclMap[key] = acl
		log.Printf("ACL %d: key='%s', criterion='%s', value='%s'", i, key, acl.Criterion.ValueString(), acl.Value.ValueString())
	}

	// Convert map back to slice, maintaining the original order for the first occurrence of each unique ACL
	var deduplicatedAcls []haproxyAclModel
	seenKeys := make(map[string]bool)

	for _, acl := range acls {
		key := acl.AclName.ValueString()
		if !seenKeys[key] {
			deduplicatedAcls = append(deduplicatedAcls, acl)
			seenKeys[key] = true
		}
	}

	log.Printf("Deduplicated ACLs: %d original -> %d unique", len(acls), len(deduplicatedAcls))
	log.Printf("ACL order based on array position: %s", r.formatAclOrder(deduplicatedAcls))

	return deduplicatedAcls
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *ACLManager) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return aclCriterionNoneValue
	}

	var order []string
	for _, acl := range acls {
		order = append(order, acl.AclName.ValueString())
	}
	return strings.Join(order, " → ")
}

// updateAclsWithIndexing handles the complex logic of updating ACLs while maintaining order
func (r *ACLManager) updateAclsWithIndexing(ctx context.Context, parentType, parentName string, existingAcls []ACLPayload, newAcls []haproxyAclModel) error {
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
