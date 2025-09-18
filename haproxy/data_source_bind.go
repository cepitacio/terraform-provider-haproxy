package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the data source implements the expected interfaces.
var (
	_ datasource.DataSource              = &bindDataSource{}
	_ datasource.DataSourceWithConfigure = &bindDataSource{}
)

// NewBindDataSource is a helper function to simplify the provider implementation.
func NewBindDataSource() datasource.DataSource {
	return &bindDataSource{}
}

// bindDataSource is the data source implementation.
type bindDataSource struct {
	client *HAProxyClient
}

// bindDataSourceModel maps the data source schema data.
type bindDataSourceModel struct {
	Binds      types.String `tfsdk:"binds"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
}

// Metadata returns the data source type name.
func (d *bindDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind"
}

// Schema defines the schema for the data source.
func (d *bindDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all binds from a specific frontend.\n\n## Example Usage\n\n```hcl\n# Get all binds from a frontend\ndata \"haproxy_bind\" \"frontend_binds\" {\n  parent_type = \"frontend\"\n  parent_name = \"web_frontend\"\n}\n\n# Use the binds data\noutput \"bind_count\" {\n  value = length(jsondecode(data.haproxy_bind.frontend_binds.binds))\n}\n\noutput \"bind_addresses\" {\n  value = [for bind in jsondecode(data.haproxy_bind.frontend_binds.binds) : bind.address]\n}\n```",
		Attributes: map[string]schema.Attribute{
			"binds": schema.StringAttribute{
				Computed:    true,
				Description: "Complete bind data from HAProxy API as JSON string",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "Parent type (frontend or backend)",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "Parent name",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *bindDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bindDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bindDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	binds, err := d.client.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Binds",
			"Could not read HAProxy Binds, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert binds to JSON for dynamic output
	jsonData, err := json.Marshal(binds)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal binds to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Binds = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// bindSingleDataSource is the single bind data source implementation.
type bindSingleDataSource struct {
	client *HAProxyClient
}

// bindSingleDataSourceModel maps the single bind data source schema data.
type bindSingleDataSourceModel struct {
	Bind       types.String `tfsdk:"bind"`
	Name       types.String `tfsdk:"name"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
}

// NewBindSingleDataSource is a helper function to simplify the provider implementation.
func NewBindSingleDataSource() datasource.DataSource {
	return &bindSingleDataSource{}
}

// Metadata returns the data source type name.
func (d *bindSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bind_single"
}

// Schema defines the schema for the single bind data source.
func (d *bindSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a single bind by name from a specific frontend.\n\n## Example Usage\n\n```hcl\n# Get a specific bind from a frontend\ndata \"haproxy_bind_single\" \"http_bind\" {\n  name        = \"0.0.0.0:80\"\n  parent_type = \"frontend\"\n  parent_name = \"web_frontend\"\n}\n\n# Use the bind data\noutput \"bind_address\" {\n  value = jsondecode(data.haproxy_bind_single.http_bind.bind).address\n}\n\noutput \"bind_port\" {\n  value = jsondecode(data.haproxy_bind_single.http_bind.bind).port\n}\n```",
		Attributes: map[string]schema.Attribute{
			"bind": schema.StringAttribute{
				Computed:    true,
				Description: "Complete bind data from HAProxy API as JSON string",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Bind name",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "Parent type (frontend or backend)",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "Parent name",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *bindSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bindSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bindSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	bind, err := d.client.ReadBind(ctx, name, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Bind",
			"Could not read HAProxy Bind, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert bind to JSON for dynamic output
	jsonData, err := json.Marshal(bind)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal bind to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Bind = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
