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
	_ datasource.DataSource = &backendsDataSource{}
)

// BackendSingleDataSource defines the single data source implementation.
type BackendSingleDataSource struct {
	client *HAProxyClient
}

// BackendSingleDataSourceModel describes the single data source data model.
type BackendSingleDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Backend types.String `tfsdk:"backend"`
}

// Metadata returns the single data source type name.
func (d *BackendSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_single"
}

// Schema defines the schema for the single data source.
func (d *BackendSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Single Backend data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Backend identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Backend name",
				Required:            true,
			},
			"backend": schema.StringAttribute{
				MarkdownDescription: "Complete backend data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *BackendSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *BackendSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackendSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the backends
	backends, err := d.client.ReadBackends(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read backends, got error: %s", err))
		return
	}

	// Find the specific backend by name
	var foundBackend *BackendPayload
	for _, backend := range backends {
		if backend.Name == data.Name.ValueString() {
			foundBackend = &backend
			break
		}
	}

	if foundBackend == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Backend with name %s not found", data.Name.ValueString()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(foundBackend.Name)
	data.Name = types.StringValue(foundBackend.Name)

	// Convert backend to JSON for dynamic output
	jsonData, err := json.Marshal(foundBackend)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal backend to JSON, got error: %s", err))
		return
	}
	data.Backend = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewBackendsDataSource is a helper function to simplify the provider implementation.
func NewBackendsDataSource() datasource.DataSource {
	return &backendsDataSource{}
}

// NewBackendSingleDataSource creates a new single backend data source
func NewBackendSingleDataSource() datasource.DataSource {
	return &BackendSingleDataSource{}
}

// backendsDataSource is the data source implementation.
type backendsDataSource struct {
	client *HAProxyClient
}

// backendsDataSourceModel maps the data source schema data.
type backendsDataSourceModel struct {
	Backends types.String `tfsdk:"backends"`
}

// Metadata returns the data source type name.
func (d *backendsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backends"
}

// Schema defines the schema for the data source.
func (d *backendsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backends": schema.StringAttribute{
				Computed:    true,
				Description: "Complete backends data from HAProxy API as JSON string",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *backendsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *backendsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state backendsDataSourceModel
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

	backends, err := d.client.ReadBackends(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backends",
			"Could not read backends, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert backends to JSON for dynamic output
	jsonData, err := json.Marshal(backends)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal backends to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Backends = types.StringValue(string(jsonData))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
