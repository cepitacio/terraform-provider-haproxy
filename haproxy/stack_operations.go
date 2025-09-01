package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StackOperations handles all CRUD operations for the haproxy_stack resource
type StackOperations struct {
	client                 *HAProxyClient
	aclManager             *ACLManager
	frontendManager        *FrontendManager
	backendManager         *BackendManager
	httpRequestRuleManager *HttpRequestRuleManager
}

// NewStackOperations creates a new StackOperations instance
func NewStackOperations(client *HAProxyClient, aclManager *ACLManager, frontendManager *FrontendManager, backendManager *BackendManager, httpRequestRuleManager *HttpRequestRuleManager) *StackOperations {
	return &StackOperations{
		client:                 client,
		aclManager:             aclManager,
		backendManager:         backendManager,
		frontendManager:        frontendManager,
		httpRequestRuleManager: httpRequestRuleManager,
	}
}

// Create performs the create operation for the haproxy_stack resource
func (o *StackOperations) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Creating HAProxy stack")

	// Begin a single transaction for all resources
	tflog.Info(ctx, "Beginning single transaction for all resources")
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return err
	}
	tflog.Info(ctx, "Transaction created", map[string]interface{}{"transaction_id": transactionID})
	defer func() {
		// Rollback transaction if we encounter an error
		if err != nil {
			tflog.Error(ctx, "Rolling back transaction due to error", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
			if rollbackErr := o.client.RollbackTransaction(transactionID); rollbackErr != nil {
				tflog.Error(ctx, "Failed to rollback transaction", map[string]interface{}{"error": rollbackErr.Error()})
			}
		}
	}()

	// Create backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Creating backend in transaction", map[string]interface{}{"transaction_id": transactionID})
		if err := o.backendManager.CreateBackendInTransaction(ctx, transactionID, data.Backend); err != nil {
			resp.Diagnostics.AddError("Error creating backend", err.Error())
			return err
		}
		tflog.Info(ctx, "Backend created successfully in transaction", map[string]interface{}{"transaction_id": transactionID})
	}

	// Create server if specified
	if data.Server != nil && data.Backend != nil {
		// Server needs parent backend
		if err := o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), &ServerPayload{
			Name:    data.Server.Name.ValueString(),
			Address: data.Server.Address.ValueString(),
			Port:    int64(data.Server.Port.ValueInt64()),
		}); err != nil {
			resp.Diagnostics.AddError("Error creating server", err.Error())
			return err
		}
	}

	// Create frontend if specified
	if data.Frontend != nil {
		if err := o.frontendManager.CreateFrontendInTransaction(ctx, transactionID, data.Frontend); err != nil {
			resp.Diagnostics.AddError("Error creating frontend", err.Error())
			return err
		}
	}

	// Note: ACLs are now handled within frontend/backend creation, not here

	// Create HTTP Request Rules AFTER ACLs (so they can reference existing ACLs)
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		if err := o.httpRequestRuleManager.CreateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
			resp.Diagnostics.AddError("Error creating HTTP request rules", err.Error())
			return err
		}
	}

	// Commit the transaction
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err := o.client.CommitTransaction(transactionID); err != nil {
		tflog.Error(ctx, "Failed to commit transaction", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
		return err
	}
	tflog.Info(ctx, "Transaction committed successfully", map[string]interface{}{"transaction_id": transactionID})

	tflog.Info(ctx, "HAProxy stack created successfully")
	return nil
}

// Read performs the read operation for the haproxy_stack resource
func (o *StackOperations) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Reading HAProxy stack")

	// Read backend if specified
	if data.Backend != nil {
		_, err := o.backendManager.ReadBackend(ctx, data.Backend.Name.ValueString(), data.Backend)
		if err != nil {
			resp.Diagnostics.AddError("Error reading backend", err.Error())
			return err
		}
	}

	// Read server if specified
	if data.Server != nil && data.Backend != nil {
		// Server reading would need to be implemented
		tflog.Info(ctx, "Server reading not yet implemented")
	}

	// Read frontend if specified
	if data.Frontend != nil {
		_, err := o.frontendManager.ReadFrontend(ctx, data.Frontend.Name.ValueString(), data.Frontend)
		if err != nil {
			resp.Diagnostics.AddError("Error reading frontend", err.Error())
			return err
		}
	}

	// Read ACLs if specified
	if len(data.Acls) > 0 {
		// ACLs reading would need to be implemented
		tflog.Info(ctx, "ACLs reading not yet implemented")
	}

	tflog.Info(ctx, "HAProxy stack read successfully")
	return nil
}

