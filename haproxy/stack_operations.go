package haproxy

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// StackOperations handles all CRUD operations for the haproxy_stack resource
type StackOperations struct {
	client                 *HAProxyClient
	aclManager             *ACLManager
	frontendManager        *FrontendManager
	backendManager         *BackendManager
	httpRequestRuleManager *HttpRequestRuleManager
	bindManager            *BindManager
}

// NewStackOperations creates a new StackOperations instance
func NewStackOperations(client *HAProxyClient, aclManager *ACLManager, frontendManager *FrontendManager, backendManager *BackendManager, httpRequestRuleManager *HttpRequestRuleManager, bindManager *BindManager) *StackOperations {
	return &StackOperations{
		client:                 client,
		aclManager:             aclManager,
		backendManager:         backendManager,
		frontendManager:        frontendManager,
		httpRequestRuleManager: httpRequestRuleManager,
		bindManager:            bindManager,
	}
}

// Create performs the create operation for the haproxy_stack resource
func (o *StackOperations) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Creating HAProxy stack")

	// Begin a single transaction for all resources
	tflog.Info(ctx, "Beginning single transaction for all resources")
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return err
	}
	tflog.Info(ctx, "Transaction created", map[string]interface{}{"transaction_id": transactionID})
	defer func() {
		// Rollback transaction if we encounter an error
		if err != nil {
			tflog.Error(ctx, "Rolling back transaction due to error", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
			if rollbackErr := o.client.RollbackTransaction(transactionID); rollbackErr != nil {
				tflog.Error(ctx, "Failed to rollback transaction", map[string]interface{}{"error": rollbackErr.Error()})
			}
		}
	}()

	// Create backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Creating backend in transaction", map[string]interface{}{"transaction_id": transactionID})
		if err := o.backendManager.CreateBackendInTransaction(ctx, transactionID, data.Backend); err != nil {
			resp.Diagnostics.AddError("Error creating backend", err.Error())
			return err
		}
		tflog.Info(ctx, "Backend created successfully in transaction", map[string]interface{}{"transaction_id": transactionID})
	}

	// Create servers if specified
	if len(data.Servers) > 0 && data.Backend != nil {
		for _, server := range data.Servers {
			if err := o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), &ServerPayload{
				Name:    server.Name.ValueString(),
				Address: server.Address.ValueString(),
				Port:    int64(server.Port.ValueInt64()),
			}); err != nil {
				resp.Diagnostics.AddError("Error creating server", err.Error())
				return err
			}
		}
	}

	// Create frontend if specified
	if data.Frontend != nil {
		if err := o.frontendManager.CreateFrontendInTransaction(ctx, transactionID, data.Frontend); err != nil {
			resp.Diagnostics.AddError("Error creating frontend", err.Error())
			return err
		}
	}

	// Create binds for frontend if specified
	if data.Frontend != nil && data.Frontend.Bind != nil && len(data.Frontend.Bind) > 0 {
		if err := o.bindManager.CreateBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Bind); err != nil {
			resp.Diagnostics.AddError("Error creating binds", err.Error())
			return err
		}
	}

	// Note: ACLs are now handled within frontend/backend creation, not here

	// Create HTTP Request Rules AFTER ACLs (so they can reference existing ACLs)
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		if err := o.httpRequestRuleManager.CreateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
			resp.Diagnostics.AddError("Error creating HTTP request rules", err.Error())
			return err
		}
	}

	// Commit the transaction
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err := o.client.CommitTransaction(transactionID); err != nil {
		tflog.Error(ctx, "Failed to commit transaction", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
		return err
	}
	tflog.Info(ctx, "Transaction committed successfully", map[string]interface{}{"transaction_id": transactionID})

	tflog.Info(ctx, "HAProxy stack created successfully")
	return nil
}

