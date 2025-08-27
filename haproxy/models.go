package haproxy

// FrontendPayload is the payload for the frontend resource.
type FrontendPayload struct {
	Name                     string              `json:"name"`
	DefaultBackend           string              `json:"default_backend"`
	HttpConnectionMode       string              `json:"http-connection-mode,omitempty"`
	AcceptInvalidHttpRequest string              `json:"accept_invalid_http_request,omitempty"`
	MaxConn                  int64               `json:"maxconn,omitempty"`
	Mode                     string              `json:"mode,omitempty"`
	Backlog                  int64               `json:"backlog,omitempty"`
	HttpKeepAliveTimeout     int64               `json:"http-keep-alive-timeout,omitempty"`
	HttpRequestTimeout       int64               `json:"http-request-timeout,omitempty"`
	HttpUseProxyHeader       string              `json:"http-use-proxy-header,omitempty"`
	HttpLog                  bool                `json:"httplog,omitempty"`
	HttpsLog                 string              `json:"httpslog,omitempty"`
	ErrorLogFormat           string              `json:"error_log_format,omitempty"`
	LogFormat                string              `json:"log_format,omitempty"`
	LogFormatSd              string              `json:"log_format_sd,omitempty"`
	MonitorUri               string              `json:"monitor_uri,omitempty"`
	TcpLog                   bool                `json:"tcplog,omitempty"`
	From                     string              `json:"from,omitempty"`
	ClientTimeout            int64               `json:"client_timeout,omitempty"`
	HttpUseHtx               string              `json:"http_use_htx,omitempty"`
	HttpIgnoreProbes         string              `json:"http_ignore_probes,omitempty"`
	LogTag                   string              `json:"log_tag,omitempty"`
	Clflog                   bool                `json:"clflog,omitempty"`
	Contstats                string              `json:"contstats,omitempty"`
	Dontlognull              string              `json:"dontlognull,omitempty"`
	LogSeparateErrors        string              `json:"log_separate_errors,omitempty"`
	OptionHttpServerClose    string              `json:"option_http_server_close,omitempty"`
	OptionHttpclose          string              `json:"option_httpclose,omitempty"`
	OptionHttpKeepAlive      string              `json:"option_http_keep_alive,omitempty"`
	OptionDontlogNormal      string              `json:"option_dontlog_normal,omitempty"`
	OptionLogasap            string              `json:"option_logasap,omitempty"`
	OptionTcplog             string              `json:"option_tcplog,omitempty"`
	OptionSocketStats        string              `json:"option_socket_stats,omitempty"`
	OptionForwardfor         string              `json:"option_forwardfor,omitempty"`
	TimeoutClient            int64               `json:"timeout_client,omitempty"`
	TimeoutHttpKeepAlive     int64               `json:"timeout_http_keep_alive,omitempty"`
	TimeoutHttpRequest       int64               `json:"timeout_http_request,omitempty"`
	TimeoutCont              int64               `json:"timeout_cont,omitempty"`
	TimeoutTarpit            int64               `json:"timeout_tarpit,omitempty"`
	StatsOptions             StatsOptionsPayload `json:"stats_options,omitempty"`
	MonitorFail              MonitorFailPayload  `json:"monitor_fail,omitempty"`
}

// StatsOptionsPayload is the payload for the stats_options resource.
type StatsOptionsPayload struct {
	StatsEnable      bool   `json:"stats_enable,omitempty"`
	StatsHideVersion bool   `json:"stats_hide_version,omitempty"`
	StatsShowLegends bool   `json:"stats_show_legends,omitempty"`
	StatsShowNode    bool   `json:"stats_show_node,omitempty"`
	StatsUri         string `json:"stats_uri,omitempty"`
	StatsRealm       string `json:"stats_realm,omitempty"`
	StatsAuth        string `json:"stats_auth,omitempty"`
	StatsRefresh     string `json:"stats_refresh,omitempty"`
}

// MonitorFailPayload is the payload for the monitor_fail resource.
type MonitorFailPayload struct {
	Cond     string `json:"cond"`
	CondTest string `json:"cond_test"`
}

