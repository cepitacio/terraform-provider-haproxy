package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"terraform-provider-haproxy/internal/transaction"
	"terraform-provider-haproxy/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceHaproxyServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHaproxyServerCreate,
		ReadContext:   resourceHaproxyServerRead,
		UpdateContext: resourceHaproxyServerUpdate,
		DeleteContext: resourceHaproxyServerDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the server. It must be unique and cannot be changed.",
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The port of the server. Constraints: Min 1┃Max 65535",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The address of the server. Pattern: ^[^\\s]+$",
			},
			"parent_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the parent object",
			},
			"parent_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of the parent object. Allowed: backend|ring|peers",
			},
			"send_proxy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "To send a Proxy Protocol header to the backend server. Allowed: enabled|disabled",
			},
			"check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "To enable health check for the server. Allowed: enabled|disabled",
			},
			"check_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "To enable health check ssl if different port is used. Allowed: enabled|disabled",
			},
			"inter": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The inter value is the time interval in milliseconds between two consecutive health checks.",
			},
			"rise": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The rise value states that a server will be considered as operational after consecutive successful health checks.",
			},
			"fall": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The fall value states that a server will be considered as failed after consecutive unsuccessful health checks.",
			},
			"ssl": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Enables ssl",
			},
			"ssl_cafile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ssl certificate ca file. Pattern: ^[^\\s]+$",
			},
			"ssl_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ssl certificate. Pattern: ^[^\\s]+$",
			},
			"ssl_max_ver": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ssl max version. Allowed: SSLv3┃TLSv1.0┃TLSv1.1┃TLSv1.2┃TLSv1.3",
			},
			"ssl_min_ver": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ssl min version. Allowed: SSLv3┃TLSv1.0┃TLSv1.1┃TLSv1.2┃TLSv1.3",
			},
			"ssl_reuse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Reuse ssl existion connection. Allowed: enabled┃disabled",
			},
			"verify": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The certificate verification for backend servers. Allowed: none┃required",
			},
			"health_check_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The health check port of the server. Constraints: Min 1┃Max 65535",
			},
			"weight": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The weight of the server",
			},
			"ciphers": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ciphers to support",
			},
			"ciphersuites": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ciphersuites to support",
			},
			"force_sslv3": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State of SSLv3 protocol support for the SSL. Allowed: enabled┃disabled",
			},
			"force_tlsv10": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State of TLSv1.0 protocol support for the SSL. Allowed: enabled┃disabled",
			},
			"force_tlsv11": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State of TLSv1.1 protocol. Allowed: enabled┃disabled",
			},
			"force_tlsv12": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State of TLSv1.2 protocol. Allowed: enabled┃disabled",
			},
			"force_tlsv13": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State of TLSv1.3 protocol. Allowed: enabled┃disabled",
			},
		},
	}
}

func resourceHaproxyServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	serverName := d.Get("name").(string)
	sendProxy := d.Get("send_proxy").(bool)
	check := d.Get("check").(bool)
	checkSsl := d.Get("check_ssl").(bool)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	payload := ServerPayload{
		Name:              serverName,
		Address:           d.Get("address").(string),
		Port:              d.Get("port").(int),
		Inter:             d.Get("inter").(int),
		Rise:              d.Get("rise").(int),
		Fall:              d.Get("fall").(int),
		Ssl:               d.Get("ssl").(string),
		Ssl_cafile:        d.Get("ssl_cafile").(string),
		Ssl_certificate:   d.Get("ssl_certificate").(string),
		Ssl_max_ver:       d.Get("ssl_max_ver").(string),
		Ssl_min_ver:       d.Get("ssl_min_ver").(string),
		Ssl_reuse:         d.Get("ssl_reuse").(string),
		Verify:            d.Get("verify").(string),
		Health_check_port: d.Get("health_check_port").(int),
		Weight:            d.Get("weight").(int),
		Ciphersuites:      d.Get("ciphersuites").(string),
		Force_sslv3:       d.Get("force_sslv3").(string),
		Force_tlsv10:      d.Get("force_tlsv10").(string),
		Force_tlsv11:      d.Get("force_tlsv11").(string),
		Force_tlsv12:      d.Get("force_tlsv12").(string),
		Force_tlsv13:      d.Get("force_tlsv13").(string),
	}

	// Check sendProxy field
	if sendProxy {
		payload.SendProxy = utils.BoolToStr(sendProxy)
	}

	// Check check field
	if check {
		payload.Check = utils.BoolToStr(check)
	}

	// Check check field
	if checkSsl {
		payload.CheckSsl = utils.BoolToStr(checkSsl)
	}

	payloadJSON, err := utils.MarshalNonZeroFields(payload)
	if err != nil {
		return diag.Errorf("failed to marshal payload: %s", err)
	}

	configMap := m.(map[string]interface{})
	serverConfig := configMap["server"].(*ConfigServer)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return serverConfig.AddServerConfiguration(payloadJSON, transactionID, parentName, parentType)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during ACL configuration")
	if len(diags) > 0 {
		return diags
	}

	return resourceHaproxyServerRead(ctx, d, m)
}

func resourceHaproxyServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	serverName := d.Get("name").(string)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	configMap := m.(map[string]interface{})
	serverConfig := configMap["server"].(*ConfigServer)

	resp, err := serverConfig.GetAServerConfiguration(serverName, parentName, parentType)
	if err != nil {
		if strings.Contains(err.Error(), "missing object") || resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
	}
	diags = utils.HandleHTTPResponse(resp, err, "Error during server configuration")
	if len(diags) > 0 {
		return diags
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	defer resp.Body.Close()

	var serverWrapper ServerWrapper
	err = json.Unmarshal(bodyBytes, &serverWrapper)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing bind configuration: %w", err))
	}

	d.Set("address", serverWrapper.Data.Address)
	d.Set("port", serverWrapper.Data.Port)
	d.Set("send-proxy", serverWrapper.Data.SendProxy)
	d.Set("timeout", serverWrapper.Data.Timeout)
	d.Set("check", serverWrapper.Data.Check)
	d.Set("check-ssl", serverWrapper.Data.CheckSsl)
	d.Set("inter", serverWrapper.Data.Inter)
	d.Set("rise", serverWrapper.Data.Rise)
	d.Set("fall", serverWrapper.Data.Fall)
	d.Set("ssl", serverWrapper.Data.Ssl)
	d.Set("ssl_cafile", serverWrapper.Data.Ssl_cafile)
	d.Set("ssl_certificate", serverWrapper.Data.Ssl_certificate)
	d.Set("ssl_max_ver", serverWrapper.Data.Ssl_max_ver)
	d.Set("ssl_min_ver", serverWrapper.Data.Ssl_min_ver)
	d.Set("ssl_reuse", serverWrapper.Data.Ssl_reuse)
	d.Set("verify", serverWrapper.Data.Verify)
	d.Set("health_check_port", serverWrapper.Data.Health_check_port)
	d.Set("weight", serverWrapper.Data.Weight)
	d.Set("ciphersuites", serverWrapper.Data.Ciphersuites)
	d.Set("force_sslv3", serverWrapper.Data.Force_sslv3)
	d.Set("force_tlsv10", serverWrapper.Data.Force_tlsv10)
	d.Set("force_tlsv11", serverWrapper.Data.Force_tlsv11)
	d.Set("force_tlsv12", serverWrapper.Data.Force_tlsv12)
	d.Set("force_tlsv13", serverWrapper.Data.Force_tlsv13)
	d.SetId(serverName)
	return diags
}

func resourceHaproxyServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	serverName := d.Get("name").(string)
	sendProxy := d.Get("send_proxy").(bool)
	check := d.Get("check").(bool)
	checkSsl := d.Get("check_ssl").(bool)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	payload := ServerPayload{
		Name:              serverName,
		Address:           d.Get("address").(string),
		Port:              d.Get("port").(int),
		Inter:             d.Get("inter").(int),
		Rise:              d.Get("rise").(int),
		Fall:              d.Get("fall").(int),
		Ssl:               d.Get("ssl").(string),
		Ssl_cafile:        d.Get("ssl_cafile").(string),
		Ssl_certificate:   d.Get("ssl_certificate").(string),
		Ssl_max_ver:       d.Get("ssl_max_ver").(string),
		Ssl_min_ver:       d.Get("ssl_min_ver").(string),
		Ssl_reuse:         d.Get("ssl_reuse").(string),
		Verify:            d.Get("verify").(string),
		Health_check_port: d.Get("health_check_port").(int),
		Weight:            d.Get("weight").(int),
		Ciphersuites:      d.Get("ciphersuites").(string),
		Force_sslv3:       d.Get("force_sslv3").(string),
		Force_tlsv10:      d.Get("force_tlsv10").(string),
		Force_tlsv11:      d.Get("force_tlsv11").(string),
		Force_tlsv12:      d.Get("force_tlsv12").(string),
		Force_tlsv13:      d.Get("force_tlsv13").(string),
	}

	// Check sendProxy field
	if sendProxy {
		payload.SendProxy = utils.BoolToStr(sendProxy)
	}

	// Check check field
	if check {
		payload.Check = utils.BoolToStr(check)
	}

	// Check check_ssl field
	if checkSsl {
		payload.CheckSsl = utils.BoolToStr(checkSsl)
	}

	payloadJSON, err := utils.MarshalNonZeroFields(payload)

	if err != nil {
		return diag.Errorf("failed to marshal payload: %s", err)
	}

	configMap := m.(map[string]interface{})
	serverConfig := configMap["server"].(*ConfigServer)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return serverConfig.UpdateServerConfiguration(serverName, payloadJSON, transactionID, parentName, parentType)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during ACL configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId(serverName)
	return resourceHaproxyServerRead(ctx, d, m)
}

func resourceHaproxyServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	serverName := d.Get("name").(string)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	configMap := m.(map[string]interface{})
	serverConfig := configMap["server"].(*ConfigServer)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)
	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return serverConfig.DeleteServerConfiguration(serverName, transactionID, parentName, parentType)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during ACL configuration")
	if len(diags) > 0 {
		return diags
	}

	d.SetId("")
	return diags
}
