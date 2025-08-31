package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

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
	client *HAProxyClient
}

// haproxyStackResourceModel maps the resource schema data.
type haproxyStackResourceModel struct {
	Name     types.String          `tfsdk:"name"`
	Backend  *haproxyBackendModel  `tfsdk:"backend"`
	Server   *haproxyServerModel   `tfsdk:"server"`
	Frontend *haproxyFrontendModel `tfsdk:"frontend"`
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
	Name          types.String                `tfsdk:"name"`
	Address       types.String                `tfsdk:"address"`
	Port          types.Int64                 `tfsdk:"port"`
	Check         types.String                `tfsdk:"check"`
	Backup        types.String                `tfsdk:"backup"`
	Maxconn       types.Int64                 `tfsdk:"maxconn"`
	Weight        types.Int64                 `tfsdk:"weight"`
	Rise          types.Int64                 `tfsdk:"rise"`
	Fall          types.Int64                 `tfsdk:"fall"`
	Inter         types.Int64                 `tfsdk:"inter"`
	Fastinter     types.Int64                 `tfsdk:"fastinter"`
	Downinter     types.Int64                 `tfsdk:"downinter"`
	Ssl           types.String                `tfsdk:"ssl"`
	Verify        types.String                `tfsdk:"verify"`
	Cookie        types.String                `tfsdk:"cookie"`
	Disabled      types.Bool                  `tfsdk:"disabled"`
	HttpchkParams []haproxyHttpchkParamsModel `tfsdk:"httpchk_params"`
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

// haproxyStatsOptionsModel maps the stats_options block schema data.
type haproxyStatsOptionsModel struct {
	StatsEnable types.Bool   `tfsdk:"stats_enable"`
	StatsUri    types.String `tfsdk:"stats_uri"`
	StatsRealm  types.String `tfsdk:"stats_realm"`
	StatsAuth   types.String `tfsdk:"stats_auth"`
}

// Metadata returns the resource type name.
func (r *haproxyStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Schema defines the schema for the resource.
func (r *haproxyStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a complete HAProxy stack (backend, server, frontend) in a single transaction.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the stack.",
			},
		},
		Blocks: map[string]schema.Block{
			"backend": schema.SingleNestedBlock{
				Description: "Backend configuration.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required:    true,
						Description: "The name of the backend.",
					},
					"mode": schema.StringAttribute{
						Required:    true,
						Description: "The mode of the backend (http, tcp).",
					},
					"adv_check": schema.StringAttribute{
						Optional:    true,
						Description: "The advanced check of the backend. Allowed: ssl-hello-chk|smtpchk|ldap-check|mysql-check|pgsql-check|tcp-check|redis-check",
					},
					"http_connection_mode": schema.StringAttribute{
						Optional:    true,
						Description: "The http connection mode of the backend. Allowed: httpclose|http-server-close|http-keep-alive",
					},
					"server_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "The server timeout for the backend in milliseconds.",
					},
					"check_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "The check timeout for the backend in milliseconds.",
					},
					"connect_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "The connect timeout for the backend in milliseconds.",
					},
					"queue_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "The queue timeout for the backend in milliseconds.",
					},
					"tunnel_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "The tunnel timeout for the backend in milliseconds.",
					},
					"tarpit_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "The tarpit timeout for the backend in milliseconds.",
					},
					"checkcache": schema.StringAttribute{
						Optional:    true,
						Description: "The checkcache of the backend.",
					},
					"retries": schema.Int64Attribute{
						Optional:    true,
						Description: "The retries of the backend.",
					},
					"stick_table": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Stick table configuration for the backend.",
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required:    true,
								Description: "The type of the stick table.",
							},
							"size": schema.Int64Attribute{
								Optional:    true,
								Description: "The size of the stick table.",
							},
							"expire": schema.Int64Attribute{
								Optional:    true,
								Description: "The expiration time of the stick table.",
							},
							"nopurge": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether to disable purging of the stick table.",
							},
							"peers": schema.StringAttribute{
								Optional:    true,
								Description: "The peers for the stick table.",
							},
						},
					},
				},
				Blocks: map[string]schema.Block{
					"default_server": schema.SingleNestedBlock{
						Description: "Default server configuration for the backend. Note: SSL/TLS protocol control fields have different support levels between Data Plane API v2 and v3. Fields not supported by your HAProxy version will be silently ignored.",
						Attributes: map[string]schema.Attribute{
							"ssl": schema.StringAttribute{
								Optional:    true,
								Description: "Whether SSL is enabled for the default server. Use 'enabled' or 'disabled'.",
							},
							"ssl_cafile": schema.StringAttribute{
								Optional:    true,
								Description: "The SSL CA file for the default server.",
							},
							"ssl_certificate": schema.StringAttribute{
								Optional:    true,
								Description: "The SSL certificate for the default server.",
							},
							"ssl_max_ver": schema.StringAttribute{
								Optional:    true,
								Description: "The maximum SSL version for the default server.",
							},
							"ssl_min_ver": schema.StringAttribute{
								Optional:    true,
								Description: "The minimum SSL version for the default server.",
							},
							"ssl_reuse": schema.StringAttribute{
								Optional:    true,
								Description: "Whether SSL session reuse is enabled.",
							},
							"ciphers": schema.StringAttribute{
								Optional:    true,
								Description: "The SSL ciphers for the default server.",
							},
							"ciphersuites": schema.StringAttribute{
								Optional:    true,
								Description: "The SSL ciphersuites for the default server.",
							},
							"verify": schema.StringAttribute{
								Optional:    true,
								Description: "The SSL verification setting.",
							},
							"sslv3": schema.StringAttribute{
								Optional:    true,
								Description: "Whether SSLv3 is enabled. Use 'enabled' or 'disabled'. Only supported in Data Plane API v3.",
							},
							"tlsv10": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.0 is enabled. Use 'enabled' or 'disabled'.",
							},
							"tlsv11": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.1 is enabled. Use 'enabled' or 'disabled'.",
							},
							"tlsv12": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.2 is enabled. Use 'enabled' or 'disabled'.",
							},
							"tlsv13": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.3 is enabled. Use 'enabled' or 'disabled'.",
							},
							"no_sslv3": schema.StringAttribute{
								Optional:    true,
								Description: "Whether SSLv3 is disabled. Use 'enabled' or 'disabled'. Only supported in Data Plane API v2 (deprecated).",
							},
							"no_tlsv10": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.0 is disabled. Use 'enabled' or 'disabled'.",
							},
							"no_tlsv11": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.1 is disabled. Use 'enabled' or 'disabled'.",
							},
							"no_tlsv12": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.2 is disabled. Use 'enabled' or 'disabled'.",
							},
							"no_tlsv13": schema.StringAttribute{
								Optional:    true,
								Description: "Whether TLSv1.3 is disabled. Use 'enabled' or 'disabled'.",
							},
							"force_sslv3": schema.StringAttribute{
								Optional:    true,
								Description: "Whether to force SSLv3. Use 'enabled' or 'disabled'. Only supported in Data Plane API v3.",
							},
							"force_tlsv10": schema.StringAttribute{
								Optional:    true,
								Description: "Whether to force TLSv1.0. Use 'enabled' or 'disabled'.",
							},
							"force_tlsv11": schema.StringAttribute{
								Optional:    true,
								Description: "Whether to force TLSv1.1. Use 'enabled' or 'disabled'.",
							},
							"force_tlsv12": schema.StringAttribute{
								Optional:    true,
								Description: "Whether to force TLSv1.2. Use 'enabled' or 'disabled'.",
							},
							"force_tlsv13": schema.StringAttribute{
								Optional:    true,
								Description: "Whether to force TLSv1.3. Use 'enabled' or 'disabled'.",
							},
							"force_strict_sni": schema.StringAttribute{
								Optional:    true,
								Description: "Whether to force strict SNI.",
							},
						},
					},
					"balance": schema.ListNestedBlock{
						Description: "Load balancing configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"algorithm": schema.StringAttribute{
									Required:    true,
									Description: "The load balancing algorithm. Allowed: roundrobin|static-rr|leastconn|first|source|uri|url_param|hdr|rdp-cookie",
								},
								"url_param": schema.StringAttribute{
									Optional:    true,
									Description: "The URL parameter for load balancing.",
								},
							},
						},
					},
					"httpchk_params": schema.ListNestedBlock{
						Description: "HTTP health check parameters for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"method": schema.StringAttribute{
									Optional:    true,
									Description: "The HTTP method for health checks. Allowed: HEAD|PUT|POST|GET|TRACE|OPTIONS",
								},
								"uri": schema.StringAttribute{
									Optional:    true,
									Description: "The URI for health checks.",
								},
								"version": schema.StringAttribute{
									Optional:    true,
									Description: "The HTTP version for health checks.",
								},
							},
						},
					},
					"forwardfor": schema.ListNestedBlock{
						Description: "Forward for configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"enabled": schema.StringAttribute{
									Required:    true,
									Description: "The state of the forwardfor. Allowed: enabled|disabled",
								},
							},
						},
					},
					"httpcheck": schema.ListNestedBlock{
						Description: "HTTP health check configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the httpcheck.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the httpcheck.",
								},
								"method": schema.StringAttribute{
									Optional:    true,
									Description: "The HTTP method for the health check.",
								},
								"uri": schema.StringAttribute{
									Optional:    true,
									Description: "The URI for the health check.",
								},
								"version": schema.StringAttribute{
									Optional:    true,
									Description: "The HTTP version for the health check.",
								},
								"timeout": schema.Int64Attribute{
									Optional:    true,
									Description: "The timeout for the health check in milliseconds.",
								},
								"match": schema.StringAttribute{
									Optional:    true,
									Description: "The match condition for the health check.",
								},
								"pattern": schema.StringAttribute{
									Optional:    true,
									Description: "The pattern to match for the health check.",
								},
								"addr": schema.StringAttribute{
									Optional:    true,
									Description: "The address for the health check.",
								},
								"port": schema.Int64Attribute{
									Optional:    true,
									Description: "The port for the health check.",
								},
								"exclamation_mark": schema.StringAttribute{
									Optional:    true,
									Description: "The exclamation mark for the health check.",
								},
								"log_level": schema.StringAttribute{
									Optional:    true,
									Description: "The log level for the health check.",
								},
								"send_proxy": schema.StringAttribute{
									Optional:    true,
									Description: "The send proxy for the health check.",
								},
								"via_socks4": schema.StringAttribute{
									Optional:    true,
									Description: "The via socks4 for the health check.",
								},
								"check_comment": schema.StringAttribute{
									Optional:    true,
									Description: "The check comment for the health check.",
								},
							},
						},
					},
					"tcp_check": schema.ListNestedBlock{
						Description: "TCP health check configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the tcp_check.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the tcp_check.",
								},
								"action": schema.StringAttribute{
									Optional:    true,
									Description: "The action of the tcp_check.",
								},
								"cond": schema.StringAttribute{
									Optional:    true,
									Description: "The condition of the tcp_check.",
								},
								"cond_test": schema.StringAttribute{
									Optional:    true,
									Description: "The condition test of the tcp_check.",
								},
							},
						},
					},

					"http_request_rule": schema.ListNestedBlock{
						Description: "HTTP request rule configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the http-request rule.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the http-request rule.",
								},
								"cond": schema.StringAttribute{
									Optional:    true,
									Description: "The condition of the http-request rule.",
								},
								"cond_test": schema.StringAttribute{
									Optional:    true,
									Description: "The condition test of the http-request rule.",
								},
								"hdr_name": schema.StringAttribute{
									Optional:    true,
									Description: "The header name of the http-request rule.",
								},
								"hdr_format": schema.StringAttribute{
									Optional:    true,
									Description: "The header format of the http-request rule.",
								},
								"redir_type": schema.StringAttribute{
									Optional:    true,
									Description: "The redirection type of the http-request rule.",
								},
								"redir_value": schema.StringAttribute{
									Optional:    true,
									Description: "The redirection value of the http-request rule.",
								},
							},
						},
					},
					"http_response_rule": schema.ListNestedBlock{
						Description: "HTTP response rule configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the http-response rule.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the http-response rule.",
								},
								"cond": schema.StringAttribute{
									Optional:    true,
									Description: "The condition of the http-response rule.",
								},
								"cond_test": schema.StringAttribute{
									Optional:    true,
									Description: "The condition test of the http-response rule.",
								},
								"hdr_name": schema.StringAttribute{
									Optional:    true,
									Description: "The header name of the http-response rule.",
								},
								"hdr_format": schema.StringAttribute{
									Optional:    true,
									Description: "The header format of the http-response rule.",
								},
							},
						},
					},
					"tcp_request_rule": schema.ListNestedBlock{
						Description: "TCP request rule configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the tcp-request rule.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the tcp-request rule.",
								},
								"action": schema.StringAttribute{
									Optional:    true,
									Description: "The action of the tcp-request rule.",
								},
								"cond": schema.StringAttribute{
									Optional:    true,
									Description: "The condition of the tcp-request rule.",
								},
								"cond_test": schema.StringAttribute{
									Optional:    true,
									Description: "The condition test of the tcp-request rule.",
								},
							},
						},
					},
					"tcp_response_rule": schema.ListNestedBlock{
						Description: "TCP response rule configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the tcp-response rule.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the tcp-response rule.",
								},
								"action": schema.StringAttribute{
									Optional:    true,
									Description: "The action of the tcp-response rule.",
								},
								"cond": schema.StringAttribute{
									Optional:    true,
									Description: "The condition of the tcp-response rule.",
								},
								"cond_test": schema.StringAttribute{
									Optional:    true,
									Description: "The condition test of the tcp-response rule.",
								},
							},
						},
					},

					"stick_rule": schema.ListNestedBlock{
						Description: "Stick rule configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the stick rule.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of the stick rule.",
								},
								"table": schema.StringAttribute{
									Optional:    true,
									Description: "The table for the stick rule.",
								},
								"pattern": schema.StringAttribute{
									Optional:    true,
									Description: "The pattern for the stick rule.",
								},
							},
						},
					},
					"stats_options": schema.ListNestedBlock{
						Description: "Stats options configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"stats_enable": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether to enable stats for the backend.",
								},
								"stats_uri": schema.StringAttribute{
									Optional:    true,
									Description: "The stats URI for the backend.",
								},
								"stats_realm": schema.StringAttribute{
									Optional:    true,
									Description: "The stats realm for the backend.",
								},
								"stats_auth": schema.StringAttribute{
									Optional:    true,
									Description: "The stats authentication for the backend.",
								},
							},
						},
					},
					"acls": schema.ListNestedBlock{
						Description: "Access Control List (ACL) configuration blocks for content switching and decision making in the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"acl_name": schema.StringAttribute{
									Required:    true,
									Description: "The name of the ACL rule.",
								},
								"criterion": schema.StringAttribute{
									Required:    true,
									Description: "The criterion for the ACL rule (e.g., 'path', 'hdr', 'src').",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "The value for the ACL rule.",
								},
								"index": schema.Int64Attribute{
									Optional:    true,
									Description: "The index/order of the ACL rule. If not specified, will be auto-assigned.",
								},
							},
						},
					},
				},
			},
			"server": schema.SingleNestedBlock{
				Description: "Server configuration.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required:    true,
						Description: "The name of the server.",
					},
					"address": schema.StringAttribute{
						Required:    true,
						Description: "The address of the server.",
					},
					"port": schema.Int64Attribute{
						Required:    true,
						Description: "The port of the server.",
					},
					"check": schema.StringAttribute{
						Optional:    true,
						Description: "Whether to enable health checks for the server.",
					},
					"backup": schema.StringAttribute{
						Optional:    true,
						Description: "Whether the server is a backup server.",
					},
					"maxconn": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum number of connections for the server.",
					},
					"weight": schema.Int64Attribute{
						Optional:    true,
						Description: "Load balancing weight for the server.",
					},
					"rise": schema.Int64Attribute{
						Optional:    true,
						Description: "Number of successful health checks to mark server as up.",
					},
					"fall": schema.Int64Attribute{
						Optional:    true,
						Description: "Number of failed health checks to mark server as down.",
					},
					"inter": schema.Int64Attribute{
						Optional:    true,
						Description: "Health check interval in milliseconds.",
					},
					"fastinter": schema.Int64Attribute{
						Optional:    true,
						Description: "Fast health check interval in milliseconds.",
					},
					"downinter": schema.Int64Attribute{
						Optional:    true,
						Description: "Down health check interval in milliseconds.",
					},
					"ssl": schema.StringAttribute{
						Optional:    true,
						Description: "SSL configuration for the server.",
					},
					"verify": schema.StringAttribute{
						Optional:    true,
						Description: "SSL verification for the server.",
					},
					"cookie": schema.StringAttribute{
						Optional:    true,
						Description: "Cookie value for the server.",
					},
					"disabled": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether the server is disabled.",
					},
				},
				Blocks: map[string]schema.Block{
					"httpchk_params": schema.ListNestedBlock{
						Description: "HTTP health check parameters for the server.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"method": schema.StringAttribute{
									Optional:    true,
									Description: "The HTTP method for health checks. Allowed: HEAD|PUT|POST|GET|TRACE|OPTIONS",
								},
								"uri": schema.StringAttribute{
									Optional:    true,
									Description: "The URI for health checks.",
								},
								"version": schema.StringAttribute{
									Optional:    true,
									Description: "The HTTP version for health checks.",
								},
							},
						},
					},
				},
			},
			"frontend": schema.SingleNestedBlock{
				Description: "Frontend configuration.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required:    true,
						Description: "The name of the frontend.",
					},
					"mode": schema.StringAttribute{
						Required:    true,
						Description: "The mode of the frontend (http, tcp).",
					},
					"default_backend": schema.StringAttribute{
						Required:    true,
						Description: "The default backend for the frontend.",
					},
					"maxconn": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum number of connections for the frontend.",
					},
					"backlog": schema.Int64Attribute{
						Optional:    true,
						Description: "Backlog setting for the frontend.",
					},
					"ssl": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether SSL is enabled for the frontend.",
					},
					"ssl_certificate": schema.StringAttribute{
						Optional:    true,
						Description: "The SSL certificate for the frontend.",
					},
					"ssl_cafile": schema.StringAttribute{
						Optional:    true,
						Description: "The SSL CA file for the frontend.",
					},
					"ssl_max_ver": schema.StringAttribute{
						Optional:    true,
						Description: "The maximum SSL version for the frontend.",
					},
					"ssl_min_ver": schema.StringAttribute{
						Optional:    true,
						Description: "The minimum SSL version for the frontend.",
					},
					"ciphers": schema.StringAttribute{
						Optional:    true,
						Description: "The SSL ciphers for the frontend.",
					},
					"ciphersuites": schema.StringAttribute{
						Optional:    true,
						Description: "The SSL ciphersuites for the frontend.",
					},
					"verify": schema.StringAttribute{
						Optional:    true,
						Description: "The SSL verification setting for the frontend.",
					},
					"accept_proxy": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to accept proxy protocol.",
					},
					"defer_accept": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to defer accept.",
					},
					"tcp_user_timeout": schema.Int64Attribute{
						Optional:    true,
						Description: "TCP user timeout in milliseconds.",
					},
					"tfo": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether TCP Fast Open is enabled.",
					},
					"v4v6": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to support both IPv4 and IPv6.",
					},
					"v6only": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to support only IPv6.",
					},
				},
				Blocks: map[string]schema.Block{
					"bind": schema.ListNestedBlock{
						Description: "Bind configuration for the frontend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "The name of the bind.",
								},
								"address": schema.StringAttribute{
									Required:    true,
									Description: "The address to bind to.",
								},
								"port": schema.Int64Attribute{
									Optional:    true,
									Description: "The port to bind to.",
								},
								"port_range_end": schema.Int64Attribute{
									Optional:    true,
									Description: "The end of the port range.",
								},
								"transparent": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether to enable transparent binding.",
								},
								"mode": schema.StringAttribute{
									Optional:    true,
									Description: "The mode of the bind (http, tcp).",
								},
								"maxconn": schema.Int64Attribute{
									Optional:    true,
									Description: "Maximum connections for the bind.",
								},
								"ssl": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether SSL is enabled for the bind.",
								},
							},
						},
					},
					"stats_options": schema.ListNestedBlock{
						Description: "Stats options for the frontend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"stats_enable": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether to enable stats.",
								},
								"stats_uri": schema.StringAttribute{
									Optional:    true,
									Description: "The stats URI.",
								},
								"stats_realm": schema.StringAttribute{
									Optional:    true,
									Description: "The stats realm.",
								},
								"stats_auth": schema.StringAttribute{
									Optional:    true,
									Description: "The stats authentication.",
								},
							},
						},
					},
					"acls": schema.ListNestedBlock{
						Description: "Access Control List (ACL) configuration blocks for content switching and decision making.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"acl_name": schema.StringAttribute{
									Required:    true,
									Description: "The name of the ACL rule.",
								},
								"criterion": schema.StringAttribute{
									Required:    true,
									Description: "The criterion for the ACL rule (e.g., 'path', 'hdr', 'src').",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "The value for the ACL rule.",
								},
								"index": schema.Int64Attribute{
									Optional:    true,
									Description: "The index/order of the ACL rule. If not specified, will be auto-assigned.",
								},
							},
						},
					},
				},
			},
		},
	}
}

