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
	_ datasource.DataSource = &frontendsDataSource{}
)

// NewFrontendsDataSource is a helper function to simplify the provider implementation.
func NewFrontendsDataSource() datasource.DataSource {
	return &frontendsDataSource{}
}

// frontendsDataSource is the data source implementation.
type frontendsDataSource struct {
	client *HAProxyClient
}

// frontendsDataSourceModel maps the data source schema data.
type frontendsDataSourceModel struct {
	Frontends []frontendDataSourceModel `tfsdk:"frontends"`
}

type frontendDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *frontendsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontends"
}

// Schema defines the schema for the data source.
func (d *frontendsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"frontends": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the frontend.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *frontendsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *frontendsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state frontendsDataSourceModel
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

	frontends, err := d.client.ReadFrontends(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading frontends",
			"Could not read frontends, unexpected error: "+err.Error(),
		)
		return
	}

	for _, frontend := range frontends {
		state.Frontends = append(state.Frontends, frontendDataSourceModel{
			Name: types.StringValue(frontend.Name),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
