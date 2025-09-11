package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TcpRequestRuleDataSource defines the data source implementation.
type TcpRequestRuleDataSource struct {
	client *HAProxyClient
}

// TcpRequestRuleDataSourceModel describes the data source data model.
type TcpRequestRuleDataSourceModel struct {
	TcpRequestRules types.String `tfsdk:"tcp_request_rules"`
	ParentType      types.String `tfsdk:"parent_type"`
	ParentName      types.String `tfsdk:"parent_name"`
}

// Metadata returns the data source type name.
func (d *TcpRequestRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_request_rule"
}

// Schema defines the schema for the data source.
func (d *TcpRequestRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TCP Request Rule data source",

		Attributes: map[string]schema.Attribute{
			"tcp_request_rules": schema.StringAttribute{
				Computed:    true,
				Description: "Complete TCP request rules data from HAProxy API as JSON string",
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

// Configure adds the provider configured client to the data source.
func (d *TcpRequestRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *TcpRequestRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TcpRequestRuleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	// Read the rules
	rules, err := d.client.ReadTcpRequestRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP request rules, got error: %s", err))
		return
	}

	// Fix index field - use array position if API returns 0 for all rules
	for i := range rules {
		rules[i].Index = int64(i)
	}

	// Convert rules to JSON for dynamic output
	jsonData, err := json.Marshal(rules)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal TCP request rules to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.TcpRequestRules = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// TcpRequestRuleSingleDataSource defines the single data source implementation.
type TcpRequestRuleSingleDataSource struct {
	client *HAProxyClient
}

// TcpRequestRuleSingleDataSourceModel describes the single data source data model.
type TcpRequestRuleSingleDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ParentType     types.String `tfsdk:"parent_type"`
	ParentName     types.String `tfsdk:"parent_name"`
	Index          types.Int64  `tfsdk:"index"`
	TcpRequestRule types.String `tfsdk:"tcp_request_rule"`
}

// Metadata returns the single data source type name.
func (d *TcpRequestRuleSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_request_rule_single"
}

// Schema defines the schema for the single data source.
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
			"tcp_request_rule": schema.StringAttribute{
				MarkdownDescription: "Complete TCP request rule data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *TcpRequestRuleSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
}

// Read refreshes the Terraform state with the latest data for single data source.
func (d *TcpRequestRuleSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TcpRequestRuleSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the rules directly from API
	rules, err := d.client.ReadTcpRequestRules(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP request rules, got error: %s", err))
		return
	}

	// Find the specific rule by array position (more predictable than API index)
	var foundRule *TcpRequestRulePayload
	if data.Index.ValueInt64() < int64(len(rules)) {
		foundRule = &rules[data.Index.ValueInt64()]
	}

	if foundRule == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("TCP request rule at position %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Fix the index to use array position instead of API index
	foundRule.Index = data.Index.ValueInt64()

	// Convert TCP request rule to JSON for dynamic output
	jsonData, err := json.Marshal(foundRule)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal TCP request rule to JSON, got error: %s", err))
		return
	}
	data.TcpRequestRule = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewTcpRequestRuleDataSource creates a new TCP request rule data source
func NewTcpRequestRuleDataSource() datasource.DataSource {
	return &TcpRequestRuleDataSource{}
}

// NewTcpRequestRuleSingleDataSource creates a new single TCP request rule data source
func NewTcpRequestRuleSingleDataSource() datasource.DataSource {
	return &TcpRequestRuleSingleDataSource{}
}
