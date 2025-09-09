package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewNameserverDataSource() datasource.DataSource {
	return &nameserverDataSource{}
}

type nameserverDataSource struct {
	client *HAProxyClient
}

type nameserverDataSourceModel struct {
	Nameservers []nameserverItemModel `tfsdk:"nameservers"`
}

type nameserverItemModel struct {
	Name types.String `tfsdk:"name"`
}

func (d *nameserverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nameserver"
}

func (d *nameserverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"nameservers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the nameserver.",
						},
					},
				},
			},
			"resolver": schema.StringAttribute{
				Required:    true,
				Description: "The resolver name to get nameservers for.",
			},
		},
	}
}

func (d *nameserverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *nameserverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nameserverDataSourceModel

	var config struct {
		Resolver types.String `tfsdk:"resolver"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resolverName := config.Resolver.ValueString()

	nameservers, err := d.client.ReadNameservers(ctx, resolverName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Nameservers",
			"Could not read HAProxy Nameservers, unexpected error: "+err.Error(),
		)
		return
	}

	for _, nameserver := range nameservers {
		state.Nameservers = append(state.Nameservers, nameserverItemModel{
			Name: types.StringValue(nameserver.Name),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
