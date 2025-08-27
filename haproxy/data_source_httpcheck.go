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
	Httpchecks []httpcheckItemModel `tfsdk:"httpchecks"`
}

type httpcheckItemModel struct {
	Index           types.String `tfsdk:"index"`
	Addr            types.String `tfsdk:"addr"`
	Match           types.String `tfsdk:"match"`
	Pattern         types.String `tfsdk:"pattern"`
	Type            types.String `tfsdk:"type"`
	Method          types.String `tfsdk:"method"`
	Port            types.Int64  `tfsdk:"port"`
	Uri             types.String `tfsdk:"uri"`
	Version         types.String `tfsdk:"version"`
	ExclamationMark types.String `tfsdk:"exclamation_mark"`
	LogLevel        types.String `tfsdk:"log_level"`
	SendProxy       types.String `tfsdk:"send_proxy"`
	ViaSocks4       types.String `tfsdk:"via_socks4"`
	CheckComment    types.String `tfsdk:"check_comment"`
}

func (d *httpcheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_httpcheck"
}

func (d *httpcheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"httpchecks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.StringAttribute{
							Computed:    true,
							Description: "The index of the HTTP check.",
						},
						"addr": schema.StringAttribute{
							Computed:    true,
							Description: "The address of the HTTP check.",
						},
						"match": schema.StringAttribute{
							Computed:    true,
							Description: "The match condition of the HTTP check.",
						},
						"pattern": schema.StringAttribute{
							Computed:    true,
							Description: "The pattern of the HTTP check.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the HTTP check.",
						},
						"method": schema.StringAttribute{
							Computed:    true,
							Description: "The HTTP method of the check.",
						},
						"port": schema.Int64Attribute{
							Computed:    true,
							Description: "The port of the HTTP check.",
						},
						"uri": schema.StringAttribute{
							Computed:    true,
							Description: "The URI of the HTTP check.",
						},
						"version": schema.StringAttribute{
							Computed:    true,
							Description: "The HTTP version of the check.",
						},
						"exclamation_mark": schema.StringAttribute{
							Computed:    true,
							Description: "The exclamation mark flag of the HTTP check.",
						},
						"log_level": schema.StringAttribute{
							Computed:    true,
							Description: "The log level of the HTTP check.",
						},
						"send_proxy": schema.StringAttribute{
							Computed:    true,
							Description: "The send proxy flag of the HTTP check.",
						},
						"via_socks4": schema.StringAttribute{
							Computed:    true,
							Description: "The via socks4 flag of the HTTP check.",
						},
						"check_comment": schema.StringAttribute{
							Computed:    true,
							Description: "The comment of the HTTP check.",
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
			Index:           types.StringValue(fmt.Sprintf("%d", httpcheck.Index)),
			Addr:            types.StringValue(httpcheck.Addr),
			Match:           types.StringValue(httpcheck.Match),
			Pattern:         types.StringValue(httpcheck.Pattern),
			Type:            types.StringValue(httpcheck.Type),
			Method:          types.StringValue(httpcheck.Method),
			Port:            types.Int64Value(httpcheck.Port),
			Uri:             types.StringValue(httpcheck.Uri),
			Version:         types.StringValue(httpcheck.Version),
			ExclamationMark: types.StringValue(httpcheck.ExclamationMark),
			LogLevel:        types.StringValue(httpcheck.LogLevel),
			SendProxy:       types.StringValue(httpcheck.SendProxy),
			ViaSocks4:       types.StringValue(httpcheck.ViaSocks4),
			CheckComment:    types.StringValue(httpcheck.CheckComment),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
