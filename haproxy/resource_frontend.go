package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &frontendResource{}
)

// NewFrontendResource is a helper function to simplify the provider implementation.
func NewFrontendResource() resource.Resource {
	return &frontendResource{}
}

// frontendResource is the resource implementation.
type frontendResource struct {
	client *HAProxyClient
}

// frontendResourceModel maps the resource schema data.
type frontendResourceModel struct {
	Name                     types.String `tfsdk:"name"`
	DefaultBackend           types.String `tfsdk:"default_backend"`
	HttpConnectionMode       types.String `tfsdk:"http_connection_mode"`
	AcceptInvalidHttpRequest types.String `tfsdk:"accept_invalid_http_request"`
	MaxConn                  types.Int64  `tfsdk:"maxconn"`
	Mode                     types.String `tfsdk:"mode"`
	Backlog                  types.Int64  `tfsdk:"backlog"`
	HttpKeepAliveTimeout     types.Int64  `tfsdk:"http_keep_alive_timeout"`
	HttpRequestTimeout       types.Int64  `tfsdk:"http_request_timeout"`
	HttpUseProxyHeader       types.String `tfsdk:"http_use_proxy_header"`
	HttpLog                  types.Bool   `tfsdk:"httplog"`
	HttpsLog                 types.String `tfsdk:"httpslog"`
	ErrorLogFormat           types.String `tfsdk:"error_log_format"`
	LogFormat                types.String `tfsdk:"log_format"`
	LogFormatSd              types.String `tfsdk:"log_format_sd"`
	MonitorUri               types.String `tfsdk:"monitor_uri"`
	TcpLog                   types.Bool   `tfsdk:"tcplog"`
	From                     types.String `tfsdk:"from"`
	Binds                    types.List   `tfsdk:"bind"`
	MonitorFail              types.List   `tfsdk:"monitor_fail"`
	Acls                     types.List   `tfsdk:"acl"`
	HttpRequestRules         types.List   `tfsdk:"httprequestrule"`
	HttpResponseRules        types.List   `tfsdk:"httpresponserule"`
	TcpRequestRules          types.List   `tfsdk:"tcprequestrule"`
	TcpResponseRules         types.List   `tfsdk:"tcpresponserule"`
	ClientTimeout            types.Int64  `tfsdk:"client_timeout"`
	HttpUseHtx               types.String `tfsdk:"http_use_htx"`
	HttpIgnoreProbes         types.String `tfsdk:"http_ignore_probes"`
	LogTag                   types.String `tfsdk:"log_tag"`
	Clflog                   types.Bool   `tfsdk:"clflog"`
	Contstats                types.String `tfsdk:"contstats"`
	Dontlognull              types.String `tfsdk:"dontlognull"`
	LogSeparateErrors        types.String `tfsdk:"log_separate_errors"`
	OptionHttpServerClose    types.String `tfsdk:"option_http_server_close"`
	OptionHttpclose          types.String `tfsdk:"option_httpclose"`
	OptionHttpKeepAlive      types.String `tfsdk:"option_http_keep_alive"`
	OptionDontlogNormal      types.String `tfsdk:"option_dontlog_normal"`
	OptionLogasap            types.String `tfsdk:"option_logasap"`
	OptionTcplog             types.String `tfsdk:"option_tcplog"`
	OptionSocketStats        types.String `tfsdk:"option_socket_stats"`
	OptionForwardfor         types.String `tfsdk:"option_forwardfor"`
	TimeoutClient            types.Int64  `tfsdk:"timeout_client"`
	TimeoutHttpKeepAlive     types.Int64  `tfsdk:"timeout_http_keep_alive"`
	TimeoutHttpRequest       types.Int64  `tfsdk:"timeout_http_request"`
	TimeoutCont              types.Int64  `tfsdk:"timeout_cont"`
	TimeoutTarpit            types.Int64  `tfsdk:"timeout_tarpit"`
	StatsOptions             types.Object `tfsdk:"stats_options"`
}

// bindResourceModel maps the resource schema data.
type bindResourceModel struct {
	Name                 types.String `tfsdk:"name"`
	Port                 types.Int64  `tfsdk:"port"`
	PortRangeEnd         types.Int64  `tfsdk:"port_range_end"`
	Address              types.String `tfsdk:"address"`
	Transparent          types.Bool   `tfsdk:"transparent"`
	Mode                 types.String `tfsdk:"mode"`
	Maxconn              types.Int64  `tfsdk:"maxconn"`
	User                 types.String `tfsdk:"user"`
	Group                types.String `tfsdk:"group"`
	ForceSslv3           types.Bool   `tfsdk:"force_sslv3"`
	ForceTlsv10          types.Bool   `tfsdk:"force_tlsv10"`
	ForceTlsv11          types.Bool   `tfsdk:"force_tlsv11"`
	ForceTlsv12          types.Bool   `tfsdk:"force_tlsv12"`
	ForceTlsv13          types.Bool   `tfsdk:"force_tlsv13"`
	ForceStrictSni       types.String `tfsdk:"force_strict_sni"`
	Ssl                  types.Bool   `tfsdk:"ssl"`
	SslCafile            types.String `tfsdk:"ssl_cafile"`
	SslMaxVer            types.String `tfsdk:"ssl_max_ver"`
	SslMinVer            types.String `tfsdk:"ssl_min_ver"`
	SslCertificate       types.String `tfsdk:"ssl_certificate"`
	Ciphers              types.String `tfsdk:"ciphers"`
	Ciphersuites         types.String `tfsdk:"ciphersuites"`
	AcceptProxy          types.Bool   `tfsdk:"accept_proxy"`
	Allow0rtt            types.Bool   `tfsdk:"allow_0rtt"`
	Alpn                 types.String `tfsdk:"alpn"`
	Backlog              types.String `tfsdk:"backlog"`
	CaIgnoreErr          types.String `tfsdk:"ca_ignore_err"`
	CaSignFile           types.String `tfsdk:"ca_sign_file"`
	CaSignPass           types.String `tfsdk:"ca_sign_pass"`
	CaVerifyFile         types.String `tfsdk:"ca_verify_file"`
	CrlFile              types.String `tfsdk:"crl_file"`
	CrtIgnoreErr         types.String `tfsdk:"crt_ignore_err"`
	CrtList              types.String `tfsdk:"crt_list"`
	DeferAccept          types.Bool   `tfsdk:"defer_accept"`
	ExposeViaAgent       types.Bool   `tfsdk:"expose_via_agent"`
	GenerateCertificates types.Bool   `tfsdk:"generate_certificates"`
	Gid                  types.Int64  `tfsdk:"gid"`
	Id                   types.String `tfsdk:"id"`
	Interface            types.String `tfsdk:"interface"`
	Level                types.String `tfsdk:"level"`
	LogProto             types.String `tfsdk:"log_proto"`
	Mdev                 types.String `tfsdk:"mdev"`
	Namespace            types.String `tfsdk:"namespace"`
	Nice                 types.Int64  `tfsdk:"nice"`
	NoCaNames            types.Bool   `tfsdk:"no_ca_names"`
	NoSslv3              types.Bool   `tfsdk:"no_sslv3"`
	NoTlsv10             types.Bool   `tfsdk:"no_tlsv10"`
	NoTlsv11             types.Bool   `tfsdk:"no_tlsv11"`
	NoTlsv12             types.Bool   `tfsdk:"no_tlsv12"`
	NoTlsv13             types.Bool   `tfsdk:"no_tlsv13"`
	// New v3 fields (non-deprecated)
	Sslv3  types.Bool `tfsdk:"sslv3"`
	Tlsv10 types.Bool `tfsdk:"tlsv10"`
	Tlsv11 types.Bool `tfsdk:"tlsv11"`
	Tlsv12 types.Bool `tfsdk:"tlsv12"`
	Tlsv13 types.Bool `tfsdk:"tlsv13"`

	Npn                 types.String `tfsdk:"npn"`
	PreferClientCiphers types.Bool   `tfsdk:"prefer_client_ciphers"`
	Process             types.String `tfsdk:"process"`
	Proto               types.String `tfsdk:"proto"`
	SeverityOutput      types.String `tfsdk:"severity_output"`
	StrictSni           types.Bool   `tfsdk:"strict_sni"`
	TcpUserTimeout      types.Int64  `tfsdk:"tcp_user_timeout"`
	Tfo                 types.Bool   `tfsdk:"tfo"`
	TlsTicketKeys       types.String `tfsdk:"tls_ticket_keys"`
	Uid                 types.String `tfsdk:"uid"`
	V4v6                types.Bool   `tfsdk:"v4v6"`
	V6only              types.Bool   `tfsdk:"v6only"`
	Verify              types.String `tfsdk:"verify"`
	Metadata            types.String `tfsdk:"metadata"`
}

func (b bindResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                  types.StringType,
		"port":                  types.Int64Type,
		"port_range_end":        types.Int64Type,
		"address":               types.StringType,
		"transparent":           types.BoolType,
		"mode":                  types.StringType,
		"maxconn":               types.Int64Type,
		"user":                  types.StringType,
		"group":                 types.StringType,
		"force_sslv3":           types.BoolType,
		"force_tlsv10":          types.BoolType,
		"force_tlsv11":          types.BoolType,
		"force_tlsv12":          types.BoolType,
		"force_tlsv13":          types.BoolType,
		"force_strict_sni":      types.StringType,
		"ssl":                   types.BoolType,
		"ssl_cafile":            types.StringType,
		"ssl_max_ver":           types.StringType,
		"ssl_min_ver":           types.StringType,
		"ssl_certificate":       types.StringType,
		"ciphers":               types.StringType,
		"ciphersuites":          types.StringType,
		"accept_proxy":          types.BoolType,
		"allow_0rtt":            types.BoolType,
		"alpn":                  types.StringType,
		"backlog":               types.StringType,
		"ca_ignore_err":         types.StringType,
		"ca_sign_file":          types.StringType,
		"ca_sign_pass":          types.StringType,
		"ca_verify_file":        types.StringType,
		"crl_file":              types.StringType,
		"crt_ignore_err":        types.StringType,
		"crt_list":              types.StringType,
		"defer_accept":          types.BoolType,
		"expose_via_agent":      types.BoolType,
		"generate_certificates": types.BoolType,
		"gid":                   types.Int64Type,
		"id":                    types.StringType,
		"interface":             types.StringType,
		"level":                 types.StringType,
		"log_proto":             types.StringType,
		"mdev":                  types.StringType,
		"namespace":             types.StringType,
		"nice":                  types.Int64Type,
		"no_ca_names":           types.BoolType,
		"no_sslv3":              types.BoolType,
		"no_tlsv10":             types.BoolType,
		"no_tlsv11":             types.BoolType,
		"no_tlsv12":             types.BoolType,
		"no_tlsv13":             types.BoolType,
		// New v3 fields
		"sslv3":  types.BoolType,
		"tlsv10": types.BoolType,
		"tlsv11": types.BoolType,
		"tlsv12": types.BoolType,
		"tlsv13": types.BoolType,

		"npn":                   types.StringType,
		"prefer_client_ciphers": types.BoolType,
		"process":               types.StringType,
		"proto":                 types.StringType,
		"severity_output":       types.StringType,
		"strict_sni":            types.BoolType,
		"tcp_user_timeout":      types.Int64Type,
		"tfo":                   types.BoolType,
		"tls_ticket_keys":       types.StringType,
		"uid":                   types.StringType,
		"v4v6":                  types.BoolType,
		"v6only":                types.BoolType,
		"verify":                types.StringType,
		"metadata":              types.StringType,
	}
}

// frontendAclResourceModel maps the resource schema data.
type frontendAclResourceModel struct {
	AclName   types.String `tfsdk:"acl_name"`
	Index     types.Int64  `tfsdk:"index"`
	Criterion types.String `tfsdk:"criterion"`
	Value     types.String `tfsdk:"value"`
}

func (a frontendAclResourceModel) GetIndex() int64 {
	return a.Index.ValueInt64()
}

func (a frontendAclResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"acl_name":  types.StringType,
		"index":     types.Int64Type,
		"criterion": types.StringType,
		"value":     types.StringType,
	}
}

// httpRequestRuleResourceModel maps the resource schema data.
type httpRequestRuleResourceModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Type         types.String `tfsdk:"type"`
	Cond         types.String `tfsdk:"cond"`
	CondTest     types.String `tfsdk:"cond_test"`
	HdrName      types.String `tfsdk:"hdr_name"`
	HdrFormat    types.String `tfsdk:"hdr_format"`
	RedirType    types.String `tfsdk:"redir_type"`
	RedirValue   types.String `tfsdk:"redir_value"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	StatusReason types.String `tfsdk:"status_reason"`
}

func (h httpRequestRuleResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h httpRequestRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":         types.Int64Type,
		"type":          types.StringType,
		"cond":          types.StringType,
		"cond_test":     types.StringType,
		"hdr_name":      types.StringType,
		"hdr_format":    types.StringType,
		"redir_type":    types.StringType,
		"redir_value":   types.StringType,
		"status_code":   types.Int64Type,
		"status_reason": types.StringType,
	}
}

// httpResponseRuleResourceModel maps the resource schema data.
type httpResponseRuleResourceModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Type         types.String `tfsdk:"type"`
	Cond         types.String `tfsdk:"cond"`
	CondTest     types.String `tfsdk:"cond_test"`
	HdrName      types.String `tfsdk:"hdr_name"`
	HdrFormat    types.String `tfsdk:"hdr_format"`
	RedirType    types.String `tfsdk:"redir_type"`
	RedirValue   types.String `tfsdk:"redir_value"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	StatusReason types.String `tfsdk:"status_reason"`
}

