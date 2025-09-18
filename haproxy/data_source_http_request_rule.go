package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HttpRequestRuleSingleDataSource defines the single data source implementation.
type HttpRequestRuleSingleDataSource struct {
	client *HAProxyClient
}

// HttpRequestRuleSingleDataSourceModel describes the single data source data model.
type HttpRequestRuleSingleDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	ParentType      types.String `tfsdk:"parent_type"`
	ParentName      types.String `tfsdk:"parent_name"`
	Index           types.Int64  `tfsdk:"index"`
	HttpRequestRule types.String `tfsdk:"http_request_rule"`
}

// Metadata returns the single data source type name.
func (d *HttpRequestRuleSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_request_rule_single"
}

// Schema defines the schema for the single data source.
func (d *HttpRequestRuleSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single HTTP request rule by index from a specific parent (frontend or backend).\n\n## Example Usage\n\n```hcl\n# Get a specific HTTP request rule from a frontend\ndata \"haproxy_http_request_rule_single\" \"allow_rule\" {\n  parent_type = \"frontend\"\n  parent_name = \"web_frontend\"\n  index       = 0\n}\n\n# Use the rule data\noutput \"rule_type\" {\n  value = jsondecode(data.haproxy_http_request_rule_single.allow_rule.http_request_rule).type\n}\n\noutput \"rule_cond\" {\n  value = jsondecode(data.haproxy_http_request_rule_single.allow_rule.http_request_rule).cond\n}\n```",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "HTTP Request Rule identifier",
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
				MarkdownDescription: "HTTP Request Rule index",
				Required:            true,
			},
			"http_request_rule": schema.StringAttribute{
				MarkdownDescription: "Complete HTTP request rule data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *HttpRequestRuleSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *HttpRequestRuleSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HttpRequestRuleSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the HTTP request rules
	rules, err := d.client.ReadHttpRequestRules(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HTTP request rules, got error: %s", err))
		return
	}

	// Find the specific rule by array position (more predictable than API index)
	var foundRule *HttpRequestRulePayload
	if data.Index.ValueInt64() < int64(len(rules)) {
		foundRule = &rules[data.Index.ValueInt64()]
	}

	if foundRule == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("HTTP request rule at position %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Fix the index to use array position instead of API index
	foundRule.Index = data.Index.ValueInt64()

	// Convert HTTP request rule to JSON for dynamic output
	jsonData, err := json.Marshal(foundRule)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal HTTP request rule to JSON, got error: %s", err))
		return
	}
	data.HttpRequestRule = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func NewHttpRequestRuleDataSource() datasource.DataSource {
	return &httpRequestRuleDataSource{}
}

// NewHttpRequestRuleSingleDataSource creates a new single HTTP request rule data source
func NewHttpRequestRuleSingleDataSource() datasource.DataSource {
	return &HttpRequestRuleSingleDataSource{}
}

type httpRequestRuleDataSource struct {
	client *HAProxyClient
}

type httpRequestRuleDataSourceModel struct {
	HttpRequestRules types.String `tfsdk:"http_request_rules"`
	ParentType       types.String `tfsdk:"parent_type"`
	ParentName       types.String `tfsdk:"parent_name"`
}

func (d *httpRequestRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_request_rule"
}

func (d *httpRequestRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all HTTP request rules from a specific parent (frontend or backend).\n\n## Example Usage\n\n```hcl\n# Get all HTTP request rules from a frontend\ndata \"haproxy_http_request_rule\" \"frontend_rules\" {\n  parent_type = \"frontend\"\n  parent_name = \"web_frontend\"\n}\n\n# Use the rules data\noutput \"rule_count\" {\n  value = length(jsondecode(data.haproxy_http_request_rule.frontend_rules.http_request_rules))\n}\n\noutput \"rule_types\" {\n  value = [for rule in jsondecode(data.haproxy_http_request_rule.frontend_rules.http_request_rules) : rule.type]\n}\n```",
		Attributes: map[string]schema.Attribute{
			"http_request_rules": schema.StringAttribute{
				Computed:    true,
				Description: "Complete HTTP request rules data from HAProxy API as JSON string",
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

func (d *httpRequestRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state httpRequestRuleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	httpRequestRules, err := d.client.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy HTTP Request Rules",
			"Could not read HAProxy HTTP Request Rules, unexpected error: "+err.Error(),
		)
		return
	}

	// Fix index field - use array position if API returns 0 for all rules
	for i := range httpRequestRules {
		httpRequestRules[i].Index = int64(i)
	}

	// Convert rules to JSON for dynamic output
	jsonData, err := json.Marshal(httpRequestRules)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal HTTP request rules to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.HttpRequestRules = types.StringValue(string(jsonData))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
