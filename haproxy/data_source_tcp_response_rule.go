package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TcpResponseRuleDataSource defines the data source implementation.
type TcpResponseRuleDataSource struct {
	client *HAProxyClient
}

// TcpResponseRuleDataSourceModel describes the data source data model.
type TcpResponseRuleDataSourceModel struct {
	TcpResponseRules []TcpResponseRuleItemModel `tfsdk:"tcp_response_rules"`
}

type TcpResponseRuleItemModel struct {
	Index types.Int64 `tfsdk:"index"`
}

// Metadata returns the data source type name.
func (d *TcpResponseRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_response_rule"
}

// Schema defines the schema for the data source.
func (d *TcpResponseRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TCP Response Rule data source",

		Attributes: map[string]schema.Attribute{
			"tcp_response_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the TCP response rule.",
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

// Configure adds the provider configured client to the data source.
func (d *TcpResponseRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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
func (d *TcpResponseRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TcpResponseRuleDataSourceModel

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

	// Read the rules
	manager := NewTcpResponseRuleManager(d.client)
	rules, err := manager.Read(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP response rules, got error: %s", err))
		return
	}

	// Convert all rules to data source model
	for _, rule := range rules {
		state.TcpResponseRules = append(state.TcpResponseRules, TcpResponseRuleItemModel{
			Index: rule.Index,
		})
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// NewTcpResponseRuleDataSource creates a new TCP response rule data source
func NewTcpResponseRuleDataSource() datasource.DataSource {
	return &TcpResponseRuleDataSource{}
}
