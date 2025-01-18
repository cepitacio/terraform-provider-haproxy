package acl

// Config defines variable for haproxy configuration
type ConfigAcl struct {
	Username string
	Password string
	BaseURL  string
	SSL      bool
}

// AclWrapper wraps nested API response data under the "data" key.
type AclWrapper struct {
	Data []AclPayload `json:"data"`
}

type AclPayload struct {
	AclName   string `json:"acl_name"`
	Criterion string `json:"criterion"`
	Index     int    `json:"index"`
	Value     string `json:"value"`
}
