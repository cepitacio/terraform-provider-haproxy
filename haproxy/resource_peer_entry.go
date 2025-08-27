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
	_ resource.Resource = &peerEntryResource{}
)

// NewPeerEntryResource is a helper function to simplify the provider implementation.
func NewPeerEntryResource() resource.Resource {
	return &peerEntryResource{}
}

// peerEntryResource is the resource implementation.
type peerEntryResource struct {
	client *HAProxyClient
}

// peerEntryResourceModel maps the resource schema data.
type peerEntryResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	Peers   types.String `tfsdk:"peers"`
}

// Metadata returns the resource type name.
func (r *peerEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_peer_entry"
}

// Schema defines the schema for the resource.
func (r *peerEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the peer entry. It must be unique and cannot be changed.",
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "The address of the peer entry.",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "The port of the peer entry.",
			},
			"peers": schema.StringAttribute{
				Required:    true,
				Description: "The peers to which the peer entry belongs.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *peerEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *peerEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan peerEntryResourceModel
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

	payload := &PeerEntryPayload{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
		Port:    plan.Port.ValueInt64(),
	}

	err := r.client.CreatePeerEntry(ctx, plan.Peers.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating peer_entry",
			"Could not create peer_entry, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *peerEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state peerEntryResourceModel
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

	peerEntry, err := r.client.ReadPeerEntry(ctx, state.Name.ValueString(), state.Peers.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading peer_entry",
			"Could not read peer_entry, unexpected error: "+err.Error(),
		)
		return
	}

	if peerEntry == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(peerEntry.Name)
	state.Address = types.StringValue(peerEntry.Address)
	state.Port = types.Int64Value(peerEntry.Port)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *peerEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan peerEntryResourceModel
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

	payload := &PeerEntryPayload{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
		Port:    plan.Port.ValueInt64(),
	}

	err := r.client.UpdatePeerEntry(ctx, plan.Name.ValueString(), plan.Peers.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating peer_entry",
			"Could not update peer_entry, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *peerEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state peerEntryResourceModel
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

	err := r.client.DeletePeerEntry(ctx, state.Name.ValueString(), state.Peers.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting peer_entry",
			"Could not delete peer_entry, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
