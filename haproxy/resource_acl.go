package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// ReadACLs reads ACLs from HAProxy for a given parent
func (r *ACLManager) ReadACLs(ctx context.Context, parentType string, parentName string) ([]ACLPayload, error) {
	return r.client.ReadACLs(ctx, parentType, parentName)
}

// UpdateACLs handles the complex logic of updating ACLs while maintaining order
func (r *ACLManager) UpdateACLs(ctx context.Context, parentType string, parentName string, newAcls []haproxyAclModel) error {
	// Read existing ACLs from HAProxy
	existingAcls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing ACLs: %w", err)
	}

	// Process new ACLs with proper indexing
	sortedNewAcls := r.processAclsBlock(newAcls)

	// Create maps for efficient lookup
	existingAclMap := make(map[string]*ACLPayload)
	for i := range existingAcls {
		existingAclMap[existingAcls[i].AclName] = &existingAcls[i]
	}

	// Track which ACLs we've processed to avoid duplicates
	processedAcls := make(map[string]bool)

	// Track ACLs that need to be recreated due to index changes
	var aclsToRecreate []haproxyAclModel

	// First pass: identify ACLs that need index changes and mark them for recreation
	// Also detect renames by matching content and position swaps
	for _, newAcl := range sortedNewAcls {
		newAclName := newAcl.AclName.ValueString()
		newAclIndex := newAcl.Index.ValueInt64()
		newAclContent := fmt.Sprintf("%s:%s", newAcl.Criterion.ValueString(), newAcl.Value.ValueString())

		// First check if this ACL exists by name
		if existingAcl, exists := existingAclMap[newAclName]; exists {
			// Check if the index has changed
			if existingAcl.Index != newAclIndex {
				// Index has changed - check if this is just a position swap
				// Look for another ACL that might have moved to this ACL's old position
				var isPositionSwap bool
				var swappedAcl *ACLPayload

				for _, otherExistingAcl := range existingAcls {
					if otherExistingAcl.AclName != newAclName && !processedAcls[otherExistingAcl.AclName] {
						otherContent := fmt.Sprintf("%s:%s", otherExistingAcl.Criterion, otherExistingAcl.Value)
						if otherContent == newAclContent {
							// This is a position swap - same content, different positions
							isPositionSwap = true
							swappedAcl = &otherExistingAcl
							break
						}
					}
				}

				if isPositionSwap {
					// This is a position swap, not a real change
					log.Printf("ACL '%s' position swapped with '%s' (same content, different positions) - no update needed",
						newAclName, swappedAcl.AclName)
					// Mark both as processed since they're just swapped
					processedAcls[newAclName] = true
					processedAcls[swappedAcl.AclName] = true
				} else {
					// Real index change, mark for recreation
					log.Printf("ACL '%s' index changed from %d to %d, will recreate",
						newAclName, existingAcl.Index, newAclIndex)
					aclsToRecreate = append(aclsToRecreate, newAcl)
				}
			} else if existingAcl.Criterion != newAcl.Criterion.ValueString() || existingAcl.Value != newAcl.Value.ValueString() {
				// Index is the same but content has changed, update in place
				aclPayload := ACLPayload{
					AclName:   newAcl.AclName.ValueString(),
					Criterion: newAcl.Criterion.ValueString(),
					Value:     newAcl.Value.ValueString(),
					Index:     existingAcl.Index, // Keep the existing index
				}

				log.Printf("Updating existing ACL '%s' at index %d", aclPayload.AclName, aclPayload.Index)
				err := r.client.UpdateAcl(ctx, existingAcl.Index, parentType, parentName, &aclPayload)
				if err != nil {
					return fmt.Errorf("failed to update ACL '%s': %w", aclPayload.AclName, err)
				}
			} else {
				// ACL is identical, no changes needed
				log.Printf("ACL '%s' at index %d is unchanged", newAclName, existingAcl.Index)
			}

			// Mark this ACL as processed
			processedAcls[newAclName] = true
		} else {
			// ACL doesn't exist by name, check if it's a rename by matching content
			var renamedAcl *ACLPayload
			for _, existingAcl := range existingAcls {
				existingContent := fmt.Sprintf("%s:%s", existingAcl.Criterion, existingAcl.Value)
				if existingContent == newAclContent && !processedAcls[existingAcl.AclName] {
					// This is a rename - same content, different name
					renamedAcl = &existingAcl
					break
				}
			}

			if renamedAcl != nil {
				// This is a rename, update the name while keeping the same index and content
				log.Printf("ACL renamed from '%s' to '%s' at index %d", renamedAcl.AclName, newAclName, renamedAcl.Index)
				aclPayload := ACLPayload{
					AclName:   newAclName,
					Criterion: renamedAcl.Criterion,
					Value:     renamedAcl.Value,
					Index:     renamedAcl.Index, // Keep the same index
				}

				err := r.client.UpdateAcl(ctx, renamedAcl.Index, parentType, parentName, &aclPayload)
				if err != nil {
					return fmt.Errorf("failed to rename ACL from '%s' to '%s': %w", renamedAcl.AclName, newAclName, err)
				}

				// Mark both as processed
				processedAcls[renamedAcl.AclName] = true
				processedAcls[newAclName] = true
			} else {
				// This is a completely new ACL, mark for creation
				log.Printf("ACL '%s' is new, will create", newAclName)
			}
		}
	}

	// Second pass: delete all ACLs that need to be recreated (due to index changes)
	// Delete in reverse order (highest index first) to avoid shifting issues
	for _, newAcl := range aclsToRecreate {
		newAclName := newAcl.AclName.ValueString()
		if existingAcl, exists := existingAclMap[newAclName]; exists {
			log.Printf("Deleting ACL '%s' at old index %d for recreation", newAclName, existingAcl.Index)
			err := r.client.DeleteAcl(ctx, existingAcl.Index, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete ACL '%s' at old index %d: %w", newAclName, existingAcl.Index, err)
			}
		}
	}

	// Third pass: create all ACLs that need to be recreated at their new positions
	// Use the user-specified index to maintain order
	for _, newAcl := range aclsToRecreate {
		newAclName := newAcl.AclName.ValueString()
		newAclIndex := newAcl.Index.ValueInt64()

		log.Printf("Creating ACL '%s' at user-specified index %d", newAclName, newAclIndex)
		aclPayload := ACLPayload{
			AclName:   newAcl.AclName.ValueString(),
			Criterion: newAcl.Criterion.ValueString(),
			Value:     newAcl.Value.ValueString(),
			Index:     newAclIndex, // Use the user-specified index
		}

		err = r.client.CreateAcl(ctx, parentType, parentName, &aclPayload)
		if err != nil {
			return fmt.Errorf("failed to create ACL '%s' at index %d: %w", newAclName, newAclIndex)
		}
	}

	// Delete ACLs that are no longer needed (not in the new configuration)
	// Delete in reverse order (highest index first) to avoid shifting issues
	var aclsToDelete []ACLPayload
	for _, existingAcl := range existingAcls {
		if !processedAcls[existingAcl.AclName] {
			aclsToDelete = append(aclsToDelete, existingAcl)
		}
	}

	// Sort by index in descending order (highest first)
	sort.Slice(aclsToDelete, func(i, j int) bool {
		return aclsToDelete[i].Index > aclsToDelete[j].Index
	})

	// Delete ACLs in reverse order
	for _, aclToDelete := range aclsToDelete {
		log.Printf("Deleting ACL '%s' at index %d (no longer needed)", aclToDelete.AclName, aclToDelete.Index)
		err := r.client.DeleteAcl(ctx, aclToDelete.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete ACL '%s': %w", aclToDelete.AclName, err)
		}
	}

	// Create new ACLs that don't exist yet
	// Use the user-specified index to maintain order
	for _, newAcl := range sortedNewAcls {
		newAclName := newAcl.AclName.ValueString()
		if !processedAcls[newAclName] {
			// This is a new ACL, create it with the user-specified index
			newAclIndex := newAcl.Index.ValueInt64()
			aclPayload := ACLPayload{
				AclName:   newAcl.AclName.ValueString(),
				Criterion: newAcl.Criterion.ValueString(),
				Value:     newAcl.Value.ValueString(),
				Index:     newAclIndex, // Use the user-specified index
			}

			log.Printf("Creating new ACL '%s' at index %d", aclPayload.AclName, aclPayload.Index)
			err := r.client.CreateAcl(ctx, parentType, parentName, &aclPayload)
			if err != nil {
				return fmt.Errorf("failed to create ACL '%s': %w", aclPayload.AclName, err)
			}
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
