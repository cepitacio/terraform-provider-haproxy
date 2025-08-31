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
	_ resource.Resource = &haproxyStackResource{}
)

// NewHaproxyStackResource is a helper function to simplify the provider implementation.
func NewHaproxyStackResource() resource.Resource {
	return &haproxyStackResource{}
}

// haproxyStackResource is the resource implementation.
type haproxyStackResource struct {
	stackManager *StackManager
}

// haproxyStackResourceModel maps the resource schema data.
type haproxyStackResourceModel struct {
	Name     types.String          `tfsdk:"name"`
	Backend  *haproxyBackendModel  `tfsdk:"backend"`
	Server   *haproxyServerModel   `tfsdk:"server"`
	Frontend *haproxyFrontendModel `tfsdk:"frontend"`
	Acls     []haproxyAclModel     `tfsdk:"acls"`
}

// haproxyBackendModel maps the backend block schema data.
type haproxyBackendModel struct {
	Name               types.String                   `tfsdk:"name"`
	Mode               types.String                   `tfsdk:"mode"`
	AdvCheck           types.String                   `tfsdk:"adv_check"`
	HttpConnectionMode types.String                   `tfsdk:"http_connection_mode"`
	ServerTimeout      types.Int64                    `tfsdk:"server_timeout"`
	CheckTimeout       types.Int64                    `tfsdk:"check_timeout"`
	ConnectTimeout     types.Int64                    `tfsdk:"connect_timeout"`
	QueueTimeout       types.Int64                    `tfsdk:"queue_timeout"`
	TunnelTimeout      types.Int64                    `tfsdk:"tunnel_timeout"`
	TarpitTimeout      types.Int64                    `tfsdk:"tarpit_timeout"`
	Checkcache         types.String                   `tfsdk:"checkcache"`
	Retries            types.Int64                    `tfsdk:"retries"`
	Balance            []haproxyBalanceModel          `tfsdk:"balance"`
	HttpchkParams      []haproxyHttpchkParamsModel    `tfsdk:"httpchk_params"`
	Forwardfor         []haproxyForwardforModel       `tfsdk:"forwardfor"`
	Httpcheck          []haproxyHttpcheckModel        `tfsdk:"httpcheck"`
	TcpCheck           []haproxyTcpCheckModel         `tfsdk:"tcp_check"`
	Acls               []haproxyAclModel              `tfsdk:"acls"`
	HttpRequestRule    []haproxyHttpRequestRuleModel  `tfsdk:"http_request_rule"`
	HttpResponseRule   []haproxyHttpResponseRuleModel `tfsdk:"http_response_rule"`
	TcpRequestRule     []haproxyTcpRequestRuleModel   `tfsdk:"tcp_request_rule"`
	TcpResponseRule    []haproxyTcpResponseRuleModel  `tfsdk:"tcp_response_rule"`
	DefaultServer      *haproxyDefaultServerModel     `tfsdk:"default_server"`
	StickTable         *haproxyStickTableModel        `tfsdk:"stick_table"`
	StickRule          []haproxyStickRuleModel        `tfsdk:"stick_rule"`
	StatsOptions       []haproxyStatsOptionsModel     `tfsdk:"stats_options"`
}

// haproxyDefaultServerModel maps the default_server block schema data.
type haproxyDefaultServerModel struct {
	Ssl            types.String `tfsdk:"ssl"`
	SslCafile      types.String `tfsdk:"ssl_cafile"`
	SslCertificate types.String `tfsdk:"ssl_certificate"`
	SslMaxVer      types.String `tfsdk:"ssl_max_ver"`
	SslMinVer      types.String `tfsdk:"ssl_min_ver"`
	SslReuse       types.String `tfsdk:"ssl_reuse"`
	Ciphers        types.String `tfsdk:"ciphers"`
	Ciphersuites   types.String `tfsdk:"ciphersuites"`
	Verify         types.String `tfsdk:"verify"`
	Sslv3          types.String `tfsdk:"sslv3"`
	Tlsv10         types.String `tfsdk:"tlsv10"`
	Tlsv11         types.String `tfsdk:"tlsv11"`
	Tlsv12         types.String `tfsdk:"tlsv12"`
	Tlsv13         types.String `tfsdk:"tlsv13"`
	NoSslv3        types.String `tfsdk:"no_sslv3"`
	NoTlsv10       types.String `tfsdk:"no_tlsv10"`
	NoTlsv11       types.String `tfsdk:"no_tlsv11"`
	NoTlsv12       types.String `tfsdk:"no_tlsv12"`
	NoTlsv13       types.String `tfsdk:"no_tlsv13"`
	ForceSslv3     types.String `tfsdk:"force_sslv3"`
	ForceTlsv10    types.String `tfsdk:"force_tlsv10"`
	ForceTlsv11    types.String `tfsdk:"force_tlsv11"`
	ForceTlsv12    types.String `tfsdk:"force_tlsv12"`
	ForceTlsv13    types.String `tfsdk:"force_tlsv13"`
	ForceStrictSni types.String `tfsdk:"force_strict_sni"`
}