// Read performs the read operation for the haproxy_stack resource
func (o *StackOperations) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Reading HAProxy stack")

	// Read backend if specified
	if data.Backend != nil {
		_, err := o.backendManager.ReadBackend(ctx, data.Backend.Name.ValueString(), data.Backend)
		if err != nil {
			resp.Diagnostics.AddError("Error reading backend", err.Error())
			return err
		}
	}

	// Read servers if specified
	if len(data.Servers) > 0 && data.Backend != nil {
		// Server reading would need to be implemented
		tflog.Info(ctx, "Servers reading not yet implemented")
	}

	// Read frontend if specified
	if data.Frontend != nil {
		_, err := o.frontendManager.ReadFrontend(ctx, data.Frontend.Name.ValueString(), data.Frontend)
		if err != nil {
			resp.Diagnostics.AddError("Error reading frontend", err.Error())
			return err
		}
	}

	// Read binds for frontend if specified
	if data.Frontend != nil {
		binds, err := o.bindManager.ReadBinds(ctx, "frontend", data.Frontend.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading binds", err.Error())
			return err
		}

		// Get API version to determine which fields to map
		apiVersion := o.client.GetAPIVersion()
		// Create a map of binds by name for easy lookup
		bindMap := make(map[string]BindPayload)
		for _, bind := range binds {
			bindMap[bind.Name] = bind
		}

		// Debug: Log the bind map
		log.Printf("DEBUG: Bind map created with %d binds:", len(bindMap))
		for name, bind := range bindMap {
			log.Printf("DEBUG: Bind '%s': %+v", name, bind)
		}

		// Convert binds to model format in the same order as configuration
		// Store the original configuration binds to preserve order
		originalConfigBinds := data.Frontend.Bind
		data.Frontend.Bind = make([]haproxyBindModel, len(originalConfigBinds))
		log.Printf("DEBUG: Processing %d configuration binds:", len(originalConfigBinds))
		for i, configBind := range originalConfigBinds {
			bindName := configBind.Name.ValueString()
			log.Printf("DEBUG: Looking for bind '%s' in bind map", bindName)
			if bind, exists := bindMap[bindName]; exists {
				log.Printf("DEBUG: Found bind '%s' in HAProxy, mapping fields", bindName)

				// Start with the original configuration values
				data.Frontend.Bind[i] = configBind

				// Override only the fields that were explicitly set in the user's configuration
				// Always update these core fields from HAProxy
				data.Frontend.Bind[i].Name = types.StringValue(bind.Name)
				data.Frontend.Bind[i].Address = types.StringValue(bind.Address)
				data.Frontend.Bind[i].Port = types.Int64Value(*bind.Port)

				// Only override fields that were explicitly set in the user's config
				if !configBind.PortRangeEnd.IsNull() && bind.PortRangeEnd != nil {
					data.Frontend.Bind[i].PortRangeEnd = types.Int64Value(*bind.PortRangeEnd)
				}
				if !configBind.Transparent.IsNull() {
					data.Frontend.Bind[i].Transparent = types.BoolValue(bind.Transparent)
				}
				if !configBind.Mode.IsNull() && bind.Mode != "" {
					data.Frontend.Bind[i].Mode = types.StringValue(bind.Mode)
				}
				if !configBind.Maxconn.IsNull() {
					data.Frontend.Bind[i].Maxconn = types.Int64Value(bind.Maxconn)
				}
				if !configBind.Ssl.IsNull() {
					data.Frontend.Bind[i].Ssl = types.BoolValue(bind.Ssl)
				}
				if !configBind.SslCafile.IsNull() && bind.SslCafile != "" {
					data.Frontend.Bind[i].SslCafile = types.StringValue(bind.SslCafile)
				}
				if !configBind.SslCertificate.IsNull() && bind.SslCertificate != "" {
					data.Frontend.Bind[i].SslCertificate = types.StringValue(bind.SslCertificate)
				}
				if !configBind.SslMaxVer.IsNull() && bind.SslMaxVer != "" {
					data.Frontend.Bind[i].SslMaxVer = types.StringValue(bind.SslMaxVer)
				}
				if !configBind.SslMinVer.IsNull() && bind.SslMinVer != "" {
					data.Frontend.Bind[i].SslMinVer = types.StringValue(bind.SslMinVer)
				}
				if !configBind.Ciphers.IsNull() && bind.Ciphers != "" {
					data.Frontend.Bind[i].Ciphers = types.StringValue(bind.Ciphers)
				}
				if !configBind.Ciphersuites.IsNull() && bind.Ciphersuites != "" {
					data.Frontend.Bind[i].Ciphersuites = types.StringValue(bind.Ciphersuites)
				}
				if !configBind.Verify.IsNull() && bind.Verify != "" {
					data.Frontend.Bind[i].Verify = types.StringValue(bind.Verify)
				}
				if !configBind.AcceptProxy.IsNull() {
					data.Frontend.Bind[i].AcceptProxy = types.BoolValue(bind.AcceptProxy)
				}
				if !configBind.Allow0rtt.IsNull() {
					data.Frontend.Bind[i].Allow0rtt = types.BoolValue(bind.Allow0rtt)
				}
				if !configBind.Alpn.IsNull() && bind.Alpn != "" {
					data.Frontend.Bind[i].Alpn = types.StringValue(bind.Alpn)
				}
				if !configBind.Backlog.IsNull() && bind.Backlog != "" {
					data.Frontend.Bind[i].Backlog = types.StringValue(bind.Backlog)
				}
				if !configBind.DeferAccept.IsNull() {
					data.Frontend.Bind[i].DeferAccept = types.BoolValue(bind.DeferAccept)
				}
				if !configBind.GenerateCertificates.IsNull() {
					data.Frontend.Bind[i].GenerateCertificates = types.BoolValue(bind.GenerateCertificates)
				}
				if !configBind.Gid.IsNull() {
					data.Frontend.Bind[i].Gid = types.Int64Value(bind.Gid)
				}
				if !configBind.Group.IsNull() && bind.Group != "" {
					data.Frontend.Bind[i].Group = types.StringValue(bind.Group)
				}
				if !configBind.Id.IsNull() && bind.Id != "" {
					data.Frontend.Bind[i].Id = types.StringValue(bind.Id)
				}
				if !configBind.Interface.IsNull() && bind.Interface != "" {
					data.Frontend.Bind[i].Interface = types.StringValue(bind.Interface)
				}
				if !configBind.Level.IsNull() && bind.Level != "" {
					data.Frontend.Bind[i].Level = types.StringValue(bind.Level)
				}
				if !configBind.Namespace.IsNull() && bind.Namespace != "" {
					data.Frontend.Bind[i].Namespace = types.StringValue(bind.Namespace)
				}
				if !configBind.Nice.IsNull() {
					data.Frontend.Bind[i].Nice = types.Int64Value(bind.Nice)
				}
				if !configBind.NoCaNames.IsNull() {
					data.Frontend.Bind[i].NoCaNames = types.BoolValue(bind.NoCaNames)
				}
				if !configBind.Npn.IsNull() && bind.Npn != "" {
					data.Frontend.Bind[i].Npn = types.StringValue(bind.Npn)
				}
				if !configBind.PreferClientCiphers.IsNull() {
					data.Frontend.Bind[i].PreferClientCiphers = types.BoolValue(bind.PreferClientCiphers)
				}
				// Process field - only supported in v2, not v3
				if apiVersion == "v2" && !configBind.Process.IsNull() {
					data.Frontend.Bind[i].Process = types.StringValue(bind.Process)
				}
				if !configBind.Proto.IsNull() && bind.Proto != "" {
					data.Frontend.Bind[i].Proto = types.StringValue(bind.Proto)
				}
				if !configBind.SeverityOutput.IsNull() && bind.SeverityOutput != "" {
					data.Frontend.Bind[i].SeverityOutput = types.StringValue(bind.SeverityOutput)
				}
				if !configBind.StrictSni.IsNull() {
					data.Frontend.Bind[i].StrictSni = types.BoolValue(bind.StrictSni)
				}
				if !configBind.TcpUserTimeout.IsNull() {
					data.Frontend.Bind[i].TcpUserTimeout = types.Int64Value(bind.TcpUserTimeout)
				}
				if !configBind.Tfo.IsNull() {
					data.Frontend.Bind[i].Tfo = types.BoolValue(bind.Tfo)
				}
				if !configBind.TlsTicketKeys.IsNull() && bind.TlsTicketKeys != "" {
					data.Frontend.Bind[i].TlsTicketKeys = types.StringValue(bind.TlsTicketKeys)
				}
				if !configBind.Uid.IsNull() && bind.Uid != "" {
					data.Frontend.Bind[i].Uid = types.StringValue(bind.Uid)
				}
				if !configBind.User.IsNull() && bind.User != "" {
					data.Frontend.Bind[i].User = types.StringValue(bind.User)
				}
				if !configBind.V4v6.IsNull() {
					data.Frontend.Bind[i].V4v6 = types.BoolValue(bind.V4v6)
				}
				if !configBind.V6only.IsNull() {
					data.Frontend.Bind[i].V6only = types.BoolValue(bind.V6only)
				}

				// v3 fields - only override if explicitly set in config
				if !configBind.Sslv3.IsNull() {
					data.Frontend.Bind[i].Sslv3 = types.BoolValue(bind.Sslv3)
				}
				if !configBind.Tlsv10.IsNull() {
					data.Frontend.Bind[i].Tlsv10 = types.BoolValue(bind.Tlsv10)
				}
				if !configBind.Tlsv11.IsNull() {
					data.Frontend.Bind[i].Tlsv11 = types.BoolValue(bind.Tlsv11)
				}
				// TLS version fields - not supported in either v2 or v3 for binds
				// (HAProxy doesn't store these fields, so keep original config values)
				if !configBind.TlsTickets.IsNull() && bind.TlsTickets != "" {
					data.Frontend.Bind[i].TlsTickets = types.StringValue(bind.TlsTickets)
				}
				if !configBind.ForceStrictSni.IsNull() && bind.ForceStrictSni != "" {
					data.Frontend.Bind[i].ForceStrictSni = types.StringValue(bind.ForceStrictSni)
				}
				if !configBind.NoStrictSni.IsNull() {
					data.Frontend.Bind[i].NoStrictSni = types.BoolValue(bind.NoStrictSni)
				}
				if !configBind.GuidPrefix.IsNull() && bind.GuidPrefix != "" {
					data.Frontend.Bind[i].GuidPrefix = types.StringValue(bind.GuidPrefix)
				}
				if !configBind.IdlePing.IsNull() && bind.IdlePing != nil {
					data.Frontend.Bind[i].IdlePing = types.Int64Value(*bind.IdlePing)
				}
				if !configBind.QuicCcAlgo.IsNull() && bind.QuicCcAlgo != "" {
					data.Frontend.Bind[i].QuicCcAlgo = types.StringValue(bind.QuicCcAlgo)
				}
				if !configBind.QuicForceRetry.IsNull() {
					data.Frontend.Bind[i].QuicForceRetry = types.BoolValue(bind.QuicForceRetry)
				}
				if !configBind.QuicSocket.IsNull() && bind.QuicSocket != "" {
					data.Frontend.Bind[i].QuicSocket = types.StringValue(bind.QuicSocket)
				}
				if !configBind.QuicCcAlgoBurstSize.IsNull() && bind.QuicCcAlgoBurstSize != nil {
					data.Frontend.Bind[i].QuicCcAlgoBurstSize = types.Int64Value(*bind.QuicCcAlgoBurstSize)
				}
				if !configBind.QuicCcAlgoMaxWindow.IsNull() && bind.QuicCcAlgoMaxWindow != nil {
					data.Frontend.Bind[i].QuicCcAlgoMaxWindow = types.Int64Value(*bind.QuicCcAlgoMaxWindow)
				}
				// Metadata field - not supported in either v2 or v3 for binds
				// (HAProxy doesn't store this field, so keep original config value)

				// v2 fields (deprecated in v3) - only override if explicitly set in config
				if !configBind.NoSslv3.IsNull() && bind.NoSslv3 {
					data.Frontend.Bind[i].NoSslv3 = types.BoolValue(bind.NoSslv3)
				}
				if !configBind.ForceSslv3.IsNull() && bind.ForceSslv3 {
					data.Frontend.Bind[i].ForceSslv3 = types.BoolValue(bind.ForceSslv3)
				}
				if !configBind.ForceTlsv10.IsNull() && bind.ForceTlsv10 {
					data.Frontend.Bind[i].ForceTlsv10 = types.BoolValue(bind.ForceTlsv10)
				}
				if !configBind.ForceTlsv11.IsNull() && bind.ForceTlsv11 {
					data.Frontend.Bind[i].ForceTlsv11 = types.BoolValue(bind.ForceTlsv11)
				}
				if !configBind.ForceTlsv12.IsNull() && bind.ForceTlsv12 {
					data.Frontend.Bind[i].ForceTlsv12 = types.BoolValue(bind.ForceTlsv12)
				}
				if !configBind.ForceTlsv13.IsNull() && bind.ForceTlsv13 {
					data.Frontend.Bind[i].ForceTlsv13 = types.BoolValue(bind.ForceTlsv13)
				}
				if !configBind.NoTlsv10.IsNull() && bind.NoTlsv10 {
					data.Frontend.Bind[i].NoTlsv10 = types.BoolValue(bind.NoTlsv10)
				}
				if !configBind.NoTlsv11.IsNull() && bind.NoTlsv11 {
					data.Frontend.Bind[i].NoTlsv11 = types.BoolValue(bind.NoTlsv11)
				}
				if !configBind.NoTlsv12.IsNull() && bind.NoTlsv12 {
					data.Frontend.Bind[i].NoTlsv12 = types.BoolValue(bind.NoTlsv12)
				}
				if !configBind.NoTlsv13.IsNull() && bind.NoTlsv13 {
					data.Frontend.Bind[i].NoTlsv13 = types.BoolValue(bind.NoTlsv13)
				}
				if !configBind.NoTlsTickets.IsNull() && bind.NoTlsTickets {
					data.Frontend.Bind[i].NoTlsTickets = types.BoolValue(bind.NoTlsTickets)
				}
			} else {
				// Bind not found in HAProxy, keep the configuration values
				log.Printf("DEBUG: Bind '%s' not found in HAProxy, keeping configuration values", bindName)
				data.Frontend.Bind[i] = configBind
			}
		}
	}

	// Read ACLs if specified
	if len(data.Acls) > 0 {
		// ACLs reading would need to be implemented
		tflog.Info(ctx, "ACLs reading not yet implemented")
	}

	tflog.Info(ctx, "HAProxy stack read successfully")
	return nil
}

