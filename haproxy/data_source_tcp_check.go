package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTcpCheckDataSource() datasource.DataSource {
	return &tcpCheckDataSource{}
}

type tcpCheckDataSource struct {
	client *HAProxyClient
}

type tcpCheckDataSourceModel struct {
	TcpChecks []tcpCheckItemModel `tfsdk:"tcp_checks"`
}

type tcpCheckItemModel struct {
	Index      types.Int64  `tfsdk:"index"`
	Action     types.String `tfsdk:"action"`
	Comment    types.String `tfsdk:"comment"`
	Port       types.Int64  `tfsdk:"port"`
	Address    types.String `tfsdk:"address"`
	Data       types.String `tfsdk:"data"`
	MinRecv    types.Int64  `tfsdk:"min_recv"`
	OnSuccess  types.String `tfsdk:"on_success"`
	OnError    types.String `tfsdk:"on_error"`
	StatusCode types.String `tfsdk:"status_code"`
	Timeout    types.Int64  `tfsdk:"timeout"`
	LogLevel   types.String `tfsdk:"log_level"`
}

func (d *tcpCheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_check"
}

func (d *tcpCheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tcp_checks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Computed:    true,
							Description: "The index of the TCP check.",
						},
						"action": schema.StringAttribute{
							Computed:    true,
							Description: "The action of the TCP check.",
						},
						"comment": schema.StringAttribute{
							Computed:    true,
							Description: "The comment of the TCP check.",
						},
						"port": schema.Int64Attribute{
							Computed:    true,
							Description: "The port of the TCP check.",
						},
						"address": schema.StringAttribute{
							Computed:    true,
							Description: "The address of the TCP check.",
						},
						"data": schema.StringAttribute{
							Computed:    true,
							Description: "The data of the TCP check.",
						},
						"min_recv": schema.Int64Attribute{
							Computed:    true,
							Description: "The minimum receive bytes of the TCP check.",
						},
						"on_success": schema.StringAttribute{
							Computed:    true,
							Description: "The on success action of the TCP check.",
						},
						"on_error": schema.StringAttribute{
							Computed:    true,
							Description: "The on error action of the TCP check.",
						},
						"status_code": schema.StringAttribute{
							Computed:    true,
							Description: "The status code of the TCP check.",
						},
						"timeout": schema.Int64Attribute{
							Computed:    true,
							Description: "The timeout of the TCP check.",
						},
						"log_level": schema.StringAttribute{
							Computed:    true,
							Description: "The log level of the TCP check.",
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

func (d *tcpCheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tcpCheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tcpCheckDataSourceModel

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

	tcpChecks, err := d.client.ReadTcpChecks(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy TCP Checks",
			"Could not read HAProxy TCP Checks, unexpected error: "+err.Error(),
		)
		return
	}

	for _, tcpCheck := range tcpChecks {
		state.TcpChecks = append(state.TcpChecks, tcpCheckItemModel{
			Index:      types.Int64Value(tcpCheck.Index),
			Action:     types.StringValue(tcpCheck.Action),
			Comment:    types.StringValue(tcpCheck.Comment),
			Port:       types.Int64Value(tcpCheck.Port),
			Address:    types.StringValue(tcpCheck.Address),
			Data:       types.StringValue(tcpCheck.Data),
			MinRecv:    types.Int64Value(tcpCheck.MinRecv),
			OnSuccess:  types.StringValue(tcpCheck.OnSuccess),
			OnError:    types.StringValue(tcpCheck.OnError),
			StatusCode: types.StringValue(tcpCheck.StatusCode),
			Timeout:    types.Int64Value(tcpCheck.Timeout),
			LogLevel:   types.StringValue(tcpCheck.LogLevel),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