// haproxyServerModel maps the server block schema data.
type haproxyServerModel struct {
	Name      types.String `tfsdk:"name"`
	Address   types.String `tfsdk:"address"`
	Port      types.Int64  `tfsdk:"port"`
	Check     types.String `tfsdk:"check"`
	Backup    types.String `tfsdk:"backup"`
	Maxconn   types.Int64  `tfsdk:"maxconn"`
	Weight    types.Int64  `tfsdk:"weight"`
	Rise      types.Int64  `tfsdk:"rise"`
	Fall      types.Int64  `tfsdk:"fall"`
	Inter     types.Int64  `tfsdk:"inter"`
	Fastinter types.Int64  `tfsdk:"fastinter"`
	Downinter types.Int64  `tfsdk:"downinter"`
	Ssl       types.String `tfsdk:"ssl"`
	Verify    types.String `tfsdk:"verify"`
	Cookie    types.String `tfsdk:"cookie"`
	Disabled  types.Bool   `tfsdk:"disabled"`
}

// haproxyFrontendModel maps the frontend block schema data.
type haproxyFrontendModel struct {
	Name           types.String               `tfsdk:"name"`
	Mode           types.String               `tfsdk:"mode"`
	DefaultBackend types.String               `tfsdk:"default_backend"`
	Maxconn        types.Int64                `tfsdk:"maxconn"`
	Backlog        types.Int64                `tfsdk:"backlog"`
	Ssl            types.Bool                 `tfsdk:"ssl"`
	SslCertificate types.String               `tfsdk:"ssl_certificate"`
	SslCafile      types.String               `tfsdk:"ssl_cafile"`
	SslMaxVer      types.String               `tfsdk:"ssl_max_ver"`
	SslMinVer      types.String               `tfsdk:"ssl_min_ver"`
	Ciphers        types.String               `tfsdk:"ciphers"`
	Ciphersuites   types.String               `tfsdk:"ciphersuites"`
	Verify         types.String               `tfsdk:"verify"`
	AcceptProxy    types.Bool                 `tfsdk:"accept_proxy"`
	DeferAccept    types.Bool                 `tfsdk:"defer_accept"`
	TcpUserTimeout types.Int64                `tfsdk:"tcp_user_timeout"`
	Tfo            types.Bool                 `tfsdk:"tfo"`
	V4v6           types.Bool                 `tfsdk:"v4v6"`
	V6only         types.Bool                 `tfsdk:"v6only"`
	Bind           []haproxyBindModel         `tfsdk:"bind"`
	StatsOptions   []haproxyStatsOptionsModel `tfsdk:"stats_options"`
	Acls           []haproxyAclModel          `tfsdk:"acls"`
}

// haproxyBalanceModel maps the balance block schema data.
type haproxyBalanceModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	UrlParam  types.String `tfsdk:"url_param"`
}

// haproxyHttpchkParamsModel maps the httpchk_params block schema data.
type haproxyHttpchkParamsModel struct {
	Method  types.String `tfsdk:"method"`
	Uri     types.String `tfsdk:"uri"`
	Version types.String `tfsdk:"version"`
}

// haproxyForwardforModel maps the forwardfor block schema data.
type haproxyForwardforModel struct {
	Enabled types.String `tfsdk:"enabled"`
}

