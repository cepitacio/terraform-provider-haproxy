package haproxy

import (
	"context"
	"fmt"
	"log"

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

// equalBoolPtr compares two bool pointers for equality
func equalBoolPtr(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// serverNeedsUpdate checks if a server needs to be updated
func (o *StackOperations) serverNeedsUpdate(existing ServerPayload, desired ServerPayload) bool {
	// Check all the fields that can be configured
	// Note: We don't compare Disabled field since HAProxy doesn't support it
	return existing.Address != desired.Address ||
		existing.Port != desired.Port ||
		existing.Check != desired.Check ||
		existing.Maxconn != desired.Maxconn ||
		existing.Weight != desired.Weight ||
		existing.Backup != desired.Backup ||
		existing.Cookie != desired.Cookie ||
		existing.Downinter != desired.Downinter ||
		existing.Fall != desired.Fall ||
		existing.Fastinter != desired.Fastinter ||
		existing.Inter != desired.Inter ||
		existing.Rise != desired.Rise ||
		existing.Ssl != desired.Ssl ||
		existing.Verify != desired.Verify
}

// convertServerPayloadToModel converts a ServerPayload to haproxyServerModel
func (o *StackOperations) convertServerPayloadToModel(server ServerPayload) haproxyServerModel {
	model := haproxyServerModel{
		// Note: Name is now the map key, not a field
		Address: types.StringValue(server.Address),
		Port:    types.Int64Value(server.Port),
	}

	// Set optional fields if they have values
	if server.Check != "" {
		model.Check = types.StringValue(server.Check)
	}
	if server.Maxconn != 0 {
		model.Maxconn = types.Int64Value(server.Maxconn)
	}
	if server.Weight != 0 {
		model.Weight = types.Int64Value(server.Weight)
	}
	if server.Backup != "" {
		model.Backup = types.StringValue(server.Backup)
	}
	if server.Rise != 0 {
		model.Rise = types.Int64Value(server.Rise)
	}
	if server.Fall != 0 {
		model.Fall = types.Int64Value(server.Fall)
	}
	if server.Inter != 0 {
		model.Inter = types.Int64Value(server.Inter)
	}
	if server.Fastinter != 0 {
		model.Fastinter = types.Int64Value(server.Fastinter)
	}
	if server.Downinter != 0 {
		model.Downinter = types.Int64Value(server.Downinter)
	}
	if server.Ssl != "" {
		model.Ssl = types.StringValue(server.Ssl)
	}
	if server.Verify != "" {
		model.Verify = types.StringValue(server.Verify)
	}
	if server.Cookie != "" {
		model.Cookie = types.StringValue(server.Cookie)
	}
	// Note: HAProxy doesn't support the 'disabled' field for servers
	// This field has been removed from the schema

	// SSL/TLS Protocol Control (v3 fields)
	if server.Sslv3 != "" {
		model.Sslv3 = types.StringValue(server.Sslv3)
	}
	if server.Tlsv10 != "" {
		model.Tlsv10 = types.StringValue(server.Tlsv10)
	}
	if server.Tlsv11 != "" {
		model.Tlsv11 = types.StringValue(server.Tlsv11)
	}
	if server.Tlsv12 != "" {
		model.Tlsv12 = types.StringValue(server.Tlsv12)
	}
	if server.Tlsv13 != "" {
		model.Tlsv13 = types.StringValue(server.Tlsv13)
	}

	// SSL/TLS Protocol Control (deprecated v2 fields)
	if server.NoSslv3 != "" {
		model.NoSslv3 = types.StringValue(server.NoSslv3)
	}
	if server.NoTlsv10 != "" {
		model.NoTlsv10 = types.StringValue(server.NoTlsv10)
	}
	if server.NoTlsv11 != "" {
		model.NoTlsv11 = types.StringValue(server.NoTlsv11)
	}
	if server.NoTlsv12 != "" {
		model.NoTlsv12 = types.StringValue(server.NoTlsv12)
	}
	if server.NoTlsv13 != "" {
		model.NoTlsv13 = types.StringValue(server.NoTlsv13)
	}
	if server.ForceSslv3 != "" {
		model.ForceSslv3 = types.StringValue(server.ForceSslv3)
	}
	if server.ForceTlsv10 != "" {
		model.ForceTlsv10 = types.StringValue(server.ForceTlsv10)
	}
	if server.ForceTlsv11 != "" {
		model.ForceTlsv11 = types.StringValue(server.ForceTlsv11)
	}
	if server.ForceTlsv12 != "" {
		model.ForceTlsv12 = types.StringValue(server.ForceTlsv12)
	}
	if server.ForceTlsv13 != "" {
		model.ForceTlsv13 = types.StringValue(server.ForceTlsv13)
	}

	return model
}

// convertServerModelToPayload converts a haproxyServerModel to ServerPayload
func (o *StackOperations) convertServerModelToPayload(serverName string, server haproxyServerModel) *ServerPayload {
	payload := &ServerPayload{
		Name:    serverName,
		Address: server.Address.ValueString(),
		Port:    server.Port.ValueInt64(),
	}

	// Set optional fields if they have values
	if !server.Check.IsNull() && !server.Check.IsUnknown() {
		payload.Check = server.Check.ValueString()
	}
	if !server.Backup.IsNull() && !server.Backup.IsUnknown() {
		payload.Backup = server.Backup.ValueString()
	}
	if !server.Maxconn.IsNull() && !server.Maxconn.IsUnknown() {
		payload.Maxconn = server.Maxconn.ValueInt64()
	}
	if !server.Weight.IsNull() && !server.Weight.IsUnknown() {
		payload.Weight = server.Weight.ValueInt64()
	}
	if !server.Rise.IsNull() && !server.Rise.IsUnknown() {
		payload.Rise = server.Rise.ValueInt64()
	}
	if !server.Fall.IsNull() && !server.Fall.IsUnknown() {
		payload.Fall = server.Fall.ValueInt64()
	}
	if !server.Inter.IsNull() && !server.Inter.IsUnknown() {
		payload.Inter = server.Inter.ValueInt64()
	}
	if !server.Fastinter.IsNull() && !server.Fastinter.IsUnknown() {
		payload.Fastinter = server.Fastinter.ValueInt64()
	}
	if !server.Downinter.IsNull() && !server.Downinter.IsUnknown() {
		payload.Downinter = server.Downinter.ValueInt64()
	}
	if !server.Ssl.IsNull() && !server.Ssl.IsUnknown() {
		payload.Ssl = server.Ssl.ValueString()
	}
	if !server.Verify.IsNull() && !server.Verify.IsUnknown() {
		payload.Verify = server.Verify.ValueString()
	}
	if !server.Cookie.IsNull() && !server.Cookie.IsUnknown() {
		payload.Cookie = server.Cookie.ValueString()
	}
	// Note: HAProxy doesn't support the 'disabled' field for servers
	// We don't send it to HAProxy, but we allow it in the Terraform config
	// for user convenience. It will always be read as false from HAProxy.

	// SSL/TLS Protocol Control (v3 fields)
	if !server.Sslv3.IsNull() && !server.Sslv3.IsUnknown() {
		payload.Sslv3 = server.Sslv3.ValueString()
	}
	if !server.Tlsv10.IsNull() && !server.Tlsv10.IsUnknown() {
		payload.Tlsv10 = server.Tlsv10.ValueString()
	}
	if !server.Tlsv11.IsNull() && !server.Tlsv11.IsUnknown() {
		payload.Tlsv11 = server.Tlsv11.ValueString()
	}
	if !server.Tlsv12.IsNull() && !server.Tlsv12.IsUnknown() {
		payload.Tlsv12 = server.Tlsv12.ValueString()
	}
	if !server.Tlsv13.IsNull() && !server.Tlsv13.IsUnknown() {
		payload.Tlsv13 = server.Tlsv13.ValueString()
	}

	// SSL/TLS Protocol Control (deprecated v2 fields)
	if !server.NoSslv3.IsNull() && !server.NoSslv3.IsUnknown() {
		payload.NoSslv3 = server.NoSslv3.ValueString()
	}
	if !server.NoTlsv10.IsNull() && !server.NoTlsv10.IsUnknown() {
		payload.NoTlsv10 = server.NoTlsv10.ValueString()
	}
	if !server.NoTlsv11.IsNull() && !server.NoTlsv11.IsUnknown() {
		payload.NoTlsv11 = server.NoTlsv11.ValueString()
	}
	if !server.NoTlsv12.IsNull() && !server.NoTlsv12.IsUnknown() {
		payload.NoTlsv12 = server.NoTlsv12.ValueString()
	}
	if !server.NoTlsv13.IsNull() && !server.NoTlsv13.IsUnknown() {
		payload.NoTlsv13 = server.NoTlsv13.ValueString()
	}
	if !server.ForceSslv3.IsNull() && !server.ForceSslv3.IsUnknown() {
		payload.ForceSslv3 = server.ForceSslv3.ValueString()
	}
	if !server.ForceTlsv10.IsNull() && !server.ForceTlsv10.IsUnknown() {
		payload.ForceTlsv10 = server.ForceTlsv10.ValueString()
	}
	if !server.ForceTlsv11.IsNull() && !server.ForceTlsv11.IsUnknown() {
		payload.ForceTlsv11 = server.ForceTlsv11.ValueString()
	}
	if !server.ForceTlsv12.IsNull() && !server.ForceTlsv12.IsUnknown() {
		payload.ForceTlsv12 = server.ForceTlsv12.ValueString()
	}
	if !server.ForceTlsv13.IsNull() && !server.ForceTlsv13.IsUnknown() {
		payload.ForceTlsv13 = server.ForceTlsv13.ValueString()
	}

	return payload
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
		for serverName, server := range data.Servers {
			serverPayload := o.convertServerModelToPayload(serverName, server)
			tflog.Info(ctx, "Creating server", map[string]interface{}{
				"server_name":  serverName,
				"backend_name": data.Backend.Name.ValueString(),
			})
			if err := o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
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
	if data.Frontend != nil && data.Frontend.Binds != nil && len(data.Frontend.Binds) > 0 {
		if err := o.bindManager.CreateBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Binds); err != nil {
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
	if data.Backend != nil {
		tflog.Info(ctx, "Reading servers from HAProxy", map[string]interface{}{
			"backend_name":          data.Backend.Name.ValueString(),
			"current_servers_count": len(data.Servers),
		})

		servers, err := o.client.ReadServers(ctx, "backend", data.Backend.Name.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Could not read servers, preserving existing state", map[string]interface{}{"error": err.Error()})
			// Don't overwrite data.Servers if we can't read from HAProxy
			// This preserves the existing state
		} else {
			tflog.Info(ctx, "Successfully read servers from HAProxy", map[string]interface{}{
				"servers_found": len(servers),
			})
			// Convert servers to map format
			data.Servers = make(map[string]haproxyServerModel)
			for _, server := range servers {
				data.Servers[server.Name] = o.convertServerPayloadToModel(server)
				tflog.Info(ctx, "Converted server", map[string]interface{}{
					"server_name": server.Name,
				})
			}
		}
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
		originalConfigBinds := data.Frontend.Binds
		// Convert bind list to map for processing
		data.Frontend.Binds = make(map[string]haproxyBindModel)
		log.Printf("DEBUG: Processing %d configuration binds:", len(originalConfigBinds))
		for bindName, configBind := range originalConfigBinds {
			log.Printf("DEBUG: Looking for bind '%s' in bind map", bindName)
			if bind, exists := bindMap[bindName]; exists {
				log.Printf("DEBUG: Found bind '%s' in HAProxy, mapping fields", bindName)

				// Start with the original configuration values
				data.Frontend.Binds[bindName] = configBind

				// Override only the fields that were explicitly set in the user's configuration
				// Always update these core fields from HAProxy
				bindModel := data.Frontend.Binds[bindName]
				bindModel.Address = types.StringValue(bind.Address)
				bindModel.Port = types.Int64Value(*bind.Port)
				data.Frontend.Binds[bindName] = bindModel

				// Only override fields that were explicitly set in the user's config
				if !configBind.PortRangeEnd.IsNull() && bind.PortRangeEnd != nil {
					bindModel.PortRangeEnd = types.Int64Value(*bind.PortRangeEnd)
				}
				if !configBind.Transparent.IsNull() {
					bindModel.Transparent = types.BoolValue(bind.Transparent)
				}
				if !configBind.Mode.IsNull() && bind.Mode != "" {
					bindModel.Mode = types.StringValue(bind.Mode)
				}
				if !configBind.Maxconn.IsNull() {
					bindModel.Maxconn = types.Int64Value(bind.Maxconn)
				}
				if !configBind.Ssl.IsNull() {
					bindModel.Ssl = types.BoolValue(bind.Ssl)
				}
				if !configBind.SslCafile.IsNull() && bind.SslCafile != "" {
					bindModel.SslCafile = types.StringValue(bind.SslCafile)
				}
				if !configBind.SslCertificate.IsNull() && bind.SslCertificate != "" {
					bindModel.SslCertificate = types.StringValue(bind.SslCertificate)
				}
				if !configBind.SslMaxVer.IsNull() && bind.SslMaxVer != "" {
					bindModel.SslMaxVer = types.StringValue(bind.SslMaxVer)
				}
				if !configBind.SslMinVer.IsNull() && bind.SslMinVer != "" {
					bindModel.SslMinVer = types.StringValue(bind.SslMinVer)
				}
				if !configBind.Ciphers.IsNull() && bind.Ciphers != "" {
					bindModel.Ciphers = types.StringValue(bind.Ciphers)
				}
				if !configBind.Ciphersuites.IsNull() && bind.Ciphersuites != "" {
					bindModel.Ciphersuites = types.StringValue(bind.Ciphersuites)
				}
				if !configBind.Verify.IsNull() && bind.Verify != "" {
					bindModel.Verify = types.StringValue(bind.Verify)
				}
				if !configBind.AcceptProxy.IsNull() {
					bindModel.AcceptProxy = types.BoolValue(bind.AcceptProxy)
				}
				if !configBind.Allow0rtt.IsNull() {
					bindModel.Allow0rtt = types.BoolValue(bind.Allow0rtt)
				}
				if !configBind.Alpn.IsNull() && bind.Alpn != "" {
					bindModel.Alpn = types.StringValue(bind.Alpn)
				}
				if !configBind.Backlog.IsNull() && bind.Backlog != "" {
					bindModel.Backlog = types.StringValue(bind.Backlog)
				}
				if !configBind.DeferAccept.IsNull() {
					bindModel.DeferAccept = types.BoolValue(bind.DeferAccept)
				}
				if !configBind.GenerateCertificates.IsNull() {
					bindModel.GenerateCertificates = types.BoolValue(bind.GenerateCertificates)
				}
				if !configBind.Gid.IsNull() {
					bindModel.Gid = types.Int64Value(bind.Gid)
				}
				if !configBind.Group.IsNull() && bind.Group != "" {
					bindModel.Group = types.StringValue(bind.Group)
				}
				if !configBind.Id.IsNull() && bind.Id != "" {
					bindModel.Id = types.StringValue(bind.Id)
				}
				if !configBind.Interface.IsNull() && bind.Interface != "" {
					bindModel.Interface = types.StringValue(bind.Interface)
				}
				if !configBind.Level.IsNull() && bind.Level != "" {
					bindModel.Level = types.StringValue(bind.Level)
				}
				if !configBind.Namespace.IsNull() && bind.Namespace != "" {
					bindModel.Namespace = types.StringValue(bind.Namespace)
				}
				if !configBind.Nice.IsNull() {
					bindModel.Nice = types.Int64Value(bind.Nice)
				}
				if !configBind.NoCaNames.IsNull() {
					bindModel.NoCaNames = types.BoolValue(bind.NoCaNames)
				}
				if !configBind.Npn.IsNull() && bind.Npn != "" {
					bindModel.Npn = types.StringValue(bind.Npn)
				}
				if !configBind.PreferClientCiphers.IsNull() {
					bindModel.PreferClientCiphers = types.BoolValue(bind.PreferClientCiphers)
				}
				// Process field - only supported in v2, not v3
				if apiVersion == "v2" && !configBind.Process.IsNull() {
					bindModel.Process = types.StringValue(bind.Process)
				}
				if !configBind.Proto.IsNull() && bind.Proto != "" {
					bindModel.Proto = types.StringValue(bind.Proto)
				}
				if !configBind.SeverityOutput.IsNull() && bind.SeverityOutput != "" {
					bindModel.SeverityOutput = types.StringValue(bind.SeverityOutput)
				}
				if !configBind.StrictSni.IsNull() {
					bindModel.StrictSni = types.BoolValue(bind.StrictSni)
				}
				if !configBind.TcpUserTimeout.IsNull() {
					bindModel.TcpUserTimeout = types.Int64Value(bind.TcpUserTimeout)
				}
				if !configBind.Tfo.IsNull() {
					bindModel.Tfo = types.BoolValue(bind.Tfo)
				}
				if !configBind.TlsTicketKeys.IsNull() && bind.TlsTicketKeys != "" {
					bindModel.TlsTicketKeys = types.StringValue(bind.TlsTicketKeys)
				}
				if !configBind.Uid.IsNull() && bind.Uid != "" {
					bindModel.Uid = types.StringValue(bind.Uid)
				}
				if !configBind.User.IsNull() && bind.User != "" {
					bindModel.User = types.StringValue(bind.User)
				}
				if !configBind.V4v6.IsNull() {
					bindModel.V4v6 = types.BoolValue(bind.V4v6)
				}
				if !configBind.V6only.IsNull() {
					bindModel.V6only = types.BoolValue(bind.V6only)
				}

				// v3 fields - only override if explicitly set in config
				if !configBind.Sslv3.IsNull() {
					bindModel.Sslv3 = types.BoolValue(bind.Sslv3)
				}
				if !configBind.Tlsv10.IsNull() {
					bindModel.Tlsv10 = types.BoolValue(bind.Tlsv10)
				}
				if !configBind.Tlsv11.IsNull() {
					bindModel.Tlsv11 = types.BoolValue(bind.Tlsv11)
				}
				// TLS version fields - not supported in either v2 or v3 for binds
				// (HAProxy doesn't store these fields, so keep original config values)
				if !configBind.TlsTickets.IsNull() && bind.TlsTickets != "" {
					bindModel.TlsTickets = types.StringValue(bind.TlsTickets)
				}
				if !configBind.ForceStrictSni.IsNull() && bind.ForceStrictSni != "" {
					bindModel.ForceStrictSni = types.StringValue(bind.ForceStrictSni)
				}
				if !configBind.NoStrictSni.IsNull() {
					bindModel.NoStrictSni = types.BoolValue(bind.NoStrictSni)
				}
				if !configBind.GuidPrefix.IsNull() && bind.GuidPrefix != "" {
					bindModel.GuidPrefix = types.StringValue(bind.GuidPrefix)
				}
				if !configBind.IdlePing.IsNull() && bind.IdlePing != nil {
					bindModel.IdlePing = types.Int64Value(*bind.IdlePing)
				}
				if !configBind.QuicCcAlgo.IsNull() && bind.QuicCcAlgo != "" {
					bindModel.QuicCcAlgo = types.StringValue(bind.QuicCcAlgo)
				}
				if !configBind.QuicForceRetry.IsNull() {
					bindModel.QuicForceRetry = types.BoolValue(bind.QuicForceRetry)
				}
				if !configBind.QuicSocket.IsNull() && bind.QuicSocket != "" {
					bindModel.QuicSocket = types.StringValue(bind.QuicSocket)
				}
				if !configBind.QuicCcAlgoBurstSize.IsNull() && bind.QuicCcAlgoBurstSize != nil {
					bindModel.QuicCcAlgoBurstSize = types.Int64Value(*bind.QuicCcAlgoBurstSize)
				}
				if !configBind.QuicCcAlgoMaxWindow.IsNull() && bind.QuicCcAlgoMaxWindow != nil {
					bindModel.QuicCcAlgoMaxWindow = types.Int64Value(*bind.QuicCcAlgoMaxWindow)
				}
				// Metadata field - not supported in either v2 or v3 for binds
				// (HAProxy doesn't store this field, so keep original config value)

				// v2 fields (deprecated in v3) - only override if explicitly set in config
				if !configBind.NoSslv3.IsNull() && bind.NoSslv3 {
					bindModel.NoSslv3 = types.BoolValue(bind.NoSslv3)
				}
				if !configBind.ForceSslv3.IsNull() && bind.ForceSslv3 {
					bindModel.ForceSslv3 = types.BoolValue(bind.ForceSslv3)
				}
				if !configBind.ForceTlsv10.IsNull() && bind.ForceTlsv10 {
					bindModel.ForceTlsv10 = types.BoolValue(bind.ForceTlsv10)
				}
				if !configBind.ForceTlsv11.IsNull() && bind.ForceTlsv11 {
					bindModel.ForceTlsv11 = types.BoolValue(bind.ForceTlsv11)
				}
				if !configBind.ForceTlsv12.IsNull() && bind.ForceTlsv12 {
					bindModel.ForceTlsv12 = types.BoolValue(bind.ForceTlsv12)
				}
				if !configBind.ForceTlsv13.IsNull() && bind.ForceTlsv13 {
					bindModel.ForceTlsv13 = types.BoolValue(bind.ForceTlsv13)
				}
				if !configBind.NoTlsv10.IsNull() && bind.NoTlsv10 {
					bindModel.NoTlsv10 = types.BoolValue(bind.NoTlsv10)
				}
				if !configBind.NoTlsv11.IsNull() && bind.NoTlsv11 {
					bindModel.NoTlsv11 = types.BoolValue(bind.NoTlsv11)
				}
				if !configBind.NoTlsv12.IsNull() && bind.NoTlsv12 {
					bindModel.NoTlsv12 = types.BoolValue(bind.NoTlsv12)
				}
				if !configBind.NoTlsv13.IsNull() && bind.NoTlsv13 {
					bindModel.NoTlsv13 = types.BoolValue(bind.NoTlsv13)
				}
				if !configBind.NoTlsTickets.IsNull() && bind.NoTlsTickets {
					bindModel.NoTlsTickets = types.BoolValue(bind.NoTlsTickets)
				}
				// Assign the updated bind model back to the map
				data.Frontend.Binds[bindName] = bindModel
			} else {
				// Bind not found in HAProxy, keep the configuration values
				log.Printf("DEBUG: Bind '%s' not found in HAProxy, keeping configuration values", bindName)
				data.Frontend.Binds[bindName] = configBind
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

	// Update servers if specified - use name-based management
	if len(data.Servers) > 0 && data.Backend != nil {
		tflog.Info(ctx, "Updating servers", map[string]interface{}{
			"backend_name":          data.Backend.Name.ValueString(),
			"desired_servers_count": len(data.Servers),
		})

		// First, read existing servers to get current state
		existingServers, err := o.client.ReadServers(ctx, "backend", data.Backend.Name.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Could not read existing servers, proceeding with create/update", map[string]interface{}{"error": err.Error()})
			existingServers = []ServerPayload{}
		} else {
			tflog.Info(ctx, "Read existing servers from HAProxy", map[string]interface{}{
				"existing_servers_count": len(existingServers),
			})
		}

		// Create a map of existing servers by name
		existingServerMap := make(map[string]ServerPayload)
		for _, existingServer := range existingServers {
			existingServerMap[existingServer.Name] = existingServer
		}

		// Create a map of desired servers by name (data.Servers is already a map)
		desiredServerMap := data.Servers

		// Delete servers that are no longer in the desired state
		for serverName := range existingServerMap {
			if _, exists := desiredServerMap[serverName]; !exists {
				tflog.Info(ctx, "Deleting server", map[string]interface{}{"server_name": serverName})
				if err = o.client.DeleteServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverName); err != nil {
					resp.Diagnostics.AddError("Error deleting server", err.Error())
					return err
				}
			}
		}

		// Create or update servers in the desired state
		for serverName, server := range desiredServerMap {
			// Use the full conversion function to get all fields
			serverPayload := o.convertServerModelToPayload(serverName, server)

			// Debug logging for the payload being sent
			disabledStr := "nil"
			if serverPayload.Disabled != nil {
				disabledStr = fmt.Sprintf("%t", *serverPayload.Disabled)
			}
			tflog.Info(ctx, "Server payload for update", map[string]interface{}{
				"server_name": serverName,
				"disabled":    disabledStr,
				"check":       serverPayload.Check,
				"maxconn":     serverPayload.Maxconn,
				"weight":      serverPayload.Weight,
			})

			if existingServer, exists := existingServerMap[serverName]; exists {
				// Server exists, check if it needs updating
				if o.serverNeedsUpdate(existingServer, *serverPayload) {
					tflog.Info(ctx, "Updating server", map[string]interface{}{"server_name": serverName})
					if err = o.client.UpdateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
						resp.Diagnostics.AddError("Error updating server", err.Error())
						return err
					}
				} else {
					tflog.Info(ctx, "Server unchanged", map[string]interface{}{"server_name": serverName})
				}
			} else {
				// Server doesn't exist, create it
				tflog.Info(ctx, "Creating new server", map[string]interface{}{"server_name": serverName})
				if err = o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
					resp.Diagnostics.AddError("Error creating server", err.Error())
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
	if data.Frontend != nil && data.Frontend.Binds != nil {
		tflog.Info(ctx, "Updating binds", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.bindManager.UpdateBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Binds); err != nil {
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
	if data.Frontend != nil && data.Frontend.Binds != nil && len(data.Frontend.Binds) > 0 {
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

	// Delete servers if specified - use name-based management
	if len(data.Servers) > 0 && data.Backend != nil {
		// Read existing servers to get current state
		existingServers, err := o.client.ReadServers(ctx, "backend", data.Backend.Name.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Could not read existing servers for deletion", map[string]interface{}{"error": err.Error()})
			existingServers = []ServerPayload{}
		}

		// Create a map of desired servers by name (data.Servers is already a map)
		desiredServerMap := make(map[string]bool)
		for serverName := range data.Servers {
			desiredServerMap[serverName] = true
		}

		// Delete servers that are not in the desired state
		for _, existingServer := range existingServers {
			if !desiredServerMap[existingServer.Name] {
				tflog.Info(ctx, "Deleting server", map[string]interface{}{"server_name": existingServer.Name})
				if err = o.client.DeleteServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), existingServer.Name); err != nil {
					resp.Diagnostics.AddError("Error deleting server", err.Error())
					return err
				}
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
