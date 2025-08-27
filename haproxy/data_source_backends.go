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
	_ datasource.DataSource = &backendsDataSource{}
)

// NewBackendsDataSource is a helper function to simplify the provider implementation.
func NewBackendsDataSource() datasource.DataSource {
	return &backendsDataSource{}
}

// backendsDataSource is the data source implementation.
type backendsDataSource struct {
	client *HAProxyClient
}

// backendsDataSourceModel maps the data source schema data.
type backendsDataSourceModel struct {
	Backends []backendDataSourceModel `tfsdk:"backends"`
}

type backendDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *backendsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backends"
}

// Schema defines the schema for the data source.
func (d *backendsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backends": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the backend.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *backendsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *backendsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state backendsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	backends, err := d.client.ReadBackends(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backends",
			"Could not read backends, unexpected error: "+err.Error(),
		)
		return
	}

	for _, backend := range backends {
		state.Backends = append(state.Backends, backendDataSourceModel{
			Name: types.StringValue(backend.Name),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