// Update performs the update operation for the haproxy_stack resource
func (o *StackOperations) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Updating HAProxy stack")

	// Begin transaction for all updates
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return err
	}

	// Use defer to ensure rollback on error
	defer func() {
		if err != nil {
			tflog.Info(ctx, "Rolling back transaction due to error", map[string]interface{}{"transaction_id": transactionID})
			if rollbackErr := o.client.RollbackTransaction(transactionID); rollbackErr != nil {
				tflog.Error(ctx, "Error rolling back transaction", map[string]interface{}{"error": rollbackErr.Error()})
			}
		}
	}()

	// Update backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Updating backend", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.backendManager.UpdateBackendInTransaction(ctx, transactionID, data.Backend); err != nil {
			resp.Diagnostics.AddError("Error updating backend", err.Error())
			return err
		}
	}

	// Update servers if specified
	if len(data.Servers) > 0 && data.Backend != nil {
		for _, server := range data.Servers {
			serverName := server.Name.ValueString()
			tflog.Info(ctx, "Managing server", map[string]interface{}{"server_name": serverName})

			// Try to update first, if it fails with 404, create it
			serverPayload := &ServerPayload{
				Name:    server.Name.ValueString(),
				Address: server.Address.ValueString(),
				Port:    server.Port.ValueInt64(),
				Check:   server.Check.ValueString(),
				Maxconn: server.Maxconn.ValueInt64(),
				Weight:  server.Weight.ValueInt64(),
			}

			if err = o.client.UpdateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
				// If update fails with 404, try to create the server
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "does not exist") {
					tflog.Info(ctx, "Server not found, creating new server", map[string]interface{}{"server_name": serverName})
					if err = o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
						resp.Diagnostics.AddError("Error creating server", err.Error())
						return err
					}
				} else {
					resp.Diagnostics.AddError("Error updating server", err.Error())
					return err
				}
			}
		}
	}

	// Update frontend if specified
	if data.Frontend != nil {
		tflog.Info(ctx, "Updating frontend", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.frontendManager.UpdateFrontendInTransaction(ctx, transactionID, data.Frontend); err != nil {
			resp.Diagnostics.AddError("Error updating frontend", err.Error())
			return err
		}
	}

	// Update binds for frontend if specified
	if data.Frontend != nil && data.Frontend.Bind != nil {
		tflog.Info(ctx, "Updating binds", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.bindManager.UpdateBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Bind); err != nil {
			resp.Diagnostics.AddError("Error updating binds", err.Error())
			return err
		}
	}

	// Update HTTP Request Rules if specified
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		tflog.Info(ctx, "Updating HTTP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.UpdateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
			resp.Diagnostics.AddError("Error updating HTTP request rules", err.Error())
			return err
		}
	}

	// Commit all updates
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err = o.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
		return err
	}

	// Clear the error so defer doesn't rollback
	err = nil
	tflog.Info(ctx, "HAProxy stack updated successfully")
	return nil
}

