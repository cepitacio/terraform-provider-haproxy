package haproxy

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &backendResource{}
)

// NewBackendResource is a helper function to simplify the provider implementation.
func NewBackendResource() resource.Resource {
	return &backendResource{}
}

// backendResource is the resource implementation.
type backendResource struct {
	client *HAProxyClient
}

// backendResourceModel maps the resource schema data.
type backendResourceModel struct {
	Name               types.String `tfsdk:"name"`
	Mode               types.String `tfsdk:"mode"`
	Forwardfor         types.Object `tfsdk:"forwardfor"`
	Balance            types.Object `tfsdk:"balance"`
	HttpchkParams      types.Object `tfsdk:"httpchk_params"`
	HttpConnectionMode types.String `tfsdk:"http_connection_mode"`
	Acls               types.List   `tfsdk:"acl"`
	HttpRequestRules   types.List   `tfsdk:"http_request_rule"`
	HttpResponseRules  types.List   `tfsdk:"http_response_rule"`
	TcpRequestRules    types.List   `tfsdk:"tcp_request_rule"`
	TcpResponseRules   types.List   `tfsdk:"tcp_response_rule"`
	Httpchecks         types.List   `tfsdk:"httpcheck"`
	TcpChecks          types.List   `tfsdk:"tcp_check"`
	AdvCheck           types.String `tfsdk:"adv_check"`
	ServerTimeout      types.Int64  `tfsdk:"server_timeout"`
	CheckTimeout       types.Int64  `tfsdk:"check_timeout"`
	ConnectTimeout     types.Int64  `tfsdk:"connect_timeout"`
	QueueTimeout       types.Int64  `tfsdk:"queue_timeout"`
	TunnelTimeout      types.Int64  `tfsdk:"tunnel_timeout"`
	TarpitTimeout      types.Int64  `tfsdk:"tarpit_timeout"`
	CheckCache         types.String `tfsdk:"checkcache"`
	Retries            types.Int64  `tfsdk:"retries"`

	// Default Server Configuration (for SSL/TLS settings)
	DefaultServer types.Object `tfsdk:"default_server"`
	StickTable    types.Object `tfsdk:"stick_table"`
	StickRules    types.List   `tfsdk:"stick_rule"`
	StatsOptions  types.Object `tfsdk:"stats_options"`
}

func (b backendResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                 types.StringType,
		"mode":                 types.StringType,
		"forwardfor":           types.ObjectType{},
		"balance":              types.ObjectType{},
		"httpchk_params":       types.ObjectType{},
		"http_connection_mode": types.StringType,
		"acl":                  types.ListType{ElemType: types.ObjectType{}},
		"http_request_rule":    types.ListType{ElemType: types.ObjectType{}},
		"http_response_rule":   types.ListType{ElemType: types.ObjectType{}},
		"tcp_request_rule":     types.ListType{ElemType: types.ObjectType{}},
		"tcp_response_rule":    types.ListType{ElemType: types.ObjectType{}},
		"httpcheck":            types.ListType{ElemType: types.ObjectType{}},
		"tcp_check":            types.ListType{ElemType: types.ObjectType{}},
		"adv_check":            types.StringType,
		"server_timeout":       types.Int64Type,
		"check_timeout":        types.Int64Type,
		"connect_timeout":      types.Int64Type,
		"queue_timeout":        types.Int64Type,
		"tunnel_timeout":       types.Int64Type,
		"tarpit_timeout":       types.Int64Type,
		"checkcache":           types.StringType,
		"retries":              types.Int64Type,

		// Default Server Configuration (for SSL/TLS settings)
		"default_server": types.ObjectType{},
		"stick_table":    types.ObjectType{},
		"stick_rule":     types.ListType{ElemType: types.ObjectType{}},
		"stats_options":  types.ObjectType{},
	}
}

// backendAclResourceModel maps the resource schema data.
type backendAclResourceModel struct {
	AclName   types.String `tfsdk:"acl_name"`
	Index     types.Int64  `tfsdk:"index"`
	Criterion types.String `tfsdk:"criterion"`
	Value     types.String `tfsdk:"value"`
}

func (a backendAclResourceModel) GetIndex() int64 {
	return a.Index.ValueInt64()
}

func (a backendAclResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"acl_name":  types.StringType,
		"index":     types.Int64Type,
		"criterion": types.StringType,
		"value":     types.StringType,
	}
}

// backendHttpRequestRuleResourceModel maps the resource schema data.
type backendHttpRequestRuleResourceModel struct {
	Index                types.Int64  `tfsdk:"index"`
	Type                 types.String `tfsdk:"type"`
	AclFile              types.String `tfsdk:"acl_file"`
	AclKeyfmt            types.String `tfsdk:"acl_keyfmt"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	CacheName            types.String `tfsdk:"cache_name"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	Expr                 types.String `tfsdk:"expr"`
	HdrFormat            types.String `tfsdk:"hdr_format"`
	HdrMatch             types.String `tfsdk:"hdr_match"`
	HdrMethod            types.String `tfsdk:"hdr_method"`
	HdrName              types.String `tfsdk:"hdr_name"`
	LogLevel             types.String `tfsdk:"log_level"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	MapFile              types.String `tfsdk:"map_file"`
	MapKeyfmt            types.String `tfsdk:"map_keyfmt"`
	MapValuefmt          types.String `tfsdk:"map_valuefmt"`
	MarkValue            types.String `tfsdk:"mark_value"`
	MethodFmt            types.String `tfsdk:"method_fmt"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	PathFmt              types.String `tfsdk:"path_fmt"`
	PathMatch            types.String `tfsdk:"path_match"`
	QueryFmt             types.String `tfsdk:"query_fmt"`
	RedirCode            types.Int64  `tfsdk:"redir_code"`
	RedirType            types.String `tfsdk:"redir_type"`
	RedirValue           types.String `tfsdk:"redir_value"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScId                 types.Int64  `tfsdk:"sc_id"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	Service              types.String `tfsdk:"service"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	StatusCode           types.Int64  `tfsdk:"status_code"`
	StatusReason         types.String `tfsdk:"status_reason"`
	Timeout              types.String `tfsdk:"timeout"`
	TimeoutValue         types.Int64  `tfsdk:"timeout_value"`
	TosValue             types.String `tfsdk:"tos_value"`
	TrackScKey           types.String `tfsdk:"track_sc_key"`
	TrackScTable         types.String `tfsdk:"track_sc_table"`
	UriFmt               types.String `tfsdk:"uri_fmt"`
	UriMatch             types.String `tfsdk:"uri_match"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	WaitTime             types.Int64  `tfsdk:"wait_time"`
}

