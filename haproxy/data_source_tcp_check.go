package haproxy

import (
	"context"
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
	TcpChecks []TcpCheckItemModel `tfsdk:"tcp_checks"`
}

type TcpCheckItemModel struct {
	Index types.Int64 `tfsdk:"index"`
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
			"tcp_checks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the TCP check.",
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
func (d *TcpCheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TcpCheckDataSourceModel

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

	// Read the checks
	manager := NewTcpCheckManager(d.client)
	checks, err := manager.Read(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP checks, got error: %s", err))
		return
	}

	// Convert all checks to data source model
	for _, check := range checks {
		state.TcpChecks = append(state.TcpChecks, TcpCheckItemModel{
			Index: check.Index,
		})
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// NewTcpCheckDataSource creates a new TCP check data source
func NewTcpCheckDataSource() datasource.DataSource {
	return &TcpCheckDataSource{}
}
