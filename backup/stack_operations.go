package haproxy

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// CRUD Operations for haproxy_stack resource
// =============================================================================

// createAllResources creates all resources in a single transaction
func (r *haproxyStackResource) createAllResources(ctx context.Context, plan *haproxyStackResourceModel) error {
	// Create payload for single transaction
	allResources := &AllResourcesPayload{
		Backend: r.backendManager.processBackendBlock(plan.Backend),
		Servers: []ServerResource{
			{
				ParentType: "backend",
				ParentName: plan.Backend.Name.ValueString(),
				Payload: &ServerPayload{
					Name:      plan.Server.Name.ValueString(),
					Address:   plan.Server.Address.ValueString(),
					Port:      plan.Server.Port.ValueInt64(),
					Check:     plan.Server.Check.ValueString(),
					Backup:    plan.Server.Backup.ValueString(),
					Maxconn:   plan.Server.Maxconn.ValueInt64(),
					Weight:    plan.Server.Weight.ValueInt64(),
					Rise:      plan.Server.Rise.ValueInt64(),
					Fall:      plan.Server.Fall.ValueInt64(),
					Inter:     plan.Server.Inter.ValueInt64(),
					Fastinter: plan.Server.Fastinter.ValueInt64(),
					Downinter: plan.Server.Downinter.ValueInt64(),
					Ssl:       plan.Server.Ssl.ValueString(),
					Verify:    plan.Server.Verify.ValueString(),
					Cookie:    plan.Server.Cookie.ValueString(),
					Disabled:  plan.Server.Disabled.ValueBool(),
				},
			},
		},
		Frontend: r.frontendManager.processFrontendBlock(plan.Frontend),
	}

	// Prepare Frontend ACLs for the transaction
	if plan.Frontend.Acls != nil && len(plan.Frontend.Acls) > 0 {
		sortedAcls := r.aclManager.processAclsBlock(plan.Frontend.Acls)
		nextIndex := int64(0)

		for _, acl := range sortedAcls {
			aclResource := ACLResource{
				ParentType: "frontend",
				ParentName: plan.Frontend.Name.ValueString(),
				Payload: &ACLPayload{
					AclName:   acl.AclName.ValueString(),
					Criterion: acl.Criterion.ValueString(),
					Value:     acl.Value.ValueString(),
					Index:     nextIndex,
				},
			}
			allResources.Acls = append(allResources.Acls, aclResource)
			nextIndex++
		}
	}

	// Prepare Backend ACLs for the transaction
	if plan.Backend.Acls != nil && len(plan.Backend.Acls) > 0 {
		sortedAcls := r.aclManager.processAclsBlock(plan.Backend.Acls)
		nextIndex := int64(0)

		for _, acl := range sortedAcls {
			aclResource := ACLResource{
				ParentType: "backend",
				ParentName: plan.Backend.Name.ValueString(),
				Payload: &ACLPayload{
					AclName:   acl.AclName.ValueString(),
					Criterion: acl.Criterion.ValueString(),
					Value:     acl.Value.ValueString(),
					Index:     nextIndex,
				},
			}
			allResources.Acls = append(allResources.Acls, aclResource)
			nextIndex++
		}
	}

	// Create all resources in single transaction
	return r.client.CreateAllResourcesInSingleTransaction(ctx, allResources)
}

