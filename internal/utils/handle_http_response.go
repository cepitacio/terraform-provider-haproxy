package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func HandleHTTPResponse(resp *http.Response, err error, summary string) diag.Diagnostics {
	var diags diag.Diagnostics
	if err != nil {
		fmt.Printf("Error occurred during transaction: %v\n", err)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   fmt.Sprintf("Error: %s", err.Error()),
		})
		return diags
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		fmt.Println("Response is not in the success range")
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			fmt.Println("Error reading response body:", readErr)
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  summary,
			Detail:   fmt.Sprintf("HTTP Error: %d, Response Body: %s", resp.StatusCode, string(body)),
		})
		return diags
	}

	return diags
}
