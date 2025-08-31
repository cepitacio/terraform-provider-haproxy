package haproxy

import (
	"context"
	"fmt"

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
	if data.Backend == nil && data.Server == nil && data.Frontend == nil && len(data.Acls) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"At least one of backend, server, frontend, or acls must be specified",
		)
		return
	}

	// Validate server has parent backend
	if data.Server != nil && data.Backend == nil {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Server must have a parent backend specified",
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

	// Validate ACL indices are unique within their parent
	if len(data.Acls) > 0 {
		indices := make(map[int64]bool)
		for _, acl := range data.Acls {
			if indices[acl.Index.ValueInt64()] {
				resp.Diagnostics.AddError(
					"Invalid Configuration",
					fmt.Sprintf("Duplicate ACL index %d found", acl.Index.ValueInt64()),
				)
				return
			}
			indices[acl.Index.ValueInt64()] = true
		}
	}

	tflog.Info(ctx, "Resource configuration validation passed")
}
