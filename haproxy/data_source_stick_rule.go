package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewStickRuleDataSource() datasource.DataSource {
	return &stickRuleDataSource{}
}

type stickRuleDataSource struct {
	client *HAProxyClient
}

type stickRuleDataSourceModel struct {
	StickRules []stickRuleItemModel `tfsdk:"stick_rules"`
}

type stickRuleItemModel struct {
	Index   types.Int64  `tfsdk:"index"`
	Type    types.String `tfsdk:"type"`
	Cond    types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
	Pattern types.String `tfsdk:"pattern"`
	Table   types.String `tfsdk:"table"`
	Backend types.String `tfsdk:"backend"`
}

func (d *stickRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stick_rule"
}

func (d *stickRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"stick_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the stick rule.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the stick rule.",
						},
						"cond": schema.StringAttribute{
							Computed:    true,
							Description: "The condition of the stick rule.",
						},
						"cond_test": schema.StringAttribute{
							Computed:    true,
							Description: "The condition test of the stick rule.",
						},
						"pattern": schema.StringAttribute{
							Computed:    true,
							Description: "The pattern of the stick rule.",
						},
						"table": schema.StringAttribute{
							Computed:    true,
							Description: "The table of the stick rule.",
						},
						"backend": schema.StringAttribute{
							Computed:    true,
							Description: "The backend of the stick rule.",
						},
					},
				},
			},
			"backend": schema.StringAttribute{
				Required:    true,
				Description: "The backend name to get stick rules for.",
			},
		},
	}
}

func (d *stickRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *stickRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state stickRuleDataSourceModel

	var config struct {
		Backend types.String `tfsdk:"backend"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backendName := config.Backend.ValueString()

	stickRules, err := d.client.ReadStickRules(ctx, backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Stick Rules",
			"Could not read HAProxy Stick Rules, unexpected error: "+err.Error(),
		)
		return
	}

	for _, rule := range stickRules {
		state.StickRules = append(state.StickRules, stickRuleItemModel{
			Index:    types.Int64Value(rule.Index),
			Type:     types.StringValue(rule.Type),
			Cond:     types.StringValue(rule.Cond),
			CondTest: types.StringValue(rule.CondTest),
			Pattern:  types.StringValue(rule.Pattern),
			Table:    types.StringValue(rule.Table),
			Backend:  types.StringValue(backendName),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
} 