// ValidateConfig validates the configuration during plan time
func (r *haproxyStackResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config haproxyStackResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate SSL/TLS configuration if default_server is configured
	if config.Backend != nil && config.Backend.DefaultServer != nil {
		// Create a temporary client to get API version for validation
		// We'll use the provider's default API version for validation
		tempClient := &HAProxyClient{
			apiVersion: "v2", // Default to v2 for validation
		}

		if err := r.validateSSLConfiguration(config.Backend.DefaultServer, tempClient); err != nil {
			resp.Diagnostics.AddError(
				"Invalid SSL/TLS Configuration",
				err.Error(),
			)
		}
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

	r.client = client
}

// Create resource.
func (r *haproxyStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan haproxyStackResourceModel
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

	// Check if backend is provided
	if plan.Backend == nil {
		resp.Diagnostics.AddError(
			"Missing backend configuration",
			"Backend configuration is required for haproxy_stack resource",
		)
		return
	}

	// Check if server is provided
	if plan.Server == nil {
		resp.Diagnostics.AddError(
			"Missing server configuration",
			"Server configuration is required for haproxy_stack resource",
		)
		return
	}

	// Check if frontend is provided
	if plan.Frontend == nil {
		resp.Diagnostics.AddError(
			"Missing frontend configuration",
			"Frontend configuration is required for haproxy_stack resource",
		)
		return
	}

	// Validate SSL/TLS configuration based on HAProxy API version
	if err := r.validateSSLConfiguration(plan.Backend.DefaultServer, r.client); err != nil {
		resp.Diagnostics.AddError(
			"Invalid SSL/TLS configuration",
			err.Error(),
		)
		return
	}

	// Create payload for single transaction with version-aware DefaultServer
	allResources := &AllResourcesPayload{
		Backend: &BackendPayload{
			Name:               plan.Backend.Name.ValueString(),
			Mode:               plan.Backend.Mode.ValueString(),
			AdvCheck:           r.determineAdvCheckForAPI(plan.Backend.AdvCheck, plan.Backend.HttpchkParams),
			HttpConnectionMode: plan.Backend.HttpConnectionMode.ValueString(),
			ServerTimeout:      plan.Backend.ServerTimeout.ValueInt64(),
			CheckTimeout:       plan.Backend.CheckTimeout.ValueInt64(),
			ConnectTimeout:     plan.Backend.ConnectTimeout.ValueInt64(),
			QueueTimeout:       plan.Backend.QueueTimeout.ValueInt64(),
			TunnelTimeout:      plan.Backend.TunnelTimeout.ValueInt64(),
			TarpitTimeout:      plan.Backend.TarpitTimeout.ValueInt64(),
			CheckCache:         plan.Backend.Checkcache.ValueString(),
			Retries:            plan.Backend.Retries.ValueInt64(),

			// Process nested blocks
			Balance:       r.processBalanceBlock(plan.Backend.Balance),
			HttpchkParams: r.processHttpchkParamsBlock(plan.Backend.HttpchkParams),

			Forwardfor: r.processForwardforBlock(plan.Backend.Forwardfor),

			DefaultServer: func() *DefaultServerPayload {
				if plan.Backend.DefaultServer == nil {
					return nil
				}
				return &DefaultServerPayload{
					// Core SSL fields (supported in both v2 and v3)
					Ssl:            plan.Backend.DefaultServer.Ssl.ValueString(),
					SslCafile:      plan.Backend.DefaultServer.SslCafile.ValueString(),
					SslCertificate: plan.Backend.DefaultServer.SslCertificate.ValueString(),
					SslMaxVer:      plan.Backend.DefaultServer.SslMaxVer.ValueString(),
					SslMinVer:      plan.Backend.DefaultServer.SslMinVer.ValueString(),
					SslReuse:       plan.Backend.DefaultServer.SslReuse.ValueString(),
					Ciphers:        plan.Backend.DefaultServer.Ciphers.ValueString(),
					Ciphersuites:   plan.Backend.DefaultServer.Ciphersuites.ValueString(),
					Verify:         plan.Backend.DefaultServer.Verify.ValueString(),

					// Protocol control fields (v3 only)
					Sslv3:  plan.Backend.DefaultServer.Sslv3.ValueString(),
					Tlsv10: plan.Backend.DefaultServer.Tlsv10.ValueString(),
					Tlsv11: plan.Backend.DefaultServer.Tlsv11.ValueString(),
					Tlsv12: plan.Backend.DefaultServer.Tlsv12.ValueString(),
					Tlsv13: plan.Backend.DefaultServer.Tlsv13.ValueString(),

					// Deprecated fields (v2 only) - translate to force fields
					NoSslv3:  plan.Backend.DefaultServer.NoSslv3.ValueString(),
					NoTlsv10: r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv10.ValueString()),
					NoTlsv11: r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv11.ValueString()),
					NoTlsv12: r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv12.ValueString()),
					NoTlsv13: r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv13.ValueString()),

					// Force fields (v3 only)
					ForceSslv3:     plan.Backend.DefaultServer.ForceSslv3.ValueString(),
					ForceTlsv10:    plan.Backend.DefaultServer.ForceTlsv10.ValueString(),
					ForceTlsv11:    plan.Backend.DefaultServer.ForceTlsv11.ValueString(),
					ForceTlsv12:    plan.Backend.DefaultServer.ForceTlsv12.ValueString(),
					ForceTlsv13:    plan.Backend.DefaultServer.ForceTlsv13.ValueString(),
					ForceStrictSni: plan.Backend.DefaultServer.ForceStrictSni.ValueString(),
				}
			}(),
		},
		Servers: []ServerResource{
			{
				ParentType: "backend",
				ParentName: plan.Backend.Name.ValueString(),
				Payload: &ServerPayload{
					Name:      plan.Server.Name.ValueString(),
					Address:   plan.Server.Address.ValueString(),
					Port:      plan.Server.Port.ValueInt64(),
					Check:     plan.Server.Check.ValueString(),
					Backup:    plan.Server.Backup.ValueString(),
					Maxconn:   plan.Server.Maxconn.ValueInt64(),
					Weight:    plan.Server.Weight.ValueInt64(),
					Rise:      plan.Server.Rise.ValueInt64(),
					Fall:      plan.Server.Fall.ValueInt64(),
					Inter:     plan.Server.Inter.ValueInt64(),
					Fastinter: plan.Server.Fastinter.ValueInt64(),
					Downinter: plan.Server.Downinter.ValueInt64(),
					Ssl:       plan.Server.Ssl.ValueString(),
					Verify:    plan.Server.Verify.ValueString(),
					Cookie:    plan.Server.Cookie.ValueString(),
					Disabled:  plan.Server.Disabled.ValueBool(),
				},
			},
		},
		Frontend: &FrontendPayload{
			Name:           plan.Frontend.Name.ValueString(),
			Mode:           plan.Frontend.Mode.ValueString(),
			DefaultBackend: plan.Frontend.DefaultBackend.ValueString(),
			MaxConn:        plan.Frontend.Maxconn.ValueInt64(),
			Backlog:        plan.Frontend.Backlog.ValueInt64(),
		},
	}

	// Debug: Log the payload being sent to HAProxy
	log.Printf("DEBUG: Backend payload being sent to HAProxy:")
	log.Printf("  AdvCheck: %+v", allResources.Backend.AdvCheck)
	log.Printf("  HttpchkParams: %+v", allResources.Backend.HttpchkParams)
	log.Printf("  Balance: %+v", allResources.Backend.Balance)
	log.Printf("  Forwardfor: %+v", allResources.Backend.Forwardfor)
	log.Printf("  Original AdvCheck from plan: %+v", plan.Backend.AdvCheck.ValueString())
	log.Printf("  HttpchkParams from plan: %+v", len(plan.Backend.HttpchkParams))

	// Prepare Frontend ACLs for the transaction
	if plan.Frontend.Acls != nil && len(plan.Frontend.Acls) > 0 {
		// Sort ACLs by index to ensure proper order
		sortedAcls := r.processAclsBlock(plan.Frontend.Acls)

		// Create ACLs with sequential indices starting from 0
		nextIndex := int64(0) // Start from 0 as required by HAProxy

		for _, acl := range sortedAcls {
			aclResource := ACLResource{
				ParentType: "frontend",
				ParentName: plan.Frontend.Name.ValueString(),
				Payload: &ACLPayload{
					AclName:   acl.AclName.ValueString(),
					Criterion: acl.Criterion.ValueString(),
					Value:     acl.Value.ValueString(),
					Index:     nextIndex, // Use sequential index starting from 0
				},
			}
			allResources.Acls = append(allResources.Acls, aclResource)
			nextIndex++ // Increment for next ACL
		}
	}

	// Prepare Backend ACLs for the transaction
	if plan.Backend.Acls != nil && len(plan.Backend.Acls) > 0 {
		// Sort ACLs by index to ensure proper order
		sortedAcls := r.processAclsBlock(plan.Backend.Acls)

		// Create ACLs with sequential indices starting from 0
		nextIndex := int64(0) // Start from 0 as required by HAProxy

		for _, acl := range sortedAcls {
			aclResource := ACLResource{
				ParentType: "backend",
				ParentName: plan.Backend.Name.ValueString(),
				Payload: &ACLPayload{
					AclName:   acl.AclName.ValueString(),
					Criterion: acl.Criterion.ValueString(),
					Value:     acl.Value.ValueString(),
					Index:     nextIndex, // Use sequential index starting from 0
				},
			}
			allResources.Acls = append(allResources.Acls, aclResource)
			nextIndex++ // Increment for next ACL
		}
	}

	// Create all resources in single transaction (including ACLs)
	err := r.client.CreateAllResourcesInSingleTransaction(ctx, allResources)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HAProxy stack",
			"Could not create HAProxy stack, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource.
