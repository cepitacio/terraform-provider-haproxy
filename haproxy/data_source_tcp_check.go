package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TcpCheckDataSource defines the data source implementation.
type TcpCheckDataSource struct {
	client *HAProxyClient
}

// TcpCheckDataSourceModel describes the data source data model.
type TcpCheckDataSourceModel struct {
	TcpChecks  types.String `tfsdk:"tcp_checks"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
}

// Metadata returns the data source type name.
func (d *TcpCheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_check"
}

// Schema defines the schema for the data source.
func (d *TcpCheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TCP Check data source",

		Attributes: map[string]schema.Attribute{
			"tcp_checks": schema.StringAttribute{
				Computed:    true,
				Description: "Complete TCP checks data from HAProxy API as JSON string",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "The parent type (frontend or backend).",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "The parent name to get TCP checks for.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TcpCheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TcpCheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TcpCheckDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	// Read the checks directly from API
	checks, err := d.client.ReadTcpChecks(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP checks, got error: %s", err))
		return
	}

	// Fix index field - use array position if API returns 0 for all rules
	for i := range checks {
		checks[i].Index = int64(i)
	}

	// Convert checks to JSON for dynamic output
	jsonData, err := json.Marshal(checks)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal TCP checks to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.TcpChecks = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// TcpCheckSingleDataSource defines the single data source implementation.
type TcpCheckSingleDataSource struct {
	client *HAProxyClient
}

// TcpCheckSingleDataSourceModel describes the single data source data model.
type TcpCheckSingleDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
	Index      types.Int64  `tfsdk:"index"`
	TcpCheck   types.String `tfsdk:"tcp_check"`
}

// Metadata returns the single data source type name.
func (d *TcpCheckSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_check_single"
}

// Schema defines the schema for the single data source.
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
			"tcp_check": schema.StringAttribute{
				MarkdownDescription: "Complete TCP check data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *TcpCheckSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TcpCheckSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TcpCheckSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the checks directly from API
	checks, err := d.client.ReadTcpChecks(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP checks, got error: %s", err))
		return
	}

	// Find the specific check by array position (more predictable than API index)
	var foundCheck *TcpCheckPayload
	if data.Index.ValueInt64() < int64(len(checks)) {
		foundCheck = &checks[data.Index.ValueInt64()]
	}

	if foundCheck == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("TCP check at position %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Fix the index to use array position instead of API index
	foundCheck.Index = data.Index.ValueInt64()

	// Convert TCP check to JSON for dynamic output
	jsonData, err := json.Marshal(foundCheck)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal TCP check to JSON, got error: %s", err))
		return
	}
	data.TcpCheck = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewTcpCheckDataSource creates a new TCP check data source
func NewTcpCheckDataSource() datasource.DataSource {
	return &TcpCheckDataSource{}
}

// NewTcpCheckSingleDataSource creates a new single TCP check data source
func NewTcpCheckSingleDataSource() datasource.DataSource {
	return &TcpCheckSingleDataSource{}
}