// haproxyHttpcheckModel maps the httpcheck block schema data.
type haproxyHttpcheckModel struct {
	Index           types.Int64  `tfsdk:"index"`
	Type            types.String `tfsdk:"type"`
	Method          types.String `tfsdk:"method"`
	Uri             types.String `tfsdk:"uri"`
	Version         types.String `tfsdk:"version"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	Match           types.String `tfsdk:"match"`
	Pattern         types.String `tfsdk:"pattern"`
	Addr            types.String `tfsdk:"addr"`
	Port            types.Int64  `tfsdk:"port"`
	ExclamationMark types.String `tfsdk:"exclamation_mark"`
	LogLevel        types.String `tfsdk:"log_level"`
	SendProxy       types.String `tfsdk:"send_proxy"`
	ViaSocks4       types.String `tfsdk:"via_socks4"`
	CheckComment    types.String `tfsdk:"check_comment"`
}

// haproxyTcpCheckModel maps the tcp_check block schema data.
type haproxyTcpCheckModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// haproxyAclModel maps the acl block schema data.
type haproxyAclModel struct {
	AclName   types.String `tfsdk:"acl_name"`
	Index     types.Int64  `tfsdk:"index"`
	Criterion types.String `tfsdk:"criterion"`
	Value     types.String `tfsdk:"value"`
}

// haproxyHttpRequestRuleModel maps the http_request_rule block schema data.
type haproxyHttpRequestRuleModel struct {
	Index      types.Int64  `tfsdk:"index"`
	Type       types.String `tfsdk:"type"`
	RedirType  types.String `tfsdk:"redir_type"`
	RedirValue types.String `tfsdk:"redir_value"`
	Cond       types.String `tfsdk:"cond"`
	CondTest   types.String `tfsdk:"cond_test"`
	HdrName    types.String `tfsdk:"hdr_name"`
	HdrFormat  types.String `tfsdk:"hdr_format"`
}

// haproxyHttpResponseRuleModel maps the http_response_rule block schema data.
type haproxyHttpResponseRuleModel struct {
	Index     types.Int64  `tfsdk:"index"`
	Type      types.String `tfsdk:"type"`
	Cond      types.String `tfsdk:"cond"`
	CondTest  types.String `tfsdk:"cond_test"`
	HdrName   types.String `tfsdk:"hdr_name"`
	HdrFormat types.String `tfsdk:"hdr_format"`
}

// haproxyTcpRequestRuleModel maps the tcp_request_rule block schema data.
type haproxyTcpRequestRuleModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// haproxyTcpResponseRuleModel maps the tcp_response_rule block schema data.
type haproxyTcpResponseRuleModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// haproxyStickTableModel maps the stick_table block schema data.
type haproxyStickTableModel struct {
	Type    types.String `tfsdk:"type"`
	Size    types.Int64  `tfsdk:"size"`
	Expire  types.Int64  `tfsdk:"expire"`
	Nopurge types.Bool   `tfsdk:"nopurge"`
	Peers   types.String `tfsdk:"peers"`
}

// haproxyStickRuleModel maps the stick_rule block schema data.
type haproxyStickRuleModel struct {
	Index   types.Int64  `tfsdk:"index"`
	Type    types.String `tfsdk:"type"`
	Table   types.String `tfsdk:"table"`
	Pattern types.String `tfsdk:"pattern"`
}

// haproxyStatsOptionsModel maps the stats_options block schema data.
type haproxyStatsOptionsModel struct {
	StatsEnable types.Bool   `tfsdk:"stats_enable"`
	StatsUri    types.String `tfsdk:"stats_uri"`
	StatsRealm  types.String `tfsdk:"stats_realm"`
	StatsAuth   types.String `tfsdk:"stats_auth"`
}

// haproxyBindModel maps the bind block schema data.
type haproxyBindModel struct {
	Name         types.String `tfsdk:"name"`
	Address      types.String `tfsdk:"address"`
	Port         types.Int64  `tfsdk:"port"`
	PortRangeEnd types.Int64  `tfsdk:"port_range_end"`
	Transparent  types.Bool   `tfsdk:"transparent"`
	Mode         types.String `tfsdk:"mode"`
	Maxconn      types.Int64  `tfsdk:"maxconn"`
	Ssl          types.Bool   `tfsdk:"ssl"`
}

// Metadata returns the resource type name.
func (r *haproxyStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Schema defines the schema for the resource.
func (r *haproxyStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a complete HAProxy stack including backend, server, frontend, and ACLs.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the HAProxy stack.",
			},
		},
		Blocks: map[string]schema.Block{
			"backend":  GetBackendSchema(),
			"server":   GetServerSchema(),
			"frontend": GetFrontendSchema(),
			"acls":     GetACLSchema(),
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *haproxyStackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developer.", req.ProviderData),
		)
		return
	}

	// Initialize the stack manager with all required components
	aclManager := NewACLManager(client)
	frontendManager := NewFrontendManager(client)
	backendManager := NewBackendManager(client)
	r.stackManager = NewStackManager(client, aclManager, frontendManager, backendManager)
}

// Create resource.
func (r *haproxyStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if err := r.stackManager.Create(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error creating HAProxy stack", err.Error())
	}
}

// Read resource.
func (r *haproxyStackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if err := r.stackManager.Read(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy stack", err.Error())
	}
}

// Update resource.
func (r *haproxyStackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if err := r.stackManager.Update(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error updating HAProxy stack", err.Error())
	}
}

// Delete resource.
func (r *haproxyStackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.stackManager.Delete(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error deleting HAProxy stack", err.Error())
	}
}