// Delete performs the delete operation for the haproxy_stack resource
func (o *StackOperations) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Deleting HAProxy stack")

	// Begin transaction for all deletes
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return err
	}

	// Use defer to ensure rollback on error
	defer func() {
		if err != nil {
			tflog.Info(ctx, "Rolling back transaction due to error", map[string]interface{}{"transaction_id": transactionID})
			if rollbackErr := o.client.RollbackTransaction(transactionID); rollbackErr != nil {
				tflog.Error(ctx, "Error rolling back transaction", map[string]interface{}{"error": rollbackErr.Error()})
			}
		}
	}()

	// Delete HTTP Request Rules if specified
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		tflog.Info(ctx, "Deleting HTTP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.DeleteHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting HTTP request rules", err.Error())
			return err
		}
	}

	// Delete binds for frontend if specified
	if data.Frontend != nil && data.Frontend.Bind != nil && len(data.Frontend.Bind) > 0 {
		tflog.Info(ctx, "Deleting binds", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.bindManager.DeleteBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting binds", err.Error())
			return err
		}
	}

	// Delete frontend if specified
	if data.Frontend != nil {
		tflog.Info(ctx, "Deleting frontend", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.frontendManager.DeleteFrontendInTransaction(ctx, transactionID, data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting frontend", err.Error())
			return err
		}
	}

	// Delete servers if specified
	if len(data.Servers) > 0 && data.Backend != nil {
		for _, server := range data.Servers {
			tflog.Info(ctx, "Deleting server", map[string]interface{}{"server_name": server.Name.ValueString()})
			if err = o.client.DeleteServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), server.Name.ValueString()); err != nil {
				resp.Diagnostics.AddError("Error deleting server", err.Error())
				return err
			}
		}
	}

	// Delete backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Deleting backend", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.backendManager.DeleteBackendInTransaction(ctx, transactionID, data.Backend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting backend", err.Error())
			return err
		}
	}

	// Commit all deletes
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err = o.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
		return err
	}

	// Clear the error so defer doesn't rollback
	err = nil
	tflog.Info(ctx, "HAProxy stack deleted successfully")
	return nil
}
