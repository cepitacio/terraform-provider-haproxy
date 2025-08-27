package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Name             types.String `tfsdk:"name"`
	Port             types.Int64  `tfsdk:"port"`
	Address          types.String `tfsdk:"address"`
	ParentName       types.String `tfsdk:"parent_name"`
	ParentType       types.String `tfsdk:"parent_type"`
	AgentAddr        types.String `tfsdk:"agent_addr"`
	AgentCheck       types.String `tfsdk:"agent_check"`
	AgentInter       types.Int64  `tfsdk:"agent_inter"`
	AgentPort        types.Int64  `tfsdk:"agent_port"`
	AgentSend        types.String `tfsdk:"agent_send"`
	Allow0rtt        types.Bool   `tfsdk:"allow_0rtt"`
	Alpn             types.String `tfsdk:"alpn"`
	Backup           types.String `tfsdk:"backup"`
	Check            types.String `tfsdk:"check"`
	CheckAlpn        types.String `tfsdk:"check_alpn"`
	CheckSni         types.String `tfsdk:"check_sni"`
	CheckSsl         types.String `tfsdk:"check_ssl"`
	CheckViaSocks4   types.String `tfsdk:"check_via_socks4"`
	Ciphers          types.String `tfsdk:"ciphers"`
	Ciphersuites     types.String `tfsdk:"ciphersuites"`
	Cookie           types.String `tfsdk:"cookie"`
	Crt              types.String `tfsdk:"crt"`
	Downinter        types.Int64  `tfsdk:"downinter"`
	ErrorLimit       types.Int64  `tfsdk:"error_limit"`
	Fall             types.Int64  `tfsdk:"fall"`
	Fastinter        types.Int64  `tfsdk:"fastinter"`
	ForceSslv3       types.String `tfsdk:"force_sslv3"`
	ForceTlsv10      types.String `tfsdk:"force_tlsv10"`
	ForceTlsv11      types.String `tfsdk:"force_tlsv11"`
	ForceTlsv12      types.String `tfsdk:"force_tlsv12"`
	ForceTlsv13      types.String `tfsdk:"force_tlsv13"`
	HealthCheckPort  types.Int64  `tfsdk:"health_check_port"`
	InitAddr         types.String `tfsdk:"init_addr"`
	Inter            types.Int64  `tfsdk:"inter"`
	Maintenance      types.String `tfsdk:"maintenance"`
	Maxconn          types.Int64  `tfsdk:"maxconn"`
	Maxqueue         types.Int64  `tfsdk:"maxqueue"`
	Minconn          types.Int64  `tfsdk:"minconn"`
	NoSslv3          types.String `tfsdk:"no_sslv3"`
	NoTlsv10         types.String `tfsdk:"no_tlsv10"`
	NoTlsv11         types.String `tfsdk:"no_tlsv11"`
	NoTlsv12         types.String `tfsdk:"no_tlsv12"`
	NoTlsv13         types.String `tfsdk:"no_tlsv13"`
	OnError          types.String `tfsdk:"on_error"`
	OnMarkedDown     types.String `tfsdk:"on_marked_down"`
	OnMarkedUp       types.String `tfsdk:"on_marked_up"`
	PoolLowConn      types.Int64  `tfsdk:"pool_low_conn"`
	PoolMaxConn      types.Int64  `tfsdk:"pool_max_conn"`
	PoolPurgeDelay   types.Int64  `tfsdk:"pool_purge_delay"`
	Proto            types.String `tfsdk:"proto"`
	ProxyV2Options   types.List   `tfsdk:"proxy_v2_options"`
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
				Description: "State of SSLv3 protocol support for the SSL. Allowed: enabled┃disabled",
			},
			"force_tlsv10": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.0 protocol support for the SSL. Allowed: enabled┃disabled",
			},
			"force_tlsv11": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.1 protocol. Allowed: enabled┃disabled",
			},
			"force_tlsv12": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.2 protocol. Allowed: enabled┃disabled",
			},
			"force_tlsv13": schema.StringAttribute{
				Optional:    true,
				Description: "State of TLSv1.3 protocol. Allowed: enabled┃disabled",
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
				Description: "The no sslv3 of the server.",
			},
			"no_tlsv10": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv10 of the server.",
			},
			"no_tlsv11": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv11 of the server.",
			},
			"no_tlsv12": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv12 of the server.",
			},
			"no_tlsv13": schema.StringAttribute{
				Optional:    true,
				Description: "The no tlsv13 of the server.",
			},
			"on_error": schema.StringAttribute{
				Optional:    true,
				Description: "The on error of the server.",
			},
			"on_marked_down": schema.StringAttribute{
				Optional:    true,
				Description: "The on marked down of the server.",
			},
			"on_marked_up": schema.StringAttribute{
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
		Backup:           plan.Backup.ValueString(),
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
	}

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
	state.AgentAddr = types.StringValue(server.AgentAddr)
	state.AgentCheck = types.StringValue(server.AgentCheck)
	state.AgentInter = types.Int64Value(server.AgentInter)
	state.AgentPort = types.Int64Value(server.AgentPort)
	state.AgentSend = types.StringValue(server.AgentSend)
	state.Allow0rtt = types.BoolValue(server.Allow0rtt)
	state.Alpn = types.StringValue(server.Alpn)
	state.Backup = types.StringValue(server.Backup)
	state.Check = types.StringValue(server.Check)
	state.CheckAlpn = types.StringValue(server.CheckAlpn)
	state.CheckSni = types.StringValue(server.CheckSni)
	state.CheckSsl = types.StringValue(server.CheckSsl)
	state.CheckViaSocks4 = types.StringValue(server.CheckViaSocks4)
	state.Ciphers = types.StringValue(server.Ciphers)
	state.Ciphersuites = types.StringValue(server.Ciphersuites)
	state.Cookie = types.StringValue(server.Cookie)
	state.Crt = types.StringValue(server.Crt)
	state.Downinter = types.Int64Value(server.Downinter)
	state.ErrorLimit = types.Int64Value(server.ErrorLimit)
	state.Fall = types.Int64Value(server.Fall)
	state.Fastinter = types.Int64Value(server.Fastinter)
	state.ForceSslv3 = types.StringValue(server.ForceSslv3)
	state.ForceTlsv10 = types.StringValue(server.ForceTlsv10)
	state.ForceTlsv11 = types.StringValue(server.ForceTlsv11)
	state.ForceTlsv12 = types.StringValue(server.ForceTlsv12)
	state.ForceTlsv13 = types.StringValue(server.ForceTlsv13)
	state.HealthCheckPort = types.Int64Value(server.HealthCheckPort)
	state.InitAddr = types.StringValue(server.InitAddr)
	state.Inter = types.Int64Value(server.Inter)
	state.Maintenance = types.StringValue(server.Maintenance)
	state.Maxconn = types.Int64Value(server.Maxconn)
	state.Maxqueue = types.Int64Value(server.Maxqueue)
	state.Minconn = types.Int64Value(server.Minconn)
	state.NoSslv3 = types.StringValue(server.NoSslv3)
	state.NoTlsv10 = types.StringValue(server.NoTlsv10)
	state.NoTlsv11 = types.StringValue(server.NoTlsv11)
	state.NoTlsv12 = types.StringValue(server.NoTlsv12)
	state.NoTlsv13 = types.StringValue(server.NoTlsv13)
	state.OnError = types.StringValue(server.OnError)
	state.OnMarkedDown = types.StringValue(server.OnMarkedDown)
	state.OnMarkedUp = types.StringValue(server.OnMarkedUp)
	state.PoolLowConn = types.Int64Value(server.PoolLowConn)
	state.PoolMaxConn = types.Int64Value(server.PoolMaxConn)
	state.PoolPurgeDelay = types.Int64Value(server.PoolPurgeDelay)
	state.Proto = types.StringValue(server.Proto)
	state.ProxyV2Options, diags = types.ListValueFrom(ctx, types.StringType, server.ProxyV2Options)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Rise = types.Int64Value(server.Rise)
	state.SendProxy = types.StringValue(server.SendProxy)
	state.SendProxyV2 = types.StringValue(server.SendProxyV2)
	state.SendProxyV2Ssl = types.StringValue(server.SendProxyV2Ssl)
	state.SendProxyV2SslCn = types.StringValue(server.SendProxyV2SslCn)
	state.Slowstart = types.Int64Value(server.Slowstart)
	state.Sni = types.StringValue(server.Sni)
	state.Source = types.StringValue(server.Source)
	state.Ssl = types.StringValue(server.Ssl)
	state.SslCafile = types.StringValue(server.SslCafile)
	state.SslCertificate = types.StringValue(server.SslCertificate)
	state.SslMaxVer = types.StringValue(server.SslMaxVer)
	state.SslMinVer = types.StringValue(server.SslMinVer)
	state.SslReuse = types.StringValue(server.SslReuse)
	state.Stick = types.StringValue(server.Stick)
	state.Tfo = types.StringValue(server.Tfo)
	state.TlsTickets = types.StringValue(server.TlsTickets)
	state.Track = types.StringValue(server.Track)
	state.Verify = types.StringValue(server.Verify)
	state.Weight = types.Int64Value(server.Weight)

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

	payload := &ServerPayload{
		Name:            plan.Name.ValueString(),
		Address:         plan.Address.ValueString(),
		Port:            plan.Port.ValueInt64(),
		SendProxy:       plan.SendProxy.ValueString(),
		Check:           plan.Check.ValueString(),
		CheckSsl:        plan.CheckSsl.ValueString(),
		Inter:           plan.Inter.ValueInt64(),
		Rise:            plan.Rise.ValueInt64(),
		Fall:            plan.Fall.ValueInt64(),
		Ssl:             plan.Ssl.ValueString(),
		SslCafile:       plan.SslCafile.ValueString(),
		SslCertificate:  plan.SslCertificate.ValueString(),
		SslMaxVer:       plan.SslMaxVer.ValueString(),
		SslMinVer:       plan.SslMinVer.ValueString(),
		SslReuse:        plan.SslReuse.ValueString(),
		Verify:          plan.Verify.ValueString(),
		HealthCheckPort: plan.HealthCheckPort.ValueInt64(),
		Weight:          plan.Weight.ValueInt64(),
		Ciphers:         plan.Ciphers.ValueString(),
		Ciphersuites:    plan.Ciphersuites.ValueString(),
		ForceSslv3:      plan.ForceSslv3.ValueString(),
		ForceTlsv10:     plan.ForceTlsv10.ValueString(),
		ForceTlsv11:     plan.ForceTlsv11.ValueString(),
		ForceTlsv12:     plan.ForceTlsv12.ValueString(),
		ForceTlsv13:     plan.ForceTlsv13.ValueString(),
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
