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
	Forwardfor         types.List   `tfsdk:"forwardfor"`
	HttpConnectionMode types.String `tfsdk:"http_connection_mode"`
	Acls               types.List   `tfsdk:"acl"`
	HttpRequestRules   types.List   `tfsdk:"httprequestrule"`
	HttpResponseRules  types.List   `tfsdk:"httpresponserule"`
	TcpRequestRules    types.List   `tfsdk:"tcprequestrule"`
	TcpResponseRules   types.List   `tfsdk:"tcpresponserule"`
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

	// SSL/TLS Configuration Fields
	// Deprecated fields (API v2) - will be removed in future
	NoSslv3        types.Bool   `tfsdk:"no_sslv3"`
	NoTlsv10       types.Bool   `tfsdk:"no_tlsv10"`
	NoTlsv11       types.Bool   `tfsdk:"no_tlsv11"`
	NoTlsv12       types.Bool   `tfsdk:"no_tlsv12"`
	NoTlsv13       types.Bool   `tfsdk:"no_tlsv13"`
	ForceSslv3     types.Bool   `tfsdk:"force_sslv3"`
	ForceTlsv10    types.Bool   `tfsdk:"force_tlsv10"`
	ForceTlsv11    types.Bool   `tfsdk:"force_tlsv11"`
	ForceTlsv12    types.Bool   `tfsdk:"force_tlsv12"`
	ForceTlsv13    types.Bool   `tfsdk:"force_tlsv13"`
	ForceStrictSni types.String `tfsdk:"force_strict_sni"`

	// New v3 fields (non-deprecated)
	Sslv3  types.Bool `tfsdk:"sslv3"`
	Tlsv10 types.Bool `tfsdk:"tlsv10"`
	Tlsv11 types.Bool `tfsdk:"tlsv11"`
	Tlsv12 types.Bool `tfsdk:"tlsv12"`
	Tlsv13 types.Bool `tfsdk:"tlsv13"`

	// SSL/TLS Configuration
	Ssl            types.Bool   `tfsdk:"ssl"`
	SslCafile      types.String `tfsdk:"ssl_cafile"`
	SslCertificate types.String `tfsdk:"ssl_certificate"`
	SslMaxVer      types.String `tfsdk:"ssl_max_ver"`
	SslMinVer      types.String `tfsdk:"ssl_min_ver"`
	SslReuse       types.String `tfsdk:"ssl_reuse"`
	Ciphers        types.String `tfsdk:"ciphers"`
	Ciphersuites   types.String `tfsdk:"ciphersuites"`
	Verify         types.String `tfsdk:"verify"`

	// Block fields (defined in schema as blocks, not attributes)
	Balance       types.Object `tfsdk:"balance"`
	HttpchkParams types.Object `tfsdk:"httpchk_params"`
	Httpchecks    types.List   `tfsdk:"httpcheck"`
}

