package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ServerSingleDataSource defines the single data source implementation.
type ServerSingleDataSource struct {
	client *HAProxyClient
}

// ServerSingleDataSourceModel describes the single data source data model.
type ServerSingleDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Backend types.String `tfsdk:"backend"`
	Server  types.String `tfsdk:"server"`
}

// Metadata returns the single data source type name.
func (d *ServerSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_single"
}

// Schema defines the schema for the single data source.
func (d *ServerSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Single Server data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Required:            true,
			},
			"backend": schema.StringAttribute{
				MarkdownDescription: "Backend name",
				Required:            true,
			},
			"server": schema.StringAttribute{
				MarkdownDescription: "Complete server data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *ServerSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ServerSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the servers
	servers, err := d.client.ReadServers(ctx, "backend", data.Backend.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read servers, got error: %s", err))
		return
	}

	// Find the specific server by name
	var foundServer *ServerPayload
	for _, server := range servers {
		if server.Name == data.Name.ValueString() {
			foundServer = &server
			break
		}
	}

	if foundServer == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Server with name %s not found in backend %s", data.Name.ValueString(), data.Backend.ValueString()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(foundServer.Name)
	data.Name = types.StringValue(foundServer.Name)
	data.Backend = types.StringValue(data.Backend.ValueString())

	// Convert server to JSON for dynamic output
	jsonData, err := json.Marshal(foundServer)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal server to JSON, got error: %s", err))
		return
	}
	data.Server = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// NewServerSingleDataSource creates a new single server data source
func NewServerSingleDataSource() datasource.DataSource {
	return &ServerSingleDataSource{}
}

type serverDataSource struct {
	client *HAProxyClient
}

type serverDataSourceModel struct {
	Servers types.String `tfsdk:"servers"`
	Backend types.String `tfsdk:"backend"`
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"servers": schema.StringAttribute{
				Computed:    true,
				Description: "Complete servers data from HAProxy API as JSON string",
			},
			"backend": schema.StringAttribute{
				Required:    true,
				Description: "The backend name to get servers for.",
			},
		},
	}
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serverDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	backendName := state.Backend.ValueString()

	servers, err := d.client.ReadServers(ctx, "backend", backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Servers",
			"Could not read HAProxy Servers, unexpected error: "+err.Error(),
		)
		return
	}

	// Convert servers to JSON for dynamic output
	jsonData, err := json.Marshal(servers)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal servers to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Servers = types.StringValue(string(jsonData))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
