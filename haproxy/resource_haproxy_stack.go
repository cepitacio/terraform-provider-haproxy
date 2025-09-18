package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	apiVersion   string
}

// haproxyStackResourceModel maps the resource schema data.
type haproxyStackResourceModel struct {
	Name     types.String          `tfsdk:"name"`
	Backend  *haproxyBackendModel  `tfsdk:"backend"`
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
	Servers            map[string]haproxyServerModel  `tfsdk:"servers"` // Multiple servers
	Retries            types.Int64                    `tfsdk:"retries"`
	Balance            []haproxyBalanceModel          `tfsdk:"balance"`
	HttpchkParams      []haproxyHttpchkParamsModel    `tfsdk:"httpchk_params"`
	Forwardfor         []haproxyForwardforModel       `tfsdk:"forwardfor"`
	Httpchecks         []haproxyHttpcheckModel        `tfsdk:"http_checks"`
	TcpChecks          []haproxyTcpCheckModel         `tfsdk:"tcp_checks"`
	Acls               []haproxyAclModel              `tfsdk:"acls"`
	HttpRequestRules   []haproxyHttpRequestRuleModel  `tfsdk:"http_request_rules"`
	HttpResponseRules  []haproxyHttpResponseRuleModel `tfsdk:"http_response_rules"`
	TcpRequestRules    []haproxyTcpRequestRuleModel   `tfsdk:"tcp_request_rules"`
	TcpResponseRules   []haproxyTcpResponseRuleModel  `tfsdk:"tcp_response_rules"`
	DefaultServer      *haproxyDefaultServerModel     `tfsdk:"default_server"`
	StickTable         *haproxyStickTableModel        `tfsdk:"stick_table"`
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
	// Note: Name is now the map key, not a field
	Address        types.String `tfsdk:"address"`
	Port           types.Int64  `tfsdk:"port"`
	Check          types.String `tfsdk:"check"`
	Backup         types.String `tfsdk:"backup"`
	Maxconn        types.Int64  `tfsdk:"maxconn"`
	Weight         types.Int64  `tfsdk:"weight"`
	Rise           types.Int64  `tfsdk:"rise"`
	Fall           types.Int64  `tfsdk:"fall"`
	Inter          types.Int64  `tfsdk:"inter"`
	Fastinter      types.Int64  `tfsdk:"fastinter"`
	Downinter      types.Int64  `tfsdk:"downinter"`
	Ssl            types.String `tfsdk:"ssl"`
	SslCertificate types.String `tfsdk:"ssl_certificate"`
	SslCafile      types.String `tfsdk:"ssl_cafile"`
	SslMaxVer      types.String `tfsdk:"ssl_max_ver"`
	SslMinVer      types.String `tfsdk:"ssl_min_ver"`
	Verify         types.String `tfsdk:"verify"`
	Cookie         types.String `tfsdk:"cookie"`

	// SSL/TLS Protocol Control (v3 fields)
	Sslv3  types.String `tfsdk:"sslv3"`
	Tlsv10 types.String `tfsdk:"tlsv10"`
	Tlsv11 types.String `tfsdk:"tlsv11"`
	Tlsv12 types.String `tfsdk:"tlsv12"`
	Tlsv13 types.String `tfsdk:"tlsv13"`
	// SSL/TLS Protocol Control (deprecated v2 fields)
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

// haproxyFrontendModel maps the frontend block schema data.
type haproxyFrontendModel struct {
	Name              types.String                   `tfsdk:"name"`
	Mode              types.String                   `tfsdk:"mode"`
	DefaultBackend    types.String                   `tfsdk:"default_backend"`
	Maxconn           types.Int64                    `tfsdk:"maxconn"`
	Backlog           types.Int64                    `tfsdk:"backlog"`
	Ssl               types.Bool                     `tfsdk:"ssl"`
	SslCertificate    types.String                   `tfsdk:"ssl_certificate"`
	SslCafile         types.String                   `tfsdk:"ssl_cafile"`
	SslMaxVer         types.String                   `tfsdk:"ssl_max_ver"`
	SslMinVer         types.String                   `tfsdk:"ssl_min_ver"`
	Ciphers           types.String                   `tfsdk:"ciphers"`
	Ciphersuites      types.String                   `tfsdk:"ciphersuites"`
	Verify            types.String                   `tfsdk:"verify"`
	AcceptProxy       types.Bool                     `tfsdk:"accept_proxy"`
	DeferAccept       types.Bool                     `tfsdk:"defer_accept"`
	TcpUserTimeout    types.Int64                    `tfsdk:"tcp_user_timeout"`
	Tfo               types.Bool                     `tfsdk:"tfo"`
	V4v6              types.Bool                     `tfsdk:"v4v6"`
	V6only            types.Bool                     `tfsdk:"v6only"`
	MonitorUri        types.String                   `tfsdk:"monitor_uri"`
	Binds             map[string]haproxyBindModel    `tfsdk:"binds"`
	Acls              []haproxyAclModel              `tfsdk:"acls"`
	HttpRequestRules  []haproxyHttpRequestRuleModel  `tfsdk:"http_request_rules"`
	HttpResponseRules []haproxyHttpResponseRuleModel `tfsdk:"http_response_rules"`
	TcpRequestRules   []haproxyTcpRequestRuleModel   `tfsdk:"tcp_request_rules"`
	StatsOptions      []haproxyStatsOptionsModel     `tfsdk:"stats_options"`
	MonitorFail       []haproxyMonitorFailModel      `tfsdk:"monitor_fail"`
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
	Type            types.String `tfsdk:"type"`
	Addr            types.String `tfsdk:"addr"`
	Alpn            types.String `tfsdk:"alpn"`
	Body            types.String `tfsdk:"body"`
	BodyLogFormat   types.String `tfsdk:"body_log_format"`
	CheckComment    types.String `tfsdk:"check_comment"`
	Default         types.Bool   `tfsdk:"default"`
	ErrorStatus     types.String `tfsdk:"error_status"`
	ExclamationMark types.Bool   `tfsdk:"exclamation_mark"`
	Headers         types.List   `tfsdk:"headers"`
	Linger          types.Bool   `tfsdk:"linger"`
	Match           types.String `tfsdk:"match"`
	Method          types.String `tfsdk:"method"`
	MinRecv         types.Int64  `tfsdk:"min_recv"`
	OkStatus        types.String `tfsdk:"ok_status"`
	OnError         types.String `tfsdk:"on_error"`
	OnSuccess       types.String `tfsdk:"on_success"`
	Pattern         types.String `tfsdk:"pattern"`
	Port            types.Int64  `tfsdk:"port"`
	PortString      types.String `tfsdk:"port_string"`
	Proto           types.String `tfsdk:"proto"`
	SendProxy       types.Bool   `tfsdk:"send_proxy"`
	Sni             types.String `tfsdk:"sni"`
	Ssl             types.Bool   `tfsdk:"ssl"`
	StatusCode      types.String `tfsdk:"status_code"`
	ToutStatus      types.String `tfsdk:"tout_status"`
	Uri             types.String `tfsdk:"uri"`
	UriLogFormat    types.String `tfsdk:"uri_log_format"`
	VarExpr         types.String `tfsdk:"var_expr"`
	VarFormat       types.String `tfsdk:"var_format"`
	VarName         types.String `tfsdk:"var_name"`
	VarScope        types.String `tfsdk:"var_scope"`
	Version         types.String `tfsdk:"version"`
	ViaSocks4       types.Bool   `tfsdk:"via_socks4"`
}

// haproxyTcpCheckModel maps the tcp_check block schema data.
type haproxyTcpCheckModel struct {
	Action          types.String `tfsdk:"action"`
	Addr            types.String `tfsdk:"addr"`
	Alpn            types.String `tfsdk:"alpn"`
	CheckComment    types.String `tfsdk:"check_comment"`
	Data            types.String `tfsdk:"data"`
	Default         types.Bool   `tfsdk:"default"`
	ErrorStatus     types.String `tfsdk:"error_status"`
	ExclamationMark types.Bool   `tfsdk:"exclamation_mark"`
	Fmt             types.String `tfsdk:"fmt"`
	HexFmt          types.String `tfsdk:"hex_fmt"`
	HexString       types.String `tfsdk:"hex_string"`
	Linger          types.Bool   `tfsdk:"linger"`
	Match           types.String `tfsdk:"match"`
	MinRecv         types.Int64  `tfsdk:"min_recv"`
	OkStatus        types.String `tfsdk:"ok_status"`
	OnError         types.String `tfsdk:"on_error"`
	OnSuccess       types.String `tfsdk:"on_success"`
	Pattern         types.String `tfsdk:"pattern"`
	Port            types.Int64  `tfsdk:"port"`
	PortString      types.String `tfsdk:"port_string"`
	Proto           types.String `tfsdk:"proto"`
	SendProxy       types.Bool   `tfsdk:"send_proxy"`
	Sni             types.String `tfsdk:"sni"`
	Ssl             types.Bool   `tfsdk:"ssl"`
	StatusCode      types.String `tfsdk:"status_code"`
	ToutStatus      types.String `tfsdk:"tout_status"`
	VarExpr         types.String `tfsdk:"var_expr"`
	VarFmt          types.String `tfsdk:"var_fmt"`
	VarName         types.String `tfsdk:"var_name"`
	VarScope        types.String `tfsdk:"var_scope"`
	ViaSocks4       types.Bool   `tfsdk:"via_socks4"`
}

// haproxyHttpResponseRuleModel maps the http_response_rule block schema data.
type haproxyHttpResponseRuleModel struct {
	Index                types.Int64  `tfsdk:"index"` // For backward compatibility with existing state
	Type                 types.String `tfsdk:"type"`
	Action               types.String `tfsdk:"action"`
	RedirType            types.String `tfsdk:"redir_type"`
	RedirValue           types.String `tfsdk:"redir_value"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	HdrName              types.String `tfsdk:"hdr_name"`
	HdrFormat            types.String `tfsdk:"hdr_format"`
	HdrMatch             types.String `tfsdk:"hdr_match"`
	HdrMethod            types.String `tfsdk:"hdr_method"`
	RedirCode            types.Int64  `tfsdk:"redir_code"`
	RedirOption          types.String `tfsdk:"redir_option"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	AclFile              types.String `tfsdk:"acl_file"`
	AclKeyfmt            types.String `tfsdk:"acl_keyfmt"`
	CacheName            types.String `tfsdk:"cache_name"`
	CaptureId            types.Int64  `tfsdk:"capture_id"`
	CaptureLen           types.Int64  `tfsdk:"capture_len"`
	CaptureSample        types.String `tfsdk:"capture_sample"`
	DenyStatus           types.Int64  `tfsdk:"deny_status"`
	Expr                 types.String `tfsdk:"expr"`
	LogLevel             types.String `tfsdk:"log_level"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	MapFile              types.String `tfsdk:"map_file"`
	MapKeyfmt            types.String `tfsdk:"map_keyfmt"`
	MapValuefmt          types.String `tfsdk:"map_valuefmt"`
	MarkValue            types.String `tfsdk:"mark_value"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	ReturnContent        types.String `tfsdk:"return_content"`
	ReturnContentFormat  types.String `tfsdk:"return_content_format"`
	ReturnContentType    types.String `tfsdk:"return_content_type"`
	ReturnStatusCode     types.Int64  `tfsdk:"return_status_code"`
	RstTtl               types.Int64  `tfsdk:"rst_ttl"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScId                 types.Int64  `tfsdk:"sc_id"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	Status               types.Int64  `tfsdk:"status"`
	StatusReason         types.String `tfsdk:"status_reason"`
	StrictMode           types.String `tfsdk:"strict_mode"`
	Timeout              types.String `tfsdk:"timeout"`
	TimeoutType          types.String `tfsdk:"timeout_type"`
	TosValue             types.String `tfsdk:"tos_value"`
	TrackScKey           types.String `tfsdk:"track_sc_key"`
	TrackScStickCounter  types.Int64  `tfsdk:"track_sc_stick_counter"`
	TrackScTable         types.String `tfsdk:"track_sc_table"`
	VarExpr              types.String `tfsdk:"var_expr"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	WaitAtLeast          types.Int64  `tfsdk:"wait_at_least"`
	WaitTime             types.Int64  `tfsdk:"wait_time"`
	AuthRealm            types.String `tfsdk:"auth_realm"`
	HintName             types.String `tfsdk:"hint_name"`
	HintFormat           types.String `tfsdk:"hint_format"`
}

// haproxyTcpRequestRuleModel maps the tcp_request_rule block schema data.
type haproxyTcpRequestRuleModel struct {
	Type                 types.String `tfsdk:"type"`
	Action               types.String `tfsdk:"action"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	Expr                 types.String `tfsdk:"expr"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	LogLevel             types.String `tfsdk:"log_level"`
	MarkValue            types.String `tfsdk:"mark_value"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	TosValue             types.String `tfsdk:"tos_value"`
	CaptureLen           types.Int64  `tfsdk:"capture_len"`
	CaptureSample        types.String `tfsdk:"capture_sample"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	ResolveProtocol      types.String `tfsdk:"resolve_protocol"`
	ResolveResolvers     types.String `tfsdk:"resolve_resolvers"`
	ResolveVar           types.String `tfsdk:"resolve_var"`
	RstTtl               types.Int64  `tfsdk:"rst_ttl"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScIncId              types.String `tfsdk:"sc_inc_id"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	ServerName           types.String `tfsdk:"server_name"`
	ServiceName          types.String `tfsdk:"service_name"`
	VarName              types.String `tfsdk:"var_name"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarScope             types.String `tfsdk:"var_scope"`
	VarExpr              types.String `tfsdk:"var_expr"`
	Index                types.Int64  `tfsdk:"index"`
	// Additional fields from schema
	HdrName             types.String `tfsdk:"hdr_name"`
	HdrFormat           types.String `tfsdk:"hdr_format"`
	HdrMatch            types.String `tfsdk:"hdr_match"`
	HdrMethod           types.String `tfsdk:"hdr_method"`
	RedirType           types.String `tfsdk:"redir_type"`
	RedirValue          types.String `tfsdk:"redir_value"`
	RedirCode           types.Int64  `tfsdk:"redir_code"`
	RedirOption         types.String `tfsdk:"redir_option"`
	AclFile             types.String `tfsdk:"acl_file"`
	AclKeyfmt           types.String `tfsdk:"acl_keyfmt"`
	AuthRealm           types.String `tfsdk:"auth_realm"`
	CacheName           types.String `tfsdk:"cache_name"`
	CaptureId           types.Int64  `tfsdk:"capture_id"`
	DenyStatus          types.Int64  `tfsdk:"deny_status"`
	HintFormat          types.String `tfsdk:"hint_format"`
	HintName            types.String `tfsdk:"hint_name"`
	MapFile             types.String `tfsdk:"map_file"`
	MapKeyfmt           types.String `tfsdk:"map_keyfmt"`
	MapValuefmt         types.String `tfsdk:"map_valuefmt"`
	ReturnContent       types.String `tfsdk:"return_content"`
	ReturnContentFormat types.String `tfsdk:"return_content_format"`
	ReturnContentType   types.String `tfsdk:"return_content_type"`
	ReturnStatusCode    types.Int64  `tfsdk:"return_status_code"`
	ScExpr              types.String `tfsdk:"sc_expr"`
	ScId                types.Int64  `tfsdk:"sc_id"`
	SpoeEngine          types.String `tfsdk:"spoe_engine"`
	SpoeGroup           types.String `tfsdk:"spoe_group"`
	Status              types.Int64  `tfsdk:"status"`
	StatusReason        types.String `tfsdk:"status_reason"`
	StrictMode          types.String `tfsdk:"strict_mode"`
	TimeoutType         types.String `tfsdk:"timeout_type"`
	TrackScKey          types.String `tfsdk:"track_sc_key"`
	TrackScStickCounter types.Int64  `tfsdk:"track_sc_stick_counter"`
	TrackScTable        types.String `tfsdk:"track_sc_table"`
	WaitAtLeast         types.Int64  `tfsdk:"wait_at_least"`
	WaitTime            types.Int64  `tfsdk:"wait_time"`
}

// haproxyTcpResponseRuleModel maps the tcp_response_rule block schema data.
type haproxyTcpResponseRuleModel struct {
	Type                 types.String `tfsdk:"type"`
	Action               types.String `tfsdk:"action"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	Expr                 types.String `tfsdk:"expr"`
	LogLevel             types.String `tfsdk:"log_level"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	MarkValue            types.String `tfsdk:"mark_value"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	RstTtl               types.Int64  `tfsdk:"rst_ttl"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScId                 types.Int64  `tfsdk:"sc_id"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	TosValue             types.String `tfsdk:"tos_value"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	VarExpr              types.String `tfsdk:"var_expr"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	CaptureLen           types.Int64  `tfsdk:"capture_len"`
	CaptureSample        types.String `tfsdk:"capture_sample"`
	Index                types.Int64  `tfsdk:"index"`
	// Additional fields from schema
	HdrName             types.String `tfsdk:"hdr_name"`
	HdrFormat           types.String `tfsdk:"hdr_format"`
	HdrMatch            types.String `tfsdk:"hdr_match"`
	HdrMethod           types.String `tfsdk:"hdr_method"`
	RedirType           types.String `tfsdk:"redir_type"`
	RedirValue          types.String `tfsdk:"redir_value"`
	RedirCode           types.Int64  `tfsdk:"redir_code"`
	RedirOption         types.String `tfsdk:"redir_option"`
	AclFile             types.String `tfsdk:"acl_file"`
	AclKeyfmt           types.String `tfsdk:"acl_keyfmt"`
	AuthRealm           types.String `tfsdk:"auth_realm"`
	CacheName           types.String `tfsdk:"cache_name"`
	CaptureId           types.Int64  `tfsdk:"capture_id"`
	DenyStatus          types.Int64  `tfsdk:"deny_status"`
	HintFormat          types.String `tfsdk:"hint_format"`
	HintName            types.String `tfsdk:"hint_name"`
	MapFile             types.String `tfsdk:"map_file"`
	MapKeyfmt           types.String `tfsdk:"map_keyfmt"`
	MapValuefmt         types.String `tfsdk:"map_valuefmt"`
	ReturnContent       types.String `tfsdk:"return_content"`
	ReturnContentFormat types.String `tfsdk:"return_content_format"`
	ReturnContentType   types.String `tfsdk:"return_content_type"`
	ReturnStatusCode    types.Int64  `tfsdk:"return_status_code"`
	Status              types.Int64  `tfsdk:"status"`
	StatusReason        types.String `tfsdk:"status_reason"`
	StrictMode          types.String `tfsdk:"strict_mode"`
	TimeoutType         types.String `tfsdk:"timeout_type"`
	TrackScKey          types.String `tfsdk:"track_sc_key"`
	TrackScStickCounter types.Int64  `tfsdk:"track_sc_stick_counter"`
	TrackScTable        types.String `tfsdk:"track_sc_table"`
	WaitAtLeast         types.Int64  `tfsdk:"wait_at_least"`
	WaitTime            types.Int64  `tfsdk:"wait_time"`
}

// haproxyStickTableModel maps the stick_table block schema data.
type haproxyStickTableModel struct {
	Type    types.String `tfsdk:"type"`
	Size    types.Int64  `tfsdk:"size"`
	Expire  types.Int64  `tfsdk:"expire"`
	Nopurge types.Bool   `tfsdk:"nopurge"`
	Peers   types.String `tfsdk:"peers"`
}

// haproxyStatsOptionsModel maps the stats_options block schema data.
type haproxyStatsOptionsModel struct {
	StatsEnable types.Bool   `tfsdk:"stats_enable"`
	StatsUri    types.String `tfsdk:"stats_uri"`
	StatsRealm  types.String `tfsdk:"stats_realm"`
	StatsAuth   types.String `tfsdk:"stats_auth"`
}

// haproxyMonitorFailModel maps the monitor_fail block schema data.
type haproxyMonitorFailModel struct {
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// Metadata returns the resource type name.
func (r *haproxyStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Schema defines the schema for the resource.
func (r *haproxyStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Use default v3 if apiVersion is not set (latest version)
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
			"frontend": GetFrontendSchema(),
		},
		MarkdownDescription: "Manages a complete HAProxy stack including backend, server, frontend, and ACLs.\n\n## Example Usage\n\n```hcl\nresource \"haproxy_stack\" \"web_app\" {\n  name = \"web_application\"\n  \n  backend {\n    name = \"web_backend\"\n    mode = \"http\"\n    \n    # Backend ACLs\n    acls {\n      acl_name = \"is_api\"\n      criterion = \"path\"\n      value     = \"/api\"\n    }\n    \n    # HTTP request rules\n    http_request_rules {\n      type      = \"allow\"\n      cond      = \"if\"\n      cond_test = \"is_api\"\n    }\n    \n    # Health checks\n    http_checks {\n      type = \"connect\"\n      addr = \"127.0.0.1\"\n      port = 80\n    }\n    \n    # Servers (nested under backend)\n    servers = {\n      \"web_server_1\" = {\n        address = \"192.168.1.10\"\n        port    = 8080\n        check   = \"enabled\"\n        weight  = 100\n      }\n      \n      \"web_server_2\" = {\n        address = \"192.168.1.11\"\n        port    = 8080\n        check   = \"enabled\"\n        weight  = 100\n      }\n    }\n  }\n  \n  frontend {\n    name           = \"web_frontend\"\n    mode           = \"http\"\n    default_backend = \"web_backend\"\n    \n    # Frontend ACLs\n    acls {\n      acl_name = \"is_admin\"\n      criterion = \"path\"\n      value     = \"/admin\"\n    }\n    \n    # Bind configuration\n    binds = {\n      http_bind = {\n        address = \"0.0.0.0\"\n        port    = 80\n      }\n    }\n  }\n}\n```",
	}
}

