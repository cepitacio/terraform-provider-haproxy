package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &serverResource{}
)

// NewServerResource is a helper function to simplify the provider implementation.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

// serverResource is the resource implementation.
type serverResource struct {
	client *HAProxyClient
}

// serverStandaloneResourceModel maps the resource schema data.
type serverStandaloneResourceModel struct {
	Name            types.String `tfsdk:"name"`
	Port            types.Int64  `tfsdk:"port"`
	Address         types.String `tfsdk:"address"`
	ParentName      types.String `tfsdk:"parent_name"`
	ParentType      types.String `tfsdk:"parent_type"`
	AgentAddr       types.String `tfsdk:"agent_addr"`
	AgentCheck      types.String `tfsdk:"agent_check"`
	AgentInter      types.Int64  `tfsdk:"agent_inter"`
	AgentPort       types.Int64  `tfsdk:"agent_port"`
	AgentSend       types.String `tfsdk:"agent_send"`
	Allow0rtt       types.Bool   `tfsdk:"allow_0rtt"`
	Alpn            types.String `tfsdk:"alpn"`
	Backup          types.String `tfsdk:"backup"`
	Check           types.String `tfsdk:"check"`
	CheckAlpn       types.String `tfsdk:"check_alpn"`
	CheckSni        types.String `tfsdk:"check_sni"`
	CheckSsl        types.String `tfsdk:"check_ssl"`
	CheckViaSocks4  types.String `tfsdk:"check_via_socks4"`
	Ciphers         types.String `tfsdk:"ciphers"`
	Ciphersuites    types.String `tfsdk:"ciphersuites"`
	Cookie          types.String `tfsdk:"cookie"`
	Crt             types.String `tfsdk:"crt"`
	Downinter       types.Int64  `tfsdk:"downinter"`
	ErrorLimit      types.Int64  `tfsdk:"error_limit"`
	Fall            types.Int64  `tfsdk:"fall"`
	Fastinter       types.Int64  `tfsdk:"fastinter"`
	ForceSslv3      types.String `tfsdk:"force_sslv3"`
	ForceTlsv10     types.String `tfsdk:"force_tlsv10"`
	ForceTlsv11     types.String `tfsdk:"force_tlsv11"`
	ForceTlsv12     types.String `tfsdk:"force_tlsv12"`
	ForceTlsv13     types.String `tfsdk:"force_tlsv13"`
	ForceStrictSni  types.String `tfsdk:"force_strict_sni"`
	HealthCheckPort types.Int64  `tfsdk:"health_check_port"`
	InitAddr        types.String `tfsdk:"init_addr"`
	Inter           types.Int64  `tfsdk:"inter"`
	Maintenance     types.String `tfsdk:"maintenance"`
	Maxconn         types.Int64  `tfsdk:"maxconn"`
	Maxqueue        types.Int64  `tfsdk:"maxqueue"`
	Minconn         types.Int64  `tfsdk:"minconn"`
	NoSslv3         types.String `tfsdk:"no_sslv3"`
	NoTlsv10        types.String `tfsdk:"no_tlsv10"`
	NoTlsv11        types.String `tfsdk:"no_tlsv11"`
	NoTlsv12        types.String `tfsdk:"no_tlsv12"`
	NoTlsv13        types.String `tfsdk:"no_tlsv13"`
	// New v3 fields (non-deprecated)
	Sslv3            types.String `tfsdk:"sslv3"`
	Tlsv10           types.String `tfsdk:"tlsv10"`
	Tlsv11           types.String `tfsdk:"tlsv11"`
	Tlsv12           types.String `tfsdk:"tlsv12"`
	Tlsv13           types.String `tfsdk:"tlsv13"`
	OnError          types.String `tfsdk:"onerror"`
	OnMarkedDown     types.String `tfsdk:"onmarkeddown"`
	OnMarkedUp       types.String `tfsdk:"onmarkedup"`
	PoolLowConn      types.Int64  `tfsdk:"pool_low_conn"`
	PoolMaxConn      types.Int64  `tfsdk:"pool_max_conn"`
	PoolPurgeDelay   types.Int64  `tfsdk:"pool_purge_delay"`
	Proto            types.String `tfsdk:"proto"`
	ProxyV2Options   types.List   `tfsdk:"proxy_v2_options"`
	Redir            types.String `tfsdk:"redir"`
	Rise             types.Int64  `tfsdk:"rise"`
	SendProxy        types.String `tfsdk:"send_proxy"`
	SendProxyV2      types.String `tfsdk:"send_proxy_v2"`
	SendProxyV2Ssl   types.String `tfsdk:"send_proxy_v2_ssl"`
	SendProxyV2SslCn types.String `tfsdk:"send_proxy_v2_ssl_cn"`
	Slowstart        types.Int64  `tfsdk:"slowstart"`
	Sni              types.String `tfsdk:"sni"`
	Source           types.String `tfsdk:"source"`
	Ssl              types.String `tfsdk:"ssl"`
	SslCafile        types.String `tfsdk:"ssl_cafile"`
	SslCertificate   types.String `tfsdk:"ssl_certificate"`
	SslMaxVer        types.String `tfsdk:"ssl_max_ver"`
	SslMinVer        types.String `tfsdk:"ssl_min_ver"`
	SslReuse         types.String `tfsdk:"ssl_reuse"`
	Stick            types.String `tfsdk:"stick"`
	Tfo              types.String `tfsdk:"tfo"`
	TlsTickets       types.String `tfsdk:"tls_tickets"`
	Track            types.String `tfsdk:"track"`
	Verify           types.String `tfsdk:"verify"`
	Weight           types.Int64  `tfsdk:"weight"`
	Disabled         types.Bool   `tfsdk:"disabled"`
	LogProto         types.String `tfsdk:"log_proto"`
	Observe          types.String `tfsdk:"observe"`
	VerifyHost       types.String `tfsdk:"verifyhost"`
	HttpchkParams    types.Object `tfsdk:"httpchk_params"`
}

