package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StackValidation handles all validation logic for the haproxy_stack resource
type StackValidation struct{}

// NewStackValidation creates a new StackValidation instance
func NewStackValidation() *StackValidation {
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
	if data.Backend == nil && len(data.Servers) == 0 && data.Frontend == nil && len(data.Acls) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"At least one of backend, servers, frontend, or acls must be specified",
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

	// Validate ACLs have parent resource
	if len(data.Acls) > 0 && data.Frontend == nil && data.Backend == nil {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"ACLs must have a parent frontend or backend specified",
		)
		return
	}

	// ACL indices are now handled by array position, no validation needed

	tflog.Info(ctx, "Resource configuration validation passed")
}
