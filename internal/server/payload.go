package server

import "terraform-provider-haproxy/internal/transaction"

// Config defines variable for haproxy configuration
type ConfigServer struct {
	Username      string
	Password      string
	BaseURL       string
	SSL           bool
	TransactionID string
}

// ServerWrapper wraps nested API response data under the "data" key.
type ServerWrapper struct {
	Data ServerPayload `json:"data"`
}

type ServerPayload struct {
	Name              string `json:"name"`
	Address           string `json:"address"`
	Port              int    `json:"port"`
	SendProxy         string `json:"send-proxy"`
	Timeout           int    `json:"timeout"`
	Check             string `json:"check"`
	CheckSsl          string `json:"check-ssl"`
	Inter             int    `json:"inter"`
	Rise              int    `json:"rise"`
	Fall              int    `json:"fall"`
	Ssl               string `json:"ssl"`
	Ssl_cafile        string `json:"ssl_cafile"`
	Ssl_certificate   string `json:"ssl_certificate"`
	Ssl_max_ver       string `json:"ssl_max_ver"`
	Ssl_min_ver       string `json:"ssl_min_ver"`
	Ssl_reuse         string `json:"ssl_reuse"`
	Verify            string `json:"verify"`
	Health_check_port int    `json:"health_check_port"`
	Weight            int    `json:"weight"`
	Ciphersuites      string `json:"ciphersuites"`
	Force_sslv3       string `json:"force_sslv3"`
	Force_tlsv10      string `json:"force_tlsv10"`
	Force_tlsv11      string `json:"force_tlsv11"`
	Force_tlsv12      string `json:"force_tlsv12"`
	Force_tlsv13      string `json:"force_tlsv13"`
}

type Manager struct {
	Config *transaction.ConfigTransaction
}