// readAllResources reads all resources from HAProxy
func (r *haproxyStackResource) readAllResources(ctx context.Context, state *haproxyStackResourceModel) error {
	backendName := state.Backend.Name.ValueString()
	serverName := state.Server.Name.ValueString()
	frontendName := state.Frontend.Name.ValueString()

	// Store existing values to preserve them
	existingBackend := state.Backend
	existingServer := state.Server
	existingFrontend := state.Frontend

	// Reset the state completely to avoid drift
	*state = haproxyStackResourceModel{
		Name: types.StringValue(state.Name.ValueString()),
		Backend: &haproxyBackendModel{
			Name: types.StringValue(backendName),
		},
		Server: &haproxyServerModel{
			Name: types.StringValue(serverName),
		},
		Frontend: &haproxyFrontendModel{
			Name: types.StringValue(frontendName),
		},
	}

	// Read backend
	backend, err := r.client.ReadBackend(ctx, backendName)
	if err != nil {
		return fmt.Errorf("could not read backend: %w", err)
	}

	// Read servers
	servers, err := r.client.ReadServers(ctx, "backend", backendName)
	if err != nil {
		return fmt.Errorf("could not read servers: %w", err)
	}

	// Read frontend using FrontendManager
	frontendModel, err := r.frontendManager.ReadFrontend(ctx, frontendName, existingFrontend)
	if err != nil {
		return fmt.Errorf("could not read frontend: %w", err)
	}
	state.Frontend = frontendModel

	// Update state with actual HAProxy configuration
	if backend != nil {
		state.Backend.Mode = types.StringValue(backend.Mode)

		// Handle adv_check based on whether httpchk_params is present
		if len(existingBackend.HttpchkParams) > 0 && existingBackend.AdvCheck.IsNull() {
			state.Backend.AdvCheck = types.StringNull()
		} else if !existingBackend.AdvCheck.IsNull() && !existingBackend.AdvCheck.IsUnknown() {
			state.Backend.AdvCheck = existingBackend.AdvCheck
		} else if backend.AdvCheck != "" {
			state.Backend.AdvCheck = types.StringValue(backend.AdvCheck)
		} else {
			state.Backend.AdvCheck = types.StringNull()
		}

		// Set other fields if HAProxy returned them
		if backend.HttpConnectionMode != "" {
			state.Backend.HttpConnectionMode = types.StringValue(backend.HttpConnectionMode)
		}
		if backend.ServerTimeout != 0 {
			state.Backend.ServerTimeout = types.Int64Value(backend.ServerTimeout)
		}
		if backend.CheckTimeout != 0 {
			state.Backend.CheckTimeout = types.Int64Value(backend.CheckTimeout)
		}
		if backend.ConnectTimeout != 0 {
			state.Backend.ConnectTimeout = types.Int64Value(backend.ConnectTimeout)
		}
		if backend.QueueTimeout != 0 {
			state.Backend.QueueTimeout = types.Int64Value(backend.QueueTimeout)
		}
		if backend.TunnelTimeout != 0 {
			state.Backend.TunnelTimeout = types.Int64Value(backend.TunnelTimeout)
		}
		if backend.TarpitTimeout != 0 {
			state.Backend.TarpitTimeout = types.Int64Value(backend.TarpitTimeout)
		}
		if backend.CheckCache != "" {
			state.Backend.Checkcache = types.StringValue(backend.CheckCache)
		}
		if backend.Retries != 0 {
			state.Backend.Retries = types.Int64Value(backend.Retries)
		}

		// Handle nested blocks
		if backend.Balance != nil {
			balanceModel := haproxyBalanceModel{
				Algorithm: types.StringValue(backend.Balance.Algorithm),
			}
			if backend.Balance.UrlParam != "" {
				balanceModel.UrlParam = types.StringValue(backend.Balance.UrlParam)
			}
			state.Backend.Balance = []haproxyBalanceModel{balanceModel}
		} else if existingBackend.Balance != nil && len(existingBackend.Balance) > 0 {
			state.Backend.Balance = existingBackend.Balance
		}

		if backend.HttpchkParams != nil {
			httpchkModel := haproxyHttpchkParamsModel{}
			if backend.HttpchkParams.Method != "" {
				httpchkModel.Method = types.StringValue(backend.HttpchkParams.Method)
			}
			if backend.HttpchkParams.Uri != "" {
				httpchkModel.Uri = types.StringValue(backend.HttpchkParams.Uri)
			}
			if backend.HttpchkParams.Version != "" {
				httpchkModel.Version = types.StringValue(backend.HttpchkParams.Version)
			}
			state.Backend.HttpchkParams = []haproxyHttpchkParamsModel{httpchkModel}
		} else if existingBackend.HttpchkParams != nil && len(existingBackend.HttpchkParams) > 0 {
			state.Backend.HttpchkParams = existingBackend.HttpchkParams
		}

		if backend.Forwardfor != nil {
			forwardforModel := haproxyForwardforModel{}
			if backend.Forwardfor.Enabled != "" {
				forwardforModel.Enabled = types.StringValue(backend.Forwardfor.Enabled)
			}
			state.Backend.Forwardfor = []haproxyForwardforModel{forwardforModel}
		} else if existingBackend.Forwardfor != nil && len(existingBackend.Forwardfor) > 0 {
			state.Backend.Forwardfor = existingBackend.Forwardfor
		}
	}

	// Update server state
	if len(servers) > 0 {
		server := servers[0]
		if server.Name != "" {
			state.Server.Name = types.StringValue(server.Name)
		}
		if server.Address != "" {
			state.Server.Address = types.StringValue(server.Address)
		}
		if server.Port != 0 {
			state.Server.Port = types.Int64Value(server.Port)
		}
		if server.Check != "" {
			state.Server.Check = types.StringValue(server.Check)
		}
		if server.Backup != "" {
			state.Server.Backup = types.StringValue(server.Backup)
		}
		if server.Maxconn != 0 {
			state.Server.Maxconn = types.Int64Value(server.Maxconn)
		}
		if server.Weight != 0 {
			state.Server.Weight = types.Int64Value(server.Weight)
		}
		if server.Rise != 0 {
			state.Server.Rise = types.Int64Value(server.Rise)
		}
		if server.Fall != 0 {
			state.Server.Fall = types.Int64Value(server.Fall)
		}
		if server.Inter != 0 {
			state.Server.Inter = types.Int64Value(server.Inter)
		}
		if server.Fastinter != 0 {
			state.Server.Fastinter = types.Int64Value(server.Fastinter)
		}
		if server.Downinter != 0 {
			state.Server.Downinter = types.Int64Value(server.Downinter)
		}
		if server.Ssl != "" {
			state.Server.Ssl = types.StringValue(server.Ssl)
		}
		if server.Verify != "" {
			state.Server.Verify = types.StringValue(server.Verify)
		}
		if server.Cookie != "" {
			state.Server.Cookie = types.StringValue(server.Cookie)
		}
		state.Server.Disabled = types.BoolValue(server.Disabled)
	} else if existingServer != nil {
		state.Server = existingServer
	}

	// Handle Backend ACLs - preserve user's exact configuration from state
	if existingBackend.Acls != nil && len(existingBackend.Acls) > 0 {
		state.Backend.Acls = existingBackend.Acls
	}

	return nil
}