func (r *haproxyStackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state haproxyStackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Store the names we need to read
	backendName := state.Backend.Name.ValueString()
	serverName := state.Server.Name.ValueString()
	frontendName := state.Frontend.Name.ValueString()

	// Store existing values to preserve them
	existingBackend := state.Backend
	existingServer := state.Server
	existingFrontend := state.Frontend

	// Reset the state completely to avoid drift
	state = haproxyStackResourceModel{
		Name: types.StringValue(state.Name.ValueString()),
		Backend: &haproxyBackendModel{
			Name: types.StringValue(backendName),
		},
		Server: &haproxyServerModel{
			Name: types.StringValue(serverName),
		},
		Frontend: &haproxyFrontendModel{
			Name: types.StringValue(frontendName),
		},
	}

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Read all resources from HAProxy
	backend, err := r.client.ReadBackend(ctx, backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backend",
			"Could not read backend, unexpected error: "+err.Error(),
		)
		return
	}

	servers, err := r.client.ReadServers(ctx, "backend", backendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading servers",
			"Could not read servers, unexpected error: "+err.Error(),
		)
		return
	}

	frontend, err := r.client.ReadFrontend(ctx, frontendName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading frontend",
			"Could not read frontend, unexpected error: "+err.Error(),
		)
		return
	}

	// Read ACLs for the frontend
	var frontendAcls []AclPayload
	if frontend != nil {
		frontendAcls, err = r.client.ReadAcls(ctx, "frontend", frontendName)
		if err != nil {
			log.Printf("Warning: Failed to read ACLs for frontend %s: %v", frontendName, err)
			// Continue without ACLs if reading fails
		}
	}

	// Read ACLs for the backend
	var backendAcls []AclPayload
	backendAcls, err = r.client.ReadAcls(ctx, "backend", backendName)
	if err != nil {
		log.Printf("Warning: Failed to read ACLs for backend %s: %v", backendName, err)
		// Continue without ACLs if reading fails
	}

	// Update state with actual HAProxy configuration
	if backend != nil {
		state.Backend.Mode = types.StringValue(backend.Mode)

		// Handle adv_check based on whether httpchk_params is present
		if len(existingBackend.HttpchkParams) > 0 && existingBackend.AdvCheck.IsNull() {
			// If httpchk_params is configured and adv_check was not explicitly set,
			// adv_check should be "httpchk" but we don't store it in state since it's auto-managed
			state.Backend.AdvCheck = types.StringNull()
		} else if !existingBackend.AdvCheck.IsNull() && !existingBackend.AdvCheck.IsUnknown() {
			// Preserve the explicitly configured adv_check value
			state.Backend.AdvCheck = existingBackend.AdvCheck
		} else if backend.AdvCheck != "" {
			// Only set adv_check if HAProxy returned it and no explicit configuration
			state.Backend.AdvCheck = types.StringValue(backend.AdvCheck)
		} else {
			state.Backend.AdvCheck = types.StringNull()
		}

		// Only set fields if HAProxy actually returned them
		if backend.HttpConnectionMode != "" {
			state.Backend.HttpConnectionMode = types.StringValue(backend.HttpConnectionMode)
		}
		if backend.ServerTimeout != 0 {
			state.Backend.ServerTimeout = types.Int64Value(backend.ServerTimeout)
		}
		if backend.CheckTimeout != 0 {
			state.Backend.CheckTimeout = types.Int64Value(backend.CheckTimeout)
		}
		if backend.ConnectTimeout != 0 {
			state.Backend.ConnectTimeout = types.Int64Value(backend.ConnectTimeout)
		}
		if backend.QueueTimeout != 0 {
			state.Backend.QueueTimeout = types.Int64Value(backend.QueueTimeout)
		}
		if backend.TunnelTimeout != 0 {
			state.Backend.TunnelTimeout = types.Int64Value(backend.TunnelTimeout)
		}
		if backend.TarpitTimeout != 0 {
			state.Backend.TarpitTimeout = types.Int64Value(backend.TarpitTimeout)
		}
		if backend.CheckCache != "" {
			state.Backend.Checkcache = types.StringValue(backend.CheckCache)
		}
		if backend.Retries != 0 {
			state.Backend.Retries = types.Int64Value(backend.Retries)
		}

		// Handle default_server configuration
		if backend.DefaultServer != nil {
			// Initialize DefaultServer only if we have data to set
			state.Backend.DefaultServer = &haproxyDefaultServerModel{}

			// Only set fields that HAProxy actually returned (non-empty)
			if backend.DefaultServer.Ssl != "" {
				state.Backend.DefaultServer.Ssl = types.StringValue(backend.DefaultServer.Ssl)
			}
			if backend.DefaultServer.SslCafile != "" {
				state.Backend.DefaultServer.SslCafile = types.StringValue(backend.DefaultServer.SslCafile)
			}
			if backend.DefaultServer.SslCertificate != "" {
				state.Backend.DefaultServer.SslCertificate = types.StringValue(backend.DefaultServer.SslCertificate)
			}
			if backend.DefaultServer.SslMaxVer != "" {
				state.Backend.DefaultServer.SslMaxVer = types.StringValue(backend.DefaultServer.SslMaxVer)
			}
			if backend.DefaultServer.SslMinVer != "" {
				state.Backend.DefaultServer.SslMinVer = types.StringValue(backend.DefaultServer.SslMinVer)
			}
			if backend.DefaultServer.SslReuse != "" {
				state.Backend.DefaultServer.SslReuse = types.StringValue(backend.DefaultServer.SslReuse)
			}
			if backend.DefaultServer.Ciphers != "" {
				state.Backend.DefaultServer.Ciphers = types.StringValue(backend.DefaultServer.Ciphers)
			}
			if backend.DefaultServer.Ciphersuites != "" {
				state.Backend.DefaultServer.Ciphersuites = types.StringValue(backend.DefaultServer.Ciphersuites)
			}
			if backend.DefaultServer.Verify != "" {
				state.Backend.DefaultServer.Verify = types.StringValue(backend.DefaultServer.Verify)
			}
			if backend.DefaultServer.Sslv3 != "" {
				state.Backend.DefaultServer.Sslv3 = types.StringValue(backend.DefaultServer.Sslv3)
			}
			if backend.DefaultServer.Tlsv10 != "" {
				state.Backend.DefaultServer.Tlsv10 = types.StringValue(backend.DefaultServer.Tlsv10)
			}
			if backend.DefaultServer.Tlsv11 != "" {
				state.Backend.DefaultServer.Tlsv11 = types.StringValue(backend.DefaultServer.Tlsv11)
			}
			if backend.DefaultServer.Tlsv12 != "" {
				state.Backend.DefaultServer.Tlsv12 = types.StringValue(backend.DefaultServer.Tlsv12)
			}
			if backend.DefaultServer.Tlsv13 != "" {
				state.Backend.DefaultServer.Tlsv13 = types.StringValue(backend.DefaultServer.Tlsv13)
			}

			// Handle no_tlsv* fields - preserve existing state if HAProxy doesn't return them
			// These fields get translated to force_tlsv* internally by HAProxy
			if backend.DefaultServer.NoSslv3 != "" {
				state.Backend.DefaultServer.NoSslv3 = types.StringValue(backend.DefaultServer.NoSslv3)
			} else if existingBackend.DefaultServer != nil && !existingBackend.DefaultServer.NoSslv3.IsNull() && !existingBackend.DefaultServer.NoSslv3.IsUnknown() {
				// Preserve existing state to avoid unnecessary updates
				state.Backend.DefaultServer.NoSslv3 = existingBackend.DefaultServer.NoSslv3
			}
			if backend.DefaultServer.NoTlsv10 != "" {
				state.Backend.DefaultServer.NoTlsv10 = types.StringValue(backend.DefaultServer.NoTlsv10)
			} else if existingBackend.DefaultServer != nil && !existingBackend.DefaultServer.NoTlsv10.IsNull() && !existingBackend.DefaultServer.NoTlsv10.IsUnknown() {
				// Preserve existing state to avoid unnecessary updates
				state.Backend.DefaultServer.NoTlsv10 = existingBackend.DefaultServer.NoTlsv10
			}
			if backend.DefaultServer.NoTlsv11 != "" {
				state.Backend.DefaultServer.NoTlsv11 = types.StringValue(backend.DefaultServer.NoTlsv11)
			} else if existingBackend.DefaultServer != nil && !existingBackend.DefaultServer.NoTlsv11.IsNull() && !existingBackend.DefaultServer.NoTlsv11.IsUnknown() {
				// Preserve existing state to avoid unnecessary updates
				state.Backend.DefaultServer.NoTlsv11 = existingBackend.DefaultServer.NoTlsv11
			}
			if backend.DefaultServer.NoTlsv12 != "" {
				state.Backend.DefaultServer.NoTlsv12 = types.StringValue(backend.DefaultServer.NoTlsv12)
			} else if existingBackend.DefaultServer != nil && !existingBackend.DefaultServer.NoTlsv12.IsNull() && !existingBackend.DefaultServer.NoTlsv12.IsUnknown() {
				// Preserve existing state to avoid unnecessary updates
				state.Backend.DefaultServer.NoTlsv12 = existingBackend.DefaultServer.NoTlsv12
			}
			if backend.DefaultServer.NoTlsv13 != "" {
				state.Backend.DefaultServer.NoTlsv13 = types.StringValue(backend.DefaultServer.NoTlsv13)
			} else if existingBackend.DefaultServer != nil && !existingBackend.DefaultServer.NoTlsv13.IsNull() && !existingBackend.DefaultServer.NoTlsv13.IsUnknown() {
				// Preserve existing state to avoid unnecessary updates
				state.Backend.DefaultServer.NoTlsv13 = existingBackend.DefaultServer.NoTlsv13
			}

			// Handle force_tlsv* fields
			if backend.DefaultServer.ForceSslv3 != "" {
				state.Backend.DefaultServer.ForceSslv3 = types.StringValue(backend.DefaultServer.ForceSslv3)
			}
			if backend.DefaultServer.ForceTlsv10 != "" {
				state.Backend.DefaultServer.ForceTlsv10 = types.StringValue(backend.DefaultServer.ForceTlsv10)
			}
			if backend.DefaultServer.ForceTlsv11 != "" {
				state.Backend.DefaultServer.ForceTlsv11 = types.StringValue(backend.DefaultServer.ForceTlsv11)
			}
			if backend.DefaultServer.ForceTlsv12 != "" {
				state.Backend.DefaultServer.ForceTlsv12 = types.StringValue(backend.DefaultServer.ForceTlsv12)
			}
			if backend.DefaultServer.ForceTlsv13 != "" {
				state.Backend.DefaultServer.ForceTlsv13 = types.StringValue(backend.DefaultServer.ForceTlsv13)
			}
			if backend.DefaultServer.ForceStrictSni != "" {
				state.Backend.DefaultServer.ForceStrictSni = types.StringValue(backend.DefaultServer.ForceStrictSni)
			}
		}

		// Translate force_tlsv* fields back to no_tlsv* fields for state consistency
		// Only translate if we don't already have preserved values
		if backend.DefaultServer != nil && state.Backend.DefaultServer != nil {
			// Translate force_tlsv* back to no_tlsv* only if we don't have preserved values
			if backend.DefaultServer.ForceTlsv10 != "" && (state.Backend.DefaultServer.NoTlsv10.IsNull() || state.Backend.DefaultServer.NoTlsv10.IsUnknown()) {
				state.Backend.DefaultServer.NoTlsv10 = types.StringValue(r.translateForceTlsToNoTls(backend.DefaultServer.ForceTlsv10))
			}
			if backend.DefaultServer.ForceTlsv11 != "" && (state.Backend.DefaultServer.NoTlsv11.IsNull() || state.Backend.DefaultServer.NoTlsv11.IsUnknown()) {
				state.Backend.DefaultServer.NoTlsv11 = types.StringValue(r.translateForceTlsToNoTls(backend.DefaultServer.ForceTlsv11))
			}
			if backend.DefaultServer.ForceTlsv12 != "" && (state.Backend.DefaultServer.NoTlsv12.IsNull() || state.Backend.DefaultServer.NoTlsv12.IsUnknown()) {
				state.Backend.DefaultServer.NoTlsv12 = types.StringValue(r.translateForceTlsToNoTls(backend.DefaultServer.ForceTlsv12))
			}
			if backend.DefaultServer.ForceTlsv13 != "" && (state.Backend.DefaultServer.NoTlsv13.IsNull() || state.Backend.DefaultServer.NoTlsv13.IsUnknown()) {
				state.Backend.DefaultServer.NoTlsv13 = types.StringValue(r.translateForceTlsToNoTls(backend.DefaultServer.ForceTlsv13))
			}
		}

		// Handle nested blocks
		if backend.Balance != nil {
			balanceModel := haproxyBalanceModel{
				Algorithm: types.StringValue(backend.Balance.Algorithm),
			}

			// Only set UrlParam if it has a value
			if backend.Balance.UrlParam != "" {
				balanceModel.UrlParam = types.StringValue(backend.Balance.UrlParam)
			}

			state.Backend.Balance = []haproxyBalanceModel{balanceModel}
		} else if existingBackend.Balance != nil && len(existingBackend.Balance) > 0 {
			// Preserve existing balance configuration if HAProxy didn't return it
			state.Backend.Balance = existingBackend.Balance
		}

		if backend.HttpchkParams != nil {
			httpchkModel := haproxyHttpchkParamsModel{}

			// Only set fields if they have values
			if backend.HttpchkParams.Method != "" {
				httpchkModel.Method = types.StringValue(backend.HttpchkParams.Method)
			}
			if backend.HttpchkParams.Uri != "" {
				httpchkModel.Uri = types.StringValue(backend.HttpchkParams.Uri)
			}
			if backend.HttpchkParams.Version != "" {
				httpchkModel.Version = types.StringValue(backend.HttpchkParams.Version)
			}

			state.Backend.HttpchkParams = []haproxyHttpchkParamsModel{httpchkModel}
		} else if existingBackend.HttpchkParams != nil && len(existingBackend.HttpchkParams) > 0 {
			// Preserve existing httpchk_params configuration if HAProxy didn't return it
			state.Backend.HttpchkParams = existingBackend.HttpchkParams
		}

		if backend.Forwardfor != nil {
			forwardforModel := haproxyForwardforModel{}

			// Only set Enabled if it has a value
			if backend.Forwardfor.Enabled != "" {
				forwardforModel.Enabled = types.StringValue(backend.Forwardfor.Enabled)
			}

			state.Backend.Forwardfor = []haproxyForwardforModel{forwardforModel}
		} else if existingBackend.Forwardfor != nil && len(existingBackend.Forwardfor) > 0 {
			// Preserve existing forwardfor configuration if HAProxy didn't return it
			state.Backend.Forwardfor = existingBackend.Forwardfor
		}
	}

	if len(servers) > 0 {
		server := servers[0] // Take first server

		// Only set fields if HAProxy actually returned them
		if server.Name != "" {
			state.Server.Name = types.StringValue(server.Name)
		}
		if server.Address != "" {
			state.Server.Address = types.StringValue(server.Address)
		}
		if server.Port != 0 {
			state.Server.Port = types.Int64Value(server.Port)
		}
		if server.Check != "" {
			state.Server.Check = types.StringValue(server.Check)
		}
		if server.Backup != "" {
			state.Server.Backup = types.StringValue(server.Backup)
		}
		if server.Maxconn != 0 {
			state.Server.Maxconn = types.Int64Value(server.Maxconn)
		}
		if server.Weight != 0 {
			state.Server.Weight = types.Int64Value(server.Weight)
		}
		if server.Rise != 0 {
			state.Server.Rise = types.Int64Value(server.Rise)
		}
		if server.Fall != 0 {
			state.Server.Fall = types.Int64Value(server.Fall)
		}
		if server.Inter != 0 {
			state.Server.Inter = types.Int64Value(server.Inter)
		}
		if server.Fastinter != 0 {
			state.Server.Fastinter = types.Int64Value(server.Fastinter)
		}
		if server.Downinter != 0 {
			state.Server.Downinter = types.Int64Value(server.Downinter)
		}
		if server.Ssl != "" {
			state.Server.Ssl = types.StringValue(server.Ssl)
		}
		if server.Verify != "" {
			state.Server.Verify = types.StringValue(server.Verify)
		}
		if server.Cookie != "" {
			state.Server.Cookie = types.StringValue(server.Cookie)
		}
		// For boolean fields, we need to check if they're not the zero value
		// Since HAProxy might not return these fields if they're false
		state.Server.Disabled = types.BoolValue(server.Disabled)
	} else if existingServer != nil {
		// Preserve existing server configuration if HAProxy didn't return it
		state.Server = existingServer
	}

	if frontend != nil {
		// Only set fields if HAProxy actually returned them
		if frontend.Mode != "" {
			state.Frontend.Mode = types.StringValue(frontend.Mode)
		}
		if frontend.DefaultBackend != "" {
			state.Frontend.DefaultBackend = types.StringValue(frontend.DefaultBackend)
		}
		if frontend.MaxConn != 0 {
			state.Frontend.Maxconn = types.Int64Value(frontend.MaxConn)
		}
		if frontend.Backlog != 0 {
			state.Frontend.Backlog = types.Int64Value(frontend.Backlog)
		}

		// Handle Frontend ACLs - ALWAYS preserve user's exact configuration from state
		if existingFrontend.Acls != nil && len(existingFrontend.Acls) > 0 {
			// ALWAYS use the existing ACLs from state to preserve user's exact order
			log.Printf("DEBUG: Using existing frontend ACLs from state to preserve user's exact order: %s", r.formatAclOrder(existingFrontend.Acls))
			state.Frontend.Acls = existingFrontend.Acls
		} else if len(frontendAcls) > 0 {
			// Only create new ACLs if there are no existing ones in state
			log.Printf("DEBUG: No existing ACLs in state, creating from HAProxy response")
			var aclModels []haproxyAclModel
			for _, acl := range frontendAcls {
				aclModels = append(aclModels, haproxyAclModel{
					AclName:   types.StringValue(acl.AclName),
					Criterion: types.StringValue(acl.Criterion),
					Value:     types.StringValue(acl.Value),
					Index:     types.Int64Value(acl.Index),
				})
			}
			state.Frontend.Acls = aclModels
			log.Printf("Frontend ACLs created from HAProxy: %s", r.formatAclOrder(aclModels))
		}
	} else if existingFrontend != nil {
		// Preserve existing frontend configuration if HAProxy didn't return it
		state.Frontend = existingFrontend
	}

	// Handle Backend ACLs - ALWAYS preserve user's exact configuration from state
	if existingBackend.Acls != nil && len(existingBackend.Acls) > 0 {
		// ALWAYS use the existing ACLs from state to preserve user's exact order
		log.Printf("DEBUG: Using existing backend ACLs from state to preserve user's exact order: %s", r.formatAclOrder(existingBackend.Acls))
		state.Backend.Acls = existingBackend.Acls
	} else if len(backendAcls) > 0 {
		// Only create new ACLs if there are no existing ones in state
		log.Printf("DEBUG: No existing backend ACLs in state, creating from HAProxy response")
		var backendAclModels []haproxyAclModel
		for _, acl := range backendAcls {
			backendAclModels = append(backendAclModels, haproxyAclModel{
				AclName:   types.StringValue(acl.AclName),
				Criterion: types.StringValue(acl.Criterion),
				Value:     types.StringValue(acl.Value),
				Index:     types.Int64Value(acl.Index),
			})
		}
		state.Backend.Acls = backendAclModels
		log.Printf("Backend ACLs created from HAProxy: %s", r.formatAclOrder(backendAclModels))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource.
func (r *haproxyStackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan haproxyStackResourceModel
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

	// Check if backend is provided
	if plan.Backend == nil {
		resp.Diagnostics.AddError(
			"Missing backend configuration",
			"Backend configuration is required for haproxy_stack resource",
		)
		return
	}

	// Check if server is provided
	if plan.Server == nil {
		resp.Diagnostics.AddError(
			"Missing server configuration",
			"Server configuration is required for haproxy_stack resource",
		)
		return
	}

	// Check if frontend is provided
	if plan.Frontend == nil {
		resp.Diagnostics.AddError(
			"Missing frontend configuration",
			"Frontend configuration is required for haproxy_stack resource",
		)
		return
	}

	// Validate SSL/TLS configuration based on HAProxy API version
	if err := r.validateSSLConfiguration(plan.Backend.DefaultServer, r.client); err != nil {
		resp.Diagnostics.AddError(
			"Invalid SSL/TLS configuration",
			err.Error(),
		)
		return
	}

	// Create payload for single transaction with ALL configuration
	allResources := &AllResourcesPayload{
		Backend: &BackendPayload{
			Name:               plan.Backend.Name.ValueString(),
			Mode:               plan.Backend.Mode.ValueString(),
			AdvCheck:           r.determineAdvCheckForAPI(plan.Backend.AdvCheck, plan.Backend.HttpchkParams),
			HttpConnectionMode: plan.Backend.HttpConnectionMode.ValueString(),
			ServerTimeout:      plan.Backend.ServerTimeout.ValueInt64(),
			CheckTimeout:       plan.Backend.CheckTimeout.ValueInt64(),
			ConnectTimeout:     plan.Backend.ConnectTimeout.ValueInt64(),
			QueueTimeout:       plan.Backend.QueueTimeout.ValueInt64(),
			TunnelTimeout:      plan.Backend.TunnelTimeout.ValueInt64(),
			TarpitTimeout:      plan.Backend.TarpitTimeout.ValueInt64(),
			CheckCache:         plan.Backend.Checkcache.ValueString(),
			Retries:            plan.Backend.Retries.ValueInt64(),

			// Process nested blocks
			Balance:       r.processBalanceBlock(plan.Backend.Balance),
			HttpchkParams: r.processHttpchkParamsBlock(plan.Backend.HttpchkParams),

			Forwardfor: r.processForwardforBlock(plan.Backend.Forwardfor),

			DefaultServer: func() *DefaultServerPayload {
				if plan.Backend.DefaultServer == nil {
					return nil
				}
				return &DefaultServerPayload{
					Ssl:            plan.Backend.DefaultServer.Ssl.ValueString(),
					SslCafile:      plan.Backend.DefaultServer.SslCafile.ValueString(),
					SslCertificate: plan.Backend.DefaultServer.SslCertificate.ValueString(),
					SslMaxVer:      plan.Backend.DefaultServer.SslMaxVer.ValueString(),
					SslMinVer:      plan.Backend.DefaultServer.SslMinVer.ValueString(),
					SslReuse:       plan.Backend.DefaultServer.SslReuse.ValueString(),
					Ciphers:        plan.Backend.DefaultServer.Ciphers.ValueString(),
					Ciphersuites:   plan.Backend.DefaultServer.Ciphersuites.ValueString(),
					Verify:         plan.Backend.DefaultServer.Verify.ValueString(),
					Sslv3:          plan.Backend.DefaultServer.Sslv3.ValueString(),
					Tlsv10:         plan.Backend.DefaultServer.Tlsv10.ValueString(),
					Tlsv11:         plan.Backend.DefaultServer.Tlsv11.ValueString(),
					Tlsv12:         plan.Backend.DefaultServer.Tlsv12.ValueString(),
					Tlsv13:         plan.Backend.DefaultServer.Tlsv13.ValueString(),
					NoSslv3:        plan.Backend.DefaultServer.NoSslv3.ValueString(),
					NoTlsv10:       r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv10.ValueString()),
					NoTlsv11:       r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv11.ValueString()),
					NoTlsv12:       r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv12.ValueString()),
					NoTlsv13:       r.translateNoTlsToForceTls(plan.Backend.DefaultServer.NoTlsv13.ValueString()),
					ForceSslv3:     plan.Backend.DefaultServer.ForceSslv3.ValueString(),
					ForceTlsv10:    plan.Backend.DefaultServer.ForceTlsv10.ValueString(),
					ForceTlsv11:    plan.Backend.DefaultServer.ForceTlsv11.ValueString(),
					ForceTlsv12:    plan.Backend.DefaultServer.ForceTlsv12.ValueString(),
					ForceTlsv13:    plan.Backend.DefaultServer.ForceTlsv13.ValueString(),
					ForceStrictSni: plan.Backend.DefaultServer.ForceStrictSni.ValueString(),
				}
			}(),
		},
		Servers: []ServerResource{
			{
				ParentType: "backend",
				ParentName: plan.Backend.Name.ValueString(),
				Payload: &ServerPayload{
					Name:      plan.Server.Name.ValueString(),
					Address:   plan.Server.Address.ValueString(),
					Port:      plan.Server.Port.ValueInt64(),
					Check:     plan.Server.Check.ValueString(),
					Backup:    plan.Server.Backup.ValueString(),
					Maxconn:   plan.Server.Maxconn.ValueInt64(),
					Weight:    plan.Server.Weight.ValueInt64(),
					Rise:      plan.Server.Rise.ValueInt64(),
					Fall:      plan.Server.Fall.ValueInt64(),
					Inter:     plan.Server.Inter.ValueInt64(),
					Fastinter: plan.Server.Fastinter.ValueInt64(),
					Downinter: plan.Server.Downinter.ValueInt64(),
					Ssl:       plan.Server.Ssl.ValueString(),
					Verify:    plan.Server.Verify.ValueString(),
					Cookie:    plan.Server.Cookie.ValueString(),
					Disabled:  plan.Server.Disabled.ValueBool(),
				},
			},
		},
		Frontend: &FrontendPayload{
			Name:           plan.Frontend.Name.ValueString(),
			Mode:           plan.Frontend.Mode.ValueString(),
			DefaultBackend: plan.Frontend.DefaultBackend.ValueString(),
			MaxConn:        plan.Frontend.Maxconn.ValueInt64(),
			Backlog:        plan.Frontend.Backlog.ValueInt64(),
		},
	}

	// Update backend
	err := r.client.UpdateBackend(ctx, plan.Backend.Name.ValueString(), allResources.Backend)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy stack",
			"Could not update backend, unexpected error: "+err.Error(),
		)
		return
	}

	// Update server
	err = r.client.UpdateServer(ctx, plan.Server.Name.ValueString(), "backend", plan.Backend.Name.ValueString(), allResources.Servers[0].Payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy stack",
			"Could not update server, unexpected error: "+err.Error(),
		)
		return
	}

	// Update frontend
	err = r.client.UpdateFrontend(ctx, plan.Frontend.Name.ValueString(), allResources.Frontend)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HAProxy stack",
			"Could not update frontend, unexpected error: "+err.Error(),
		)
		return
	}

	// Update Frontend ACLs - this requires careful handling of order and changes
	if plan.Frontend.Acls != nil && len(plan.Frontend.Acls) > 0 {
		if err := r.updateACLs(ctx, "frontend", plan.Frontend.Name.ValueString(), plan.Frontend.Acls); err != nil {
			resp.Diagnostics.AddError(
				"Error updating frontend ACLs",
				fmt.Sprintf("Could not update frontend ACLs: %s", err.Error()),
			)
			return
		}
	}

	// Update Backend ACLs - this requires careful handling of order and changes
	if plan.Backend.Acls != nil && len(plan.Backend.Acls) > 0 {
		if err := r.updateACLs(ctx, "backend", plan.Backend.Name.ValueString(), plan.Backend.Acls); err != nil {
			resp.Diagnostics.AddError(
				"Error updating backend ACLs",
				fmt.Sprintf("Could not update backend ACLs: %s", err.Error()),
			)
			return
		}
	}

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// getAdvCheckValue safely extracts the adv_check value from the plan
func (r *haproxyStackResource) getAdvCheckValue(advCheck types.String) string {
	if advCheck.IsNull() || advCheck.IsUnknown() {
		return ""
	}
	return advCheck.ValueString()
}

