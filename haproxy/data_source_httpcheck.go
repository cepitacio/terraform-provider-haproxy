package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HttpcheckSingleDataSource defines the single data source implementation.
type HttpcheckSingleDataSource struct {
	client *HAProxyClient
}

// HttpcheckSingleDataSourceModel describes the single data source data model.
type HttpcheckSingleDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
	Index      types.Int64  `tfsdk:"index"`
	Httpcheck  types.String `tfsdk:"httpcheck"`
}

// Metadata returns the single data source type name.
func (d *HttpcheckSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_httpcheck_single"
}

// Schema defines the schema for the single data source.
func (d *HttpcheckSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Single HTTP Check data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "HTTP Check identifier",
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
				MarkdownDescription: "HTTP Check index",
				Required:            true,
			},
			"httpcheck": schema.StringAttribute{
				MarkdownDescription: "Complete HTTP check data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *HttpcheckSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *HttpcheckSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HttpcheckSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the HTTP checks
	httpchecks, err := d.client.ReadHttpchecks(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HTTP checks, got error: %s", err))
		return
	}

	// Find the specific HTTP check by array position (more predictable than API index)
	var foundHttpcheck *HttpcheckPayload
	if data.Index.ValueInt64() < int64(len(httpchecks)) {
		foundHttpcheck = &httpchecks[data.Index.ValueInt64()]
	}

	if foundHttpcheck == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("HTTP check at position %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Fix the index to use array position instead of API index
	foundHttpcheck.Index = data.Index.ValueInt64()

	// Convert HTTP check to JSON for dynamic output
	jsonData, err := json.Marshal(foundHttpcheck)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal HTTP check to JSON, got error: %s", err))
		return
	}
	data.Httpcheck = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func NewHttpcheckDataSource() datasource.DataSource {
	return &httpcheckDataSource{}
}

// NewHttpcheckSingleDataSource creates a new single HTTP check data source
func NewHttpcheckSingleDataSource() datasource.DataSource {
	return &HttpcheckSingleDataSource{}
}

type httpcheckDataSource struct {
	client *HAProxyClient
}

type httpcheckDataSourceModel struct {
	Httpchecks types.String `tfsdk:"http_checks"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
}

func (d *httpcheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_httpcheck"
}

func (d *httpcheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http_checks": schema.StringAttribute{
				Computed:    true,
				Description: "Complete HTTP checks data from HAProxy API as JSON string",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "The parent type (frontend or backend).",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "The parent name to get HTTP checks for.",
			},
		},
	}
}

func (d *httpcheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *httpcheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state httpcheckDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	httpchecks, err := d.client.ReadHttpchecks(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy HTTP Checks",
			"Could not read HAProxy HTTP Checks, unexpected error: "+err.Error(),
		)
		return
	}

	// Fix index field - use array position if API returns 0 for all rules
	for i := range httpchecks {
		httpchecks[i].Index = int64(i)
	}

	// Convert checks to JSON for dynamic output
	jsonData, err := json.Marshal(httpchecks)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal HTTP checks to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Httpchecks = types.StringValue(string(jsonData))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