// updateAllResources updates all resources in a single transaction
func (r *haproxyStackResource) updateAllResources(ctx context.Context, plan *haproxyStackResourceModel) error {
	// Create payload for single transaction
	allResources := &AllResourcesPayload{
		Backend: r.backendManager.processBackendBlock(plan.Backend),
		Servers: []ServerResource{
			{
				ParentType: "backend",
				ParentName: plan.Backend.Name.ValueString(),
				Payload: &ServerPayload{
					Name:      plan.Server.Name.ValueString(),
					Address:   plan.Server.Address.ValueString(),
					Port:      plan.Server.Port.ValueInt64(),
					Check:     plan.Server.Check.ValueString(),
					Backup:    plan.Server.Backup.ValueString(),
					Maxconn:   plan.Server.Maxconn.ValueInt64(),
					Weight:    plan.Server.Weight.ValueInt64(),
					Rise:      plan.Server.Rise.ValueInt64(),
					Fall:      plan.Server.Fall.ValueInt64(),
					Inter:     plan.Server.Inter.ValueInt64(),
					Fastinter: plan.Server.Fastinter.ValueInt64(),
					Downinter: plan.Server.Downinter.ValueInt64(),
					Ssl:       plan.Server.Ssl.ValueString(),
					Verify:    plan.Server.Verify.ValueString(),
					Cookie:    plan.Server.Cookie.ValueString(),
					Disabled:  plan.Server.Disabled.ValueBool(),
				},
			},
		},
		Frontend: r.frontendManager.processFrontendBlock(plan.Frontend),
	}

	// Update backend
	err := r.client.UpdateBackend(ctx, plan.Backend.Name.ValueString(), allResources.Backend)
	if err != nil {
		return fmt.Errorf("could not update backend: %w", err)
	}

	// Update server
	err = r.client.UpdateServer(ctx, plan.Server.Name.ValueString(), "backend", plan.Backend.Name.ValueString(), allResources.Servers[0].Payload)
	if err != nil {
		return fmt.Errorf("could not update server: %w", err)
	}

	// Update frontend
	err = r.client.UpdateFrontend(ctx, plan.Frontend.Name.ValueString(), allResources.Frontend)
	if err != nil {
		return fmt.Errorf("could not update frontend: %w", err)
	}

	// Update Frontend ACLs
	if plan.Frontend.Acls != nil && len(plan.Frontend.Acls) > 0 {
		if err := r.aclManager.UpdateACLs(ctx, "frontend", plan.Frontend.Name.ValueString(), plan.Frontend.Acls); err != nil {
			return fmt.Errorf("could not update frontend ACLs: %w", err)
		}
	}

	// Update Backend ACLs
	if plan.Backend.Acls != nil && len(plan.Backend.Acls) > 0 {
		if err := r.aclManager.UpdateACLs(ctx, "backend", plan.Backend.Name.ValueString(), plan.Backend.Acls); err != nil {
			return fmt.Errorf("could not update backend ACLs: %w", err)
		}
	}

	return nil
}

