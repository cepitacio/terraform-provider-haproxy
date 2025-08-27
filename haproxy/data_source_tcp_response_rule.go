package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTcpResponseRuleDataSource() datasource.DataSource {
	return &tcpResponseRuleDataSource{}
}

type tcpResponseRuleDataSource struct {
	client *HAProxyClient
}

type tcpResponseRuleDataSourceModel struct {
	TcpResponseRules []tcpResponseRuleItemModel `tfsdk:"tcp_response_rules"`
}

type tcpResponseRuleItemModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Action       types.String `tfsdk:"action"`
	Cond         types.String `tfsdk:"cond"`
	CondTest     types.String `tfsdk:"cond_test"`
	HdrName      types.String `tfsdk:"hdr_name"`
	HdrFormat    types.String `tfsdk:"hdr_format"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	StatusReason types.String `tfsdk:"status_reason"`
}

func (d *tcpResponseRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_response_rule"
}

func (d *tcpResponseRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tcp_response_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the TCP response rule.",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "The action of the TCP response rule.",
						},
						"cond": schema.StringAttribute{
							Computed:    true,
							Description: "The condition of the TCP response rule.",
						},
						"cond_test": schema.StringAttribute{
							Computed:    true,
							Description: "The condition test of the TCP response rule.",
						},
						"hdr_name": schema.StringAttribute{
							Computed:    true,
							Description: "The header name of the TCP response rule.",
						},
						"hdr_format": schema.StringAttribute{
							Computed:    true,
							Description: "The header format of the TCP response rule.",
						},
						"status_code": schema.Int64Attribute{
							Computed:    true,
							Description: "The status code of the TCP response rule.",
						},
						"status_reason": schema.StringAttribute{
							Computed:    true,
							Description: "The status reason of the TCP response rule.",
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
				Description: "The parent name to get TCP response rules for.",
			},
		},
	}
}

func (d *tcpResponseRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tcpResponseRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tcpResponseRuleDataSourceModel

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

	tcpResponseRules, err := d.client.ReadTcpResponseRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy TCP Response Rules",
			"Could not read HAProxy TCP Response Rules, unexpected error: "+err.Error(),
		)
		return
	}

	for _, rule := range tcpResponseRules {
		state.TcpResponseRules = append(state.TcpResponseRules, tcpResponseRuleItemModel{
			Index:        types.Int64Value(rule.Index),
			Action:       types.StringValue(rule.Action),
			Cond:         types.StringValue(rule.Cond),
			CondTest:     types.StringValue(rule.CondTest),
			HdrName:      types.StringValue(""), // Not available in model
			HdrFormat:    types.StringValue(""), // Not available in model
			StatusCode:   types.Int64Value(0),   // Not available in model
			StatusReason: types.StringValue(""), // Not available in model
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
