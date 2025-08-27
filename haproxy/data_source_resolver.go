package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &resolverDataSource{}
)

// NewResolverDataSource is a helper function to simplify the provider implementation.
func NewResolverDataSource() datasource.DataSource {
	return &resolverDataSource{}
}

// resolverDataSource is the data source implementation.
type resolverDataSource struct {
	client *HAProxyClient
}

// resolverDataSourceModel maps the data source schema data.
type resolverDataSourceModel struct {
	Name                types.String `tfsdk:"name"`
	AcceptedPayloadSize types.Int64  `tfsdk:"accepted_payload_size"`
	HoldNx              types.Int64  `tfsdk:"hold_nx"`
	HoldObsolete        types.Int64  `tfsdk:"hold_obsolete"`
	HoldOther           types.Int64  `tfsdk:"hold_other"`
	HoldRefused         types.Int64  `tfsdk:"hold_refused"`
	HoldTimeout         types.Int64  `tfsdk:"hold_timeout"`
	HoldValid           types.Int64  `tfsdk:"hold_valid"`
	ResolveRetries      types.Int64  `tfsdk:"resolve_retries"`
	TimeoutResolve      types.Int64  `tfsdk:"timeout_resolve"`
	TimeoutRetry        types.Int64  `tfsdk:"timeout_retry"`
}

// Metadata returns the data source type name.
func (d *resolverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resolver"
}

// Schema defines the schema for the data source.
func (d *resolverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the resolver. It must be unique and cannot be changed.",
			},
			"accepted_payload_size": schema.Int64Attribute{
				Computed:    true,
				Description: "The accepted payload size of the resolver.",
			},
			"hold_nx": schema.Int64Attribute{
				Computed:    true,
				Description: "The hold nx of the resolver.",
			},
			"hold_obsolete": schema.Int64Attribute{
				Computed:    true,
				Description: "The hold obsolete of the resolver.",
			},
			"hold_other": schema.Int64Attribute{
				Computed:    true,
				Description: "The hold other of the resolver.",
			},
			"hold_refused": schema.Int64Attribute{
				Computed:    true,
				Description: "The hold refused of the resolver.",
			},
			"hold_timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "The hold timeout of the resolver.",
			},
			"hold_valid": schema.Int64Attribute{
				Computed:    true,
				Description: "The hold valid of the resolver.",
			},
			"resolve_retries": schema.Int64Attribute{
				Computed:    true,
				Description: "The resolve retries of the resolver.",
			},
			"timeout_resolve": schema.Int64Attribute{
				Computed:    true,
				Description: "The timeout resolve of the resolver.",
			},
			"timeout_retry": schema.Int64Attribute{
				Computed:    true,
				Description: "The timeout retry of the resolver.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *resolverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *resolverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state resolverDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolver, err := d.client.ReadResolver(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Resolver",
			"Could not read HAProxy Resolver, unexpected error: "+err.Error(),
		)
		return
	}

	if resolver == nil {
		resp.Diagnostics.AddError(
			"HAProxy Resolver Not Found",
			"Could not find HAProxy Resolver",
		)
		return
	}

	state.Name = types.StringValue(resolver.Name)
	state.AcceptedPayloadSize = types.Int64Value(resolver.AcceptedPayloadSize)
	state.HoldNx = types.Int64Value(resolver.HoldNx)
	state.HoldObsolete = types.Int64Value(resolver.HoldObsolete)
	state.HoldOther = types.Int64Value(resolver.HoldOther)
	state.HoldRefused = types.Int64Value(resolver.HoldRefused)
	state.HoldTimeout = types.Int64Value(resolver.HoldTimeout)
	state.HoldValid = types.Int64Value(resolver.HoldValid)
	state.ResolveRetries = types.Int64Value(resolver.ResolveRetries)
	state.TimeoutResolve = types.Int64Value(resolver.TimeoutResolve)
	state.TimeoutRetry = types.Int64Value(resolver.TimeoutRetry)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
