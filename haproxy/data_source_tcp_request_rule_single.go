package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TcpRequestRuleSingleDataSource defines the data source implementation.
type TcpRequestRuleSingleDataSource struct {
	client *HAProxyClient
}

// TcpRequestRuleSingleDataSourceModel describes the data source data model.
type TcpRequestRuleSingleDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
	Index      types.Int64  `tfsdk:"index"`
	Type       types.String `tfsdk:"type"`
	Action     types.String `tfsdk:"action"`
	Cond       types.String `tfsdk:"cond"`
	CondTest   types.String `tfsdk:"cond_test"`
}

// Metadata returns the data source type name.
func (d *TcpRequestRuleSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_request_rule_single"
}

// Schema defines the schema for the data source.
func (d *TcpRequestRuleSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Single TCP Request Rule data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "TCP request rule identifier",
				Computed:            true,
			},
			"parent_type": schema.StringAttribute{
				MarkdownDescription: "Parent type (frontend or backend)",
				Required:            true,
			},
			"parent_name": schema.StringAttribute{
				MarkdownDescription: "Parent name",
				Required:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Rule index",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Rule type",
				Computed:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Rule action",
				Computed:            true,
			},
			"cond": schema.StringAttribute{
				MarkdownDescription: "Condition",
				Computed:            true,
			},
			"cond_test": schema.StringAttribute{
				MarkdownDescription: "Condition test",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TcpRequestRuleSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TcpRequestRuleSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TcpRequestRuleSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the rules
	manager := NewTcpRequestRuleManager(d.client)
	rules, err := manager.Read(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP request rules, got error: %s", err))
		return
	}

	// Find the specific rule by index
	var foundRule *TcpRequestRuleResourceModel
	for _, rule := range rules {
		if rule.Index.ValueInt64() == data.Index.ValueInt64() {
			foundRule = &rule
			break
		}
	}

	if foundRule == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("TCP request rule with index %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = foundRule.ID
	data.Type = foundRule.Type
	data.Action = foundRule.Action
	data.Cond = foundRule.Cond
	data.CondTest = foundRule.CondTest

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewTcpRequestRuleSingleDataSource creates a new single TCP request rule data source
func NewTcpRequestRuleSingleDataSource() datasource.DataSource {
	return &TcpRequestRuleSingleDataSource{}
}