// Configure adds the provider configured client to the resource.
func (r *haproxyStackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developer.", req.ProviderData),
		)
		return
	}

	// Initialize the stack manager with all required components
	aclManager := CreateACLManager(providerData.Client)
	frontendManager := CreateFrontendManager(providerData.Client)
	backendManager := CreateBackendManager(providerData.Client)
	r.stackManager = CreateStackManager(providerData.Client, aclManager, frontendManager, backendManager)

	// Store the API version for schema generation
	r.apiVersion = providerData.APIVersion
}

// Create resource.
func (r *haproxyStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Validate configuration based on API version before creating
	if err := r.validateConfigForAPIVersion(ctx, req, resp); err != nil {
		return // Validation failed, don't proceed with creation
	}

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
	// Validate configuration based on API version before updating
	if err := r.validateConfigForAPIVersionUpdate(ctx, req, resp); err != nil {
		return // Validation failed, don't proceed with update
	}

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

// validateConfigForAPIVersion validates the configuration based on API version
func (r *haproxyStackResource) validateConfigForAPIVersion(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) error {
	// Get the API version from the provider configuration
	apiVersion := r.apiVersion
	if apiVersion == "" {
		apiVersion = "v3" // Default to v3
	}

	// Get the configuration data
	var config haproxyStackResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return fmt.Errorf("failed to get configuration")
	}

	// Validate based on API version
	if apiVersion == "v2" {
		// Check for v3 fields that are not supported in v2
		if config.Backend != nil && config.Backend.DefaultServer != nil {
			validateDefaultServerV2ForCreate(ctx, &resp.Diagnostics, config.Backend.DefaultServer, "backend.default_server")
		}

		if config.Backend != nil && len(config.Backend.Servers) > 0 {
			for serverName, server := range config.Backend.Servers {
				validateServerV2ForCreate(ctx, &resp.Diagnostics, &server, fmt.Sprintf("backend.servers[%s]", serverName))
			}
		}

		if config.Frontend != nil {
			for bindName, bind := range config.Frontend.Binds {
				validateBindV2ForCreate(ctx, &resp.Diagnostics, bind, fmt.Sprintf("frontend.binds[%s]", bindName))
			}

		}
	} else if apiVersion == "v3" {
		// Check for v2 fields that are deprecated in v3
		if config.Backend != nil && config.Backend.DefaultServer != nil {
			validateDefaultServerV3ForCreate(ctx, &resp.Diagnostics, config.Backend.DefaultServer, "backend.default_server")
		}

		if config.Backend != nil && len(config.Backend.Servers) > 0 {
			for serverName, server := range config.Backend.Servers {
				validateServerV3ForCreate(ctx, &resp.Diagnostics, &server, fmt.Sprintf("backend.servers[%s]", serverName))
			}
		}

		if config.Frontend != nil {
			for bindName, bind := range config.Frontend.Binds {
				validateBindV3ForCreate(ctx, &resp.Diagnostics, bind, fmt.Sprintf("frontend.binds[%s]", bindName))
			}
		}
	}

	// Check if validation produced any errors
	if resp.Diagnostics.HasError() {
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

// validateConfigForAPIVersionUpdate validates the configuration based on API version for updates
func (r *haproxyStackResource) validateConfigForAPIVersionUpdate(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) error {
	// Get the API version from the provider configuration
	apiVersion := r.apiVersion
	if apiVersion == "" {
		apiVersion = "v3" // Default to v3
	}

	// Get the configuration data
	var config haproxyStackResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return fmt.Errorf("failed to get configuration")
	}

	// Validate based on API version
	if apiVersion == "v2" {
		// Check for v3 fields that are not supported in v2
		if config.Backend != nil && config.Backend.DefaultServer != nil {
			validateDefaultServerV2ForCreate(ctx, &resp.Diagnostics, config.Backend.DefaultServer, "backend.default_server")
		}

		if config.Backend != nil && len(config.Backend.Servers) > 0 {
			for serverName, server := range config.Backend.Servers {
				validateServerV2ForCreate(ctx, &resp.Diagnostics, &server, fmt.Sprintf("backend.servers[%s]", serverName))
			}
		}

		if config.Frontend != nil {
			for bindName, bind := range config.Frontend.Binds {
				validateBindV2ForCreate(ctx, &resp.Diagnostics, bind, fmt.Sprintf("frontend.binds[%s]", bindName))
			}

		}
	} else if apiVersion == "v3" {
		// Check for v2 fields that are deprecated in v3
		if config.Backend != nil && config.Backend.DefaultServer != nil {
			validateDefaultServerV3ForCreate(ctx, &resp.Diagnostics, config.Backend.DefaultServer, "backend.default_server")
		}

		if config.Backend != nil && len(config.Backend.Servers) > 0 {
			for serverName, server := range config.Backend.Servers {
				validateServerV3ForCreate(ctx, &resp.Diagnostics, &server, fmt.Sprintf("backend.servers[%s]", serverName))
			}
		}

		if config.Frontend != nil {
			for bindName, bind := range config.Frontend.Binds {
				validateBindV3ForCreate(ctx, &resp.Diagnostics, bind, fmt.Sprintf("frontend.binds[%s]", bindName))
			}
		}
	}

	// Check if validation produced any errors
	if resp.Diagnostics.HasError() {
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

// Create-specific validation functions that work with diag.Diagnostics
func validateDefaultServerV2ForCreate(ctx context.Context, diags *diag.Diagnostics, defaultServer *haproxyDefaultServerModel, pathPrefix string) {
	if !defaultServer.Sslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("sslv3"),
			"Unsupported field in v2",
			"Field 'sslv3' is not supported in Data Plane API v2. Use 'no_sslv3' or 'force_sslv3' instead.",
		)
	}
	if !defaultServer.Tlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv10"),
			"Unsupported field in v2",
			"Field 'tlsv10' is not supported in Data Plane API v2. Use 'no_tlsv10' or 'force_tlsv10' instead.",
		)
	}
	if !defaultServer.Tlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv11"),
			"Unsupported field in v2",
			"Field 'tlsv11' is not supported in Data Plane API v2. Use 'no_tlsv11' or 'force_tlsv11' instead.",
		)
	}
	if !defaultServer.Tlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv12"),
			"Unsupported field in v2",
			"Field 'tlsv12' is not supported in Data Plane API v2. Use 'no_tlsv12' or 'force_tlsv12' instead.",
		)
	}
	if !defaultServer.Tlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv13"),
			"Unsupported field in v2",
			"Field 'tlsv13' is not supported in Data Plane API v2. Use 'no_tlsv13' or 'force_tlsv13' instead.",
		)
	}
}