func (h backendHttpRequestRuleResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h backendHttpRequestRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":                  types.Int64Type,
		"type":                   types.StringType,
		"acl_file":               types.StringType,
		"acl_keyfmt":             types.StringType,
		"bandwidth_limit_name":   types.StringType,
		"bandwidth_limit_period": types.StringType,
		"bandwidth_limit_limit":  types.StringType,
		"cache_name":             types.StringType,
		"cond":                   types.StringType,
		"cond_test":              types.StringType,
		"expr":                   types.StringType,
		"hdr_format":             types.StringType,
		"hdr_match":              types.StringType,
		"hdr_method":             types.StringType,
		"hdr_name":               types.StringType,
		"log_level":              types.StringType,
		"lua_action":             types.StringType,
		"lua_params":             types.StringType,
		"map_file":               types.StringType,
		"map_keyfmt":             types.StringType,
		"map_valuefmt":           types.StringType,
		"mark_value":             types.StringType,
		"method_fmt":             types.StringType,
		"nice_value":             types.Int64Type,
		"path_fmt":               types.StringType,
		"path_match":             types.StringType,
		"query_fmt":              types.StringType,
		"redir_code":             types.Int64Type,
		"redir_type":             types.StringType,
		"redir_value":            types.StringType,
		"sc_expr":                types.StringType,
		"sc_id":                  types.Int64Type,
		"sc_idx":                 types.Int64Type,
		"sc_int":                 types.Int64Type,
		"service":                types.StringType,
		"spoe_engine":            types.StringType,
		"spoe_group":             types.StringType,
		"status_code":            types.Int64Type,
		"status_reason":          types.StringType,
		"timeout":                types.StringType,
		"timeout_value":          types.Int64Type,
		"tos_value":              types.StringType,
		"track_sc_key":           types.StringType,
		"track_sc_table":         types.StringType,
		"uri_fmt":                types.StringType,
		"uri_match":              types.StringType,
		"var_name":               types.StringType,
		"var_scope":              types.StringType,
		"wait_time":              types.Int64Type,
	}
}

// backendHttpResponseRuleResourceModel maps the resource schema data.
type backendHttpResponseRuleResourceModel struct {
	Index                types.Int64  `tfsdk:"index"`
	Type                 types.String `tfsdk:"type"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	HdrName              types.String `tfsdk:"hdr_name"`
	HdrFormat            types.String `tfsdk:"hdr_format"`
	RedirType            types.String `tfsdk:"redir_type"`
	RedirValue           types.String `tfsdk:"redir_value"`
	StatusCode           types.Int64  `tfsdk:"status_code"`
	StatusReason         types.String `tfsdk:"status_reason"`
	RedirCode            types.Int64  `tfsdk:"redir_code"`
	HdrMethod            types.String `tfsdk:"hdr_method"`
	PathFmt              types.String `tfsdk:"path_fmt"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScId                 types.Int64  `tfsdk:"sc_id"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	CacheName            types.String `tfsdk:"cache_name"`
	MapKeyfmt            types.String `tfsdk:"map_keyfmt"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	TrackScKey           types.String `tfsdk:"track_sc_key"`
	UriFmt               types.String `tfsdk:"uri_fmt"`
	UriMatch             types.String `tfsdk:"uri_match"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	MethodFmt            types.String `tfsdk:"method_fmt"`
	AclKeyfmt            types.String `tfsdk:"acl_keyfmt"`
	PathMatch            types.String `tfsdk:"path_match"`
	TosValue             types.String `tfsdk:"tos_value"`
	AclFile              types.String `tfsdk:"acl_file"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	QueryFmt             types.String `tfsdk:"query_fmt"`
	MapFile              types.String `tfsdk:"map_file"`
	MapValuefmt          types.String `tfsdk:"map_valuefmt"`
	MarkValue            types.String `tfsdk:"mark_value"`
	Service              types.String `tfsdk:"service"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	TrackScTable         types.String `tfsdk:"track_sc_table"`
	LogLevel             types.String `tfsdk:"log_level"`
	LuaAction            types.String `tfsdk:"lua_action"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	Timeout              types.String `tfsdk:"timeout"`
	Expr                 types.String `tfsdk:"expr"`
	LuaParams            types.String `tfsdk:"lua_params"`
	TimeoutValue         types.Int64  `tfsdk:"timeout_value"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	WaitTime             types.Int64  `tfsdk:"wait_time"`
	HdrMatch             types.String `tfsdk:"hdr_match"`
}

func (h backendHttpResponseRuleResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h backendHttpResponseRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":                  types.Int64Type,
		"type":                   types.StringType,
		"cond":                   types.StringType,
		"cond_test":              types.StringType,
		"hdr_name":               types.StringType,
		"hdr_format":             types.StringType,
		"redir_type":             types.StringType,
		"redir_value":            types.StringType,
		"status_code":            types.Int64Type,
		"status_reason":          types.StringType,
		"redir_code":             types.Int64Type,
		"hdr_method":             types.StringType,
		"path_fmt":               types.StringType,
		"sc_expr":                types.StringType,
		"sc_id":                  types.Int64Type,
		"bandwidth_limit_period": types.StringType,
		"cache_name":             types.StringType,
		"map_keyfmt":             types.StringType,
		"sc_idx":                 types.Int64Type,
		"track_sc_key":           types.StringType,
		"uri_fmt":                types.StringType,
		"uri_match":              types.StringType,
		"bandwidth_limit_limit":  types.StringType,
		"method_fmt":             types.StringType,
		"acl_keyfmt":             types.StringType,
		"path_match":             types.StringType,
		"tos_value":              types.StringType,
		"acl_file":               types.StringType,
		"bandwidth_limit_name":   types.StringType,
		"nice_value":             types.Int64Type,
		"query_fmt":              types.StringType,
		"map_file":               types.StringType,
		"map_valuefmt":           types.StringType,
		"mark_value":             types.StringType,
		"service":                types.StringType,
		"spoe_engine":            types.StringType,
		"track_sc_table":         types.StringType,
		"log_level":              types.StringType,
		"lua_action":             types.StringType,
		"sc_int":                 types.Int64Type,
		"spoe_group":             types.StringType,
		"timeout":                types.StringType,
		"expr":                   types.StringType,
		"lua_params":             types.StringType,
		"timeout_value":          types.Int64Type,
		"var_name":               types.StringType,
		"var_scope":              types.StringType,
		"wait_time":              types.Int64Type,
		"hdr_match":              types.StringType,
	}
}

// backendTcpRequestRuleResourceModel maps the resource schema data.
type backendTcpRequestRuleResourceModel struct {
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

func (t backendTcpRequestRuleResourceModel) GetIndex() int64 {
	return t.Index.ValueInt64()
}

func (t backendTcpRequestRuleResourceModel) attrTypes() map[string]attr.Type {
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

// backendTcpResponseRuleResourceModel maps the resource schema data.
type backendTcpResponseRuleResourceModel struct {
	Index     types.Int64  `tfsdk:"index"`
	Type      types.String `tfsdk:"type"`
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

func (t backendTcpResponseRuleResourceModel) GetIndex() int64 {
	return t.Index.ValueInt64()
}

func (t backendTcpResponseRuleResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":       types.Int64Type,
		"type":        types.StringType,
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

// backendHttpcheckResourceModel maps the resource schema data.
type backendHttpcheckResourceModel struct {
	Index           types.Int64  `tfsdk:"index"`
	Addr            types.String `tfsdk:"addr"`
	Match           types.String `tfsdk:"match"`
	Pattern         types.String `tfsdk:"pattern"`
	Type            types.String `tfsdk:"type"`
	Method          types.String `tfsdk:"method"`
	Port            types.Int64  `tfsdk:"port"`
	Uri             types.String `tfsdk:"uri"`
	Version         types.String `tfsdk:"version"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	ExclamationMark types.String `tfsdk:"exclamation_mark"`
	LogLevel        types.String `tfsdk:"log_level"`
	SendProxy       types.String `tfsdk:"send_proxy"`
	ViaSocks4       types.String `tfsdk:"via_socks4"`
	CheckComment    types.String `tfsdk:"check_comment"`
}

func (h backendHttpcheckResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h backendHttpcheckResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":            types.Int64Type,
		"addr":             types.StringType,
		"match":            types.StringType,
		"pattern":          types.StringType,
		"type":             types.StringType,
		"method":           types.StringType,
		"port":             types.Int64Type,
		"uri":              types.StringType,
		"version":          types.StringType,
		"timeout":          types.Int64Type,
		"exclamation_mark": types.StringType,
		"log_level":        types.StringType,
		"send_proxy":       types.StringType,
		"via_socks4":       types.StringType,
		"check_comment":    types.StringType,
	}
}

