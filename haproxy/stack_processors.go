package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// StackProcessors handles all data processing logic for the haproxy_stack resource
type StackProcessors struct{}

// NewStackProcessors creates a new StackProcessors instance
func NewStackProcessors() *StackProcessors {
	return &StackProcessors{}
}

// ProcessCreateRequest processes the create request data
func (p *StackProcessors) ProcessCreateRequest(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) (*haproxyStackResourceModel, error) {
	var data haproxyStackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to get plan data")
	}

	// Process and validate the data
	if err := p.processStackData(ctx, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// ProcessReadRequest processes the read request data
func (p *StackProcessors) ProcessReadRequest(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) (*haproxyStackResourceModel, error) {
	var data haproxyStackResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to get state data")
	}

	// Process and validate the data
	if err := p.processStackData(ctx, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// ProcessUpdateRequest processes the update request data
func (p *StackProcessors) ProcessUpdateRequest(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) (*haproxyStackResourceModel, error) {
	var plan, state haproxyStackResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to get plan data")
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to get state data")
	}

	// Process and validate the plan data
	if err := p.processStackData(ctx, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// ProcessDeleteRequest processes the delete request data
func (p *StackProcessors) ProcessDeleteRequest(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) (*haproxyStackResourceModel, error) {
	var data haproxyStackResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to get state data")
	}

	// Process and validate the data
	if err := p.processStackData(ctx, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// processStackData processes and validates the stack data
func (p *StackProcessors) processStackData(ctx context.Context, data *haproxyStackResourceModel) error {
	// Ensure required fields are set
	if data.Name.IsNull() || data.Name.IsUnknown() {
		return fmt.Errorf("stack name is required")
	}

	// Process backend data if present
	if data.Backend != nil {
		if err := p.processBackendData(ctx, data.Backend); err != nil {
			return fmt.Errorf("failed to process backend data: %w", err)
		}
	}

	// Process servers data if present
	if len(data.Servers) > 0 {
		if err := p.processServersData(ctx, data.Servers); err != nil {
			return fmt.Errorf("failed to process servers data: %w", err)
		}
	}

	// Process frontend data if present
	if data.Frontend != nil {
		if err := p.processFrontendData(ctx, data.Frontend); err != nil {
			return fmt.Errorf("failed to process frontend data: %w", err)
		}
	}

	// Process ACLs data if present
	if len(data.Acls) > 0 {
		if err := p.processACLsData(ctx, data.Acls); err != nil {
			return fmt.Errorf("failed to process ACLs data: %w", err)
		}
	}

	return nil
}

// processBackendData processes backend-specific data
func (p *StackProcessors) processBackendData(ctx context.Context, backend *haproxyBackendModel) error {
	if backend.Name.IsNull() || backend.Name.IsUnknown() {
		return fmt.Errorf("backend name is required")
	}

	if backend.Mode.IsNull() || backend.Mode.IsUnknown() {
		return fmt.Errorf("backend mode is required")
	}

	// Validate mode values
	mode := backend.Mode.ValueString()
	if mode != "http" && mode != "tcp" {
		return fmt.Errorf("backend mode must be 'http' or 'tcp', got: %s", mode)
	}

	return nil
}

// processServerData processes server-specific data
func (p *StackProcessors) processServerData(ctx context.Context, server *haproxyServerModel) error {
	// Note: server name is now the map key, not a field

	if server.Address.IsNull() || server.Address.IsUnknown() {
		return fmt.Errorf("server address is required")
	}

	if server.Port.IsNull() || server.Port.IsUnknown() {
		return fmt.Errorf("server port is required")
	}

	// Validate port range
	port := server.Port.ValueInt64()
	if port < 1 || port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got: %d", port)
	}

	return nil
}

// processServersData processes multiple servers data
func (p *StackProcessors) processServersData(ctx context.Context, servers map[string]haproxyServerModel) error {
	for serverName, server := range servers {
		if err := p.processServerData(ctx, &server); err != nil {
			return fmt.Errorf("server[%s]: %w", serverName, err)
		}
	}
	return nil
}

// processFrontendData processes frontend-specific data
func (p *StackProcessors) processFrontendData(ctx context.Context, frontend *haproxyFrontendModel) error {
	if frontend.Name.IsNull() || frontend.Name.IsUnknown() {
		return fmt.Errorf("frontend name is required")
	}

	if frontend.Mode.IsNull() || frontend.Mode.IsUnknown() {
		return fmt.Errorf("frontend mode is required")
	}

	// Validate mode values
	mode := frontend.Mode.ValueString()
	if mode != "http" && mode != "tcp" {
		return fmt.Errorf("frontend mode must be 'http' or 'tcp', got: %s", mode)
	}

	return nil
}

// processACLsData processes ACLs-specific data
func (p *StackProcessors) processACLsData(ctx context.Context, acls []haproxyAclModel) error {
	for i, acl := range acls {
		if acl.AclName.IsNull() || acl.AclName.IsUnknown() {
			return fmt.Errorf("ACL %d: acl_name is required", i)
		}

		if acl.Criterion.IsNull() || acl.Criterion.IsUnknown() {
			return fmt.Errorf("ACL %d: criterion is required", i)
		}

		if acl.Value.IsNull() || acl.Value.IsUnknown() {
			return fmt.Errorf("ACL %d: value is required", i)
		}

		if acl.Index.IsNull() || acl.Index.IsUnknown() {
			return fmt.Errorf("ACL %d: index is required", i)
		}

		// Validate index is non-negative
		if acl.Index.ValueInt64() < 0 {
			return fmt.Errorf("ACL %d: index must be non-negative, got: %d", i, acl.Index.ValueInt64())
		}
	}

	return nil
}
