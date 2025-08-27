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
	_ datasource.DataSource = &globalDataSource{}
)

// NewGlobalDataSource is a helper function to simplify the provider implementation.
func NewGlobalDataSource() datasource.DataSource {
	return &globalDataSource{}
}

// globalDataSource is the data source implementation.
type globalDataSource struct {
	client *HAProxyClient
}

// globalDataSourceModel maps the data source schema data.
type globalDataSourceModel struct {
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

// Metadata returns the data source type name.
func (d *globalDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global"
}

// Schema defines the schema for the data source.
func (d *globalDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the global configuration. It must be unique and cannot be changed.",
			},
			"maxconn": schema.Int64Attribute{
				Computed:    true,
				Description: "The max connection of the global configuration.",
			},
			"daemon": schema.StringAttribute{
				Computed:    true,
				Description: "The daemon mode of the global configuration.",
			},
			"stats_timeout": schema.Int64Attribute{
				Computed:    true,
				Description: "The stats timeout of the global configuration.",
			},
			"tune_ssl_default_dh_param": schema.Int64Attribute{
				Computed:    true,
				Description: "The tune ssl default dh param of the global configuration.",
			},
			"ssl_default_bind_ciphers": schema.StringAttribute{
				Computed:    true,
				Description: "The ssl default bind ciphers of the global configuration.",
			},
			"ssl_default_bind_options": schema.StringAttribute{
				Computed:    true,
				Description: "The ssl default bind options of the global configuration.",
			},
			"ssl_default_server_ciphers": schema.StringAttribute{
				Computed:    true,
				Description: "The ssl default server ciphers of the global configuration.",
			},
			"ssl_default_server_options": schema.StringAttribute{
				Computed:    true,
				Description: "The ssl default server options of the global configuration.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *globalDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *globalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state globalDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	global, err := d.client.ReadGlobal(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Global",
			"Could not read HAProxy Global, unexpected error: "+err.Error(),
		)
		return
	}

	if global == nil {
		resp.Diagnostics.AddError(
			"HAProxy Global Not Found",
			"Could not find HAProxy Global",
		)
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
