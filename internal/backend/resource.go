package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"terraform-provider-haproxy/internal/common_services/httpcheck"
	"terraform-provider-haproxy/internal/transaction"
	"terraform-provider-haproxy/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceHaproxyBackend() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHaproxyBackendCreate,
		ReadContext:   resourceHaproxyBackendRead,
		UpdateContext: resourceHaproxyBackendUpdate,
		DeleteContext: resourceHaproxyBackendDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the backend. It must be unique and cannot be changed.",
			},
			"mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The mode of the backend. Allowed: http|tcp|log",
			},
			"adv_check": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The adv_check of the backend. Allowed: ssl-hello-chk|smtpchk|ldap-check|mysql-check|pgsql-check|tcp-check|redis-check|httpchk",
			},
			"http_connection_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The http_connection_mode of the backend. Allowed: http-keep-alive|httpclose|http-server-close",
			},
			"server_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The server_timeout of the backend.",
			},
			"check_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The check_timeout of the backend.",
			},
			"connect_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The connect_timeout of the backend.",
			},
			"queue_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The queue_timeout of the backend.",
			},
			"tunnel_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The tunnel_timeout of the backend.",
			},
			"tarpit_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The tarpit_timeout of the backend.",
			},
			"check_cache": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The check_cache of the backend.",
			},
			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The retries of the backend.",
			},
			"balance": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The balance of the backend.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"algorithm": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "roundrobin",
							Description: "The algorithm of the balance. Allowed: first|hash|hdr|leastconn|random|rdp-cookie|roundrobin|source|static-rr|uri|url_param",
						},
						"url_param": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The url param if algorith is url_param. Pattern: ^[^\\s]+$",
						},
						// "url_param_check_post": {
						// 	Type:        schema.TypeInt,
						// 	Optional:    true,
						// 	Default:     nil,
						// 	Description: "The url param check post if algorith is url_param.",
						// },
						// "url_param_max_wait": {
						// 	Type:        schema.TypeInt,
						// 	Default:     nil,
						// 	Optional:    true,
						// 	Description: "The url param max wait if algorith is url_param.",
						// },
					},
				},
			},
			"httpchk_params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The httpchk_params of the backend.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The method of the httpchk_params. Allowed: HEAD|PUT|POST|GET|TRACE|PATCH|DELETE|CONNECT|OPTIONS",
						},
						"uri": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The uri of the httpchk_params. Pattern: ^[^ ]*$",
						},
						"version": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The version of the httpchk_params.",
						},
					},
				},
			},
			"forwardfor": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The forwardfor of the backend.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The enabled of the forwardfor. Allowed: enabled",
						},
					},
				},
			},
			"httpcheck": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "httpchecks feature for backend",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The index for httpcheck",
						},
						"addr": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The addr for httpcheck. Pattern: ^[^\\s]+$",
						},
						"match": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The status code for httpcheck response. Allowed: status|rstatus|hdr|fhdr|string|rstring",
						},
						"pattern": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "The status code for httpcheck response",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The type for httpcheck. Allowed: comment|connect|disable-on-404|expect|send|send-state|set-var|set-var-fmt|unset-var",
						},
						"method": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The method for httpcheck. Allowed: HEAD|PUT|POST|GET|TRACE|PATCH|DELETE|CONNECT|OPTIONS",
						},
					},
				},
			},
		},
	}
}

func resourceHaproxyBackendCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	backendName := d.Get("name").(string)
	parentName := backendName
	parentType := "backend"

	var (
		algorithm string
		version   string
		uri       string
		method    string
		enabled   string
		url_param string
		// url_param_check_post int
		// url_param_max_wait   int
	)

	// Handle balance
	balanceItem := d.Get("balance").(*schema.Set).List()
	if len(balanceItem) > 0 {
		balanceMap := balanceItem[0].(map[string]interface{})
		algorithm = balanceMap["algorithm"].(string)

	}

	// Handle httpchk_params
	httpchkItem := d.Get("httpchk_params").(*schema.Set).List()
	if len(httpchkItem) > 0 {
		httpchkMap := httpchkItem[0].(map[string]interface{})
		version = httpchkMap["version"].(string)
		uri = httpchkMap["uri"].(string)
		method = httpchkMap["method"].(string)
	}

	// Handle forwardfor
	forwardforItem := d.Get("forwardfor").(*schema.Set).List()
	if len(forwardforItem) > 0 {
		forwardforMap := forwardforItem[0].(map[string]interface{})
		enabled = forwardforMap["enabled"].(string)
	}

	payload := BackendPayload{
		Name:               backendName,
		Mode:               d.Get("mode").(string),
		AdvCheck:           d.Get("adv_check").(string),
		HttpConnectionMode: d.Get("http_connection_mode").(string),
		ServerTimeout:      d.Get("server_timeout").(int),
		CheckTimeout:       d.Get("check_timeout").(int),
		ConnectTimeout:     d.Get("connect_timeout").(int),
		QueueTimeout:       d.Get("queue_timeout").(int),
		TunnelTimeout:      d.Get("tunnel_timeout").(int),
		TarpitTimeout:      d.Get("tarpit_timeout").(int),
		Retries:            d.Get("retries").(int),
		CheckCache:         d.Get("check_cache").(string),
	}

	// Check if Balance data is available
	if balanceItem != nil {
		payload.Balance = Balance{
			Algorithm: algorithm,
			UrlParam:  url_param,
			// UrlParamCheckPost: url_param_check_post,
			// UrlParamMaxWait:   url_param_max_wait,
		}
	}

	// Check if HttpchkParams data is available
	if httpchkItem != nil {
		payload.HttpchkParams = HttpchkParams{
			Method:  method,
			Uri:     uri,
			Version: version,
		}
	}

	// Check if Forwardfor data is available
	if forwardforItem != nil {
		payload.Forwardfor = ForwardFor{
			Enabled: enabled,
		}
	}

	// httpchecksItems := d.Get("httpcheck").(*schema.Set).List()
	// fmt.Println("Length of httpcheck:", httpchecksItems)

	payloadJSON, err := utils.MarshalNonZeroFields(payload)
	if err != nil {
		return diag.Errorf("failed to marshal payload: %s", err)
	}

	fmt.Println("payloadJSON", string(payloadJSON))
	httpcheckItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "httpcheck")

	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httpcheckItems payload",
				Detail:   fmt.Sprintf("Error while marshaling httpcheckItem payload: %s", err.Error()),
			},
		}
	}

	configMap := m.(map[string]interface{})
	backendConfig := configMap["backend"].(*ConfigBackend)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	httpcheckConfig := configMap["httpcheck"].(*httpcheck.ConfigHttpCheck)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		resp_backend, err := backendConfig.AddBackendConfiguration(payloadJSON, transactionID)
		diags = utils.HandleHTTPResponse(resp_backend, err, "Error during creating backend configuration")
		if len(diags) > 0 {
			utils.PrintDiags(diags)
			return resp_backend, err
		}

		for _, item := range httpcheckItems {
			httpcheck_payloadJSON, _ := json.Marshal(item)
			lastResp, err := httpcheckConfig.AddAHttpCheckConfiguration(httpcheck_payloadJSON, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, fmt.Sprintf("Error during creating httpcheck configuration for parent: %s", parentName))
			if len(diags) > 0 {
				return lastResp, err
			}
		}

		return resp_backend, err
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during backend create configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId(backendName)
	return resourceHaproxyBackendRead(ctx, d, m)
}

func resourceHaproxyBackendRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	backendName := d.Get("name").(string)
	parentName := backendName
	parentType := "backend"
	configMap := m.(map[string]interface{})
	backendConfig := configMap["backend"].(*ConfigBackend)

	// httpcheck read starts here
	httpcheckConfig := configMap["httpcheck"].(*httpcheck.ConfigHttpCheck)

	httpcheckItems, err := utils.FetchAndSortSchemaItemsByIndex(d.Get, "httpcheck")
	fmt.Println("Length of httpcheckItems:", len(httpcheckItems))
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httpcheckItems payload",
				Detail:   fmt.Sprintf("Error while marshaling httpcheckItem payload: %s", err.Error()),
			},
		}
	}

	updatedItems := make([]map[string]interface{}, 0, len(httpcheckItems))
	var lastResp *http.Response
	processed := make(map[string]map[string]bool)
	for range httpcheckItems {
		if _, exists := processed[parentName]; exists {
			if processed[backendName][parentType] {
				continue
			}
		} else {
			processed[parentName] = make(map[string]bool)
		}

		processed[parentName]["backend"] = true
		lastResp, err = httpcheckConfig.GetAllHttpCheckConfiguration(parentName, parentType)
		diags = utils.HandleHTTPResponse(lastResp, err, "Error during httpcheck configuration")
		if diags != nil {
			return diags
		}
		var httpcheckWrapper httpcheck.HttpCheckWrapper
		if err := json.NewDecoder(lastResp.Body).Decode(&httpcheckWrapper); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing httpcheck configuration: %w", err))
		}
		defer lastResp.Body.Close()

		for _, item := range httpcheckWrapper.Data {
			updatedhttpcheck := map[string]interface{}{
				"index":   item.Index,
				"match":   item.Match,
				"pattern": item.Pattern,
				"type":    item.Type,
				"method":  item.Method,
				"addr":    item.Addr,
			}
			updatedItems = append(updatedItems, updatedhttpcheck)
		}
	}

	fmt.Printf("All httpchecks items from API: %+v\n", updatedItems)
	d.Set("httpcheck", updatedItems)

	//read a backend
	resp, err := backendConfig.GetABackendConfiguration(backendName)

	if err != nil {
		if strings.Contains(err.Error(), "missing object") || resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
	}

	diags = utils.HandleHTTPResponse(resp, err, "Error during Backend configuration")

	if len(diags) > 0 {
		return diags
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	defer resp.Body.Close()

	var backendWrapper BackendWrapper
	fmt.Printf("Raw JSON Response: %s\n", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &backendWrapper)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing backend configuration: %w", err))
	}

	baLance := backendWrapper.Data.Balance

	if baLance == (Balance{}) {
		d.Set("monitor_fail", nil)
	} else {
		d.Set("balance", []interface{}{
			map[string]interface{}{
				"algorithm": baLance.Algorithm,
				"url_param": baLance.UrlParam,
			},
		})
	}

	httpchkParams := backendWrapper.Data.HttpchkParams

	if httpchkParams == (HttpchkParams{}) {
		d.Set("httpchk_params", nil)
	} else {
		d.Set("httpchk_params", []interface{}{
			map[string]interface{}{
				"method":  httpchkParams.Method,
				"uri":     httpchkParams.Uri,
				"version": httpchkParams.Version,
			},
		})
	}

	forwardFor := backendWrapper.Data.Forwardfor

	if forwardFor == (ForwardFor{}) {
		d.Set("forwardfor", nil)
	} else {
		d.Set("forwardfor", []interface{}{
			map[string]interface{}{
				"enabled": forwardFor.Enabled,
			},
		})
	}

	d.Set("name", backendName)
	d.Set("mode", backendWrapper.Data.Mode)
	d.Set("adv_check", backendWrapper.Data.AdvCheck)
	d.Set("http_connection_mode", backendWrapper.Data.HttpConnectionMode)
	d.Set("server_timeout", backendWrapper.Data.ServerTimeout)
	d.Set("check_timeout", backendWrapper.Data.CheckTimeout)
	d.Set("connect_timeout", backendWrapper.Data.ConnectTimeout)
	d.Set("queue_timeout", backendWrapper.Data.QueueTimeout)
	d.Set("tunnel_timeout", backendWrapper.Data.TunnelTimeout)
	d.Set("tarpit_timeout", backendWrapper.Data.TarpitTimeout)
	d.Set("check_cache", backendWrapper.Data.CheckCache)
	d.Set("retries", backendWrapper.Data.Retries)

	return diags
}

func resourceHaproxyBackendUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	backendName := d.Get("name").(string)
	parentName := backendName
	parentType := "backend"

	var (
		algorithm string
		version   string
		uri       string
		method    string
		enabled   string
		url_param string
	// 	url_param_check_post int
	// 	url_param_max_wait   int
	)

	// Read values for balance
	balanceItem := utils.GetFirstItemValue(d.Get, "balance")
	if balanceItem != nil {
		balance := d.Get("balance").(*schema.Set).List()
		algorithm = balance[0].(map[string]interface{})["algorithm"].(string)
		url_param = balance[0].(map[string]interface{})["url_param"].(string)
		// url_param_check_post = balance[0].(map[string]interface{})["url_param_check_post"].(int)
		// url_param_max_wait = balance[0].(map[string]interface{})["url_param_max_wait"].(int)
	}

	//Read values for httpchk_params
	httpchkItem := utils.GetFirstItemValue(d.Get, "httpchk_params")
	if httpchkItem != nil {
		httpchk_params := d.Get("httpchk_params").(*schema.Set).List()
		version = httpchk_params[0].(map[string]interface{})["version"].(string)
		uri = httpchk_params[0].(map[string]interface{})["uri"].(string)
		method = httpchk_params[0].(map[string]interface{})["method"].(string)
	}

	//Read values for forwardfor
	forwardforItem := utils.GetFirstItemValue(d.Get, "forwardfor")
	if forwardforItem != nil {
		forwardfor := d.Get("forwardfor").(*schema.Set).List()
		enabled = forwardfor[0].(map[string]interface{})["enabled"].(string)
	}

	payload := BackendPayload{
		Name:               backendName,
		Mode:               d.Get("mode").(string),
		AdvCheck:           d.Get("adv_check").(string),
		HttpConnectionMode: d.Get("http_connection_mode").(string),
		ServerTimeout:      d.Get("server_timeout").(int),
		CheckTimeout:       d.Get("check_timeout").(int),
		ConnectTimeout:     d.Get("connect_timeout").(int),
		QueueTimeout:       d.Get("queue_timeout").(int),
		TunnelTimeout:      d.Get("tunnel_timeout").(int),
		TarpitTimeout:      d.Get("tarpit_timeout").(int),
		Retries:            d.Get("retries").(int),
		CheckCache:         d.Get("check_cache").(string),
	}

	// Check if Balance data is available
	if balanceItem != nil {
		payload.Balance = Balance{
			Algorithm: algorithm,
			UrlParam:  url_param,
			// UrlParamCheckPost: url_param_check_post,
			// UrlParamMaxWait:   url_param_max_wait,
		}
	}

	// // Check if HttpchkParams data is available
	if httpchkItem != nil {
		payload.HttpchkParams = HttpchkParams{
			Method:  method,
			Uri:     uri,
			Version: version,
		}
	}

	// Check if Forwardfor data is available
	if forwardforItem != nil {
		payload.Forwardfor = ForwardFor{
			Enabled: enabled,
		}
	}

	payloadJSON, err := utils.MarshalNonZeroFields(payload)

	if err != nil {
		return diag.Errorf("failed to fetch and sort schema items: %s", err)
	}

	httpcheckItemUpdates, httpcheckItemCreations, httpcheckItemDeletions, err := utils.GetResourceswithIndexToBeUpdated(d, "httpcheck")
	if err != nil {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to marshal httpcheckupdates payload",
				Detail:   fmt.Sprintf("Error while marshaling httpcheckupdates payload: %s", err.Error()),
			},
		}
	}

	configMap := m.(map[string]interface{})
	backendConfig := configMap["backend"].(*ConfigBackend)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resourceConfig := configMap["httpcheck"].(*httpcheck.ConfigHttpCheck)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		resp_backend, err := backendConfig.UpdateBackendConfiguration(backendName, payloadJSON, transactionID)
		if len(diags) > 0 {
			utils.PrintDiags(diags)
			return resp_backend, err
		}

		for _, item := range httpcheckItemUpdates {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(resourceConfig, "UpdateAHttpCheckConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during update. httpcheck configuration update failed.")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httpcheckItemCreations {
			lastResp, err := utils.ProcessUpdateResourceswithoutIndex(resourceConfig, "AddAHttpCheckConfiguration", item, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during update. httpcheck configuration updated failed.")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}
		for _, item := range httpcheckItemDeletions {
			lastResp, err := utils.ProcessUpdateResourceswithIndex(resourceConfig, "DeleteAHttpCheckConfiguration", []map[string]interface{}{item}, transactionID, parentName, parentType)
			diags = utils.HandleHTTPResponse(lastResp, err, "Error during update. httpcheck configuration delete failed.")
			if len(diags) > 0 {
				utils.PrintDiags(diags)
				return lastResp, err
			}
		}

		return resp_backend, err
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during updating httpcheck configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId(backendName)
	return resourceHaproxyBackendRead(ctx, d, m)
}

func resourceHaproxyBackendDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	backendName := d.Get("name").(string)

	configMap := m.(map[string]interface{})
	backendConfig := configMap["backend"].(*ConfigBackend)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return backendConfig.DeleteBackendConfiguration(backendName, transactionID)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during deleting httpcheck configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId("")
	return nil
}
