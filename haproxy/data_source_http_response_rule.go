package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewHttpResponseRuleDataSource() datasource.DataSource {
	return &httpResponseRuleDataSource{}
}

type httpResponseRuleDataSource struct {
	client *HAProxyClient
}

type httpResponseRuleDataSourceModel struct {
	HttpResponseRules []httpResponseRuleItemModel `tfsdk:"http_response_rules"`
}

type httpResponseRuleItemModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Type         types.String `tfsdk:"type"`
	Action       types.String `tfsdk:"action"`
	Cond         types.String `tfsdk:"cond"`
	CondTest     types.String `tfsdk:"cond_test"`
	HdrName      types.String `tfsdk:"hdr_name"`
	HdrFormat    types.String `tfsdk:"hdr_format"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	StatusReason types.String `tfsdk:"status_reason"`
}

func (d *httpResponseRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_response_rule"
}

func (d *httpResponseRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http_response_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the HTTP response rule.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the HTTP response rule.",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "The action of the HTTP response rule.",
						},
						"cond": schema.StringAttribute{
							Computed:    true,
							Description: "The condition of the HTTP response rule.",
						},
						"cond_test": schema.StringAttribute{
							Computed:    true,
							Description: "The condition test of the HTTP response rule.",
						},
						"hdr_name": schema.StringAttribute{
							Computed:    true,
							Description: "The header name of the HTTP response rule.",
						},
						"hdr_format": schema.StringAttribute{
							Computed:    true,
							Description: "The header format of the HTTP response rule.",
						},
						"status_code": schema.Int64Attribute{
							Computed:    true,
							Description: "The status code of the HTTP response rule.",
						},
						"status_reason": schema.StringAttribute{
							Computed:    true,
							Description: "The status reason of the HTTP response rule.",
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
				Description: "The parent name to get HTTP response rules for.",
			},
		},
	}
}

func (d *httpResponseRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *httpResponseRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state httpResponseRuleDataSourceModel

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

	httpResponseRules, err := d.client.ReadHttpResponseRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy HTTP Response Rules",
			"Could not read HAProxy HTTP Response Rules, unexpected error: "+err.Error(),
		)
		return
	}

	for _, rule := range httpResponseRules {
		state.HttpResponseRules = append(state.HttpResponseRules, httpResponseRuleItemModel{
			Index:        types.Int64Value(rule.Index),
			Type:         types.StringValue(rule.Type),
			Action:       types.StringValue(rule.Type), // Use Type as Action for now
			Cond:         types.StringValue(rule.Cond),
			CondTest:     types.StringValue(rule.CondTest),
			HdrName:      types.StringValue(rule.HdrName),
			HdrFormat:    types.StringValue(rule.HdrFormat),
			StatusCode:   types.Int64Value(rule.StatusCode),
			StatusReason: types.StringValue(rule.StatusReason),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