// determineAdvCheckForAPI determines the correct adv_check value for the HAProxy API
// When httpchk_params is present, we need adv_check = "httpchk" for the API to accept it
// But we preserve the user's original intent by sending both configurations
func (r *haproxyStackResource) determineAdvCheckForAPI(advCheck types.String, httpchkParams []haproxyHttpchkParamsModel) string {
	// If httpchk_params is present, the API requires adv_check = "httpchk"
	// This is the only way to make both configurations work together
	if len(httpchkParams) > 0 {
		return "httpchk"
	}

	// Otherwise, use the user's configured adv_check value
	return r.getAdvCheckValue(advCheck)
}

// determineAdvCheck determines the correct adv_check value based on configuration
func (r *haproxyStackResource) determineAdvCheck(advCheck string, httpchkParams []haproxyHttpchkParamsModel) string {
	// If adv_check is explicitly configured, use it
	if advCheck != "" {
		return advCheck
	}

	// If httpchk_params is present, we need to set adv_check = "httpchk"
	// for HAProxy to apply the HTTP health check parameters
	if len(httpchkParams) > 0 {
		return "httpchk"
	}

	// Return empty string (no adv_check) if neither is configured
	return ""
}

// processBalanceBlock processes the balance block configuration
func (r *haproxyStackResource) processBalanceBlock(balance []haproxyBalanceModel) *Balance {
	if len(balance) == 0 {
		return nil
	}

	// Take the first balance configuration
	bal := balance[0]
	return &Balance{
		Algorithm: bal.Algorithm.ValueString(),
		UrlParam:  bal.UrlParam.ValueString(),
	}
}

