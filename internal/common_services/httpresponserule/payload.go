package httpresponserule

// Config defines variable for haproxy configuration
type ConfigHttpResponseRule struct {
	Username string
	Password string
	BaseURL  string
	SSL      bool
}

// HttpRequestRuleWrapper wraps nested API response data under the "data" key.
type HttpResponseRuleWrapper struct {
	Data []HttpResponseRulePayload `json:"data"`
}

type HttpResponseRulePayload struct {
	Index                int    `json:"index"`
	Type                 string `json:"type"`
	AclFile              string `json:"acl_file"`
	AclKeyFmt            string `json:"acl_keyfmt"`
	AuthRealm            string `json:"auth_realm"`
	BandwidthLimitLimit  string `json:"bandwidth_limit_limit"`
	BandwidthLimitName   string `json:"bandwidth_limit_name"`
	BandwidthLimitPeriod string `json:"bandwidth_limit_period"`
	Cond                 string `json:"cond"`
	CondTest             string `json:"cond_test"`
	// DenyStatus          int    `json:"deny_status"`
	Expr      string `json:"expr"`
	HdrFormat string `json:"hdr_format"`
	HdrMatch  string `json:"hdr_match"`
	HdrMethod string `json:"hdr_method"`
	HdrName   string `json:"hdr_name"`
	LogLevel  string `json:"log_level"`
	// ReturnStatusCode    int    `json:"return_status_code"`
	ServiceName string `json:"service_name"`
	Timeout     string `json:"timeout"`
	TimeoutType string `json:"timeout_type"`
	UriFmt      string `json:"uri-fmt"`
	UriMatch    string `json:"uri-match"`
	VarExpr     string `json:"var_expr"`
	VarFormat   string `json:"var_format"`
	VarName     string `json:"var_name"`
	VarScope    string `json:"var_scope"`
	WaitAtLeast int    `json:"wait_at_least"`
	WaitTime    int    `json:"wait_time"`
	RedirType   string `json:"redir_type"`
	RedirValue  string `json:"redir_value"`
	ParentName  string `json:"parent_name"`
	ParentType  string `json:"parent_type"`
}
