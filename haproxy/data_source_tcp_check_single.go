package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TcpCheckSingleDataSource defines the data source implementation.
type TcpCheckSingleDataSource struct {
	client *HAProxyClient
}

// TcpCheckSingleDataSourceModel describes the data source data model.
type TcpCheckSingleDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
	Index      types.Int64  `tfsdk:"index"`
	Action     types.String `tfsdk:"action"`
	Data       types.String `tfsdk:"data"`
	Match      types.String `tfsdk:"match"`
	Pattern    types.String `tfsdk:"pattern"`
}

// Metadata returns the data source type name.
func (d *TcpCheckSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_check_single"
}

// Schema defines the schema for the data source.
func (d *TcpCheckSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Single TCP Check data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "TCP check identifier",
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
				MarkdownDescription: "Check index",
				Required:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Check action",
				Computed:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Data",
				Computed:            true,
			},
			"match": schema.StringAttribute{
				MarkdownDescription: "Match",
				Computed:            true,
			},
			"pattern": schema.StringAttribute{
				MarkdownDescription: "Pattern",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TcpCheckSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TcpCheckSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TcpCheckSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the checks
	manager := NewTcpCheckManager(d.client)
	checks, err := manager.Read(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP checks, got error: %s", err))
		return
	}

	// Find the specific check by index
	var foundCheck *TcpCheckResourceModel
	for _, check := range checks {
		if check.Index.ValueInt64() == data.Index.ValueInt64() {
			foundCheck = &check
			break
		}
	}

	if foundCheck == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("TCP check with index %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = foundCheck.ID
	data.Action = foundCheck.Action
	data.Data = foundCheck.Data
	data.Match = foundCheck.Match
	data.Pattern = foundCheck.Pattern

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewTcpCheckSingleDataSource creates a new single TCP check data source
func NewTcpCheckSingleDataSource() datasource.DataSource {
	return &TcpCheckSingleDataSource{}
}
