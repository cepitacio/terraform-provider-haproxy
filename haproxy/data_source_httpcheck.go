package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewHttpcheckDataSource() datasource.DataSource {
	return &httpcheckDataSource{}
}

type httpcheckDataSource struct {
	client *HAProxyClient
}

type httpcheckDataSourceModel struct {
	Httpchecks []httpcheckItemModel `tfsdk:"http_checks"`
}

type httpcheckItemModel struct {
	Index   types.String `tfsdk:"index"`
	Type    types.String `tfsdk:"type"`
	Method  types.String `tfsdk:"method"`
	Uri     types.String `tfsdk:"uri"`
	Version types.String `tfsdk:"version"`
}

func (d *httpcheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_httpcheck"
}

func (d *httpcheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http_checks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.StringAttribute{
							Computed:    true,
							Description: "The index of the HTTP check.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the HTTP check.",
						},
						"method": schema.StringAttribute{
							Computed:    true,
							Description: "The HTTP method of the check.",
						},
						"uri": schema.StringAttribute{
							Computed:    true,
							Description: "The URI of the HTTP check.",
						},
						"version": schema.StringAttribute{
							Computed:    true,
							Description: "The HTTP version of the check.",
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
				Description: "The parent name to get HTTP checks for.",
			},
		},
	}
}

func (d *httpcheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *httpcheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state httpcheckDataSourceModel

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

	httpchecks, err := d.client.ReadHttpchecks(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy HTTP Checks",
			"Could not read HAProxy HTTP Checks, unexpected error: "+err.Error(),
		)
		return
	}

	for _, httpcheck := range httpchecks {
		state.Httpchecks = append(state.Httpchecks, httpcheckItemModel{
			Index:   types.StringValue(fmt.Sprintf("%d", httpcheck.Index)),
			Type:    types.StringValue(httpcheck.Type),
			Method:  types.StringValue(httpcheck.Method),
			Uri:     types.StringValue(httpcheck.Uri),
			Version: types.StringValue(httpcheck.Version),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
