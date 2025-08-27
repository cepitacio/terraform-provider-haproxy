package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPeerEntryDataSource() datasource.DataSource {
	return &peerEntryDataSource{}
}

type peerEntryDataSource struct {
	client *HAProxyClient
}

type peerEntryDataSourceModel struct {
	PeerEntries []peerEntryItemModel `tfsdk:"peer_entries"`
}

type peerEntryItemModel struct {
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
	Peers   types.String `tfsdk:"peers"`
}

func (d *peerEntryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_peer_entry"
}

func (d *peerEntryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"peer_entries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the peer entry.",
						},
						"address": schema.StringAttribute{
							Computed:    true,
							Description: "The address of the peer entry.",
						},
						"port": schema.Int64Attribute{
							Computed:    true,
							Description: "The port of the peer entry.",
						},
						"peers": schema.StringAttribute{
							Computed:    true,
							Description: "The peers group of the peer entry.",
						},
					},
				},
			},
			"peers": schema.StringAttribute{
				Required:    true,
				Description: "The peers group name to get peer entries for.",
			},
		},
	}
}

func (d *peerEntryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *peerEntryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state peerEntryDataSourceModel

	var config struct {
		Peers types.String `tfsdk:"peers"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	peersName := config.Peers.ValueString()

	peerEntries, err := d.client.ReadPeerEntries(ctx, peersName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy Peer Entries",
			"Could not read HAProxy Peer Entries, unexpected error: "+err.Error(),
		)
		return
	}

	for _, peerEntry := range peerEntries {
		state.PeerEntries = append(state.PeerEntries, peerEntryItemModel{
			Name:    types.StringValue(peerEntry.Name),
			Address: types.StringValue(peerEntry.Address),
			Port:    types.Int64Value(peerEntry.Port),
			Peers:   types.StringValue(peersName),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
} 