func (b backendResourceModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                 types.StringType,
		"mode":                 types.StringType,
		"forwardfor":           types.ListType{ElemType: types.ObjectType{}},
		"http_connection_mode": types.StringType,
		"acl":                  types.ListType{ElemType: types.ObjectType{}},
		"httprequestrule":      types.ListType{ElemType: types.ObjectType{}},
		"httpresponserule":     types.ListType{ElemType: types.ObjectType{}},
		"tcprequestrule":       types.ListType{ElemType: types.ObjectType{}},
		"tcpresponserule":      types.ListType{ElemType: types.ObjectType{}},
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

		// SSL/TLS Configuration Fields
		"no_sslv3":         types.BoolType,
		"no_tlsv10":        types.BoolType,
		"no_tlsv11":        types.BoolType,
		"no_tlsv12":        types.BoolType,
		"no_tlsv13":        types.BoolType,
		"force_sslv3":      types.BoolType,
		"force_tlsv10":     types.BoolType,
		"force_tlsv11":     types.BoolType,
		"force_tlsv12":     types.BoolType,
		"force_tlsv13":     types.BoolType,
		"force_strict_sni": types.StringType,

		// New v3 fields (non-deprecated)
		"sslv3":  types.BoolType,
		"tlsv10": types.BoolType,
		"tlsv11": types.BoolType,
		"tlsv12": types.BoolType,
		"tlsv13": types.BoolType,

		// SSL/TLS Configuration
		"ssl":             types.BoolType,
		"ssl_cafile":      types.StringType,
		"ssl_certificate": types.StringType,
		"ssl_max_ver":     types.StringType,
		"ssl_min_ver":     types.StringType,
		"ssl_reuse":       types.StringType,
		"ciphers":         types.StringType,
		"ciphersuites":    types.StringType,
		"verify":          types.StringType,

		// Block fields
		"balance":        types.ObjectType{},
		"httpchk_params": types.ObjectType{},
		"httpcheck":      types.ListType{ElemType: types.ObjectType{}},
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

func (h backendHttpRequestRuleResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h backendHttpRequestRuleResourceModel) attrTypes() map[string]attr.Type {
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

// backendHttpResponseRuleResourceModel maps the resource schema data.
type backendHttpResponseRuleResourceModel struct {
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

func (h backendHttpResponseRuleResourceModel) GetIndex() int64 {
	return h.Index.ValueInt64()
}

func (h backendHttpResponseRuleResourceModel) attrTypes() map[string]attr.Type {
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
			"acl": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
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
			"httprequestrule": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the http-request rule",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the http-request rule",
						},
						"acl_file": schema.StringAttribute{
							Optional:    true,
							Description: "The acl file of the http-request rule",
						},
						"acl_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The acl keyfmt of the http-request rule",
						},
						"bandwidth_limit_name": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit name of the http-request rule",
						},
						"bandwidth_limit_period": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit period of the http-request rule",
						},
						"bandwidth_limit_limit": schema.StringAttribute{
							Optional:    true,
							Description: "The bandwidth limit limit of the http-request rule",
						},
						"cache_name": schema.StringAttribute{
							Optional:    true,
							Description: "The cache name of the http-request rule",
						},
						"cond": schema.StringAttribute{
							Optional:    true,
							Description: "The condition of the http-request rule",
						},
						"cond_test": schema.StringAttribute{
							Optional:    true,
							Description: "The condition test of the http-request rule",
						},
						"expr": schema.StringAttribute{
							Optional:    true,
							Description: "The expr of the http-request rule",
						},
						"hdr_format": schema.StringAttribute{
							Optional:    true,
							Description: "The header format of the http-request rule",
						},
						"hdr_match": schema.StringAttribute{
							Optional:    true,
							Description: "The header match of the http-request rule",
						},
						"hdr_method": schema.StringAttribute{
							Optional:    true,
							Description: "The header method of the http-request rule",
						},
						"hdr_name": schema.StringAttribute{
							Optional:    true,
							Description: "The header name of the http-request rule",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log level of the http-request rule",
						},
						"lua_action": schema.StringAttribute{
							Optional:    true,
							Description: "The lua action of the http-request rule",
						},
						"lua_params": schema.StringAttribute{
							Optional:    true,
							Description: "The lua params of the http-request rule",
						},
						"map_file": schema.StringAttribute{
							Optional:    true,
							Description: "The map file of the http-request rule",
						},
						"map_keyfmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map keyfmt of the http-request rule",
						},
						"map_valuefmt": schema.StringAttribute{
							Optional:    true,
							Description: "The map valuefmt of the http-request rule",
						},
						"mark_value": schema.StringAttribute{
							Optional:    true,
							Description: "The mark value of the http-request rule",
						},
						"method_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The method fmt of the http-request rule",
						},
						"nice_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The nice value of the http-request rule",
						},
						"path_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The path fmt of the http-request rule",
						},
						"path_match": schema.StringAttribute{
							Optional:    true,
							Description: "The path match of the http-request rule",
						},
						"query_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The query fmt of the http-request rule",
						},
						"redir_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The redir code of the http-request rule",
						},
						"redir_type": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection type of the http-request rule",
						},
						"redir_value": schema.StringAttribute{
							Optional:    true,
							Description: "The redirection value of the http-request rule",
						},
						"sc_expr": schema.StringAttribute{
							Optional:    true,
							Description: "The sc expr of the http-request rule",
						},
						"sc_id": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc id of the http-request rule",
						},
						"sc_idx": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc idx of the http-request rule",
						},
						"sc_int": schema.Int64Attribute{
							Optional:    true,
							Description: "The sc int of the http-request rule",
						},
						"service": schema.StringAttribute{
							Optional:    true,
							Description: "The service of the http-request rule",
						},
						"spoe_engine": schema.StringAttribute{
							Optional:    true,
							Description: "The spoe engine of the http-request rule",
						},
						"spoe_group": schema.StringAttribute{
							Optional:    true,
							Description: "The spoe group of the http-request rule",
						},
						"status_code": schema.Int64Attribute{
							Optional:    true,
							Description: "The status code of the http-request rule",
						},
						"status_reason": schema.StringAttribute{
							Optional:    true,
							Description: "The status reason of the http-request rule",
						},
						"timeout": schema.StringAttribute{
							Optional:    true,
							Description: "The timeout of the http-request rule",
						},
						"timeout_value": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout value of the http-request rule",
						},
						"tos_value": schema.StringAttribute{
							Optional:    true,
							Description: "The tos value of the http-request rule",
						},
						"track_sc_key": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc key of the http-request rule",
						},
						"track_sc_table": schema.StringAttribute{
							Optional:    true,
							Description: "The track sc table of the http-request rule",
						},
						"uri_fmt": schema.StringAttribute{
							Optional:    true,
							Description: "The uri fmt of the http-request rule",
						},
						"uri_match": schema.StringAttribute{
							Optional:    true,
							Description: "The uri match of the http-request rule",
						},
						"var_name": schema.StringAttribute{
							Optional:    true,
							Description: "The var name of the http-request rule",
						},
						"var_scope": schema.StringAttribute{
							Optional:    true,
							Description: "The var scope of the http-request rule",
						},
						"wait_time": schema.Int64Attribute{
							Optional:    true,
							Description: "The wait time of the http-request rule",
						},
					},
				},
			},
			"httpresponserule": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
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

			"tcp_check": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"index": schema.Int64Attribute{
							Required:    true,
							Description: "The index of the tcp_check",
						},
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The action of the tcp_check",
						},
						"comment": schema.StringAttribute{
							Optional:    true,
							Description: "The comment of the tcp_check",
						},
						"port": schema.Int64Attribute{
							Optional:    true,
							Description: "The port of the tcp_check",
						},
						"address": schema.StringAttribute{
							Optional:    true,
							Description: "The address of the tcp_check",
						},
						"data": schema.StringAttribute{
							Optional:    true,
							Description: "The data of the tcp_check",
						},
						"min_recv": schema.Int64Attribute{
							Optional:    true,
							Description: "The min_recv of the tcp_check",
						},
						"on_success": schema.StringAttribute{
							Optional:    true,
							Description: "The on_success of the tcp_check",
						},
						"on_error": schema.StringAttribute{
							Optional:    true,
							Description: "The on_error of the tcp_check",
						},
						"status_code": schema.StringAttribute{
							Optional:    true,
							Description: "The status_code of the tcp_check",
						},
						"timeout": schema.Int64Attribute{
							Optional:    true,
							Description: "The timeout of the tcp_check",
						},
						"log_level": schema.StringAttribute{
							Optional:    true,
							Description: "The log_level of the tcp_check",
						},
					},
				},
			},

			// SSL/TLS Configuration Fields
			"no_sslv3": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable SSLv3 protocol support. DEPRECATED: Use 'sslv3' field instead in Data Plane API v3",
			},
			"no_tlsv10": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.0 protocol support. DEPRECATED: Use 'tlsv10' field instead in Data Plane API v3",
			},
			"no_tlsv11": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.1 protocol support. DEPRECATED: Use 'tlsv11' field instead in Data Plane API v3",
			},
			"no_tlsv12": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.2 protocol support. DEPRECATED: Use 'tlsv12' field instead in Data Plane API v3",
			},
			"no_tlsv13": schema.BoolAttribute{
				Optional:    true,
				Description: "Disable TLSv1.3 protocol support. DEPRECATED: Use 'tlsv13' field instead in Data Plane API v3",
			},
			"force_sslv3": schema.BoolAttribute{
				Optional:    true,
				Description: "Force SSLv3 protocol support. DEPRECATED: Use 'sslv3' field instead in Data Plane API v3",
			},
			"force_tlsv10": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.0 protocol support. DEPRECATED: Use 'tlsv10' field instead in Data Plane API v3",
			},
			"force_tlsv11": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.1 protocol support. DEPRECATED: Use 'tlsv11' field instead in Data Plane API v3",
			},
			"force_tlsv12": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.2 protocol support. DEPRECATED: Use 'tlsv12' field instead in Data Plane API v3",
			},
			"force_tlsv13": schema.BoolAttribute{
				Optional:    true,
				Description: "Force TLSv1.3 protocol support. DEPRECATED: Use 'tlsv13' field instead in Data Plane API v3",
			},
			"force_strict_sni": schema.StringAttribute{
				Optional:    true,
				Description: "Force strict SNI. DEPRECATED: Use 'strict_sni' field instead in Data Plane API v3. Allowed: enabled|disabled",
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

			// SSL/TLS Configuration
			"ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable SSL for backend connections",
			},
			"ssl_cafile": schema.StringAttribute{
				Optional:    true,
				Description: "SSL CA file path. Pattern: ^[^\\s]+$",
			},
			"ssl_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "SSL certificate path. Pattern: ^[^\\s]+$",
			},
			"ssl_max_ver": schema.StringAttribute{
				Optional:    true,
				Description: "SSL maximum version. Allowed: SSLv3|TLSv1.0|TLSv1.1|TLSv1.2|TLSv1.3",
			},
			"ssl_min_ver": schema.StringAttribute{
				Optional:    true,
				Description: "SSL minimum version. Allowed: SSLv3|TLSv1.0|TLSv1.1|TLSv1.2|TLSv1.3",
			},
			"ssl_reuse": schema.StringAttribute{
				Optional:    true,
				Description: "SSL session reuse. Allowed: enabled|disabled",
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
			"forwardfor": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.StringAttribute{
							Required:    true,
							Description: "The state of the forwardfor. Allowed: enabled|disabled",
						},
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
						"match": schema.StringAttribute{
							Optional:    true,
							Description: "The match of the httpcheck",
						},
						"pattern": schema.StringAttribute{
							Optional:    true,
							Description: "The pattern of the httpcheck",
						},
						"type": schema.StringAttribute{
							Optional:    true,
							Description: "The type of the httpcheck",
						},
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

		// SSL/TLS Configuration Fields
		// Deprecated fields (API v2)
		NoSslv3:        plan.NoSslv3.ValueBool(),
		NoTlsv10:       plan.NoTlsv10.ValueBool(),
		NoTlsv11:       plan.NoTlsv11.ValueBool(),
		NoTlsv12:       plan.NoTlsv12.ValueBool(),
		NoTlsv13:       plan.NoTlsv13.ValueBool(),
		ForceSslv3:     plan.ForceSslv3.ValueBool(),
		ForceTlsv10:    plan.ForceTlsv10.ValueBool(),
		ForceTlsv11:    plan.ForceTlsv11.ValueBool(),
		ForceTlsv12:    plan.ForceTlsv12.ValueBool(),
		ForceTlsv13:    plan.ForceTlsv13.ValueBool(),
		ForceStrictSni: plan.ForceStrictSni.ValueString(),

		// New v3 fields (non-deprecated)
		Sslv3:  plan.Sslv3.ValueBool(),
		Tlsv10: plan.Tlsv10.ValueBool(),
		Tlsv11: plan.Tlsv11.ValueBool(),
		Tlsv12: plan.Tlsv12.ValueBool(),
		Tlsv13: plan.Tlsv13.ValueBool(),

		// SSL/TLS Configuration
		Ssl:            plan.Ssl.ValueBool(),
		SslCafile:      plan.SslCafile.ValueString(),
		SslCertificate: plan.SslCertificate.ValueString(),
		SslMaxVer:      plan.SslMaxVer.ValueString(),
		SslMinVer:      plan.SslMinVer.ValueString(),
		SslReuse:       plan.SslReuse.ValueString(),
		Ciphers:        plan.Ciphers.ValueString(),
		Ciphersuites:   plan.Ciphersuites.ValueString(),
		Verify:         plan.Verify.ValueString(),
	}

	if !plan.Forwardfor.IsNull() && len(plan.Forwardfor.Elements()) > 0 {
		var forwardforModels []struct {
			Enabled types.String `tfsdk:"enabled"`
		}
		diags := plan.Forwardfor.ElementsAs(ctx, &forwardforModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(forwardforModels) > 0 {
			payload.Forwardfor = &ForwardFor{
				Enabled: forwardforModels[0].Enabled.ValueString(),
			}
		}
	}

	// Handle balance block
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
		}
		if !balanceModel.UrlParam.IsNull() {
			payload.Balance.UrlParam = balanceModel.UrlParam.ValueString()
		}
	}

	// Handle httpchk_params block
	if !plan.HttpchkParams.IsNull() {
		var httpchkModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		diags := plan.HttpchkParams.As(ctx, &httpchkModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		payload.HttpchkParams = &HttpchkParams{
			Method: httpchkModel.Method.ValueString(),
			Uri:    httpchkModel.Uri.ValueString(),
		}
		if !httpchkModel.Version.IsNull() {
			payload.HttpchkParams.Version = httpchkModel.Version.ValueString()
		}
	}

	// Use the new transaction handling to avoid timeout issues
	transactionID, err := r.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating backend",
			"Could not begin transaction: "+err.Error(),
		)
		return
	}

	err = r.client.CreateBackendInTransaction(ctx, transactionID, payload)
	if err != nil {
		// Try to rollback the transaction
		_ = r.client.CommitTransaction(transactionID) // This will fail but that's okay
		resp.Diagnostics.AddError(
			"Error creating backend",
			"Could not create backend, unexpected error: "+err.Error(),
		)
		return
	}

	// Commit the transaction
	err = r.client.CommitTransaction(transactionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating backend",
			"Could not commit transaction: "+err.Error(),
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

	// TODO: Handle httpcheck from blocks instead of attributes
	// This is now defined as a block in the schema

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
	// Only set string fields if they have meaningful values (not empty)
	if backend.HttpConnectionMode != "" {
		state.HttpConnectionMode = types.StringValue(backend.HttpConnectionMode)
	} else {
		state.HttpConnectionMode = types.StringNull()
	}
	if backend.AdvCheck != "" {
		state.AdvCheck = types.StringValue(backend.AdvCheck)
	} else {
		state.AdvCheck = types.StringNull()
	}
	// Only set timeout fields if they have meaningful values (not zero/empty)
	if backend.ServerTimeout > 0 {
		state.ServerTimeout = types.Int64Value(backend.ServerTimeout)
	} else {
		state.ServerTimeout = types.Int64Null()
	}
	if backend.CheckTimeout > 0 {
		state.CheckTimeout = types.Int64Value(backend.CheckTimeout)
	} else {
		state.CheckTimeout = types.Int64Null()
	}
	if backend.ConnectTimeout > 0 {
		state.ConnectTimeout = types.Int64Value(backend.ConnectTimeout)
	} else {
		state.ConnectTimeout = types.Int64Null()
	}
	if backend.QueueTimeout > 0 {
		state.QueueTimeout = types.Int64Value(backend.QueueTimeout)
	} else {
		state.QueueTimeout = types.Int64Null()
	}
	if backend.TunnelTimeout > 0 {
		state.TunnelTimeout = types.Int64Value(backend.TunnelTimeout)
	} else {
		state.TunnelTimeout = types.Int64Null()
	}
	if backend.TarpitTimeout > 0 {
		state.TarpitTimeout = types.Int64Value(backend.TarpitTimeout)
	} else {
		state.TarpitTimeout = types.Int64Null()
	}
	if backend.CheckCache != "" {
		state.CheckCache = types.StringValue(backend.CheckCache)
	} else {
		state.CheckCache = types.StringNull()
	}
	if backend.Retries > 0 {
		state.Retries = types.Int64Value(backend.Retries)
	} else {
		state.Retries = types.Int64Null()
	}

	// SSL/TLS Configuration Fields
	// Deprecated fields (API v2) - only set if true
	if backend.NoSslv3 {
		state.NoSslv3 = types.BoolValue(true)
	} else {
		state.NoSslv3 = types.BoolNull()
	}
	if backend.NoTlsv10 {
		state.NoTlsv10 = types.BoolValue(true)
	} else {
		state.NoTlsv10 = types.BoolNull()
	}
	if backend.NoTlsv11 {
		state.NoTlsv11 = types.BoolValue(true)
	} else {
		state.NoTlsv11 = types.BoolNull()
	}
	if backend.NoTlsv12 {
		state.NoTlsv12 = types.BoolValue(true)
	} else {
		state.NoTlsv12 = types.BoolNull()
	}
	if backend.NoTlsv13 {
		state.NoTlsv13 = types.BoolValue(true)
	} else {
		state.NoTlsv13 = types.BoolNull()
	}
	if backend.ForceSslv3 {
		state.ForceSslv3 = types.BoolValue(true)
	} else {
		state.ForceSslv3 = types.BoolNull()
	}
	if backend.ForceTlsv10 {
		state.ForceTlsv10 = types.BoolValue(true)
	} else {
		state.ForceTlsv10 = types.BoolNull()
	}
	if backend.ForceTlsv11 {
		state.ForceTlsv11 = types.BoolValue(true)
	} else {
		state.ForceTlsv11 = types.BoolNull()
	}
	if backend.ForceTlsv12 {
		state.ForceTlsv12 = types.BoolValue(true)
	} else {
		state.ForceTlsv12 = types.BoolNull()
	}
	if backend.ForceTlsv13 {
		state.ForceTlsv13 = types.BoolValue(true)
	} else {
		state.ForceTlsv13 = types.BoolNull()
	}
	if backend.ForceStrictSni != "" {
		state.ForceStrictSni = types.StringValue(backend.ForceStrictSni)
	} else {
		state.ForceStrictSni = types.StringNull()
	}

	// New v3 fields (non-deprecated) - only set if true
	if backend.Sslv3 {
		state.Sslv3 = types.BoolValue(true)
	} else {
		state.Sslv3 = types.BoolNull()
	}
	if backend.Tlsv10 {
		state.Tlsv10 = types.BoolValue(true)
	} else {
		state.Tlsv10 = types.BoolNull()
	}
	if backend.Tlsv11 {
		state.Tlsv11 = types.BoolValue(true)
	} else {
		state.Tlsv11 = types.BoolNull()
	}
	if backend.Tlsv12 {
		state.Tlsv12 = types.BoolValue(true)
	} else {
		state.Tlsv12 = types.BoolNull()
	}
	if backend.Tlsv13 {
		state.Tlsv13 = types.BoolValue(true)
	} else {
		state.Tlsv13 = types.BoolNull()
	}

	// SSL/TLS Configuration - only set if true
	if backend.Ssl {
		state.Ssl = types.BoolValue(true)
	} else {
		state.Ssl = types.BoolNull()
	}
	// Only set SSL string fields if they have meaningful values (not empty)
	if backend.SslCafile != "" {
		state.SslCafile = types.StringValue(backend.SslCafile)
	} else {
		state.SslCafile = types.StringNull()
	}
	if backend.SslCertificate != "" {
		state.SslCertificate = types.StringValue(backend.SslCertificate)
	} else {
		state.SslCertificate = types.StringNull()
	}
	if backend.SslMaxVer != "" {
		state.SslMaxVer = types.StringValue(backend.SslMaxVer)
	} else {
		state.SslMaxVer = types.StringNull()
	}
	if backend.SslMinVer != "" {
		state.SslMinVer = types.StringValue(backend.SslMinVer)
	} else {
		state.SslMinVer = types.StringNull()
	}
	if backend.SslReuse != "" {
		state.SslReuse = types.StringValue(backend.SslReuse)
	} else {
		state.SslReuse = types.StringNull()
	}
	if backend.Ciphers != "" {
		state.Ciphers = types.StringValue(backend.Ciphers)
	} else {
		state.Ciphers = types.StringNull()
	}
	if backend.Ciphersuites != "" {
		state.Ciphersuites = types.StringValue(backend.Ciphersuites)
	} else {
		state.Ciphersuites = types.StringNull()
	}
	if backend.Verify != "" {
		state.Verify = types.StringValue(backend.Verify)
	} else {
		state.Verify = types.StringNull()
	}

	if backend.Forwardfor != nil {
		var forwardforModel struct {
			Enabled types.String `tfsdk:"enabled"`
		}
		forwardforModel.Enabled = types.StringValue(backend.Forwardfor.Enabled)
		forwardforList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"enabled": types.StringType,
			},
		}, []interface{}{forwardforModel})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Forwardfor = forwardforList
	}

	// Handle balance block
	if backend.Balance != nil {
		var balanceModel struct {
			Algorithm types.String `tfsdk:"algorithm"`
			UrlParam  types.String `tfsdk:"url_param"`
		}
		balanceModel.Algorithm = types.StringValue(backend.Balance.Algorithm)
		if backend.Balance.UrlParam != "" {
			balanceModel.UrlParam = types.StringValue(backend.Balance.UrlParam)
		}

		balanceObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"algorithm": types.StringType,
			"url_param": types.StringType,
		}, balanceModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Balance = balanceObj
	}

	// Handle httpchk_params block
	if backend.HttpchkParams != nil {
		var httpchkModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		httpchkModel.Method = types.StringValue(backend.HttpchkParams.Method)
		httpchkModel.Uri = types.StringValue(backend.HttpchkParams.Uri)
		if backend.HttpchkParams.Version != "" {
			httpchkModel.Version = types.StringValue(backend.HttpchkParams.Version)
		}

		httpchkObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"method":  types.StringType,
			"uri":     types.StringType,
			"version": types.StringType,
		}, httpchkModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.HttpchkParams = httpchkObj
	}

	// TODO: Handle balance and httpchk_params from blocks instead of attributes
	// These are now defined as blocks in the schema

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

	// TODO: Handle httpcheck from blocks instead of attributes
	// This is now defined as a block in the schema

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

		// SSL/TLS Configuration Fields
		// Deprecated fields (API v2)
		NoSslv3:        plan.NoSslv3.ValueBool(),
		NoTlsv10:       plan.NoTlsv10.ValueBool(),
		NoTlsv11:       plan.NoTlsv11.ValueBool(),
		NoTlsv12:       plan.NoTlsv12.ValueBool(),
		NoTlsv13:       plan.NoTlsv13.ValueBool(),
		ForceSslv3:     plan.ForceSslv3.ValueBool(),
		ForceTlsv10:    plan.ForceTlsv10.ValueBool(),
		ForceTlsv11:    plan.ForceTlsv11.ValueBool(),
		ForceTlsv12:    plan.ForceTlsv12.ValueBool(),
		ForceTlsv13:    plan.ForceTlsv13.ValueBool(),
		ForceStrictSni: plan.ForceStrictSni.ValueString(),

		// New v3 fields (non-deprecated)
		Sslv3:  plan.Sslv3.ValueBool(),
		Tlsv10: plan.Tlsv10.ValueBool(),
		Tlsv11: plan.Tlsv11.ValueBool(),
		Tlsv12: plan.Tlsv11.ValueBool(),
		Tlsv13: plan.Tlsv13.ValueBool(),

		// SSL/TLS Configuration
		Ssl:          plan.Ssl.ValueBool(),
		SslCafile:    plan.SslCafile.ValueString(),
		SslMinVer:    plan.SslMinVer.ValueString(),
		SslMaxVer:    plan.SslMaxVer.ValueString(),
		SslReuse:     plan.SslReuse.ValueString(),
		Ciphers:      plan.Ciphers.ValueString(),
		Ciphersuites: plan.Ciphersuites.ValueString(),
		Verify:       plan.Verify.ValueString(),
	}

	if !plan.Forwardfor.IsNull() && len(plan.Forwardfor.Elements()) > 0 {
		var forwardforModels []struct {
			Enabled types.String `tfsdk:"enabled"`
		}
		diags := plan.Forwardfor.ElementsAs(ctx, &forwardforModels, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(forwardforModels) > 0 {
			payload.Forwardfor = &ForwardFor{
				Enabled: forwardforModels[0].Enabled.ValueString(),
			}
		}
	}

	// Handle balance block
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
		}
		if !balanceModel.UrlParam.IsNull() {
			payload.Balance.UrlParam = balanceModel.UrlParam.ValueString()
		}
	}

	// Handle httpchk_params block
	if !plan.HttpchkParams.IsNull() {
		var httpchkModel struct {
			Method  types.String `tfsdk:"method"`
			Uri     types.String `tfsdk:"uri"`
			Version types.String `tfsdk:"version"`
		}
		diags := plan.HttpchkParams.As(ctx, &httpchkModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		payload.HttpchkParams = &HttpchkParams{
			Method: httpchkModel.Method.ValueString(),
			Uri:    httpchkModel.Uri.ValueString(),
		}
		if !httpchkModel.Version.IsNull() {
			payload.HttpchkParams.Version = httpchkModel.Version.ValueString()
		}
	}

	// TODO: Handle balance and httpchk_params from blocks instead of attributes
	// These are now defined as blocks in the schema

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

	// TODO: Handle httpcheck from blocks instead of attributes
	// This is now defined as a block in the schema

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