// TransactionResponse is the response from the HAProxy Data Plane API when creating a transaction.
type TransactionResponse struct {
	Version int    `json:"_version"`
	ID      string `json:"id"`
	Status  string `json:"status"`
}

// GlobalPayload is the payload for the global resource.
type GlobalPayload struct {
	Name                    string `json:"name"`
	Maxconn                 int64  `json:"maxconn,omitempty"`
	Daemon                  string `json:"daemon,omitempty"`
	StatsTimeout            int64  `json:"stats_timeout,omitempty"`
	TuneSslDefaultDhParam   int64  `json:"tune_ssl_default_dh_param,omitempty"`
	SslDefaultBindCiphers   string `json:"ssl_default_bind_ciphers,omitempty"`
	SslDefaultBindOptions   string `json:"ssl_default_bind_options,omitempty"`
	SslDefaultServerCiphers string `json:"ssl_default_server_ciphers,omitempty"`
	SslDefaultServerOptions string `json:"ssl_default_server_options,omitempty"`
}

// BackendPayload is the payload for the backend resource.
type BackendPayload struct {
	Name               string        `json:"name"`
	Mode               string        `json:"mode"`
	AdvCheck           string        `json:"adv_check"`
	HttpConnectionMode string        `json:"http_connection_mode"`
	ServerTimeout      int64         `json:"server_timeout"`
	CheckTimeout       int64         `json:"check_timeout"`
	ConnectTimeout     int64         `json:"connect_timeout"`
	QueueTimeout       int64         `json:"queue_timeout"`
	TunnelTimeout      int64         `json:"tunnel_timeout"`
	TarpitTimeout      int64         `json:"tarpit_timeout"`
	CheckCache         string        `json:"checkcache"`
	Retries            int64         `json:"retries"`
	Balance            Balance       `json:"balance"`
	HttpchkParams      HttpchkParams `json:"httpchk_params"`
	Forwardfor         ForwardFor    `json:"forwardfor"`
	
	// SSL/TLS Configuration Fields
	// Deprecated fields (API v2) - will be removed in future
	NoSslv3            bool     `json:"no_sslv3,omitempty"`
	NoTlsv10           bool     `json:"no_tlsv10,omitempty"`
	NoTlsv11           bool     `json:"no_tlsv11,omitempty"`
	NoTlsv12           bool     `json:"no_tlsv12,omitempty"`
	NoTlsv13           bool     `json:"no_tlsv13,omitempty"`
	ForceSslv3         bool     `json:"force_sslv3,omitempty"`
	ForceTlsv10        bool     `json:"force_tlsv10,omitempty"`
	ForceTlsv11        bool     `json:"force_tlsv11,omitempty"`
	ForceTlsv12        bool     `json:"force_tlsv12,omitempty"`
	ForceTlsv13        bool     `json:"force_tlsv13,omitempty"`
	ForceStrictSni     string   `json:"force_strict_sni,omitempty"`
	
	// New v3 fields (non-deprecated)
	Sslv3              bool     `json:"sslv3,omitempty"`
	Tlsv10             bool     `json:"tlsv10,omitempty"`
	Tlsv11             bool     `json:"tlsv11,omitempty"`
	Tlsv12             bool     `json:"tlsv12,omitempty"`
	Tlsv13             bool     `json:"tlsv13,omitempty"`
	
	// SSL/TLS Configuration
	Ssl                bool     `json:"ssl,omitempty"`
	SslCafile          string   `json:"ssl_cafile,omitempty"`
	SslCertificate     string   `json:"ssl_certificate,omitempty"`
	SslMaxVer          string   `json:"ssl_max_ver,omitempty"`
	SslMinVer          string   `json:"ssl_min_ver,omitempty"`
	SslReuse           string   `json:"ssl_reuse,omitempty"`
	Ciphers            string   `json:"ciphers,omitempty"`
	Ciphersuites       string   `json:"ciphersuites,omitempty"`
	Verify             string   `json:"verify,omitempty"`
}

type Balance struct {
	Algorithm string `json:"algorithm"`
	UrlParam  string `json:"url_param"`
}