// Metadata returns the resource type name.
func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

// Schema defines the schema for the resource.
func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the server. It must be unique and cannot be changed.",
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "The port of the server. Constraints: Min 1┃Max 65535",
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "The address of the server. Pattern: ^[^\\s]+$",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the parent object",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the parent object. Allowed: backend|ring|peers",
			},
			"agent_addr": schema.StringAttribute{
				Optional:    true,
				Description: "The agent address of the server.",
			},
			"agent_check": schema.StringAttribute{
				Optional:    true,
				Description: "The agent check of the server.",
			},
			"agent_inter": schema.Int64Attribute{
				Optional:    true,
				Description: "The agent inter of the server.",
			},
			"agent_port": schema.Int64Attribute{
				Optional:    true,
				Description: "The agent port of the server.",
			},
			"agent_send": schema.StringAttribute{
				Optional:    true,
				Description: "The agent send of the server.",
			},
			"allow_0rtt": schema.BoolAttribute{
				Optional:    true,
				Description: "The allow 0rtt of the server.",
			},
			"alpn": schema.StringAttribute{
				Optional:    true,
				Description: "The alpn of the server.",
			},
			"backup": schema.StringAttribute{
				Optional:    true,
				Description: "The backup of the server.",
			},
			"check": schema.StringAttribute{
				Optional:    true,
				Description: "To enable health check for the server. Allowed: enabled|disabled",
			},
			"check_alpn": schema.StringAttribute{
				Optional:    true,
				Description: "The check alpn of the server.",
			},
			"check_sni": schema.StringAttribute{
				Optional:    true,
				Description: "The check sni of the server.",
			},
			"check_ssl": schema.StringAttribute{
				Optional:    true,
				Description: "To enable health check ssl if different port is used. Allowed: enabled|disabled",
			},
			"check_via_socks4": schema.StringAttribute{
				Optional:    true,
				Description: "The check via socks4 of the server.",
			},
			"ciphers": schema.StringAttribute{
				Optional:    true,
				Description: "ciphers to support",
			},
			"ciphersuites": schema.StringAttribute{
				Optional:    true,
				Description: "ciphersuites to support",
			},
			"cookie": schema.StringAttribute{
				Optional:    true,
				Description: "The cookie of the server.",
			},
			"crt": schema.StringAttribute{
				Optional:    true,
				Description: "The crt of the server.",
			},
			"downinter": schema.Int64Attribute{
				Optional:    true,
				Description: "The downinter of the server.",
			},
			"error_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "The error limit of the server.",
			},
			"fall": schema.Int64Attribute{
				Optional:    true,
				Description: "The fall value states that a server will be considered as failed after consecutive unsuccessful health checks.",
			},
			"fastinter": schema.Int64Attribute{
				Optional:    true,
				Description: "The fastinter of the server.",
			},
			"force_sslv3": schema.StringAttribute{
				Optional:    true,
				Description: "State of SSLv3 protocol support for the SSL. Allowed: enabled┃disabled. DEPRECATED: Use 'sslv3' field instead in Data Plane API v3",
			},
			"force_tlsv10": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.0 protocol support for the SSL. Allowed: enabled┃disabled. DEPRECATED: Use 'tlsv10' field instead in Data Plane API v3",
			},
			"force_tlsv11": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.1 protocol. Allowed: enabled┃disabled. DEPRECATED: Use 'tlsv11' field instead in Data Plane API v3",
			},
			"force_tlsv12": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.2 protocol. Allowed: enabled┃disabled. DEPRECATED: Use 'tlsv12' field instead in Data Plane API v3",
			},
			"force_tlsv13": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.3 protocol. Allowed: enabled┃disabled. DEPRECATED: Use 'tlsv13' field instead in Data Plane API v3",
			},
			"force_strict_sni": schema.StringAttribute{
				Optional:    true,
				Description: "Force strict SNI. DEPRECATED: Use 'strict_sni' field instead in Data Plane API v3. Allowed: enabled|disabled",
			},
			"health_check_port": schema.Int64Attribute{
				Optional:    true,
				Description: "The health check port of the server. Constraints: Min 1┃Max 65535",
			},
			"init_addr": schema.StringAttribute{
				Optional:    true,
				Description: "The init addr of the server.",
			},
			"inter": schema.Int64Attribute{
				Optional:    true,
				Description: "The inter value is the time interval in milliseconds between two consecutive health checks.",
			},
			"maintenance": schema.StringAttribute{
				Optional:    true,
				Description: "The maintenance of the server.",
			},
			"maxconn": schema.Int64Attribute{
				Optional:    true,
				Description: "The maxconn of the server.",
			},
			"maxqueue": schema.Int64Attribute{
				Optional:    true,
				Description: "The maxqueue of the server.",
			},
			"minconn": schema.Int64Attribute{
				Optional:    true,
				Description: "The minconn of the server.",
			},
			"no_sslv3": schema.StringAttribute{
				Optional:    true,
				Description: "The no sslv3 of the server. DEPRECATED: Use 'sslv3' field instead in Data Plane API v3",
			},
			"no_tlsv10": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv10 of the server. DEPRECATED: Use 'tlsv10' field instead in Data Plane API v3",
			},
			"no_tlsv11": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv11 of the server. DEPRECATED: Use 'tlsv11' field instead in Data Plane API v3",
			},
			"no_tlsv12": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv12 of the server. DEPRECATED: Use 'tlsv12' field instead in Data Plane API v3",
			},
			"no_tlsv13": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv13 of the server. DEPRECATED: Use 'tlsv13' field instead in Data Plane API v3",
			},
			// New v3 fields (non-deprecated)
			"sslv3": schema.StringAttribute{
				Optional:    true,
				Description: "Enable SSLv3 protocol support (v3 API, replaces no_sslv3). Allowed: enabled|disabled",
			},
			"tlsv10": schema.StringAttribute{
				Optional:    true,
				Description: "Enable TLSv1.0 protocol support (v3 API, replaces no_tlsv10). Allowed: enabled|disabled",
			},
			"tlsv11": schema.StringAttribute{
				Optional:    true,
				Description: "Enable TLSv1.1 protocol support (v3 API, replaces no_tlsv11). Allowed: enabled|disabled",
			},
			"tlsv12": schema.StringAttribute{
				Optional:    true,
				Description: "Enable TLSv1.2 protocol support (v3 API, replaces no_tlsv12). Allowed: enabled|disabled",
			},
			"tlsv13": schema.StringAttribute{
				Optional:    true,
				Description: "Enable TLSv1.3 protocol support (v3 API, replaces no_tlsv13). Allowed: enabled|disabled",
			},
			"onerror": schema.StringAttribute{
				Optional:    true,
				Description: "The on error of the server.",
			},
			"onmarkeddown": schema.StringAttribute{
				Optional:    true,
				Description: "The on marked down of the server.",
			},
			"onmarkedup": schema.StringAttribute{
				Optional:    true,
				Description: "The on marked up of the server.",
			},
			"pool_low_conn": schema.Int64Attribute{
				Optional:    true,
				Description: "The pool low conn of the server.",
			},
			"pool_max_conn": schema.Int64Attribute{
				Optional:    true,
				Description: "The pool max conn of the server.",
			},
			"pool_purge_delay": schema.Int64Attribute{
				Optional:    true,
				Description: "The pool purge delay of the server.",
			},
			"proto": schema.StringAttribute{
				Optional:    true,
				Description: "The proto of the server.",
			},
			"proxy_v2_options": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "The proxy v2 options of the server.",
			},
			"redir": schema.StringAttribute{
				Optional:    true,
				Description: "The redir of the server.",
			},
			"rise": schema.Int64Attribute{
				Optional:    true,
				Description: "The rise value states that a server will be considered as operational after consecutive successful health checks.",
			},
			"send_proxy": schema.StringAttribute{
				Optional:    true,
				Description: "To send a Proxy Protocol header to the backend server. Allowed: enabled|disabled",
			},
			"send_proxy_v2": schema.StringAttribute{
				Optional:    true,
				Description: "The send proxy v2 of the server.",
			},
			"send_proxy_v2_ssl": schema.StringAttribute{
				Optional:    true,
				Description: "The send proxy v2 ssl of the server.",
			},
			"send_proxy_v2_ssl_cn": schema.StringAttribute{
				Optional:    true,
				Description: "The send proxy v2 ssl cn of the server.",
			},
			"slowstart": schema.Int64Attribute{
				Optional:    true,
				Description: "The slowstart of the server.",
			},
			"sni": schema.StringAttribute{
				Optional:    true,
				Description: "The sni of the server.",
			},
			"source": schema.StringAttribute{
				Optional:    true,
				Description: "The source of the server.",
			},
			"ssl": schema.StringAttribute{
				Optional:    true,
				Description: "Enables ssl",
			},
			"ssl_cafile": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl certificate ca file. Pattern: ^[^\\s]+$",
			},
			"ssl_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl certificate. Pattern: ^[^\\s]+$",
			},
			"ssl_max_ver": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl max version. Allowed: SSLv3┃TLSv1.0┃TLSv1.1┃TLSv1.2┃TLSv1.3",
			},
			"ssl_min_ver": schema.StringAttribute{
				Optional:    true,
				Description: "The ssl min version. Allowed: SSLv3┃TLSv1.0┃TLSv1.1┃TLSv1.2┃TLSv1.3",
			},
			"ssl_reuse": schema.StringAttribute{
				Optional:    true,
				Description: "Reuse ssl existion connection. Allowed: enabled┃disabled",
			},
			"stick": schema.StringAttribute{
				Optional:    true,
				Description: "The stick of the server.",
			},
			"tfo": schema.StringAttribute{
				Optional:    true,
				Description: "The tfo of the server.",
			},
			"tls_tickets": schema.StringAttribute{
				Optional:    true,
				Description: "The tls tickets of the server.",
			},
			"track": schema.StringAttribute{
				Optional:    true,
				Description: "The track of the server.",
			},
			"verify": schema.StringAttribute{
				Optional:    true,
				Description: "The certificate verification for backend servers. Allowed: none┃required",
			},
			"weight": schema.Int64Attribute{
				Optional:    true,
				Description: "The weight of the server",
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "The disabled state of the server",
			},
			"log_proto": schema.StringAttribute{
				Optional:    true,
				Description: "The log protocol of the server",
			},
			"observe": schema.StringAttribute{
				Optional:    true,
				Description: "The observe mode of the server",
			},

			"verifyhost": schema.StringAttribute{
				Optional:    true,
				Description: "The verify host configuration of the server",
			},
		},
		Blocks: map[string]schema.Block{
			"httpchk_params": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Optional:    true,
						Description: "The HTTP method for health checks",
					},
					"uri": schema.StringAttribute{
						Optional:    true,
						Description: "The URI for health checks",
					},
					"version": schema.StringAttribute{
						Optional:    true,
						Description: "The HTTP version for health checks",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create a new resource.
func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverStandaloneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	var proxyV2Options []string
	if !plan.ProxyV2Options.IsNull() {
		diags := plan.ProxyV2Options.ElementsAs(ctx, &proxyV2Options, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var httpchkParams *HttpchkParams
	if !plan.HttpchkParams.IsNull() {
		var httpchkParamsModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		diags := plan.HttpchkParams.As(ctx, &httpchkParamsModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		httpchkParams = &HttpchkParams{
			Method:  httpchkParamsModel.Method.ValueString(),
			Uri:     httpchkParamsModel.Uri.ValueString(),
			Version: httpchkParamsModel.Version.ValueString(),
		}
	}

	payload := &ServerPayload{
		Name:             plan.Name.ValueString(),
		Address:          plan.Address.ValueString(),
		Port:             plan.Port.ValueInt64(),
		AgentAddr:        plan.AgentAddr.ValueString(),
		AgentCheck:       plan.AgentCheck.ValueString(),
		AgentInter:       plan.AgentInter.ValueInt64(),
		AgentPort:        plan.AgentPort.ValueInt64(),
		AgentSend:        plan.AgentSend.ValueString(),
		Allow0rtt:        plan.Allow0rtt.ValueBool(),
		Alpn:             plan.Alpn.ValueString(),
		Check:            plan.Check.ValueString(),
		CheckAlpn:        plan.CheckAlpn.ValueString(),
		CheckSni:         plan.CheckSni.ValueString(),
		CheckSsl:         plan.CheckSsl.ValueString(),
		CheckViaSocks4:   plan.CheckViaSocks4.ValueString(),
		Ciphers:          plan.Ciphers.ValueString(),
		Ciphersuites:     plan.Ciphersuites.ValueString(),
		Cookie:           plan.Cookie.ValueString(),
		Crt:              plan.Crt.ValueString(),
		Downinter:        plan.Downinter.ValueInt64(),
		ErrorLimit:       plan.ErrorLimit.ValueInt64(),
		Fall:             plan.Fall.ValueInt64(),
		Fastinter:        plan.Fastinter.ValueInt64(),
		ForceSslv3:       plan.ForceSslv3.ValueString(),
		ForceTlsv10:      plan.ForceTlsv10.ValueString(),
		ForceTlsv11:      plan.ForceTlsv11.ValueString(),
		ForceTlsv12:      plan.ForceTlsv12.ValueString(),
		ForceTlsv13:      plan.ForceTlsv13.ValueString(),
		HealthCheckPort:  plan.HealthCheckPort.ValueInt64(),
		InitAddr:         plan.InitAddr.ValueString(),
		Inter:            plan.Inter.ValueInt64(),
		Maintenance:      plan.Maintenance.ValueString(),
		Maxconn:          plan.Maxconn.ValueInt64(),
		Maxqueue:         plan.Maxqueue.ValueInt64(),
		Minconn:          plan.Minconn.ValueInt64(),
		NoSslv3:          plan.NoSslv3.ValueString(),
		NoTlsv10:         plan.NoTlsv10.ValueString(),
		NoTlsv11:         plan.NoTlsv11.ValueString(),
		NoTlsv12:         plan.NoTlsv12.ValueString(),
		NoTlsv13:         plan.NoTlsv13.ValueString(),
		OnError:          plan.OnError.ValueString(),
		OnMarkedDown:     plan.OnMarkedDown.ValueString(),
		OnMarkedUp:       plan.OnMarkedUp.ValueString(),
		PoolLowConn:      plan.PoolLowConn.ValueInt64(),
		PoolMaxConn:      plan.PoolMaxConn.ValueInt64(),
		PoolPurgeDelay:   plan.PoolPurgeDelay.ValueInt64(),
		Proto:            plan.Proto.ValueString(),
		ProxyV2Options:   proxyV2Options,
		Redir:            plan.Redir.ValueString(),
		Rise:             plan.Rise.ValueInt64(),
		SendProxy:        plan.SendProxy.ValueString(),
		SendProxyV2:      plan.SendProxyV2.ValueString(),
		SendProxyV2Ssl:   plan.SendProxyV2Ssl.ValueString(),
		SendProxyV2SslCn: plan.SendProxyV2SslCn.ValueString(),
		Slowstart:        plan.Slowstart.ValueInt64(),
		Sni:              plan.Sni.ValueString(),
		Source:           plan.Source.ValueString(),
		Ssl:              plan.Ssl.ValueString(),
		SslCafile:        plan.SslCafile.ValueString(),
		SslCertificate:   plan.SslCertificate.ValueString(),
		SslMaxVer:        plan.SslMaxVer.ValueString(),
		SslMinVer:        plan.SslMinVer.ValueString(),
		SslReuse:         plan.SslReuse.ValueString(),
		Stick:            plan.Stick.ValueString(),
		Tfo:              plan.Tfo.ValueString(),
		TlsTickets:       plan.TlsTickets.ValueString(),
		Track:            plan.Track.ValueString(),
		Verify:           plan.Verify.ValueString(),
		Weight:           plan.Weight.ValueInt64(),
		Disabled:         plan.Disabled.ValueBool(),
		LogProto:         plan.LogProto.ValueString(),
		Observe:          plan.Observe.ValueString(),
		VerifyHost:       plan.VerifyHost.ValueString(),
		HttpchkParams:    httpchkParams,
	}

	// Only add backup field if it's explicitly set
	if !plan.Backup.IsNull() && plan.Backup.ValueString() != "" {
		payload.Backup = plan.Backup.ValueString()
	}

	// Use the old transaction method which has built-in retry logic
	err := r.client.CreateServer(ctx, plan.ParentType.ValueString(), plan.ParentName.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating server",
			"Could not create server, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverStandaloneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	server, err := r.client.ReadServer(ctx, state.Name.ValueString(), state.ParentType.ValueString(), state.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading server",
			"Could not read server, unexpected error: "+err.Error(),
		)
		return
	}

	if server == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(server.Name)
	state.Address = types.StringValue(server.Address)
	state.Port = types.Int64Value(server.Port)
	// Only set agent fields if they have meaningful values (not empty)
	if server.AgentAddr != "" {
		state.AgentAddr = types.StringValue(server.AgentAddr)
	} else {
		state.AgentAddr = types.StringNull()
	}
	if server.AgentCheck != "" {
		state.AgentCheck = types.StringValue(server.AgentCheck)
	} else {
		state.AgentCheck = types.StringNull()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.AgentInter > 0 {
		state.AgentInter = types.Int64Value(server.AgentInter)
	} else {
		state.AgentInter = types.Int64Null()
	}
	if server.AgentPort > 0 {
		state.AgentPort = types.Int64Value(server.AgentPort)
	} else {
		state.AgentPort = types.Int64Null()
	}
	// Only set string fields if they have meaningful values (not empty)
	if server.AgentSend != "" {
		state.AgentSend = types.StringValue(server.AgentSend)
	} else {
		state.AgentSend = types.StringNull()
	}
	// Only set boolean fields if they have meaningful values (not false)
	if server.Allow0rtt {
		state.Allow0rtt = types.BoolValue(true)
	} else {
		state.Allow0rtt = types.BoolNull()
	}
	if server.Alpn != "" {
		state.Alpn = types.StringValue(server.Alpn)
	} else {
		state.Alpn = types.StringNull()
	}
	if server.Backup != "" {
		state.Backup = types.StringValue(server.Backup)
	} else {
		state.Backup = types.StringNull()
	}
	if server.Check != "" {
		state.Check = types.StringValue(server.Check)
	} else {
		state.Check = types.StringNull()
	}
	if server.CheckAlpn != "" {
		state.CheckAlpn = types.StringValue(server.CheckAlpn)
	} else {
		state.CheckAlpn = types.StringNull()
	}
	if server.CheckSni != "" {
		state.CheckSni = types.StringValue(server.CheckSni)
	} else {
		state.CheckSni = types.StringNull()
	}
	if server.CheckSsl != "" {
		state.CheckSsl = types.StringValue(server.CheckSsl)
	} else {
		state.CheckSsl = types.StringNull()
	}
	if server.CheckViaSocks4 != "" {
		state.CheckViaSocks4 = types.StringValue(server.CheckViaSocks4)
	} else {
		state.CheckViaSocks4 = types.StringNull()
	}
	if server.Ciphers != "" {
		state.Ciphers = types.StringValue(server.Ciphers)
	} else {
		state.Ciphers = types.StringNull()
	}
	if server.Ciphersuites != "" {
		state.Ciphersuites = types.StringValue(server.Ciphersuites)
	} else {
		state.Ciphersuites = types.StringNull()
	}
	if server.Cookie != "" {
		state.Cookie = types.StringValue(server.Cookie)
	} else {
		state.Cookie = types.StringNull()
	}
	if server.Crt != "" {
		state.Crt = types.StringValue(server.Crt)
	} else {
		state.Crt = types.StringNull()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.Downinter > 0 {
		state.Downinter = types.Int64Value(server.Downinter)
	} else {
		state.Downinter = types.Int64Null()
	}
	if server.ErrorLimit > 0 {
		state.ErrorLimit = types.Int64Value(server.ErrorLimit)
	} else {
		state.ErrorLimit = types.Int64Null()
	}
	if server.Fall > 0 {
		state.Fall = types.Int64Value(server.Fall)
	} else {
		state.Fall = types.Int64Null()
	}
	if server.Fastinter > 0 {
		state.Fastinter = types.Int64Value(server.Fastinter)
	} else {
		state.Fastinter = types.Int64Null()
	}
	// Only set string fields if they have meaningful values (not empty)
	if server.ForceSslv3 != "" {
		state.ForceSslv3 = types.StringValue(server.ForceSslv3)
	} else {
		state.ForceSslv3 = types.StringNull()
	}
	if server.ForceTlsv10 != "" {
		state.ForceTlsv10 = types.StringValue(server.ForceTlsv10)
	} else {
		state.ForceTlsv10 = types.StringNull()
	}
	if server.ForceTlsv11 != "" {
		state.ForceTlsv11 = types.StringValue(server.ForceTlsv11)
	} else {
		state.ForceTlsv11 = types.StringNull()
	}
	if server.ForceTlsv12 != "" {
		state.ForceTlsv12 = types.StringValue(server.ForceTlsv12)
	} else {
		state.ForceTlsv12 = types.StringNull()
	}
	if server.ForceTlsv13 != "" {
		state.ForceTlsv13 = types.StringValue(server.ForceTlsv13)
	} else {
		state.ForceTlsv13 = types.StringNull()
	}
	if server.HealthCheckPort > 0 {
		state.HealthCheckPort = types.Int64Value(server.HealthCheckPort)
	} else {
		state.HealthCheckPort = types.Int64Null()
	}
	if server.InitAddr != "" {
		state.InitAddr = types.StringValue(server.InitAddr)
	} else {
		state.InitAddr = types.StringNull()
	}
	if server.Inter > 0 {
		state.Inter = types.Int64Value(server.Inter)
	} else {
		state.Inter = types.Int64Null()
	}
	if server.Maintenance != "" {
		state.Maintenance = types.StringValue(server.Maintenance)
	} else {
		state.Maintenance = types.StringNull()
	}
	if server.Maxconn > 0 {
		state.Maxconn = types.Int64Value(server.Maxconn)
	} else {
		state.Maxconn = types.Int64Null()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.Maxqueue > 0 {
		state.Maxqueue = types.Int64Value(server.Maxqueue)
	} else {
		state.Maxqueue = types.Int64Null()
	}
	if server.Minconn > 0 {
		state.Minconn = types.Int64Value(server.Minconn)
	} else {
		state.Minconn = types.Int64Null()
	}
	// Only set string fields if they have meaningful values (not empty)
	if server.NoSslv3 != "" {
		state.NoSslv3 = types.StringValue(server.NoSslv3)
	} else {
		state.NoSslv3 = types.StringNull()
	}
	if server.NoTlsv10 != "" {
		state.NoTlsv10 = types.StringValue(server.NoTlsv10)
	} else {
		state.NoTlsv10 = types.StringNull()
	}
	if server.NoTlsv11 != "" {
		state.NoTlsv11 = types.StringValue(server.NoTlsv11)
	} else {
		state.NoTlsv11 = types.StringNull()
	}
	if server.NoTlsv12 != "" {
		state.NoTlsv12 = types.StringValue(server.NoTlsv12)
	} else {
		state.NoTlsv12 = types.StringNull()
	}
	if server.NoTlsv13 != "" {
		state.NoTlsv13 = types.StringValue(server.NoTlsv13)
	} else {
		state.NoTlsv13 = types.StringNull()
	}
	if server.OnError != "" {
		state.OnError = types.StringValue(server.OnError)
	} else {
		state.OnError = types.StringNull()
	}
	if server.OnMarkedDown != "" {
		state.OnMarkedDown = types.StringValue(server.OnMarkedDown)
	} else {
		state.OnMarkedDown = types.StringNull()
	}
	if server.OnMarkedUp != "" {
		state.OnMarkedUp = types.StringValue(server.OnMarkedUp)
	} else {
		state.OnMarkedUp = types.StringNull()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.PoolLowConn > 0 {
		state.PoolLowConn = types.Int64Value(server.PoolLowConn)
	} else {
		state.PoolLowConn = types.Int64Null()
	}
	if server.PoolMaxConn > 0 {
		state.PoolMaxConn = types.Int64Value(server.PoolMaxConn)
	} else {
		state.PoolMaxConn = types.Int64Null()
	}
	if server.PoolPurgeDelay > 0 {
		state.PoolPurgeDelay = types.Int64Value(server.PoolPurgeDelay)
	} else {
		state.PoolPurgeDelay = types.Int64Null()
	}
	if server.Proto != "" {
		state.Proto = types.StringValue(server.Proto)
	} else {
		state.Proto = types.StringNull()
	}
	state.ProxyV2Options, diags = types.ListValueFrom(ctx, types.StringType, server.ProxyV2Options)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only set Redir field if it has meaningful values (not empty)
	if server.Redir != "" {
		state.Redir = types.StringValue(server.Redir)
	} else {
		state.Redir = types.StringNull()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.Rise > 0 {
		state.Rise = types.Int64Value(server.Rise)
	} else {
		state.Rise = types.Int64Null()
	}
	// Only set SendProxy fields if they have meaningful values (not empty)
	if server.SendProxy != "" {
		state.SendProxy = types.StringValue(server.SendProxy)
	} else {
		state.SendProxy = types.StringNull()
	}
	if server.SendProxyV2 != "" {
		state.SendProxyV2 = types.StringValue(server.SendProxyV2)
	} else {
		state.SendProxyV2 = types.StringNull()
	}
	if server.SendProxyV2Ssl != "" {
		state.SendProxyV2Ssl = types.StringValue(server.SendProxyV2Ssl)
	} else {
		state.SendProxyV2Ssl = types.StringNull()
	}
	if server.SendProxyV2SslCn != "" {
		state.SendProxyV2SslCn = types.StringValue(server.SendProxyV2SslCn)
	} else {
		state.SendProxyV2SslCn = types.StringNull()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.Slowstart > 0 {
		state.Slowstart = types.Int64Value(server.Slowstart)
	} else {
		state.Slowstart = types.Int64Null()
	}
	// Only set SSL and other string fields if they have meaningful values (not empty)
	if server.Sni != "" {
		state.Sni = types.StringValue(server.Sni)
	} else {
		state.Sni = types.StringNull()
	}
	if server.Source != "" {
		state.Source = types.StringValue(server.Source)
	} else {
		state.Source = types.StringNull()
	}
	if server.Ssl != "" {
		state.Ssl = types.StringValue(server.Ssl)
	} else {
		state.Ssl = types.StringNull()
	}
	if server.SslCafile != "" {
		state.SslCafile = types.StringValue(server.SslCafile)
	} else {
		state.SslCafile = types.StringNull()
	}
	if server.SslCertificate != "" {
		state.SslCertificate = types.StringValue(server.SslCertificate)
	} else {
		state.SslCertificate = types.StringNull()
	}
	if server.SslMaxVer != "" {
		state.SslMaxVer = types.StringValue(server.SslMaxVer)
	} else {
		state.SslMaxVer = types.StringNull()
	}
	if server.SslMinVer != "" {
		state.SslMinVer = types.StringValue(server.SslMinVer)
	} else {
		state.SslMinVer = types.StringNull()
	}
	if server.SslReuse != "" {
		state.SslReuse = types.StringValue(server.SslReuse)
	} else {
		state.SslReuse = types.StringNull()
	}
	if server.Stick != "" {
		state.Stick = types.StringValue(server.Stick)
	} else {
		state.Stick = types.StringNull()
	}
	if server.Tfo != "" {
		state.Tfo = types.StringValue(server.Tfo)
	} else {
		state.Tfo = types.StringNull()
	}
	if server.TlsTickets != "" {
		state.TlsTickets = types.StringValue(server.TlsTickets)
	} else {
		state.TlsTickets = types.StringNull()
	}
	if server.Track != "" {
		state.Track = types.StringValue(server.Track)
	} else {
		state.Track = types.StringNull()
	}
	if server.Verify != "" {
		state.Verify = types.StringValue(server.Verify)
	} else {
		state.Verify = types.StringNull()
	}
	// Only set fields if they have meaningful values (not zero)
	if server.Weight > 0 {
		state.Weight = types.Int64Value(server.Weight)
	} else {
		state.Weight = types.Int64Null()
	}

	// Handle new fields
	if server.Disabled {
		state.Disabled = types.BoolValue(server.Disabled)
	} else {
		state.Disabled = types.BoolNull()
	}
	if server.LogProto != "" {
		state.LogProto = types.StringValue(server.LogProto)
	} else {
		state.LogProto = types.StringNull()
	}
	if server.Observe != "" {
		state.Observe = types.StringValue(server.Observe)
	} else {
		state.Observe = types.StringNull()
	}
	if server.VerifyHost != "" {
		state.VerifyHost = types.StringValue(server.VerifyHost)
	} else {
		state.VerifyHost = types.StringNull()
	}

	// Handle httpchk_params block
	if server.HttpchkParams != nil {
		var httpchkParamsModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		httpchkParamsModel.Method = types.StringValue(server.HttpchkParams.Method)
		httpchkParamsModel.Uri = types.StringValue(server.HttpchkParams.Uri)
		httpchkParamsModel.Version = types.StringValue(server.HttpchkParams.Version)

		httpchkParamsObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"method":  types.StringType,
			"uri":     types.StringType,
			"version": types.StringType,
		}, map[string]attr.Value{
			"method":  httpchkParamsModel.Method,
			"uri":     httpchkParamsModel.Uri,
			"version": httpchkParamsModel.Version,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.HttpchkParams = httpchkParamsObj
	} else {
		state.HttpchkParams = types.ObjectNull(map[string]attr.Type{
			"method":  types.StringType,
			"uri":     types.StringType,
			"version": types.StringType,
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverStandaloneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Handle proxy_v2_options for update
	var proxyV2Options []string
	if !plan.ProxyV2Options.IsNull() {
		diags := plan.ProxyV2Options.ElementsAs(ctx, &proxyV2Options, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Handle httpchk_params for update
	var httpchkParams *HttpchkParams
	if !plan.HttpchkParams.IsNull() {
		var httpchkParamsModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		diags := plan.HttpchkParams.As(ctx, &httpchkParamsModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		httpchkParams = &HttpchkParams{
			Method:  httpchkParamsModel.Method.ValueString(),
			Uri:     httpchkParamsModel.Uri.ValueString(),
			Version: httpchkParamsModel.Version.ValueString(),
		}
	}

	payload := &ServerPayload{
		Name:       plan.Name.ValueString(),
		Address:    plan.Address.ValueString(),
		Port:       plan.Port.ValueInt64(),
		AgentAddr:  plan.AgentAddr.ValueString(),
		AgentCheck: plan.AgentCheck.ValueString(),
		AgentInter: plan.AgentInter.ValueInt64(),
		AgentPort:  plan.AgentPort.ValueInt64(),
		AgentSend:  plan.AgentSend.ValueString(),
		Allow0rtt:  plan.Allow0rtt.ValueBool(),
		Alpn:       plan.Alpn.ValueString(),

		Check:            plan.Check.ValueString(),
		CheckAlpn:        plan.CheckAlpn.ValueString(),
		CheckSni:         plan.CheckSni.ValueString(),
		CheckSsl:         plan.CheckSsl.ValueString(),
		CheckViaSocks4:   plan.CheckViaSocks4.ValueString(),
		Ciphers:          plan.Ciphers.ValueString(),
		Ciphersuites:     plan.Ciphersuites.ValueString(),
		Cookie:           plan.Cookie.ValueString(),
		Crt:              plan.Crt.ValueString(),
		Downinter:        plan.Downinter.ValueInt64(),
		ErrorLimit:       plan.ErrorLimit.ValueInt64(),
		Fall:             plan.Fall.ValueInt64(),
		Fastinter:        plan.Fastinter.ValueInt64(),
		ForceSslv3:       plan.ForceSslv3.ValueString(),
		ForceTlsv10:      plan.ForceTlsv10.ValueString(),
		ForceTlsv11:      plan.ForceTlsv11.ValueString(),
		ForceTlsv12:      plan.ForceTlsv12.ValueString(),
		ForceTlsv13:      plan.ForceTlsv13.ValueString(),
		ForceStrictSni:   plan.ForceStrictSni.ValueString(),
		HealthCheckPort:  plan.HealthCheckPort.ValueInt64(),
		InitAddr:         plan.InitAddr.ValueString(),
		Inter:            plan.Inter.ValueInt64(),
		Maintenance:      plan.Maintenance.ValueString(),
		Maxconn:          plan.Maxconn.ValueInt64(),
		Maxqueue:         plan.Maxqueue.ValueInt64(),
		Minconn:          plan.Minconn.ValueInt64(),
		NoSslv3:          plan.NoSslv3.ValueString(),
		NoTlsv10:         plan.NoTlsv10.ValueString(),
		NoTlsv11:         plan.NoTlsv11.ValueString(),
		NoTlsv12:         plan.NoTlsv12.ValueString(),
		NoTlsv13:         plan.NoTlsv13.ValueString(),
		Sslv3:            plan.Sslv3.ValueString(),
		Tlsv10:           plan.Tlsv10.ValueString(),
		Tlsv11:           plan.Tlsv11.ValueString(),
		Tlsv12:           plan.Tlsv12.ValueString(),
		Tlsv13:           plan.Tlsv13.ValueString(),
		OnError:          plan.OnError.ValueString(),
		OnMarkedDown:     plan.OnMarkedDown.ValueString(),
		OnMarkedUp:       plan.OnMarkedUp.ValueString(),
		PoolLowConn:      plan.PoolLowConn.ValueInt64(),
		PoolMaxConn:      plan.PoolMaxConn.ValueInt64(),
		PoolPurgeDelay:   plan.PoolPurgeDelay.ValueInt64(),
		Proto:            plan.Proto.ValueString(),
		ProxyV2Options:   proxyV2Options,
		Redir:            plan.Redir.ValueString(),
		Rise:             plan.Rise.ValueInt64(),
		SendProxy:        plan.SendProxy.ValueString(),
		SendProxyV2:      plan.SendProxyV2.ValueString(),
		SendProxyV2Ssl:   plan.SendProxyV2Ssl.ValueString(),
		SendProxyV2SslCn: plan.SendProxyV2SslCn.ValueString(),
		Slowstart:        plan.Slowstart.ValueInt64(),
		Sni:              plan.Sni.ValueString(),
		Source:           plan.Source.ValueString(),
		Ssl:              plan.Ssl.ValueString(),
		SslCafile:        plan.SslCafile.ValueString(),
		SslCertificate:   plan.SslCertificate.ValueString(),
		SslMaxVer:        plan.SslMaxVer.ValueString(),
		SslMinVer:        plan.SslMinVer.ValueString(),
		SslReuse:         plan.SslReuse.ValueString(),
		Stick:            plan.Stick.ValueString(),
		Tfo:              plan.Tfo.ValueString(),
		TlsTickets:       plan.TlsTickets.ValueString(),
		Track:            plan.Track.ValueString(),
		Verify:           plan.Verify.ValueString(),
		Weight:           plan.Weight.ValueInt64(),
		Disabled:         plan.Disabled.ValueBool(),
		LogProto:         plan.LogProto.ValueString(),
		Observe:          plan.Observe.ValueString(),
		VerifyHost:       plan.VerifyHost.ValueString(),
		HttpchkParams:    httpchkParams,
	}

	// Only add backup field if it's explicitly set
	if !plan.Backup.IsNull() && plan.Backup.ValueString() != "" {
		payload.Backup = plan.Backup.ValueString()
	}

	err := r.client.UpdateServer(ctx, plan.Name.ValueString(), plan.ParentType.ValueString(), plan.ParentName.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating server",
			"Could not update server, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverStandaloneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	err := r.client.DeleteServer(ctx, state.Name.ValueString(), state.ParentType.ValueString(), state.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting server",
			"Could not delete server, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