// processHttpchkParamsBlock processes the httpchk_params block configuration
func (r *haproxyStackResource) processHttpchkParamsBlock(httpchkParams []haproxyHttpchkParamsModel) *HttpchkParams {
	if len(httpchkParams) == 0 {
		return nil
	}

	// Take the first httpchk_params configuration
	params := httpchkParams[0]
	return &HttpchkParams{
		Method:  params.Method.ValueString(),
		Uri:     params.Uri.ValueString(),
		Version: params.Version.ValueString(),
	}
}

// processForwardforBlock processes the forwardfor block configuration
func (r *haproxyStackResource) processForwardforBlock(forwardfor []haproxyForwardforModel) *ForwardFor {
	if len(forwardfor) == 0 {
		return nil
	}

	// Take the first forwardfor configuration
	ff := forwardfor[0]
	return &ForwardFor{
		Enabled: ff.Enabled.ValueString(),
	}
}

// updateACLs handles the complex logic of updating ACLs while maintaining order
func (r *haproxyStackResource) updateACLs(ctx context.Context, parentType string, parentName string, newAcls []haproxyAclModel) error {
	// Read existing ACLs from HAProxy
	existingAcls, err := r.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing ACLs: %w", err)
	}

	// Process new ACLs with proper indexing
	sortedNewAcls := r.processAclsBlock(newAcls)

	// Create maps for efficient lookup
	existingAclMap := make(map[string]*ACLPayload)
	for i := range existingAcls {
		existingAclMap[existingAcls[i].AclName] = &existingAcls[i]
	}

	// Track which ACLs we've processed to avoid duplicates
	processedAcls := make(map[string]bool)

	// Track ACLs that need to be recreated due to index changes
	var aclsToRecreate []haproxyAclModel

	// First pass: identify ACLs that need index changes and mark them for recreation
	// Also detect renames by matching content and position swaps
	for _, newAcl := range sortedNewAcls {
		newAclName := newAcl.AclName.ValueString()
		newAclIndex := newAcl.Index.ValueInt64()
		newAclContent := fmt.Sprintf("%s:%s", newAcl.Criterion.ValueString(), newAcl.Value.ValueString())

		// First check if this ACL exists by name
		if existingAcl, exists := existingAclMap[newAclName]; exists {
			// Check if the index has changed
			if existingAcl.Index != newAclIndex {
				// Index has changed - check if this is just a position swap
				// Look for another ACL that might have moved to this ACL's old position
				var isPositionSwap bool
				var swappedAcl *ACLPayload

				for _, otherExistingAcl := range existingAcls {
					if otherExistingAcl.AclName != newAclName && !processedAcls[otherExistingAcl.AclName] {
						otherContent := fmt.Sprintf("%s:%s", otherExistingAcl.Criterion, otherExistingAcl.Value)
						if otherContent == newAclContent {
							// This is a position swap - same content, different positions
							isPositionSwap = true
							swappedAcl = &otherExistingAcl
							break
						}
					}
				}

				if isPositionSwap {
					// This is a position swap, not a real change
					log.Printf("ACL '%s' position swapped with '%s' (same content, different positions) - no update needed",
						newAclName, swappedAcl.AclName)
					// Mark both as processed since they're just swapped
					processedAcls[newAclName] = true
					processedAcls[swappedAcl.AclName] = true
				} else {
					// Real index change, mark for recreation
					log.Printf("ACL '%s' index changed from %d to %d, will recreate",
						newAclName, existingAcl.Index, newAclIndex)
					aclsToRecreate = append(aclsToRecreate, newAcl)
				}
			} else if existingAcl.Criterion != newAcl.Criterion.ValueString() || existingAcl.Value != newAcl.Value.ValueString() {
				// Index is the same but content has changed, update in place
				aclPayload := ACLPayload{
					AclName:   newAcl.AclName.ValueString(),
					Criterion: newAcl.Criterion.ValueString(),
					Value:     newAcl.Value.ValueString(),
					Index:     existingAcl.Index, // Keep the existing index
				}

				log.Printf("Updating existing ACL '%s' at index %d", aclPayload.AclName, aclPayload.Index)
				err := r.client.UpdateAcl(ctx, existingAcl.Index, parentType, parentName, &aclPayload)
				if err != nil {
					return fmt.Errorf("failed to update ACL '%s': %w", aclPayload.AclName, err)
				}
			} else {
				// ACL is identical, no changes needed
				log.Printf("ACL '%s' at index %d is unchanged", newAclName, existingAcl.Index)
			}

			// Mark this ACL as processed
			processedAcls[newAclName] = true
		} else {
			// ACL doesn't exist by name, check if it's a rename by matching content
			var renamedAcl *ACLPayload
			for _, existingAcl := range existingAcls {
				existingContent := fmt.Sprintf("%s:%s", existingAcl.Criterion, existingAcl.Value)
				if existingContent == newAclContent && !processedAcls[existingAcl.AclName] {
					// This is a rename - same content, different name
					renamedAcl = &existingAcl
					break
				}
			}

			if renamedAcl != nil {
				// This is a rename, update the name while keeping the same index and content
				log.Printf("ACL renamed from '%s' to '%s' at index %d", renamedAcl.AclName, newAclName, renamedAcl.Index)
				aclPayload := ACLPayload{
					AclName:   newAclName,
					Criterion: renamedAcl.Criterion,
					Value:     renamedAcl.Value,
					Index:     renamedAcl.Index, // Keep the same index
				}

				err := r.client.UpdateAcl(ctx, renamedAcl.Index, parentType, parentName, &aclPayload)
				if err != nil {
					return fmt.Errorf("failed to rename ACL from '%s' to '%s': %w", renamedAcl.AclName, newAclName, err)
				}

				// Mark both as processed
				processedAcls[renamedAcl.AclName] = true
				processedAcls[newAclName] = true
			} else {
				// This is a completely new ACL, mark for creation
				log.Printf("ACL '%s' is new, will create", newAclName)
			}
		}
	}

	// Second pass: delete all ACLs that need to be recreated (due to index changes)
	// Delete in reverse order (highest index first) to avoid shifting issues
	for _, newAcl := range aclsToRecreate {
		newAclName := newAcl.AclName.ValueString()
		if existingAcl, exists := existingAclMap[newAclName]; exists {
			log.Printf("Deleting ACL '%s' at old index %d for recreation", newAclName, existingAcl.Index)
			err := r.client.DeleteAcl(ctx, existingAcl.Index, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete ACL '%s' at old index %d: %w", newAclName, existingAcl.Index, err)
			}
		}
	}

	// Third pass: create all ACLs that need to be recreated at their new positions
	// Use the user-specified index to maintain order
	for _, newAcl := range aclsToRecreate {
		newAclName := newAcl.AclName.ValueString()
		newAclIndex := newAcl.Index.ValueInt64()

		log.Printf("Creating ACL '%s' at user-specified index %d", newAclName, newAclIndex)
		aclPayload := ACLPayload{
			AclName:   newAcl.AclName.ValueString(),
			Criterion: newAcl.Criterion.ValueString(),
			Value:     newAcl.Value.ValueString(),
			Index:     newAclIndex, // Use the user-specified index
		}

		err = r.client.CreateAcl(ctx, parentType, parentName, &aclPayload)
		if err != nil {
			return fmt.Errorf("failed to create ACL '%s' at index %d: %w", newAclName, newAclIndex)
		}
	}

	// Delete ACLs that are no longer needed (not in the new configuration)
	// Delete in reverse order (highest index first) to avoid shifting issues
	var aclsToDelete []ACLPayload
	for _, existingAcl := range existingAcls {
		if !processedAcls[existingAcl.AclName] {
			aclsToDelete = append(aclsToDelete, existingAcl)
		}
	}

	// Sort by index in descending order (highest first)
	sort.Slice(aclsToDelete, func(i, j int) bool {
		return aclsToDelete[i].Index > aclsToDelete[j].Index
	})

	// Delete ACLs in reverse order
	for _, aclToDelete := range aclsToDelete {
		log.Printf("Deleting ACL '%s' at index %d (no longer needed)", aclToDelete.AclName, aclToDelete.Index)
		err := r.client.DeleteAcl(ctx, aclToDelete.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete ACL '%s': %w", aclToDelete.AclName, err)
		}
	}

	// Create new ACLs that don't exist yet
	// Use the user-specified index to maintain order
	for _, newAcl := range sortedNewAcls {
		newAclName := newAcl.AclName.ValueString()
		if !processedAcls[newAclName] {
			// This is a new ACL, create it with the user-specified index
			newAclIndex := newAcl.Index.ValueInt64()
			aclPayload := ACLPayload{
				AclName:   newAcl.AclName.ValueString(),
				Criterion: newAcl.Criterion.ValueString(),
				Value:     newAcl.Value.ValueString(),
				Index:     newAclIndex, // Use the user-specified index
			}

			log.Printf("Creating new ACL '%s' at index %d", aclPayload.AclName, aclPayload.Index)
			err := r.client.CreateAcl(ctx, parentType, parentName, &aclPayload)
			if err != nil {
				return fmt.Errorf("failed to create ACL '%s': %w", aclPayload.AclName, err)
			}
		}
	}

	return nil
}

