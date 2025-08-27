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
	_ datasource.DataSource = &peersDataSource{}
)

// NewPeersDataSource is a helper function to simplify the provider implementation.
func NewPeersDataSource() datasource.DataSource {
	return &peersDataSource{}
}

// peersDataSource is the data source implementation.
type peersDataSource struct {
	client *HAProxyClient
}

// peersDataSourceModel maps the data source schema data.
type peersDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *peersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_peers"
}

// Schema defines the schema for the data source.
func (d *peersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the peers. It must be unique and cannot be changed.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *peersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *peersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state peersDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	peers, err := d.client.ReadPeers(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Peers",
			"Could not read HAProxy Peers, unexpected error: "+err.Error(),
		)
		return
	}

	if peers == nil {
		resp.Diagnostics.AddError(
			"HAProxy Peers Not Found",
			"Could not find HAProxy Peers",
		)
		return
	}

	state.Name = types.StringValue(peers.Name)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
