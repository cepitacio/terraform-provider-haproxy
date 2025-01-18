package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PrintDiags(diags diag.Diagnostics) {
	if len(diags) == 0 {
		fmt.Println("No diagnostics available")
	} else {
		for _, d := range diags {
			var severityStr string
			switch d.Severity {
			case diag.Error:
				severityStr = "Error"
			case diag.Warning:
				severityStr = "Warning"
			default:
				severityStr = "Unknown"
			}
			fmt.Printf("Severity: %s\n", severityStr)
			fmt.Printf("Summary: %s\n", d.Summary)
			fmt.Printf("Detail: %s\n", d.Detail)
		}
	}
}
