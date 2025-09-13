package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StackValidation handles all validation logic for the haproxy_stack resource
type StackValidation struct{}

// NewStackValidation creates a new StackValidation instance
func CreateStackValidation() *StackValidation {
	return &StackValidation{}
}

// ValidateResourceConfig validates the resource configuration
func (v *StackValidation) ValidateResourceConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data haproxyStackResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one resource is specified
	if data.Backend == nil && len(data.Servers) == 0 && data.Frontend == nil {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"At least one of backend, servers, or frontend must be specified",
		)
		return
	}

	// Validate servers have parent backend
	if len(data.Servers) > 0 && data.Backend == nil {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Servers must have a parent backend specified",
		)
		return
	}

	// ACLs are now validated within their parent frontend/backend blocks

	// ACL indices are now handled by array position, no validation needed

	tflog.Info(ctx, "Resource configuration validation passed")
}
