package haproxy

import (
	"context"
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
	TcpRequestRules []TcpRequestRuleItemModel `tfsdk:"tcp_request_rules"`
}

type TcpRequestRuleItemModel struct {
	Index types.Int64 `tfsdk:"index"`
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
			"tcp_request_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the TCP request rule.",
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

// Configure adds the provider configured client to the data source.
func (d *TcpRequestRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TcpRequestRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TcpRequestRuleDataSourceModel

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
	manager := NewTcpRequestRuleManager(d.client)
	rules, err := manager.Read(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP request rules, got error: %s", err))
		return
	}

	// Convert all rules to data source model
	for _, rule := range rules {
		state.TcpRequestRules = append(state.TcpRequestRules, TcpRequestRuleItemModel{
			Index: rule.Index,
		})
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// NewTcpRequestRuleDataSource creates a new TCP request rule data source
func NewTcpRequestRuleDataSource() datasource.DataSource {
	return &TcpRequestRuleDataSource{}
}
