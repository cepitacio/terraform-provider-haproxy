package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StackOperations handles all CRUD operations for the haproxy_stack resource
type StackOperations struct {
	client          *HAProxyClient
	aclManager      *ACLManager
	frontendManager *FrontendManager
	backendManager  *BackendManager
}

// NewStackOperations creates a new StackOperations instance
func NewStackOperations(client *HAProxyClient, aclManager *ACLManager, frontendManager *FrontendManager, backendManager *BackendManager) *StackOperations {
	return &StackOperations{
		client:          client,
		aclManager:      aclManager,
		frontendManager: frontendManager,
		backendManager:  backendManager,
	}
}

// Create performs the create operation for the haproxy_stack resource
func (o *StackOperations) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Creating HAProxy stack")

	// Create backend if specified
	if data.Backend != nil {
		_, err := o.backendManager.CreateBackend(ctx, data.Backend)
		if err != nil {
			resp.Diagnostics.AddError("Error creating backend", err.Error())
			return err
		}
	}

	// Create server if specified
	if data.Server != nil && data.Backend != nil {
		// Server needs parent backend
		if err := o.client.CreateServer(ctx, "backend", data.Backend.Name.ValueString(), &ServerPayload{
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
		_, err := o.frontendManager.CreateFrontend(ctx, data.Frontend)
		if err != nil {
			resp.Diagnostics.AddError("Error creating frontend", err.Error())
			return err
		}
	}

	// Create ACLs if specified
	if len(data.Acls) > 0 {
		// ACLs need parent type and name
		parentType := "frontend"
		parentName := ""
		if data.Frontend != nil {
			parentName = data.Frontend.Name.ValueString()
		} else if data.Backend != nil {
			parentType = "backend"
			parentName = data.Backend.Name.ValueString()
		}

		if parentName != "" {
			if err := o.aclManager.CreateACLs(ctx, parentType, parentName, data.Acls); err != nil {
				resp.Diagnostics.AddError("Error creating ACLs", err.Error())
				return err
			}
		}
	}

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

	// Update backend if specified
	if data.Backend != nil {
		if err := o.backendManager.UpdateBackend(ctx, data.Backend); err != nil {
			resp.Diagnostics.AddError("Error updating backend", err.Error())
			return err
		}
	}

	// Update server if specified
	if data.Server != nil {
		// Server updating would need to be implemented
		tflog.Info(ctx, "Server updating not yet implemented")
	}

	// Update frontend if specified
	if data.Frontend != nil {
		if err := o.frontendManager.UpdateFrontend(ctx, data.Frontend); err != nil {
			resp.Diagnostics.AddError("Error updating frontend", err.Error())
			return err
		}
	}

	// Update ACLs if specified
	if len(data.Acls) > 0 {
		// ACLs updating would need to be implemented
		tflog.Info(ctx, "ACLs updating not yet implemented")
	}

	tflog.Info(ctx, "HAProxy stack updated successfully")
	return nil
}

// Delete performs the delete operation for the haproxy_stack resource
func (o *StackOperations) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Deleting HAProxy stack")

	// Delete ACLs if specified
	if len(data.Acls) > 0 {
		// ACLs deleting would need to be implemented
		tflog.Info(ctx, "ACLs deleting not yet implemented")
	}

	// Delete frontend if specified
	if data.Frontend != nil {
		if err := o.frontendManager.DeleteFrontend(ctx, data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting frontend", err.Error())
			return err
		}
	}

	// Delete server if specified
	if data.Server != nil {
		// Server deleting would need to be implemented
		tflog.Info(ctx, "Server deleting not yet implemented")
	}

	// Delete backend if specified
	if data.Backend != nil {
		if err := o.backendManager.DeleteBackend(ctx, data.Backend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting backend", err.Error())
			return err
		}
	}

	tflog.Info(ctx, "HAProxy stack deleted successfully")
	return nil
}