// validateServerV2ForCreate validates that v3 fields are not used in v2 mode
func validateServerV2ForCreate(ctx context.Context, diags *diag.Diagnostics, server *haproxyServerModel, pathPrefix string) {
	if !server.Sslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("sslv3"),
			"Unsupported field in v2",
			"Field 'sslv3' is not supported in Data Plane API v2. Use 'no_sslv3' or 'force_sslv3' instead.",
		)
	}
	if !server.Tlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv10"),
			"Unsupported field in v2",
			"Field 'tlsv10' is not supported in Data Plane API v2. Use 'no_tlsv10' or 'force_tlsv10' instead.",
		)
	}
	if !server.Tlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv11"),
			"Unsupported field in v2",
			"Field 'tlsv11' is not supported in Data Plane API v2. Use 'no_tlsv11' or 'force_tlsv11' instead.",
		)
	}
	if !server.Tlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv12"),
			"Unsupported field in v2",
			"Field 'tlsv12' is not supported in Data Plane API v2. Use 'no_tlsv12' or 'force_tlsv12' instead.",
		)
	}
	if !server.Tlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv13"),
			"Unsupported field in v2",
			"Field 'tlsv13' is not supported in Data Plane API v2. Use 'no_tlsv13' or 'force_tlsv13' instead.",
		)
	}
}

