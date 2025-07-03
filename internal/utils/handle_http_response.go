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
		sanitizedError := sanitizeString(err.Error())

		// Log internally — redacted if needed
		fmt.Printf("Error occurred during transaction: %s\n", sanitizedError)

		// Show generic error to user
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   "Unexpected error occurred during request. Check your connection, credentials, or endpoint.",
		})
		return diags
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, readErr := io.ReadAll(resp.Body)
		sanitizedBody := "unable to retrieve error details"

		if readErr == nil {
			// Log internal body — redacted if needed
			fmt.Println("HTTP error response body:", sanitizeString(string(body)))

			// If it looks like an auth failure, give user a useful message
			if resp.StatusCode == 401 || strings.Contains(strings.ToLower(string(body)), "unauthorized") {
				sanitizedBody = "Authentication failed: check username or password"
			} else {
				// Otherwise show status with no internal details
				sanitizedBody = fmt.Sprintf("Request failed with status code %d", resp.StatusCode)
			}
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   sanitizedBody,
		})
		return diags
	}

	return diags
}

func sanitizeString(s string) string {
	lowered := strings.ToLower(s)

	if strings.Contains(lowered, "password") || strings.Contains(lowered, "authorization") || strings.Contains(lowered, "token") {
		return "[REDACTED: potential sensitive content]"
	}

	return s
}
