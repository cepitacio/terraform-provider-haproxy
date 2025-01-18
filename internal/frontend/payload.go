package frontend

// Config defines variable for haproxy configuration
type ConfigFrontend struct {
	Username string
	Password string
	BaseURL  string
	SSL      bool
}

// FrontendWrapper wraps nested API response data under the "data" key.
type FrontendWrapper struct {
	Data FrontendPayload `json:"data"`
}

type FrontendPayload struct {
	Name                     string      `json:"name"`
	DefaultBackend           string      `json:"default_backend"`
	HttpConnectionMode       string      `json:"http_connection_mode"`
	AcceptInvalidHttpRequest string      `json:"accept_invalid_http_request"`
	MaxConn                  int         `json:"maxconn"`
	Mode                     string      `json:"mode"`
	Backlog                  int         `json:"backlog"`
	HttpKeepAliveTimeout     int         `json:"http_keep_alive_timeout"`
	HttpRequestTimeout       int         `json:"http_request_timeout"`
	HttpUseProxyHeader       string      `json:"http_use_proxy_header"`
	HttpLog                  bool        `json:"httplog"`
	HttpsLog                 string      `json:"httpslog"`
	ErrorLogFormat           string      `json:"error_log_format"`
	LogFormat                string      `json:"log_format"`
	LogFormatSd              string      `json:"log_format_sd"`
	MonitorUri               string      `json:"monitor_uri"`
	TcpLog                   bool        `json:"tcplog"`
	From                     string      `json:"from"`
	Compression              Compression `json:"compression"`
	Forwardfor               Forwardfor  `json:"forwardfor"`
	MonitorFail              MonitorFail `json:"monitor_fail"`
}

type Compression struct {
	Algorithms []string `json:"algorithms"`
	Offload    bool     `json:"offload"`
	Types      []string `json:"types"`
}

type Forwardfor struct {
	Enabled string `json:"enabled"`
	Except  string `json:"except"`
	Header  string `json:"header"`
	Ifnone  bool   `json:"ifnone"`
}

type MonitorFail struct {
	Cond     string `json:"cond"`
	CondTest string `json:"cond_test"`
}