// validateBindV2ForCreate validates that v3 fields are not used in v2 mode
func validateBindV2ForCreate(ctx context.Context, diags *diag.Diagnostics, bind haproxyBindModel, pathPrefix string) {
	if !bind.Sslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("sslv3"),
			"Unsupported field in v2",
			"Field 'sslv3' is not supported in Data Plane API v2. Use 'force_sslv3' instead.",
		)
	}
	if !bind.Tlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv10"),
			"Unsupported field in v2",
			"Field 'tlsv10' is not supported in Data Plane API v2. Use 'force_tlsv10' instead.",
		)
	}
	if !bind.Tlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv11"),
			"Unsupported field in v2",
			"Field 'tlsv11' is not supported in Data Plane API v2. Use 'force_tlsv11' instead.",
		)
	}
	if !bind.Tlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv12"),
			"Unsupported field in v2",
			"Field 'tlsv12' is not supported in Data Plane API v2. Use 'force_tlsv12' instead.",
		)
	}
	if !bind.Tlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv13"),
			"Unsupported field in v2",
			"Field 'tlsv13' is not supported in Data Plane API v2. Use 'force_tlsv13' instead.",
		)
	}
}

// validateDefaultServerV3ForCreate validates that deprecated v2 fields are not used in v3 mode
func validateDefaultServerV3ForCreate(ctx context.Context, diags *diag.Diagnostics, defaultServer *haproxyDefaultServerModel, pathPrefix string) {
	if !defaultServer.NoSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_sslv3"),
			"Invalid field in v3 default-server",
			"Field 'no_sslv3' is not accepted in default-server sections in Data Plane API v3. Use 'sslv3' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv10"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv10' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv10' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv11"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv11' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv11' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv12"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv12' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv12' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv13"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv13' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv13' in individual server sections instead.",
		)
	}
	if !defaultServer.ForceSslv3.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_sslv3"),
			"Deprecated field in v3 default-server",
			"Field 'force_sslv3' is deprecated in Data Plane API v3. It will be converted to 'sslv3' automatically.",
		)
	}
	if !defaultServer.ForceTlsv10.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv10"),
			"Deprecated field in v3 default-server",
			"Field 'force_tlsv10' is deprecated in Data Plane API v3. It will be converted to 'tlsv10' automatically.",
		)
	}
	if !defaultServer.ForceTlsv11.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv11"),
			"Deprecated field in v3 default-server",
			"Field 'force_tlsv11' is deprecated in Data Plane API v3. It will be converted to 'tlsv11' automatically.",
		)
	}
	if !defaultServer.ForceTlsv12.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv12"),
			"Deprecated field in v3 default-server",
			"Field 'force_tlsv12' is deprecated in Data Plane API v3. It will be converted to 'tlsv12' automatically.",
		)
	}
	if !defaultServer.ForceTlsv13.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv13"),
			"Deprecated field in v3 default-server",
			"Field 'force_tlsv13' is deprecated in Data Plane API v3. It will be converted to 'tlsv13' automatically.",
		)
	}
}

