package bind

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

func ResourceHaproxyBind() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHaproxyBindCreate,
		ReadContext:   resourceHaproxyBindRead,
		UpdateContext: resourceHaproxyBindUpdate,
		DeleteContext: resourceHaproxyBindDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the bind. It must be unique and cannot be changed.",
			},
			"parent_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the parent object",
			},
			"parent_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of the parent object. Allowed: frontend|log_forward|peers",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The port of the bind",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The address of the bind"},
			"transparent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable transparent binding",
			},
			"mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "http, tcp",
			},
			"maxconn": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The max connections of the bind",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The user of the bind",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The group of the bind",
			},
			"force_sslv3": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of SSLv3 protocol support for the SSL",
			},
			"force_tlsv10": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of TLSv1.0 protocol support for the SSL",
			},
			"force_tlsv11": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of TLSv1.1 protocol",
			},
			"force_tlsv12": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of TLSv1.2 protocol",
			},
			"force_tlsv13": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of TLSv1.3 protocol",
			},
			"ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "State of SSL",
			},
			"ssl_cafile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ssl CA file. Pattern: ^[^\\s]+$",
			},
			"ssl_max_ver": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ssl max version to support. Allowed: SSLv3|TLSv1.0|TLSv1.1|TLSv1.2|TLSv1.3",
			},
			"ssl_min_ver": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "ssl min version to support. Allowed: SSLv3|TLSv1.0|TLSv1.1|TLSv1.2|TLSv1.3",
			},
			"ssl_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Path of SSL certificate. Pattern: ^[^\\s]+$",
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
		},
	}
}

func resourceHaproxyBindCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	bindName := d.Get("name").(string)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	payload := BindPayload{
		Name:           bindName,
		Address:        d.Get("address").(string),
		Port:           d.Get("port").(int),
		Maxconn:        d.Get("maxconn").(int),
		User:           d.Get("user").(string),
		Group:          d.Get("group").(string),
		Mode:           d.Get("mode").(string),
		ForceSslv3:     d.Get("force_sslv3").(bool),
		ForceTlsv10:    d.Get("force_tlsv10").(bool),
		ForceTlsv11:    d.Get("force_tlsv11").(bool),
		ForceTlsv12:    d.Get("force_tlsv12").(bool),
		ForceTlsv13:    d.Get("force_tlsv13").(bool),
		Ssl:            d.Get("ssl").(bool),
		SslCafile:      d.Get("ssl_cafile").(string),
		SslMaxVer:      d.Get("ssl_max_ver").(string),
		SslMinVer:      d.Get("ssl_min_ver").(string),
		SslCertificate: d.Get("ssl_certificate").(string),
		Ciphers:        d.Get("ciphers").(string),
		CipherSuites:   d.Get("ciphersuites").(string),
		Transparent:    d.Get("transparent").(bool),
	}

	payloadJSON, err := utils.MarshalNonZeroFields(payload)
	if err != nil {
		return diag.Errorf("failed to fetch and sort schema items: %s", err)
	}

	configMap := m.(map[string]interface{})
	BindConfig := configMap["bind"].(*ConfigBind)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)

	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return BindConfig.AddBindConfiguration(payloadJSON, transactionID, parentName, parentType)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during add bind configuration")
	if len(diags) > 0 {
		utils.PrintDiags(diags)
		return diags
	}

	d.SetId(bindName)
	return resourceHaproxyBindRead(ctx, d, m)
}

func resourceHaproxyBindRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	bindName := d.Get("name").(string)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	configMap := m.(map[string]interface{})
	BindConfig := configMap["bind"].(*ConfigBind)

	resp, err := BindConfig.GetABindConfiguration(bindName, parentName, parentType)

	if err != nil {
		if strings.Contains(err.Error(), "missing object") || resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
	}

	diags = utils.HandleHTTPResponse(resp, err, "Error during bind configuration read")
	if len(diags) > 0 {
		utils.PrintDiags(diags)
		return diags
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading response body: %w", err))
	}

	defer resp.Body.Close()

	var bindWrapper BindWrapper
	err = json.Unmarshal(bodyBytes, &bindWrapper)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing bind configuration: %w", err))
	}

	d.Set("address", bindWrapper.Data.Address)
	d.Set("port", bindWrapper.Data.Port)
	d.Set("maxconn", bindWrapper.Data.Maxconn)
	d.Set("user", bindWrapper.Data.User)
	d.Set("group", bindWrapper.Data.Group)
	d.Set("mode", bindWrapper.Data.Mode)
	d.Set("force_sslv3", bindWrapper.Data.ForceSslv3)
	d.Set("force_tlsv10", bindWrapper.Data.ForceTlsv10)
	d.Set("force_tlsv11", bindWrapper.Data.ForceTlsv11)
	d.Set("force_tlsv12", bindWrapper.Data.ForceTlsv12)
	d.Set("force_tlsv13", bindWrapper.Data.ForceTlsv13)
	d.Set("ssl", bindWrapper.Data.Ssl)
	d.Set("ssl_cafile", bindWrapper.Data.SslCafile)
	d.Set("ssl_max_ver", bindWrapper.Data.SslMaxVer)
	d.Set("ssl_min_ver", bindWrapper.Data.SslMinVer)
	d.Set("ssl_certificate", bindWrapper.Data.SslCertificate)
	d.Set("ciphers", bindWrapper.Data.Ciphers)
	d.Set("ciphersuites", bindWrapper.Data.CipherSuites)
	d.Set("transparent", bindWrapper.Data.Transparent)
	d.SetId(bindName)

	return diags
}

func resourceHaproxyBindUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	bindName := d.Get("name").(string)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	payload := BindPayload{
		Name:           bindName,
		Address:        d.Get("address").(string),
		Port:           d.Get("port").(int),
		Maxconn:        d.Get("maxconn").(int),
		User:           d.Get("user").(string),
		Group:          d.Get("group").(string),
		Mode:           d.Get("mode").(string),
		ForceSslv3:     d.Get("force_sslv3").(bool),
		ForceTlsv10:    d.Get("force_tlsv10").(bool),
		ForceTlsv11:    d.Get("force_tlsv11").(bool),
		ForceTlsv12:    d.Get("force_tlsv12").(bool),
		ForceTlsv13:    d.Get("force_tlsv13").(bool),
		Ssl:            d.Get("ssl").(bool),
		SslCafile:      d.Get("ssl_cafile").(string),
		SslMaxVer:      d.Get("ssl_max_ver").(string),
		SslMinVer:      d.Get("ssl_min_ver").(string),
		SslCertificate: d.Get("ssl_certificate").(string),
		Ciphers:        d.Get("ciphers").(string),
		CipherSuites:   d.Get("ciphersuites").(string),
		Transparent:    d.Get("transparent").(bool),
	}

	payloadJSON, err := utils.MarshalNonZeroFields(payload)
	if err != nil {
		return diag.Errorf("failed to marshal payload: %s", err)
	}

	configMap := m.(map[string]interface{})
	BindConfig := configMap["bind"].(*ConfigBind)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)

	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return BindConfig.UpdateBindConfiguration(bindName, payloadJSON, transactionID, parentName, parentType)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during add bind configuration")
	if len(diags) > 0 {
		utils.PrintDiags(diags)
		return diags
	}

	d.SetId(bindName)
	return resourceHaproxyBindRead(ctx, d, m)
}

func resourceHaproxyBindDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	bindName := d.Get("name").(string)
	parentName := d.Get("parent_name").(string)
	parentType := d.Get("parent_type").(string)

	configMap := m.(map[string]interface{})
	BindConfig := configMap["bind"].(*ConfigBind)
	tranConfig := configMap["transaction"].(*transaction.ConfigTransaction)

	resp, err := tranConfig.Transaction(func(transactionID string) (*http.Response, error) {
		return BindConfig.DeleteBindConfiguration(bindName, transactionID, parentName, parentType)
	})

	diags = utils.HandleHTTPResponse(resp, err, "Error during add bind configuration")
	if len(diags) > 0 {
		utils.PrintDiags(diags)
		return diags
	}

	d.SetId("")
	return nil
}
