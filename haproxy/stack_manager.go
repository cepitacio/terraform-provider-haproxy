package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StackManager orchestrates all stack operations and components
type StackManager struct {
	operations *StackOperations
	validation *StackValidation
	processors *StackProcessors
}

// NewStackManager creates a new StackManager instance
func CreateStackManager(client *HAProxyClient, aclManager *ACLManager, frontendManager *FrontendManager, backendManager *BackendManager) *StackManager {
	httpRequestRuleManager := CreateHttpRequestRuleManager(client)
	httpResponseRuleManager := CreateHttpResponseRuleManager(client)
	tcpRequestRuleManager := CreateTcpRequestRuleManager(client)
	tcpResponseRuleManager := CreateTcpResponseRuleManager(client)
	httpcheckManager := CreateHttpcheckManager(client)
	tcpCheckManager := CreateTcpCheckManager(client)
	bindManager := CreateBindManager(client)
	return &StackManager{
		operations: CreateStackOperations(client, aclManager, frontendManager, backendManager, httpRequestRuleManager, httpResponseRuleManager, tcpRequestRuleManager, tcpResponseRuleManager, httpcheckManager, tcpCheckManager, bindManager),
		validation: CreateStackValidation(),
		processors: CreateStackProcessors(),
	}
}

// GetSchema returns the complete schema for the haproxy_stack resource
func (m *StackManager) GetSchema(apiVersion string) schema.Schema {
	return schema.Schema{
		Description: "Manages a complete HAProxy stack including backend, server, frontend, and ACLs.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the HAProxy stack.",
			},
		},
		Blocks: map[string]schema.Block{
			"backend":  GetBackendSchema(),
			"server":   GetServerSchema(),
			"frontend": GetFrontendSchema(),
		},
	}
}

// Create handles the create operation for the haproxy_stack resource
func (m *StackManager) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) error {
	// Process the request data
	data, err := m.processors.ProcessCreateRequest(ctx, req, resp)
	if err != nil {
		resp.Diagnostics.AddError("Error processing create request", err.Error())
		return err
	}

	// Execute the create operation
	if err := m.operations.Create(ctx, req, resp, data); err != nil {
		return err
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	return nil
}

// Read handles the read operation for the haproxy_stack resource
func (m *StackManager) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) error {
	// Process the request data
	data, err := m.processors.ProcessReadRequest(ctx, req, resp)
	if err != nil {
		resp.Diagnostics.AddError("Error processing read request", err.Error())
		return err
	}

	// Execute the read operation
	if err := m.operations.Read(ctx, req, resp, data); err != nil {
		return err
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	return nil
}

// Update handles the update operation for the haproxy_stack resource
func (m *StackManager) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) error {
	// Process the request data
	data, err := m.processors.ProcessUpdateRequest(ctx, req, resp)
	if err != nil {
		resp.Diagnostics.AddError("Error processing update request", err.Error())
		return err
	}

	// Execute the update operation
	if err := m.operations.Update(ctx, req, resp, data); err != nil {
		return err
	}

	// Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
	return nil
}

// Delete handles the delete operation for the haproxy_stack resource
func (m *StackManager) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) error {
	// Process the request data
	data, err := m.processors.ProcessDeleteRequest(ctx, req, resp)
	if err != nil {
		resp.Diagnostics.AddError("Error processing delete request", err.Error())
		return err
	}

	// Execute the delete operation
	if err := m.operations.Delete(ctx, req, resp, data); err != nil {
		return err
	}

	return nil
}

// Configure handles the configuration of the stack manager
func (m *StackManager) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// This method can be used for any additional configuration
	// Currently, all configuration is handled in the constructor
	tflog.Info(ctx, "Stack manager configured successfully")
}

// Metadata returns the resource type metadata
func (m *StackManager) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}
