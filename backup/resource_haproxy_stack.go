package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &haproxyStackResource{}
)

// NewHaproxyStackResource is a helper function to simplify the provider implementation.
func NewHaproxyStackResource() resource.Resource {
	return &haproxyStackResource{}
}

// haproxyStackResource is the resource implementation.
type haproxyStackResource struct {
	client          *HAProxyClient
	aclManager      *ACLManager
	frontendManager *FrontendManager
	backendManager  *BackendManager
	serverManager   *ServerManager
}

// haproxyStackResourceModel maps the resource schema data.
type haproxyStackResourceModel struct {
	Name     types.String          `tfsdk:"name"`
	Backend  *HaproxyBackendModel  `tfsdk:"backend"`
	Server   *HaproxyServerModel   `tfsdk:"server"`
	Frontend *HaproxyFrontendModel `tfsdk:"frontend"`
}

// Metadata returns the resource type name.
func (r *haproxyStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Schema defines the schema for the resource.
func (r *haproxyStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a complete HAProxy stack (backend, server, frontend) in a single transaction.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the stack.",
			},
		},
		Blocks: map[string]schema.Block{
			"backend":  GetBackendSchema(),
			"server":   GetServerSchema(),
			"frontend": GetFrontendSchema(),
			"acls":     GetACLSchema(),
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *haproxyStackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developer.", req.ProviderData),
		)
		return
	}

	r.client = client
	r.aclManager = NewACLManager(client)
	r.frontendManager = NewFrontendManager(client)
	r.backendManager = NewBackendManager(client)
	r.serverManager = NewServerManager(client)
}

// Create resource.
func (r *haproxyStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan haproxyStackResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Check if backend is provided
	if plan.Backend == nil {
		resp.Diagnostics.AddError(
			"Missing backend configuration",
			"Backend configuration is required for haproxy_stack resource",
		)
		return
	}

	// Check if server is provided
	if plan.Server == nil {
		resp.Diagnostics.AddError(
			"Missing server configuration",
			"Server configuration is required for haproxy_stack resource",
		)
		return
	}

	// Check if frontend is provided
	if plan.Frontend == nil {
		resp.Diagnostics.AddError(
			"Missing frontend configuration",
			"Frontend configuration is required for haproxy_stack resource",
		)
		return
	}

	// Create all resources in a single transaction
	err := r.createAllResources(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resources",
			"Could not create resources in transaction, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	state := plan
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read resource.
func (r *haproxyStackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state haproxyStackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Read all resources from HAProxy
	err := r.readAllResources(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading resources",
			"Could not read resources, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource.
func (r *haproxyStackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan haproxyStackResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Update all resources in a single transaction
	err := r.updateAllResources(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating resources",
			"Could not update resources in transaction, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	state := plan
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *haproxyStackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state haproxyStackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Delete all resources in a single transaction
	err := r.deleteAllResources(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting resources",
			"Could not delete resources in transaction, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