type HttpchkParams struct {
	Method  string `json:"method"`
	Uri     string `json:"uri"`
	Version string `json:"version"`
}

type ForwardFor struct {
	Enabled string `json:"enabled"`
}

// ServerPayload is the payload for the server resource.
type ServerPayload struct {
	Name             string   `json:"name"`
	Address          string   `json:"address"`
	Port             int64    `json:"port"`
	AgentAddr        string   `json:"agent-addr,omitempty"`
	AgentCheck       string   `json:"agent-check,omitempty"`
	AgentInter       int64    `json:"agent-inter,omitempty"`
	AgentPort        int64    `json:"agent-port,omitempty"`
	AgentSend        string   `json:"agent-send,omitempty"`
	Allow0rtt        bool     `json:"allow_0rtt,omitempty"`
	Alpn             string   `json:"alpn,omitempty"`
	Backup           string   `json:"backup,omitempty"`
	Check            string   `json:"check,omitempty"`
	CheckAlpn        string   `json:"check-alpn,omitempty"`
	CheckSni         string   `json:"check-sni,omitempty"`
	CheckSsl         string   `json:"check-ssl,omitempty"`
	CheckViaSocks4   string   `json:"check-via-socks4,omitempty"`
	Ciphers          string   `json:"ciphers,omitempty"`
	Ciphersuites     string   `json:"ciphersuites,omitempty"`
	Cookie           string   `json:"cookie,omitempty"`
	Crt              string   `json:"crt,omitempty"`
	Downinter        int64    `json:"downinter,omitempty"`
	ErrorLimit       int64    `json:"error-limit,omitempty"`
	Fall             int64    `json:"fall,omitempty"`
	Fastinter        int64    `json:"fastinter,omitempty"`
	ForceSslv3       string   `json:"force_sslv3,omitempty"`
	ForceTlsv10      string   `json:"force_tlsv10,omitempty"`
	ForceTlsv11      string   `json:"force_tlsv11,omitempty"`
	ForceTlsv12      string   `json:"force_tlsv12,omitempty"`
	ForceTlsv13      string   `json:"force_tlsv13,omitempty"`
	ForceStrictSni   string   `json:"force_strict_sni,omitempty"`
	HealthCheckPort  int64    `json:"health_check_port,omitempty"`
	InitAddr         string   `json:"init-addr,omitempty"`
	Inter            int64    `json:"inter,omitempty"`
	Maintenance      string   `json:"maintenance,omitempty"`
	Maxconn          int64    `json:"maxconn,omitempty"`
	Maxqueue         int64    `json:"maxqueue,omitempty"`
	Minconn          int64    `json:"minconn,omitempty"`
	NoSslv3          string   `json:"no_sslv3,omitempty"`
	NoTlsv10         string   `json:"no_tlsv10,omitempty"`
	NoTlsv11         string   `json:"no_tlsv11,omitempty"`
	NoTlsv12         string   `json:"no_tlsv12,omitempty"`
	NoTlsv13         string   `json:"no_tlsv13,omitempty"`
	// New v3 fields (non-deprecated)
	Sslv3            string   `json:"sslv3,omitempty"`
	Tlsv10           string   `json:"tlsv10,omitempty"`
	Tlsv11           string   `json:"tlsv11,omitempty"`
	Tlsv12           string   `json:"tlsv12,omitempty"`
	Tlsv13           string   `json:"tlsv13,omitempty"`
	OnError          string   `json:"on-error,omitempty"`
	OnMarkedDown     string   `json:"on-marked-down,omitempty"`
	OnMarkedUp       string   `json:"on-marked-up,omitempty"`
	PoolLowConn      int64    `json:"pool_low_conn,omitempty"`
	PoolMaxConn      int64    `json:"pool_max_conn,omitempty"`
	PoolPurgeDelay   int64    `json:"pool_purge_delay,omitempty"`
	Proto            string   `json:"proto,omitempty"`
	ProxyV2Options   []string `json:"proxy-v2-options,omitempty"`
	Rise             int64    `json:"rise,omitempty"`
	SendProxy        string   `json:"send-proxy,omitempty"`
	SendProxyV2      string   `json:"send-proxy-v2,omitempty"`
	SendProxyV2Ssl   string   `json:"send-proxy-v2-ssl,omitempty"`
	SendProxyV2SslCn string   `json:"send-proxy-v2-ssl-cn,omitempty"`
	Slowstart        int64    `json:"slowstart,omitempty"`
	Sni              string   `json:"sni,omitempty"`
	Source           string   `json:"source,omitempty"`
	Ssl              string   `json:"ssl,omitempty"`
	SslCafile        string   `json:"ssl_cafile,omitempty"`
	SslCertificate   string   `json:"ssl_certificate,omitempty"`
	SslMaxVer        string   `json:"ssl_max_ver,omitempty"`
	SslMinVer        string   `json:"ssl_min_ver,omitempty"`
	SslReuse         string   `json:"ssl_reuse,omitempty"`
	Stick            string   `json:"stick,omitempty"`
	Tfo              string   `json:"tfo,omitempty"`
	TlsTickets       string   `json:"tls_tickets,omitempty"`
	Track            string   `json:"track,omitempty"`
	Verify           string   `json:"verify,omitempty"`
	Weight           int64    `json:"weight,omitempty"`
}