// processAclsBlock processes the ACLs block configuration with proper index normalization
// This function normalizes user-specified indices to sequential order (0, 1, 2, 3...)
// while preserving the user's intended ACL sequence
func (r *haproxyStackResource) processAclsBlock(acls []haproxyAclModel) []haproxyAclModel {
	if len(acls) == 0 {
		return nil
	}

	// Create a copy to avoid modifying the original
	normalizedAcls := make([]haproxyAclModel, len(acls))
	copy(normalizedAcls, acls)

	// Sort ACLs by user-specified index to determine the intended order
	sort.Slice(normalizedAcls, func(i, j int) bool {
		indexI := normalizedAcls[i].Index.ValueInt64()
		indexJ := normalizedAcls[j].Index.ValueInt64()
		return indexI < indexJ
	})

	// DO NOT normalize indices - preserve user's exact configuration
	// HAProxy can handle non-sequential indices, and normalization causes state drift
	log.Printf("ACL order preserved as configured: %s", r.formatAclOrder(normalizedAcls))

	return normalizedAcls
}

// formatAclOrder creates a readable string showing ACL order for logging
func (r *haproxyStackResource) formatAclOrder(acls []haproxyAclModel) string {
	if len(acls) == 0 {
		return "none"
	}

	var order []string
	for _, acl := range acls {
		order = append(order, fmt.Sprintf("%s(index:%d)", acl.AclName.ValueString(), acl.Index.ValueInt64()))
	}
	return strings.Join(order, " → ")
}

