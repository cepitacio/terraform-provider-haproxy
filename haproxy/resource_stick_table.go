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
	_ resource.Resource = &stickTableResource{}
)

// NewStickTableResource is a helper function to simplify the provider implementation.
func NewStickTableResource() resource.Resource {
	return &stickTableResource{}
}

// stickTableResource is the resource implementation.
type stickTableResource struct {
	client *HAProxyClient
}

// stickTableResourceModel maps the resource schema data.
type stickTableResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Size    types.String `tfsdk:"size"`
	Store   types.String `tfsdk:"store"`
	Peers   types.String `tfsdk:"peers"`
	NoPurge types.Bool   `tfsdk:"no_purge"`
}

// Metadata returns the resource type name.
func (r *stickTableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stick_table"
}

// Schema defines the schema for the resource.
func (r *stickTableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the stick table. It must be unique and cannot be changed.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "The type of the stick table.",
			},
			"size": schema.StringAttribute{
				Optional:    true,
				Description: "The size of the stick table.",
			},
			"store": schema.StringAttribute{
				Optional:    true,
				Description: "The store of the stick table.",
			},
			"peers": schema.StringAttribute{
				Optional:    true,
				Description: "The peers of the stick table.",
			},
			"no_purge": schema.BoolAttribute{
				Optional:    true,
				Description: "The no_purge of the stick table.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *stickTableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *stickTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stickTableResourceModel
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

	payload := &StickTablePayload{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Size:    plan.Size.ValueString(),
		Store:   plan.Store.ValueString(),
		Peers:   plan.Peers.ValueString(),
		NoPurge: plan.NoPurge.ValueBool(),
	}

	err := r.client.CreateStickTable(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating stick_table",
			"Could not create stick_table, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *stickTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stickTableResourceModel
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

	stickTable, err := r.client.ReadStickTable(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading stick_table",
			"Could not read stick_table, unexpected error: "+err.Error(),
		)
		return
	}

	if stickTable == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(stickTable.Name)
	state.Type = types.StringValue(stickTable.Type)
	state.Size = types.StringValue(stickTable.Size)
	state.Store = types.StringValue(stickTable.Store)
	state.Peers = types.StringValue(stickTable.Peers)
	state.NoPurge = types.BoolValue(stickTable.NoPurge)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *stickTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stickTableResourceModel
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

	payload := &StickTablePayload{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Size:    plan.Size.ValueString(),
		Store:   plan.Store.ValueString(),
		Peers:   plan.Peers.ValueString(),
		NoPurge: plan.NoPurge.ValueBool(),
	}

	err := r.client.UpdateStickTable(ctx, plan.Name.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating stick_table",
			"Could not update stick_table, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *stickTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stickTableResourceModel
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

	err := r.client.DeleteStickTable(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting stick_table",
			"Could not delete stick_table, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
