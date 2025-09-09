package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewBindDataSource() datasource.DataSource {
	return &bindDataSource{}
}

type bindDataSource struct {
	client *HAProxyClient
}

type bindDataSourceModel struct {
	Binds []bindItemModel `tfsdk:"binds"`
}

type bindItemModel struct {
	Name types.String `tfsdk:"name"`
}

func (d *bindDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind"
}

func (d *bindDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"binds": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the bind.",
						},
					},
				},
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "The parent type (frontend or backend).",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "The parent name to get binds for.",
			},
		},
	}
}

func (d *bindDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bindDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bindDataSourceModel

	var config struct {
		ParentType types.String `tfsdk:"parent_type"`
		ParentName types.String `tfsdk:"parent_name"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := config.ParentType.ValueString()
	parentName := config.ParentName.ValueString()

	binds, err := d.client.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Binds",
			"Could not read HAProxy Binds, unexpected error: "+err.Error(),
		)
		return
	}

	for _, bind := range binds {
		state.Binds = append(state.Binds, bindItemModel{
			Name: types.StringValue(bind.Name),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