// translateNoTlsToForceTls translates no_tlsv* fields to force_tlsv* fields
// Based on your example:
// no_tlsv10 = "enabled" → HAProxy creates force_tlsv10: "disabled"
// no_tlsv13 = "enabled" → HAProxy creates force_tlsv13: "disabled"
func (r *haproxyStackResource) translateNoTlsToForceTls(noTlsValue string) string {
	if noTlsValue == "enabled" {
		return "disabled" // "Don't allow TLSv1.x" → "Force disabled"
	} else if noTlsValue == "disabled" {
		return "enabled" // "Allow TLSv1.x" → "Force enabled"
	}
	return noTlsValue // Return as-is for other values
}

// translateForceTlsToNoTls translates force_tlsv* fields back to no_tlsv* fields
// This is the reverse of translateNoTlsToForceTls
// force_tlsv10: "disabled" → no_tlsv10 = "enabled"
// force_tlsv13: "disabled" → no_tlsv13 = "enabled"
func (r *haproxyStackResource) translateForceTlsToNoTls(forceTlsValue string) string {
	if forceTlsValue == "disabled" {
		return "enabled" // "Force disabled" → "Don't allow TLSv1.x"
	} else if forceTlsValue == "enabled" {
		return "disabled" // "Force enabled" → "Allow TLSv1.x"
	}
	return forceTlsValue // Return as-is for other values
}