func (h httpResponseRuleResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h httpResponseRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":         types.Int64Type,
		"type":          types.StringType,
		"cond":          types.StringType,
		"cond_test":     types.StringType,
		"hdr_name":      types.StringType,
		"hdr_format":    types.StringType,
		"redir_type":    types.StringType,
		"redir_value":   types.StringType,
		"status_code":   types.Int64Type,
		"status_reason": types.StringType,
	}
}

// frontendTcpRequestRuleResourceModel maps the resource schema data.
type frontendTcpRequestRuleResourceModel struct {
	Index        types.Int64  `tfsdk:"index"`
	Type         types.String `tfsdk:"type"`
	Action       types.String `tfsdk:"action"`
	Cond         types.String `tfsdk:"cond"`
	CondTest     types.String `tfsdk:"cond_test"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	LuaAction    types.String `tfsdk:"lua_action"`
	LuaParams    types.String `tfsdk:"lua_params"`
	ScId         types.Int64  `tfsdk:"sc_id"`
	ScIdx        types.Int64  `tfsdk:"sc_idx"`
	ScInt        types.Int64  `tfsdk:"sc_int"`
	ScIncGpc0    types.String `tfsdk:"sc_inc_gpc0"`
	ScIncGpc1    types.String `tfsdk:"sc_inc_gpc1"`
	ScSetGpt0    types.String `tfsdk:"sc_set_gpt0"`
	TrackScKey   types.String `tfsdk:"track_sc_key"`
	TrackScTable types.String `tfsdk:"track_sc_table"`
	VarName      types.String `tfsdk:"var_name"`
	VarScope     types.String `tfsdk:"var_scope"`
	VarExpr      types.String `tfsdk:"var_expr"`
	VarFormat    types.String `tfsdk:"var_format"`
	VarType      types.String `tfsdk:"var_type"`
}

func (t frontendTcpRequestRuleResourceModel) GetIndex() int64 {
	return t.Index.ValueInt64()
}

func (t frontendTcpRequestRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":          types.Int64Type,
		"type":           types.StringType,
		"action":         types.StringType,
		"cond":           types.StringType,
		"cond_test":      types.StringType,
		"timeout":        types.Int64Type,
		"lua_action":     types.StringType,
		"lua_params":     types.StringType,
		"sc_id":          types.Int64Type,
		"sc_idx":         types.Int64Type,
		"sc_int":         types.Int64Type,
		"sc_inc_gpc0":    types.StringType,
		"sc_inc_gpc1":    types.StringType,
		"sc_set_gpt0":    types.StringType,
		"track_sc_key":   types.StringType,
		"track_sc_table": types.StringType,
		"var_name":       types.StringType,
		"var_scope":      types.StringType,
		"var_expr":       types.StringType,
		"var_format":     types.StringType,
		"var_type":       types.StringType,
	}
}

// frontendTcpResponseRuleResourceModel maps the resource schema data.
type frontendTcpResponseRuleResourceModel struct {
	Index     types.Int64  `tfsdk:"index"`
	Action    types.String `tfsdk:"action"`
	Cond      types.String `tfsdk:"cond"`
	CondTest  types.String `tfsdk:"cond_test"`
	LuaAction types.String `tfsdk:"lua_action"`
	LuaParams types.String `tfsdk:"lua_params"`
	ScId      types.Int64  `tfsdk:"sc_id"`
	ScIdx     types.Int64  `tfsdk:"sc_idx"`
	ScInt     types.Int64  `tfsdk:"sc_int"`
	ScIncGpc0 types.String `tfsdk:"sc_inc_gpc0"`
	ScIncGpc1 types.String `tfsdk:"sc_inc_gpc1"`
	ScSetGpt0 types.String `tfsdk:"sc_set_gpt0"`
	VarName   types.String `tfsdk:"var_name"`
	VarScope  types.String `tfsdk:"var_scope"`
	VarExpr   types.String `tfsdk:"var_expr"`
	VarFormat types.String `tfsdk:"var_format"`
	VarType   types.String `tfsdk:"var_type"`
}

func (t frontendTcpResponseRuleResourceModel) GetIndex() int64 {
	return t.Index.ValueInt64()
}

func (t frontendTcpResponseRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":       types.Int64Type,
		"action":      types.StringType,
		"cond":        types.StringType,
		"cond_test":   types.StringType,
		"lua_action":  types.StringType,
		"lua_params":  types.StringType,
		"sc_id":       types.Int64Type,
		"sc_idx":      types.Int64Type,
		"sc_int":      types.Int64Type,
		"sc_inc_gpc0": types.StringType,
		"sc_inc_gpc1": types.StringType,
		"sc_set_gpt0": types.StringType,
		"var_name":    types.StringType,
		"var_scope":   types.StringType,
		"var_expr":    types.StringType,
		"var_format":  types.StringType,
		"var_type":    types.StringType,
	}
}

// Metadata returns the resource type name.
func (r *frontendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontend"
}

// Schema defines the schema for the resource.
func (r *frontendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the frontend. It must be unique and cannot be changed.",
			},
			"default_backend": schema.StringAttribute{
				Required:    true,
				Description: "The name of the default_backend. Pattern: ^[A-Za-z0-9-_.:]+$",
			},
			"http_connection_mode": schema.StringAttribute{
				Optional:    true,
				Description: "The http connection mode of the frontend. Allowed: httpclose|http-server-close|http-keep-alive",
			},
			"accept_invalid_http_request": schema.StringAttribute{
				Optional:    true,
				Description: "The accept invalid http request of the frontend. Allowed: enabled|disabled",
			},
			"maxconn": schema.Int64Attribute{
				Optional:    true,
				Description: "The max connection of the frontend.",
			},
			"mode": schema.StringAttribute{
				Optional:    true,
				Description: "The mode of the frontend. Allowed: http|tcp",
			},
			"backlog": schema.Int64Attribute{
				Optional:    true,
				Description: "The backlog of the frontend.",
			},
			"http_keep_alive_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The http keep alive timeout of the frontend.",
			},
			"http_request_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The http request timeout of the frontend.",
			},
			"http_use_proxy_header": schema.StringAttribute{
				Optional:    true,
				Description: "The http use proxy header of the frontend. Allowed: enabled|disabled",
			},
			"httplog": schema.BoolAttribute{
				Optional:    true,
				Description: "The http log of the frontend.",
			},
			"httpslog": schema.StringAttribute{
				Optional:    true,
				Description: "The https log of the frontend. Allowed: enabled|disabled",
			},
			"error_log_format": schema.StringAttribute{
				Optional:    true,
				Description: "The error log format of the frontend.",
			},
			"log_format": schema.StringAttribute{
				Optional:    true,
				Description: "The log format of the frontend.",
			},
			"log_format_sd": schema.StringAttribute{
				Optional:    true,
				Description: "The log format sd of the frontend.",
			},
			"monitor_uri": schema.StringAttribute{
				Optional:    true,
				Description: "The monitor uri of the frontend.",
			},
			"tcplog": schema.BoolAttribute{
				Optional:    true,
				Description: "The tcp log of the frontend.",
			},
			"from": schema.StringAttribute{
				Optional:    true,
				Description: "The from of the frontend.",
			},
			"client_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The client timeout of the frontend.",
			},
			"http_use_htx": schema.StringAttribute{
				Optional:    true,
				Description: "The http use htx of the frontend. Allowed: enabled|disabled",
			},
			"http_ignore_probes": schema.StringAttribute{
				Optional:    true,
				Description: "The http ignore probes of the frontend. Allowed: enabled|disabled",
			},
			"log_tag": schema.StringAttribute{
				Optional:    true,
				Description: "The log tag of the frontend.",
			},
			"clflog": schema.BoolAttribute{
				Optional:    true,
				Description: "The clflog of the frontend.",
			},
			"contstats": schema.StringAttribute{
				Optional:    true,
				Description: "The contstats of the frontend. Allowed: enabled|disabled",
			},
			"dontlognull": schema.StringAttribute{
				Optional:    true,
				Description: "The dontlognull of the frontend. Allowed: enabled|disabled",
			},
			"log_separate_errors": schema.StringAttribute{
				Optional:    true,
				Description: "The log separate errors of the frontend. Allowed: enabled|disabled",
			},
			"option_http_server_close": schema.StringAttribute{
				Optional:    true,
				Description: "The option http server close of the frontend. Allowed: enabled|disabled",
			},
			"option_httpclose": schema.StringAttribute{
				Optional:    true,
				Description: "The option httpclose of the frontend. Allowed: enabled|disabled",
			},
			"option_http_keep_alive": schema.StringAttribute{
				Optional:    true,
				Description: "The option http keep alive of the frontend. Allowed: enabled|disabled",
			},
			"option_dontlog_normal": schema.StringAttribute{
				Optional:    true,
				Description: "The option dontlog normal of the frontend. Allowed: enabled|disabled",
			},
			"option_logasap": schema.StringAttribute{
				Optional:    true,
				Description: "The option logasap of the frontend. Allowed: enabled|disabled",
			},
			"option_tcplog": schema.StringAttribute{
				Optional:    true,
				Description: "The option tcplog of the frontend. Allowed: enabled|disabled",
			},
			"option_socket_stats": schema.StringAttribute{
				Optional:    true,
				Description: "The option socket stats of the frontend. Allowed: enabled|disabled",
			},
			"option_forwardfor": schema.StringAttribute{
				Optional:    true,
				Description: "The option forwardfor of the frontend. Allowed: enabled|disabled",
			},
			"timeout_client": schema.Int64Attribute{
				Optional:    true,
				Description: "The timeout client of the frontend.",
			},
			"timeout_http_keep_alive": schema.Int64Attribute{
				Optional:    true,
				Description: "The timeout http keep alive of the frontend.",
			},
			"timeout_http_request": schema.Int64Attribute{
				Optional:    true,
				Description: "The timeout http request of the frontend.",
			},
			"timeout_cont": schema.Int64Attribute{
				Optional:    true,
				Description: "The timeout cont of the frontend.",
			},
			"timeout_tarpit": schema.Int64Attribute{
				Optional:    true,
				Description: "The timeout tarpit of the frontend.",
			},
			"stats_options": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"stats_enable": schema.BoolAttribute{
						Optional:    true,
						Description: "Enable stats.",
					},
					"stats_hide_version": schema.BoolAttribute{
						Optional:    true,
						Description: "Hide version.",
					},
					"stats_show_legends": schema.BoolAttribute{
						Optional:    true,
						Description: "Show legends.",
					},
					"stats_show_node": schema.BoolAttribute{
						Optional:    true,
						Description: "Show node.",
					},
					"stats_uri": schema.StringAttribute{
						Optional:    true,
						Description: "Stats uri.",
					},
					"stats_realm": schema.StringAttribute{
						Optional:    true,
						Description: "Stats realm.",
					},
					"stats_auth": schema.StringAttribute{
						Optional:    true,
						Description: "Stats auth.",
					},
					"stats_refresh": schema.StringAttribute{
						Optional:    true,
						Description: "Stats refresh.",
					},
				},
			},
			"bind": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the bind. It must be unique and cannot be changed.",
						},
						"port": schema.Int64Attribute{
							Optional:    true,
							Description: "The port of the bind",
						},
						"port_range_end": schema.Int64Attribute{
							Optional:    true,
							Description: "The end of the port range",
						},
						"address": schema.StringAttribute{
							Required:    true,
							Description: "The address of the bind",
						},
						"transparent": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable transparent binding",
						},
						"mode": schema.StringAttribute{
							Optional:    true,
							Description: "http, tcp",
						},
						"maxconn": schema.Int64Attribute{
							Optional:    true,
							Description: "The max connections of the bind",
						},
						"user": schema.StringAttribute{
							Optional:    true,
							Description: "The user of the bind",
						},
						"group": schema.StringAttribute{
							Optional:    true,
							Description: "The group of the bind",
						},
						"force_sslv3": schema.BoolAttribute{
							Optional:    true,
							Description: "State of SSLv3 protocol support for the SSL. DEPRECATED: Use 'sslv3' field instead in Data Plane API v3",
						},
						"force_tlsv10": schema.BoolAttribute{
							Optional:    true,
							Description: "State of TLSv1.0 protocol support for the SSL. DEPRECATED: Use 'tlsv10' field instead in Data Plane API v3",
						},
						"force_tlsv11": schema.BoolAttribute{
							Optional:    true,
							Description: "State of TLSv1.1 protocol. DEPRECATED: Use 'tlsv11' field instead in Data Plane API v3",
						},
						"force_tlsv12": schema.BoolAttribute{
							Optional:    true,
							Description: "State of TLSv1.2 protocol. DEPRECATED: Use 'tlsv12' field instead in Data Plane API v3",
						},
						"force_tlsv13": schema.BoolAttribute{
							Optional:    true,
							Description: "State of TLSv1.3 protocol. DEPRECATED: Use 'tlsv13' field instead in Data Plane API v3",
						},
						"force_strict_sni": schema.StringAttribute{
							Optional:    true,
							Description: "Force strict SNI. DEPRECATED: Use 'strict_sni' field instead in Data Plane API v3. Allowed: enabled|disabled",
						},
						"ssl": schema.BoolAttribute{
							Optional:    true,
							Description: "State of SSL",
						},
						"ssl_cafile": schema.StringAttribute{
							Optional:    true,
							Description: "ssl CA file. Pattern: ^[^\\s]+$",
						},
						"ssl_max_ver": schema.StringAttribute{
							Optional:    true,
							Description: "ssl max version to support. Allowed: SSLv3|TLSv1.0|TLSv1.1|TLSv1.2|TLSv1.3",
						},
						"ssl_min_ver": schema.StringAttribute{
							Optional:    true,
							Description: "ssl min version to support. Allowed: SSLv3|TLSv1.0|TLSv1.1|TLSv1.2|TLSv1.3",
						},
						"ssl_certificate": schema.StringAttribute{
							Optional:    true,
							Description: "Path of SSL certificate. Pattern: ^[^\\s]+$",
						},
						"ciphers": schema.StringAttribute{
							Optional:    true,
							Description: "ciphers to support",
						},
						"ciphersuites": schema.StringAttribute{
							Optional:    true,
							Description: "ciphersuites to support",
						},
						"accept_proxy": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable accept-proxy",
						},
						"allow_0rtt": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable allow-0rtt",
						},
						"alpn": schema.StringAttribute{
							Optional:    true,
							Description: "Set ALPN protocols",
						},
						"backlog": schema.StringAttribute{
							Optional:    true,
							Description: "Set backlog size",
						},
						"ca_ignore_err": schema.StringAttribute{
							Optional:    true,
							Description: "Ignore CA errors",
						},
						"ca_sign_file": schema.StringAttribute{
							Optional:    true,
							Description: "CA sign file",
						},
						"ca_sign_pass": schema.StringAttribute{
							Optional:    true,
							Description: "CA sign password",
						},
						"ca_verify_file": schema.StringAttribute{
							Optional:    true,
							Description: "CA verify file",
						},
						"crl_file": schema.StringAttribute{
							Optional:    true,
							Description: "CRL file",
						},
						"crt_ignore_err": schema.StringAttribute{
							Optional:    true,
							Description: "Ignore certificate errors",
						},
						"crt_list": schema.StringAttribute{
							Optional:    true,
							Description: "Certificate list",
						},
						"defer_accept": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable defer-accept",
						},
						"expose_via_agent": schema.BoolAttribute{
							Optional:    true,
							Description: "Expose via agent",
						},
						"generate_certificates": schema.BoolAttribute{
							Optional:    true,
							Description: "Generate certificates",
						},
						"gid": schema.Int64Attribute{
							Optional:    true,
							Description: "Set GID",
						},
						"id": schema.StringAttribute{
							Optional:    true,
							Description: "Set ID",
						},
						"interface": schema.StringAttribute{
							Optional:    true,
							Description: "Set interface",
						},
						"level": schema.StringAttribute{
							Optional:    true,
							Description: "Set level",
						},
						"log_proto": schema.StringAttribute{
							Optional:    true,
							Description: "Set log protocol",
						},
						"mdev": schema.StringAttribute{
							Optional:    true,
							Description: "Set mdev",
						},
						"namespace": schema.StringAttribute{
							Optional:    true,
							Description: "Set namespace",
						},
						"nice": schema.Int64Attribute{
							Optional:    true,
							Description: "Set nice value",
						},
						"no_ca_names": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable CA names",
						},
						"no_sslv3": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable SSLv3. DEPRECATED: Use 'sslv3' field instead in Data Plane API v3",
						},
						"no_tlsv10": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable TLSv1.0. DEPRECATED: Use 'tlsv10' field instead in Data Plane API v3",
						},
						"no_tlsv11": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable TLSv1.1. DEPRECATED: Use 'tlsv11' field instead in Data Plane API v3",
						},
						"no_tlsv12": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable TLSv1.2. DEPRECATED: Use 'tlsv12' field instead in Data Plane API v3",
						},
						"no_tlsv13": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable TLSv1.3. DEPRECATED: Use 'tlsv13' field instead in Data Plane API v3",
						},
						// New v3 fields (non-deprecated)
						"sslv3": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable SSLv3 protocol support (v3 API, replaces no_sslv3)",
						},
						"tlsv10": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable TLSv1.0 protocol support (v3 API, replaces no_tlsv10)",
						},
						"tlsv11": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable TLSv1.1 protocol support (v3 API, replaces no_tlsv11)",
						},
						"tlsv12": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable TLSv1.2 protocol support (v3 API, replaces no_tlsv12)",
						},
						"tlsv13": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable TLSv1.3 protocol support (v3 API, replaces no_tlsv13)",
						},

						"npn": schema.StringAttribute{
							Optional:    true,
							Description: "Set NPN protocols",
						},
						"prefer_client_ciphers": schema.BoolAttribute{
							Optional:    true,
							Description: "Prefer client ciphers",
						},
						"process": schema.StringAttribute{
							Optional:    true,
							Description: "Set process",
						},
						"proto": schema.StringAttribute{
							Optional:    true,
							Description: "Set proto",
						},
						"severity_output": schema.StringAttribute{
							Optional:    true,
							Description: "Set severity output",
						},
						"strict_sni": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable strict SNI",
						},
						"tcp_user_timeout": schema.Int64Attribute{
							Optional:    true,
							Description: "Set TCP user timeout",
						},
						"tfo": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable TFO",
						},
						"tls_ticket_keys": schema.StringAttribute{
							Optional:    true,
							Description: "Set TLS ticket keys",
						},
						"uid": schema.StringAttribute{
							Optional:    true,
							Description: "Set UID",
						},
						"v4v6": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable v4v6",
						},
						"v6only": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable v6only",
						},
						"verify": schema.StringAttribute{
							Optional:    true,
							Description: "Set verify",
						},
						"metadata": schema.StringAttribute{
							Optional:    true,
							Description: "Metadata for the bind",
						},
					},
				},
			},
			"tcprequestrule": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the tcp-request rule",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the tcp-request rule",
						},
						"action": schema.StringAttribute{
							Optional:    true,
							Description: "The action of the tcp-request rule",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the tcp-request rule",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the tcp-request rule",
						},
						"timeout": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout of the tcp-request rule",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The lua action of the tcp-request rule",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The lua params of the tcp-request rule",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc id of the tcp-request rule",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc idx of the tcp-request rule",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc int of the tcp-request rule",
						},
						"sc_inc_gpc0": schema.StringAttribute{
							Optional:    true,
							Description: "The sc inc gpc0 of the tcp-request rule",
						},
						"sc_inc_gpc1": schema.StringAttribute{
							Optional:    true,
							Description: "The sc inc gpc1 of the tcp-request rule",
						},
						"sc_set_gpt0": schema.StringAttribute{
							Optional:    true,
							Description: "The sc set gpt0 of the tcp-request rule",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc key of the tcp-request rule",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc table of the tcp-request rule",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The var name of the tcp-request rule",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The var scope of the tcp-request rule",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The var expr of the tcp-request rule",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The var format of the tcp-request rule",
						},
						"var_type": schema.StringAttribute{
							Optional:    true,
							Description: "The var type of the tcp-request rule",
						},
					},
				},
			},
			"tcpresponserule": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the tcp-response rule",
						},
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The action of the tcp-response rule",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the tcp-response rule",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the tcp-response rule",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The lua action of the tcp-response rule",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The lua params of the tcp-response rule",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc id of the tcp-response rule",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc idx of the tcp-response rule",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc int of the tcp-response rule",
						},
						"sc_inc_gpc0": schema.StringAttribute{
							Optional:    true,
							Description: "The sc inc gpc0 of the tcp-response rule",
						},
						"sc_inc_gpc1": schema.StringAttribute{
							Optional:    true,
							Description: "The sc inc gpc1 of the tcp-response rule",
						},
						"sc_set_gpt0": schema.StringAttribute{
							Optional:    true,
							Description: "The sc set gpt0 of the tcp-response rule",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The var name of the tcp-response rule",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The var scope of the tcp-response rule",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The var expr of the tcp-response rule",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The var format of the tcp-response rule",
						},
						"var_type": schema.StringAttribute{
							Optional:    true,
							Description: "The var type of the tcp-response rule",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"monitor_fail": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cond": schema.StringAttribute{
							Required:    true,
							Description: "The cond of the monitor_fail. Allowed: if|unless",
						},
						"cond_test": schema.StringAttribute{
							Required:    true,
							Description: "The cond_test of the monitor_fail.",
						},
					},
				},
			},
			"acl": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"acl_name": schema.StringAttribute{
							Required:    true,
							Description: "The acl name. Pattern: ^[^\\s]+$",
						},
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the acl",
						},
						"criterion": schema.StringAttribute{
							Required:    true,
							Description: "The criterion. Pattern: ^[^\\s]+$",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The value of the criterion",
						},
					},
				},
			},
			"httprequestrule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the http-request rule",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the http-request rule",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the http-request rule",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the http-request rule",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name of the http-request rule",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format of the http-request rule",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type of the http-request rule",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value of the http-request rule",
						},
						"status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The status code of the http-request rule",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-request rule",
						},
					},
				},
			},
			"httpresponserule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the http-response rule",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the http-response rule",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the http-response rule",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the http-response rule",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name of the http-response rule",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format of the http-response rule",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type of the http-response rule",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value of the http-response rule",
						},
						"status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The status code of the http-response rule",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-response rule",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *frontendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *frontendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan frontendResourceModel
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

	payload := &FrontendPayload{
		Name:                     plan.Name.ValueString(),
		DefaultBackend:           plan.DefaultBackend.ValueString(),
		HttpConnectionMode:       plan.HttpConnectionMode.ValueString(),
		AcceptInvalidHttpRequest: plan.AcceptInvalidHttpRequest.ValueString(),
		MaxConn:                  plan.MaxConn.ValueInt64(),
		Mode:                     plan.Mode.ValueString(),
		Backlog:                  plan.Backlog.ValueInt64(),
		HttpKeepAliveTimeout:     plan.HttpKeepAliveTimeout.ValueInt64(),
		HttpRequestTimeout:       plan.HttpRequestTimeout.ValueInt64(),
		HttpUseProxyHeader:       plan.HttpUseProxyHeader.ValueString(),
		HttpLog:                  plan.HttpLog.ValueBool(),
		HttpsLog:                 plan.HttpsLog.ValueString(),
		ErrorLogFormat:           plan.ErrorLogFormat.ValueString(),
		LogFormat:                plan.LogFormat.ValueString(),
		LogFormatSd:              plan.LogFormatSd.ValueString(),
		MonitorUri:               plan.MonitorUri.ValueString(),
		TcpLog:                   plan.TcpLog.ValueBool(),
		From:                     plan.From.ValueString(),
		ClientTimeout:            plan.ClientTimeout.ValueInt64(),
		HttpUseHtx:               plan.HttpUseHtx.ValueString(),
		HttpIgnoreProbes:         plan.HttpIgnoreProbes.ValueString(),
		LogTag:                   plan.LogTag.ValueString(),
		Clflog:                   plan.Clflog.ValueBool(),
		Contstats:                plan.Contstats.ValueString(),
		Dontlognull:              plan.Dontlognull.ValueString(),
		LogSeparateErrors:        plan.LogSeparateErrors.ValueString(),
		OptionHttpServerClose:    plan.OptionHttpServerClose.ValueString(),
		OptionHttpclose:          plan.OptionHttpclose.ValueString(),
		OptionHttpKeepAlive:      plan.OptionHttpKeepAlive.ValueString(),
		OptionDontlogNormal:      plan.OptionDontlogNormal.ValueString(),
		OptionLogasap:            plan.OptionLogasap.ValueString(),
		OptionTcplog:             plan.OptionTcplog.ValueString(),
		OptionSocketStats:        plan.OptionSocketStats.ValueString(),
		OptionForwardfor:         plan.OptionForwardfor.ValueString(),
		TimeoutClient:            plan.TimeoutClient.ValueInt64(),
		TimeoutHttpKeepAlive:     plan.TimeoutHttpKeepAlive.ValueInt64(),
		TimeoutHttpRequest:       plan.TimeoutHttpRequest.ValueInt64(),
		TimeoutCont:              plan.TimeoutCont.ValueInt64(),
		TimeoutTarpit:            plan.TimeoutTarpit.ValueInt64(),
	}

	if !plan.StatsOptions.IsNull() {
		var statsOptionsModel struct {
			StatsEnable      types.Bool   `tfsdk:"stats_enable"`
			StatsHideVersion types.Bool   `tfsdk:"stats_hide_version"`
			StatsShowLegends types.Bool   `tfsdk:"stats_show_legends"`
			StatsShowNode    types.Bool   `tfsdk:"stats_show_node"`
			StatsUri         types.String `tfsdk:"stats_uri"`
			StatsRealm       types.String `tfsdk:"stats_realm"`
			StatsAuth        types.String `tfsdk:"stats_auth"`
			StatsRefresh     types.String `tfsdk:"stats_refresh"`
		}
		diags := plan.StatsOptions.As(ctx, &statsOptionsModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload.StatsOptions = StatsOptionsPayload{
			StatsEnable:      statsOptionsModel.StatsEnable.ValueBool(),
			StatsHideVersion: statsOptionsModel.StatsHideVersion.ValueBool(),
			StatsShowLegends: statsOptionsModel.StatsShowLegends.ValueBool(),
			StatsShowNode:    statsOptionsModel.StatsShowNode.ValueBool(),
			StatsUri:         statsOptionsModel.StatsUri.ValueString(),
			StatsRealm:       statsOptionsModel.StatsRealm.ValueString(),
			StatsAuth:        statsOptionsModel.StatsAuth.ValueString(),
			StatsRefresh:     statsOptionsModel.StatsRefresh.ValueString(),
		}
	}

	log.Printf("Creating frontend with payload: %+v", payload)

	// Use the old transaction method which has built-in retry logic
	err := r.client.CreateFrontend(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating frontend",
			"Could not create frontend, unexpected error: "+err.Error(),
		)
		return
	}

	log.Printf("Frontend created successfully")

	if !plan.Binds.IsNull() {
		var bindModels []bindResourceModel
		diags := plan.Binds.ElementsAs(ctx, &bindModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, bindModel := range bindModels {
			bindPayload := &BindPayload{
				Name:                 bindModel.Name.ValueString(),
				Address:              bindModel.Address.ValueString(),
				Transparent:          bindModel.Transparent.ValueBool(),
				Mode:                 bindModel.Mode.ValueString(),
				Maxconn:              bindModel.Maxconn.ValueInt64(),
				User:                 bindModel.User.ValueString(),
				Group:                bindModel.Group.ValueString(),
				ForceSslv3:           bindModel.ForceSslv3.ValueBool(),
				ForceTlsv10:          bindModel.ForceTlsv10.ValueBool(),
				ForceTlsv11:          bindModel.ForceTlsv11.ValueBool(),
				ForceTlsv12:          bindModel.ForceTlsv12.ValueBool(),
				ForceTlsv13:          bindModel.ForceTlsv13.ValueBool(),
				ForceStrictSni:       bindModel.ForceStrictSni.ValueString(),
				Ssl:                  bindModel.Ssl.ValueBool(),
				SslCafile:            bindModel.SslCafile.ValueString(),
				SslMaxVer:            bindModel.SslMaxVer.ValueString(),
				SslMinVer:            bindModel.SslMinVer.ValueString(),
				SslCertificate:       bindModel.SslCertificate.ValueString(),
				Ciphers:              bindModel.Ciphers.ValueString(),
				Ciphersuites:         bindModel.Ciphersuites.ValueString(),
				AcceptProxy:          bindModel.AcceptProxy.ValueBool(),
				Allow0rtt:            bindModel.Allow0rtt.ValueBool(),
				Alpn:                 bindModel.Alpn.ValueString(),
				Backlog:              bindModel.Backlog.ValueString(),
				CaIgnoreErr:          bindModel.CaIgnoreErr.ValueString(),
				CaSignFile:           bindModel.CaSignFile.ValueString(),
				CaSignPass:           bindModel.CaSignPass.ValueString(),
				CaVerifyFile:         bindModel.CaVerifyFile.ValueString(),
				CrlFile:              bindModel.CrlFile.ValueString(),
				CrtIgnoreErr:         bindModel.CrtIgnoreErr.ValueString(),
				CrtList:              bindModel.CrtList.ValueString(),
				DeferAccept:          bindModel.DeferAccept.ValueBool(),
				ExposeViaAgent:       bindModel.ExposeViaAgent.ValueBool(),
				GenerateCertificates: bindModel.GenerateCertificates.ValueBool(),
				Gid:                  bindModel.Gid.ValueInt64(),
				Id:                   bindModel.Id.ValueString(),
				Interface:            bindModel.Interface.ValueString(),
				Level:                bindModel.Level.ValueString(),
				LogProto:             bindModel.LogProto.ValueString(),
				Mdev:                 bindModel.Mdev.ValueString(),
				Namespace:            bindModel.Namespace.ValueString(),
				Nice:                 bindModel.Nice.ValueInt64(),
				NoCaNames:            bindModel.NoCaNames.ValueBool(),
				NoSslv3:              bindModel.NoSslv3.ValueBool(),
				NoTlsv10:             bindModel.NoTlsv10.ValueBool(),
				NoTlsv11:             bindModel.NoTlsv11.ValueBool(),
				NoTlsv12:             bindModel.NoTlsv12.ValueBool(),
				NoTlsv13:             bindModel.NoTlsv13.ValueBool(),
				// New v3 fields
				Sslv3:               bindModel.Sslv3.ValueBool(),
				Tlsv10:              bindModel.Tlsv10.ValueBool(),
				Tlsv11:              bindModel.Tlsv11.ValueBool(),
				Tlsv12:              bindModel.Tlsv12.ValueBool(),
				Tlsv13:              bindModel.Tlsv13.ValueBool(),
				Npn:                 bindModel.Npn.ValueString(),
				PreferClientCiphers: bindModel.PreferClientCiphers.ValueBool(),
				Process:             bindModel.Process.ValueString(),
				Proto:               bindModel.Proto.ValueString(),
				SeverityOutput:      bindModel.SeverityOutput.ValueString(),
				StrictSni:           bindModel.StrictSni.ValueBool(),
				TcpUserTimeout:      bindModel.TcpUserTimeout.ValueInt64(),
				Tfo:                 bindModel.Tfo.ValueBool(),
				TlsTicketKeys:       bindModel.TlsTicketKeys.ValueString(),
				Uid:                 bindModel.Uid.ValueString(),
				V4v6:                bindModel.V4v6.ValueBool(),
				V6only:              bindModel.V6only.ValueBool(),
				Verify:              bindModel.Verify.ValueString(),
				Metadata:            bindModel.Metadata.ValueString(),
			}
			if !bindModel.Port.IsNull() {
				p := bindModel.Port.ValueInt64()
				bindPayload.Port = &p
			}
			if !bindModel.PortRangeEnd.IsNull() {
				p := bindModel.PortRangeEnd.ValueInt64()
				bindPayload.PortRangeEnd = &p
			}
			err := r.client.CreateBind(ctx, "frontend", plan.Name.ValueString(), bindPayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating bind",
					fmt.Sprintf("Could not create bind %s, unexpected error: %s", bindModel.Name.ValueString(), err.Error()),
				)
				return
			}
		}
	}

	if !plan.Acls.IsNull() {
		var aclModels []frontendAclResourceModel
		diags := plan.Acls.ElementsAs(ctx, &aclModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(aclModels, func(i, j int) bool {
			return aclModels[i].GetIndex() < aclModels[j].GetIndex()
		})

		for _, aclModel := range aclModels {
			aclPayload := &AclPayload{
				AclName:   aclModel.AclName.ValueString(),
				Index:     aclModel.Index.ValueInt64(),
				Criterion: aclModel.Criterion.ValueString(),
				Value:     aclModel.Value.ValueString(),
			}
			err := r.client.CreateAcl(ctx, "frontend", plan.Name.ValueString(), aclPayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating acl",
					fmt.Sprintf("Could not create acl %s, unexpected error: %s", aclModel.AclName.ValueString(), err.Error()),
				)
				return
			}
		}
	}

	// Create monitor_fail AFTER ACLs are created
	if !plan.MonitorFail.IsNull() {
		var monitorFailModels []struct {
			Cond     types.String `tfsdk:"cond"`
			CondTest types.String `tfsdk:"cond_test"`
		}
		diags := plan.MonitorFail.ElementsAs(ctx, &monitorFailModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(monitorFailModels) > 0 {
			// Update the frontend payload with monitor_fail after ACLs are created
			updatePayload := &FrontendPayload{
				Name: plan.Name.ValueString(),
				MonitorFail: &MonitorFailPayload{
					Cond:     monitorFailModels[0].Cond.ValueString(),
					CondTest: monitorFailModels[0].CondTest.ValueString(),
				},
			}
			err := r.client.UpdateFrontend(ctx, plan.Name.ValueString(), updatePayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating frontend with monitor_fail",
					fmt.Sprintf("Could not update frontend with monitor_fail, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	if !plan.HttpRequestRules.IsNull() {
		var httpRequestRuleModels []httpRequestRuleResourceModel
		diags := plan.HttpRequestRules.ElementsAs(ctx, &httpRequestRuleModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(httpRequestRuleModels, func(i, j int) bool {
			return httpRequestRuleModels[i].GetIndex() < httpRequestRuleModels[j].GetIndex()
		})

		for _, httpRequestRuleModel := range httpRequestRuleModels {
			httpRequestRulePayload := &HttpRequestRulePayload{
				Index:        httpRequestRuleModel.Index.ValueInt64(),
				Type:         httpRequestRuleModel.Type.ValueString(),
				Cond:         httpRequestRuleModel.Cond.ValueString(),
				CondTest:     httpRequestRuleModel.CondTest.ValueString(),
				HdrName:      httpRequestRuleModel.HdrName.ValueString(),
				HdrFormat:    httpRequestRuleModel.HdrFormat.ValueString(),
				RedirType:    httpRequestRuleModel.RedirType.ValueString(),
				RedirValue:   httpRequestRuleModel.RedirValue.ValueString(),
				StatusCode:   httpRequestRuleModel.StatusCode.ValueInt64(),
				StatusReason: httpRequestRuleModel.StatusReason.ValueString(),
			}
			err := r.client.CreateHttpRequestRule(ctx, "frontend", plan.Name.ValueString(), httpRequestRulePayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating http-request rule",
					fmt.Sprintf("Could not create http-request rule, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	if !plan.HttpResponseRules.IsNull() {
		var httpResponseRuleModels []httpResponseRuleResourceModel
		diags := plan.HttpResponseRules.ElementsAs(ctx, &httpResponseRuleModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(httpResponseRuleModels, func(i, j int) bool {
			return httpResponseRuleModels[i].GetIndex() < httpResponseRuleModels[j].GetIndex()
		})

		for _, httpResponseRuleModel := range httpResponseRuleModels {
			httpResponseRulePayload := &HttpResponseRulePayload{
				Index:        httpResponseRuleModel.Index.ValueInt64(),
				Type:         httpResponseRuleModel.Type.ValueString(),
				Cond:         httpResponseRuleModel.Cond.ValueString(),
				CondTest:     httpResponseRuleModel.CondTest.ValueString(),
				HdrName:      httpResponseRuleModel.HdrName.ValueString(),
				HdrFormat:    httpResponseRuleModel.HdrFormat.ValueString(),
				RedirType:    httpResponseRuleModel.RedirType.ValueString(),
				RedirValue:   httpResponseRuleModel.RedirValue.ValueString(),
				StatusCode:   httpResponseRuleModel.StatusCode.ValueInt64(),
				StatusReason: httpResponseRuleModel.StatusReason.ValueString(),
			}
			err := r.client.CreateHttpResponseRule(ctx, "frontend", plan.Name.ValueString(), httpResponseRulePayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating http-response rule",
					fmt.Sprintf("Could not create http-response rule, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	if !plan.TcpRequestRules.IsNull() {
		var tcpRequestRuleModels []frontendTcpRequestRuleResourceModel
		diags := plan.TcpRequestRules.ElementsAs(ctx, &tcpRequestRuleModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(tcpRequestRuleModels, func(i, j int) bool {
			return tcpRequestRuleModels[i].GetIndex() < tcpRequestRuleModels[j].GetIndex()
		})

		for _, tcpRequestRuleModel := range tcpRequestRuleModels {
			tcpRequestRulePayload := &TcpRequestRulePayload{
				Index:        tcpRequestRuleModel.Index.ValueInt64(),
				Type:         tcpRequestRuleModel.Type.ValueString(),
				Action:       tcpRequestRuleModel.Action.ValueString(),
				Cond:         tcpRequestRuleModel.Cond.ValueString(),
				CondTest:     tcpRequestRuleModel.CondTest.ValueString(),
				Timeout:      tcpRequestRuleModel.Timeout.ValueInt64(),
				LuaAction:    tcpRequestRuleModel.LuaAction.ValueString(),
				LuaParams:    tcpRequestRuleModel.LuaParams.ValueString(),
				ScId:         tcpRequestRuleModel.ScId.ValueInt64(),
				ScIdx:        tcpRequestRuleModel.ScIdx.ValueInt64(),
				ScInt:        tcpRequestRuleModel.ScInt.ValueInt64(),
				ScIncGpc0:    tcpRequestRuleModel.ScIncGpc0.ValueString(),
				ScIncGpc1:    tcpRequestRuleModel.ScIncGpc1.ValueString(),
				ScSetGpt0:    tcpRequestRuleModel.ScSetGpt0.ValueString(),
				TrackScKey:   tcpRequestRuleModel.TrackScKey.ValueString(),
				TrackScTable: tcpRequestRuleModel.TrackScTable.ValueString(),
				VarName:      tcpRequestRuleModel.VarName.ValueString(),
				VarScope:     tcpRequestRuleModel.VarScope.ValueString(),
				VarExpr:      tcpRequestRuleModel.VarExpr.ValueString(),
				VarFormat:    tcpRequestRuleModel.VarFormat.ValueString(),
				VarType:      tcpRequestRuleModel.VarType.ValueString(),
			}
			err := r.client.CreateTcpRequestRule(ctx, "frontend", plan.Name.ValueString(), tcpRequestRulePayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating tcp-request rule",
					fmt.Sprintf("Could not create tcp-request rule, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	if !plan.TcpResponseRules.IsNull() {
		var tcpResponseRuleModels []frontendTcpResponseRuleResourceModel
		diags := plan.TcpResponseRules.ElementsAs(ctx, &tcpResponseRuleModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(tcpResponseRuleModels, func(i, j int) bool {
			return tcpResponseRuleModels[i].GetIndex() < tcpResponseRuleModels[j].GetIndex()
		})

		for _, tcpResponseRuleModel := range tcpResponseRuleModels {
			tcpResponseRulePayload := &TcpResponseRulePayload{
				Index:     tcpResponseRuleModel.Index.ValueInt64(),
				Action:    tcpResponseRuleModel.Action.ValueString(),
				Cond:      tcpResponseRuleModel.Cond.ValueString(),
				CondTest:  tcpResponseRuleModel.CondTest.ValueString(),
				LuaAction: tcpResponseRuleModel.LuaAction.ValueString(),
				LuaParams: tcpResponseRuleModel.LuaParams.ValueString(),
				ScId:      tcpResponseRuleModel.ScId.ValueInt64(),
				ScIdx:     tcpResponseRuleModel.ScIdx.ValueInt64(),
				ScInt:     tcpResponseRuleModel.ScInt.ValueInt64(),
				ScIncGpc0: tcpResponseRuleModel.ScIncGpc0.ValueString(),
				ScIncGpc1: tcpResponseRuleModel.ScIncGpc1.ValueString(),
				ScSetGpt0: tcpResponseRuleModel.ScSetGpt0.ValueString(),
				VarName:   tcpResponseRuleModel.VarName.ValueString(),
				VarScope:  tcpResponseRuleModel.VarScope.ValueString(),
				VarExpr:   tcpResponseRuleModel.VarExpr.ValueString(),
				VarFormat: tcpResponseRuleModel.VarFormat.ValueString(),
				VarType:   tcpResponseRuleModel.VarType.ValueString(),
			}
			err := r.client.CreateTcpResponseRule(ctx, "frontend", plan.Name.ValueString(), tcpResponseRulePayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating tcp-response rule",
					fmt.Sprintf("Could not create tcp-response rule, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *frontendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state frontendResourceModel
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

	frontend, err := r.client.ReadFrontend(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading frontend",
			"Could not read frontend, unexpected error: "+err.Error(),
		)
		return
	}

	if frontend == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(frontend.Name)
	state.Mode = types.StringValue(frontend.Mode)
	// Only set fields if they have meaningful values (not empty/zero)
	if frontend.DefaultBackend != "" {
		state.DefaultBackend = types.StringValue(frontend.DefaultBackend)
	} else {
		state.DefaultBackend = types.StringNull()
	}
	if frontend.HttpConnectionMode != "" {
		state.HttpConnectionMode = types.StringValue(frontend.HttpConnectionMode)
	} else {
		state.HttpConnectionMode = types.StringNull()
	}
	if frontend.AcceptInvalidHttpRequest != "" {
		state.AcceptInvalidHttpRequest = types.StringValue(frontend.AcceptInvalidHttpRequest)
	} else {
		state.AcceptInvalidHttpRequest = types.StringNull()
	}
	if frontend.MaxConn > 0 {
		state.MaxConn = types.Int64Value(frontend.MaxConn)
	} else {
		state.MaxConn = types.Int64Null()
	}
	if frontend.Backlog > 0 {
		state.Backlog = types.Int64Value(frontend.Backlog)
	} else {
		state.Backlog = types.Int64Null()
	}
	if frontend.HttpKeepAliveTimeout > 0 {
		state.HttpKeepAliveTimeout = types.Int64Value(frontend.HttpKeepAliveTimeout)
	} else {
		state.HttpKeepAliveTimeout = types.Int64Null()
	}
	if frontend.HttpRequestTimeout > 0 {
		state.HttpRequestTimeout = types.Int64Value(frontend.HttpRequestTimeout)
	} else {
		state.HttpRequestTimeout = types.Int64Null()
	}
	if frontend.HttpUseProxyHeader != "" {
		state.HttpUseProxyHeader = types.StringValue(frontend.HttpUseProxyHeader)
	} else {
		state.HttpUseProxyHeader = types.StringNull()
	}
	if frontend.HttpLog {
		state.HttpLog = types.BoolValue(true)
	} else {
		state.HttpLog = types.BoolNull()
	}
	if frontend.HttpsLog != "" {
		state.HttpsLog = types.StringValue(frontend.HttpsLog)
	} else {
		state.HttpsLog = types.StringNull()
	}
	if frontend.ErrorLogFormat != "" {
		state.ErrorLogFormat = types.StringValue(frontend.ErrorLogFormat)
	} else {
		state.ErrorLogFormat = types.StringNull()
	}
	if frontend.LogFormat != "" {
		state.LogFormat = types.StringValue(frontend.LogFormat)
	} else {
		state.LogFormat = types.StringNull()
	}
	if frontend.LogFormatSd != "" {
		state.LogFormatSd = types.StringValue(frontend.LogFormatSd)
	} else {
		state.LogFormatSd = types.StringNull()
	}
	if frontend.MonitorUri != "" {
		state.MonitorUri = types.StringValue(frontend.MonitorUri)
	} else {
		state.MonitorUri = types.StringNull()
	}
	if frontend.TcpLog {
		state.TcpLog = types.BoolValue(true)
	} else {
		state.TcpLog = types.BoolNull()
	}
	if frontend.From != "" {
		state.From = types.StringValue(frontend.From)
	} else {
		state.From = types.StringNull()
	}
	if frontend.ClientTimeout > 0 {
		state.ClientTimeout = types.Int64Value(frontend.ClientTimeout)
	} else {
		state.ClientTimeout = types.Int64Null()
	}
	if frontend.HttpUseHtx != "" {
		state.HttpUseHtx = types.StringValue(frontend.HttpUseHtx)
	} else {
		state.HttpUseHtx = types.StringNull()
	}
	if frontend.HttpIgnoreProbes != "" {
		state.HttpIgnoreProbes = types.StringValue(frontend.HttpIgnoreProbes)
	} else {
		state.HttpIgnoreProbes = types.StringNull()
	}
	if frontend.LogTag != "" {
		state.LogTag = types.StringValue(frontend.LogTag)
	} else {
		state.LogTag = types.StringNull()
	}
	if frontend.Clflog {
		state.Clflog = types.BoolValue(true)
	} else {
		state.Clflog = types.BoolNull()
	}
	if frontend.Contstats != "" {
		state.Contstats = types.StringValue(frontend.Contstats)
	} else {
		state.Contstats = types.StringNull()
	}
	if frontend.Dontlognull != "" {
		state.Dontlognull = types.StringValue(frontend.Dontlognull)
	} else {
		state.Dontlognull = types.StringNull()
	}
	if frontend.LogSeparateErrors != "" {
		state.LogSeparateErrors = types.StringValue(frontend.LogSeparateErrors)
	} else {
		state.LogSeparateErrors = types.StringNull()
	}
	if frontend.OptionHttpServerClose != "" {
		state.OptionHttpServerClose = types.StringValue(frontend.OptionHttpServerClose)
	} else {
		state.OptionHttpServerClose = types.StringNull()
	}
	if frontend.OptionHttpclose != "" {
		state.OptionHttpclose = types.StringValue(frontend.OptionHttpclose)
	} else {
		state.OptionHttpclose = types.StringNull()
	}
	if frontend.OptionHttpKeepAlive != "" {
		state.OptionHttpKeepAlive = types.StringValue(frontend.OptionHttpKeepAlive)
	} else {
		state.OptionHttpKeepAlive = types.StringNull()
	}
	if frontend.OptionDontlogNormal != "" {
		state.OptionDontlogNormal = types.StringValue(frontend.OptionDontlogNormal)
	} else {
		state.OptionDontlogNormal = types.StringNull()
	}
	if frontend.OptionLogasap != "" {
		state.OptionLogasap = types.StringValue(frontend.OptionLogasap)
	} else {
		state.OptionLogasap = types.StringNull()
	}
	if frontend.OptionTcplog != "" {
		state.OptionTcplog = types.StringValue(frontend.OptionTcplog)
	} else {
		state.OptionTcplog = types.StringNull()
	}
	if frontend.OptionSocketStats != "" {
		state.OptionSocketStats = types.StringValue(frontend.OptionSocketStats)
	} else {
		state.OptionSocketStats = types.StringNull()
	}
	if frontend.OptionForwardfor != "" {
		state.OptionForwardfor = types.StringValue(frontend.OptionForwardfor)
	} else {
		state.OptionForwardfor = types.StringNull()
	}
	// Only set timeout fields if they have meaningful values (not zero)
	if frontend.TimeoutClient > 0 {
		state.TimeoutClient = types.Int64Value(frontend.TimeoutClient)
	} else {
		state.TimeoutClient = types.Int64Null()
	}
	if frontend.TimeoutHttpKeepAlive > 0 {
		state.TimeoutHttpKeepAlive = types.Int64Value(frontend.TimeoutHttpKeepAlive)
	} else {
		state.TimeoutHttpKeepAlive = types.Int64Null()
	}
	if frontend.TimeoutHttpRequest > 0 {
		state.TimeoutHttpRequest = types.Int64Value(frontend.TimeoutHttpRequest)
	} else {
		state.TimeoutHttpRequest = types.Int64Null()
	}
	if frontend.TimeoutCont > 0 {
		state.TimeoutCont = types.Int64Value(frontend.TimeoutCont)
	} else {
		state.TimeoutCont = types.Int64Null()
	}
	if frontend.TimeoutTarpit > 0 {
		state.TimeoutTarpit = types.Int64Value(frontend.TimeoutTarpit)
	} else {
		state.TimeoutTarpit = types.Int64Null()
	}

	if frontend.StatsOptions != (StatsOptionsPayload{}) {
		var statsOptionsModel struct {
			StatsEnable      types.Bool   `tfsdk:"stats_enable"`
			StatsHideVersion types.Bool   `tfsdk:"stats_hide_version"`
			StatsShowLegends types.Bool   `tfsdk:"stats_show_legends"`
			StatsShowNode    types.Bool   `tfsdk:"stats_show_node"`
			StatsUri         types.String `tfsdk:"stats_uri"`
			StatsRealm       types.String `tfsdk:"stats_realm"`
			StatsAuth        types.String `tfsdk:"stats_auth"`
			StatsRefresh     types.String `tfsdk:"stats_refresh"`
		}
		statsOptionsModel.StatsEnable = types.BoolValue(frontend.StatsOptions.StatsEnable)
		statsOptionsModel.StatsHideVersion = types.BoolValue(frontend.StatsOptions.StatsHideVersion)
		statsOptionsModel.StatsShowLegends = types.BoolValue(frontend.StatsOptions.StatsShowLegends)
		statsOptionsModel.StatsShowNode = types.BoolValue(frontend.StatsOptions.StatsShowNode)
		statsOptionsModel.StatsUri = types.StringValue(frontend.StatsOptions.StatsUri)
		statsOptionsModel.StatsRealm = types.StringValue(frontend.StatsOptions.StatsRealm)
		statsOptionsModel.StatsAuth = types.StringValue(frontend.StatsOptions.StatsAuth)
		statsOptionsModel.StatsRefresh = types.StringValue(frontend.StatsOptions.StatsRefresh)
		state.StatsOptions, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"stats_enable":       types.BoolType,
			"stats_hide_version": types.BoolType,
			"stats_show_legends": types.BoolType,
			"stats_show_node":    types.BoolType,
			"stats_uri":          types.StringType,
			"stats_realm":        types.StringType,
			"stats_auth":         types.StringType,
			"stats_refresh":      types.StringType,
		}, statsOptionsModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	binds, err := r.client.ReadBinds(ctx, "frontend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading binds",
			"Could not read binds, unexpected error: "+err.Error(),
		)
		return
	}

	if len(binds) > 0 {
		var bindModels []bindResourceModel
		for _, bind := range binds {
			bm := bindResourceModel{
				Name:                 types.StringValue(bind.Name),
				Address:              types.StringValue(bind.Address),
				Transparent:          types.BoolValue(bind.Transparent),
				Mode:                 types.StringValue(bind.Mode),
				Maxconn:              types.Int64Value(bind.Maxconn),
				User:                 types.StringValue(bind.User),
				Group:                types.StringValue(bind.Group),
				ForceSslv3:           types.BoolValue(bind.ForceSslv3),
				ForceTlsv10:          types.BoolValue(bind.ForceTlsv10),
				ForceTlsv11:          types.BoolValue(bind.ForceTlsv11),
				ForceTlsv12:          types.BoolValue(bind.ForceTlsv12),
				ForceTlsv13:          types.BoolValue(bind.ForceTlsv13),
				ForceStrictSni:       types.StringValue(bind.ForceStrictSni),
				Ssl:                  types.BoolValue(bind.Ssl),
				SslCafile:            types.StringValue(bind.SslCafile),
				SslMaxVer:            types.StringValue(bind.SslMaxVer),
				SslMinVer:            types.StringValue(bind.SslMinVer),
				SslCertificate:       types.StringValue(bind.SslCertificate),
				Ciphers:              types.StringValue(bind.Ciphers),
				Ciphersuites:         types.StringValue(bind.Ciphersuites),
				AcceptProxy:          types.BoolValue(bind.AcceptProxy),
				Allow0rtt:            types.BoolValue(bind.Allow0rtt),
				Alpn:                 types.StringValue(bind.Alpn),
				Backlog:              types.StringValue(bind.Backlog),
				CaIgnoreErr:          types.StringValue(bind.CaIgnoreErr),
				CaSignFile:           types.StringValue(bind.CaSignFile),
				CaSignPass:           types.StringValue(bind.CaSignPass),
				CaVerifyFile:         types.StringValue(bind.CaVerifyFile),
				CrlFile:              types.StringValue(bind.CrlFile),
				CrtIgnoreErr:         types.StringValue(bind.CrtIgnoreErr),
				CrtList:              types.StringValue(bind.CrtList),
				DeferAccept:          types.BoolValue(bind.DeferAccept),
				ExposeViaAgent:       types.BoolValue(bind.ExposeViaAgent),
				GenerateCertificates: types.BoolValue(bind.GenerateCertificates),
				Gid:                  types.Int64Value(bind.Gid),
				Id:                   types.StringValue(bind.Id),
				Interface:            types.StringValue(bind.Interface),
				Level:                types.StringValue(bind.Level),
				LogProto:             types.StringValue(bind.LogProto),
				Mdev:                 types.StringValue(bind.Mdev),
				Namespace:            types.StringValue(bind.Namespace),
				Nice:                 types.Int64Value(bind.Nice),
				NoCaNames:            types.BoolValue(bind.NoCaNames),
				NoSslv3:              types.BoolValue(bind.NoSslv3),
				NoTlsv10:             types.BoolValue(bind.NoTlsv10),
				NoTlsv11:             types.BoolValue(bind.NoTlsv11),
				NoTlsv12:             types.BoolValue(bind.NoTlsv12),
				NoTlsv13:             types.BoolValue(bind.NoTlsv13),
				// New v3 fields
				Sslv3:  types.BoolValue(bind.Sslv3),
				Tlsv10: types.BoolValue(bind.Tlsv10),
				Tlsv11: types.BoolValue(bind.Tlsv11),
				Tlsv12: types.BoolValue(bind.Tlsv12),
				Tlsv13: types.BoolValue(bind.Tlsv13),

				Npn:                 types.StringValue(bind.Npn),
				PreferClientCiphers: types.BoolValue(bind.PreferClientCiphers),
				Process:             types.StringValue(bind.Process),
				Proto:               types.StringValue(bind.Proto),
				SeverityOutput:      types.StringValue(bind.SeverityOutput),
				StrictSni:           types.BoolValue(bind.StrictSni),
				TcpUserTimeout:      types.Int64Value(bind.TcpUserTimeout),
				Tfo:                 types.BoolValue(bind.Tfo),
				TlsTicketKeys:       types.StringValue(bind.TlsTicketKeys),
				Uid:                 types.StringValue(bind.Uid),
				V4v6:                types.BoolValue(bind.V4v6),
				V6only:              types.BoolValue(bind.V6only),
				Verify:              types.StringValue(bind.Verify),
				Metadata:            types.StringValue(bind.Metadata),
			}
			if bind.Port != nil {
				bm.Port = types.Int64Value(*bind.Port)
			}
			if bind.PortRangeEnd != nil {
				bm.PortRangeEnd = types.Int64Value(*bind.PortRangeEnd)
			}
			bindModels = append(bindModels, bm)
		}
		state.Binds, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: bindResourceModel{}.attrTypes(),
		}, bindModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if frontend.MonitorFail != nil {
		var monitorFailModels []struct {
			Cond     types.String `tfsdk:"cond"`
			CondTest types.String `tfsdk:"cond_test"`
		}
		monitorFailModels = append(monitorFailModels, struct {
			Cond     types.String `tfsdk:"cond"`
			CondTest types.String `tfsdk:"cond_test"`
		}{
			Cond:     types.StringValue(frontend.MonitorFail.Cond),
			CondTest: types.StringValue(frontend.MonitorFail.CondTest),
		})
		state.MonitorFail, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"cond":      types.StringType,
				"cond_test": types.StringType,
			},
		}, monitorFailModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	acls, err := r.client.ReadAcls(ctx, "frontend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading acls",
			"Could not read acls, unexpected error: "+err.Error(),
		)
		return
	}

	if len(acls) > 0 {
		var aclModels []frontendAclResourceModel
		for _, acl := range acls {
			aclModels = append(aclModels, frontendAclResourceModel{
				AclName:   types.StringValue(acl.AclName),
				Index:     types.Int64Value(acl.Index),
				Criterion: types.StringValue(acl.Criterion),
				Value:     types.StringValue(acl.Value),
			})
		}
		state.Acls, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: frontendAclResourceModel{}.attrTypes(),
		}, aclModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	httpRequestRules, err := r.client.ReadHttpRequestRules(ctx, "frontend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading http-request rules",
			"Could not read http-request rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(httpRequestRules) > 0 {
		var httpRequestRuleModels []httpRequestRuleResourceModel
		for _, httpRequestRule := range httpRequestRules {
			httpRequestRuleModels = append(httpRequestRuleModels, httpRequestRuleResourceModel{
				Index:      types.Int64Value(httpRequestRule.Index),
				Type:       types.StringValue(httpRequestRule.Type),
				Cond:       types.StringValue(httpRequestRule.Cond),
				CondTest:   types.StringValue(httpRequestRule.CondTest),
				HdrName:    types.StringValue(httpRequestRule.HdrName),
				HdrFormat:  types.StringValue(httpRequestRule.HdrFormat),
				RedirType:  types.StringValue(httpRequestRule.RedirType),
				RedirValue: types.StringValue(httpRequestRule.RedirValue),
			})
		}
		state.HttpRequestRules, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: httpRequestRuleResourceModel{}.attrTypes(),
		}, httpRequestRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	httpResponseRules, err := r.client.ReadHttpResponseRules(ctx, "frontend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading http-response rules",
			"Could not read http-response rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(httpResponseRules) > 0 {
		var httpResponseRuleModels []httpResponseRuleResourceModel
		for _, httpResponseRule := range httpResponseRules {
			httpResponseRuleModels = append(httpResponseRuleModels, httpResponseRuleResourceModel{
				Index:        types.Int64Value(httpResponseRule.Index),
				Type:         types.StringValue(httpResponseRule.Type),
				Cond:         types.StringValue(httpResponseRule.Cond),
				CondTest:     types.StringValue(httpResponseRule.CondTest),
				HdrName:      types.StringValue(httpResponseRule.HdrName),
				HdrFormat:    types.StringValue(httpResponseRule.HdrFormat),
				RedirType:    types.StringValue(httpResponseRule.RedirType),
				RedirValue:   types.StringValue(httpResponseRule.RedirValue),
				StatusCode:   types.Int64Value(httpResponseRule.StatusCode),
				StatusReason: types.StringValue(httpResponseRule.StatusReason),
			})
		}
		state.HttpResponseRules, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: httpResponseRuleResourceModel{}.attrTypes(),
		}, httpResponseRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tcpRequestRules, err := r.client.ReadTcpRequestRules(ctx, "frontend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tcp-request rules",
			"Could not read tcp-request rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(tcpRequestRules) > 0 {
		var tcpRequestRuleModels []frontendTcpRequestRuleResourceModel
		for _, tcpRequestRule := range tcpRequestRules {
			tcpRequestRuleModels = append(tcpRequestRuleModels, frontendTcpRequestRuleResourceModel{
				Index:        types.Int64Value(tcpRequestRule.Index),
				Type:         types.StringValue(tcpRequestRule.Type),
				Action:       types.StringValue(tcpRequestRule.Action),
				Cond:         types.StringValue(tcpRequestRule.Cond),
				CondTest:     types.StringValue(tcpRequestRule.CondTest),
				Timeout:      types.Int64Value(tcpRequestRule.Timeout),
				LuaAction:    types.StringValue(tcpRequestRule.LuaAction),
				LuaParams:    types.StringValue(tcpRequestRule.LuaParams),
				ScId:         types.Int64Value(tcpRequestRule.ScId),
				ScIdx:        types.Int64Value(tcpRequestRule.ScIdx),
				ScInt:        types.Int64Value(tcpRequestRule.ScInt),
				ScIncGpc0:    types.StringValue(tcpRequestRule.ScIncGpc0),
				ScIncGpc1:    types.StringValue(tcpRequestRule.ScIncGpc1),
				ScSetGpt0:    types.StringValue(tcpRequestRule.ScSetGpt0),
				TrackScKey:   types.StringValue(tcpRequestRule.TrackScKey),
				TrackScTable: types.StringValue(tcpRequestRule.TrackScTable),
				VarName:      types.StringValue(tcpRequestRule.VarName),
				VarScope:     types.StringValue(tcpRequestRule.VarScope),
				VarExpr:      types.StringValue(tcpRequestRule.VarExpr),
				VarFormat:    types.StringValue(tcpRequestRule.VarFormat),
				VarType:      types.StringValue(tcpRequestRule.VarType),
			})
		}
		state.TcpRequestRules, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: frontendTcpRequestRuleResourceModel{}.attrTypes(),
		}, tcpRequestRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tcpResponseRules, err := r.client.ReadTcpResponseRules(ctx, "frontend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tcp-response rules",
			"Could not read tcp-response rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(tcpResponseRules) > 0 {
		var tcpResponseRuleModels []frontendTcpResponseRuleResourceModel
		for _, tcpResponseRule := range tcpResponseRules {
			tcpResponseRuleModels = append(tcpResponseRuleModels, frontendTcpResponseRuleResourceModel{
				Index:     types.Int64Value(tcpResponseRule.Index),
				Action:    types.StringValue(tcpResponseRule.Action),
				Cond:      types.StringValue(tcpResponseRule.Cond),
				CondTest:  types.StringValue(tcpResponseRule.CondTest),
				LuaAction: types.StringValue(tcpResponseRule.LuaAction),
				LuaParams: types.StringValue(tcpResponseRule.LuaParams),
				ScId:      types.Int64Value(tcpResponseRule.ScId),
				ScIdx:     types.Int64Value(tcpResponseRule.ScIdx),
				ScInt:     types.Int64Value(tcpResponseRule.ScInt),
				ScIncGpc0: types.StringValue(tcpResponseRule.ScIncGpc0),
				ScIncGpc1: types.StringValue(tcpResponseRule.ScIncGpc1),
				ScSetGpt0: types.StringValue(tcpResponseRule.ScSetGpt0),
				VarName:   types.StringValue(tcpResponseRule.VarName),
				VarScope:  types.StringValue(tcpResponseRule.VarScope),
				VarExpr:   types.StringValue(tcpResponseRule.VarExpr),
				VarFormat: types.StringValue(tcpResponseRule.VarFormat),
				VarType:   types.StringValue(tcpResponseRule.VarType),
			})
		}
		state.TcpResponseRules, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: frontendTcpResponseRuleResourceModel{}.attrTypes(),
		}, tcpResponseRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *frontendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan frontendResourceModel
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

	payload := &FrontendPayload{
		Name:                     plan.Name.ValueString(),
		DefaultBackend:           plan.DefaultBackend.ValueString(),
		HttpConnectionMode:       plan.HttpConnectionMode.ValueString(),
		AcceptInvalidHttpRequest: plan.AcceptInvalidHttpRequest.ValueString(),
		MaxConn:                  plan.MaxConn.ValueInt64(),
		Mode:                     plan.Mode.ValueString(),
		Backlog:                  plan.Backlog.ValueInt64(),
		HttpKeepAliveTimeout:     plan.HttpKeepAliveTimeout.ValueInt64(),
		HttpRequestTimeout:       plan.HttpRequestTimeout.ValueInt64(),
		HttpUseProxyHeader:       plan.HttpUseProxyHeader.ValueString(),
		HttpLog:                  plan.HttpLog.ValueBool(),
		HttpsLog:                 plan.HttpsLog.ValueString(),
		ErrorLogFormat:           plan.ErrorLogFormat.ValueString(),
		LogFormat:                plan.LogFormat.ValueString(),
		LogFormatSd:              plan.LogFormatSd.ValueString(),
		MonitorUri:               plan.MonitorUri.ValueString(),
		TcpLog:                   plan.TcpLog.ValueBool(),
		From:                     plan.From.ValueString(),
		ClientTimeout:            plan.ClientTimeout.ValueInt64(),
		HttpUseHtx:               plan.HttpUseHtx.ValueString(),
		HttpIgnoreProbes:         plan.HttpIgnoreProbes.ValueString(),
		LogTag:                   plan.LogTag.ValueString(),
		Clflog:                   plan.Clflog.ValueBool(),
		Contstats:                plan.Contstats.ValueString(),
		Dontlognull:              plan.Dontlognull.ValueString(),
		LogSeparateErrors:        plan.LogSeparateErrors.ValueString(),
		OptionHttpServerClose:    plan.OptionHttpServerClose.ValueString(),
		OptionHttpclose:          plan.OptionHttpclose.ValueString(),
		OptionHttpKeepAlive:      plan.OptionHttpKeepAlive.ValueString(),
		OptionDontlogNormal:      plan.OptionDontlogNormal.ValueString(),
		OptionLogasap:            plan.OptionLogasap.ValueString(),
		OptionTcplog:             plan.OptionTcplog.ValueString(),
		OptionSocketStats:        plan.OptionSocketStats.ValueString(),
		OptionForwardfor:         plan.OptionForwardfor.ValueString(),
		TimeoutClient:            plan.TimeoutClient.ValueInt64(),
		TimeoutHttpKeepAlive:     plan.TimeoutHttpKeepAlive.ValueInt64(),
		TimeoutHttpRequest:       plan.TimeoutHttpRequest.ValueInt64(),
		TimeoutCont:              plan.TimeoutCont.ValueInt64(),
		TimeoutTarpit:            plan.TimeoutTarpit.ValueInt64(),
	}

	if !plan.StatsOptions.IsNull() {
		var statsOptionsModel struct {
			StatsEnable      types.Bool   `tfsdk:"stats_enable"`
			StatsHideVersion types.Bool   `tfsdk:"stats_hide_version"`
			StatsShowLegends types.Bool   `tfsdk:"stats_show_legends"`
			StatsShowNode    types.Bool   `tfsdk:"stats_show_node"`
			StatsUri         types.String `tfsdk:"stats_uri"`
			StatsRealm       types.String `tfsdk:"stats_realm"`
			StatsAuth        types.String `tfsdk:"stats_auth"`
			StatsRefresh     types.String `tfsdk:"stats_refresh"`
		}
		diags := plan.StatsOptions.As(ctx, &statsOptionsModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload.StatsOptions = StatsOptionsPayload{
			StatsEnable:      statsOptionsModel.StatsEnable.ValueBool(),
			StatsHideVersion: statsOptionsModel.StatsHideVersion.ValueBool(),
			StatsShowLegends: statsOptionsModel.StatsShowLegends.ValueBool(),
			StatsShowNode:    statsOptionsModel.StatsShowNode.ValueBool(),
			StatsUri:         statsOptionsModel.StatsUri.ValueString(),
			StatsRealm:       statsOptionsModel.StatsRealm.ValueString(),
			StatsAuth:        statsOptionsModel.StatsAuth.ValueString(),
			StatsRefresh:     statsOptionsModel.StatsRefresh.ValueString(),
		}
	}

	if !plan.MonitorFail.IsNull() {
		var monitorFailModels []struct {
			Cond     types.String `tfsdk:"cond"`
			CondTest types.String `tfsdk:"cond_test"`
		}
		diags := plan.MonitorFail.ElementsAs(ctx, &monitorFailModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(monitorFailModels) > 0 {
			payload.MonitorFail = &MonitorFailPayload{
				Cond:     monitorFailModels[0].Cond.ValueString(),
				CondTest: monitorFailModels[0].CondTest.ValueString(),
			}
		}
	}

	err := r.client.UpdateFrontend(ctx, plan.Name.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating frontend",
			"Could not update frontend, unexpected error: "+err.Error(),
		)
		return
	}

	if !plan.Binds.IsNull() {
		var planBinds []bindResourceModel
		diags := plan.Binds.ElementsAs(ctx, &planBinds, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateBinds []bindResourceModel
		if !req.State.Raw.IsNull() {
			var state frontendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.Binds.IsNull() {
				diags := state.Binds.ElementsAs(ctx, &stateBinds, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planBindsMap := make(map[string]bindResourceModel)
		for _, bind := range planBinds {
			planBindsMap[bind.Name.ValueString()] = bind
		}

		stateBindsMap := make(map[string]bindResourceModel)
		for _, bind := range stateBinds {
			stateBindsMap[bind.Name.ValueString()] = bind
		}

		for name, planBind := range planBindsMap {
			stateBind, ok := stateBindsMap[name]
			if !ok {
				// Create new bind
				bindPayload := &BindPayload{
					Name:                 planBind.Name.ValueString(),
					Address:              planBind.Address.ValueString(),
					Transparent:          planBind.Transparent.ValueBool(),
					Mode:                 planBind.Mode.ValueString(),
					Maxconn:              planBind.Maxconn.ValueInt64(),
					User:                 planBind.User.ValueString(),
					Group:                planBind.Group.ValueString(),
					ForceSslv3:           planBind.ForceSslv3.ValueBool(),
					ForceTlsv10:          planBind.ForceTlsv10.ValueBool(),
					ForceTlsv11:          planBind.ForceTlsv11.ValueBool(),
					ForceTlsv12:          planBind.ForceTlsv12.ValueBool(),
					ForceTlsv13:          planBind.ForceTlsv13.ValueBool(),
					ForceStrictSni:       planBind.ForceStrictSni.ValueString(),
					Ssl:                  planBind.Ssl.ValueBool(),
					SslCafile:            planBind.SslCafile.ValueString(),
					SslMaxVer:            planBind.SslMaxVer.ValueString(),
					SslMinVer:            planBind.SslMinVer.ValueString(),
					SslCertificate:       planBind.SslCertificate.ValueString(),
					Ciphers:              planBind.Ciphers.ValueString(),
					Ciphersuites:         planBind.Ciphersuites.ValueString(),
					AcceptProxy:          planBind.AcceptProxy.ValueBool(),
					Allow0rtt:            planBind.Allow0rtt.ValueBool(),
					Alpn:                 planBind.Alpn.ValueString(),
					Backlog:              planBind.Backlog.ValueString(),
					CaIgnoreErr:          planBind.CaIgnoreErr.ValueString(),
					CaSignFile:           planBind.CaSignFile.ValueString(),
					CaSignPass:           planBind.CaSignPass.ValueString(),
					CaVerifyFile:         planBind.CaVerifyFile.ValueString(),
					CrlFile:              planBind.CrlFile.ValueString(),
					CrtIgnoreErr:         planBind.CrtIgnoreErr.ValueString(),
					CrtList:              planBind.CrtList.ValueString(),
					DeferAccept:          planBind.DeferAccept.ValueBool(),
					ExposeViaAgent:       planBind.ExposeViaAgent.ValueBool(),
					GenerateCertificates: planBind.GenerateCertificates.ValueBool(),
					Gid:                  planBind.Gid.ValueInt64(),
					Id:                   planBind.Id.ValueString(),
					Interface:            planBind.Interface.ValueString(),
					Level:                planBind.Level.ValueString(),
					LogProto:             planBind.LogProto.ValueString(),
					Mdev:                 planBind.Mdev.ValueString(),
					Namespace:            planBind.Namespace.ValueString(),
					Nice:                 planBind.Nice.ValueInt64(),
					NoCaNames:            planBind.NoCaNames.ValueBool(),
					NoSslv3:              planBind.NoSslv3.ValueBool(),
					NoTlsv10:             planBind.NoTlsv10.ValueBool(),
					NoTlsv11:             planBind.NoTlsv11.ValueBool(),
					NoTlsv12:             planBind.NoTlsv12.ValueBool(),
					NoTlsv13:             planBind.NoTlsv13.ValueBool(),
					// New v3 fields
					Sslv3:               planBind.Sslv3.ValueBool(),
					Tlsv10:              planBind.Tlsv10.ValueBool(),
					Tlsv11:              planBind.Tlsv11.ValueBool(),
					Tlsv12:              planBind.Tlsv12.ValueBool(),
					Tlsv13:              planBind.Tlsv13.ValueBool(),
					Npn:                 planBind.Npn.ValueString(),
					PreferClientCiphers: planBind.PreferClientCiphers.ValueBool(),
					Process:             planBind.Process.ValueString(),
					Proto:               planBind.Proto.ValueString(),
					SeverityOutput:      planBind.SeverityOutput.ValueString(),
					StrictSni:           planBind.StrictSni.ValueBool(),
					TcpUserTimeout:      planBind.TcpUserTimeout.ValueInt64(),
					Tfo:                 planBind.Tfo.ValueBool(),
					TlsTicketKeys:       planBind.TlsTicketKeys.ValueString(),
					Uid:                 planBind.Uid.ValueString(),
					V4v6:                planBind.V4v6.ValueBool(),
					V6only:              planBind.V6only.ValueBool(),
					Verify:              planBind.Verify.ValueString(),
					Metadata:            planBind.Metadata.ValueString(),
				}
				if !planBind.Port.IsNull() {
					p := planBind.Port.ValueInt64()
					bindPayload.Port = &p
				}
				if !planBind.PortRangeEnd.IsNull() {
					p := planBind.PortRangeEnd.ValueInt64()
					bindPayload.PortRangeEnd = &p
				}
				err := r.client.CreateBind(ctx, "frontend", plan.Name.ValueString(), bindPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating bind",
						fmt.Sprintf("Could not create bind %s, unexpected error: %s", planBind.Name.ValueString(), err.Error()),
					)
					return
				}
			} else if !planBind.Address.Equal(stateBind.Address) || !planBind.Port.Equal(stateBind.Port) {
				// Update existing bind
				bindPayload := &BindPayload{
					Name:                 planBind.Name.ValueString(),
					Address:              planBind.Address.ValueString(),
					Transparent:          planBind.Transparent.ValueBool(),
					Mode:                 planBind.Mode.ValueString(),
					Maxconn:              planBind.Maxconn.ValueInt64(),
					User:                 planBind.User.ValueString(),
					Group:                planBind.Group.ValueString(),
					ForceSslv3:           planBind.ForceSslv3.ValueBool(),
					ForceTlsv10:          planBind.ForceTlsv10.ValueBool(),
					ForceTlsv11:          planBind.ForceTlsv11.ValueBool(),
					ForceTlsv12:          planBind.ForceTlsv12.ValueBool(),
					ForceTlsv13:          planBind.ForceTlsv13.ValueBool(),
					ForceStrictSni:       planBind.ForceStrictSni.ValueString(),
					Ssl:                  planBind.Ssl.ValueBool(),
					SslCafile:            planBind.SslCafile.ValueString(),
					SslMaxVer:            planBind.SslMaxVer.ValueString(),
					SslMinVer:            planBind.SslMinVer.ValueString(),
					SslCertificate:       planBind.SslCertificate.ValueString(),
					Ciphers:              planBind.Ciphers.ValueString(),
					Ciphersuites:         planBind.Ciphersuites.ValueString(),
					AcceptProxy:          planBind.AcceptProxy.ValueBool(),
					Allow0rtt:            planBind.Allow0rtt.ValueBool(),
					Alpn:                 planBind.Alpn.ValueString(),
					Backlog:              planBind.Backlog.ValueString(),
					CaIgnoreErr:          planBind.CaIgnoreErr.ValueString(),
					CaSignFile:           planBind.CaSignFile.ValueString(),
					CaSignPass:           planBind.CaSignPass.ValueString(),
					CaVerifyFile:         planBind.CaVerifyFile.ValueString(),
					CrlFile:              planBind.CrlFile.ValueString(),
					CrtIgnoreErr:         planBind.CrtIgnoreErr.ValueString(),
					CrtList:              planBind.CrtList.ValueString(),
					DeferAccept:          planBind.DeferAccept.ValueBool(),
					ExposeViaAgent:       planBind.ExposeViaAgent.ValueBool(),
					GenerateCertificates: planBind.GenerateCertificates.ValueBool(),
					Gid:                  planBind.Gid.ValueInt64(),
					Id:                   planBind.Id.ValueString(),
					Interface:            planBind.Interface.ValueString(),
					Level:                planBind.Level.ValueString(),
					LogProto:             planBind.LogProto.ValueString(),
					Mdev:                 planBind.Mdev.ValueString(),
					Namespace:            planBind.Namespace.ValueString(),
					Nice:                 planBind.Nice.ValueInt64(),
					NoCaNames:            planBind.NoCaNames.ValueBool(),
					NoSslv3:              planBind.NoSslv3.ValueBool(),
					NoTlsv10:             planBind.NoTlsv10.ValueBool(),
					NoTlsv11:             planBind.NoTlsv11.ValueBool(),
					NoTlsv12:             planBind.NoTlsv12.ValueBool(),
					NoTlsv13:             planBind.NoTlsv13.ValueBool(),
					// New v3 fields
					Sslv3:               planBind.Sslv3.ValueBool(),
					Tlsv10:              planBind.Tlsv10.ValueBool(),
					Tlsv11:              planBind.Tlsv11.ValueBool(),
					Tlsv12:              planBind.Tlsv12.ValueBool(),
					Tlsv13:              planBind.Tlsv13.ValueBool(),
					Npn:                 planBind.Npn.ValueString(),
					PreferClientCiphers: planBind.PreferClientCiphers.ValueBool(),
					Process:             planBind.Process.ValueString(),
					Proto:               planBind.Proto.ValueString(),
					SeverityOutput:      planBind.SeverityOutput.ValueString(),
					StrictSni:           planBind.StrictSni.ValueBool(),
					TcpUserTimeout:      planBind.TcpUserTimeout.ValueInt64(),
					Tfo:                 planBind.Tfo.ValueBool(),
					TlsTicketKeys:       planBind.TlsTicketKeys.ValueString(),
					Uid:                 planBind.Uid.ValueString(),
					V4v6:                planBind.V4v6.ValueBool(),
					V6only:              planBind.V6only.ValueBool(),
					Verify:              planBind.Verify.ValueString(),
					Metadata:            planBind.Metadata.ValueString(),
				}
				if !planBind.Port.IsNull() {
					p := planBind.Port.ValueInt64()
					bindPayload.Port = &p
				}
				if !planBind.PortRangeEnd.IsNull() {
					p := planBind.PortRangeEnd.ValueInt64()
					bindPayload.PortRangeEnd = &p
				}
				err := r.client.UpdateBind(ctx, name, "frontend", plan.Name.ValueString(), bindPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating bind",
						fmt.Sprintf("Could not update bind %s, unexpected error: %s", name, err.Error()),
					)
					return
				}
			}
		}

		for name := range stateBindsMap {
			if _, ok := planBindsMap[name]; !ok {
				// Delete bind
				err := r.client.DeleteBind(ctx, name, "frontend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting bind",
						fmt.Sprintf("Could not delete bind %s, unexpected error: %s", name, err.Error()),
					)
					return
				}
			}
		}
	}

	if !plan.Acls.IsNull() {
		var planAcls []frontendAclResourceModel
		diags := plan.Acls.ElementsAs(ctx, &planAcls, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateAcls []frontendAclResourceModel
		if !req.State.Raw.IsNull() {
			var state frontendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.Acls.IsNull() {
				diags := state.Acls.ElementsAs(ctx, &stateAcls, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planAclsMap := make(map[int64]frontendAclResourceModel)
		for _, acl := range planAcls {
			planAclsMap[acl.Index.ValueInt64()] = acl
		}

		stateAclsMap := make(map[int64]frontendAclResourceModel)
		for _, acl := range stateAcls {
			stateAclsMap[acl.Index.ValueInt64()] = acl
		}

		for index, planAcl := range planAclsMap {
			stateAcl, ok := stateAclsMap[index]
			if !ok {
				// Create new acl
				aclPayload := &AclPayload{
					AclName:   planAcl.AclName.ValueString(),
					Index:     planAcl.Index.ValueInt64(),
					Criterion: planAcl.Criterion.ValueString(),
					Value:     planAcl.Value.ValueString(),
				}
				err := r.client.CreateAcl(ctx, "frontend", plan.Name.ValueString(), aclPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating acl",
						fmt.Sprintf("Could not create acl %s, unexpected error: %s", planAcl.AclName.ValueString(), err.Error()),
					)
					return
				}
			} else if !planAcl.AclName.Equal(stateAcl.AclName) || !planAcl.Criterion.Equal(stateAcl.Criterion) || !planAcl.Value.Equal(stateAcl.Value) {
				// Update existing acl
				aclPayload := &AclPayload{
					AclName:   planAcl.AclName.ValueString(),
					Index:     planAcl.Index.ValueInt64(),
					Criterion: planAcl.Criterion.ValueString(),
					Value:     planAcl.Value.ValueString(),
				}
				err := r.client.UpdateAcl(ctx, index, "frontend", plan.Name.ValueString(), aclPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating acl",
						fmt.Sprintf("Could not update acl %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateAclsMap {
			if _, ok := planAclsMap[index]; !ok {
				// Delete acl
				err := r.client.DeleteAcl(ctx, index, "frontend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting acl",
						fmt.Sprintf("Could not delete acl %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}
	}

	if !plan.HttpRequestRules.IsNull() {
		var planHttpRequestRules []httpRequestRuleResourceModel
		diags := plan.HttpRequestRules.ElementsAs(ctx, &planHttpRequestRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateHttpRequestRules []httpRequestRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state frontendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.HttpRequestRules.IsNull() {
				diags := state.HttpRequestRules.ElementsAs(ctx, &stateHttpRequestRules, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planHttpRequestRulesMap := make(map[int64]httpRequestRuleResourceModel)
		for _, rule := range planHttpRequestRules {
			planHttpRequestRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateHttpRequestRulesMap := make(map[int64]httpRequestRuleResourceModel)
		for _, rule := range stateHttpRequestRules {
			stateHttpRequestRulesMap[rule.Index.ValueInt64()] = rule
		}

		for index, planRule := range planHttpRequestRulesMap {
			stateRule, ok := stateHttpRequestRulesMap[index]
			if !ok {
				// Create new http-request rule
				rulePayload := &HttpRequestRulePayload{
					Index:      planRule.Index.ValueInt64(),
					Type:       planRule.Type.ValueString(),
					Cond:       planRule.Cond.ValueString(),
					CondTest:   planRule.CondTest.ValueString(),
					HdrName:    planRule.HdrName.ValueString(),
					HdrFormat:  planRule.HdrFormat.ValueString(),
					RedirType:  planRule.RedirType.ValueString(),
					RedirValue: planRule.RedirValue.ValueString(),
				}
				err := r.client.CreateHttpRequestRule(ctx, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating http-request rule",
						fmt.Sprintf("Could not create http-request rule, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planRule.Type.Equal(stateRule.Type) || !planRule.Cond.Equal(stateRule.Cond) || !planRule.CondTest.Equal(stateRule.CondTest) || !planRule.HdrName.Equal(stateRule.HdrName) || !planRule.HdrFormat.Equal(stateRule.HdrFormat) || !planRule.RedirType.Equal(stateRule.RedirType) || !planRule.RedirValue.Equal(stateRule.RedirValue) {
				// Update existing http-request rule
				rulePayload := &HttpRequestRulePayload{
					Index:      planRule.Index.ValueInt64(),
					Type:       planRule.Type.ValueString(),
					Cond:       planRule.Cond.ValueString(),
					CondTest:   planRule.CondTest.ValueString(),
					HdrName:    planRule.HdrName.ValueString(),
					HdrFormat:  planRule.HdrFormat.ValueString(),
					RedirType:  planRule.RedirType.ValueString(),
					RedirValue: planRule.RedirValue.ValueString(),
				}
				err := r.client.UpdateHttpRequestRule(ctx, index, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating http-request rule",
						fmt.Sprintf("Could not update http-request rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateHttpRequestRulesMap {
			if _, ok := planHttpRequestRulesMap[index]; !ok {
				// Delete http-request rule
				err := r.client.DeleteHttpRequestRule(ctx, index, "frontend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting http-request rule",
						fmt.Sprintf("Could not delete http-request rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}
	}

	if !plan.HttpResponseRules.IsNull() {
		var planHttpResponseRules []httpResponseRuleResourceModel
		diags := plan.HttpResponseRules.ElementsAs(ctx, &planHttpResponseRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateHttpResponseRules []httpResponseRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state frontendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.HttpResponseRules.IsNull() {
				diags := state.HttpResponseRules.ElementsAs(ctx, &stateHttpResponseRules, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planHttpResponseRulesMap := make(map[int64]httpResponseRuleResourceModel)
		for _, rule := range planHttpResponseRules {
			planHttpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateHttpResponseRulesMap := make(map[int64]httpResponseRuleResourceModel)
		for _, rule := range stateHttpResponseRules {
			stateHttpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		for index, planRule := range planHttpResponseRulesMap {
			stateRule, ok := stateHttpResponseRulesMap[index]
			if !ok {
				// Create new http-response rule
				rulePayload := &HttpResponseRulePayload{
					Index:        planRule.Index.ValueInt64(),
					Type:         planRule.Type.ValueString(),
					Cond:         planRule.Cond.ValueString(),
					CondTest:     planRule.CondTest.ValueString(),
					HdrName:      planRule.HdrName.ValueString(),
					HdrFormat:    planRule.HdrFormat.ValueString(),
					RedirType:    planRule.RedirType.ValueString(),
					RedirValue:   planRule.RedirValue.ValueString(),
					StatusCode:   planRule.StatusCode.ValueInt64(),
					StatusReason: planRule.StatusReason.ValueString(),
				}
				err := r.client.CreateHttpResponseRule(ctx, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating http-response rule",
						fmt.Sprintf("Could not create http-response rule, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planRule.Type.Equal(stateRule.Type) || !planRule.Cond.Equal(stateRule.Cond) || !planRule.CondTest.Equal(stateRule.CondTest) || !planRule.HdrName.Equal(stateRule.HdrName) || !planRule.HdrFormat.Equal(stateRule.HdrFormat) || !planRule.RedirType.Equal(stateRule.RedirType) || !planRule.RedirValue.Equal(stateRule.RedirValue) {
				// Update existing http-response rule
				rulePayload := &HttpResponseRulePayload{
					Index:        planRule.Index.ValueInt64(),
					Type:         planRule.Type.ValueString(),
					Cond:         planRule.Cond.ValueString(),
					CondTest:     planRule.CondTest.ValueString(),
					HdrName:      planRule.HdrName.ValueString(),
					HdrFormat:    planRule.HdrFormat.ValueString(),
					RedirType:    planRule.RedirType.ValueString(),
					RedirValue:   planRule.RedirValue.ValueString(),
					StatusCode:   planRule.StatusCode.ValueInt64(),
					StatusReason: planRule.StatusReason.ValueString(),
				}
				err := r.client.UpdateHttpResponseRule(ctx, index, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating http-response rule",
						fmt.Sprintf("Could not update http-response rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateHttpResponseRulesMap {
			if _, ok := planHttpResponseRulesMap[index]; !ok {
				// Delete http-response rule
				err := r.client.DeleteHttpResponseRule(ctx, index, "frontend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting http-response rule",
						fmt.Sprintf("Could not delete http-response rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}
	}

	if !plan.TcpRequestRules.IsNull() {
		var planTcpRequestRules []frontendTcpRequestRuleResourceModel
		diags := plan.TcpRequestRules.ElementsAs(ctx, &planTcpRequestRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateTcpRequestRules []frontendTcpRequestRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state frontendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.TcpRequestRules.IsNull() {
				diags := state.TcpRequestRules.ElementsAs(ctx, &stateTcpRequestRules, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planTcpRequestRulesMap := make(map[int64]frontendTcpRequestRuleResourceModel)
		for _, rule := range planTcpRequestRules {
			planTcpRequestRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateTcpRequestRulesMap := make(map[int64]frontendTcpRequestRuleResourceModel)
		for _, rule := range stateTcpRequestRules {
			stateTcpRequestRulesMap[rule.Index.ValueInt64()] = rule
		}

		for index, planRule := range planTcpRequestRulesMap {
			stateRule, ok := stateTcpRequestRulesMap[index]
			if !ok {
				// Create new tcp-request rule
				rulePayload := &TcpRequestRulePayload{
					Index:        planRule.Index.ValueInt64(),
					Type:         planRule.Type.ValueString(),
					Action:       planRule.Action.ValueString(),
					Cond:         planRule.Cond.ValueString(),
					CondTest:     planRule.CondTest.ValueString(),
					Timeout:      planRule.Timeout.ValueInt64(),
					LuaAction:    planRule.LuaAction.ValueString(),
					LuaParams:    planRule.LuaParams.ValueString(),
					ScId:         planRule.ScId.ValueInt64(),
					ScIdx:        planRule.ScIdx.ValueInt64(),
					ScInt:        planRule.ScInt.ValueInt64(),
					ScIncGpc0:    planRule.ScIncGpc0.ValueString(),
					ScIncGpc1:    planRule.ScIncGpc1.ValueString(),
					ScSetGpt0:    planRule.ScSetGpt0.ValueString(),
					TrackScKey:   planRule.TrackScKey.ValueString(),
					TrackScTable: planRule.TrackScTable.ValueString(),
					VarName:      planRule.VarName.ValueString(),
					VarScope:     planRule.VarScope.ValueString(),
					VarExpr:      planRule.VarExpr.ValueString(),
					VarFormat:    planRule.VarFormat.ValueString(),
					VarType:      planRule.VarType.ValueString(),
				}
				err := r.client.CreateTcpRequestRule(ctx, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating tcp-request rule",
						fmt.Sprintf("Could not create tcp-request rule, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planRule.Type.Equal(stateRule.Type) || !planRule.Action.Equal(stateRule.Action) || !planRule.Cond.Equal(stateRule.Cond) || !planRule.CondTest.Equal(stateRule.CondTest) {
				// Update existing tcp-request rule
				rulePayload := &TcpRequestRulePayload{
					Index:        planRule.Index.ValueInt64(),
					Type:         planRule.Type.ValueString(),
					Action:       planRule.Action.ValueString(),
					Cond:         planRule.Cond.ValueString(),
					CondTest:     planRule.CondTest.ValueString(),
					Timeout:      planRule.Timeout.ValueInt64(),
					LuaAction:    planRule.LuaAction.ValueString(),
					LuaParams:    planRule.LuaParams.ValueString(),
					ScId:         planRule.ScId.ValueInt64(),
					ScIdx:        planRule.ScIdx.ValueInt64(),
					ScInt:        planRule.ScInt.ValueInt64(),
					ScIncGpc0:    planRule.ScIncGpc0.ValueString(),
					ScIncGpc1:    planRule.ScIncGpc1.ValueString(),
					ScSetGpt0:    planRule.ScSetGpt0.ValueString(),
					TrackScKey:   planRule.TrackScKey.ValueString(),
					TrackScTable: planRule.TrackScTable.ValueString(),
					VarName:      planRule.VarName.ValueString(),
					VarScope:     planRule.VarScope.ValueString(),
					VarExpr:      planRule.VarExpr.ValueString(),
					VarFormat:    planRule.VarFormat.ValueString(),
					VarType:      planRule.VarType.ValueString(),
				}
				err := r.client.UpdateTcpRequestRule(ctx, index, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating tcp-request rule",
						fmt.Sprintf("Could not update tcp-request rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateTcpRequestRulesMap {
			if _, ok := planTcpRequestRulesMap[index]; !ok {
				// Delete tcp-request rule
				err := r.client.DeleteTcpRequestRule(ctx, index, "frontend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting tcp-request rule",
						fmt.Sprintf("Could not delete tcp-request rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}
	}

	if !plan.TcpResponseRules.IsNull() {
		var planTcpResponseRules []frontendTcpResponseRuleResourceModel
		diags := plan.TcpResponseRules.ElementsAs(ctx, &planTcpResponseRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateTcpResponseRules []frontendTcpResponseRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state frontendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.TcpResponseRules.IsNull() {
				diags := state.TcpResponseRules.ElementsAs(ctx, &stateTcpResponseRules, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planTcpResponseRulesMap := make(map[int64]frontendTcpResponseRuleResourceModel)
		for _, rule := range planTcpResponseRules {
			planTcpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateTcpResponseRulesMap := make(map[int64]frontendTcpResponseRuleResourceModel)
		for _, rule := range stateTcpResponseRules {
			stateTcpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		for index, planRule := range planTcpResponseRulesMap {
			stateRule, ok := stateTcpResponseRulesMap[index]
			if !ok {
				// Create new tcp-response rule
				rulePayload := &TcpResponseRulePayload{
					Index:     planRule.Index.ValueInt64(),
					Action:    planRule.Action.ValueString(),
					Cond:      planRule.Cond.ValueString(),
					CondTest:  planRule.CondTest.ValueString(),
					LuaAction: planRule.LuaAction.ValueString(),
					LuaParams: planRule.LuaParams.ValueString(),
					ScId:      planRule.ScId.ValueInt64(),
					ScIdx:     planRule.ScIdx.ValueInt64(),
					ScInt:     planRule.ScInt.ValueInt64(),
					ScIncGpc0: planRule.ScIncGpc0.ValueString(),
					ScIncGpc1: planRule.ScIncGpc1.ValueString(),
					ScSetGpt0: planRule.ScSetGpt0.ValueString(),
					VarName:   planRule.VarName.ValueString(),
					VarScope:  planRule.VarScope.ValueString(),
					VarExpr:   planRule.VarExpr.ValueString(),
					VarFormat: planRule.VarFormat.ValueString(),
					VarType:   planRule.VarType.ValueString(),
				}
				err := r.client.CreateTcpResponseRule(ctx, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating tcp-response rule",
						fmt.Sprintf("Could not create tcp-response rule, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planRule.Action.Equal(stateRule.Action) || !planRule.Cond.Equal(stateRule.Cond) || !planRule.CondTest.Equal(stateRule.CondTest) {
				// Update existing tcp-response rule
				rulePayload := &TcpResponseRulePayload{
					Index:     planRule.Index.ValueInt64(),
					Action:    planRule.Action.ValueString(),
					Cond:      planRule.Cond.ValueString(),
					CondTest:  planRule.CondTest.ValueString(),
					LuaAction: planRule.LuaAction.ValueString(),
					LuaParams: planRule.LuaParams.ValueString(),
					ScId:      planRule.ScId.ValueInt64(),
					ScIdx:     planRule.ScIdx.ValueInt64(),
					ScInt:     planRule.ScInt.ValueInt64(),
					ScIncGpc0: planRule.ScIncGpc0.ValueString(),
					ScIncGpc1: planRule.ScIncGpc1.ValueString(),
					ScSetGpt0: planRule.ScSetGpt0.ValueString(),
					VarName:   planRule.VarName.ValueString(),
					VarScope:  planRule.VarScope.ValueString(),
					VarExpr:   planRule.VarExpr.ValueString(),
					VarFormat: planRule.VarFormat.ValueString(),
					VarType:   planRule.VarType.ValueString(),
				}
				err := r.client.UpdateTcpResponseRule(ctx, index, "frontend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating tcp-response rule",
						fmt.Sprintf("Could not update tcp-response rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateTcpResponseRulesMap {
			if _, ok := planTcpResponseRulesMap[index]; !ok {
				// Delete tcp-response rule
				err := r.client.DeleteTcpResponseRule(ctx, index, "frontend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting tcp-response rule",
						fmt.Sprintf("Could not delete tcp-response rule %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete resource.
func (r *frontendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state frontendResourceModel
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

	err := r.client.DeleteFrontend(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting frontend",
			"Could not delete frontend, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
