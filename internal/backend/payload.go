package backend

import "terraform-provider-haproxy/internal/transaction"

// Config defines variable for backend configuration
type ConfigBackend struct {
	Username      string
	Password      string
	BaseURL       string
	SSL           bool
	TransactionID string
}

// BackendWrapper wraps nested API response data under the "data" key.
type BackendWrapper struct {
	Data BackendPayload `json:"data"`
}

type BackendPayload struct {
	Name               string        `json:"name"`
	Mode               string        `json:"mode"`
	AdvCheck           string        `json:"adv_check"`
	HttpConnectionMode string        `json:"http_connection_mode"`
	ServerTimeout      int           `json:"server_timeout"`
	CheckTimeout       int           `json:"check_timeout"`
	ConnectTimeout     int           `json:"connect_timeout"`
	QueueTimeout       int           `json:"queue_timeout"`
	TunnelTimeout      int           `json:"tunnel_timeout"`
	TarpitTimeout      int           `json:"tarpit_timeout"`
	CheckCache         string        `json:"checkcache"`
	Retries            int           `json:"retries"`
	Balance            Balance       `json:"balance"`
	HttpchkParams      HttpchkParams `json:"httpchk_params"`
	Forwardfor         ForwardFor    `json:"forwardfor"`
	HttpCheck          HttpCheck     `json:"http-check"`
}

type Balance struct {
	Algorithm string `json:"algorithm"`
	UrlParam  string `json:"url_param"`
	// UrlParamCheckPost int    `json:"url_param_check_post"`
	// UrlParamMaxWait   int    `json:"url_param_max_wait"`
}

type HttpchkParams struct {
	Method  string `json:"method"`
	Uri     string `json:"uri"`
	Version string `json:"version"`
}

type ForwardFor struct {
	Enabled string `json:"enabled"`
}

type HttpCheck struct {
	Index   int    `json:"index"`
	Match   string `json:"match"`
	Pattern string `json:"pattern"`
	Type    string `json:"type"`
	Address string `json:"address"`
	Port    *int   `json:"port,omitempty"`
	Method  string `json:"method"`
}

type Manager struct {
	Config *transaction.ConfigTransaction
}
