package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func HandleHTTPResponse(resp *http.Response, err error, summary string) diag.Diagnostics {
	var diags diag.Diagnostics

	if err != nil {
		detail := sanitizeErrorMessage(err.Error())
		fmt.Printf("Error occurred during transaction: %s\n", detail)

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   detail,
		})
		return diags
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, readErr := io.ReadAll(resp.Body)
		detail := fmt.Sprintf("Request failed with status code %d", resp.StatusCode)

		if readErr == nil {
			bodyStr := string(body)
			detail = sanitizeErrorMessage(bodyStr)
			fmt.Printf("HTTP error response body: %s\n", detail)
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   detail,
		})
		return diags
	}

	return diags
}

func sanitizeErrorMessage(s string) string {
	lowered := strings.ToLower(s)

	if strings.Contains(lowered, "password") ||
		strings.Contains(lowered, "authorization") ||
		strings.Contains(lowered, "token") {
		return "An error occurred. Please check your credentials, username, password, and endpoint. Sensitive details have been hidden for security."
	}

	return s
}
