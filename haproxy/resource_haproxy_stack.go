package haproxy

import (
	"context"
	"fmt"
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
	Acl                []haproxyAclModel              `tfsdk:"acl"`
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
					"acl": schema.ListNestedBlock{
						Description: "Access Control List configuration for the backend.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"acl_name": schema.StringAttribute{
									Required:    true,
									Description: "The acl name. Pattern: ^[^\\s]+$",
								},
								"index": schema.Int64Attribute{
									Required:    true,
									Description: "The index of the acl.",
								},
								"criterion": schema.StringAttribute{
									Required:    true,
									Description: "The criterion. Pattern: ^[^\\s]+$",
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "The value of the criterion.",
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
				},
			},
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
	if err := r.validateSSLConfiguration(plan.Backend.DefaultServer); err != nil {
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
			AdvCheck:           plan.Backend.AdvCheck.ValueString(),
			HttpConnectionMode: plan.Backend.HttpConnectionMode.ValueString(),
			ServerTimeout:      plan.Backend.ServerTimeout.ValueInt64(),
			CheckTimeout:       plan.Backend.CheckTimeout.ValueInt64(),
			ConnectTimeout:     plan.Backend.ConnectTimeout.ValueInt64(),
			QueueTimeout:       plan.Backend.QueueTimeout.ValueInt64(),
			TunnelTimeout:      plan.Backend.TunnelTimeout.ValueInt64(),
			TarpitTimeout:      plan.Backend.TarpitTimeout.ValueInt64(),
			CheckCache:         plan.Backend.Checkcache.ValueString(),
			Retries:            plan.Backend.Retries.ValueInt64(),
			DefaultServer: &DefaultServerPayload{
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

				// Deprecated fields (v2 only)
				NoSslv3:  plan.Backend.DefaultServer.NoSslv3.ValueString(),
				NoTlsv10: plan.Backend.DefaultServer.NoTlsv10.ValueString(),
				NoTlsv11: plan.Backend.DefaultServer.NoTlsv11.ValueString(),
				NoTlsv12: plan.Backend.DefaultServer.NoTlsv12.ValueString(),
				NoTlsv13: plan.Backend.DefaultServer.NoTlsv13.ValueString(),

				// Force fields (v3 only)
				ForceSslv3:     plan.Backend.DefaultServer.ForceSslv3.ValueString(),
				ForceTlsv10:    plan.Backend.DefaultServer.ForceTlsv10.ValueString(),
				ForceTlsv11:    plan.Backend.DefaultServer.ForceTlsv11.ValueString(),
				ForceTlsv12:    plan.Backend.DefaultServer.ForceTlsv12.ValueString(),
				ForceTlsv13:    plan.Backend.DefaultServer.ForceTlsv13.ValueString(),
				ForceStrictSni: plan.Backend.DefaultServer.ForceStrictSni.ValueString(),
			},
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

	// Create all resources in single transaction
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

	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"HAProxy client has not been configured; please report this issue to the provider developer",
		)
		return
	}

	// Read all resources from HAProxy
	backend, err := r.client.ReadBackend(ctx, state.Backend.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backend",
			"Could not read backend, unexpected error: "+err.Error(),
		)
		return
	}

	servers, err := r.client.ReadServers(ctx, "backend", state.Backend.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading servers",
			"Could not read servers, unexpected error: "+err.Error(),
		)
		return
	}

	frontend, err := r.client.ReadFrontend(ctx, state.Frontend.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading frontend",
			"Could not read frontend, unexpected error: "+err.Error(),
		)
		return
	}

	// Update state with actual HAProxy configuration
	if backend != nil {
		state.Backend.Mode = types.StringValue(backend.Mode)
		state.Backend.AdvCheck = types.StringValue(backend.AdvCheck)
		state.Backend.HttpConnectionMode = types.StringValue(backend.HttpConnectionMode)
		state.Backend.ServerTimeout = types.Int64Value(backend.ServerTimeout)
		state.Backend.CheckTimeout = types.Int64Value(backend.CheckTimeout)
		state.Backend.ConnectTimeout = types.Int64Value(backend.ConnectTimeout)
		state.Backend.QueueTimeout = types.Int64Value(backend.QueueTimeout)
		state.Backend.TunnelTimeout = types.Int64Value(backend.TunnelTimeout)
		state.Backend.TarpitTimeout = types.Int64Value(backend.TarpitTimeout)
		state.Backend.Checkcache = types.StringValue(backend.CheckCache)
		state.Backend.Retries = types.Int64Value(backend.Retries)

		// Handle default_server configuration
		if backend.DefaultServer != nil {
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
			if backend.DefaultServer.NoSslv3 != "" {
				state.Backend.DefaultServer.NoSslv3 = types.StringValue(backend.DefaultServer.NoSslv3)
			}
			if backend.DefaultServer.NoTlsv10 != "" {
				state.Backend.DefaultServer.NoTlsv10 = types.StringValue(backend.DefaultServer.NoTlsv10)
			}
			if backend.DefaultServer.NoTlsv11 != "" {
				state.Backend.DefaultServer.NoTlsv11 = types.StringValue(backend.DefaultServer.NoTlsv11)
			}
			if backend.DefaultServer.NoTlsv12 != "" {
				state.Backend.DefaultServer.NoTlsv12 = types.StringValue(backend.DefaultServer.NoTlsv12)
			}
			if backend.DefaultServer.NoTlsv13 != "" {
				state.Backend.DefaultServer.NoTlsv13 = types.StringValue(backend.DefaultServer.NoTlsv13)
			}
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
	}

	if len(servers) > 0 && len(servers) > 0 {
		server := servers[0] // Take first server
		state.Server.Name = types.StringValue(server.Name)
		state.Server.Address = types.StringValue(server.Address)
		state.Server.Port = types.Int64Value(server.Port)
		state.Server.Check = types.StringValue(server.Check)
		state.Server.Backup = types.StringValue(server.Backup)
		state.Server.Maxconn = types.Int64Value(server.Maxconn)
		state.Server.Weight = types.Int64Value(server.Weight)
		state.Server.Rise = types.Int64Value(server.Rise)
		state.Server.Fall = types.Int64Value(server.Fall)
		state.Server.Inter = types.Int64Value(server.Inter)
		state.Server.Fastinter = types.Int64Value(server.Fastinter)
		state.Server.Downinter = types.Int64Value(server.Downinter)
		state.Server.Ssl = types.StringValue(server.Ssl)
		state.Server.Verify = types.StringValue(server.Verify)
		state.Server.Cookie = types.StringValue(server.Cookie)
		state.Server.Disabled = types.BoolValue(server.Disabled)
	}

	if frontend != nil {
		state.Frontend.Mode = types.StringValue(frontend.Mode)
		state.Frontend.DefaultBackend = types.StringValue(frontend.DefaultBackend)
		state.Frontend.Maxconn = types.Int64Value(frontend.MaxConn)
		state.Frontend.Backlog = types.Int64Value(frontend.Backlog)
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
	if err := r.validateSSLConfiguration(plan.Backend.DefaultServer); err != nil {
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
			AdvCheck:           plan.Backend.AdvCheck.ValueString(),
			HttpConnectionMode: plan.Backend.HttpConnectionMode.ValueString(),
			ServerTimeout:      plan.Backend.ServerTimeout.ValueInt64(),
			CheckTimeout:       plan.Backend.CheckTimeout.ValueInt64(),
			ConnectTimeout:     plan.Backend.ConnectTimeout.ValueInt64(),
			QueueTimeout:       plan.Backend.QueueTimeout.ValueInt64(),
			TunnelTimeout:      plan.Backend.TunnelTimeout.ValueInt64(),
			TarpitTimeout:      plan.Backend.TarpitTimeout.ValueInt64(),
			CheckCache:         plan.Backend.Checkcache.ValueString(),
			Retries:            plan.Backend.Retries.ValueInt64(),
			DefaultServer: &DefaultServerPayload{
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
				NoTlsv10:       plan.Backend.DefaultServer.NoTlsv10.ValueString(),
				NoTlsv11:       plan.Backend.DefaultServer.NoTlsv11.ValueString(),
				NoTlsv12:       plan.Backend.DefaultServer.NoTlsv12.ValueString(),
				NoTlsv13:       plan.Backend.DefaultServer.NoTlsv13.ValueString(),
				ForceSslv3:     plan.Backend.DefaultServer.ForceSslv3.ValueString(),
				ForceTlsv10:    plan.Backend.DefaultServer.ForceTlsv10.ValueString(),
				ForceTlsv11:    plan.Backend.DefaultServer.ForceTlsv11.ValueString(),
				ForceTlsv12:    plan.Backend.DefaultServer.ForceTlsv12.ValueString(),
				ForceTlsv13:    plan.Backend.DefaultServer.ForceTlsv13.ValueString(),
				ForceStrictSni: plan.Backend.DefaultServer.ForceStrictSni.ValueString(),
			},
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

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// validateSSLConfiguration validates SSL/TLS configuration based on HAProxy API version
func (r *haproxyStackResource) validateSSLConfiguration(defaultServer *haproxyDefaultServerModel) error {
	// TODO: Detect HAProxy API version dynamically
	// For now, we'll use a configuration flag or environment variable
	// This should be enhanced to detect the actual API version from HAProxy

	// Note: v2-only fields (no_sslv3, no_tlsv10, etc.) are supported in v2
	// but we don't need to validate them as they're always allowed

	// Check if we're using v3-only fields
	v3Fields := []struct {
		name  string
		value types.String
	}{
		{"sslv3", defaultServer.Sslv3},
		{"tlsv10", defaultServer.Tlsv10},
		{"tlsv11", defaultServer.Tlsv11},
		{"tlsv12", defaultServer.Tlsv12},
		{"tlsv13", defaultServer.Tlsv13},
		{"force_sslv3", defaultServer.ForceSslv3},
		{"force_tlsv10", defaultServer.ForceTlsv10},
		{"force_tlsv11", defaultServer.ForceTlsv11},
		{"force_tlsv12", defaultServer.ForceTlsv12},
		{"force_tlsv13", defaultServer.ForceTlsv13},
		{"force_strict_sni", defaultServer.ForceStrictSni},
	}

	// For now, we'll fail if v3 fields are used (assuming v2)
	// This should be configurable or auto-detected
	var v3FieldErrors []string
	for _, field := range v3Fields {
		if !field.value.IsNull() && field.value.ValueString() != "" {
			v3FieldErrors = append(v3FieldErrors, fmt.Sprintf("Field '%s' is only supported in Data Plane API v3, but you appear to be using v2", field.name))
		}
	}

	if len(v3FieldErrors) > 0 {
		return fmt.Errorf("SSL/TLS configuration validation failed:\n%s\n\nTo fix this:\n1. Upgrade to HAProxy Data Plane API v3, OR\n2. Remove the unsupported fields from your configuration", strings.Join(v3FieldErrors, "\n"))
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

	// Delete all resources in proper order (frontend → backend → server)
	// Delete frontend first
	err := r.client.DeleteFrontend(ctx, state.Frontend.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting frontend",
			"Could not delete frontend, unexpected error: "+err.Error(),
		)
		return
	}

	// Delete backend (this will also delete associated servers)
	err = r.client.DeleteBackend(ctx, state.Backend.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backend",
			"Could not delete backend, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
