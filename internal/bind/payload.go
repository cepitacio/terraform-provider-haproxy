package bind

// Config defines variable for haproxy configuration
type ConfigBind struct {
	Username string
	Password string
	BaseURL  string
	SSL      bool
}

// BindWrapper wraps nested API response data under the "data" key.
type BindWrapper struct {
	Data BindPayload `json:"data"`
}

type BindPayload struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	Port           int    `json:"port"`
	Maxconn        int    `json:"maxconn"`
	User           string `json:"user"`
	Group          string `json:"group"`
	Mode           string `json:"mode"`
	ForceSslv3     bool   `json:"force_sslv3"`
	ForceTlsv10    bool   `json:"force_tlsv10"`
	ForceTlsv11    bool   `json:"force_tlsv11"`
	ForceTlsv12    bool   `json:"force_tlsv12"`
	ForceTlsv13    bool   `json:"force_tlsv13"`
	Ssl            bool   `json:"ssl"`
	SslCafile      string `json:"ssl_cafile"`
	SslMaxVer      string `json:"ssl_max_ver"`
	SslMinVer      string `json:"ssl_min_ver"`
	SslCertificate string `json:"ssl_certificate"`
	Ciphers        string `json:"ciphers"`
	CipherSuites   string `json:"ciphersuites"`
	Transparent    bool   `json:"transparent"`
}
