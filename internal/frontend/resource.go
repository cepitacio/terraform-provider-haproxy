package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"terraform-provider-haproxy/internal/common_services/acl"
	"terraform-provider-haproxy/internal/common_services/httprequestrule"
	"terraform-provider-haproxy/internal/common_services/httpresponserule"
	"terraform-provider-haproxy/internal/transaction"
	"terraform-provider-haproxy/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceHaproxyFrontend() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceHaproxyFrontendCreate,
		ReadContext:   ResourceHaproxyFrontendRead,
		UpdateContext: ResourceHaproxyFrontendUpdate,
		DeleteContext: ResourceHaproxyFrontendDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the frontend. It must be unique and cannot be changed.",
			},
			"default_backend": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the default_backend. Pattern: ^[A-Za-z0-9-_.:]+$",
			},
			"http_connection_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The http connection mode of the frontend. Allowed: httpclose|http-server-close|http-keep-alive",
			},
			"accept_invalid_http_request": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The accept invalid http request of the frontend. Allowed: enabled|disabled",
			},
			"maxconn": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The max connection of the frontend.",
			},
			"mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The mode of the frontend. Allowed: http|tcp",
			},
			"backlog": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The backlog of the frontend.",
			},
			"http_keep_alive_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The http keep alive timeout of the frontend.",
			},
			"http_request_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The http request timeout of the frontend.",
			},
			"http_use_proxy_header": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The http use proxy header of the frontend. Allowed: enabled|disabled",
			},
			"httplog": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "The http log of the frontend.",
			},
			"httpslog": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The https log of the frontend. Allowed: enabled|disabled",
			},
			"error_log_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The error log format of the frontend.",
			},
			"log_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The log format of the frontend.",
			},
			"log_format_sd": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The log format sd of the frontend.",
			},
			"monitor_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The monitor uri of the frontend.",
			},
			"tcplog": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "The tcp log of the frontend.",
			},
			"from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The from of the frontend.",
			},
			"monitor_fail": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The monitor_fail of the frontend.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cond": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The cond of the monitor_fail. Allowed: if|unless",
						},
						"cond_test": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The cond_test of the monitor_fail.",
						},
					},
				},
			},
			// "compression": {
			// 	Type:        schema.TypeSet,
			// 	Optional:    true,
			// 	Description: "The compression of the frontend.",
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"algorithms": {
			// 				Type:        schema.TypeList,
			// 				Optional:    true,
			// 				Description: "The algorithms of the compression.",
			// 				Elem: &schema.Schema{
			// 					Type: schema.TypeString,
			// 				},
			// 			},
			// 			"offload": {
			// 				Type:        schema.TypeBool,
			// 				Optional:    true,
			// 				Description: "The offload of the compression.",
			// 			},
			// 			"types": {
			// 				Type:        schema.TypeList,
			// 				Optional:    true,
			// 				Description: "The types of the compression.",
			// 				Elem: &schema.Schema{
			// 					Type: schema.TypeString,
			// 				},
			// 			},
			// 		},
			// 	},
			// },
			// "forwardfor": {
			// 	Type:        schema.TypeSet,
			// 	Optional:    true,
			// 	Description: "The forwardfor of the frontend.",
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"enabled": {
			// 				Type:        schema.TypeString,
			// 				Optional:    true,
			// 				Description: "The enabled of the forwardfor.",
			// 			},
			// 			"except": {
			// 				Type:        schema.TypeString,
			// 				Optional:    true,
			// 				Description: "The except of the forwardfor.",
			// 			},
			// 			"header": {
			// 				Type:        schema.TypeString,
			// 				Optional:    true,
			// 				Description: "The header of the forwardfor.",
			// 			},
			// 			"ifnone": {
			// 				Type:        schema.TypeBool,
			// 				Optional:    true,
			// 				Description: "The ifnone of the forwardfor.",
			// 			},
			// 		},
			// 	},
			// },
			"acl": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of ACLs to configure",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"acl_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The acl name. Pattern: ^[^\\s]+$",
						},
						"index": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The index of the acl",
						},
						"criterion": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The criterion. Pattern: ^[^\\s]+$",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The value of the criterion",
						},
					},
				},
			},
			"httprequestrule": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of httprequest to configure",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The index of the httpresponserules in the parent object starting at 0",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of the check, Allowed: Allowed: add-acl|add-header|allow|auth|cache-use|capture|del-acl|del-header|del-map|deny|disable-l7-retry|do-resolve|early-hint|lua|normalize-uri|redirect|reject|replace-header|replace-path|replace-pathq|replace-uri|replace-value|return|sc-add-gpc|sc-inc-gpc|sc-inc-gpc0|sc-inc-gpc1|sc-set-gpt|sc-set-gpt0|send-spoe-group|set-dst|set-dst-port|set-header|set-log-level|set-map|set-mark|set-method|set-nice|set-path|set-pathq|set-priority-class|set-priority-offset|set-query|set-src|set-src-port|set-timeout|set-tos|set-uri|set-var|silent-drop|strict-mode|tarpit|track-sc0|track-sc1|track-sc2|track-sc|unset-var|use-service|wait-for-body|wait-for-handshake|set-bandwidth-limit",
						},
						"cond": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The condition of the httpresponserules. Allowed: if|unless",
						},
						"cond_test": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The condition test of the httpresponserules",
						},
						"hdr_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The header name of the httpresponserules",
						},
						"hdr_format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The header format of the httpresponserules. Pattern: ^[^\\s]+$",
						},
						"redir_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The redirection type of the httpresponserules. Allowed: location|prefix|scheme",
						},
						"redir_value": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The redirection value of the httpresponserules. Pattern: ^[^\\s]+$",
						},
					},
				},
			},
			"httpresponserule": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of httprequest to configure",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The index of the httpresponserules in the parent object starting at 0",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of the check, Allowed: Allowed: add-acl|add-header|allow|cache-store|capture|del-acl|del-header|del-map|deny|lua|redirect|replace-header|replace-value|return|sc-add-gpc|sc-inc-gpc|sc-inc-gpc0|sc-inc-gpc1|sc-set-gpt|sc-set-gpt0|send-spoe-group|set-header|set-log-level|set-map|set-mark|set-nice|set-status|set-timeout|set-tos|set-var|set-var-fmt|silent-drop|strict-mode|track-sc0|track-sc1|track-sc2|track-sc|unset-var|wait-for-body|set-bandwidth-limit",
						},
						"cond": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The condition of the httpresponserules. Allowed: if|unless",
						},
						"cond_test": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The condition test of the httpresponserules",
						},
						"hdr_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The header name of the httpresponserules",
						},
						"hdr_format": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The header format of the httpresponserules",
						},
						"redir_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The redirection type of the httpresponserules. Allowed: location|prefix|scheme",
						},
						"redir_value": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The redirection value of the httpresponserules. Pattern: ^[^\\s]+$",
						},
					},
				},
			},
		},
	}
}

func ResourceHaproxyFrontendCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	frontendName := d.Get("name").(string)
	parentName := frontendName
	parentType := "frontend"
	var (
		// compressionOffload bool = false
		// // forwardforEnabled     bool = false
		// forwardforExcept      string
		// forwardforHeader      string
		// forwardforIfnone      bool = false
		// compressionAlgorithms []string
		// compressionTypes      []string
		// enabledStr            string = "enabled"
		monitorfailCond     string
		monitorfailcondTest string
	)

	// compressionItem := utils.GetFirstItemValue(d.Get, "compression")
	// if compressionItem != nil {
	// 	// Read the compression block
	// 	compression := d.Get("compression").(*schema.Set).List()
	// 	compressionData := compression[0].(map[string]interface{})
	// 	compressionAlgorithmsRaw := compressionData["algorithms"].([]interface{})
	// 	for _, algorithm := range compressionAlgorithmsRaw {
	// 		compressionAlgorithms = append(compressionAlgorithms, algorithm.(string))
	// 	}

	// 	compressionOffload = compressionData["offload"].(bool)

	// 	// Corrected handling of the 'types' attribute
	// 	compressionTypesRaw := compressionData["types"].([]interface{})
	// 	for _, t := range compressionTypesRaw {
	// 		compressionTypes = append(compressionTypes, t.(string))
	// 	}
	// }

	// forwardforItem := utils.GetFirstItemValue(d.Get, "forwardfor")
	// if forwardforItem != nil {
	// 	//Read the forwardfor block
	// 	forwardfor := d.Get("forwardfor").(*schema.Set).List()
	// 	forwardforEnabled = forwardfor[0].(map[string]interface{})["enabled"].(bool)
	// 	forwardforExcept = forwardfor[0].(map[string]interface{})["except"].(string)
	// 	forwardforHeader = forwardfor[0].(map[string]interface{})["header"].(string)
	// 	forwardforIfnone = forwardfor[0].(map[string]interface{})["ifnone"].(bool)
	// 	enabledStr = utils.BoolToStr(forwardforEnabled)
	// }

	monitorfailItem := utils.GetFirstItemValue(d.Get, "monitor_fail")

	if monitorfailItem != nil {
		//Read the monitorfail block
		monitor_fail := d.Get("monitor_fail").(*schema.Set).List()
		monitorfailCond = monitor_fail[0].(map[string]interface{})["cond"].(string)
		monitorfailcondTest = monitor_fail[0].(map[string]interface{})["cond_test"].(string)
	}

	payload := FrontendPayload{
		Name:                     frontendName,
		DefaultBackend:           d.Get("default_backend").(string),
		HttpConnectionMode:       d.Get("http_connection_mode").(string),
		MaxConn:                  d.Get("maxconn").(int),
		Mode:                     d.Get("mode").(string),
		Backlog:                  d.Get("backlog").(int),
		HttpKeepAliveTimeout:     d.Get("http_keep_alive_timeout").(int),
		HttpRequestTimeout:       d.Get("http_request_timeout").(int),
		HttpLog:                  d.Get("httplog").(bool),
		HttpsLog:                 d.Get("httpslog").(string),
		ErrorLogFormat:           d.Get("error_log_format").(string),
		LogFormat:                d.Get("log_format").(string),
		LogFormatSd:              d.Get("log_format_sd").(string),
		MonitorUri:               d.Get("monitor_uri").(string),
		TcpLog:                   d.Get("tcplog").(bool),
		From:                     d.Get("from").(string),
		AcceptInvalidHttpRequest: d.Get("accept_invalid_http_request").(string),
		HttpUseProxyHeader:       d.Get("http_use_proxy_header").(string),
	}

	// Check if AcceptInvalidHttpRequest is set
	// if acceptInvalidHttpRequest {
	// 	payload.AcceptInvalidHttpRequest = utils.BoolToStr(acceptInvalidHttpRequest)
	// }

	// Check if compression is set
	// if compressionItem != nil {
	// 	payload.Compression = Compression{
	// 		Algorithms: compressionAlgorithms,
	// 		Offload:    compressionOffload,
	// 		Types:      compressionTypes,
	// 	}
	// }

	// Check if forwardfor is set
	// if forwardforItem != nil {
	// 	payload.Forwardfor = Forwardfor{
	// 		Enabled: enabledStr,
	// 		Except:  forwardforExcept,
	// 		Header:  forwardforHeader,
	// 		Ifnone:  forwardforIfnone,
	// 	}
	// }

	//Check if monitorfail is set
	if monitorfailItem != nil {
		payload.MonitorFail = MonitorFail{
			Cond:     monitorfailCond,
			CondTest: monitorfailcondTest,
		}
	}
	payloadJSON, err := utils.MarshalNonZeroFields(payload)
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal frontend payload",
				Detail:   fmt.Sprintf("Error while marshaling frontend create payload: %s", err.Error()),
			},
		}
	}

	configMap := m.(map[string]interface{})
	frontendConfig := configMap["frontend"].(*ConfigFrontend)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	aclConfig := configMap["acl"].(*acl.ConfigAcl)
	httprequestruleConfig := configMap["httprequestrule"].(*httprequestrule.ConfigHttpRequestRule)
	httpresponseruleConfig := configMap["httpresponserule"].(*httpresponserule.ConfigHttpResponseRule)

	aclItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "acl")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal aclItems payload",
				Detail:   fmt.Sprintf("Error while marshaling aclsItem payload: %s", err.Error()),
			},
		}
	}

	httprequestruleItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "httprequestrule")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httprequestrulesItem payload",
				Detail:   fmt.Sprintf("Error while marshaling httprequestrulesItem payload: %s", err.Error()),
			},
		}
	}

	httpresponseruleItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "httpresponserule")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httpresponserulesItem payload",
				Detail:   fmt.Sprintf("Error while marshaling httpresponserulesItem payload: %s", err.Error()),
			},
		}
	}

	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		front_resp, err := frontendConfig.AddFrontendConfiguration(payloadJSON, transactionID)
		diags = utils.HandleHTTPResponse(front_resp, err, "Error during frontend configuration")
		if len(diags) > 0 {
			utils.PrintDiags(diags)
			return front_resp, err
		}

		// var lastResp *http.Response
		for _, item := range aclItems {
			acl_payloadJSON, _ := json.Marshal(item)

			lastResp, err := aclConfig.AddAnAclConfiguration(acl_payloadJSON, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during acl configuration")
			if len(diags) > 0 {
				return lastResp, err
			}
		}

		for _, item := range httprequestruleItems {
			payloadJSON, _ := json.Marshal(item)

			lastResp, err := httprequestruleConfig.AddAHttpRequestRuleConfiguration(payloadJSON, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during httprequestrule configuration")
			if diags != nil {
				return lastResp, err
			}
		}
		for _, item := range httpresponseruleItems {
			payloadJSON, _ := json.Marshal(item)

			lastResp, err := httpresponseruleConfig.AddAHttpResponseRuleConfiguration(payloadJSON, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during httpresponserule configuration")
			if diags != nil {
				return lastResp, err
			}
		}
		return front_resp, err
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during transaction commit")
	utils.PrintDiags(diags)
	if len(diags) > 0 {
		return diags
	}
	d.SetId(frontendName)
	return ResourceHaproxyFrontendRead(ctx, d, m)
}

func ResourceHaproxyFrontendRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	frontendName := d.Get("name").(string)
	parentName := frontendName
	parentType := "frontend"
	configMap := m.(map[string]interface{})
	frontendConfig := configMap["frontend"].(*ConfigFrontend)

	aclresourceConfig := configMap["acl"].(*acl.ConfigAcl)
	httprequestruleresourceConfig := configMap["httprequestrule"].(*httprequestrule.ConfigHttpRequestRule)
	httpresponseruleresourceConfig := configMap["httpresponserule"].(*httpresponserule.ConfigHttpResponseRule)

	//acl read starts here
	aclItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "acl")
	fmt.Println("Length of aclItems:", len(aclItems))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal aclItems payload",
				Detail:   fmt.Sprintf("Error while marshaling aclItem payload: %s", err.Error()),
			},
		}
	}

	aclupdatedItems := make([]map[string]interface{}, 0, len(aclItems))
	var acllastResp *http.Response
	aclprocessed := make(map[string]map[string]bool)
	for range aclItems {

		if _, exists := aclprocessed[parentName]; exists {
			if aclprocessed[parentName][parentType] {
				continue
			}
		} else {
			aclprocessed[parentName] = make(map[string]bool)
		}

		aclprocessed[parentName][parentType] = true
		acllastResp, err = aclresourceConfig.GetAllAclConfiguration(parentName, parentType)
		diags = utils.HandleHTTPResponse(acllastResp, err, "Error during acl configuration")
		if diags != nil {
			return diags
		}
		var aclWrapper acl.AclWrapper
		if err := json.NewDecoder(acllastResp.Body).Decode(&aclWrapper); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing acl configuration: %w", err))
		}
		defer acllastResp.Body.Close()

		for _, item := range aclWrapper.Data {
			updatedacl := map[string]interface{}{
				"acl_name":  item.AclName,
				"criterion": item.Criterion,
				"index":     item.Index,
				"value":     item.Value,
			}
			aclupdatedItems = append(aclupdatedItems, updatedacl)
		}
	}
	fmt.Printf("All acls items from API: %+v\n", aclupdatedItems)
	d.Set("acl", aclupdatedItems)

	// httprequestrule read starts here
	httprequestruleItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "httprequestrule")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httprequestruleItems payload",
				Detail:   fmt.Sprintf("Error while marshaling httprequestruleItem payload: %s", err.Error()),
			},
		}
	}

	httprequestruleupdatedItems := make([]map[string]interface{}, 0, len(httprequestruleItems))
	var httprequestrulelastResp *http.Response
	httprequestruleprocessed := make(map[string]map[string]bool)
	for range httprequestruleItems {

		// Check if the pair has already been processed
		if _, exists := httprequestruleprocessed[parentName]; exists {
			if httprequestruleprocessed[parentName][parentType] {
				// Skip this combination as it is already processed
				continue
			}
		} else {
			// Initialize map for this parentName
			httprequestruleprocessed[parentName] = make(map[string]bool)
		}

		// Mark the pair as processed
		httprequestruleprocessed[parentName][parentType] = true
		httprequestrulelastResp, err = httprequestruleresourceConfig.GetAllHttpRequestRuleConfiguration(parentName, parentType)
		diags = utils.HandleHTTPResponse(httprequestrulelastResp, err, "Error during httpresponserule configuration")
		if diags != nil {
			return diags
		}
		var httprequestrulesWrapper httprequestrule.HttpRequestRuleWrapper
		if err := json.NewDecoder(httprequestrulelastResp.Body).Decode(&httprequestrulesWrapper); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing httpresponserule configuration: %w", err))
		}
		defer httprequestrulelastResp.Body.Close() // Move it after decoding

		// Iterate over aclWrapper.Data and process
		for _, item := range httprequestrulesWrapper.Data {
			updatedHttprequestrule := map[string]interface{}{
				"index":       item.Index,
				"type":        item.Type,
				"cond":        item.Cond,
				"cond_test":   item.CondTest,
				"hdr_name":    item.HdrName,
				"hdr_format":  item.HdrFormat,
				"redir_type":  item.RedirType,
				"redir_value": item.RedirValue,
			}
			httprequestruleupdatedItems = append(httprequestruleupdatedItems, updatedHttprequestrule)
		}
	}
	fmt.Printf("All Httprequestrules items from API: %+v\n", httprequestruleupdatedItems)
	d.Set("httprequestrule", httprequestruleupdatedItems)

	// httpresponse read starts here
	httpresonseruleItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "httpresponserule")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httpresponseruleItems payload",
				Detail:   fmt.Sprintf("Error while marshaling httpresponseruleItem payload: %s", err.Error()),
			},
		}
	}

	httpresponseruleupdatedItems := make([]map[string]interface{}, 0, len(httpresonseruleItems))
	var httpresponserulelastResp *http.Response
	httpresponseruleprocessed := make(map[string]map[string]bool)
	for range httpresonseruleItems {

		// Check if the pair has already been processed
		if _, exists := httpresponseruleprocessed[parentName]; exists {
			if httpresponseruleprocessed[parentName][parentType] {
				// Skip this combination as it is already processed
				continue
			}
		} else {
			// Initialize map for this parentName
			httpresponseruleprocessed[parentName] = make(map[string]bool)
		}

		// Mark the pair as processed
		httpresponseruleprocessed[parentName][parentType] = true
		httpresponserulelastResp, err = httpresponseruleresourceConfig.GetAllHttpResponseRuleConfiguration(parentName, parentType)
		diags = utils.HandleHTTPResponse(httpresponserulelastResp, err, "Error during httpresponserule configuration")
		if diags != nil {
			return diags
		}
		var httpreponserulesWrapper httpresponserule.HttpResponseRuleWrapper
		if err := json.NewDecoder(httpresponserulelastResp.Body).Decode(&httpreponserulesWrapper); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing httpresponserule configuration: %w", err))
		}
		defer httpresponserulelastResp.Body.Close() // Move it after decoding

		// Iterate over aclWrapper.Data and process
		for _, item := range httpreponserulesWrapper.Data {
			updatedHttprequestrule := map[string]interface{}{
				"index":       item.Index,
				"type":        item.Type,
				"cond":        item.Cond,
				"cond_test":   item.CondTest,
				"hdr_name":    item.HdrName,
				"hdr_format":  item.HdrFormat,
				"redir_type":  item.RedirType,
				"redir_value": item.RedirValue,
			}
			httpresponseruleupdatedItems = append(httpresponseruleupdatedItems, updatedHttprequestrule)
		}
	}
	fmt.Printf("All Httpresponserules items from API: %+v\n", httpresponseruleupdatedItems)
	d.Set("httpresponserule", httpresponseruleupdatedItems)

	//frontend read starts here
	resp, err := frontendConfig.GetAFrontendConfiguration(frontendName)

	if err != nil {
		if strings.Contains(err.Error(), "missing object") || resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
	}

	diags = utils.HandleHTTPResponse(resp, err, "Error during frontend configuration")

	if len(diags) > 0 {
		return diags
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	defer resp.Body.Close()

	var frontendWrapper FrontendWrapper

	err = json.Unmarshal(bodyBytes, &frontendWrapper)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing bind configuration: %w", err))
	}

	monitorFail := frontendWrapper.Data.MonitorFail

	if monitorFail == (MonitorFail{}) {
		d.Set("monitor_fail", nil)
	} else {
		d.Set("monitor_fail", []interface{}{
			map[string]interface{}{
				"cond":      monitorFail.Cond,
				"cond_test": monitorFail.CondTest,
			},
		})
	}

	d.SetId(frontendName)
	d.Set("default_backend", frontendWrapper.Data.DefaultBackend)
	d.Set("http_connection_mode", frontendWrapper.Data.HttpConnectionMode)
	d.Set("accept_invalid_http_request", frontendWrapper.Data.AcceptInvalidHttpRequest)
	d.Set("maxconn", frontendWrapper.Data.MaxConn)
	d.Set("mode", frontendWrapper.Data.Mode)
	d.Set("backlog", frontendWrapper.Data.Backlog)
	d.Set("http_keep_alive_timeout", frontendWrapper.Data.HttpKeepAliveTimeout)
	d.Set("http_request_timeout", frontendWrapper.Data.HttpRequestTimeout)
	d.Set("http_use_proxy_header", frontendWrapper.Data.HttpUseProxyHeader)
	d.Set("httplog", frontendWrapper.Data.HttpLog)
	d.Set("httpslog", frontendWrapper.Data.HttpsLog)
	d.Set("error_log_format", frontendWrapper.Data.ErrorLogFormat)
	d.Set("log_format", frontendWrapper.Data.LogFormat)
	d.Set("log_format_sd", frontendWrapper.Data.LogFormatSd)
	d.Set("monitor_uri", frontendWrapper.Data.MonitorUri)
	d.Set("tcplog", frontendWrapper.Data.TcpLog)
	d.Set("from", frontendWrapper.Data.From)

	return diags
}

func ResourceHaproxyFrontendUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	frontendName := d.Get("name").(string)
	parentName := frontendName
	parentType := "frontend"
	// acceptInvalidHttpRequest := d.Get("accept_invalid_http_request").(string)
	// httpslog := d.Get("httpslog").(bool)
	// httpUseProxyHeader := d.Get("http_use_proxy_header").(bool)

	var (
		// compressionOffload    bool = false
		// forwardforEnabled     bool = false
		// forwardforExcept      string
		// forwardforHeader      string
		// forwardforIfnone      bool = false
		// compressionAlgorithms []string
		// compressionTypes      []string
		// enabledStr            string = "enabled"
		monitorfailCond     string
		monitorfailcondTest string
	)

	// compressionItem := utils.GetFirstItemValue(d.Get, "compression")
	// if compressionItem != nil {
	// 	// Read the compression block
	// 	compression := d.Get("compression").(*schema.Set).List()
	// 	compressionData := compression[0].(map[string]interface{})
	// 	compressionAlgorithmsRaw := compressionData["algorithms"].([]interface{})
	// 	for _, algorithm := range compressionAlgorithmsRaw {
	// 		compressionAlgorithms = append(compressionAlgorithms, algorithm.(string))
	// 	}

	// 	compressionOffload = compressionData["offload"].(bool)

	// 	// Corrected handling of the 'types' attribute
	// 	compressionTypesRaw := compressionData["types"].([]interface{})
	// 	for _, t := range compressionTypesRaw {
	// 		compressionTypes = append(compressionTypes, t.(string))
	// 	}
	// }

	// forwardforItem := utils.GetFirstItemValue(d.Get, "forwardfor")
	// if forwardforItem != nil {
	// 	//Read the forwardfor block
	// 	forwardfor := d.Get("forwardfor").(*schema.Set).List()
	// 	forwardforEnabled = forwardfor[0].(map[string]interface{})["enabled"].(bool)
	// 	forwardforExcept = forwardfor[0].(map[string]interface{})["except"].(string)
	// 	forwardforHeader = forwardfor[0].(map[string]interface{})["header"].(string)
	// 	forwardforIfnone = forwardfor[0].(map[string]interface{})["ifnone"].(bool)
	// 	enabledStr = utils.BoolToStr(forwardforEnabled)
	// }

	monitorfailItem := utils.GetFirstItemValue(d.Get, "monitor_fail")

	if monitorfailItem != nil {
		//Read the monitorfail block
		monitor_fail := d.Get("monitor_fail").(*schema.Set).List()
		monitorfailCond = monitor_fail[0].(map[string]interface{})["cond"].(string)
		monitorfailcondTest = monitor_fail[0].(map[string]interface{})["cond_test"].(string)
	}

	payload := FrontendPayload{
		Name:                 frontendName,
		DefaultBackend:       d.Get("default_backend").(string),
		HttpConnectionMode:   d.Get("http_connection_mode").(string),
		MaxConn:              d.Get("maxconn").(int),
		Mode:                 d.Get("mode").(string),
		Backlog:              d.Get("backlog").(int),
		HttpKeepAliveTimeout: d.Get("http_keep_alive_timeout").(int),
		HttpRequestTimeout:   d.Get("http_request_timeout").(int),
		HttpLog:              d.Get("httplog").(bool),
		ErrorLogFormat:       d.Get("error_log_format").(string),
		LogFormat:            d.Get("log_format").(string),
		LogFormatSd:          d.Get("log_format_sd").(string),
		MonitorUri:           d.Get("monitor_uri").(string),
		TcpLog:               d.Get("tcplog").(bool),
		From:                 d.Get("from").(string),
	}

	// Check httpslog is set
	// if httpslog {
	// 	payload.HttpsLog = utils.BoolToStr(httpslog)
	// }

	// Check httpUseProxyHeader is set
	// if httpUseProxyHeader {
	// 	payload.HttpUseProxyHeader = utils.BoolToStr(httpUseProxyHeader)
	// }

	// Check if AcceptInvalidHttpRequest is set
	// if acceptInvalidHttpRequest {
	// 	payload.AcceptInvalidHttpRequest = utils.BoolToStr(acceptInvalidHttpRequest)
	// }

	// // Check if compression is set
	// if compressionItem != nil {
	// 	payload.Compression = Compression{
	// 		Algorithms: compressionAlgorithms,
	// 		Offload:    compressionOffload,
	// 		Types:      compressionTypes,
	// 	}
	// }

	// // Check if forwardfor is set
	// if forwardforItem != nil {
	// 	payload.Forwardfor = Forwardfor{
	// 		Enabled: enabledStr,
	// 		Except:  forwardforExcept,
	// 		Header:  forwardforHeader,
	// 		Ifnone:  forwardforIfnone,
	// 	}
	// }

	// Check if monitorfail is set
	if monitorfailItem != nil {
		payload.MonitorFail = MonitorFail{
			Cond:     monitorfailCond,
			CondTest: monitorfailcondTest,
		}
	}

	payloadJSON, err := utils.MarshalNonZeroFields(payload)

	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal generic payload",
				Detail:   fmt.Sprintf("Error while marshaling generic payload: %s", err.Error()),
			},
		}
	}

	aclItemUpdates, aclItemCreations, aclItemDeletions, err := utils.GetResourceswithIndexToBeUpdated(d, "acl")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal aclupdates payload",
				Detail:   fmt.Sprintf("Error while marshaling aclupdates payload: %s", err.Error()),
			},
		}
	}
	httprequestruleItemUpdates, httprequestruleItemCreations, httprequestruleItemDeletions, err := utils.GetResourceswithIndexToBeUpdated(d, "httprequestrule")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal aclupdates payload",
				Detail:   fmt.Sprintf("Error while marshaling httprequestruleupdates payload: %s", err.Error()),
			},
		}
	}
	httpresponseruleItemUpdates, httpresponseruleItemCreations, httpresponseruleItemDeletions, err := utils.GetResourceswithIndexToBeUpdated(d, "httpresponserule")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal aclupdates payload",
				Detail:   fmt.Sprintf("Error while marshaling httpresponseupdates payload: %s", err.Error()),
			},
		}
	}

	configMap := m.(map[string]interface{})
	frontendConfig := configMap["frontend"].(*ConfigFrontend)
	aclConfig := configMap["acl"].(*acl.ConfigAcl)
	httprequestrulesConfig := configMap["httprequestrule"].(*httprequestrule.ConfigHttpRequestRule)
	// tcprequestrulesConfig := configMap["tcprequestrule"].(*tcprequestrule.ConfigTcpRequestRules)
	httpresponserulesConfig := configMap["httpresponserule"].(*httpresponserule.ConfigHttpResponseRule)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		//acls
		resp_frontend, err := frontendConfig.UpdateFrontendConfiguration(frontendName, payloadJSON, transactionID)
		diags = utils.HandleHTTPResponse(resp_frontend, err, "Error during frontend configuration")

		if len(diags) > 0 {
			utils.PrintDiags(diags)
			return resp_frontend, err
		}

		for _, item := range aclItemUpdates {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(aclConfig, "UpdateAnAclConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Upated ACL configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range aclItemCreations {
			lastResp, err := utils.ProcessUpdateResourceswithoutIndex(aclConfig, "AddAnAclConfiguration", item, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Add ACL configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range aclItemDeletions {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(aclConfig, "DeleteAnAclConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Delete ACL configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httprequestruleItemUpdates {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(httprequestrulesConfig, "UpdateAHttpRequestRuleConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Upated Httprequestrule configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httprequestruleItemCreations {
			lastResp, err := utils.ProcessUpdateResourceswithoutIndex(httprequestrulesConfig, "AddAHttpRequestRuleConfiguration", item, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Add Httprequestrule configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httprequestruleItemDeletions {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(httprequestrulesConfig, "DeleteAHttpRequestRuleConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Delete Httprequestrule configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httpresponseruleItemUpdates {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(httpresponserulesConfig, "UpdateAHttpResponseRuleConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Upated ACL configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httpresponseruleItemCreations {
			lastResp, err := utils.ProcessUpdateResourceswithoutIndex(httpresponserulesConfig, "AddAHttpResponseRuleConfiguration", item, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Add ACL configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httpresponseruleItemDeletions {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(httpresponserulesConfig, "DeleteAHttpResponseRuleConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during Delete Httpresponserule configuration")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		return resp_frontend, err
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during frontend configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId(frontendName)
	return ResourceHaproxyFrontendRead(ctx, d, m)
}

func ResourceHaproxyFrontendDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	frontendName := d.Get("name").(string)

	configMap := m.(map[string]interface{})
	frontendConfig := configMap["frontend"].(*ConfigFrontend)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return frontendConfig.DeleteFrontendConfiguration(frontendName, transactionID)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during frontend configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId("")
	return diags
}
