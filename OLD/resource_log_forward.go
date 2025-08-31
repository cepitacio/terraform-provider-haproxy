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
	_ resource.Resource = &logForwardResource{}
)

// NewLogForwardResource is a helper function to simplify the provider implementation.
func NewLogForwardResource() resource.Resource {
	return &logForwardResource{}
}

// logForwardResource is the resource implementation.
type logForwardResource struct {
	client *HAProxyClient
}

// logForwardResourceModel maps the resource schema data.
type logForwardResourceModel struct {
	Name     types.String `tfsdk:"name"`
	Backlog  types.Int64  `tfsdk:"backlog"`
	Maxconn  types.Int64  `tfsdk:"maxconn"`
	Timeout  types.Int64  `tfsdk:"timeout"`
	Loglevel types.String `tfsdk:"loglevel"`
}

// Metadata returns the resource type name.
func (r *logForwardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_forward"
}

// Schema defines the schema for the resource.
func (r *logForwardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the log forward. It must be unique and cannot be changed.",
			},
			"backlog": schema.Int64Attribute{
				Optional:    true,
				Description: "The backlog of the log forward.",
			},
			"maxconn": schema.Int64Attribute{
				Optional:    true,
				Description: "The max connections of the log forward.",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The timeout of the log forward.",
			},
			"loglevel": schema.StringAttribute{
				Optional:    true,
				Description: "The log level of the log forward.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *logForwardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *logForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan logForwardResourceModel
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

	payload := &LogForwardPayload{
		Name:     plan.Name.ValueString(),
		Backlog:  plan.Backlog.ValueInt64(),
		Maxconn:  plan.Maxconn.ValueInt64(),
		Timeout:  plan.Timeout.ValueInt64(),
		Loglevel: plan.Loglevel.ValueString(),
	}

	err := r.client.CreateLogForward(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating log forward",
			"Could not create log forward, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *logForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state logForwardResourceModel
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

	logForward, err := r.client.ReadLogForward(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading log forward",
			"Could not read log forward, unexpected error: "+err.Error(),
		)
		return
	}

	if logForward == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(logForward.Name)
	state.Backlog = types.Int64Value(logForward.Backlog)
	state.Maxconn = types.Int64Value(logForward.Maxconn)
	state.Timeout = types.Int64Value(logForward.Timeout)
	state.Loglevel = types.StringValue(logForward.Loglevel)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *logForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan logForwardResourceModel
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

	payload := &LogForwardPayload{
		Name:     plan.Name.ValueString(),
		Backlog:  plan.Backlog.ValueInt64(),
		Maxconn:  plan.Maxconn.ValueInt64(),
		Timeout:  plan.Timeout.ValueInt64(),
		Loglevel: plan.Loglevel.ValueString(),
	}

	err := r.client.UpdateLogForward(ctx, plan.Name.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating log forward",
			"Could not update log forward, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *logForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state logForwardResourceModel
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

	err := r.client.DeleteLogForward(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting log forward",
			"Could not delete log forward, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
