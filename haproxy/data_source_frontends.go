package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &frontendsDataSource{}
)

// FrontendSingleDataSource defines the single data source implementation.
type FrontendSingleDataSource struct {
	client *HAProxyClient
}

// FrontendSingleDataSourceModel describes the single data source data model.
type FrontendSingleDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Frontend types.String `tfsdk:"frontend"`
}

// Metadata returns the single data source type name.
func (d *FrontendSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontend_single"
}

// Schema defines the schema for the single data source.
func (d *FrontendSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Single Frontend data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Frontend identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Frontend name",
				Required:            true,
			},
			"frontend": schema.StringAttribute{
				MarkdownDescription: "Complete frontend data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *FrontendSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *FrontendSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FrontendSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the frontends
	frontends, err := d.client.ReadFrontends(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read frontends, got error: %s", err))
		return
	}

	// Find the specific frontend by name
	var foundFrontend *FrontendPayload
	for _, frontend := range frontends {
		if frontend.Name == data.Name.ValueString() {
			foundFrontend = &frontend
			break
		}
	}

	if foundFrontend == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Frontend with name %s not found", data.Name.ValueString()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(foundFrontend.Name)
	data.Name = types.StringValue(foundFrontend.Name)

	// Convert frontend to JSON for dynamic output
	jsonData, err := json.Marshal(foundFrontend)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal frontend to JSON, got error: %s", err))
		return
	}
	data.Frontend = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewFrontendsDataSource is a helper function to simplify the provider implementation.
func NewFrontendsDataSource() datasource.DataSource {
	return &frontendsDataSource{}
}

// NewFrontendSingleDataSource creates a new single frontend data source
func NewFrontendSingleDataSource() datasource.DataSource {
	return &FrontendSingleDataSource{}
}

// frontendsDataSource is the data source implementation.
type frontendsDataSource struct {
	client *HAProxyClient
}

// frontendsDataSourceModel maps the data source schema data.
type frontendsDataSourceModel struct {
	Frontends types.String `tfsdk:"frontends"`
}

// Metadata returns the data source type name.
func (d *frontendsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontends"
}

// Schema defines the schema for the data source.
func (d *frontendsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"frontends": schema.StringAttribute{
				Computed:    true,
				Description: "Complete frontends data from HAProxy API as JSON string",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *frontendsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *frontendsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state frontendsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	frontends, err := d.client.ReadFrontends(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading frontends",
			"Could not read frontends, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert frontends to JSON for dynamic output
	jsonData, err := json.Marshal(frontends)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal frontends to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Frontends = types.StringValue(string(jsonData))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