// Update performs the update operation for the haproxy_stack resource
func (o *StackOperations) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Updating HAProxy stack")

	// Begin transaction for all updates
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return err
	}

	// Use defer to ensure rollback on error
	defer func() {
		if err != nil {
			tflog.Info(ctx, "Rolling back transaction due to error", map[string]interface{}{"transaction_id": transactionID})
			if rollbackErr := o.client.RollbackTransaction(transactionID); rollbackErr != nil {
				tflog.Error(ctx, "Error rolling back transaction", map[string]interface{}{"error": rollbackErr.Error()})
			}
		}
	}()

	// Update backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Updating backend", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.backendManager.UpdateBackendInTransaction(ctx, transactionID, data.Backend); err != nil {
			resp.Diagnostics.AddError("Error updating backend", err.Error())
			return err
		}
	}

	// Update server if specified
	if data.Server != nil && data.Backend != nil {
		tflog.Info(ctx, "Updating server", map[string]interface{}{"server_name": data.Server.Name.ValueString()})
		if err = o.client.UpdateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), &ServerPayload{
			Name:    data.Server.Name.ValueString(),
			Address: data.Server.Address.ValueString(),
			Port:    data.Server.Port.ValueInt64(),
			Check:   data.Server.Check.ValueString(),
			Maxconn: data.Server.Maxconn.ValueInt64(),
			Weight:  data.Server.Weight.ValueInt64(),
		}); err != nil {
			resp.Diagnostics.AddError("Error updating server", err.Error())
			return err
		}
	}

	// Update frontend if specified
	if data.Frontend != nil {
		tflog.Info(ctx, "Updating frontend", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.frontendManager.UpdateFrontendInTransaction(ctx, transactionID, data.Frontend); err != nil {
			resp.Diagnostics.AddError("Error updating frontend", err.Error())
			return err
		}
	}

	// Update HTTP Request Rules if specified
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		tflog.Info(ctx, "Updating HTTP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.UpdateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
			resp.Diagnostics.AddError("Error updating HTTP request rules", err.Error())
			return err
		}
	}

	// Commit all updates
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err = o.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
		return err
	}

	// Clear the error so defer doesn't rollback
	err = nil
	tflog.Info(ctx, "HAProxy stack updated successfully")
	return nil
}

// Delete performs the delete operation for the haproxy_stack resource
func (o *StackOperations) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Deleting HAProxy stack")

	// Begin transaction for all deletes
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return err
	}

	// Use defer to ensure rollback on error
	defer func() {
		if err != nil {
			tflog.Info(ctx, "Rolling back transaction due to error", map[string]interface{}{"transaction_id": transactionID})
			if rollbackErr := o.client.RollbackTransaction(transactionID); rollbackErr != nil {
				tflog.Error(ctx, "Error rolling back transaction", map[string]interface{}{"error": rollbackErr.Error()})
			}
		}
	}()

	// Delete HTTP Request Rules if specified
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		tflog.Info(ctx, "Deleting HTTP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.DeleteHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting HTTP request rules", err.Error())
			return err
		}
	}

	// Delete frontend if specified
	if data.Frontend != nil {
		tflog.Info(ctx, "Deleting frontend", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.frontendManager.DeleteFrontendInTransaction(ctx, transactionID, data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting frontend", err.Error())
			return err
		}
	}

	// Delete server if specified
	if data.Server != nil && data.Backend != nil {
		tflog.Info(ctx, "Deleting server", map[string]interface{}{"server_name": data.Server.Name.ValueString()})
		if err = o.client.DeleteServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Server.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting server", err.Error())
			return err
		}
	}

	// Delete backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Deleting backend", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.backendManager.DeleteBackendInTransaction(ctx, transactionID, data.Backend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting backend", err.Error())
			return err
		}
	}

	// Commit all deletes
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err = o.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
		return err
	}

	// Clear the error so defer doesn't rollback
	err = nil
	tflog.Info(ctx, "HAProxy stack deleted successfully")
	return nil
}
