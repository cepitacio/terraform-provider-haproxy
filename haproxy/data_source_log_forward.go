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
	_ datasource.DataSource = &logForwardDataSource{}
)

// NewLogForwardDataSource is a helper function to simplify the provider implementation.
func NewLogForwardDataSource() datasource.DataSource {
	return &logForwardDataSource{}
}

// logForwardDataSource is the data source implementation.
type logForwardDataSource struct {
	client *HAProxyClient
}

// logForwardDataSourceModel maps the data source schema data.
type logForwardDataSourceModel struct {
	Name     types.String `tfsdk:"name"`
	Backlog  types.Int64  `tfsdk:"backlog"`
	Maxconn  types.Int64  `tfsdk:"maxconn"`
	Timeout  types.Int64  `tfsdk:"timeout"`
	Loglevel types.String `tfsdk:"loglevel"`
}

// Metadata returns the data source type name.
func (d *logForwardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_forward"
}

// Schema defines the schema for the data source.
func (d *logForwardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the log_forward. It must be unique and cannot be changed.",
			},
			"backlog": schema.Int64Attribute{
				Computed:    true,
				Description: "The backlog of the log_forward.",
			},
			"maxconn": schema.Int64Attribute{
				Computed:    true,
				Description: "The maxconn of the log_forward.",
			},
			"timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "The timeout of the log_forward.",
			},
			"loglevel": schema.StringAttribute{
				Computed:    true,
				Description: "The loglevel of the log_forward.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *logForwardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *logForwardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state logForwardDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	logForward, err := d.client.ReadLogForward(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Log Forward",
			"Could not read HAProxy Log Forward, unexpected error: "+err.Error(),
		)
		return
	}

	if logForward == nil {
		resp.Diagnostics.AddError(
			"HAProxy Log Forward Not Found",
			"Could not find HAProxy Log Forward",
		)
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