// validateServerV3ForCreate validates that deprecated v2 fields are not used in v3 mode
func validateServerV3ForCreate(ctx context.Context, diags *diag.Diagnostics, server *haproxyServerModel, pathPrefix string) {
	if !server.NoSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_sslv3"),
			"Deprecated field in v3",
			"Field 'no_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !server.NoTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv10"),
			"Deprecated field in v3",
			"Field 'no_tlsv10' is deprecated in Data Plane API v3. Use 'tlsv10' instead.",
		)
	}
	if !server.NoTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv11"),
			"Deprecated field in v3",
			"Field 'no_tlsv11' is deprecated in Data Plane API v3. Use 'tlsv11' instead.",
		)
	}
	if !server.NoTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv12"),
			"Deprecated field in v3",
			"Field 'no_tlsv12' is deprecated in Data Plane API v3. Use 'tlsv12' instead.",
		)
	}
	if !server.NoTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv13"),
			"Deprecated field in v3",
			"Field 'no_tlsv13' is deprecated in Data Plane API v3. Use 'tlsv13' instead.",
		)
	}
	if !server.ForceSslv3.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_sslv3"),
			"Deprecated field in v3",
			"Field 'force_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !server.ForceTlsv10.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv10"),
			"Deprecated field in v3",
			"Field 'force_tlsv10' is deprecated in Data Plane API v3. Use 'tlsv10' instead.",
		)
	}
	if !server.ForceTlsv11.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv11"),
			"Deprecated field in v3",
			"Field 'force_tlsv11' is deprecated in Data Plane API v3. Use 'tlsv11' instead.",
		)
	}
	if !server.ForceTlsv12.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv12"),
			"Deprecated field in v3",
			"Field 'force_tlsv12' is deprecated in Data Plane API v3. Use 'tlsv12' instead.",
		)
	}
	if !server.ForceTlsv13.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv13"),
			"Deprecated field in v3",
			"Field 'force_tlsv13' is deprecated in Data Plane API v3. Use 'tlsv13' instead.",
		)
	}
}

