package haproxy

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"terraform-provider-haproxy/haproxy/utils"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Global mutex for all HAProxy transactions
var globalTransactionMutex sync.Mutex

// StackOperations handles all CRUD operations for the haproxy_stack resource
type StackOperations struct {
	client                  *HAProxyClient
	aclManager              *ACLManager
	frontendManager         *FrontendManager
	backendManager          *BackendManager
	httpRequestRuleManager  *HttpRequestRuleManager
	httpResponseRuleManager *HttpResponseRuleManager
	tcpRequestRuleManager   *TcpRequestRuleManager
	tcpResponseRuleManager  *TcpResponseRuleManager
	httpcheckManager        *HttpcheckManager
	tcpCheckManager         *TcpCheckManager
	bindManager             *BindManager
}

// CreateStackOperations creates a new StackOperations instance
func CreateStackOperations(client *HAProxyClient, aclManager *ACLManager, frontendManager *FrontendManager, backendManager *BackendManager, httpRequestRuleManager *HttpRequestRuleManager, httpResponseRuleManager *HttpResponseRuleManager, tcpRequestRuleManager *TcpRequestRuleManager, tcpResponseRuleManager *TcpResponseRuleManager, httpcheckManager *HttpcheckManager, tcpCheckManager *TcpCheckManager, bindManager *BindManager) *StackOperations {
	stackOps := &StackOperations{
		client:                  client,
		aclManager:              aclManager,
		backendManager:          backendManager,
		frontendManager:         frontendManager,
		httpRequestRuleManager:  httpRequestRuleManager,
		httpResponseRuleManager: httpResponseRuleManager,
		tcpRequestRuleManager:   tcpRequestRuleManager,
		tcpResponseRuleManager:  tcpResponseRuleManager,
		httpcheckManager:        httpcheckManager,
		tcpCheckManager:         tcpCheckManager,
		bindManager:             bindManager,
	}

	return stackOps
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
	// SSL certificate fields - HAProxy doesn't return these, so set to null
	// Terraform will manage these values from your configuration
	model.SslCertificate = types.StringNull()
	model.SslCafile = types.StringNull()
	model.SslMaxVer = types.StringNull()
	model.SslMinVer = types.StringNull()
	if server.Verify != "" {
		model.Verify = types.StringValue(server.Verify)
	}
	if server.Cookie != "" {
		model.Cookie = types.StringValue(server.Cookie)
	}
	// HAProxy doesn't support server disabling - field ignored
	// This field has been removed from the schema

	// SSL/TLS Protocol Control (v3 fields) - only set if explicitly configured
	// Don't set default values returned by HAProxy to avoid unwanted changes
	// These fields should only be set if they were explicitly configured by the user
	// HAProxy returns "enabled" as default, but we don't want to manage that
	// For now, we'll set them to null to avoid managing default values
	model.Sslv3 = types.StringNull()
	model.Tlsv10 = types.StringNull()
	model.Tlsv11 = types.StringNull()
	model.Tlsv12 = types.StringNull()
	model.Tlsv13 = types.StringNull()

	// SSL/TLS Protocol Control (deprecated v2 fields) - only set if explicitly configured
	// Don't set default values returned by HAProxy to avoid unwanted changes
	// These fields should only be set if they were explicitly configured by the user
	// HAProxy returns "enabled" as default, but we don't want to manage that
	// For now, we'll set them to null to avoid managing default values
	model.NoSslv3 = types.StringNull()
	model.NoTlsv10 = types.StringNull()
	model.NoTlsv11 = types.StringNull()
	model.NoTlsv12 = types.StringNull()
	model.NoTlsv13 = types.StringNull()

	// Force TLS fields - HAProxy doesn't return these, so set to null
	// Terraform will manage these values from your configuration
	model.ForceSslv3 = types.StringNull()
	model.ForceTlsv10 = types.StringNull()
	model.ForceTlsv11 = types.StringNull()
	model.ForceTlsv12 = types.StringNull()
	model.ForceTlsv13 = types.StringNull()

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
	if !server.SslCertificate.IsNull() && !server.SslCertificate.IsUnknown() {
		payload.SslCertificate = server.SslCertificate.ValueString()
	}
	if !server.SslCafile.IsNull() && !server.SslCafile.IsUnknown() {
		payload.SslCafile = server.SslCafile.ValueString()
	}
	if !server.SslMaxVer.IsNull() && !server.SslMaxVer.IsUnknown() {
		payload.SslMaxVer = server.SslMaxVer.ValueString()
	}
	if !server.SslMinVer.IsNull() && !server.SslMinVer.IsUnknown() {
		payload.SslMinVer = server.SslMinVer.ValueString()
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
	// Only send force_tlsv* fields when explicitly set to "enabled"
	if !server.ForceSslv3.IsNull() && !server.ForceSslv3.IsUnknown() && server.ForceSslv3.ValueString() == "enabled" {
		payload.ForceSslv3 = "enabled"
	}
	if !server.ForceTlsv10.IsNull() && !server.ForceTlsv10.IsUnknown() && server.ForceTlsv10.ValueString() == "enabled" {
		payload.ForceTlsv10 = "enabled"
	}
	if !server.ForceTlsv11.IsNull() && !server.ForceTlsv11.IsUnknown() && server.ForceTlsv11.ValueString() == "enabled" {
		payload.ForceTlsv11 = "enabled"
	}
	if !server.ForceTlsv12.IsNull() && !server.ForceTlsv12.IsUnknown() && server.ForceTlsv12.ValueString() == "enabled" {
		payload.ForceTlsv12 = "enabled"
	}
	if !server.ForceTlsv13.IsNull() && !server.ForceTlsv13.IsUnknown() && server.ForceTlsv13.ValueString() == "enabled" {
		payload.ForceTlsv13 = "enabled"
	}

	return payload
}

// Create performs the create operation for the haproxy_stack resource
func (o *StackOperations) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *haproxyStackResourceModel) error {
	// Serialize all HAProxy operations to prevent transaction conflicts
	globalTransactionMutex.Lock()
	defer globalTransactionMutex.Unlock()

	return o.createSingle(ctx, req, resp, data)
}

// isTransactionRetryableError checks if an error is retryable for transaction operations
func (o *StackOperations) isTransactionRetryableError(err error) bool {
	// Debug logging
	tflog.Info(context.Background(), "Checking if error is retryable", map[string]interface{}{
		"error_type": fmt.Sprintf("%T", err),
		"error_msg":  err.Error(),
	})

	// Check for CustomError first
	if customErr, ok := err.(*utils.CustomError); ok && customErr.APIError != nil {
		tflog.Info(context.Background(), "Found CustomError", map[string]interface{}{
			"code":    customErr.APIError.Code,
			"message": customErr.APIError.Message,
		})
		// Check for transaction does not exist (400)
		if customErr.APIError.Code == 400 && strings.Contains(customErr.APIError.Message, "transaction does not exist") {
			tflog.Info(context.Background(), "Detected retryable CustomError: transaction does not exist")
			return true
		}
		// Check for transaction outdated (406)
		if customErr.APIError.Code == 406 && strings.Contains(customErr.APIError.Message, "transaction") && strings.Contains(customErr.APIError.Message, "is outdated and cannot be committed") {
			tflog.Info(context.Background(), "Detected retryable CustomError: transaction outdated")
			return true
		}
		// Check for version mismatch (409)
		if customErr.APIError.Code == 409 && strings.Contains(customErr.APIError.Message, "version mismatch") {
			tflog.Info(context.Background(), "Detected retryable CustomError: version mismatch")
			return true
		}
		// Check for version or transaction not specified (400)
		if customErr.APIError.Code == 400 && strings.Contains(customErr.APIError.Message, "version or transaction not specified") {
			tflog.Info(context.Background(), "Detected retryable CustomError: version or transaction not specified")
			return true
		}
	}

	// Check for regular errors that contain retryable error messages
	errStr := err.Error()
	tflog.Info(context.Background(), "Checking error string for retryable patterns", map[string]interface{}{
		"error_string":                        errStr,
		"contains_transaction_does_not_exist": strings.Contains(errStr, "transaction does not exist"),
		"contains_transaction_outdated":       strings.Contains(errStr, "transaction") && strings.Contains(errStr, "is outdated and cannot be committed"),
		"contains_version_mismatch":           strings.Contains(errStr, "version mismatch"),
		"contains_version_not_specified":      strings.Contains(errStr, "version or transaction not specified"),
	})

	if strings.Contains(errStr, "transaction does not exist") ||
		strings.Contains(errStr, "transaction") && strings.Contains(errStr, "is outdated and cannot be committed") ||
		strings.Contains(errStr, "version mismatch") ||
		strings.Contains(errStr, "version or transaction not specified") {
		tflog.Info(context.Background(), "Detected retryable error from string matching")
		return true
	}

	tflog.Info(context.Background(), "Error is not retryable")
	return false
}

