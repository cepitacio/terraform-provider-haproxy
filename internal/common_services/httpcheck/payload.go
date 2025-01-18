package httpcheck

// Config defines variable for haproxy configuration
type ConfigHttpCheck struct {
	Username string
	Password string
	BaseURL  string
	SSL      bool
}

// HttpCheckWrapper wraps nested API response data under the "data" key.
type HttpCheckWrapper struct {
	Data []HttpCheckPayload `json:"data"`
}

type HttpCheckPayload struct {
	Index   int    `json:"index"`
	Match   string `json:"match"`
	Pattern string `json:"pattern"`
	Type    string `json:"type"`
	Addr    string `json:"addr"`
	Port    *int   `json:"port,omitempty"`
	Method  string `json:"method"`
}
type Headers struct {
	Name string `json:"name"`
	Fmt  string `json:"fmt"`
}
