package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTcpRequestRuleDataSource() datasource.DataSource {
	return &tcpRequestRuleDataSource{}
}

type tcpRequestRuleDataSource struct {
	client *HAProxyClient
}

type tcpRequestRuleDataSourceModel struct {
	TcpRequestRules []tcpRequestRuleItemModel `tfsdk:"tcp_request_rules"`
}

type tcpRequestRuleItemModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Action       types.String `tfsdk:"action"`
	Cond         types.String `tfsdk:"cond"`
	CondTest     types.String `tfsdk:"cond_test"`
	HdrName      types.String `tfsdk:"hdr_name"`
	HdrFormat    types.String `tfsdk:"hdr_format"`
	RedirType    types.String `tfsdk:"redir_type"`
	RedirValue   types.String `tfsdk:"redir_value"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	StatusReason types.String `tfsdk:"status_reason"`
}

func (d *tcpRequestRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_request_rule"
}

func (d *tcpRequestRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tcp_request_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the TCP request rule.",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "The action of the TCP request rule.",
						},
						"cond": schema.StringAttribute{
							Computed:    true,
							Description: "The condition of the TCP request rule.",
						},
						"cond_test": schema.StringAttribute{
							Computed:    true,
							Description: "The condition test of the TCP request rule.",
						},
						"hdr_name": schema.StringAttribute{
							Computed:    true,
							Description: "The header name of the TCP request rule.",
						},
						"hdr_format": schema.StringAttribute{
							Computed:    true,
							Description: "The header format of the TCP request rule.",
						},
						"redir_type": schema.StringAttribute{
							Computed:    true,
							Description: "The redirect type of the TCP request rule.",
						},
						"redir_value": schema.StringAttribute{
							Computed:    true,
							Description: "The redirect value of the TCP request rule.",
						},
						"status_code": schema.Int64Attribute{
							Computed:    true,
							Description: "The status code of the TCP request rule.",
						},
						"status_reason": schema.StringAttribute{
							Computed:    true,
							Description: "The status reason of the TCP request rule.",
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
				Description: "The parent name to get TCP request rules for.",
			},
		},
	}
}

func (d *tcpRequestRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tcpRequestRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tcpRequestRuleDataSourceModel

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

	tcpRequestRules, err := d.client.ReadTcpRequestRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy TCP Request Rules",
			"Could not read HAProxy TCP Request Rules, unexpected error: "+err.Error(),
		)
		return
	}

	for _, rule := range tcpRequestRules {
		state.TcpRequestRules = append(state.TcpRequestRules, tcpRequestRuleItemModel{
			Index:        types.Int64Value(rule.Index),
			Action:       types.StringValue(rule.Action),
			Cond:         types.StringValue(rule.Cond),
			CondTest:     types.StringValue(rule.CondTest),
			HdrName:      types.StringValue(""), // Not available in model
			HdrFormat:    types.StringValue(""), // Not available in model
			RedirType:    types.StringValue(""), // Not available in model
			RedirValue:   types.StringValue(""), // Not available in model
			StatusCode:   types.Int64Value(0),   // Not available in model
			StatusReason: types.StringValue(""), // Not available in model
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