// deleteAllResources deletes all resources in a single transaction
func (r *haproxyStackResource) deleteAllResources(ctx context.Context, state *haproxyStackResourceModel) error {
	log.Printf("Starting deletion of all resources in single transaction")

	// Prepare the complete resources payload for deletion
	resources := &AllResourcesPayload{
		Frontend: &FrontendPayload{
			Name: state.Frontend.Name.ValueString(),
		},
		Backend: &BackendPayload{
			Name: state.Backend.Name.ValueString(),
		},
	}

	// Add Frontend ACLs if they exist in the state
	if state.Frontend.Acls != nil && len(state.Frontend.Acls) > 0 {
		log.Printf("Including %d frontend ACLs in deletion transaction", len(state.Frontend.Acls))

		existingAcls, err := r.client.ReadACLs(ctx, "frontend", state.Frontend.Name.ValueString())
		if err == nil && len(existingAcls) > 0 {
			acls := make([]ACLResource, len(existingAcls))
			for i, acl := range existingAcls {
				acls[i] = ACLResource{
					ParentType: "frontend",
					ParentName: state.Frontend.Name.ValueString(),
					Payload:    &acl,
				}
			}
			resources.Acls = append(resources.Acls, acls...)
			log.Printf("Successfully mapped %d existing frontend ACLs for deletion", len(acls))
		} else {
			log.Printf("No existing frontend ACLs found in HAProxy (state had %d ACLs): %v", len(state.Frontend.Acls), err)
		}
	}

	// Add Backend ACLs if they exist in the state
	if state.Backend.Acls != nil && len(state.Backend.Acls) > 0 {
		log.Printf("Including %d backend ACLs in deletion transaction", len(state.Backend.Acls))

		existingAcls, err := r.client.ReadACLs(ctx, "backend", state.Backend.Name.ValueString())
		if err == nil && len(existingAcls) > 0 {
			acls := make([]ACLResource, len(existingAcls))
			for i, acl := range existingAcls {
				acls[i] = ACLResource{
					ParentType: "backend",
					ParentName: state.Backend.Name.ValueString(),
					Payload:    &acl,
				}
			}
			resources.Acls = append(resources.Acls, acls...)
			log.Printf("Successfully mapped %d existing backend ACLs for deletion", len(acls))
		} else {
			log.Printf("No existing backend ACLs found in HAProxy (state had %d ACLs): %v", len(state.Backend.Acls), err)
		}
	}

	// Delete everything in one transaction with retry logic
	err := r.client.DeleteAllResourcesInSingleTransaction(ctx, resources)
	if err != nil {
		return fmt.Errorf("could not delete resources in transaction: %w", err)
	}

	log.Printf("Successfully deleted all resources in single transaction")
	return nil
}
