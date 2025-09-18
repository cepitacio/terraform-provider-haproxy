package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HttpResponseRuleSingleDataSource defines the single data source implementation.
type HttpResponseRuleSingleDataSource struct {
	client *HAProxyClient
}

// HttpResponseRuleSingleDataSourceModel describes the single data source data model.
type HttpResponseRuleSingleDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	ParentType       types.String `tfsdk:"parent_type"`
	ParentName       types.String `tfsdk:"parent_name"`
	Index            types.Int64  `tfsdk:"index"`
	HttpResponseRule types.String `tfsdk:"http_response_rule"`
}

// Metadata returns the single data source type name.
func (d *HttpResponseRuleSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_response_rule_single"
}

// Schema defines the schema for the single data source.
func (d *HttpResponseRuleSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single HTTP response rule by index from a specific parent (frontend or backend).\n\n## Example Usage\n\n```hcl\n# Get a specific HTTP response rule from a backend\ndata \"haproxy_http_response_rule_single\" \"set_header_rule\" {\n  parent_type = \"backend\"\n  parent_name = \"web_backend\"\n  index       = 0\n}\n\n# Use the rule data\noutput \"rule_type\" {\n  value = jsondecode(data.haproxy_http_response_rule_single.set_header_rule.http_response_rule).type\n}\n```",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "HTTP Response Rule identifier",
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
				MarkdownDescription: "HTTP Response Rule index",
				Required:            true,
			},
			"http_response_rule": schema.StringAttribute{
				MarkdownDescription: "Complete HTTP response rule data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *HttpResponseRuleSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *HttpResponseRuleSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HttpResponseRuleSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the HTTP response rules
	rules, err := d.client.ReadHttpResponseRules(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HTTP response rules, got error: %s", err))
		return
	}

	// Find the specific rule by array position (more predictable than API index)
	var foundRule *HttpResponseRulePayload
	if data.Index.ValueInt64() < int64(len(rules)) {
		foundRule = &rules[data.Index.ValueInt64()]
	}

	if foundRule == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("HTTP response rule at position %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Fix the index to use array position instead of API index
	foundRule.Index = data.Index.ValueInt64()

	// Convert HTTP response rule to JSON for dynamic output
	jsonData, err := json.Marshal(foundRule)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal HTTP response rule to JSON, got error: %s", err))
		return
	}
	data.HttpResponseRule = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func NewHttpResponseRuleDataSource() datasource.DataSource {
	return &httpResponseRuleDataSource{}
}

// NewHttpResponseRuleSingleDataSource creates a new single HTTP response rule data source
func NewHttpResponseRuleSingleDataSource() datasource.DataSource {
	return &HttpResponseRuleSingleDataSource{}
}

type httpResponseRuleDataSource struct {
	client *HAProxyClient
}

type httpResponseRuleDataSourceModel struct {
	HttpResponseRules types.String `tfsdk:"http_response_rules"`
	ParentType        types.String `tfsdk:"parent_type"`
	ParentName        types.String `tfsdk:"parent_name"`
}

func (d *httpResponseRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_response_rule"
}

func (d *httpResponseRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all HTTP response rules from a specific parent (frontend or backend).\n\n## Example Usage\n\n```hcl\n# Get all HTTP response rules from a backend\ndata \"haproxy_http_response_rule\" \"backend_rules\" {\n  parent_type = \"backend\"\n  parent_name = \"web_backend\"\n}\n\n# Use the rules data\noutput \"rule_count\" {\n  value = length(jsondecode(data.haproxy_http_response_rule.backend_rules.http_response_rules))\n}\n```",
		Attributes: map[string]schema.Attribute{
			"http_response_rules": schema.StringAttribute{
				Computed:    true,
				Description: "Complete HTTP response rules data from HAProxy API as JSON string",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "The parent type (backend or frontend).",
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

func (d *httpResponseRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state httpResponseRuleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	httpResponseRules, err := d.client.ReadHttpResponseRules(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy HTTP Response Rules",
			"Could not read HAProxy HTTP Response Rules, unexpected error: "+err.Error(),
		)
		return
	}

	// Fix index field - use array position if API returns 0 for all rules
	for i := range httpResponseRules {
		httpResponseRules[i].Index = int64(i)
	}

	// Convert rules to JSON for dynamic output
	jsonData, err := json.Marshal(httpResponseRules)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal HTTP response rules to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.HttpResponseRules = types.StringValue(string(jsonData))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
