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
	_ resource.Resource = &stickRuleResource{}
)

// NewStickRuleResource is a helper function to simplify the provider implementation.
func NewStickRuleResource() resource.Resource {
	return &stickRuleResource{}
}

// stickRuleResource is the resource implementation.
type stickRuleResource struct {
	client *HAProxyClient
}

// stickRuleResourceModel maps the resource schema data.
type stickRuleResourceModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
	Pattern  types.String `tfsdk:"pattern"`
	Table    types.String `tfsdk:"table"`
	Backend  types.String `tfsdk:"backend"`
}

// Metadata returns the resource type name.
func (r *stickRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stick_rule"
}

// Schema defines the schema for the resource.
func (r *stickRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"index": schema.Int64Attribute{
				Required:    true,
				Description: "The index of the stick rule.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the stick rule.",
			},
			"cond": schema.StringAttribute{
				Optional:    true,
				Description: "The condition of the stick rule.",
			},
			"cond_test": schema.StringAttribute{
				Optional:    true,
				Description: "The condition test of the stick rule.",
			},
			"pattern": schema.StringAttribute{
				Optional:    true,
				Description: "The pattern of the stick rule.",
			},
			"table": schema.StringAttribute{
				Optional:    true,
				Description: "The table of the stick rule.",
			},
			"backend": schema.StringAttribute{
				Required:    true,
				Description: "The backend to which the stick rule belongs.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *stickRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *stickRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan stickRuleResourceModel
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

	payload := &StickRulePayload{
		Index:    plan.Index.ValueInt64(),
		Type:     plan.Type.ValueString(),
		Cond:     plan.Cond.ValueString(),
		CondTest: plan.CondTest.ValueString(),
		Pattern:  plan.Pattern.ValueString(),
		Table:    plan.Table.ValueString(),
	}

	err := r.client.CreateStickRule(ctx, plan.Backend.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating stick_rule",
			"Could not create stick_rule, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *stickRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stickRuleResourceModel
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

	stickRule, err := r.client.ReadStickRule(ctx, state.Index.ValueInt64(), state.Backend.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading stick_rule",
			"Could not read stick_rule, unexpected error: "+err.Error(),
		)
		return
	}

	if stickRule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Index = types.Int64Value(stickRule.Index)
	state.Type = types.StringValue(stickRule.Type)
	state.Cond = types.StringValue(stickRule.Cond)
	state.CondTest = types.StringValue(stickRule.CondTest)
	state.Pattern = types.StringValue(stickRule.Pattern)
	state.Table = types.StringValue(stickRule.Table)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *stickRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stickRuleResourceModel
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

	payload := &StickRulePayload{
		Index:    plan.Index.ValueInt64(),
		Type:     plan.Type.ValueString(),
		Cond:     plan.Cond.ValueString(),
		CondTest: plan.CondTest.ValueString(),
		Pattern:  plan.Pattern.ValueString(),
		Table:    plan.Table.ValueString(),
	}

	err := r.client.UpdateStickRule(ctx, plan.Index.ValueInt64(), plan.Backend.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating stick_rule",
			"Could not update stick_rule, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *stickRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state stickRuleResourceModel
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

	err := r.client.DeleteStickRule(ctx, state.Index.ValueInt64(), state.Backend.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting stick_rule",
			"Could not delete stick_rule, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
