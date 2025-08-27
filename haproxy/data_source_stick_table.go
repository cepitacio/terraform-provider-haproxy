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
	_ datasource.DataSource = &stickTableDataSource{}
)

// NewStickTableDataSource is a helper function to simplify the provider implementation.
func NewStickTableDataSource() datasource.DataSource {
	return &stickTableDataSource{}
}

// stickTableDataSource is the data source implementation.
type stickTableDataSource struct {
	client *HAProxyClient
}

// stickTableDataSourceModel maps the data source schema data.
type stickTableDataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Size    types.String `tfsdk:"size"`
	Store   types.String `tfsdk:"store"`
	Peers   types.String `tfsdk:"peers"`
	NoPurge types.Bool   `tfsdk:"no_purge"`
}

// Metadata returns the data source type name.
func (d *stickTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stick_table"
}

// Schema defines the schema for the data source.
func (d *stickTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the stick_table. It must be unique and cannot be changed.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the stick_table.",
			},
			"size": schema.StringAttribute{
				Computed:    true,
				Description: "The size of the stick_table.",
			},
			"store": schema.StringAttribute{
				Computed:    true,
				Description: "The store of the stick_table.",
			},
			"peers": schema.StringAttribute{
				Computed:    true,
				Description: "The peers of the stick_table.",
			},
			"no_purge": schema.BoolAttribute{
				Computed:    true,
				Description: "The no_purge of the stick_table.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *stickTableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *stickTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state stickTableDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stickTable, err := d.client.ReadStickTable(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Stick Table",
			"Could not read HAProxy Stick Table, unexpected error: "+err.Error(),
		)
		return
	}

	if stickTable == nil {
		resp.Diagnostics.AddError(
			"HAProxy Stick Table Not Found",
			"Could not find HAProxy Stick Table",
		)
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
