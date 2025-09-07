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
	client                  *HAProxyClient
	aclManager              *ACLManager
	frontendManager         *FrontendManager
	backendManager          *BackendManager
	httpRequestRuleManager  *HttpRequestRuleManager
	httpResponseRuleManager *HttpResponseRuleManager
	bindManager             *BindManager
}

// NewStackOperations creates a new StackOperations instance
func NewStackOperations(client *HAProxyClient, aclManager *ACLManager, frontendManager *FrontendManager, backendManager *BackendManager, httpRequestRuleManager *HttpRequestRuleManager, httpResponseRuleManager *HttpResponseRuleManager, bindManager *BindManager) *StackOperations {
	return &StackOperations{
		client:                  client,
		aclManager:              aclManager,
		backendManager:          backendManager,
		frontendManager:         frontendManager,
		httpRequestRuleManager:  httpRequestRuleManager,
		httpResponseRuleManager: httpResponseRuleManager,
		bindManager:             bindManager,
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
	// Disabled field comparison skipped - HAProxy doesn't support server disabling
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
		// Server name is the map key, not a field in the payload
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
	// HAProxy doesn't support server disabling - field ignored
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

// convertHttpRequestRulePayloadToModel converts an HttpRequestRulePayload to haproxyHttpRequestRuleModel
func (o *StackOperations) convertHttpRequestRulePayloadToModel(rule HttpRequestRulePayload) haproxyHttpRequestRuleModel {
	model := haproxyHttpRequestRuleModel{
		Type: types.StringValue(rule.Type),
	}

	// Set optional fields if they have values
	if rule.Cond != "" {
		model.Cond = types.StringValue(rule.Cond)
	}
	if rule.CondTest != "" {
		model.CondTest = types.StringValue(rule.CondTest)
	}
	if rule.HdrName != "" {
		model.HdrName = types.StringValue(rule.HdrName)
	}
	if rule.HdrFormat != "" {
		model.HdrFormat = types.StringValue(rule.HdrFormat)
	}
	if rule.RedirType != "" {
		model.RedirType = types.StringValue(rule.RedirType)
	}
	if rule.RedirValue != "" {
		model.RedirValue = types.StringValue(rule.RedirValue)
	}

	return model
}

// convertHttpResponseRulePayloadToModel converts an HttpResponseRulePayload to haproxyHttpResponseRuleModel
func (o *StackOperations) convertHttpResponseRulePayloadToModel(rule HttpResponseRulePayload) haproxyHttpResponseRuleModel {
	model := haproxyHttpResponseRuleModel{
		Type: types.StringValue(rule.Type),
	}

	// Set optional fields if they have values
	if rule.Cond != "" {
		model.Cond = types.StringValue(rule.Cond)
	}
	if rule.CondTest != "" {
		model.CondTest = types.StringValue(rule.CondTest)
	}
	if rule.HdrName != "" {
		model.HdrName = types.StringValue(rule.HdrName)
	}
	if rule.HdrFormat != "" {
		model.HdrFormat = types.StringValue(rule.HdrFormat)
	}
	if rule.HdrMethod != "" {
		model.HdrMethod = types.StringValue(rule.HdrMethod)
	}
	if rule.RedirType != "" {
		model.RedirType = types.StringValue(rule.RedirType)
	}
	if rule.RedirValue != "" {
		model.RedirValue = types.StringValue(rule.RedirValue)
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
	// HAProxy doesn't support server disabling - field ignored
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

	// Create ACLs if specified - handle both frontend and backend ACLs
	if data.Frontend != nil && data.Frontend.Acls != nil && len(data.Frontend.Acls) > 0 {
		tflog.Info(ctx, "Creating frontend ACLs in transaction", map[string]interface{}{"transaction_id": transactionID})
		if err := o.aclManager.CreateACLsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Acls); err != nil {
			resp.Diagnostics.AddError("Error creating frontend ACLs", err.Error())
			return err
		}
	}

	if data.Backend != nil && data.Backend.Acls != nil && len(data.Backend.Acls) > 0 {
		tflog.Info(ctx, "Creating backend ACLs in transaction", map[string]interface{}{"transaction_id": transactionID})
		if err := o.aclManager.CreateACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.Acls); err != nil {
			resp.Diagnostics.AddError("Error creating backend ACLs", err.Error())
			return err
		}
	}

	// Create HTTP Request Rules AFTER ACLs (so they can reference existing ACLs)
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		if err := o.httpRequestRuleManager.CreateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
			resp.Diagnostics.AddError("Error creating HTTP request rules", err.Error())
			return err
		}
	}

	// Create Backend HTTP Request Rules AFTER ACLs (so they can reference existing ACLs)
	if data.Backend != nil && data.Backend.HttpRequestRule != nil && len(data.Backend.HttpRequestRule) > 0 {
		if err := o.httpRequestRuleManager.CreateHttpRequestRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.HttpRequestRule); err != nil {
			resp.Diagnostics.AddError("Error creating backend HTTP request rules", err.Error())
			return err
		}
	}

	// Create HTTP Response Rules AFTER HTTP Request Rules
	if data.Frontend != nil && data.Frontend.HttpResponseRules != nil && len(data.Frontend.HttpResponseRules) > 0 {
		if err := o.httpResponseRuleManager.CreateHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpResponseRules); err != nil {
			resp.Diagnostics.AddError("Error creating HTTP response rules", err.Error())
			return err
		}
	}

	// Create Backend HTTP Response Rules AFTER HTTP Request Rules
	if data.Backend != nil && data.Backend.HttpResponseRule != nil && len(data.Backend.HttpResponseRule) > 0 {
		if err := o.httpResponseRuleManager.CreateHttpResponseRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.HttpResponseRule); err != nil {
			resp.Diagnostics.AddError("Error creating backend HTTP response rules", err.Error())
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

	// HTTP Request and Response Rules are managed by Terraform state
	// We don't read them from HAProxy to avoid state drift issues
	// The Terraform state is the source of truth for these rules

	tflog.Info(ctx, "HAProxy stack read successfully")
	return nil
}