// BindPayload is the payload for the bind resource.
type BindPayload struct {
	Name                 string `json:"name"`
	Address              string `json:"address"`
	Port                 *int64 `json:"port,omitempty"`
	PortRangeEnd         *int64 `json:"port-range-end,omitempty"`
	Maxconn              int64  `json:"maxconn,omitempty"`
	User                 string `json:"user,omitempty"`
	Group                string `json:"group,omitempty"`
	Mode                 string `json:"mode,omitempty"`
	ForceSslv3           bool   `json:"force_sslv3,omitempty"`
	ForceTlsv10          bool   `json:"force_tlsv10,omitempty"`
	ForceTlsv11          bool   `json:"force_tlsv11,omitempty"`
	ForceTlsv12          bool   `json:"force_tlsv12,omitempty"`
	ForceTlsv13          bool   `json:"force_tlsv13,omitempty"`
	ForceStrictSni       string `json:"force_strict_sni,omitempty"`
	Ssl                  bool   `json:"ssl,omitempty"`
	SslCafile            string `json:"ssl_cafile,omitempty"`
	SslMaxVer            string `json:"ssl_max_ver,omitempty"`
	SslMinVer            string `json:"ssl_min_ver,omitempty"`
	SslCertificate       string `json:"ssl_certificate,omitempty"`
	Ciphers              string `json:"ciphers,omitempty"`
	Ciphersuites         string `json:"ciphersuites,omitempty"`
	Transparent          bool   `json:"transparent,omitempty"`
	AcceptProxy          bool   `json:"accept_proxy,omitempty"`
	Allow0rtt            bool   `json:"allow_0rtt,omitempty"`
	Alpn                 string `json:"alpn,omitempty"`
	Backlog              string `json:"backlog,omitempty"`
	CaIgnoreErr          string `json:"ca_ignore_err,omitempty"`
	CaSignFile           string `json:"ca_sign_file,omitempty"`
	CaSignPass           string `json:"ca_sign_pass,omitempty"`
	CaVerifyFile         string `json:"ca_verify_file,omitempty"`
	CrlFile              string `json:"crl_file,omitempty"`
	CrtIgnoreErr         string `json:"crt_ignore_err,omitempty"`
	CrtList              string `json:"crt_list,omitempty"`
	DeferAccept          bool   `json:"defer_accept,omitempty"`
	ExposeViaAgent       bool   `json:"expose_via_agent,omitempty"`
	GenerateCertificates bool   `json:"generate_certificates,omitempty"`
	Gid                  int64  `json:"gid,omitempty"`
	Id                   string `json:"id,omitempty"`
	Interface            string `json:"interface,omitempty"`
	Level                string `json:"level,omitempty"`
	LogProto             string `json:"log_proto,omitempty"`
	Mdev                 string `json:"mdev,omitempty"`
	Namespace            string `json:"namespace,omitempty"`
	Nice                 int64  `json:"nice,omitempty"`
	NoCaNames            bool   `json:"no_ca_names,omitempty"`
	NoSslv3              bool   `json:"no_sslv3,omitempty"`
	NoTlsv10             bool   `json:"no_tlsv10,omitempty"`
	NoTlsv11             bool   `json:"no_tlsv11,omitempty"`
	NoTlsv12             bool   `json:"no_tlsv12,omitempty"`
	NoTlsv13             bool   `json:"no_tlsv13,omitempty"`
	// New v3 fields (non-deprecated)
	Sslv3                bool   `json:"sslv3,omitempty"`
	Tlsv10               bool   `json:"tlsv10,omitempty"`
	Tlsv11               bool   `json:"tlsv11,omitempty"`
	Tlsv12               bool   `json:"tlsv12,omitempty"`
	Tlsv13               bool   `json:"tlsv13,omitempty"`
	Npn                  string `json:"npn,omitempty"`
	PreferClientCiphers  bool   `json:"prefer_client_ciphers,omitempty"`
	Process              string `json:"process,omitempty"`
	Proto                string `json:"proto,omitempty"`
	SeverityOutput       string `json:"severity_output,omitempty"`
	StrictSni            bool   `json:"strict_sni,omitempty"`
	TcpUserTimeout       int64  `json:"tcp_user_timeout,omitempty"`
	Tfo                  bool   `json:"tfo,omitempty"`
	TlsTicketKeys        string `json:"tls_ticket_keys,omitempty"`
	Uid                  string `json:"uid,omitempty"`
	V4v6                 bool   `json:"v4v6,omitempty"`
	V6only               bool   `json:"v6only,omitempty"`
	Verify               string `json:"verify,omitempty"`
	Metadata             string `json:"metadata,omitempty"`
}