// validateSSLConfiguration validates SSL/TLS configuration based on HAProxy API version
func (r *haproxyStackResource) validateSSLConfiguration(defaultServer *haproxyDefaultServerModel, client *HAProxyClient) error {
	// If no DefaultServer configuration, nothing to validate
	if defaultServer == nil {
		return nil
	}

	// Get the API version from the client
	apiVersion := client.GetAPIVersion()
	if apiVersion == "" {
		apiVersion = "v2" // Default to v2 if not specified
	}

	// v3-only fields that are not supported in v2
	v3OnlyFields := []struct {
		name  string
		value types.String
	}{
		{"sslv3", defaultServer.Sslv3},
		{"tlsv10", defaultServer.Tlsv10},
		{"tlsv11", defaultServer.Tlsv11},
		{"tlsv12", defaultServer.Tlsv12},
		{"tlsv13", defaultServer.Tlsv13},
	}

	// Check if we're using v3-only fields with v2 API
	if apiVersion == "v2" {
		var v3FieldErrors []string
		for _, field := range v3OnlyFields {
			if !field.value.IsNull() && field.value.ValueString() != "" {
				v3FieldErrors = append(v3FieldErrors, fmt.Sprintf("Field '%s' is only supported in Data Plane API v3, but you are using v2", field.name))
			}
		}

		if len(v3FieldErrors) > 0 {
			return fmt.Errorf("SSL/TLS configuration validation failed:\n%s\n\nTo fix this:\n1. Upgrade to HAProxy Data Plane API v3, OR\n2. Remove the unsupported fields from your configuration", strings.Join(v3FieldErrors, "\n"))
		}
	}

	return nil
}

// Delete resource.
func (r *haproxyStackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state haproxyStackResourceModel
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

	// Delete ALL resources in a single transaction with robust error handling
	// This ensures atomic deletion: ACLs → Frontend → Servers → Backend
	// The transaction will handle missing ACLs gracefully and retry on concurrency issues

	log.Printf("Starting deletion of all resources in single transaction")

	// Prepare the complete resources payload for deletion
	resources := &AllResourcesPayload{
		Frontend: &FrontendPayload{
			Name: state.Frontend.Name.ValueString(),
		},
		Backend: &BackendPayload{
			Name: state.Backend.Name.ValueString(),
		},
	}

	// Add Frontend ACLs if they exist in the state
	if state.Frontend.Acls != nil && len(state.Frontend.Acls) > 0 {
		log.Printf("Including %d frontend ACLs in deletion transaction", len(state.Frontend.Acls))

		// Read existing ACLs to get their current indices
		existingAcls, err := r.client.ReadACLs(ctx, "frontend", state.Frontend.Name.ValueString())
		if err == nil && len(existingAcls) > 0 {
			// Map existing ACLs to ACLResource format for transaction
			acls := make([]ACLResource, len(existingAcls))
			for i, acl := range existingAcls {
				acls[i] = ACLResource{
					ParentType: "frontend",
					ParentName: state.Frontend.Name.ValueString(),
					Payload:    &acl,
				}
			}
			resources.Acls = append(resources.Acls, acls...)
			log.Printf("Successfully mapped %d existing frontend ACLs for deletion", len(acls))
		} else {
			log.Printf("No existing frontend ACLs found in HAProxy (state had %d ACLs): %v", len(state.Frontend.Acls), err)
		}
	}

	// Add Backend ACLs if they exist in the state
	if state.Backend.Acls != nil && len(state.Backend.Acls) > 0 {
		log.Printf("Including %d backend ACLs in deletion transaction", len(state.Backend.Acls))

		// Read existing ACLs to get their current indices
		existingAcls, err := r.client.ReadACLs(ctx, "backend", state.Backend.Name.ValueString())
		if err == nil && len(existingAcls) > 0 {
			// Map existing ACLs to ACLResource format for transaction
			acls := make([]ACLResource, len(existingAcls))
			for i, acl := range existingAcls {
				acls[i] = ACLResource{
					ParentType: "backend",
					ParentName: state.Backend.Name.ValueString(),
					Payload:    &acl,
				}
			}
			resources.Acls = append(resources.Acls, acls...)
			log.Printf("Successfully mapped %d existing backend ACLs for deletion", len(acls))
		} else {
			log.Printf("No existing backend ACLs found in HAProxy (state had %d ACLs): %v", len(state.Backend.Acls), err)
		}
	}

	// Delete everything in one transaction with retry logic
	err := r.client.DeleteAllResourcesInSingleTransaction(ctx, resources)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting resources",
			"Could not delete resources in transaction, unexpected error: "+err.Error(),
		)
		return
	}

	log.Printf("Successfully deleted all resources in single transaction")

	resp.State.RemoveResource(ctx)
}