// Update performs the update operation for the haproxy_stack resource
func (o *StackOperations) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *haproxyStackResourceModel) error {
	// Get the previous state to compare with the plan
	var state haproxyStackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return fmt.Errorf("failed to get state data")
	}
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

	// Update backend only if it changed in the plan
	if data.Backend != nil {
		// Check if backend changed by comparing plan vs state
		backendChanged := o.backendChanged(ctx, data.Backend, state.Backend)
		if backendChanged {
			tflog.Info(ctx, "Backend changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			if err = o.backendManager.UpdateBackendInTransaction(ctx, transactionID, data.Backend); err != nil {
				resp.Diagnostics.AddError("Error updating backend", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "Backend unchanged, skipping update")
		}
	}

	// Update servers only if they changed in the plan
	if len(data.Servers) > 0 && data.Backend != nil {
		// Check if servers changed by comparing plan vs state
		serversChanged := o.serversChanged(ctx, data.Servers, state.Servers)
		if serversChanged {
			tflog.Info(ctx, "Servers changed, updating", map[string]interface{}{
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
		} else {
			tflog.Info(ctx, "Servers unchanged, skipping update")
		}
	}

	// Update frontend only if it changed in the plan
	if data.Frontend != nil {
		// Check if frontend changed by comparing plan vs state
		frontendChanged := o.frontendChanged(ctx, data.Frontend, state.Frontend)
		if frontendChanged {
			tflog.Info(ctx, "Frontend changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.frontendManager.UpdateFrontendInTransaction(ctx, transactionID, data.Frontend); err != nil {
				resp.Diagnostics.AddError("Error updating frontend", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "Frontend unchanged, skipping update")
		}
	}

	// Update binds only if they changed in the plan
	if data.Frontend != nil && data.Frontend.Binds != nil {
		// Check if binds changed by comparing plan vs state
		bindsChanged := o.bindsChanged(ctx, data.Frontend.Binds, state.Frontend.Binds)
		if bindsChanged {
			tflog.Info(ctx, "Binds changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.bindManager.UpdateBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Binds); err != nil {
				resp.Diagnostics.AddError("Error updating binds", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "Binds unchanged, skipping update")
		}
	}

	// Update frontend ACLs only if they changed in the plan
	if data.Frontend != nil && data.Frontend.Acls != nil && len(data.Frontend.Acls) > 0 {
		// Check if frontend ACLs changed by comparing plan vs state
		frontendACLsChanged := o.aclsChanged(ctx, data.Frontend.Acls, state.Frontend.Acls)
		if frontendACLsChanged {
			tflog.Info(ctx, "Frontend ACLs changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.aclManager.UpdateACLsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Acls); err != nil {
				resp.Diagnostics.AddError("Error updating frontend ACLs", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "Frontend ACLs unchanged, skipping update")
		}
	}

	// Update backend ACLs only if they changed in the plan
	if data.Backend != nil && data.Backend.Acls != nil && len(data.Backend.Acls) > 0 {
		// Check if backend ACLs changed by comparing plan vs state
		backendACLsChanged := o.aclsChanged(ctx, data.Backend.Acls, state.Backend.Acls)
		if backendACLsChanged {
			tflog.Info(ctx, "Backend ACLs changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			if err = o.aclManager.UpdateACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.Acls); err != nil {
				resp.Diagnostics.AddError("Error updating backend ACLs", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "Backend ACLs unchanged, skipping update")
		}
	}

	// Update HTTP Request Rules only if they changed in the plan
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		// Check if HTTP Request Rules changed by comparing plan vs state
		httpRequestRulesChanged := o.httpRequestRulesChanged(ctx, data.Frontend.HttpRequestRules, state.Frontend.HttpRequestRules)
		if httpRequestRulesChanged {
			tflog.Info(ctx, "HTTP request rules changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.httpRequestRuleManager.UpdateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
				resp.Diagnostics.AddError("Error updating HTTP request rules", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "HTTP request rules unchanged, skipping update")
		}
	}

	// Update HTTP Response Rules only if they changed in the plan
	if data.Frontend != nil && data.Frontend.HttpResponseRules != nil && len(data.Frontend.HttpResponseRules) > 0 {
		// Check if HTTP Response Rules changed by comparing plan vs state
		httpResponseRulesChanged := o.httpResponseRulesChanged(ctx, data.Frontend.HttpResponseRules, state.Frontend.HttpResponseRules)
		if httpResponseRulesChanged {
			tflog.Info(ctx, "HTTP response rules changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.httpResponseRuleManager.UpdateHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpResponseRules); err != nil {
				resp.Diagnostics.AddError("Error updating HTTP response rules", err.Error())
				return err
			}
		} else {
			tflog.Info(ctx, "HTTP response rules unchanged, skipping update")
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

// httpRequestRulesChanged compares plan vs state HTTP request rules to detect changes
func (o *StackOperations) httpRequestRulesChanged(ctx context.Context, planRules []haproxyHttpRequestRuleModel, stateRules []haproxyHttpRequestRuleModel) bool {
	// If counts are different, there's definitely a change
	if len(planRules) != len(stateRules) {
		tflog.Info(ctx, "HTTP request rules count changed", map[string]interface{}{
			"plan_count":  len(planRules),
			"state_count": len(stateRules),
		})
		return true
	}

	// Compare each rule
	for i, planRule := range planRules {
		if i >= len(stateRules) {
			return true
		}
		stateRule := stateRules[i]

		// Compare ALL key fields comprehensively (only confirmed existing fields)
		if planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Cond.ValueString() != stateRule.Cond.ValueString() ||
			planRule.CondTest.ValueString() != stateRule.CondTest.ValueString() ||
			planRule.HdrName.ValueString() != stateRule.HdrName.ValueString() ||
			planRule.HdrFormat.ValueString() != stateRule.HdrFormat.ValueString() ||
			planRule.RedirType.ValueString() != stateRule.RedirType.ValueString() ||
			planRule.RedirValue.ValueString() != stateRule.RedirValue.ValueString() {
			tflog.Info(ctx, "HTTP request rule changed", map[string]interface{}{
				"rule_index": i,
				"plan_type":  planRule.Type.ValueString(),
				"state_type": stateRule.Type.ValueString(),
			})
			return true
		}
	}

	return false
}

// httpResponseRulesChanged compares plan vs state HTTP response rules to detect changes
func (o *StackOperations) httpResponseRulesChanged(ctx context.Context, planRules []haproxyHttpResponseRuleModel, stateRules []haproxyHttpResponseRuleModel) bool {
	// If counts are different, there's definitely a change
	if len(planRules) != len(stateRules) {
		tflog.Info(ctx, "HTTP response rules count changed", map[string]interface{}{
			"plan_count":  len(planRules),
			"state_count": len(stateRules),
		})
		return true
	}

	// Compare each rule
	for i, planRule := range planRules {
		if i >= len(stateRules) {
			return true
		}
		stateRule := stateRules[i]

		// Compare ALL key fields comprehensively (only confirmed existing fields)
		if planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Cond.ValueString() != stateRule.Cond.ValueString() ||
			planRule.CondTest.ValueString() != stateRule.CondTest.ValueString() ||
			planRule.HdrName.ValueString() != stateRule.HdrName.ValueString() ||
			planRule.HdrFormat.ValueString() != stateRule.HdrFormat.ValueString() ||
			planRule.HdrMethod.ValueString() != stateRule.HdrMethod.ValueString() ||
			planRule.RedirType.ValueString() != stateRule.RedirType.ValueString() ||
			planRule.RedirValue.ValueString() != stateRule.RedirValue.ValueString() {
			tflog.Info(ctx, "HTTP response rule changed", map[string]interface{}{
				"rule_index": i,
				"plan_type":  planRule.Type.ValueString(),
				"state_type": stateRule.Type.ValueString(),
			})
			return true
		}
	}

	return false
}

// aclsChanged compares plan vs state ACLs to detect changes
func (o *StackOperations) aclsChanged(ctx context.Context, planACLs []haproxyAclModel, stateACLs []haproxyAclModel) bool {
	// If counts are different, there's definitely a change
	if len(planACLs) != len(stateACLs) {
		tflog.Info(ctx, "ACLs count changed", map[string]interface{}{
			"plan_count":  len(planACLs),
			"state_count": len(stateACLs),
		})
		return true
	}

	// Compare each ACL
	for i, planACL := range planACLs {
		if i >= len(stateACLs) {
			return true
		}
		stateACL := stateACLs[i]

		// Compare ALL fields
		if planACL.AclName.ValueString() != stateACL.AclName.ValueString() ||
			planACL.Criterion.ValueString() != stateACL.Criterion.ValueString() ||
			planACL.Value.ValueString() != stateACL.Value.ValueString() ||
			planACL.Index.ValueInt64() != stateACL.Index.ValueInt64() {
			tflog.Info(ctx, "ACL changed", map[string]interface{}{
				"acl_index":  i,
				"plan_name":  planACL.AclName.ValueString(),
				"state_name": stateACL.AclName.ValueString(),
			})
			return true
		}
	}

	return false
}

// frontendChanged compares plan vs state frontend to detect changes
func (o *StackOperations) frontendChanged(ctx context.Context, planFrontend *haproxyFrontendModel, stateFrontend *haproxyFrontendModel) bool {
	// If one is nil and the other isn't, there's a change
	if (planFrontend == nil) != (stateFrontend == nil) {
		tflog.Info(ctx, "Frontend nil state changed", map[string]interface{}{
			"plan_nil":  planFrontend == nil,
			"state_nil": stateFrontend == nil,
		})
		return true
	}

	// If both are nil, no change
	if planFrontend == nil && stateFrontend == nil {
		return false
	}

	// Compare ALL basic fields comprehensively
	if planFrontend.Name.ValueString() != stateFrontend.Name.ValueString() ||
		planFrontend.Mode.ValueString() != stateFrontend.Mode.ValueString() ||
		planFrontend.DefaultBackend.ValueString() != stateFrontend.DefaultBackend.ValueString() ||
		planFrontend.Maxconn.ValueInt64() != stateFrontend.Maxconn.ValueInt64() ||
		planFrontend.Backlog.ValueInt64() != stateFrontend.Backlog.ValueInt64() ||
		planFrontend.Ssl.ValueBool() != stateFrontend.Ssl.ValueBool() ||
		planFrontend.SslCertificate.ValueString() != stateFrontend.SslCertificate.ValueString() ||
		planFrontend.SslCafile.ValueString() != stateFrontend.SslCafile.ValueString() ||
		planFrontend.SslMaxVer.ValueString() != stateFrontend.SslMaxVer.ValueString() ||
		planFrontend.SslMinVer.ValueString() != stateFrontend.SslMinVer.ValueString() ||
		planFrontend.Ciphers.ValueString() != stateFrontend.Ciphers.ValueString() ||
		planFrontend.Ciphersuites.ValueString() != stateFrontend.Ciphersuites.ValueString() ||
		planFrontend.Verify.ValueString() != stateFrontend.Verify.ValueString() ||
		planFrontend.AcceptProxy.ValueBool() != stateFrontend.AcceptProxy.ValueBool() ||
		planFrontend.DeferAccept.ValueBool() != stateFrontend.DeferAccept.ValueBool() ||
		planFrontend.TcpUserTimeout.ValueInt64() != stateFrontend.TcpUserTimeout.ValueInt64() ||
		planFrontend.Tfo.ValueBool() != stateFrontend.Tfo.ValueBool() ||
		planFrontend.V4v6.ValueBool() != stateFrontend.V4v6.ValueBool() ||
		planFrontend.V6only.ValueBool() != stateFrontend.V6only.ValueBool() {
		tflog.Info(ctx, "Frontend basic fields changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	// Compare MonitorFail field
	if o.monitorFailChanged(ctx, planFrontend.MonitorFail, stateFrontend.MonitorFail) {
		tflog.Info(ctx, "Frontend MonitorFail changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	// Compare Binds field
	if o.bindsChanged(ctx, planFrontend.Binds, stateFrontend.Binds) {
		tflog.Info(ctx, "Frontend Binds changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	// Compare Acls field
	if o.aclsChanged(ctx, planFrontend.Acls, stateFrontend.Acls) {
		tflog.Info(ctx, "Frontend Acls changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	// Compare HttpRequestRules field
	if o.httpRequestRulesChanged(ctx, planFrontend.HttpRequestRules, stateFrontend.HttpRequestRules) {
		tflog.Info(ctx, "Frontend HttpRequestRules changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	// Compare HttpResponseRules field
	if o.httpResponseRulesChanged(ctx, planFrontend.HttpResponseRules, stateFrontend.HttpResponseRules) {
		tflog.Info(ctx, "Frontend HttpResponseRules changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	// Compare StatsOptions field
	if o.statsOptionsChanged(ctx, planFrontend.StatsOptions, stateFrontend.StatsOptions) {
		tflog.Info(ctx, "Frontend StatsOptions changed", map[string]interface{}{
			"plan_name":  planFrontend.Name.ValueString(),
			"state_name": stateFrontend.Name.ValueString(),
		})
		return true
	}

	return false
}

// monitorFailChanged compares plan vs state monitor fail to detect changes
func (o *StackOperations) monitorFailChanged(ctx context.Context, planMonitorFail []haproxyMonitorFailModel, stateMonitorFail []haproxyMonitorFailModel) bool {
	// If lengths are different, there's a change
	if len(planMonitorFail) != len(stateMonitorFail) {
		return true
	}

	// If both are empty, no change
	if len(planMonitorFail) == 0 && len(stateMonitorFail) == 0 {
		return false
	}

	// Compare each monitor fail entry
	for i, planMF := range planMonitorFail {
		if i >= len(stateMonitorFail) {
			return true
		}
		stateMF := stateMonitorFail[i]

		if planMF.Cond.ValueString() != stateMF.Cond.ValueString() ||
			planMF.CondTest.ValueString() != stateMF.CondTest.ValueString() {
			return true
		}
	}

	return false
}

// statsOptionsChanged compares plan vs state stats options to detect changes
func (o *StackOperations) statsOptionsChanged(ctx context.Context, planStatsOptions []haproxyStatsOptionsModel, stateStatsOptions []haproxyStatsOptionsModel) bool {
	// If lengths are different, there's a change
	if len(planStatsOptions) != len(stateStatsOptions) {
		return true
	}

	// If both are empty, no change
	if len(planStatsOptions) == 0 && len(stateStatsOptions) == 0 {
		return false
	}

	// Compare each stats options entry
	for i, planStats := range planStatsOptions {
		if i >= len(stateStatsOptions) {
			return true
		}
		stateStats := stateStatsOptions[i]

		if planStats.StatsEnable.ValueBool() != stateStats.StatsEnable.ValueBool() ||
			planStats.StatsUri.ValueString() != stateStats.StatsUri.ValueString() ||
			planStats.StatsRealm.ValueString() != stateStats.StatsRealm.ValueString() ||
			planStats.StatsAuth.ValueString() != stateStats.StatsAuth.ValueString() {
			return true
		}
	}

	return false
}

// backendChanged compares plan vs state backend to detect changes
func (o *StackOperations) backendChanged(ctx context.Context, planBackend *haproxyBackendModel, stateBackend *haproxyBackendModel) bool {
	// If one is nil and the other isn't, there's a change
	if (planBackend == nil) != (stateBackend == nil) {
		tflog.Info(ctx, "Backend nil state changed", map[string]interface{}{
			"plan_nil":  planBackend == nil,
			"state_nil": stateBackend == nil,
		})
		return true
	}

	// If both are nil, no change
	if planBackend == nil && stateBackend == nil {
		return false
	}

	// Compare ALL basic fields comprehensively
	if planBackend.Name.ValueString() != stateBackend.Name.ValueString() ||
		planBackend.Mode.ValueString() != stateBackend.Mode.ValueString() ||
		planBackend.AdvCheck.ValueString() != stateBackend.AdvCheck.ValueString() ||
		planBackend.HttpConnectionMode.ValueString() != stateBackend.HttpConnectionMode.ValueString() ||
		planBackend.ServerTimeout.ValueInt64() != stateBackend.ServerTimeout.ValueInt64() ||
		planBackend.CheckTimeout.ValueInt64() != stateBackend.CheckTimeout.ValueInt64() ||
		planBackend.ConnectTimeout.ValueInt64() != stateBackend.ConnectTimeout.ValueInt64() ||
		planBackend.QueueTimeout.ValueInt64() != stateBackend.QueueTimeout.ValueInt64() ||
		planBackend.TunnelTimeout.ValueInt64() != stateBackend.TunnelTimeout.ValueInt64() ||
		planBackend.TarpitTimeout.ValueInt64() != stateBackend.TarpitTimeout.ValueInt64() ||
		planBackend.Checkcache.ValueString() != stateBackend.Checkcache.ValueString() ||
		planBackend.Retries.ValueInt64() != stateBackend.Retries.ValueInt64() {
		tflog.Info(ctx, "Backend basic fields changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare Balance field
	if o.balanceChanged(ctx, planBackend.Balance, stateBackend.Balance) {
		tflog.Info(ctx, "Backend Balance changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare HttpchkParams field
	if o.httpchkParamsChanged(ctx, planBackend.HttpchkParams, stateBackend.HttpchkParams) {
		tflog.Info(ctx, "Backend HttpchkParams changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare Forwardfor field
	if o.forwardforChanged(ctx, planBackend.Forwardfor, stateBackend.Forwardfor) {
		tflog.Info(ctx, "Backend Forwardfor changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare Httpcheck field
	if o.httpcheckChanged(ctx, planBackend.Httpcheck, stateBackend.Httpcheck) {
		tflog.Info(ctx, "Backend Httpcheck changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare TcpCheck field
	if o.tcpCheckChanged(ctx, planBackend.TcpCheck, stateBackend.TcpCheck) {
		tflog.Info(ctx, "Backend TcpCheck changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare Acls field
	if o.aclsChanged(ctx, planBackend.Acls, stateBackend.Acls) {
		tflog.Info(ctx, "Backend Acls changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare HttpRequestRule field
	if o.httpRequestRulesChanged(ctx, planBackend.HttpRequestRule, stateBackend.HttpRequestRule) {
		tflog.Info(ctx, "Backend HttpRequestRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare HttpResponseRule field
	if o.httpResponseRulesChanged(ctx, planBackend.HttpResponseRule, stateBackend.HttpResponseRule) {
		tflog.Info(ctx, "Backend HttpResponseRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare TcpRequestRule field
	if o.tcpRequestRuleChanged(ctx, planBackend.TcpRequestRule, stateBackend.TcpRequestRule) {
		tflog.Info(ctx, "Backend TcpRequestRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare TcpResponseRule field
	if o.tcpResponseRuleChanged(ctx, planBackend.TcpResponseRule, stateBackend.TcpResponseRule) {
		tflog.Info(ctx, "Backend TcpResponseRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare DefaultServer field
	if o.defaultServerChanged(ctx, planBackend.DefaultServer, stateBackend.DefaultServer) {
		tflog.Info(ctx, "Backend DefaultServer changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare StickTable field
	if o.stickTableChanged(ctx, planBackend.StickTable, stateBackend.StickTable) {
		tflog.Info(ctx, "Backend StickTable changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare StickRule field
	if o.stickRuleChanged(ctx, planBackend.StickRule, stateBackend.StickRule) {
		tflog.Info(ctx, "Backend StickRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare StatsOptions field
	if o.statsOptionsChanged(ctx, planBackend.StatsOptions, stateBackend.StatsOptions) {
		tflog.Info(ctx, "Backend StatsOptions changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	return false
}

// balanceChanged compares plan vs state balance to detect changes
func (o *StackOperations) balanceChanged(ctx context.Context, planBalance []haproxyBalanceModel, stateBalance []haproxyBalanceModel) bool {
	// If lengths are different, there's a change
	if len(planBalance) != len(stateBalance) {
		return true
	}

	// If both are empty, no change
	if len(planBalance) == 0 && len(stateBalance) == 0 {
		return false
	}

	// Compare each balance entry
	for i, planBal := range planBalance {
		if i >= len(stateBalance) {
			return true
		}
		stateBal := stateBalance[i]

		if planBal.Algorithm.ValueString() != stateBal.Algorithm.ValueString() ||
			planBal.UrlParam.ValueString() != stateBal.UrlParam.ValueString() {
			return true
		}
	}

	return false
}

// httpchkParamsChanged compares plan vs state httpchk params to detect changes
func (o *StackOperations) httpchkParamsChanged(ctx context.Context, planHttpchkParams []haproxyHttpchkParamsModel, stateHttpchkParams []haproxyHttpchkParamsModel) bool {
	// If lengths are different, there's a change
	if len(planHttpchkParams) != len(stateHttpchkParams) {
		return true
	}

	// If both are empty, no change
	if len(planHttpchkParams) == 0 && len(stateHttpchkParams) == 0 {
		return false
	}

	// Compare each httpchk params entry
	for i, planParams := range planHttpchkParams {
		if i >= len(stateHttpchkParams) {
			return true
		}
		stateParams := stateHttpchkParams[i]

		if planParams.Method.ValueString() != stateParams.Method.ValueString() ||
			planParams.Uri.ValueString() != stateParams.Uri.ValueString() ||
			planParams.Version.ValueString() != stateParams.Version.ValueString() {
			return true
		}
	}

	return false
}

// forwardforChanged compares plan vs state forwardfor to detect changes
func (o *StackOperations) forwardforChanged(ctx context.Context, planForwardfor []haproxyForwardforModel, stateForwardfor []haproxyForwardforModel) bool {
	// If lengths are different, there's a change
	if len(planForwardfor) != len(stateForwardfor) {
		return true
	}

	// If both are empty, no change
	if len(planForwardfor) == 0 && len(stateForwardfor) == 0 {
		return false
	}

	// Compare each forwardfor entry
	for i, planFF := range planForwardfor {
		if i >= len(stateForwardfor) {
			return true
		}
		stateFF := stateForwardfor[i]

		if planFF.Enabled.ValueString() != stateFF.Enabled.ValueString() {
			return true
		}
	}

	return false
}

// httpcheckChanged compares plan vs state httpcheck to detect changes
func (o *StackOperations) httpcheckChanged(ctx context.Context, planHttpcheck []haproxyHttpcheckModel, stateHttpcheck []haproxyHttpcheckModel) bool {
	// If lengths are different, there's a change
	if len(planHttpcheck) != len(stateHttpcheck) {
		return true
	}

	// If both are empty, no change
	if len(planHttpcheck) == 0 && len(stateHttpcheck) == 0 {
		return false
	}

	// Compare each httpcheck entry
	for i, planCheck := range planHttpcheck {
		if i >= len(stateHttpcheck) {
			return true
		}
		stateCheck := stateHttpcheck[i]

		// Compare ALL fields from haproxyHttpcheckModel comprehensively
		if planCheck.Index.ValueInt64() != stateCheck.Index.ValueInt64() ||
			planCheck.Type.ValueString() != stateCheck.Type.ValueString() ||
			planCheck.Method.ValueString() != stateCheck.Method.ValueString() ||
			planCheck.Uri.ValueString() != stateCheck.Uri.ValueString() ||
			planCheck.Version.ValueString() != stateCheck.Version.ValueString() ||
			planCheck.Timeout.ValueInt64() != stateCheck.Timeout.ValueInt64() ||
			planCheck.Match.ValueString() != stateCheck.Match.ValueString() ||
			planCheck.Pattern.ValueString() != stateCheck.Pattern.ValueString() ||
			planCheck.Addr.ValueString() != stateCheck.Addr.ValueString() ||
			planCheck.Port.ValueInt64() != stateCheck.Port.ValueInt64() ||
			planCheck.ExclamationMark.ValueString() != stateCheck.ExclamationMark.ValueString() ||
			planCheck.LogLevel.ValueString() != stateCheck.LogLevel.ValueString() ||
			planCheck.SendProxy.ValueString() != stateCheck.SendProxy.ValueString() ||
			planCheck.ViaSocks4.ValueString() != stateCheck.ViaSocks4.ValueString() ||
			planCheck.CheckComment.ValueString() != stateCheck.CheckComment.ValueString() {
			return true
		}
	}

	return false
}

// tcpCheckChanged compares plan vs state tcp check to detect changes
func (o *StackOperations) tcpCheckChanged(ctx context.Context, planTcpCheck []haproxyTcpCheckModel, stateTcpCheck []haproxyTcpCheckModel) bool {
	// If lengths are different, there's a change
	if len(planTcpCheck) != len(stateTcpCheck) {
		return true
	}

	// If both are empty, no change
	if len(planTcpCheck) == 0 && len(stateTcpCheck) == 0 {
		return false
	}

	// Compare each tcp check entry
	for i, planCheck := range planTcpCheck {
		if i >= len(stateTcpCheck) {
			return true
		}
		stateCheck := stateTcpCheck[i]

		if planCheck.Index.ValueInt64() != stateCheck.Index.ValueInt64() ||
			planCheck.Type.ValueString() != stateCheck.Type.ValueString() ||
			planCheck.Action.ValueString() != stateCheck.Action.ValueString() ||
			planCheck.Cond.ValueString() != stateCheck.Cond.ValueString() ||
			planCheck.CondTest.ValueString() != stateCheck.CondTest.ValueString() {
			return true
		}
	}

	return false
}

// tcpRequestRuleChanged compares plan vs state tcp request rule to detect changes
func (o *StackOperations) tcpRequestRuleChanged(ctx context.Context, planTcpRequestRule []haproxyTcpRequestRuleModel, stateTcpRequestRule []haproxyTcpRequestRuleModel) bool {
	// If lengths are different, there's a change
	if len(planTcpRequestRule) != len(stateTcpRequestRule) {
		return true
	}

	// If both are empty, no change
	if len(planTcpRequestRule) == 0 && len(stateTcpRequestRule) == 0 {
		return false
	}

	// Compare each tcp request rule entry
	for i, planRule := range planTcpRequestRule {
		if i >= len(stateTcpRequestRule) {
			return true
		}
		stateRule := stateTcpRequestRule[i]

		if planRule.Index.ValueInt64() != stateRule.Index.ValueInt64() ||
			planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Action.ValueString() != stateRule.Action.ValueString() ||
			planRule.Cond.ValueString() != stateRule.Cond.ValueString() ||
			planRule.CondTest.ValueString() != stateRule.CondTest.ValueString() {
			return true
		}
	}

	return false
}

// tcpResponseRuleChanged compares plan vs state tcp response rule to detect changes
func (o *StackOperations) tcpResponseRuleChanged(ctx context.Context, planTcpResponseRule []haproxyTcpResponseRuleModel, stateTcpResponseRule []haproxyTcpResponseRuleModel) bool {
	// If lengths are different, there's a change
	if len(planTcpResponseRule) != len(stateTcpResponseRule) {
		return true
	}

	// If both are empty, no change
	if len(planTcpResponseRule) == 0 && len(stateTcpResponseRule) == 0 {
		return false
	}

	// Compare each tcp response rule entry
	for i, planRule := range planTcpResponseRule {
		if i >= len(stateTcpResponseRule) {
			return true
		}
		stateRule := stateTcpResponseRule[i]

		if planRule.Index.ValueInt64() != stateRule.Index.ValueInt64() ||
			planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Action.ValueString() != stateRule.Action.ValueString() ||
			planRule.Cond.ValueString() != stateRule.Cond.ValueString() ||
			planRule.CondTest.ValueString() != stateRule.CondTest.ValueString() {
			return true
		}
	}

	return false
}

// defaultServerChanged compares plan vs state default server to detect changes
func (o *StackOperations) defaultServerChanged(ctx context.Context, planDefaultServer *haproxyDefaultServerModel, stateDefaultServer *haproxyDefaultServerModel) bool {
	// If one is nil and the other isn't, there's a change
	if (planDefaultServer == nil) != (stateDefaultServer == nil) {
		return true
	}

	// If both are nil, no change
	if planDefaultServer == nil && stateDefaultServer == nil {
		return false
	}

	// Compare all default server fields
	if planDefaultServer.Ssl.ValueString() != stateDefaultServer.Ssl.ValueString() ||
		planDefaultServer.Verify.ValueString() != stateDefaultServer.Verify.ValueString() ||
		planDefaultServer.SslCafile.ValueString() != stateDefaultServer.SslCafile.ValueString() ||
		planDefaultServer.SslCertificate.ValueString() != stateDefaultServer.SslCertificate.ValueString() ||
		planDefaultServer.SslMaxVer.ValueString() != stateDefaultServer.SslMaxVer.ValueString() ||
		planDefaultServer.SslMinVer.ValueString() != stateDefaultServer.SslMinVer.ValueString() ||
		planDefaultServer.Ciphers.ValueString() != stateDefaultServer.Ciphers.ValueString() ||
		planDefaultServer.Ciphersuites.ValueString() != stateDefaultServer.Ciphersuites.ValueString() ||
		planDefaultServer.Sslv3.ValueString() != stateDefaultServer.Sslv3.ValueString() ||
		planDefaultServer.Tlsv10.ValueString() != stateDefaultServer.Tlsv10.ValueString() ||
		planDefaultServer.Tlsv11.ValueString() != stateDefaultServer.Tlsv11.ValueString() ||
		planDefaultServer.Tlsv12.ValueString() != stateDefaultServer.Tlsv12.ValueString() ||
		planDefaultServer.Tlsv13.ValueString() != stateDefaultServer.Tlsv13.ValueString() ||
		planDefaultServer.NoSslv3.ValueString() != stateDefaultServer.NoSslv3.ValueString() ||
		planDefaultServer.NoTlsv10.ValueString() != stateDefaultServer.NoTlsv10.ValueString() ||
		planDefaultServer.NoTlsv11.ValueString() != stateDefaultServer.NoTlsv11.ValueString() ||
		planDefaultServer.NoTlsv12.ValueString() != stateDefaultServer.NoTlsv12.ValueString() ||
		planDefaultServer.NoTlsv13.ValueString() != stateDefaultServer.NoTlsv13.ValueString() ||
		planDefaultServer.ForceSslv3.ValueString() != stateDefaultServer.ForceSslv3.ValueString() ||
		planDefaultServer.ForceTlsv10.ValueString() != stateDefaultServer.ForceTlsv10.ValueString() ||
		planDefaultServer.ForceTlsv11.ValueString() != stateDefaultServer.ForceTlsv11.ValueString() ||
		planDefaultServer.ForceTlsv12.ValueString() != stateDefaultServer.ForceTlsv12.ValueString() ||
		planDefaultServer.ForceTlsv13.ValueString() != stateDefaultServer.ForceTlsv13.ValueString() ||
		planDefaultServer.ForceStrictSni.ValueString() != stateDefaultServer.ForceStrictSni.ValueString() ||
		planDefaultServer.SslReuse.ValueString() != stateDefaultServer.SslReuse.ValueString() {
		return true
	}

	return false
}

// stickTableChanged compares plan vs state stick table to detect changes
func (o *StackOperations) stickTableChanged(ctx context.Context, planStickTable *haproxyStickTableModel, stateStickTable *haproxyStickTableModel) bool {
	// If one is nil and the other isn't, there's a change
	if (planStickTable == nil) != (stateStickTable == nil) {
		return true
	}

	// If both are nil, no change
	if planStickTable == nil && stateStickTable == nil {
		return false
	}

	// Compare all stick table fields
	if planStickTable.Type.ValueString() != stateStickTable.Type.ValueString() ||
		planStickTable.Size.ValueInt64() != stateStickTable.Size.ValueInt64() ||
		planStickTable.Expire.ValueInt64() != stateStickTable.Expire.ValueInt64() ||
		planStickTable.Nopurge.ValueBool() != stateStickTable.Nopurge.ValueBool() ||
		planStickTable.Peers.ValueString() != stateStickTable.Peers.ValueString() {
		return true
	}

	return false
}

// stickRuleChanged compares plan vs state stick rule to detect changes
func (o *StackOperations) stickRuleChanged(ctx context.Context, planStickRule []haproxyStickRuleModel, stateStickRule []haproxyStickRuleModel) bool {
	// If lengths are different, there's a change
	if len(planStickRule) != len(stateStickRule) {
		return true
	}

	// If both are empty, no change
	if len(planStickRule) == 0 && len(stateStickRule) == 0 {
		return false
	}

	// Compare each stick rule entry
	for i, planRule := range planStickRule {
		if i >= len(stateStickRule) {
			return true
		}
		stateRule := stateStickRule[i]

		if planRule.Index.ValueInt64() != stateRule.Index.ValueInt64() ||
			planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Table.ValueString() != stateRule.Table.ValueString() ||
			planRule.Pattern.ValueString() != stateRule.Pattern.ValueString() {
			return true
		}
	}

	return false
}

// bindsChanged compares plan vs state binds to detect changes
func (o *StackOperations) bindsChanged(ctx context.Context, planBinds map[string]haproxyBindModel, stateBinds map[string]haproxyBindModel) bool {
	// If counts are different, there's definitely a change
	if len(planBinds) != len(stateBinds) {
		tflog.Info(ctx, "Binds count changed", map[string]interface{}{
			"plan_count":  len(planBinds),
			"state_count": len(stateBinds),
		})
		return true
	}

	// Compare each bind by name
	for bindName, planBind := range planBinds {
		stateBind, exists := stateBinds[bindName]
		if !exists {
			tflog.Info(ctx, "Bind added", map[string]interface{}{"bind_name": bindName})
			return true
		}

		// Compare ALL fields from haproxyBindModel comprehensively
		if planBind.Address.ValueString() != stateBind.Address.ValueString() ||
			planBind.Port.ValueInt64() != stateBind.Port.ValueInt64() ||
			planBind.PortRangeEnd.ValueInt64() != stateBind.PortRangeEnd.ValueInt64() ||
			planBind.Transparent.ValueBool() != stateBind.Transparent.ValueBool() ||
			planBind.Mode.ValueString() != stateBind.Mode.ValueString() ||
			planBind.Maxconn.ValueInt64() != stateBind.Maxconn.ValueInt64() ||
			planBind.Ssl.ValueBool() != stateBind.Ssl.ValueBool() ||
			planBind.SslCafile.ValueString() != stateBind.SslCafile.ValueString() ||
			planBind.SslCertificate.ValueString() != stateBind.SslCertificate.ValueString() ||
			planBind.SslMaxVer.ValueString() != stateBind.SslMaxVer.ValueString() ||
			planBind.SslMinVer.ValueString() != stateBind.SslMinVer.ValueString() ||
			planBind.Ciphers.ValueString() != stateBind.Ciphers.ValueString() ||
			planBind.Ciphersuites.ValueString() != stateBind.Ciphersuites.ValueString() ||
			planBind.Verify.ValueString() != stateBind.Verify.ValueString() ||
			planBind.AcceptProxy.ValueBool() != stateBind.AcceptProxy.ValueBool() ||
			planBind.Allow0rtt.ValueBool() != stateBind.Allow0rtt.ValueBool() ||
			planBind.Alpn.ValueString() != stateBind.Alpn.ValueString() ||
			planBind.Backlog.ValueString() != stateBind.Backlog.ValueString() ||
			planBind.DeferAccept.ValueBool() != stateBind.DeferAccept.ValueBool() ||
			planBind.GenerateCertificates.ValueBool() != stateBind.GenerateCertificates.ValueBool() ||
			planBind.Gid.ValueInt64() != stateBind.Gid.ValueInt64() ||
			planBind.Group.ValueString() != stateBind.Group.ValueString() ||
			planBind.Id.ValueString() != stateBind.Id.ValueString() ||
			planBind.Interface.ValueString() != stateBind.Interface.ValueString() ||
			planBind.Level.ValueString() != stateBind.Level.ValueString() ||
			planBind.Namespace.ValueString() != stateBind.Namespace.ValueString() ||
			planBind.Nice.ValueInt64() != stateBind.Nice.ValueInt64() ||
			planBind.NoCaNames.ValueBool() != stateBind.NoCaNames.ValueBool() ||
			planBind.Npn.ValueString() != stateBind.Npn.ValueString() ||
			planBind.PreferClientCiphers.ValueBool() != stateBind.PreferClientCiphers.ValueBool() ||
			planBind.Process.ValueString() != stateBind.Process.ValueString() ||
			planBind.Proto.ValueString() != stateBind.Proto.ValueString() ||
			planBind.SeverityOutput.ValueString() != stateBind.SeverityOutput.ValueString() ||
			planBind.StrictSni.ValueBool() != stateBind.StrictSni.ValueBool() ||
			planBind.TcpUserTimeout.ValueInt64() != stateBind.TcpUserTimeout.ValueInt64() ||
			planBind.Tfo.ValueBool() != stateBind.Tfo.ValueBool() ||
			planBind.TlsTicketKeys.ValueString() != stateBind.TlsTicketKeys.ValueString() ||
			planBind.Uid.ValueString() != stateBind.Uid.ValueString() ||
			planBind.User.ValueString() != stateBind.User.ValueString() ||
			planBind.V4v6.ValueBool() != stateBind.V4v6.ValueBool() ||
			planBind.V6only.ValueBool() != stateBind.V6only.ValueBool() ||
			// v3 fields (non-deprecated)
			planBind.Sslv3.ValueBool() != stateBind.Sslv3.ValueBool() ||
			planBind.Tlsv10.ValueBool() != stateBind.Tlsv10.ValueBool() ||
			planBind.Tlsv11.ValueBool() != stateBind.Tlsv11.ValueBool() ||
			planBind.Tlsv12.ValueBool() != stateBind.Tlsv12.ValueBool() ||
			planBind.Tlsv13.ValueBool() != stateBind.Tlsv13.ValueBool() ||
			planBind.TlsTickets.ValueString() != stateBind.TlsTickets.ValueString() ||
			planBind.ForceStrictSni.ValueString() != stateBind.ForceStrictSni.ValueString() ||
			planBind.NoStrictSni.ValueBool() != stateBind.NoStrictSni.ValueBool() ||
			planBind.GuidPrefix.ValueString() != stateBind.GuidPrefix.ValueString() ||
			planBind.IdlePing.ValueInt64() != stateBind.IdlePing.ValueInt64() ||
			planBind.QuicCcAlgo.ValueString() != stateBind.QuicCcAlgo.ValueString() ||
			planBind.QuicForceRetry.ValueBool() != stateBind.QuicForceRetry.ValueBool() ||
			planBind.QuicSocket.ValueString() != stateBind.QuicSocket.ValueString() ||
			planBind.QuicCcAlgoBurstSize.ValueInt64() != stateBind.QuicCcAlgoBurstSize.ValueInt64() ||
			planBind.QuicCcAlgoMaxWindow.ValueInt64() != stateBind.QuicCcAlgoMaxWindow.ValueInt64() ||
			planBind.Metadata.ValueString() != stateBind.Metadata.ValueString() ||
			// v2 fields (deprecated in v3)
			planBind.NoSslv3.ValueBool() != stateBind.NoSslv3.ValueBool() ||
			planBind.ForceSslv3.ValueBool() != stateBind.ForceSslv3.ValueBool() ||
			planBind.ForceTlsv10.ValueBool() != stateBind.ForceTlsv10.ValueBool() ||
			planBind.ForceTlsv11.ValueBool() != stateBind.ForceTlsv11.ValueBool() ||
			planBind.ForceTlsv12.ValueBool() != stateBind.ForceTlsv12.ValueBool() ||
			planBind.ForceTlsv13.ValueBool() != stateBind.ForceTlsv13.ValueBool() ||
			planBind.NoTlsv10.ValueBool() != stateBind.NoTlsv10.ValueBool() ||
			planBind.NoTlsv11.ValueBool() != stateBind.NoTlsv11.ValueBool() ||
			planBind.NoTlsv12.ValueBool() != stateBind.NoTlsv12.ValueBool() ||
			planBind.NoTlsv13.ValueBool() != stateBind.NoTlsv13.ValueBool() ||
			planBind.NoTlsTickets.ValueBool() != stateBind.NoTlsTickets.ValueBool() {
			tflog.Info(ctx, "Bind changed", map[string]interface{}{
				"bind_name":     bindName,
				"plan_address":  planBind.Address.ValueString(),
				"state_address": stateBind.Address.ValueString(),
			})
			return true
		}
	}

	// Check for removed binds
	for bindName := range stateBinds {
		if _, exists := planBinds[bindName]; !exists {
			tflog.Info(ctx, "Bind removed", map[string]interface{}{"bind_name": bindName})
			return true
		}
	}

	return false
}

// serversChanged compares plan vs state servers to detect changes
func (o *StackOperations) serversChanged(ctx context.Context, planServers map[string]haproxyServerModel, stateServers map[string]haproxyServerModel) bool {
	// If counts are different, there's definitely a change
	if len(planServers) != len(stateServers) {
		tflog.Info(ctx, "Servers count changed", map[string]interface{}{
			"plan_count":  len(planServers),
			"state_count": len(stateServers),
		})
		return true
	}

	// Compare each server by name
	for serverName, planServer := range planServers {
		stateServer, exists := stateServers[serverName]
		if !exists {
			tflog.Info(ctx, "Server added", map[string]interface{}{"server_name": serverName})
			return true
		}

		// Compare ALL fields comprehensively
		if planServer.Address.ValueString() != stateServer.Address.ValueString() ||
			planServer.Port.ValueInt64() != stateServer.Port.ValueInt64() ||
			planServer.Check.ValueString() != stateServer.Check.ValueString() ||
			planServer.Backup.ValueString() != stateServer.Backup.ValueString() ||
			planServer.Maxconn.ValueInt64() != stateServer.Maxconn.ValueInt64() ||
			planServer.Weight.ValueInt64() != stateServer.Weight.ValueInt64() ||
			planServer.Rise.ValueInt64() != stateServer.Rise.ValueInt64() ||
			planServer.Fall.ValueInt64() != stateServer.Fall.ValueInt64() ||
			planServer.Inter.ValueInt64() != stateServer.Inter.ValueInt64() ||
			planServer.Fastinter.ValueInt64() != stateServer.Fastinter.ValueInt64() ||
			planServer.Downinter.ValueInt64() != stateServer.Downinter.ValueInt64() ||
			planServer.Ssl.ValueString() != stateServer.Ssl.ValueString() ||
			planServer.Verify.ValueString() != stateServer.Verify.ValueString() ||
			planServer.Cookie.ValueString() != stateServer.Cookie.ValueString() ||
			planServer.Sslv3.ValueString() != stateServer.Sslv3.ValueString() ||
			planServer.Tlsv10.ValueString() != stateServer.Tlsv10.ValueString() ||
			planServer.Tlsv11.ValueString() != stateServer.Tlsv11.ValueString() ||
			planServer.Tlsv12.ValueString() != stateServer.Tlsv12.ValueString() ||
			planServer.Tlsv13.ValueString() != stateServer.Tlsv13.ValueString() ||
			planServer.NoSslv3.ValueString() != stateServer.NoSslv3.ValueString() ||
			planServer.NoTlsv10.ValueString() != stateServer.NoTlsv10.ValueString() ||
			planServer.NoTlsv11.ValueString() != stateServer.NoTlsv11.ValueString() ||
			planServer.NoTlsv12.ValueString() != stateServer.NoTlsv12.ValueString() ||
			planServer.NoTlsv13.ValueString() != stateServer.NoTlsv13.ValueString() ||
			planServer.ForceSslv3.ValueString() != stateServer.ForceSslv3.ValueString() ||
			planServer.ForceTlsv10.ValueString() != stateServer.ForceTlsv10.ValueString() ||
			planServer.ForceTlsv11.ValueString() != stateServer.ForceTlsv11.ValueString() ||
			planServer.ForceTlsv12.ValueString() != stateServer.ForceTlsv12.ValueString() ||
			planServer.ForceTlsv13.ValueString() != stateServer.ForceTlsv13.ValueString() ||
			planServer.ForceStrictSni.ValueString() != stateServer.ForceStrictSni.ValueString() {
			tflog.Info(ctx, "Server changed", map[string]interface{}{
				"server_name":   serverName,
				"plan_address":  planServer.Address.ValueString(),
				"state_address": stateServer.Address.ValueString(),
			})
			return true
		}
	}

	// Check for removed servers
	for serverName := range stateServers {
		if _, exists := planServers[serverName]; !exists {
			tflog.Info(ctx, "Server removed", map[string]interface{}{"server_name": serverName})
			return true
		}
	}

	return false
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

	// Delete ACLs if specified - handle both frontend and backend ACLs
	if data.Frontend != nil && data.Frontend.Acls != nil && len(data.Frontend.Acls) > 0 {
		tflog.Info(ctx, "Deleting frontend ACLs", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.aclManager.DeleteACLsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting frontend ACLs", err.Error())
			return err
		}
	}

	if data.Backend != nil && data.Backend.Acls != nil && len(data.Backend.Acls) > 0 {
		tflog.Info(ctx, "Deleting backend ACLs", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.aclManager.DeleteACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting backend ACLs", err.Error())
			return err
		}
	}

	// Delete HTTP Request Rules if specified
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		tflog.Info(ctx, "Deleting HTTP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.DeleteHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting HTTP request rules", err.Error())
			return err
		}
	}

	// Delete HTTP Response Rules if specified
	if data.Frontend != nil && data.Frontend.HttpResponseRules != nil && len(data.Frontend.HttpResponseRules) > 0 {
		tflog.Info(ctx, "Deleting HTTP response rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpResponseRuleManager.DeleteHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error deleting HTTP response rules", err.Error())
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