// backendTcpCheckResourceModel maps the resource schema data.
type backendTcpCheckResourceModel struct {
	Index      types.Int64  `tfsdk:"index"`
	Action     types.String `tfsdk:"action"`
	Comment    types.String `tfsdk:"comment"`
	Cond       types.String `tfsdk:"cond"`
	CondTest   types.String `tfsdk:"cond_test"`
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

func (t backendTcpCheckResourceModel) GetIndex() int64 {
	return t.Index.ValueInt64()
}

func (t backendTcpCheckResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":       types.Int64Type,
		"action":      types.StringType,
		"comment":     types.StringType,
		"cond":        types.StringType,
		"cond_test":   types.StringType,
		"port":        types.Int64Type,
		"address":     types.StringType,
		"data":        types.StringType,
		"min_recv":    types.Int64Type,
		"on_success":  types.StringType,
		"on_error":    types.StringType,
		"status_code": types.StringType,
		"timeout":     types.Int64Type,
		"log_level":   types.StringType,
	}
}

// Metadata returns the resource type name.
func (r *backendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend"
}

// Schema defines the schema for the resource.
func (r *backendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the backend. It must be unique and cannot be changed.",
			},
			"mode": schema.StringAttribute{
				Optional:    true,
				Description: "The mode of the backend. Allowed: http|tcp",
			},
			"http_connection_mode": schema.StringAttribute{
				Optional:    true,
				Description: "The http connection mode of the backend. Allowed: httpclose|http-server-close|http-keep-alive",
			},
			"adv_check": schema.StringAttribute{
				Optional:    true,
				Description: "The advanced check of the backend. Allowed: ssl-hello-chk|smtpchk|ldap-check|mysql-check|pgsql-check|tcp-check|redis-check",
			},
			"server_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The server timeout of the backend.",
			},
			"check_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The check timeout of the backend.",
			},
			"connect_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The connect timeout of the backend.",
			},
			"queue_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The queue timeout of the backend.",
			},
			"tunnel_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The tunnel timeout of the backend.",
			},
			"tarpit_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The tarpit timeout of the backend.",
			},
			"checkcache": schema.StringAttribute{
				Optional:    true,
				Description: "The checkcache of the backend.",
			},
			"retries": schema.Int64Attribute{
				Optional:    true,
				Description: "The retries of the backend.",
			},
			"no_sslv3": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable SSLv3 (deprecated in API v2)",
			},
			"no_tlsv10": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.0 (deprecated in API v2)",
			},
			"no_tlsv11": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.1 (deprecated in API v2)",
			},
			"no_tlsv12": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.2 (deprecated in API v2)",
			},
			"no_tlsv13": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.3 (deprecated in API v2)",
			},
			"force_sslv3": schema.BoolAttribute{
				Optional:    true,
				Description: "Force SSLv3 (deprecated in API v2)",
			},
			"force_tlsv10": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.0 (deprecated in API v2)",
			},
			"force_tlsv11": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.1 (deprecated in API v2)",
			},
			"force_tlsv12": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.2 (deprecated in API v2)",
			},
			"force_tlsv13": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.3 (deprecated in API v2)",
			},
			"force_strict_sni": schema.StringAttribute{
				Optional:    true,
				Description: "Force strict SNI (deprecated in API v2)",
			},

			// SSL/TLS Configuration (v3 fields)
			"sslv3": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable SSLv3 (API v3, replaces no_sslv3)",
			},
			"tlsv10": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable TLSv1.0 (API v3, replaces no_tlsv10)",
			},
			"tlsv11": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable TLSv1.1 (API v3, replaces no_tlsv11)",
			},
			"tlsv12": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable TLSv1.2 (API v3, replaces no_tlsv12)",
			},
			"tlsv13": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable TLSv1.3 (API v3, replaces no_tlsv13)",
			},

			"ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable SSL/TLS",
			},
			"ssl_cafile": schema.StringAttribute{
				Optional:    true,
				Description: "SSL CA file path",
			},
			"ssl_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "SSL certificate file path",
			},
			"ssl_max_ver": schema.StringAttribute{
				Optional:    true,
				Description: "SSL maximum version",
			},
			"ssl_min_ver": schema.StringAttribute{
				Optional:    true,
				Description: "SSL minimum version",
			},
			"ssl_reuse": schema.StringAttribute{
				Optional:    true,
				Description: "SSL reuse setting",
			},
			"ciphers": schema.StringAttribute{
				Optional:    true,
				Description: "SSL ciphers to support",
			},
			"ciphersuites": schema.StringAttribute{
				Optional:    true,
				Description: "SSL ciphersuites to support",
			},
			"verify": schema.StringAttribute{
				Optional:    true,
				Description: "SSL certificate verification. Allowed: none|required",
			},
		},
		Blocks: map[string]schema.Block{
			"forwardfor": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"enabled": schema.StringAttribute{
						Required:    true,
						Description: "The state of the forwardfor. Allowed: enabled|disabled",
					},
				},
			},
			"balance": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"algorithm": schema.StringAttribute{
						Required:    true,
						Description: "The algorithm of the balance. Allowed: roundrobin|static-rr|leastconn|first|source|uri|url_param|hdr|rdp-cookie",
					},
					"url_param": schema.StringAttribute{
						Optional:    true,
						Description: "The url_param of the balance.",
					},
				},
			},
			"httpchk_params": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Optional:    true,
						Description: "The method of the httpchk_params. Allowed: HEAD|PUT|POST|GET|TRACE|OPTIONS",
					},
					"uri": schema.StringAttribute{
						Optional:    true,
						Description: "The uri of the httpchk_params.",
					},
					"version": schema.StringAttribute{
						Optional:    true,
						Description: "The version of the httpchk_params.",
					},
				},
			},
			"httpcheck": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the httpcheck",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the httpcheck",
						},
						"method": schema.StringAttribute{
				Optional:    true,
							Description: "The HTTP method for the health check",
			},
						"uri": schema.StringAttribute{
				Optional:    true,
							Description: "The URI for the health check",
			},
						"version": schema.StringAttribute{
				Optional:    true,
							Description: "The HTTP version for the health check",
			},
						"timeout": schema.Int64Attribute{
				Optional:    true,
							Description: "The timeout for the health check in milliseconds",
			},
						"match": schema.StringAttribute{
				Optional:    true,
							Description: "The match condition for the health check",
			},
						"pattern": schema.StringAttribute{
				Optional:    true,
							Description: "The pattern to match for the health check",
			},
						"addr": schema.StringAttribute{
				Optional:    true,
							Description: "The address for the health check",
			},
						"port": schema.Int64Attribute{
				Optional:    true,
							Description: "The port for the health check",
			},
						"exclamation_mark": schema.StringAttribute{
				Optional:    true,
							Description: "The exclamation mark for the health check",
			},
						"log_level": schema.StringAttribute{
				Optional:    true,
							Description: "The log level for the health check",
						},
						"send_proxy": schema.StringAttribute{
							Optional:    true,
							Description: "The send proxy for the health check",
						},
						"via_socks4": schema.StringAttribute{
							Optional:    true,
							Description: "The via socks4 for the health check",
						},
						"check_comment": schema.StringAttribute{
							Optional:    true,
							Description: "The check comment for the health check",
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
			"http_request_rule": schema.ListNestedBlock{
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
						"redir_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The redirection code of the http-request rule",
						},
						"status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The status code of the http-request rule",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-request rule",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method of the http-request rule",
						},
						"path_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The path format of the http-request rule",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The sc expression of the http-request rule",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc id of the http-request rule",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period of the http-request rule",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name of the http-request rule",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map key format of the http-request rule",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc index of the http-request rule",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc key of the http-request rule",
						},
						"uri_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The uri format of the http-request rule",
						},
						"uri_match": schema.StringAttribute{
							Optional:    true,
							Description: "The uri match of the http-request rule",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit of the http-request rule",
						},
						"method_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The method format of the http-request rule",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The acl key format of the http-request rule",
						},
						"path_match": schema.StringAttribute{
							Optional:    true,
							Description: "The path match of the http-request rule",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The tos value of the http-request rule",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The acl file of the http-request rule",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name of the http-request rule",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value of the http-request rule",
						},
						"query_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The query format of the http-request rule",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file of the http-request rule",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map value format of the http-request rule",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value of the http-request rule",
						},
						"service": schema.StringAttribute{
							Optional:    true,
							Description: "The service of the http-request rule",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The spoe engine of the http-request rule",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc table of the http-request rule",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level of the http-request rule",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The lua action of the http-request rule",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc int of the http-request rule",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The spoe group of the http-request rule",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout of the http-request rule",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression of the http-request rule",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The lua params of the http-request rule",
						},
						"timeout_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout value of the http-request rule",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name of the http-request rule",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope of the http-request rule",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time of the http-request rule",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match of the http-request rule",
					},
				},
			},
			},
			"http_response_rule": schema.ListNestedBlock{
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
						"status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The status code of the http-response rule",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-response rule",
						},
						"redir_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The redirection code of the http-response rule",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type of the http-response rule",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value of the http-response rule",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method of the http-response rule",
						},
						"path_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The path format of the http-response rule",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The sc expression of the http-response rule",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc id of the http-response rule",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period of the http-response rule",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name of the http-response rule",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map key format of the http-response rule",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc index of the http-response rule",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc key of the http-response rule",
						},
						"uri_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The uri format of the http-response rule",
						},
						"uri_match": schema.StringAttribute{
							Optional:    true,
							Description: "The uri match of the http-response rule",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit of the http-response rule",
						},
						"method_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The method format of the http-response rule",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The acl key format of the http-response rule",
						},
						"path_match": schema.StringAttribute{
							Optional:    true,
							Description: "The path match of the http-response rule",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The tos value of the http-response rule",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The acl file of the http-response rule",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name of the http-response rule",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value of the http-response rule",
						},
						"query_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The query format of the http-response rule",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file of the http-response rule",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map value format of the http-response rule",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value of the http-response rule",
						},
						"service": schema.StringAttribute{
							Optional:    true,
							Description: "The service of the http-response rule",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The spoe engine of the http-response rule",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc table of the http-response rule",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level of the http-response rule",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The lua action of the http-response rule",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc int of the http-response rule",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The spoe group of the http-response rule",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout of the http-response rule",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expression of the http-response rule",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The lua params of the http-response rule",
						},
						"timeout_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout value of the http-response rule",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The variable name of the http-response rule",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope of the http-response rule",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time of the http-response rule",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match of the http-response rule",
						},
					},
				},
			},
			"tcp_request_rule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
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
			"tcp_response_rule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the tcp-response rule",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the tcp-response rule",
						},
						"action": schema.StringAttribute{
							Optional:    true,
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
							Description: "The variable name of the tcp-response rule",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The variable scope of the tcp-response rule",
						},
						"var_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The variable expression of the tcp-response rule",
						},
						"var_format": schema.StringAttribute{
							Optional:    true,
							Description: "The variable format of the tcp-response rule",
						},
						"var_type": schema.StringAttribute{
							Optional:    true,
							Description: "The variable type of the tcp-response rule",
						},
					},
				},
			},
			"tcp_check": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the tcp check",
						},
						"action": schema.StringAttribute{
							Optional:    true,
							Description: "The action of the tcp check",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the tcp check",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the tcp check",
						},
						"comment": schema.StringAttribute{
							Optional:    true,
							Description: "The comment of the tcp check",
						},
						"port": schema.Int64Attribute{
							Optional:    true,
							Description: "The port of the tcp check",
						},
						"address": schema.StringAttribute{
							Optional:    true,
							Description: "The address of the tcp check",
						},
						"data": schema.StringAttribute{
							Optional:    true,
							Description: "The data of the tcp check",
						},
						"min_recv": schema.Int64Attribute{
							Optional:    true,
							Description: "The minimum receive size of the tcp check",
						},
						"on_success": schema.StringAttribute{
							Optional:    true,
							Description: "The action on success of the tcp check",
						},
						"on_error": schema.StringAttribute{
							Optional:    true,
							Description: "The action on error of the tcp check",
						},
						"status_code": schema.StringAttribute{
							Optional:    true,
							Description: "The status code of the tcp check",
						},
						"timeout": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout of the tcp check in milliseconds",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level of the tcp check",
						},
					},
				},
			},
			"stick_table": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:    true,
						Description: "The type of the stick table. Allowed: ip|ipv6|string|integer",
					},
					"size": schema.Int64Attribute{
						Required:    true,
						Description: "The size of the stick table",
					},
					"expire": schema.StringAttribute{
				Optional:    true,
						Description: "The expire time of the stick table",
			},
					"store": schema.StringAttribute{
				Optional:    true,
						Description: "The store configuration of the stick table",
					},
				},
			},
			"stick_rule": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the stick rule",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the stick rule. Allowed: match|on|if|unless",
						},
						"table": schema.StringAttribute{
				Optional:    true,
							Description: "The table name for the stick rule",
			},
						"pattern": schema.StringAttribute{
				Optional:    true,
							Description: "The pattern for the stick rule",
			},
						"cond": schema.StringAttribute{
				Optional:    true,
							Description: "The condition for the stick rule",
			},
						"cond_test": schema.StringAttribute{
				Optional:    true,
							Description: "The condition test for the stick rule",
			},
			},
			},
			},
			"stats_options": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"stats_enable": schema.BoolAttribute{
				Optional:    true,
						Description: "Enable statistics",
			},
					"stats_hide_version": schema.BoolAttribute{
				Optional:    true,
						Description: "Hide version in statistics",
			},
					"stats_show_legends": schema.BoolAttribute{
				Optional:    true,
						Description: "Show legends in statistics",
			},
					"stats_show_node": schema.BoolAttribute{
				Optional:    true,
						Description: "Show node in statistics",
			},
					"stats_uri": schema.StringAttribute{
				Optional:    true,
						Description: "Statistics URI",
			},
					"stats_realm": schema.StringAttribute{
				Optional:    true,
						Description: "Statistics realm",
			},
					"stats_auth": schema.StringAttribute{
				Optional:    true,
						Description: "Statistics authentication",
			},
					"stats_refresh": schema.StringAttribute{
				Optional:    true,
						Description: "Statistics refresh rate",
			},
			},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *backendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *backendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendResourceModel
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

	payload := &BackendPayload{
		Name:               plan.Name.ValueString(),
		Mode:               plan.Mode.ValueString(),
		HttpConnectionMode: plan.HttpConnectionMode.ValueString(),
		AdvCheck:           plan.AdvCheck.ValueString(),
		ServerTimeout:      plan.ServerTimeout.ValueInt64(),
		CheckTimeout:       plan.CheckTimeout.ValueInt64(),
		ConnectTimeout:     plan.ConnectTimeout.ValueInt64(),
		QueueTimeout:       plan.QueueTimeout.ValueInt64(),
		TunnelTimeout:      plan.TunnelTimeout.ValueInt64(),
		TarpitTimeout:      plan.TarpitTimeout.ValueInt64(),
		CheckCache:         plan.CheckCache.ValueString(),
		Retries:            plan.Retries.ValueInt64(),

		// Default Server Configuration (for SSL/TLS settings)
		DefaultServer: &DefaultServerPayload{
			// SSL/TLS Configuration
			Ssl:            "enabled", // Default value
			SslCafile:      "",
			SslCertificate: "",
			SslMaxVer:      "",
			SslMinVer:      "",
			SslReuse:       "",
			Ciphers:        "",
			Ciphersuites:   "",
			Verify:         "",

			// SSL/TLS Protocol Control (v3 fields)
			Sslv3:  "disabled",
			Tlsv10: "disabled",
			Tlsv11: "disabled",
			Tlsv12: "enabled",
			Tlsv13: "enabled",

			// SSL/TLS Protocol Control (deprecated v2 fields)
			NoSslv3:        "enabled",
			NoTlsv10:       "enabled",
			NoTlsv11:       "enabled",
			NoTlsv12:       "disabled",
			NoTlsv13:       "disabled",
			ForceSslv3:     "disabled",
			ForceTlsv10:    "disabled",
			ForceTlsv11:    "disabled",
			ForceTlsv12:    "disabled",
			ForceTlsv13:    "disabled",
			ForceStrictSni: "",
		},
	}

	if !plan.Forwardfor.IsNull() {
		var forwardforModel struct {
			Enabled types.String `tfsdk:"enabled"`
		}
		diags := plan.Forwardfor.As(ctx, &forwardforModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload.Forwardfor = &ForwardFor{
			Enabled: forwardforModel.Enabled.ValueString(),
		}
	}

	if !plan.Balance.IsNull() {
		var balanceModel struct {
			Algorithm types.String `tfsdk:"algorithm"`
			UrlParam  types.String `tfsdk:"url_param"`
		}
		diags := plan.Balance.As(ctx, &balanceModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload.Balance = &Balance{
			Algorithm: balanceModel.Algorithm.ValueString(),
			UrlParam:  balanceModel.UrlParam.ValueString(),
		}
	}

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
		payload.HttpchkParams = &HttpchkParams{
			Method:  httpchkParamsModel.Method.ValueString(),
			Uri:     httpchkParamsModel.Uri.ValueString(),
			Version: httpchkParamsModel.Version.ValueString(),
		}
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
		payload.StatsOptions = &StatsOptionsPayload{
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

	err := r.client.CreateBackend(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating backend",
			"Could not create backend, unexpected error: "+err.Error(),
		)
		return
	}

	if !plan.Acls.IsNull() {
		var aclModels []backendAclResourceModel
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
			err := r.client.CreateAcl(ctx, "backend", plan.Name.ValueString(), aclPayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating acl",
					fmt.Sprintf("Could not create acl %s, unexpected error: %s", aclModel.AclName.ValueString(), err.Error()),
				)
				return
			}
		}
	}

	if !plan.HttpRequestRules.IsNull() {
		var httpRequestRuleModels []backendHttpRequestRuleResourceModel
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
			err := r.client.CreateHttpRequestRule(ctx, "backend", plan.Name.ValueString(), httpRequestRulePayload)
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
		var httpResponseRuleModels []backendHttpResponseRuleResourceModel
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
			err := r.client.CreateHttpResponseRule(ctx, "backend", plan.Name.ValueString(), httpResponseRulePayload)
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
		var tcpRequestRuleModels []backendTcpRequestRuleResourceModel
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
			err := r.client.CreateTcpRequestRule(ctx, "backend", plan.Name.ValueString(), tcpRequestRulePayload)
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
		var tcpResponseRuleModels []backendTcpResponseRuleResourceModel
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
				Type:      tcpResponseRuleModel.Type.ValueString(),
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
			err := r.client.CreateTcpResponseRule(ctx, "backend", plan.Name.ValueString(), tcpResponseRulePayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating tcp-response rule",
					fmt.Sprintf("Could not create tcp-response rule, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	if !plan.Httpchecks.IsNull() {
		var httpcheckModels []backendHttpcheckResourceModel
		diags := plan.Httpchecks.ElementsAs(ctx, &httpcheckModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(httpcheckModels, func(i, j int) bool {
			return httpcheckModels[i].GetIndex() < httpcheckModels[j].GetIndex()
		})

		for _, httpcheckModel := range httpcheckModels {
			httpcheckPayload := &HttpcheckPayload{
				Index:           httpcheckModel.Index.ValueInt64(),
				Addr:            httpcheckModel.Addr.ValueString(),
				Match:           httpcheckModel.Match.ValueString(),
				Pattern:         httpcheckModel.Pattern.ValueString(),
				Type:            httpcheckModel.Type.ValueString(),
				Method:          httpcheckModel.Method.ValueString(),
				Port:            httpcheckModel.Port.ValueInt64(),
				Uri:             httpcheckModel.Uri.ValueString(),
				Version:         httpcheckModel.Version.ValueString(),
				ExclamationMark: httpcheckModel.ExclamationMark.ValueString(),
				LogLevel:        httpcheckModel.LogLevel.ValueString(),
				SendProxy:       httpcheckModel.SendProxy.ValueString(),
				ViaSocks4:       httpcheckModel.ViaSocks4.ValueString(),
				CheckComment:    httpcheckModel.CheckComment.ValueString(),
			}
			err := r.client.CreateHttpcheck(ctx, "backend", plan.Name.ValueString(), httpcheckPayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating httpcheck",
					fmt.Sprintf("Could not create httpcheck, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	if !plan.TcpChecks.IsNull() {
		var tcpCheckModels []backendTcpCheckResourceModel
		diags := plan.TcpChecks.ElementsAs(ctx, &tcpCheckModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		sort.Slice(tcpCheckModels, func(i, j int) bool {
			return tcpCheckModels[i].GetIndex() < tcpCheckModels[j].GetIndex()
		})

		for _, tcpCheckModel := range tcpCheckModels {
			tcpCheckPayload := &TcpCheckPayload{
				Index:      tcpCheckModel.Index.ValueInt64(),
				Action:     tcpCheckModel.Action.ValueString(),
				Comment:    tcpCheckModel.Comment.ValueString(),
				Port:       tcpCheckModel.Port.ValueInt64(),
				Address:    tcpCheckModel.Address.ValueString(),
				Data:       tcpCheckModel.Data.ValueString(),
				MinRecv:    tcpCheckModel.MinRecv.ValueInt64(),
				OnSuccess:  tcpCheckModel.OnSuccess.ValueString(),
				OnError:    tcpCheckModel.OnError.ValueString(),
				StatusCode: tcpCheckModel.StatusCode.ValueString(),
				Timeout:    tcpCheckModel.Timeout.ValueInt64(),
				LogLevel:   tcpCheckModel.LogLevel.ValueString(),
			}
			err := r.client.CreateTcpCheck(ctx, "backend", plan.Name.ValueString(), tcpCheckPayload)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating tcp_check",
					fmt.Sprintf("Could not create tcp_check, unexpected error: %s", err.Error()),
				)
				return
			}
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Read resource information.
func (r *backendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backendResourceModel
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

	backend, err := r.client.ReadBackend(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backend",
			"Could not read backend, unexpected error: "+err.Error(),
		)
		return
	}

	if backend == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(backend.Name)
	state.Mode = types.StringValue(backend.Mode)
	state.HttpConnectionMode = types.StringValue(backend.HttpConnectionMode)
	state.AdvCheck = types.StringValue(backend.AdvCheck)
	state.ServerTimeout = types.Int64Value(backend.ServerTimeout)
	state.CheckTimeout = types.Int64Value(backend.CheckTimeout)
	state.ConnectTimeout = types.Int64Value(backend.ConnectTimeout)
	state.QueueTimeout = types.Int64Value(backend.QueueTimeout)
	state.TunnelTimeout = types.Int64Value(backend.TunnelTimeout)
	state.TarpitTimeout = types.Int64Value(backend.TarpitTimeout)
	state.CheckCache = types.StringValue(backend.CheckCache)
	state.Retries = types.Int64Value(backend.Retries)

	// Default Server Configuration (for SSL/TLS settings)
	if backend.DefaultServer != nil {
		state.DefaultServer, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"addr":            types.StringType,
			"port":            types.Int64Type,
			"ssl":             types.StringType,
			"ssl_cafile":      types.StringType,
			"ssl_certificate": types.StringType,
			"ssl_max_ver":     types.StringType,
			"ssl_min_ver":     types.StringType,
			"ssl_reuse":       types.StringType,
			"ciphers":         types.StringType,
			"ciphersuites":    types.StringType,
			"verify":          types.StringType,
		}, backend.DefaultServer)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if backend.Forwardfor != nil {
		var forwardforModel struct {
			Enabled types.String `tfsdk:"enabled"`
		}
		forwardforModel.Enabled = types.StringValue(backend.Forwardfor.Enabled)
		state.Forwardfor, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"enabled": types.StringType,
		}, forwardforModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if backend.Balance != nil {
		var balanceModel struct {
			Algorithm types.String `tfsdk:"algorithm"`
			UrlParam  types.String `tfsdk:"url_param"`
		}
		balanceModel.Algorithm = types.StringValue(backend.Balance.Algorithm)
		balanceModel.UrlParam = types.StringValue(backend.Balance.UrlParam)
		state.Balance, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"algorithm": types.StringType,
			"url_param": types.StringType,
		}, balanceModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if backend.HttpchkParams != nil {
		var httpchkParamsModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		httpchkParamsModel.Method = types.StringValue(backend.HttpchkParams.Method)
		httpchkParamsModel.Uri = types.StringValue(backend.HttpchkParams.Uri)
		httpchkParamsModel.Version = types.StringValue(backend.HttpchkParams.Version)
		state.HttpchkParams, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
			"method":  types.StringType,
			"uri":     types.StringType,
			"version": types.StringType,
		}, httpchkParamsModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	acls, err := r.client.ReadAcls(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading acls",
			"Could not read acls, unexpected error: "+err.Error(),
		)
		return
	}

	if len(acls) > 0 {
		var aclModels []backendAclResourceModel
		for _, acl := range acls {
			aclModels = append(aclModels, backendAclResourceModel{
				AclName:   types.StringValue(acl.AclName),
				Index:     types.Int64Value(acl.Index),
				Criterion: types.StringValue(acl.Criterion),
				Value:     types.StringValue(acl.Value),
			})
		}
		state.Acls, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: backendAclResourceModel{}.attrTypes(),
		}, aclModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	httpRequestRules, err := r.client.ReadHttpRequestRules(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading http-request rules",
			"Could not read http-request rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(httpRequestRules) > 0 {
		var httpRequestRuleModels []backendHttpRequestRuleResourceModel
		for _, httpRequestRule := range httpRequestRules {
			httpRequestRuleModels = append(httpRequestRuleModels, backendHttpRequestRuleResourceModel{
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
			AttrTypes: backendHttpRequestRuleResourceModel{}.attrTypes(),
		}, httpRequestRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	httpResponseRules, err := r.client.ReadHttpResponseRules(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading http-response rules",
			"Could not read http-response rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(httpResponseRules) > 0 {
		var httpResponseRuleModels []backendHttpResponseRuleResourceModel
		for _, httpResponseRule := range httpResponseRules {
			httpResponseRuleModels = append(httpResponseRuleModels, backendHttpResponseRuleResourceModel{
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
			AttrTypes: backendHttpResponseRuleResourceModel{}.attrTypes(),
		}, httpResponseRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tcpRequestRules, err := r.client.ReadTcpRequestRules(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tcp-request rules",
			"Could not read tcp-request rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(tcpRequestRules) > 0 {
		var tcpRequestRuleModels []backendTcpRequestRuleResourceModel
		for _, tcpRequestRule := range tcpRequestRules {
			tcpRequestRuleModels = append(tcpRequestRuleModels, backendTcpRequestRuleResourceModel{
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
			AttrTypes: backendTcpRequestRuleResourceModel{}.attrTypes(),
		}, tcpRequestRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tcpResponseRules, err := r.client.ReadTcpResponseRules(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tcp-response rules",
			"Could not read tcp-response rules, unexpected error: "+err.Error(),
		)
		return
	}

	if len(tcpResponseRules) > 0 {
		var tcpResponseRuleModels []backendTcpResponseRuleResourceModel
		for _, tcpResponseRule := range tcpResponseRules {
			tcpResponseRuleModels = append(tcpResponseRuleModels, backendTcpResponseRuleResourceModel{
				Index:     types.Int64Value(tcpResponseRule.Index),
				Type:      types.StringValue(tcpResponseRule.Type),
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
			AttrTypes: backendTcpResponseRuleResourceModel{}.attrTypes(),
		}, tcpResponseRuleModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	httpchecks, err := r.client.ReadHttpchecks(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading httpchecks",
			"Could not read httpchecks, unexpected error: "+err.Error(),
		)
		return
	}

	if len(httpchecks) > 0 {
		var httpcheckModels []backendHttpcheckResourceModel
		for _, httpcheck := range httpchecks {
			httpcheckModels = append(httpcheckModels, backendHttpcheckResourceModel{
				Index:           types.Int64Value(httpcheck.Index),
				Addr:            types.StringValue(httpcheck.Addr),
				Match:           types.StringValue(httpcheck.Match),
				Pattern:         types.StringValue(httpcheck.Pattern),
				Type:            types.StringValue(httpcheck.Type),
				Method:          types.StringValue(httpcheck.Method),
				Port:            types.Int64Value(httpcheck.Port),
				Uri:             types.StringValue(httpcheck.Uri),
				Version:         types.StringValue(httpcheck.Version),
				ExclamationMark: types.StringValue(httpcheck.ExclamationMark),
				LogLevel:        types.StringValue(httpcheck.LogLevel),
				SendProxy:       types.StringValue(httpcheck.SendProxy),
				ViaSocks4:       types.StringValue(httpcheck.ViaSocks4),
				CheckComment:    types.StringValue(httpcheck.CheckComment),
			})
		}
		state.Httpchecks, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: backendHttpcheckResourceModel{}.attrTypes(),
		}, httpcheckModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tcpChecks, err := r.client.ReadTcpChecks(ctx, "backend", state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tcp_checks",
			"Could not read tcp_checks, unexpected error: "+err.Error(),
		)
		return
	}

	if len(tcpChecks) > 0 {
		var tcpCheckModels []backendTcpCheckResourceModel
		for _, tcpCheck := range tcpChecks {
			tcpCheckModels = append(tcpCheckModels, backendTcpCheckResourceModel{
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
		state.TcpChecks, diags = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: backendTcpCheckResourceModel{}.attrTypes(),
		}, tcpCheckModels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update resource information.
func (r *backendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backendResourceModel
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

	payload := &BackendPayload{
		Name:               plan.Name.ValueString(),
		Mode:               plan.Mode.ValueString(),
		HttpConnectionMode: plan.HttpConnectionMode.ValueString(),
		AdvCheck:           plan.AdvCheck.ValueString(),
		ServerTimeout:      plan.ServerTimeout.ValueInt64(),
		CheckTimeout:       plan.CheckTimeout.ValueInt64(),
		ConnectTimeout:     plan.ConnectTimeout.ValueInt64(),
		QueueTimeout:       plan.QueueTimeout.ValueInt64(),
		TunnelTimeout:      plan.TunnelTimeout.ValueInt64(),
		TarpitTimeout:      plan.TarpitTimeout.ValueInt64(),
		CheckCache:         plan.CheckCache.ValueString(),
		Retries:            plan.Retries.ValueInt64(),

		// Default Server Configuration (for SSL/TLS settings)
		DefaultServer: &DefaultServerPayload{
			// SSL/TLS Configuration
			Ssl:            "enabled", // Default value
			SslCafile:      "",
			SslCertificate: "",
			SslMaxVer:      "",
			SslMinVer:      "",
			SslReuse:       "",
			Ciphers:        "",
			Ciphersuites:   "",
			Verify:         "",

			// SSL/TLS Protocol Control (v3 fields)
			Sslv3:  "disabled",
			Tlsv10: "disabled",
			Tlsv11: "disabled",
			Tlsv12: "enabled",
			Tlsv13: "enabled",

			// SSL/TLS Protocol Control (deprecated v2 fields)
			NoSslv3:        "enabled",
			NoTlsv10:       "enabled",
			NoTlsv11:       "enabled",
			NoTlsv12:       "disabled",
			NoTlsv13:       "disabled",
			ForceSslv3:     "disabled",
			ForceTlsv10:    "disabled",
			ForceTlsv11:    "disabled",
			ForceTlsv12:    "disabled",
			ForceTlsv13:    "disabled",
			ForceStrictSni: "",
		},
	}

	if !plan.Forwardfor.IsNull() {
		var forwardforModel struct {
			Enabled types.String `tfsdk:"enabled"`
		}
		diags := plan.Forwardfor.As(ctx, &forwardforModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload.Forwardfor = &ForwardFor{
			Enabled: forwardforModel.Enabled.ValueString(),
		}
	}

	if !plan.Balance.IsNull() {
		var balanceModel struct {
			Algorithm types.String `tfsdk:"algorithm"`
			UrlParam  types.String `tfsdk:"url_param"`
		}
		diags := plan.Balance.As(ctx, &balanceModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload.Balance = &Balance{
			Algorithm: balanceModel.Algorithm.ValueString(),
			UrlParam:  balanceModel.UrlParam.ValueString(),
		}
	}

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
		payload.HttpchkParams = &HttpchkParams{
			Method:  httpchkParamsModel.Method.ValueString(),
			Uri:     httpchkParamsModel.Uri.ValueString(),
			Version: httpchkParamsModel.Version.ValueString(),
		}
	}

	err := r.client.UpdateBackend(ctx, plan.Name.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating backend",
			"Could not update backend, unexpected error: "+err.Error(),
		)
		return
	}

	if !plan.Acls.IsNull() {
		var planAcls []backendAclResourceModel
		diags := plan.Acls.ElementsAs(ctx, &planAcls, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateAcls []backendAclResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
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

		planAclsMap := make(map[int64]backendAclResourceModel)
		for _, acl := range planAcls {
			planAclsMap[acl.Index.ValueInt64()] = acl
		}

		stateAclsMap := make(map[int64]backendAclResourceModel)
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
				err := r.client.CreateAcl(ctx, "backend", plan.Name.ValueString(), aclPayload)
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
				err := r.client.UpdateAcl(ctx, index, "backend", plan.Name.ValueString(), aclPayload)
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
				err := r.client.DeleteAcl(ctx, index, "backend", plan.Name.ValueString())
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
		var planHttpRequestRules []backendHttpRequestRuleResourceModel
		diags := plan.HttpRequestRules.ElementsAs(ctx, &planHttpRequestRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateHttpRequestRules []backendHttpRequestRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
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

		planHttpRequestRulesMap := make(map[int64]backendHttpRequestRuleResourceModel)
		for _, rule := range planHttpRequestRules {
			planHttpRequestRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateHttpRequestRulesMap := make(map[int64]backendHttpRequestRuleResourceModel)
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
				err := r.client.CreateHttpRequestRule(ctx, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.UpdateHttpRequestRule(ctx, index, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.DeleteHttpRequestRule(ctx, index, "backend", plan.Name.ValueString())
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
		var planHttpResponseRules []backendHttpResponseRuleResourceModel
		diags := plan.HttpResponseRules.ElementsAs(ctx, &planHttpResponseRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateHttpResponseRules []backendHttpResponseRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
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

		planHttpResponseRulesMap := make(map[int64]backendHttpResponseRuleResourceModel)
		for _, rule := range planHttpResponseRules {
			planHttpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateHttpResponseRulesMap := make(map[int64]backendHttpResponseRuleResourceModel)
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
				err := r.client.CreateHttpResponseRule(ctx, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.UpdateHttpResponseRule(ctx, index, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.DeleteHttpResponseRule(ctx, index, "backend", plan.Name.ValueString())
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
		var planTcpRequestRules []backendTcpRequestRuleResourceModel
		diags := plan.TcpRequestRules.ElementsAs(ctx, &planTcpRequestRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateTcpRequestRules []backendTcpRequestRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
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

		planTcpRequestRulesMap := make(map[int64]backendTcpRequestRuleResourceModel)
		for _, rule := range planTcpRequestRules {
			planTcpRequestRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateTcpRequestRulesMap := make(map[int64]backendTcpRequestRuleResourceModel)
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
				err := r.client.CreateTcpRequestRule(ctx, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.UpdateTcpRequestRule(ctx, index, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.DeleteTcpRequestRule(ctx, index, "backend", plan.Name.ValueString())
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
		var planTcpResponseRules []backendTcpResponseRuleResourceModel
		diags := plan.TcpResponseRules.ElementsAs(ctx, &planTcpResponseRules, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateTcpResponseRules []backendTcpResponseRuleResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
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

		planTcpResponseRulesMap := make(map[int64]backendTcpResponseRuleResourceModel)
		for _, rule := range planTcpResponseRules {
			planTcpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		stateTcpResponseRulesMap := make(map[int64]backendTcpResponseRuleResourceModel)
		for _, rule := range stateTcpResponseRules {
			stateTcpResponseRulesMap[rule.Index.ValueInt64()] = rule
		}

		for index, planRule := range planTcpResponseRulesMap {
			stateRule, ok := stateTcpResponseRulesMap[index]
			if !ok {
				// Create new tcp-response rule
				rulePayload := &TcpResponseRulePayload{
					Index:     planRule.Index.ValueInt64(),
					Type:      planRule.Type.ValueString(),
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
				err := r.client.CreateTcpResponseRule(ctx, "backend", plan.Name.ValueString(), rulePayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating tcp-response rule",
						fmt.Sprintf("Could not create tcp-response rule, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planRule.Type.Equal(stateRule.Type) || !planRule.Action.Equal(stateRule.Action) || !planRule.Cond.Equal(stateRule.Cond) || !planRule.CondTest.Equal(stateRule.CondTest) {
				// Update existing tcp-response rule
				rulePayload := &TcpResponseRulePayload{
					Index:     planRule.Index.ValueInt64(),
					Type:      planRule.Type.ValueString(),
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
				err := r.client.UpdateTcpResponseRule(ctx, index, "backend", plan.Name.ValueString(), rulePayload)
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
				err := r.client.DeleteTcpResponseRule(ctx, index, "backend", plan.Name.ValueString())
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

	if !plan.Httpchecks.IsNull() {
		var planHttpchecks []backendHttpcheckResourceModel
		diags := plan.Httpchecks.ElementsAs(ctx, &planHttpchecks, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateHttpchecks []backendHttpcheckResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.Httpchecks.IsNull() {
				diags := state.Httpchecks.ElementsAs(ctx, &stateHttpchecks, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planHttpchecksMap := make(map[int64]backendHttpcheckResourceModel)
		for _, httpcheck := range planHttpchecks {
			planHttpchecksMap[httpcheck.Index.ValueInt64()] = httpcheck
		}

		stateHttpchecksMap := make(map[int64]backendHttpcheckResourceModel)
		for _, httpcheck := range stateHttpchecks {
			stateHttpchecksMap[httpcheck.Index.ValueInt64()] = httpcheck
		}

		for index, planHttpcheck := range planHttpchecksMap {
			stateHttpcheck, ok := stateHttpchecksMap[index]
			if !ok {
				// Create new httpcheck
				httpcheckPayload := &HttpcheckPayload{
					Index:           planHttpcheck.Index.ValueInt64(),
					Addr:            planHttpcheck.Addr.ValueString(),
					Match:           planHttpcheck.Match.ValueString(),
					Pattern:         planHttpcheck.Pattern.ValueString(),
					Type:            planHttpcheck.Type.ValueString(),
					Method:          planHttpcheck.Method.ValueString(),
					Port:            planHttpcheck.Port.ValueInt64(),
					Uri:             planHttpcheck.Uri.ValueString(),
					Version:         planHttpcheck.Version.ValueString(),
					ExclamationMark: planHttpcheck.ExclamationMark.ValueString(),
					LogLevel:        planHttpcheck.LogLevel.ValueString(),
					SendProxy:       planHttpcheck.SendProxy.ValueString(),
					ViaSocks4:       planHttpcheck.ViaSocks4.ValueString(),
					CheckComment:    planHttpcheck.CheckComment.ValueString(),
				}
				err := r.client.CreateHttpcheck(ctx, "backend", plan.Name.ValueString(), httpcheckPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating httpcheck",
						fmt.Sprintf("Could not create httpcheck, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planHttpcheck.Type.Equal(stateHttpcheck.Type) {
				// Update existing httpcheck
				httpcheckPayload := &HttpcheckPayload{
					Index:           planHttpcheck.Index.ValueInt64(),
					Addr:            planHttpcheck.Addr.ValueString(),
					Match:           planHttpcheck.Match.ValueString(),
					Pattern:         planHttpcheck.Pattern.ValueString(),
					Type:            planHttpcheck.Type.ValueString(),
					Method:          planHttpcheck.Method.ValueString(),
					Port:            planHttpcheck.Port.ValueInt64(),
					Uri:             planHttpcheck.Uri.ValueString(),
					Version:         planHttpcheck.Version.ValueString(),
					ExclamationMark: planHttpcheck.ExclamationMark.ValueString(),
					LogLevel:        planHttpcheck.LogLevel.ValueString(),
					SendProxy:       planHttpcheck.SendProxy.ValueString(),
					ViaSocks4:       planHttpcheck.ViaSocks4.ValueString(),
					CheckComment:    planHttpcheck.CheckComment.ValueString(),
				}
				err := r.client.UpdateHttpcheck(ctx, index, "backend", plan.Name.ValueString(), httpcheckPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating httpcheck",
						fmt.Sprintf("Could not update httpcheck %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateHttpchecksMap {
			if _, ok := planHttpchecksMap[index]; !ok {
				// Delete httpcheck
				err := r.client.DeleteHttpcheck(ctx, index, "backend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting httpcheck",
						fmt.Sprintf("Could not delete httpcheck %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}
	}

	if !plan.TcpChecks.IsNull() {
		var planTcpChecks []backendTcpCheckResourceModel
		diags := plan.TcpChecks.ElementsAs(ctx, &planTcpChecks, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateTcpChecks []backendTcpCheckResourceModel
		if !req.State.Raw.IsNull() {
			var state backendResourceModel
			diags := req.State.Get(ctx, &state)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !state.TcpChecks.IsNull() {
				diags := state.TcpChecks.ElementsAs(ctx, &stateTcpChecks, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		planTcpChecksMap := make(map[int64]backendTcpCheckResourceModel)
		for _, tcpCheck := range planTcpChecks {
			planTcpChecksMap[tcpCheck.Index.ValueInt64()] = tcpCheck
		}

		stateTcpChecksMap := make(map[int64]backendTcpCheckResourceModel)
		for _, tcpCheck := range stateTcpChecks {
			stateTcpChecksMap[tcpCheck.Index.ValueInt64()] = tcpCheck
		}

		for index, planTcpCheck := range planTcpChecksMap {
			stateTcpCheck, ok := stateTcpChecksMap[index]
			if !ok {
				// Create new tcp_check
				tcpCheckPayload := &TcpCheckPayload{
					Index:      planTcpCheck.Index.ValueInt64(),
					Action:     planTcpCheck.Action.ValueString(),
					Comment:    planTcpCheck.Comment.ValueString(),
					Port:       planTcpCheck.Port.ValueInt64(),
					Address:    planTcpCheck.Address.ValueString(),
					Data:       planTcpCheck.Data.ValueString(),
					MinRecv:    planTcpCheck.MinRecv.ValueInt64(),
					OnSuccess:  planTcpCheck.OnSuccess.ValueString(),
					OnError:    planTcpCheck.OnError.ValueString(),
					StatusCode: planTcpCheck.StatusCode.ValueString(),
					Timeout:    planTcpCheck.Timeout.ValueInt64(),
					LogLevel:   planTcpCheck.LogLevel.ValueString(),
				}
				err := r.client.CreateTcpCheck(ctx, "backend", plan.Name.ValueString(), tcpCheckPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating tcp_check",
						fmt.Sprintf("Could not create tcp_check, unexpected error: %s", err.Error()),
					)
					return
				}
			} else if !planTcpCheck.Action.Equal(stateTcpCheck.Action) {
				// Update existing tcp_check
				tcpCheckPayload := &TcpCheckPayload{
					Index:      planTcpCheck.Index.ValueInt64(),
					Action:     planTcpCheck.Action.ValueString(),
					Comment:    planTcpCheck.Comment.ValueString(),
					Port:       planTcpCheck.Port.ValueInt64(),
					Address:    planTcpCheck.Address.ValueString(),
					Data:       planTcpCheck.Data.ValueString(),
					MinRecv:    planTcpCheck.MinRecv.ValueInt64(),
					OnSuccess:  planTcpCheck.OnSuccess.ValueString(),
					OnError:    planTcpCheck.OnError.ValueString(),
					StatusCode: planTcpCheck.StatusCode.ValueString(),
					Timeout:    planTcpCheck.Timeout.ValueInt64(),
					LogLevel:   planTcpCheck.LogLevel.ValueString(),
				}
				err := r.client.UpdateTcpCheck(ctx, index, "backend", plan.Name.ValueString(), tcpCheckPayload)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error updating tcp_check",
						fmt.Sprintf("Could not update tcp_check %d, unexpected error: %s", index, err.Error()),
					)
					return
				}
			}
		}

		for index := range stateTcpChecksMap {
			if _, ok := planTcpChecksMap[index]; !ok {
				// Delete tcp_check
				err := r.client.DeleteTcpCheck(ctx, index, "backend", plan.Name.ValueString())
				if err != nil {
					resp.Diagnostics.AddError(
						"Error deleting tcp_check",
						fmt.Sprintf("Could not delete tcp_check %d, unexpected error: %s", index, err.Error()),
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
func (r *backendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backendResourceModel
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

	err := r.client.DeleteBackend(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backend",
			"Could not delete backend, unexpected error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}
