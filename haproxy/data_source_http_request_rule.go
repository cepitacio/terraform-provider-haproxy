package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewHttpRequestRuleDataSource() datasource.DataSource {
	return &httpRequestRuleDataSource{}
}

type httpRequestRuleDataSource struct {
	client *HAProxyClient
}

type httpRequestRuleDataSourceModel struct {
	HttpRequestRules []httpRequestRuleItemModel `tfsdk:"http_request_rules"`
}

type httpRequestRuleItemModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Type         types.String `tfsdk:"type"`
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

func (d *httpRequestRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_request_rule"
}

func (d *httpRequestRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http_request_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the HTTP request rule.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the HTTP request rule.",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "The action of the HTTP request rule.",
						},
						"cond": schema.StringAttribute{
							Computed:    true,
							Description: "The condition of the HTTP request rule.",
						},
						"cond_test": schema.StringAttribute{
							Computed:    true,
							Description: "The condition test of the HTTP request rule.",
						},
						"hdr_name": schema.StringAttribute{
							Computed:    true,
							Description: "The header name of the HTTP request rule.",
						},
						"hdr_format": schema.StringAttribute{
							Computed:    true,
							Description: "The header format of the HTTP request rule.",
						},
						"redir_type": schema.StringAttribute{
							Computed:    true,
							Description: "The redirect type of the HTTP request rule.",
						},
						"redir_value": schema.StringAttribute{
							Computed:    true,
							Description: "The redirect value of the HTTP request rule.",
						},
						"status_code": schema.Int64Attribute{
							Computed:    true,
							Description: "The status code of the HTTP request rule.",
						},
						"status_reason": schema.StringAttribute{
							Computed:    true,
							Description: "The status reason of the HTTP request rule.",
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
				Description: "The parent name to get HTTP request rules for.",
			},
		},
	}
}

func (d *httpRequestRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *httpRequestRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state httpRequestRuleDataSourceModel

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

	httpRequestRules, err := d.client.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy HTTP Request Rules",
			"Could not read HAProxy HTTP Request Rules, unexpected error: "+err.Error(),
		)
		return
	}

	for _, rule := range httpRequestRules {
		state.HttpRequestRules = append(state.HttpRequestRules, httpRequestRuleItemModel{
			Index:        types.Int64Value(rule.Index),
			Type:         types.StringValue(rule.Type),
			Action:       types.StringValue(rule.Type), // Use Type as Action for now
			Cond:         types.StringValue(rule.Cond),
			CondTest:     types.StringValue(rule.CondTest),
			HdrName:      types.StringValue(rule.HdrName),
			HdrFormat:    types.StringValue(rule.HdrFormat),
			RedirType:    types.StringValue(rule.RedirType),
			RedirValue:   types.StringValue(rule.RedirValue),
			StatusCode:   types.Int64Value(rule.StatusCode),
			StatusReason: types.StringValue(rule.StatusReason),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