// AclPayload is the payload for the acl resource.
type AclPayload struct {
	AclName   string `json:"acl_name"`
	Index     int64  `json:"index"`
	Criterion string `json:"criterion"`
	Value     string `json:"value"`
}

// HttpRequestRulePayload is the payload for the httprequestrule resource.
type HttpRequestRulePayload struct {
	Index                int64  `json:"index"`
	Type                 string `json:"type"`
	AclFile              string `json:"acl_file,omitempty"`
	AclKeyfmt            string `json:"acl_keyfmt,omitempty"`
	BandwidthLimitName   string `json:"bandwidth_limit_name,omitempty"`
	BandwidthLimitPeriod string `json:"bandwidth_limit_period,omitempty"`
	BandwidthLimitLimit  string `json:"bandwidth_limit_limit,omitempty"`
	CacheName            string `json:"cache_name,omitempty"`
	Cond                 string `json:"cond,omitempty"`
	CondTest             string `json:"cond_test,omitempty"`
	Expr                 string `json:"expr,omitempty"`
	HdrFormat            string `json:"hdr_format,omitempty"`
	HdrMatch             string `json:"hdr_match,omitempty"`
	HdrMethod            string `json:"hdr_method,omitempty"`
	HdrName              string `json:"hdr_name,omitempty"`
	LogLevel             string `json:"log_level,omitempty"`
	LuaAction            string `json:"lua_action,omitempty"`
	LuaParams            string `json:"lua_params,omitempty"`
	MapFile              string `json:"map_file,omitempty"`
	MapKeyfmt            string `json:"map_keyfmt,omitempty"`
	MapValuefmt          string `json:"map_valuefmt,omitempty"`
	MarkValue            string `json:"mark_value,omitempty"`
	MethodFmt            string `json:"method_fmt,omitempty"`
	NiceValue            int64  `json:"nice_value,omitempty"`
	PathFmt              string `json:"path_fmt,omitempty"`
	PathMatch            string `json:"path_match,omitempty"`
	QueryFmt             string `json:"query_fmt,omitempty"`
	RedirCode            int64  `json:"redir_code,omitempty"`
	RedirType            string `json:"redir_type,omitempty"`
	RedirValue           string `json:"redir_value,omitempty"`
	ScExpr               string `json:"sc_expr,omitempty"`
	ScId                 int64  `json:"sc_id,omitempty"`
	ScIdx                int64  `json:"sc_idx,omitempty"`
	ScInt                int64  `json:"sc_int,omitempty"`
	Service              string `json:"service,omitempty"`
	SpoeEngine           string `json:"spoe_engine,omitempty"`
	SpoeGroup            string `json:"spoe_group,omitempty"`
	StatusCode           int64  `json:"status_code,omitempty"`
	StatusReason         string `json:"status_reason,omitempty"`
	Timeout              string `json:"timeout,omitempty"`
	TimeoutValue         int64  `json:"timeout_value,omitempty"`
	TosValue             string `json:"tos_value,omitempty"`
	TrackScKey           string `json:"track_sc_key,omitempty"`
	TrackScTable         string `json:"track_sc_table,omitempty"`
	UriFmt               string `json:"uri_fmt,omitempty"`
	UriMatch             string `json:"uri_match,omitempty"`
	VarName              string `json:"var_name,omitempty"`
	VarScope             string `json:"var_scope,omitempty"`
	WaitTime             int64  `json:"wait_time,omitempty"`
}