// createSingle performs a single create operation with transaction retry logic
func (o *StackOperations) createSingle(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *haproxyStackResourceModel) error {
	// Retry the entire operation if transaction becomes outdated
	for {
		err := o.createSingleInternal(ctx, req, resp, data)
		if err == nil {
			return nil
		}

		// Check if this is a retryable transaction error
		if o.isTransactionRetryableError(err) {
			tflog.Info(ctx, "Transaction outdated, retrying entire operation", map[string]interface{}{"error": err.Error()})
			continue
		}

		// Non-retryable error, return it
		return err
	}
}

// createSingleInternal performs the actual create operation without retry
func (o *StackOperations) createSingleInternal(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Creating HAProxy stack")

	// Begin a single transaction for all resources
	tflog.Info(ctx, "Beginning single transaction for all resources")
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
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
			return fmt.Errorf("error creating backend: %w", err)
		}
		tflog.Info(ctx, "Backend created successfully in transaction", map[string]interface{}{"transaction_id": transactionID})
	}

	// Create servers if specified
	if data.Backend != nil && len(data.Backend.Servers) > 0 {
		for serverName, server := range data.Backend.Servers {
			serverPayload := o.convertServerModelToPayload(serverName, server)
			tflog.Info(ctx, "Creating server", map[string]interface{}{
				"server_name":  serverName,
				"backend_name": data.Backend.Name.ValueString(),
			})
			if err := o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
				return fmt.Errorf("error creating server %s: %w", serverName, err)
			}
		}
	}

	// Create frontend if specified
	if data.Frontend != nil {
		if err := o.frontendManager.CreateFrontendInTransaction(ctx, transactionID, data.Frontend); err != nil {
			return fmt.Errorf("error creating frontend: %w", err)
		}
	}

	// Create binds for frontend if specified
	if data.Frontend != nil && data.Frontend.Binds != nil && len(data.Frontend.Binds) > 0 {
		if err := o.bindManager.CreateBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Binds); err != nil {
			return fmt.Errorf("error creating binds: %w", err)
		}
	}

	// Create ACLs if specified - handle both frontend and backend ACLs
	if data.Frontend != nil && data.Frontend.Acls != nil && len(data.Frontend.Acls) > 0 {
		tflog.Info(ctx, "Creating frontend ACLs in transaction", map[string]interface{}{"transaction_id": transactionID})
		if err := o.aclManager.CreateACLsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.Acls); err != nil {
			return fmt.Errorf("error creating frontend ACLs: %w", err)
		}
	}

	if data.Backend != nil && data.Backend.Acls != nil && len(data.Backend.Acls) > 0 {
		tflog.Info(ctx, "Creating backend ACLs in transaction", map[string]interface{}{"transaction_id": transactionID})
		if err := o.aclManager.CreateACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.Acls); err != nil {
			return fmt.Errorf("error creating backend ACLs: %w", err)
		}
	}

	// Create HTTP Request Rules AFTER ACLs (so they can reference existing ACLs)
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		if err := o.httpRequestRuleManager.CreateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
			return fmt.Errorf("error creating HTTP request rules: %w", err)
		}
	}

	// Create Backend HTTP Request Rules AFTER ACLs (so they can reference existing ACLs)
	if data.Backend != nil && data.Backend.HttpRequestRules != nil && len(data.Backend.HttpRequestRules) > 0 {
		if err := o.httpRequestRuleManager.CreateHttpRequestRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.HttpRequestRules); err != nil {
			return fmt.Errorf("error creating backend HTTP request rules: %w", err)
		}
	}

	// Create Frontend HTTP Response Rules AFTER HTTP Request Rules
	if data.Frontend != nil && data.Frontend.HttpResponseRules != nil && len(data.Frontend.HttpResponseRules) > 0 {
		if err := o.httpResponseRuleManager.CreateHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpResponseRules); err != nil {
			return fmt.Errorf("error creating frontend HTTP response rules: %w", err)
		}
	}

	// Create Frontend TCP Request Rules AFTER HTTP Response Rules
	if data.Frontend != nil && data.Frontend.TcpRequestRules != nil && len(data.Frontend.TcpRequestRules) > 0 {
		tcpRequestRules := o.convertTcpRequestRulesToResourceModels(data.Frontend.TcpRequestRules, "frontend", data.Frontend.Name.ValueString())
		if err := o.tcpRequestRuleManager.Create(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), tcpRequestRules); err != nil {
			return fmt.Errorf("error creating frontend TCP request rules: %w", err)
		}
	}

	// Create Backend HTTP Response Rules AFTER HTTP Request Rules
	if data.Backend != nil && data.Backend.HttpResponseRules != nil && len(data.Backend.HttpResponseRules) > 0 {
		if err := o.httpResponseRuleManager.CreateHttpResponseRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.HttpResponseRules); err != nil {
			return fmt.Errorf("error creating backend HTTP response rules: %w", err)
		}
	}

	// Create Backend TCP Request Rules AFTER HTTP Response Rules
	if data.Backend != nil && data.Backend.TcpRequestRules != nil && len(data.Backend.TcpRequestRules) > 0 {
		tcpRequestRules := o.convertTcpRequestRulesToResourceModels(data.Backend.TcpRequestRules, "backend", data.Backend.Name.ValueString())
		if err := o.tcpRequestRuleManager.Create(ctx, transactionID, "backend", data.Backend.Name.ValueString(), tcpRequestRules); err != nil {
			return fmt.Errorf("error creating backend TCP request rules: %w", err)
		}
	}

	// Create Backend TCP Response Rules AFTER TCP Request Rules
	if data.Backend != nil && data.Backend.TcpResponseRules != nil && len(data.Backend.TcpResponseRules) > 0 {
		tcpResponseRules := o.convertTcpResponseRulesToResourceModels(data.Backend.TcpResponseRules, "backend", data.Backend.Name.ValueString())
		if err := o.tcpResponseRuleManager.Create(ctx, transactionID, "backend", data.Backend.Name.ValueString(), tcpResponseRules); err != nil {
			return fmt.Errorf("error creating backend TCP response rules: %w", err)
		}
	}

	// Create Backend HTTP Checks AFTER TCP Response Rules
	if data.Backend != nil && data.Backend.Httpchecks != nil && len(data.Backend.Httpchecks) > 0 {
		httpChecks := o.convertHttpchecksToResourceModels(data.Backend.Httpchecks, "backend", data.Backend.Name.ValueString())
		if err := o.httpcheckManager.Create(ctx, transactionID, "backend", data.Backend.Name.ValueString(), httpChecks); err != nil {
			return fmt.Errorf("error creating backend HTTP checks: %w", err)
		}
	}

	// Create Backend TCP Checks AFTER HTTP Checks
	if data.Backend != nil && data.Backend.TcpChecks != nil && len(data.Backend.TcpChecks) > 0 {
		tcpChecks := o.convertTcpChecksToResourceModels(data.Backend.TcpChecks, "backend", data.Backend.Name.ValueString())
		if err := o.tcpCheckManager.Create(ctx, transactionID, "backend", data.Backend.Name.ValueString(), tcpChecks); err != nil {
			return fmt.Errorf("error creating backend TCP checks: %w", err)
		}
	}

	// Commit the transaction
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err := o.client.CommitTransaction(transactionID); err != nil {
		// Check if this is a transaction timeout (expected in parallel operations)
		if strings.Contains(err.Error(), "406") && strings.Contains(err.Error(), "outdated") {
			tflog.Warn(ctx, "Transaction timed out (expected in parallel operations)", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		} else {
			tflog.Error(ctx, "Failed to commit transaction", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		}
		return fmt.Errorf("error committing transaction: %w", err)
	}
	tflog.Info(ctx, "Transaction committed successfully", map[string]interface{}{"transaction_id": transactionID})

	tflog.Info(ctx, "HAProxy stack created successfully")
	return nil
}

// Read performs the read operation for the haproxy_stack resource
func (o *StackOperations) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Reading HAProxy stack - READ FUNCTION CALLED")

	// Read backend if specified
	if data.Backend != nil {
		_, err := o.backendManager.ReadBackend(ctx, data.Backend.Name.ValueString(), data.Backend)
		if err != nil {
			return fmt.Errorf("error reading backend: %w", err)
		}
	}

	// Read servers if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Reading servers from HAProxy", map[string]interface{}{
			"backend_name":          data.Backend.Name.ValueString(),
			"current_servers_count": len(data.Backend.Servers),
		})

		servers, err := o.client.ReadServers(ctx, "backend", data.Backend.Name.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Could not read servers, preserving existing state", map[string]interface{}{"error": err.Error()})
			// Don't overwrite data.Backend.Servers if we can't read from HAProxy
			// This preserves the existing state
		} else {
			tflog.Info(ctx, "Successfully read servers from HAProxy", map[string]interface{}{
				"servers_found": len(servers),
			})
			// Convert servers to map format, preserving existing values for fields HAProxy doesn't return
			if data.Backend.Servers == nil {
				data.Backend.Servers = make(map[string]haproxyServerModel)
			}
			for _, server := range servers {
				// Preserve existing values for fields HAProxy doesn't return
				existingServer := data.Backend.Servers[server.Name]
				newServer := o.convertServerPayloadToModel(server)

				// Preserve user-configured values for fields HAProxy doesn't return
				if !existingServer.ForceSslv3.IsNull() && !existingServer.ForceSslv3.IsUnknown() {
					newServer.ForceSslv3 = existingServer.ForceSslv3
				}
				if !existingServer.ForceTlsv10.IsNull() && !existingServer.ForceTlsv10.IsUnknown() {
					newServer.ForceTlsv10 = existingServer.ForceTlsv10
				}
				if !existingServer.ForceTlsv11.IsNull() && !existingServer.ForceTlsv11.IsUnknown() {
					newServer.ForceTlsv11 = existingServer.ForceTlsv11
				}
				if !existingServer.ForceTlsv12.IsNull() && !existingServer.ForceTlsv12.IsUnknown() {
					newServer.ForceTlsv12 = existingServer.ForceTlsv12
				}
				if !existingServer.ForceTlsv13.IsNull() && !existingServer.ForceTlsv13.IsUnknown() {
					newServer.ForceTlsv13 = existingServer.ForceTlsv13
				}
				if !existingServer.SslCertificate.IsNull() && !existingServer.SslCertificate.IsUnknown() {
					newServer.SslCertificate = existingServer.SslCertificate
				}
				if !existingServer.SslMaxVer.IsNull() && !existingServer.SslMaxVer.IsUnknown() {
					newServer.SslMaxVer = existingServer.SslMaxVer
				}
				if !existingServer.SslMinVer.IsNull() && !existingServer.SslMinVer.IsUnknown() {
					newServer.SslMinVer = existingServer.SslMinVer
				}

				data.Backend.Servers[server.Name] = newServer
				tflog.Info(ctx, "Converted server", map[string]interface{}{
					"server_name": server.Name,
				})
			}
		}
	}

	// Read TCP checks if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Reading TCP checks from HAProxy", map[string]interface{}{
			"backend_name": data.Backend.Name.ValueString(),
		})

		tcpChecks, err := o.client.ReadTcpChecks(ctx, "backend", data.Backend.Name.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Could not read TCP checks, preserving existing state", map[string]interface{}{"error": err.Error()})
			// Don't overwrite data.Backend.TcpChecks if we can't read from HAProxy
		} else {
			tflog.Info(ctx, "Successfully read TCP checks from HAProxy", map[string]interface{}{
				"tcp_checks_found": len(tcpChecks),
			})
			// Debug: Log the actual TCP checks from HAProxy
			for i, tcpCheck := range tcpChecks {
				tflog.Info(ctx, "HAProxy TCP check", map[string]interface{}{
					"index":   i,
					"action":  tcpCheck.Action,
					"addr":    tcpCheck.Addr,
					"port":    tcpCheck.Port,
					"data":    tcpCheck.Data,
					"pattern": tcpCheck.Pattern,
				})
			}
			// Convert TCP checks to model format
			data.Backend.TcpChecks = make([]haproxyTcpCheckModel, len(tcpChecks))
			for i, tcpCheck := range tcpChecks {
				data.Backend.TcpChecks[i] = o.convertTcpCheckPayloadToStackModel(tcpCheck)
			}

			// Debug: Log the final state after conversion
			tflog.Info(ctx, "Final TCP checks state after Read", map[string]interface{}{
				"tcp_checks_count": len(data.Backend.TcpChecks),
			})
			for i, tcpCheck := range data.Backend.TcpChecks {
				tflog.Info(ctx, "Final TCP check state", map[string]interface{}{
					"index":   i,
					"action":  tcpCheck.Action.ValueString(),
					"addr":    tcpCheck.Addr.ValueString(),
					"port":    tcpCheck.Port.ValueInt64(),
					"data":    tcpCheck.Data.ValueString(),
					"pattern": tcpCheck.Pattern.ValueString(),
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

				// Set TLS fields to null to avoid managing default values
				// These should only be set if explicitly configured by the user
				bindModel.Tlsv10 = types.BoolNull()
				bindModel.Tlsv11 = types.BoolNull()
				bindModel.Tlsv12 = types.BoolNull()
				bindModel.Tlsv13 = types.BoolNull()
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

	// ACLs are now handled within frontend/backend blocks

	// HTTP Request and Response Rules are managed by Terraform state
	// We don't read them from HAProxy to avoid state drift issues
	// The Terraform state is the source of truth for these rules

	tflog.Info(ctx, "HAProxy stack read successfully")
	return nil
}

// Update performs the update operation for the haproxy_stack resource
func (o *StackOperations) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *haproxyStackResourceModel) error {
	// Serialize all HAProxy operations to prevent transaction conflicts
	globalTransactionMutex.Lock()
	defer globalTransactionMutex.Unlock()

	return o.updateSingle(ctx, req, resp, data)
}

// updateSingle performs a single update operation with transaction retry logic
func (o *StackOperations) updateSingle(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *haproxyStackResourceModel) error {
	// Retry the entire operation if transaction becomes outdated
	for {
		err := o.updateSingleInternal(ctx, req, resp, data)
		if err == nil {
			return nil
		}

		// Check if this is a retryable transaction error
		if o.isTransactionRetryableError(err) {
			tflog.Info(ctx, "Transaction outdated, retrying entire operation", map[string]interface{}{"error": err.Error()})
			continue
		}

		// Non-retryable error, return it
		return err
	}
}

// updateSingleInternal performs the actual update operation without retry
func (o *StackOperations) updateSingleInternal(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse, data *haproxyStackResourceModel) error {
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
		return fmt.Errorf("error beginning transaction: %w", err)
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
				return fmt.Errorf("error updating backend: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend unchanged, skipping update")
		}
	}

	// Update servers only if they changed in the plan
	if data.Backend != nil && len(data.Backend.Servers) > 0 {
		// Check if servers changed by comparing plan vs state
		var stateServers map[string]haproxyServerModel
		if state.Backend != nil {
			stateServers = state.Backend.Servers
		}
		serversChanged := o.serversChanged(ctx, data.Backend.Servers, stateServers)
		if serversChanged {
			tflog.Info(ctx, "Servers changed, updating", map[string]interface{}{
				"backend_name":          data.Backend.Name.ValueString(),
				"desired_servers_count": len(data.Backend.Servers),
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

			// Create a map of desired servers by name (data.Backend.Servers is already a map)
			desiredServerMap := data.Backend.Servers

			// Delete servers that are no longer in the desired state
			for serverName := range existingServerMap {
				if _, exists := desiredServerMap[serverName]; !exists {
					tflog.Info(ctx, "Deleting server", map[string]interface{}{"server_name": serverName})
					if err = o.client.DeleteServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverName); err != nil {
						return fmt.Errorf("error deleting server %s: %w", serverName, err)
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
							return fmt.Errorf("error updating server %s: %w", serverName, err)
						}
					} else {
						tflog.Info(ctx, "Server unchanged", map[string]interface{}{"server_name": serverName})
					}
				} else {
					// Server doesn't exist, create it
					tflog.Info(ctx, "Creating new server", map[string]interface{}{"server_name": serverName})
					if err = o.client.CreateServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), serverPayload); err != nil {
						return fmt.Errorf("error creating server %s: %w", serverName, err)
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
				return fmt.Errorf("error updating frontend: %w", err)
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
				return fmt.Errorf("error updating binds: %w", err)
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
				return fmt.Errorf("error updating frontend ACLs: %w", err)
			}
		} else {
			tflog.Info(ctx, "Frontend ACLs unchanged, skipping update")
		}
	} else if data.Frontend != nil && state.Frontend != nil && state.Frontend.Acls != nil && len(state.Frontend.Acls) > 0 {
		// Handle frontend ACLs deletion - plan has no ACLs but state does
		tflog.Info(ctx, "Frontend ACLs removed, deleting", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.aclManager.DeleteACLsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting frontend ACLs: %w", err)
		}
	}

	// Update backend ACLs only if they changed in the plan
	if data.Backend != nil && data.Backend.Acls != nil && len(data.Backend.Acls) > 0 {
		// Check if backend ACLs changed by comparing plan vs state
		backendACLsChanged := o.aclsChanged(ctx, data.Backend.Acls, state.Backend.Acls)
		if backendACLsChanged {
			tflog.Info(ctx, "Backend ACLs changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			if err = o.aclManager.UpdateACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.Acls); err != nil {
				return fmt.Errorf("error updating backend ACLs: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend ACLs unchanged, skipping update")
		}
	} else if data.Backend != nil && state.Backend != nil && state.Backend.Acls != nil && len(state.Backend.Acls) > 0 {
		// Handle backend ACLs deletion - plan has no ACLs but state does
		tflog.Info(ctx, "Backend ACLs removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.aclManager.DeleteACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend ACLs: %w", err)
		}
	}

	// Update HTTP Request Rules only if they changed in the plan
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		// Check if HTTP Request Rules changed by comparing plan vs state
		httpRequestRulesChanged := o.httpRequestRulesChanged(ctx, data.Frontend.HttpRequestRules, state.Frontend.HttpRequestRules)
		if httpRequestRulesChanged {
			tflog.Info(ctx, "HTTP request rules changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.httpRequestRuleManager.UpdateHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpRequestRules); err != nil {
				return fmt.Errorf("error updating HTTP request rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "HTTP request rules unchanged, skipping update")
		}
	} else if data.Frontend != nil && state.Frontend != nil && state.Frontend.HttpRequestRules != nil && len(state.Frontend.HttpRequestRules) > 0 {
		// Handle HTTP request rules deletion - plan has no rules but state does
		tflog.Info(ctx, "HTTP request rules removed, deleting", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.DeleteHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting HTTP request rules: %w", err)
		}
	}

	// Update Backend HTTP Request Rules only if they changed in the plan
	if data.Backend != nil && data.Backend.HttpRequestRules != nil && len(data.Backend.HttpRequestRules) > 0 {
		// Check if HTTP Request Rules changed by comparing plan vs state
		httpRequestRulesChanged := o.httpRequestRulesChanged(ctx, data.Backend.HttpRequestRules, state.Backend.HttpRequestRules)
		if httpRequestRulesChanged {
			tflog.Info(ctx, "Backend HTTP request rules changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			if err = o.httpRequestRuleManager.UpdateHttpRequestRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.HttpRequestRules); err != nil {
				return fmt.Errorf("error updating backend HTTP request rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend HTTP request rules unchanged, skipping update")
		}
	} else if data.Backend != nil && state.Backend != nil && state.Backend.HttpRequestRules != nil && len(state.Backend.HttpRequestRules) > 0 {
		// Handle backend HTTP request rules deletion - plan has no rules but state does
		tflog.Info(ctx, "Backend HTTP request rules removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.httpRequestRuleManager.DeleteHttpRequestRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend HTTP request rules: %w", err)
		}
	}

	// Update Frontend HTTP Response Rules only if they changed in the plan
	if data.Frontend != nil && data.Frontend.HttpResponseRules != nil && len(data.Frontend.HttpResponseRules) > 0 {
		// Check if HTTP Response Rules changed by comparing plan vs state
		httpResponseRulesChanged := o.httpResponseRulesChanged(ctx, data.Frontend.HttpResponseRules, state.Frontend.HttpResponseRules)
		if httpResponseRulesChanged {
			tflog.Info(ctx, "Frontend HTTP response rules changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			if err = o.httpResponseRuleManager.UpdateHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), data.Frontend.HttpResponseRules); err != nil {
				return fmt.Errorf("error updating frontend HTTP response rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "Frontend HTTP response rules unchanged, skipping update")
		}
	} else if data.Frontend != nil && state.Frontend != nil && state.Frontend.HttpResponseRules != nil && len(state.Frontend.HttpResponseRules) > 0 {
		// Handle frontend HTTP response rules deletion - plan has no rules but state does
		tflog.Info(ctx, "Frontend HTTP response rules removed, deleting", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpResponseRuleManager.DeleteHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting frontend HTTP response rules: %w", err)
		}
	}

	// Update Backend HTTP Response Rules only if they changed in the plan
	if data.Backend != nil && data.Backend.HttpResponseRules != nil && len(data.Backend.HttpResponseRules) > 0 {
		// Check if HTTP Response Rules changed by comparing plan vs state
		httpResponseRulesChanged := o.httpResponseRulesChanged(ctx, data.Backend.HttpResponseRules, state.Backend.HttpResponseRules)
		if httpResponseRulesChanged {
			tflog.Info(ctx, "Backend HTTP response rules changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			if err = o.httpResponseRuleManager.UpdateHttpResponseRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), data.Backend.HttpResponseRules); err != nil {
				return fmt.Errorf("error updating backend HTTP response rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend HTTP response rules unchanged, skipping update")
		}
	} else if data.Backend != nil && state.Backend != nil && state.Backend.HttpResponseRules != nil && len(state.Backend.HttpResponseRules) > 0 {
		// Handle backend HTTP response rules deletion - plan has no rules but state does
		tflog.Info(ctx, "Backend HTTP response rules removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.httpResponseRuleManager.DeleteHttpResponseRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend HTTP response rules: %w", err)
		}
	}

	// Update Frontend TCP Request Rules only if they changed in the plan
	if data.Frontend != nil && data.Frontend.TcpRequestRules != nil && len(data.Frontend.TcpRequestRules) > 0 {
		// Check if TCP Request Rules changed by comparing plan vs state
		tcpRequestRulesChanged := o.tcpRequestRuleChanged(ctx, data.Frontend.TcpRequestRules, state.Frontend.TcpRequestRules)
		if tcpRequestRulesChanged {
			tflog.Info(ctx, "Frontend TCP request rules changed, updating", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
			tcpRequestRules := o.convertTcpRequestRulesToResourceModels(data.Frontend.TcpRequestRules, "frontend", data.Frontend.Name.ValueString())
			if err = o.tcpRequestRuleManager.Update(ctx, transactionID, "frontend", data.Frontend.Name.ValueString(), tcpRequestRules); err != nil {
				return fmt.Errorf("error updating frontend TCP request rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "Frontend TCP request rules unchanged, skipping update")
		}
	} else if data.Frontend != nil && state.Frontend != nil && state.Frontend.TcpRequestRules != nil && len(state.Frontend.TcpRequestRules) > 0 {
		// Handle frontend TCP request rules deletion - plan has no rules but state does
		tflog.Info(ctx, "Frontend TCP request rules removed, deleting", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.tcpRequestRuleManager.Delete(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting frontend TCP request rules: %w", err)
		}
	}

	// Update Backend TCP Request Rules only if they changed in the plan
	if data.Backend != nil && data.Backend.TcpRequestRules != nil && len(data.Backend.TcpRequestRules) > 0 {
		// Check if TCP Request Rules changed by comparing plan vs state
		tcpRequestRulesChanged := o.tcpRequestRuleChanged(ctx, data.Backend.TcpRequestRules, state.Backend.TcpRequestRules)
		if tcpRequestRulesChanged {
			tflog.Info(ctx, "Backend TCP request rules changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			tcpRequestRules := o.convertTcpRequestRulesToResourceModels(data.Backend.TcpRequestRules, "backend", data.Backend.Name.ValueString())
			if err = o.tcpRequestRuleManager.Update(ctx, transactionID, "backend", data.Backend.Name.ValueString(), tcpRequestRules); err != nil {
				return fmt.Errorf("error updating backend TCP request rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend TCP request rules unchanged, skipping update")
		}
	} else if data.Backend != nil && state.Backend != nil && state.Backend.TcpRequestRules != nil && len(state.Backend.TcpRequestRules) > 0 {
		// Handle backend TCP request rules deletion - plan has no rules but state does
		tflog.Info(ctx, "Backend TCP request rules removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.tcpRequestRuleManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend TCP request rules: %w", err)
		}
	}

	// Update Backend TCP Response Rules only if they changed in the plan
	if data.Backend != nil && data.Backend.TcpResponseRules != nil && len(data.Backend.TcpResponseRules) > 0 {
		// Check if TCP Response Rules changed by comparing plan vs state
		tcpResponseRulesChanged := o.tcpResponseRuleChanged(ctx, data.Backend.TcpResponseRules, state.Backend.TcpResponseRules)
		if tcpResponseRulesChanged {
			tflog.Info(ctx, "Backend TCP response rules changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			tcpResponseRules := o.convertTcpResponseRulesToResourceModels(data.Backend.TcpResponseRules, "backend", data.Backend.Name.ValueString())
			if err = o.tcpResponseRuleManager.Update(ctx, transactionID, "backend", data.Backend.Name.ValueString(), tcpResponseRules); err != nil {
				return fmt.Errorf("error updating backend TCP response rules: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend TCP response rules unchanged, skipping update")
		}
	} else if data.Backend != nil && state.Backend != nil && state.Backend.TcpResponseRules != nil && len(state.Backend.TcpResponseRules) > 0 {
		// Handle backend TCP response rules deletion - plan has no rules but state does
		tflog.Info(ctx, "Backend TCP response rules removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.tcpResponseRuleManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend TCP response rules: %w", err)
		}
	}

	// Update Backend HTTP Checks only if they changed in the plan
	if data.Backend != nil && data.Backend.Httpchecks != nil && len(data.Backend.Httpchecks) > 0 {
		// Check if HTTP Checks changed by comparing plan vs state
		httpcheckChanged := o.httpcheckChanged(ctx, data.Backend.Httpchecks, state.Backend.Httpchecks)
		if httpcheckChanged {
			tflog.Info(ctx, "Backend HTTP checks changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			httpChecks := o.convertHttpchecksToResourceModels(data.Backend.Httpchecks, "backend", data.Backend.Name.ValueString())
			if err = o.httpcheckManager.Update(ctx, transactionID, "backend", data.Backend.Name.ValueString(), httpChecks); err != nil {
				return fmt.Errorf("error updating backend HTTP checks: %w", err)
			}
		} else {
			tflog.Info(ctx, "Backend HTTP checks unchanged, skipping update")
		}
	} else if data.Backend != nil && state.Backend != nil && state.Backend.Httpchecks != nil && len(state.Backend.Httpchecks) > 0 {
		// Handle HTTP checks deletion - plan has no HTTP checks but state does
		tflog.Info(ctx, "Backend HTTP checks removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.httpcheckManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend HTTP checks: %w", err)
		}
	}

	// Handle Backend TCP Checks
	if data.Backend != nil && state.Backend != nil {
		// Debug logging for TCP checks
		tflog.Info(ctx, "TCP checks processing", map[string]interface{}{
			"backend_name":         data.Backend.Name.ValueString(),
			"state_tcp_checks_len": len(state.Backend.TcpChecks),
			"data_tcp_checks_len":  len(data.Backend.TcpChecks),
		})

		// Check for deletion first - plan has no TCP checks but state does
		if len(state.Backend.TcpChecks) > 0 && len(data.Backend.TcpChecks) == 0 {
			// Handle TCP checks deletion
			tflog.Info(ctx, "Backend TCP checks removed, deleting", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
			if err = o.tcpCheckManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
				return fmt.Errorf("error deleting backend TCP checks: %w", err)
			}
		} else if len(data.Backend.TcpChecks) > 0 {
			// Check if TCP Checks changed by comparing plan vs state
			tcpCheckChanged := o.tcpCheckChanged(ctx, data.Backend.TcpChecks, state.Backend.TcpChecks)
			if tcpCheckChanged {
				tflog.Info(ctx, "Backend TCP checks changed, updating", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
				tcpChecks := o.convertTcpChecksToResourceModels(data.Backend.TcpChecks, "backend", data.Backend.Name.ValueString())
				if err = o.tcpCheckManager.Update(ctx, transactionID, "backend", data.Backend.Name.ValueString(), tcpChecks); err != nil {
					return fmt.Errorf("error updating backend TCP checks: %w", err)
				}

				// Debug: Read back the TCP checks to see what HAProxy actually stored
				tflog.Info(ctx, "Reading TCP checks back from HAProxy after update")
				readTcpChecks, err := o.client.ReadTcpChecks(ctx, "backend", data.Backend.Name.ValueString())
				if err != nil {
					tflog.Warn(ctx, "Could not read TCP checks back from HAProxy", map[string]interface{}{"error": err.Error()})
				} else {
					tflog.Info(ctx, "HAProxy returned TCP checks after update", map[string]interface{}{
						"tcp_checks_count": len(readTcpChecks),
					})
					for i, tcpCheck := range readTcpChecks {
						tflog.Info(ctx, "HAProxy TCP check after update", map[string]interface{}{
							"index":   i,
							"action":  tcpCheck.Action,
							"addr":    tcpCheck.Addr,
							"port":    tcpCheck.Port,
							"data":    tcpCheck.Data,
							"pattern": tcpCheck.Pattern,
						})
					}
				}
			} else {
				tflog.Info(ctx, "Backend TCP checks unchanged, skipping update")
			}
		}
	}

	// Commit all updates
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err := o.client.CommitTransaction(transactionID); err != nil {
		// Check if this is a transaction timeout (expected in parallel operations)
		if strings.Contains(err.Error(), "406") && strings.Contains(err.Error(), "outdated") {
			tflog.Warn(ctx, "Transaction timed out (expected in parallel operations)", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		} else {
			tflog.Error(ctx, "Failed to commit transaction", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		}
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

	// Compare TcpRequestRules field
	if o.tcpRequestRuleChanged(ctx, planFrontend.TcpRequestRules, stateFrontend.TcpRequestRules) {
		tflog.Info(ctx, "Frontend TcpRequestRules changed", map[string]interface{}{
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

	// Compare Httpchecks field
	if o.httpcheckChanged(ctx, planBackend.Httpchecks, stateBackend.Httpchecks) {
		tflog.Info(ctx, "Backend Httpcheck changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare TcpChecks field
	if o.tcpCheckChanged(ctx, planBackend.TcpChecks, stateBackend.TcpChecks) {
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

	// Compare HttpRequestRules field
	if o.httpRequestRulesChanged(ctx, planBackend.HttpRequestRules, stateBackend.HttpRequestRules) {
		tflog.Info(ctx, "Backend HttpRequestRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare HttpResponseRules field
	if o.httpResponseRulesChanged(ctx, planBackend.HttpResponseRules, stateBackend.HttpResponseRules) {
		tflog.Info(ctx, "Backend HttpResponseRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare TcpRequestRules field
	if o.tcpRequestRuleChanged(ctx, planBackend.TcpRequestRules, stateBackend.TcpRequestRules) {
		tflog.Info(ctx, "Backend TcpRequestRule changed", map[string]interface{}{
			"plan_name":  planBackend.Name.ValueString(),
			"state_name": stateBackend.Name.ValueString(),
		})
		return true
	}

	// Compare TcpResponseRules field
	if o.tcpResponseRuleChanged(ctx, planBackend.TcpResponseRules, stateBackend.TcpResponseRules) {
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
		if planCheck.Type.ValueString() != stateCheck.Type.ValueString() ||
			planCheck.Addr.ValueString() != stateCheck.Addr.ValueString() ||
			planCheck.Alpn.ValueString() != stateCheck.Alpn.ValueString() ||
			planCheck.Body.ValueString() != stateCheck.Body.ValueString() ||
			planCheck.BodyLogFormat.ValueString() != stateCheck.BodyLogFormat.ValueString() ||
			planCheck.CheckComment.ValueString() != stateCheck.CheckComment.ValueString() ||
			planCheck.Default.ValueBool() != stateCheck.Default.ValueBool() ||
			planCheck.ErrorStatus.ValueString() != stateCheck.ErrorStatus.ValueString() ||
			planCheck.ExclamationMark.ValueBool() != stateCheck.ExclamationMark.ValueBool() ||
			planCheck.Linger.ValueBool() != stateCheck.Linger.ValueBool() ||
			planCheck.Match.ValueString() != stateCheck.Match.ValueString() ||
			planCheck.Method.ValueString() != stateCheck.Method.ValueString() ||
			planCheck.MinRecv.ValueInt64() != stateCheck.MinRecv.ValueInt64() ||
			planCheck.OkStatus.ValueString() != stateCheck.OkStatus.ValueString() ||
			planCheck.OnError.ValueString() != stateCheck.OnError.ValueString() ||
			planCheck.OnSuccess.ValueString() != stateCheck.OnSuccess.ValueString() ||
			planCheck.Pattern.ValueString() != stateCheck.Pattern.ValueString() ||
			planCheck.Port.ValueInt64() != stateCheck.Port.ValueInt64() ||
			planCheck.PortString.ValueString() != stateCheck.PortString.ValueString() ||
			planCheck.Proto.ValueString() != stateCheck.Proto.ValueString() ||
			planCheck.SendProxy.ValueBool() != stateCheck.SendProxy.ValueBool() ||
			planCheck.Sni.ValueString() != stateCheck.Sni.ValueString() ||
			planCheck.Ssl.ValueBool() != stateCheck.Ssl.ValueBool() ||
			planCheck.StatusCode.ValueString() != stateCheck.StatusCode.ValueString() ||
			planCheck.ToutStatus.ValueString() != stateCheck.ToutStatus.ValueString() {
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

		if planCheck.Action.ValueString() != stateCheck.Action.ValueString() ||
			planCheck.Addr.ValueString() != stateCheck.Addr.ValueString() ||
			planCheck.Alpn.ValueString() != stateCheck.Alpn.ValueString() ||
			planCheck.CheckComment.ValueString() != stateCheck.CheckComment.ValueString() ||
			planCheck.Data.ValueString() != stateCheck.Data.ValueString() ||
			planCheck.Default.ValueBool() != stateCheck.Default.ValueBool() ||
			planCheck.ErrorStatus.ValueString() != stateCheck.ErrorStatus.ValueString() ||
			planCheck.ExclamationMark.ValueBool() != stateCheck.ExclamationMark.ValueBool() ||
			planCheck.Fmt.ValueString() != stateCheck.Fmt.ValueString() ||
			planCheck.HexFmt.ValueString() != stateCheck.HexFmt.ValueString() ||
			planCheck.HexString.ValueString() != stateCheck.HexString.ValueString() ||
			planCheck.Linger.ValueBool() != stateCheck.Linger.ValueBool() ||
			planCheck.Match.ValueString() != stateCheck.Match.ValueString() ||
			planCheck.MinRecv.ValueInt64() != stateCheck.MinRecv.ValueInt64() ||
			planCheck.OkStatus.ValueString() != stateCheck.OkStatus.ValueString() ||
			planCheck.OnError.ValueString() != stateCheck.OnError.ValueString() ||
			planCheck.OnSuccess.ValueString() != stateCheck.OnSuccess.ValueString() ||
			planCheck.Pattern.ValueString() != stateCheck.Pattern.ValueString() ||
			planCheck.Port.ValueInt64() != stateCheck.Port.ValueInt64() ||
			planCheck.PortString.ValueString() != stateCheck.PortString.ValueString() ||
			planCheck.Proto.ValueString() != stateCheck.Proto.ValueString() ||
			planCheck.SendProxy.ValueBool() != stateCheck.SendProxy.ValueBool() ||
			planCheck.Sni.ValueString() != stateCheck.Sni.ValueString() ||
			planCheck.Ssl.ValueBool() != stateCheck.Ssl.ValueBool() ||
			planCheck.StatusCode.ValueString() != stateCheck.StatusCode.ValueString() ||
			planCheck.ToutStatus.ValueString() != stateCheck.ToutStatus.ValueString() ||
			planCheck.VarExpr.ValueString() != stateCheck.VarExpr.ValueString() ||
			planCheck.VarFmt.ValueString() != stateCheck.VarFmt.ValueString() ||
			planCheck.VarName.ValueString() != stateCheck.VarName.ValueString() ||
			planCheck.VarScope.ValueString() != stateCheck.VarScope.ValueString() ||
			planCheck.ViaSocks4.ValueBool() != stateCheck.ViaSocks4.ValueBool() {
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

		if planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Action.ValueString() != stateRule.Action.ValueString() ||
			planRule.Cond.ValueString() != stateRule.Cond.ValueString() ||
			planRule.CondTest.ValueString() != stateRule.CondTest.ValueString() ||
			planRule.Expr.ValueString() != stateRule.Expr.ValueString() ||
			planRule.Timeout.ValueInt64() != stateRule.Timeout.ValueInt64() ||
			planRule.LuaAction.ValueString() != stateRule.LuaAction.ValueString() ||
			planRule.LuaParams.ValueString() != stateRule.LuaParams.ValueString() ||
			planRule.LogLevel.ValueString() != stateRule.LogLevel.ValueString() ||
			planRule.MarkValue.ValueString() != stateRule.MarkValue.ValueString() ||
			planRule.NiceValue.ValueInt64() != stateRule.NiceValue.ValueInt64() ||
			planRule.TosValue.ValueString() != stateRule.TosValue.ValueString() ||
			planRule.CaptureLen.ValueInt64() != stateRule.CaptureLen.ValueInt64() ||
			planRule.CaptureSample.ValueString() != stateRule.CaptureSample.ValueString() ||
			planRule.BandwidthLimitLimit.ValueString() != stateRule.BandwidthLimitLimit.ValueString() ||
			planRule.BandwidthLimitName.ValueString() != stateRule.BandwidthLimitName.ValueString() ||
			planRule.BandwidthLimitPeriod.ValueString() != stateRule.BandwidthLimitPeriod.ValueString() ||
			planRule.ResolveProtocol.ValueString() != stateRule.ResolveProtocol.ValueString() ||
			planRule.ResolveResolvers.ValueString() != stateRule.ResolveResolvers.ValueString() ||
			planRule.ResolveVar.ValueString() != stateRule.ResolveVar.ValueString() ||
			planRule.RstTtl.ValueInt64() != stateRule.RstTtl.ValueInt64() ||
			planRule.ScIdx.ValueInt64() != stateRule.ScIdx.ValueInt64() ||
			planRule.ScIncId.ValueString() != stateRule.ScIncId.ValueString() ||
			planRule.ScInt.ValueInt64() != stateRule.ScInt.ValueInt64() ||
			planRule.ServerName.ValueString() != stateRule.ServerName.ValueString() ||
			planRule.ServiceName.ValueString() != stateRule.ServiceName.ValueString() ||
			planRule.VarName.ValueString() != stateRule.VarName.ValueString() ||
			planRule.VarFormat.ValueString() != stateRule.VarFormat.ValueString() ||
			planRule.VarScope.ValueString() != stateRule.VarScope.ValueString() {
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

		if planRule.Type.ValueString() != stateRule.Type.ValueString() ||
			planRule.Action.ValueString() != stateRule.Action.ValueString() ||
			planRule.Cond.ValueString() != stateRule.Cond.ValueString() ||
			planRule.CondTest.ValueString() != stateRule.CondTest.ValueString() ||
			planRule.Expr.ValueString() != stateRule.Expr.ValueString() ||
			planRule.LogLevel.ValueString() != stateRule.LogLevel.ValueString() ||
			planRule.LuaAction.ValueString() != stateRule.LuaAction.ValueString() ||
			planRule.LuaParams.ValueString() != stateRule.LuaParams.ValueString() ||
			planRule.MarkValue.ValueString() != stateRule.MarkValue.ValueString() ||
			planRule.NiceValue.ValueInt64() != stateRule.NiceValue.ValueInt64() ||
			planRule.RstTtl.ValueInt64() != stateRule.RstTtl.ValueInt64() ||
			planRule.ScExpr.ValueString() != stateRule.ScExpr.ValueString() ||
			planRule.ScId.ValueInt64() != stateRule.ScId.ValueInt64() ||
			planRule.ScIdx.ValueInt64() != stateRule.ScIdx.ValueInt64() ||
			planRule.ScInt.ValueInt64() != stateRule.ScInt.ValueInt64() ||
			planRule.SpoeEngine.ValueString() != stateRule.SpoeEngine.ValueString() ||
			planRule.SpoeGroup.ValueString() != stateRule.SpoeGroup.ValueString() ||
			planRule.Timeout.ValueInt64() != stateRule.Timeout.ValueInt64() ||
			planRule.TosValue.ValueString() != stateRule.TosValue.ValueString() ||
			planRule.VarFormat.ValueString() != stateRule.VarFormat.ValueString() ||
			planRule.VarName.ValueString() != stateRule.VarName.ValueString() ||
			planRule.VarScope.ValueString() != stateRule.VarScope.ValueString() ||
			planRule.BandwidthLimitLimit.ValueString() != stateRule.BandwidthLimitLimit.ValueString() ||
			planRule.BandwidthLimitName.ValueString() != stateRule.BandwidthLimitName.ValueString() ||
			planRule.BandwidthLimitPeriod.ValueString() != stateRule.BandwidthLimitPeriod.ValueString() {
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
	// Handle nil stateServers (first time creation)
	if stateServers == nil {
		if len(planServers) > 0 {
			tflog.Info(ctx, "Servers added (first time)", map[string]interface{}{
				"plan_count": len(planServers),
			})
			return true
		}
		return false
	}

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
	// Serialize all HAProxy operations to prevent transaction conflicts
	globalTransactionMutex.Lock()
	defer globalTransactionMutex.Unlock()

	return o.deleteSingle(ctx, req, resp, data)
}

// deleteSingle performs a single delete operation with transaction retry logic
func (o *StackOperations) deleteSingle(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, data *haproxyStackResourceModel) error {
	// Retry the entire operation if transaction becomes outdated
	for {
		err := o.deleteSingleInternal(ctx, req, resp, data)
		if err == nil {
			return nil
		}

		// Check if this is a retryable transaction error
		if o.isTransactionRetryableError(err) {
			tflog.Info(ctx, "Transaction outdated, retrying entire operation", map[string]interface{}{"error": err.Error()})
			continue
		}

		// Non-retryable error, return it
		return err
	}
}

// deleteSingleInternal performs the actual delete operation without retry
func (o *StackOperations) deleteSingleInternal(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse, data *haproxyStackResourceModel) error {
	tflog.Info(ctx, "Deleting HAProxy stack")

	// Begin transaction for all deletes
	transactionID, err := o.client.BeginTransaction()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
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
			return fmt.Errorf("error deleting frontend ACLs: %w", err)
		}
	}

	if data.Backend != nil && data.Backend.Acls != nil && len(data.Backend.Acls) > 0 {
		tflog.Info(ctx, "Deleting backend ACLs", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.aclManager.DeleteACLsInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend ACLs: %w", err)
		}
	}

	// Delete HTTP Request Rules if specified
	if data.Frontend != nil && data.Frontend.HttpRequestRules != nil && len(data.Frontend.HttpRequestRules) > 0 {
		tflog.Info(ctx, "Deleting HTTP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpRequestRuleManager.DeleteHttpRequestRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting HTTP request rules: %w", err)
		}
	}

	// Delete Frontend HTTP Response Rules if specified
	if data.Frontend != nil && data.Frontend.HttpResponseRules != nil && len(data.Frontend.HttpResponseRules) > 0 {
		tflog.Info(ctx, "Deleting frontend HTTP response rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.httpResponseRuleManager.DeleteHttpResponseRulesInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting frontend HTTP response rules: %w", err)
		}
	}

	// Delete Backend HTTP Response Rules if specified
	if data.Backend != nil && data.Backend.HttpResponseRules != nil && len(data.Backend.HttpResponseRules) > 0 {
		tflog.Info(ctx, "Deleting backend HTTP response rules", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.httpResponseRuleManager.DeleteHttpResponseRulesInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend HTTP response rules: %w", err)
		}
	}

	// Delete Frontend TCP Request Rules if specified
	if data.Frontend != nil && data.Frontend.TcpRequestRules != nil && len(data.Frontend.TcpRequestRules) > 0 {
		tflog.Info(ctx, "Deleting frontend TCP request rules", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.tcpRequestRuleManager.Delete(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting frontend TCP request rules: %w", err)
		}
	}

	// Delete Backend TCP Request Rules if specified
	if data.Backend != nil && data.Backend.TcpRequestRules != nil && len(data.Backend.TcpRequestRules) > 0 {
		tflog.Info(ctx, "Deleting backend TCP request rules", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.tcpRequestRuleManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend TCP request rules: %w", err)
		}
	}

	// Delete Backend TCP Response Rules if specified
	if data.Backend != nil && data.Backend.TcpResponseRules != nil && len(data.Backend.TcpResponseRules) > 0 {
		tflog.Info(ctx, "Deleting backend TCP response rules", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.tcpResponseRuleManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend TCP response rules: %w", err)
		}
	}

	// Delete Backend HTTP Checks if specified
	if data.Backend != nil && data.Backend.Httpchecks != nil && len(data.Backend.Httpchecks) > 0 {
		tflog.Info(ctx, "Deleting backend HTTP checks", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.httpcheckManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend HTTP checks: %w", err)
		}
	}

	// Delete Backend TCP Checks if specified
	if data.Backend != nil && data.Backend.TcpChecks != nil && len(data.Backend.TcpChecks) > 0 {
		tflog.Info(ctx, "Deleting backend TCP checks", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.tcpCheckManager.Delete(ctx, transactionID, "backend", data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend TCP checks: %w", err)
		}
	}

	// Delete binds for frontend if specified
	if data.Frontend != nil && data.Frontend.Binds != nil && len(data.Frontend.Binds) > 0 {
		tflog.Info(ctx, "Deleting binds", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.bindManager.DeleteBindsInTransaction(ctx, transactionID, "frontend", data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting binds: %w", err)
		}
	}

	// Delete frontend if specified
	if data.Frontend != nil {
		tflog.Info(ctx, "Deleting frontend", map[string]interface{}{"frontend_name": data.Frontend.Name.ValueString()})
		if err = o.frontendManager.DeleteFrontendInTransaction(ctx, transactionID, data.Frontend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting frontend: %w", err)
		}
	}

	// Delete servers if specified - use name-based management
	if data.Backend != nil && len(data.Backend.Servers) > 0 {
		// Read existing servers to get current state
		existingServers, err := o.client.ReadServers(ctx, "backend", data.Backend.Name.ValueString())
		if err != nil {
			tflog.Warn(ctx, "Could not read existing servers for deletion", map[string]interface{}{"error": err.Error()})
			existingServers = []ServerPayload{}
		}

		// Create a map of desired servers by name (data.Backend.Servers is already a map)
		desiredServerMap := make(map[string]bool)
		for serverName := range data.Backend.Servers {
			desiredServerMap[serverName] = true
		}

		// Delete servers that are not in the desired state
		for _, existingServer := range existingServers {
			if !desiredServerMap[existingServer.Name] {
				tflog.Info(ctx, "Deleting server", map[string]interface{}{"server_name": existingServer.Name})
				if err = o.client.DeleteServerInTransaction(ctx, transactionID, "backend", data.Backend.Name.ValueString(), existingServer.Name); err != nil {
					return fmt.Errorf("error deleting server %s: %w", existingServer.Name, err)
				}
			}
		}
	}

	// Delete backend if specified
	if data.Backend != nil {
		tflog.Info(ctx, "Deleting backend", map[string]interface{}{"backend_name": data.Backend.Name.ValueString()})
		if err = o.backendManager.DeleteBackendInTransaction(ctx, transactionID, data.Backend.Name.ValueString()); err != nil {
			return fmt.Errorf("error deleting backend: %w", err)
		}
	}

	// Commit all deletes
	tflog.Info(ctx, "Committing transaction", map[string]interface{}{"transaction_id": transactionID})
	if err := o.client.CommitTransaction(transactionID); err != nil {
		// Check if this is a transaction timeout (expected in parallel operations)
		if strings.Contains(err.Error(), "406") && strings.Contains(err.Error(), "outdated") {
			tflog.Warn(ctx, "Transaction timed out (expected in parallel operations)", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		} else {
			tflog.Error(ctx, "Failed to commit transaction", map[string]interface{}{"transaction_id": transactionID, "error": err.Error()})
		}
		return err
	}

	// Clear the error so defer doesn't rollback
	err = nil
	tflog.Info(ctx, "HAProxy stack deleted successfully")
	return nil
}

// convertTcpRequestRulesToResourceModels converts stack models to resource models
func (o *StackOperations) convertTcpRequestRulesToResourceModels(stackRules []haproxyTcpRequestRuleModel, parentType, parentName string) []TcpRequestRuleResourceModel {
	resourceRules := make([]TcpRequestRuleResourceModel, len(stackRules))
	for i, stackRule := range stackRules {
		resourceRules[i] = TcpRequestRuleResourceModel{
			ID:                   types.StringValue(fmt.Sprintf("%s/%s/tcp_request_rule/%d", parentType, parentName, i)),
			ParentType:           types.StringValue(parentType),
			ParentName:           types.StringValue(parentName),
			Index:                types.Int64Value(int64(i)),
			Type:                 stackRule.Type,
			Action:               stackRule.Action,
			Cond:                 stackRule.Cond,
			CondTest:             stackRule.CondTest,
			Expr:                 stackRule.Expr,
			Timeout:              stackRule.Timeout,
			LuaAction:            stackRule.LuaAction,
			LuaParams:            stackRule.LuaParams,
			LogLevel:             stackRule.LogLevel,
			MarkValue:            stackRule.MarkValue,
			NiceValue:            stackRule.NiceValue,
			TosValue:             stackRule.TosValue,
			CaptureLen:           stackRule.CaptureLen,
			CaptureSample:        stackRule.CaptureSample,
			BandwidthLimitLimit:  stackRule.BandwidthLimitLimit,
			BandwidthLimitName:   stackRule.BandwidthLimitName,
			BandwidthLimitPeriod: stackRule.BandwidthLimitPeriod,
			ResolveProtocol:      stackRule.ResolveProtocol,
			ResolveResolvers:     stackRule.ResolveResolvers,
			ResolveVar:           stackRule.ResolveVar,
			RstTtl:               stackRule.RstTtl,
			ScIdx:                stackRule.ScIdx,
			ScIncId:              stackRule.ScIncId,
			ScInt:                stackRule.ScInt,
			ServerName:           stackRule.ServerName,
			ServiceName:          stackRule.ServiceName,
			VarName:              stackRule.VarName,
			VarFormat:            stackRule.VarFormat,
			VarScope:             stackRule.VarScope,
			VarExpr:              stackRule.VarExpr,
		}
	}
	return resourceRules
}

// convertTcpResponseRulesToResourceModels converts stack models to resource models
func (o *StackOperations) convertTcpResponseRulesToResourceModels(stackRules []haproxyTcpResponseRuleModel, parentType, parentName string) []TcpResponseRuleResourceModel {
	resourceRules := make([]TcpResponseRuleResourceModel, len(stackRules))
	for i, stackRule := range stackRules {
		resourceRules[i] = TcpResponseRuleResourceModel{
			ID:                   types.StringValue(fmt.Sprintf("%s/%s/tcp_response_rule/%d", parentType, parentName, i)),
			ParentType:           types.StringValue(parentType),
			ParentName:           types.StringValue(parentName),
			Index:                types.Int64Value(int64(i)),
			Type:                 stackRule.Type,
			Action:               stackRule.Action,
			Cond:                 stackRule.Cond,
			CondTest:             stackRule.CondTest,
			Expr:                 stackRule.Expr,
			LogLevel:             stackRule.LogLevel,
			LuaAction:            stackRule.LuaAction,
			LuaParams:            stackRule.LuaParams,
			MarkValue:            stackRule.MarkValue,
			NiceValue:            stackRule.NiceValue,
			RstTtl:               stackRule.RstTtl,
			ScExpr:               stackRule.ScExpr,
			ScId:                 stackRule.ScId,
			ScIdx:                stackRule.ScIdx,
			ScInt:                stackRule.ScInt,
			SpoeEngine:           stackRule.SpoeEngine,
			SpoeGroup:            stackRule.SpoeGroup,
			Timeout:              stackRule.Timeout,
			TosValue:             stackRule.TosValue,
			VarFormat:            stackRule.VarFormat,
			VarName:              stackRule.VarName,
			VarScope:             stackRule.VarScope,
			VarExpr:              stackRule.VarExpr,
			BandwidthLimitLimit:  stackRule.BandwidthLimitLimit,
			BandwidthLimitName:   stackRule.BandwidthLimitName,
			BandwidthLimitPeriod: stackRule.BandwidthLimitPeriod,
		}
	}
	return resourceRules
}

// convertHttpchecksToResourceModels converts stack models to resource models
func (o *StackOperations) convertHttpchecksToResourceModels(stackChecks []haproxyHttpcheckModel, parentType, parentName string) []HttpcheckResourceModel {
	resourceChecks := make([]HttpcheckResourceModel, len(stackChecks))
	for i, stackCheck := range stackChecks {
		resourceChecks[i] = HttpcheckResourceModel{
			ID:              types.StringValue(fmt.Sprintf("%s/%s/httpcheck/%d", parentType, parentName, i)),
			ParentType:      types.StringValue(parentType),
			ParentName:      types.StringValue(parentName),
			Index:           types.Int64Value(int64(i)),
			Type:            stackCheck.Type,
			Addr:            stackCheck.Addr,
			Alpn:            stackCheck.Alpn,
			Body:            stackCheck.Body,
			BodyLogFormat:   stackCheck.BodyLogFormat,
			CheckComment:    stackCheck.CheckComment,
			Default:         stackCheck.Default,
			ErrorStatus:     stackCheck.ErrorStatus,
			ExclamationMark: stackCheck.ExclamationMark,
			Headers:         stackCheck.Headers,
			Linger:          stackCheck.Linger,
			Match:           stackCheck.Match,
			Method:          stackCheck.Method,
			MinRecv:         stackCheck.MinRecv,
			OkStatus:        stackCheck.OkStatus,
			OnError:         stackCheck.OnError,
			OnSuccess:       stackCheck.OnSuccess,
			Pattern:         stackCheck.Pattern,
			Port:            stackCheck.Port,
			PortString:      stackCheck.PortString,
			Proto:           stackCheck.Proto,
			SendProxy:       stackCheck.SendProxy,
			Sni:             stackCheck.Sni,
			Ssl:             stackCheck.Ssl,
			StatusCode:      stackCheck.StatusCode,
			ToutStatus:      stackCheck.ToutStatus,
			Uri:             stackCheck.Uri,
			UriLogFormat:    stackCheck.UriLogFormat,
			VarExpr:         stackCheck.VarExpr,
			VarFormat:       stackCheck.VarFormat,
			VarName:         stackCheck.VarName,
			VarScope:        stackCheck.VarScope,
			Version:         stackCheck.Version,
		}
	}
	return resourceChecks
}

// convertTcpChecksToResourceModels converts stack models to resource models
func (o *StackOperations) convertTcpChecksToResourceModels(stackChecks []haproxyTcpCheckModel, parentType, parentName string) []TcpCheckResourceModel {
	resourceChecks := make([]TcpCheckResourceModel, len(stackChecks))
	for i, stackCheck := range stackChecks {
		resourceChecks[i] = TcpCheckResourceModel{
			ID:              types.StringValue(fmt.Sprintf("%s/%s/tcp_check/%d", parentType, parentName, i)),
			ParentType:      types.StringValue(parentType),
			ParentName:      types.StringValue(parentName),
			Index:           types.Int64Value(int64(i)),
			Action:          stackCheck.Action,
			Addr:            stackCheck.Addr,
			Alpn:            stackCheck.Alpn,
			CheckComment:    stackCheck.CheckComment,
			Data:            stackCheck.Data,
			Default:         stackCheck.Default,
			ErrorStatus:     stackCheck.ErrorStatus,
			ExclamationMark: stackCheck.ExclamationMark,
			Fmt:             stackCheck.Fmt,
			HexFmt:          stackCheck.HexFmt,
			HexString:       stackCheck.HexString,
			Linger:          stackCheck.Linger,
			Match:           stackCheck.Match,
			MinRecv:         stackCheck.MinRecv,
			OkStatus:        stackCheck.OkStatus,
			OnError:         stackCheck.OnError,
			OnSuccess:       stackCheck.OnSuccess,
			Pattern:         stackCheck.Pattern,
			Port:            o.getTcpCheckPort(stackCheck),
			PortString:      stackCheck.PortString,
			Proto:           stackCheck.Proto,
			SendProxy:       stackCheck.SendProxy,
			Sni:             stackCheck.Sni,
			Ssl:             stackCheck.Ssl,
			StatusCode:      stackCheck.StatusCode,
			ToutStatus:      stackCheck.ToutStatus,
			VarExpr:         stackCheck.VarExpr,
			VarFmt:          stackCheck.VarFmt,
			VarName:         stackCheck.VarName,
			VarScope:        stackCheck.VarScope,
			ViaSocks4:       stackCheck.ViaSocks4,
		}
	}
	return resourceChecks
}

// getTcpCheckPort returns the appropriate port value for TCP checks
// Return the port value if it's set, otherwise return null
func (o *StackOperations) getTcpCheckPort(stackCheck haproxyTcpCheckModel) types.Int64 {
	if !stackCheck.Port.IsNull() && !stackCheck.Port.IsUnknown() {
		return stackCheck.Port
	}
	return types.Int64Null()
}

// convertTcpCheckPayloadToStackModel converts a TcpCheckPayload to haproxyTcpCheckModel
func (o *StackOperations) convertTcpCheckPayloadToStackModel(payload TcpCheckPayload) haproxyTcpCheckModel {
	model := haproxyTcpCheckModel{
		Action: types.StringValue(payload.Action),
	}

	// For connect actions, HAProxy combines addr and port into addr field as "addr:port"
	// We need to split this back to separate addr and port fields for Terraform state
	if payload.Action == "connect" && payload.Addr != "" {
		// Check if addr contains port (format: "addr:port")
		if strings.Contains(payload.Addr, ":") {
			parts := strings.Split(payload.Addr, ":")
			if len(parts) == 2 {
				model.Addr = types.StringValue(parts[0])
				if port, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					model.Port = types.Int64Value(port)
				}
			} else {
				model.Addr = types.StringValue(payload.Addr)
			}
		} else {
			model.Addr = types.StringValue(payload.Addr)
		}
	} else {
		// For other actions, use addr as-is
		if payload.Addr != "" {
			model.Addr = types.StringValue(payload.Addr)
		}
	}

	// For connect actions, HAProxy always returns port=0, so we don't set the port field
	// For other actions, set the port field if it's non-zero
	if payload.Action != "connect" && payload.Port > 0 {
		model.Port = types.Int64Value(payload.Port)
	}
	if payload.Alpn != "" {
		model.Alpn = types.StringValue(payload.Alpn)
	}
	if payload.CheckComment != "" {
		model.CheckComment = types.StringValue(payload.CheckComment)
	}
	if payload.Data != "" {
		model.Data = types.StringValue(payload.Data)
	}
	if payload.Default {
		model.Default = types.BoolValue(payload.Default)
	}
	if payload.ErrorStatus != "" {
		model.ErrorStatus = types.StringValue(payload.ErrorStatus)
	}
	if payload.ExclamationMark {
		model.ExclamationMark = types.BoolValue(payload.ExclamationMark)
	}
	if payload.Fmt != "" {
		model.Fmt = types.StringValue(payload.Fmt)
	}
	if payload.HexFmt != "" {
		model.HexFmt = types.StringValue(payload.HexFmt)
	}
	if payload.HexString != "" {
		model.HexString = types.StringValue(payload.HexString)
	}
	if payload.Linger {
		model.Linger = types.BoolValue(payload.Linger)
	}
	if payload.Match != "" {
		model.Match = types.StringValue(payload.Match)
	}
	if payload.MinRecv != 0 {
		model.MinRecv = types.Int64Value(payload.MinRecv)
	}
	if payload.OkStatus != "" {
		model.OkStatus = types.StringValue(payload.OkStatus)
	}
	if payload.OnError != "" {
		model.OnError = types.StringValue(payload.OnError)
	}
	if payload.OnSuccess != "" {
		model.OnSuccess = types.StringValue(payload.OnSuccess)
	}
	if payload.Pattern != "" {
		model.Pattern = types.StringValue(payload.Pattern)
	}
	if payload.PortString != "" {
		model.PortString = types.StringValue(payload.PortString)
	}
	if payload.Proto != "" {
		model.Proto = types.StringValue(payload.Proto)
	}
	if payload.SendProxy {
		model.SendProxy = types.BoolValue(payload.SendProxy)
	}
	if payload.Sni != "" {
		model.Sni = types.StringValue(payload.Sni)
	}
	if payload.Ssl {
		model.Ssl = types.BoolValue(payload.Ssl)
	}
	if payload.StatusCode != "" {
		model.StatusCode = types.StringValue(payload.StatusCode)
	}
	if payload.ToutStatus != "" {
		model.ToutStatus = types.StringValue(payload.ToutStatus)
	}
	if payload.VarExpr != "" {
		model.VarExpr = types.StringValue(payload.VarExpr)
	}
	if payload.VarFmt != "" {
		model.VarFmt = types.StringValue(payload.VarFmt)
	}
	if payload.VarName != "" {
		model.VarName = types.StringValue(payload.VarName)
	}
	if payload.VarScope != "" {
		model.VarScope = types.StringValue(payload.VarScope)
	}
	if payload.ViaSocks4 {
		model.ViaSocks4 = types.BoolValue(payload.ViaSocks4)
	}

	return model
}