// validateBindV3ForCreate validates that deprecated v2 fields are not used in v3 mode
func validateBindV3ForCreate(ctx context.Context, diags *diag.Diagnostics, bind haproxyBindModel, pathPrefix string) {
	if !bind.NoSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_sslv3"),
			"Deprecated field in v3",
			"Field 'no_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !bind.ForceSslv3.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_sslv3"),
			"Deprecated field in v3",
			"Field 'force_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !bind.ForceTlsv10.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv10"),
			"Deprecated field in v3",
			"Field 'force_tlsv10' is deprecated in Data Plane API v3. Use 'tlsv10' instead.",
		)
	}
	if !bind.ForceTlsv11.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv11"),
			"Deprecated field in v3",
			"Field 'force_tlsv11' is deprecated in Data Plane API v3. Use 'tlsv11' instead.",
		)
	}
	if !bind.ForceTlsv12.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv12"),
			"Deprecated field in v3",
			"Field 'force_tlsv12' is deprecated in Data Plane API v3. Use 'tlsv12' instead.",
		)
	}
	if !bind.ForceTlsv13.IsNull() {
		diags.AddAttributeWarning(
			path.Root(pathPrefix).AtName("force_tlsv13"),
			"Deprecated field in v3",
			"Field 'force_tlsv13' is deprecated in Data Plane API v3. Use 'tlsv13' instead.",
		)
	}
}
