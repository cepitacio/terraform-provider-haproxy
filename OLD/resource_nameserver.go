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
	_ resource.Resource = &nameserverResource{}
)

// NewNameserverResource is a helper function to simplify the provider implementation.
func NewNameserverResource() resource.Resource {
	return &nameserverResource{}
}

// nameserverResource is the resource implementation.
type nameserverResource struct {
	client *HAProxyClient
}

// nameserverResourceModel maps the resource schema data.
type nameserverResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Address  types.String `tfsdk:"address"`
	Port     types.Int64  `tfsdk:"port"`
	Resolver types.String `tfsdk:"resolver"`
}

// Metadata returns the resource type name.
func (r *nameserverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameserver"
}

// Schema defines the schema for the resource.
func (r *nameserverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the nameserver. It must be unique and cannot be changed.",
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "The address of the nameserver.",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "The port of the nameserver.",
			},
			"resolver": schema.StringAttribute{
				Required:    true,
				Description: "The resolver to which the nameserver belongs.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *nameserverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create a new resource.
func (r *nameserverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nameserverResourceModel
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

	payload := &NameserverPayload{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
		Port:    plan.Port.ValueInt64(),
	}

	err := r.client.CreateNameserver(ctx, plan.Resolver.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating nameserver",
			"Could not create nameserver, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *nameserverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nameserverResourceModel
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

	nameserver, err := r.client.ReadNameserver(ctx, state.Name.ValueString(), state.Resolver.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading nameserver",
			"Could not read nameserver, unexpected error: "+err.Error(),
		)
		return
	}

	if nameserver == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(nameserver.Name)
	state.Address = types.StringValue(nameserver.Address)
	state.Port = types.Int64Value(nameserver.Port)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *nameserverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nameserverResourceModel
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

	payload := &NameserverPayload{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
		Port:    plan.Port.ValueInt64(),
	}

	err := r.client.UpdateNameserver(ctx, plan.Name.ValueString(), plan.Resolver.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating nameserver",
			"Could not update nameserver, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *nameserverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nameserverResourceModel
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

	err := r.client.DeleteNameserver(ctx, state.Name.ValueString(), state.Resolver.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting nameserver",
			"Could not delete nameserver, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
