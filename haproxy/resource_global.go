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
	_ resource.Resource = &globalResource{}
)

// NewGlobalResource is a helper function to simplify the provider implementation.
func NewGlobalResource() resource.Resource {
	return &globalResource{}
}

// globalResource is the resource implementation.
type globalResource struct {
	client *HAProxyClient
}

// globalResourceModel maps the resource schema data.
type globalResourceModel struct {
	Name                    types.String `tfsdk:"name"`
	Maxconn                 types.Int64  `tfsdk:"maxconn"`
	Daemon                  types.String `tfsdk:"daemon"`
	StatsTimeout            types.Int64  `tfsdk:"stats_timeout"`
	TuneSslDefaultDhParam   types.Int64  `tfsdk:"tune_ssl_default_dh_param"`
	SslDefaultBindCiphers   types.String `tfsdk:"ssl_default_bind_ciphers"`
	SslDefaultBindOptions   types.String `tfsdk:"ssl_default_bind_options"`
	SslDefaultServerCiphers types.String `tfsdk:"ssl_default_server_ciphers"`
	SslDefaultServerOptions types.String `tfsdk:"ssl_default_server_options"`
}

// Metadata returns the resource type name.
func (r *globalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global"
}

// Schema defines the schema for the resource.
func (r *globalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the global configuration. It must be unique and cannot be changed.",
			},
			"maxconn": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum number of concurrent connections.",
			},
			"daemon": schema.StringAttribute{
				Optional:    true,
				Description: "The daemon mode of the global configuration.",
			},
			"stats_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The stats timeout of the global configuration.",
			},
			"tune_ssl_default_dh_param": schema.Int64Attribute{
				Optional:    true,
				Description: "The tune ssl default dh param of the global configuration.",
			},
			"ssl_default_bind_ciphers": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl default bind ciphers of the global configuration.",
			},
			"ssl_default_bind_options": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl default bind options of the global configuration.",
			},
			"ssl_default_server_ciphers": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl default server ciphers of the global configuration.",
			},
			"ssl_default_server_options": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl default server options of the global configuration.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *globalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *globalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan globalResourceModel
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

	payload := &GlobalPayload{
		Name:                    plan.Name.ValueString(),
		Maxconn:                 plan.Maxconn.ValueInt64(),
		Daemon:                  plan.Daemon.ValueString(),
		StatsTimeout:            plan.StatsTimeout.ValueInt64(),
		TuneSslDefaultDhParam:   plan.TuneSslDefaultDhParam.ValueInt64(),
		SslDefaultBindCiphers:   plan.SslDefaultBindCiphers.ValueString(),
		SslDefaultBindOptions:   plan.SslDefaultBindOptions.ValueString(),
		SslDefaultServerCiphers: plan.SslDefaultServerCiphers.ValueString(),
		SslDefaultServerOptions: plan.SslDefaultServerOptions.ValueString(),
	}

	err := r.client.CreateGlobal(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating global",
			"Could not create global, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *globalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state globalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	global, err := r.client.ReadGlobal(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading global",
			"Could not read global, unexpected error: "+err.Error(),
		)
		return
	}

	if global == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(global.Name)
	state.Maxconn = types.Int64Value(global.Maxconn)
	state.Daemon = types.StringValue(global.Daemon)
	state.StatsTimeout = types.Int64Value(global.StatsTimeout)
	state.TuneSslDefaultDhParam = types.Int64Value(global.TuneSslDefaultDhParam)
	state.SslDefaultBindCiphers = types.StringValue(global.SslDefaultBindCiphers)
	state.SslDefaultBindOptions = types.StringValue(global.SslDefaultBindOptions)
	state.SslDefaultServerCiphers = types.StringValue(global.SslDefaultServerCiphers)
	state.SslDefaultServerOptions = types.StringValue(global.SslDefaultServerOptions)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *globalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan globalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := &GlobalPayload{
		Name:                    plan.Name.ValueString(),
		Maxconn:                 plan.Maxconn.ValueInt64(),
		Daemon:                  plan.Daemon.ValueString(),
		StatsTimeout:            plan.StatsTimeout.ValueInt64(),
		TuneSslDefaultDhParam:   plan.TuneSslDefaultDhParam.ValueInt64(),
		SslDefaultBindCiphers:   plan.SslDefaultBindCiphers.ValueString(),
		SslDefaultBindOptions:   plan.SslDefaultBindOptions.ValueString(),
		SslDefaultServerCiphers: plan.SslDefaultServerCiphers.ValueString(),
		SslDefaultServerOptions: plan.SslDefaultServerOptions.ValueString(),
	}

	err := r.client.UpdateGlobal(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating global",
			"Could not update global, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *globalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state globalResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlobal(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting global",
			"Could not delete global, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