// ResolverPayload is the payload for the resolver resource.
type ResolverPayload struct {
	Name                string `json:"name"`
	AcceptedPayloadSize int64  `json:"accepted_payload_size,omitempty"`
	HoldNx              int64  `json:"hold_nx,omitempty"`
	HoldObsolete        int64  `json:"hold_obsolete,omitempty"`
	HoldOther           int64  `json:"hold_other,omitempty"`
	HoldRefused         int64  `json:"hold_refused,omitempty"`
	HoldTimeout         int64  `json:"hold_timeout,omitempty"`
	HoldValid           int64  `json:"hold_valid,omitempty"`
	ResolveRetries      int64  `json:"resolve_retries,omitempty"`
	TimeoutResolve      int64  `json:"timeout_resolve,omitempty"`
	TimeoutRetry        int64  `json:"timeout_retry,omitempty"`
}

// NameserverPayload is the payload for the nameserver resource.
type NameserverPayload struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int64  `json:"port,omitempty"`
}

// PeersPayload is the payload for the peers resource.
type PeersPayload struct {
	Name string `json:"name"`
}

// PeerEntryPayload is the payload for the peer_entry resource.
type PeerEntryPayload struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int64  `json:"port,omitempty"`
}

// StickRulePayload is the payload for the stick_rule resource.
type StickRulePayload struct {
	Index    int64  `json:"index"`
	Type     string `json:"type"`
	Cond     string `json:"cond,omitempty"`
	CondTest string `json:"cond_test,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	Table    string `json:"table,omitempty"`
}

// StickTablePayload is the payload for the stick_table resource.
type StickTablePayload struct {
	Name    string `json:"name"`
	Type    string `json:"type,omitempty"`
	Size    string `json:"size,omitempty"`
	Store   string `json:"store,omitempty"`
	Peers   string `json:"peers,omitempty"`
	NoPurge bool   `json:"no_purge,omitempty"`
}

// HttpcheckPayload is the payload for the httpcheck resource.
type HttpcheckPayload struct {
	Index           int64  `json:"index"`
	Addr            string `json:"addr,omitempty"`
	Match           string `json:"match,omitempty"`
	Pattern         string `json:"pattern,omitempty"`
	Type            string `json:"type,omitempty"`
	Method          string `json:"method,omitempty"`
	Port            int64  `json:"port,omitempty"`
	Uri             string `json:"uri,omitempty"`
	Version         string `json:"version,omitempty"`
	ExclamationMark string `json:"exclamation_mark,omitempty"`
	LogLevel        string `json:"log_level,omitempty"`
	SendProxy       string `json:"send_proxy,omitempty"`
	ViaSocks4       string `json:"via_socks4,omitempty"`
	CheckComment    string `json:"check_comment,omitempty"`
}

// HttpResponseRulePayload is the payload for the httpresponserule resource.
type HttpResponseRulePayload struct {
	Index        int64  `json:"index"`
	Type         string `json:"type"`
	Cond         string `json:"cond,omitempty"`
	CondTest     string `json:"cond_test,omitempty"`
	HdrName      string `json:"hdr_name,omitempty"`
	HdrFormat    string `json:"hdr_format,omitempty"`
	RedirType    string `json:"redir_type,omitempty"`
	RedirValue   string `json:"redir_value,omitempty"`
	StatusCode   int64  `json:"status_code,omitempty"`
	StatusReason string `json:"status_reason,omitempty"`
}

// TcpCheckPayload is the payload for the tcpcheck resource.
type TcpCheckPayload struct {
	Index      int64  `json:"index"`
	Action     string `json:"action"`
	Comment    string `json:"comment,omitempty"`
	Port       int64  `json:"port,omitempty"`
	Address    string `json:"address,omitempty"`
	Data       string `json:"data,omitempty"`
	MinRecv    int64  `json:"min_recv,omitempty"`
	OnSuccess  string `json:"on_success,omitempty"`
	OnError    string `json:"on_error,omitempty"`
	StatusCode string `json:"status_code,omitempty"`
	Timeout    int64  `json:"timeout,omitempty"`
	LogLevel   string `json:"log_level,omitempty"`
}

// TcpRequestRulePayload is the payload for the tcprequestrule resource.
type TcpRequestRulePayload struct {
	Index        int64  `json:"index"`
	Type         string `json:"type"`
	Action       string `json:"action"`
	Cond         string `json:"cond,omitempty"`
	CondTest     string `json:"cond_test,omitempty"`
	Timeout      int64  `json:"timeout,omitempty"`
	LuaAction    string `json:"lua_action,omitempty"`
	LuaParams    string `json:"lua_params,omitempty"`
	ScId         int64  `json:"sc_id,omitempty"`
	ScIdx        int64  `json:"sc_idx,omitempty"`
	ScInt        int64  `json:"sc_int,omitempty"`
	ScIncGpc0    string `json:"sc_inc_gpc0,omitempty"`
	ScIncGpc1    string `json:"sc_inc_gpc1,omitempty"`
	ScSetGpt0    string `json:"sc_set_gpt0,omitempty"`
	TrackScKey   string `json:"track_sc_key,omitempty"`
	TrackScTable string `json:"track_sc_table,omitempty"`
	VarName      string `json:"var_name,omitempty"`
	VarScope     string `json:"var_scope,omitempty"`
	VarExpr      string `json:"var_expr,omitempty"`
	VarFormat    string `json:"var_format,omitempty"`
	VarType      string `json:"var_type,omitempty"`
}

// TcpResponseRulePayload is the payload for the tcpresponserule resource.
type TcpResponseRulePayload struct {
	Index     int64  `json:"index"`
	Action    string `json:"action"`
	Cond      string `json:"cond,omitempty"`
	CondTest  string `json:"cond_test,omitempty"`
	LuaAction string `json:"lua_action,omitempty"`
	LuaParams string `json:"lua_params,omitempty"`
	ScId      int64  `json:"sc_id,omitempty"`
	ScIdx     int64  `json:"sc_idx,omitempty"`
	ScInt     int64  `json:"sc_int,omitempty"`
	ScIncGpc0 string `json:"sc_inc_gpc0,omitempty"`
	ScIncGpc1 string `json:"sc_inc_gpc1,omitempty"`
	ScSetGpt0 string `json:"sc_set_gpt0,omitempty"`
	VarName   string `json:"var_name,omitempty"`
	VarScope  string `json:"var_scope,omitempty"`
	VarExpr   string `json:"var_expr,omitempty"`
	VarFormat string `json:"var_format,omitempty"`
	VarType   string `json:"var_type,omitempty"`
}

// LogForwardPayload is the payload for the logforward resource.
type LogForwardPayload struct {
	Name     string `json:"name"`
	Backlog  int64  `json:"backlog,omitempty"`
	Maxconn  int64  `json:"maxconn,omitempty"`
	Timeout  int64  `json:"timeout,omitempty"`
	Loglevel string `json:"loglevel,omitempty"`
}
