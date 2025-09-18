package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	tcpCheckTypeConnectValue = "connect"
)

// TcpCheckResource defines the resource implementation.
type TcpCheckResource struct {
	client *HAProxyClient
}

// TcpCheckResourceModel describes the resource data model.
type TcpCheckResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ParentType      types.String `tfsdk:"parent_type"`
	ParentName      types.String `tfsdk:"parent_name"`
	Index           types.Int64  `tfsdk:"index"`
	Action          types.String `tfsdk:"action"`
	Addr            types.String `tfsdk:"addr"`
	Alpn            types.String `tfsdk:"alpn"`
	CheckComment    types.String `tfsdk:"check_comment"`
	Data            types.String `tfsdk:"data"`
	Default         types.Bool   `tfsdk:"default"`
	ErrorStatus     types.String `tfsdk:"error_status"`
	ExclamationMark types.Bool   `tfsdk:"exclamation_mark"`
	Fmt             types.String `tfsdk:"fmt"`
	HexFmt          types.String `tfsdk:"hex_fmt"`
	HexString       types.String `tfsdk:"hex_string"`
	Linger          types.Bool   `tfsdk:"linger"`
	Match           types.String `tfsdk:"match"`
	MinRecv         types.Int64  `tfsdk:"min_recv"`
	OkStatus        types.String `tfsdk:"ok_status"`
	OnError         types.String `tfsdk:"on_error"`
	OnSuccess       types.String `tfsdk:"on_success"`
	Pattern         types.String `tfsdk:"pattern"`
	Port            types.Int64  `tfsdk:"port"`
	PortString      types.String `tfsdk:"port_string"`
	Proto           types.String `tfsdk:"proto"`
	SendProxy       types.Bool   `tfsdk:"send_proxy"`
	Sni             types.String `tfsdk:"sni"`
	Ssl             types.Bool   `tfsdk:"ssl"`
	StatusCode      types.String `tfsdk:"status_code"`
	ToutStatus      types.String `tfsdk:"tout_status"`
	VarExpr         types.String `tfsdk:"var_expr"`
	VarFmt          types.String `tfsdk:"var_fmt"`
	VarName         types.String `tfsdk:"var_name"`
	VarScope        types.String `tfsdk:"var_scope"`
	ViaSocks4       types.Bool   `tfsdk:"via_socks4"`
}

// TcpCheckManager manages TCP checks
type TcpCheckManager struct {
	client *HAProxyClient
}

// NewTcpCheckManager creates a new TCP check manager
func CreateTcpCheckManager(client *HAProxyClient) *TcpCheckManager {
	return &TcpCheckManager{
		client: client,
	}
}

// Create creates TCP checks
func (r *TcpCheckManager) Create(ctx context.Context, transactionID, parentType, parentName string, checks []TcpCheckResourceModel) error {
	if len(checks) == 0 {
		return nil
	}

	log.Printf("Creating %d TCP checks for %s %s", len(checks), parentType, parentName)

	// Sort checks by index to ensure proper ordering
	sortedChecks := r.processTcpCheckBlock(checks)

	// Convert all checks to payloads
	allPayloads := make([]TcpCheckPayload, 0, len(sortedChecks))
	for i, check := range sortedChecks {
		checkPayload := r.convertToTcpCheckPayload(&check, i)
		allPayloads = append(allPayloads, *checkPayload)
	}

	// Send all checks in one request (same for both v2 and v3)
	if err := r.client.CreateAllTcpChecksInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		// Return the original error to preserve error type for retry logic
		return err
	}

	log.Printf("Created all %d TCP checks for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
	return nil
}

// Read reads TCP checks
func (r *TcpCheckManager) Read(ctx context.Context, parentType, parentName string) ([]TcpCheckResourceModel, error) {
	log.Printf("Reading TCP checks for %s %s", parentType, parentName)

	payloads, err := r.client.ReadTcpChecks(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read TCP checks for %s %s: %w", parentType, parentName, err)
	}

	// Convert payloads to resource models
	checks := make([]TcpCheckResourceModel, 0, len(payloads))
	for _, payload := range payloads {
		check := r.convertFromTcpCheckPayload(payload, parentType, parentName)
		checks = append(checks, check)
	}

	log.Printf("Read %d TCP checks for %s %s", len(checks), parentType, parentName)
	return checks, nil
}

// Update updates TCP checks
func (r *TcpCheckManager) Update(ctx context.Context, transactionID, parentType, parentName string, checks []TcpCheckResourceModel) error {
	if len(checks) == 0 {
		// If no checks, delete all existing checks
		return r.Delete(ctx, transactionID, parentType, parentName)
	}

	log.Printf("Updating %d TCP checks for %s %s", len(checks), parentType, parentName)

	// Sort new checks by index to ensure proper ordering
	sortedChecks := r.processTcpCheckBlock(checks)

	// Convert new checks to payloads
	desiredPayloads := make([]TcpCheckPayload, 0, len(sortedChecks))
	for i, check := range sortedChecks {
		checkPayload := r.convertToTcpCheckPayload(&check, i)
		desiredPayloads = append(desiredPayloads, *checkPayload)
	}

	// Use delete-all-then-create-all pattern (same as http_request_rules)
	// First, delete all existing TCP checks to avoid duplicates
	if err := r.deleteAllTcpChecksInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing TCP checks for %s %s: %w", parentType, parentName, err)
	}

	// Then create all desired checks using the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
	if err := r.client.CreateAllTcpChecksInTransaction(ctx, transactionID, parentType, parentName, desiredPayloads); err != nil {
		return fmt.Errorf("failed to create new TCP checks for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Updated %d TCP checks for %s %s in transaction %s (delete-then-create)", len(desiredPayloads), parentType, parentName, transactionID)
	return nil
}

// Delete deletes TCP checks
func (r *TcpCheckManager) Delete(ctx context.Context, transactionID, parentType, parentName string) error {
	log.Printf("Deleting all TCP checks for %s %s", parentType, parentName)

	// Read existing checks to get their indices
	existingChecks, err := r.client.ReadTcpChecks(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing TCP checks for deletion: %w", err)
	}

	// Delete each check by index in reverse order (highest index first) to avoid shifting issues
	sort.Slice(existingChecks, func(i, j int) bool {
		return existingChecks[i].Index > existingChecks[j].Index
	})

	for _, check := range existingChecks {
		if err := r.client.DeleteTcpCheckInTransaction(ctx, transactionID, check.Index, parentType, parentName); err != nil {
			return fmt.Errorf("failed to delete TCP check at index %d: %w", check.Index, err)
		}
	}

	log.Printf("Deleted %d TCP checks for %s %s", len(existingChecks), parentType, parentName)
	return nil
}

// deleteAllTcpChecksInTransaction deletes all TCP checks for a parent resource using an existing transaction ID
func (r *TcpCheckManager) deleteAllTcpChecksInTransaction(ctx context.Context, transactionID string, parentType, parentName string) error {
	checks, err := r.client.ReadTcpChecks(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read TCP checks for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Index > checks[j].Index
	})

	for _, check := range checks {
		log.Printf("Deleting TCP check at index %d in transaction %s", check.Index, transactionID)
		err := r.client.DeleteTcpCheckInTransaction(ctx, transactionID, check.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete TCP check at index %d: %w", check.Index, err)
		}
	}

	return nil
}

// processTcpCheckBlock processes and sorts TCP checks
func (r *TcpCheckManager) processTcpCheckBlock(checks []TcpCheckResourceModel) []TcpCheckResourceModel {
	// Sort checks by index to ensure proper ordering
	sortedChecks := make([]TcpCheckResourceModel, len(checks))
	copy(sortedChecks, checks)

	// Sort by index
	for i := 0; i < len(sortedChecks)-1; i++ {
		for j := i + 1; j < len(sortedChecks); j++ {
			if sortedChecks[i].Index.ValueInt64() > sortedChecks[j].Index.ValueInt64() {
				sortedChecks[i], sortedChecks[j] = sortedChecks[j], sortedChecks[i]
			}
		}
	}

	return sortedChecks
}

// convertToTcpCheckPayload converts a resource model to a payload
func (r *TcpCheckManager) convertToTcpCheckPayload(check *TcpCheckResourceModel, index int) *TcpCheckPayload {
	payload := &TcpCheckPayload{
		Index:  int64(index),
		Action: check.Action.ValueString(),
	}

	// Set optional fields if they have values
	if !check.Addr.IsNull() && !check.Addr.IsUnknown() {
		addr := check.Addr.ValueString()
		// For connect action, include port in addr field if both addr and port are specified
		if check.Action.ValueString() == tcpCheckTypeConnectValue && !check.Port.IsNull() && !check.Port.IsUnknown() && check.Port.ValueInt64() > 0 {
			port := check.Port.ValueInt64()
			// Only add port if it's not already in the addr
			if !strings.Contains(addr, ":") {
				addr = fmt.Sprintf("%s:%d", addr, port)
			}
		}
		payload.Addr = addr
	}
	if !check.Alpn.IsNull() && !check.Alpn.IsUnknown() {
		payload.Alpn = check.Alpn.ValueString()
	}
	if !check.CheckComment.IsNull() && !check.CheckComment.IsUnknown() {
		payload.CheckComment = check.CheckComment.ValueString()
	}
	if !check.Data.IsNull() && !check.Data.IsUnknown() {
		payload.Data = check.Data.ValueString()
	}
	if !check.Default.IsNull() && !check.Default.IsUnknown() {
		payload.Default = check.Default.ValueBool()
	}
	if !check.ErrorStatus.IsNull() && !check.ErrorStatus.IsUnknown() {
		payload.ErrorStatus = check.ErrorStatus.ValueString()
	}
	if !check.ExclamationMark.IsNull() && !check.ExclamationMark.IsUnknown() {
		payload.ExclamationMark = check.ExclamationMark.ValueBool()
	}
	if !check.Fmt.IsNull() && !check.Fmt.IsUnknown() {
		payload.Fmt = check.Fmt.ValueString()
	}
	if !check.HexFmt.IsNull() && !check.HexFmt.IsUnknown() {
		payload.HexFmt = check.HexFmt.ValueString()
	}
	if !check.HexString.IsNull() && !check.HexString.IsUnknown() {
		payload.HexString = check.HexString.ValueString()
	}
	if !check.Linger.IsNull() && !check.Linger.IsUnknown() {
		payload.Linger = check.Linger.ValueBool()
	}
	if !check.Match.IsNull() && !check.Match.IsUnknown() {
		payload.Match = check.Match.ValueString()
	}
	if !check.MinRecv.IsNull() && !check.MinRecv.IsUnknown() {
		payload.MinRecv = check.MinRecv.ValueInt64()
	}
	if !check.OkStatus.IsNull() && !check.OkStatus.IsUnknown() {
		payload.OkStatus = check.OkStatus.ValueString()
	}
	if !check.OnError.IsNull() && !check.OnError.IsUnknown() {
		payload.OnError = check.OnError.ValueString()
	}
	if !check.OnSuccess.IsNull() && !check.OnSuccess.IsUnknown() {
		payload.OnSuccess = check.OnSuccess.ValueString()
	}
	if !check.Pattern.IsNull() && !check.Pattern.IsUnknown() {
		payload.Pattern = check.Pattern.ValueString()
	}
	if !check.Port.IsNull() && !check.Port.IsUnknown() {
		port := check.Port.ValueInt64()
		action := check.Action.ValueString()

		if action == tcpCheckTypeConnectValue {
			// For connect actions, don't set port field at all
			// Port is handled in the addr field as addr:port
		} else {
			// For other actions, set port field
			if port > 0 {
				payload.Port = port
			} else {
				// For send/expect actions, set port to 1 to satisfy HAProxy's requirement
				payload.Port = 1
			}
		}
	}
	if !check.PortString.IsNull() && !check.PortString.IsUnknown() {
		payload.PortString = check.PortString.ValueString()
	}
	if !check.Proto.IsNull() && !check.Proto.IsUnknown() {
		payload.Proto = check.Proto.ValueString()
	}
	if !check.SendProxy.IsNull() && !check.SendProxy.IsUnknown() {
		payload.SendProxy = check.SendProxy.ValueBool()
	}
	if !check.Sni.IsNull() && !check.Sni.IsUnknown() {
		payload.Sni = check.Sni.ValueString()
	}
	if !check.Ssl.IsNull() && !check.Ssl.IsUnknown() {
		payload.Ssl = check.Ssl.ValueBool()
	}
	if !check.StatusCode.IsNull() && !check.StatusCode.IsUnknown() {
		payload.StatusCode = check.StatusCode.ValueString()
	}
	if !check.ToutStatus.IsNull() && !check.ToutStatus.IsUnknown() {
		payload.ToutStatus = check.ToutStatus.ValueString()
	}
	if !check.VarExpr.IsNull() && !check.VarExpr.IsUnknown() {
		payload.VarExpr = check.VarExpr.ValueString()
	}
	if !check.VarFmt.IsNull() && !check.VarFmt.IsUnknown() {
		payload.VarFmt = check.VarFmt.ValueString()
	}
	if !check.VarName.IsNull() && !check.VarName.IsUnknown() {
		payload.VarName = check.VarName.ValueString()
	}
	if !check.VarScope.IsNull() && !check.VarScope.IsUnknown() {
		payload.VarScope = check.VarScope.ValueString()
	}
	if !check.ViaSocks4.IsNull() && !check.ViaSocks4.IsUnknown() {
		payload.ViaSocks4 = check.ViaSocks4.ValueBool()
	}

	return payload
}

// convertFromTcpCheckPayload converts a payload to a resource model
func (r *TcpCheckManager) convertFromTcpCheckPayload(payload TcpCheckPayload, parentType, parentName string) TcpCheckResourceModel {
	check := TcpCheckResourceModel{
		ID:         types.StringValue(fmt.Sprintf("%s/%s/tcp_check/%d", parentType, parentName, payload.Index)),
		ParentType: types.StringValue(parentType),
		ParentName: types.StringValue(parentName),
		Index:      types.Int64Value(payload.Index),
		Action:     types.StringValue(payload.Action),
	}

	// For connect actions, HAProxy combines addr and port into addr field as "addr:port"
	// We need to split this back to separate addr and port fields for Terraform state
	if payload.Action == tcpCheckTypeConnectValue && payload.Addr != "" {
		// Check if addr contains port (format: "addr:port")
		if strings.Contains(payload.Addr, ":") {
			parts := strings.Split(payload.Addr, ":")
			if len(parts) == 2 {
				check.Addr = types.StringValue(parts[0])
				if port, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					check.Port = types.Int64Value(port)
				}
			} else {
				check.Addr = types.StringValue(payload.Addr)
			}
		} else {
			check.Addr = types.StringValue(payload.Addr)
		}
	} else {
		// For other actions, use addr as-is
		if payload.Addr != "" {
			check.Addr = types.StringValue(payload.Addr)
		}
	}

	// For connect actions, HAProxy always returns port=0, so we don't set the port field
	// For other actions, set the port field if it's non-zero
	if payload.Action != tcpCheckTypeConnectValue && payload.Port > 0 {
		check.Port = types.Int64Value(payload.Port)
	}
	if payload.Alpn != "" {
		check.Alpn = types.StringValue(payload.Alpn)
	}
	if payload.CheckComment != "" {
		check.CheckComment = types.StringValue(payload.CheckComment)
	}
	if payload.Data != "" {
		check.Data = types.StringValue(payload.Data)
	}
	if payload.Default {
		check.Default = types.BoolValue(payload.Default)
	}
	if payload.ErrorStatus != "" {
		check.ErrorStatus = types.StringValue(payload.ErrorStatus)
	}
	if payload.ExclamationMark {
		check.ExclamationMark = types.BoolValue(payload.ExclamationMark)
	}
	if payload.Fmt != "" {
		check.Fmt = types.StringValue(payload.Fmt)
	}
	if payload.HexFmt != "" {
		check.HexFmt = types.StringValue(payload.HexFmt)
	}
	if payload.HexString != "" {
		check.HexString = types.StringValue(payload.HexString)
	}
	if payload.Linger {
		check.Linger = types.BoolValue(payload.Linger)
	}
	if payload.Match != "" {
		check.Match = types.StringValue(payload.Match)
	}
	if payload.MinRecv != 0 {
		check.MinRecv = types.Int64Value(payload.MinRecv)
	}
	if payload.OkStatus != "" {
		check.OkStatus = types.StringValue(payload.OkStatus)
	}
	if payload.OnError != "" {
		check.OnError = types.StringValue(payload.OnError)
	}
	if payload.OnSuccess != "" {
		check.OnSuccess = types.StringValue(payload.OnSuccess)
	}
	if payload.Pattern != "" {
		check.Pattern = types.StringValue(payload.Pattern)
	}
	if payload.Port != 0 {
		check.Port = types.Int64Value(payload.Port)
	}
	if payload.PortString != "" {
		check.PortString = types.StringValue(payload.PortString)
	}
	if payload.Proto != "" {
		check.Proto = types.StringValue(payload.Proto)
	}
	if payload.SendProxy {
		check.SendProxy = types.BoolValue(payload.SendProxy)
	}
	if payload.Sni != "" {
		check.Sni = types.StringValue(payload.Sni)
	}
	if payload.Ssl {
		check.Ssl = types.BoolValue(payload.Ssl)
	}
	if payload.StatusCode != "" {
		check.StatusCode = types.StringValue(payload.StatusCode)
	}
	if payload.ToutStatus != "" {
		check.ToutStatus = types.StringValue(payload.ToutStatus)
	}
	if payload.VarExpr != "" {
		check.VarExpr = types.StringValue(payload.VarExpr)
	}
	if payload.VarFmt != "" {
		check.VarFmt = types.StringValue(payload.VarFmt)
	}
	if payload.VarName != "" {
		check.VarName = types.StringValue(payload.VarName)
	}
	if payload.VarScope != "" {
		check.VarScope = types.StringValue(payload.VarScope)
	}
	if payload.ViaSocks4 {
		check.ViaSocks4 = types.BoolValue(payload.ViaSocks4)
	}

	return check
}

// Metadata returns the resource type name.
func (r *TcpCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_check"
}

// GetTcpCheckSchema returns the schema for the tcp_check block
func GetTcpCheckSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "TCP check configuration.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"action": schema.StringAttribute{
					Required:    true,
					Description: "Check action.",
				},
				"addr": schema.StringAttribute{
					Optional:    true,
					Description: "Address.",
				},
				"alpn": schema.StringAttribute{
					Optional:    true,
					Description: "ALPN.",
				},
				"check_comment": schema.StringAttribute{
					Optional:    true,
					Description: "Check comment.",
				},
				"data": schema.StringAttribute{
					Optional:    true,
					Description: "Data.",
				},
				"default": schema.BoolAttribute{
					Optional:    true,
					Description: "Default.",
				},
				"error_status": schema.StringAttribute{
					Optional:    true,
					Description: "Error status.",
				},
				"exclamation_mark": schema.BoolAttribute{
					Optional:    true,
					Description: "Exclamation mark.",
				},
				"fmt": schema.StringAttribute{
					Optional:    true,
					Description: "Format.",
				},
				"hex_fmt": schema.StringAttribute{
					Optional:    true,
					Description: "Hex format.",
				},
				"hex_string": schema.StringAttribute{
					Optional:    true,
					Description: "Hex string.",
				},
				"linger": schema.BoolAttribute{
					Optional:    true,
					Description: "Linger.",
				},
				"match": schema.StringAttribute{
					Optional:    true,
					Description: "Match.",
				},
				"min_recv": schema.Int64Attribute{
					Optional:    true,
					Description: "Min recv.",
				},
				"ok_status": schema.StringAttribute{
					Optional:    true,
					Description: "OK status.",
				},
				"on_error": schema.StringAttribute{
					Optional:    true,
					Description: "On error.",
				},
				"on_success": schema.StringAttribute{
					Optional:    true,
					Description: "On success.",
				},
				"pattern": schema.StringAttribute{
					Optional:    true,
					Description: "Pattern.",
				},
				"port": schema.Int64Attribute{
					Optional:    true,
					Description: "Port.",
				},
				"port_string": schema.StringAttribute{
					Optional:    true,
					Description: "Port string.",
				},
				"proto": schema.StringAttribute{
					Optional:    true,
					Description: "Protocol.",
				},
				"send_proxy": schema.BoolAttribute{
					Optional:    true,
					Description: "Send proxy.",
				},
				"sni": schema.StringAttribute{
					Optional:    true,
					Description: "SNI.",
				},
				"ssl": schema.BoolAttribute{
					Optional:    true,
					Description: "SSL.",
				},
				"status_code": schema.StringAttribute{
					Optional:    true,
					Description: "Status code.",
				},
				"tout_status": schema.StringAttribute{
					Optional:    true,
					Description: "Timeout status.",
				},
				"var_expr": schema.StringAttribute{
					Optional:    true,
					Description: "Variable expression.",
				},
				"var_fmt": schema.StringAttribute{
					Optional:    true,
					Description: "Variable format.",
				},
				"var_name": schema.StringAttribute{
					Optional:    true,
					Description: "Variable name.",
				},
				"var_scope": schema.StringAttribute{
					Optional:    true,
					Description: "Variable scope.",
				},
				"via_socks4": schema.BoolAttribute{
					Optional:    true,
					Description: "Via SOCKS4.",
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *TcpCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *TcpCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TcpCheckResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Individual TCP check resources should not be used - use haproxy_stack instead
	resp.Diagnostics.AddError("Invalid Usage", "TCP check resources should not be used directly. Use haproxy_stack resource instead.")
}

// Read refreshes the Terraform state with the latest data.
func (r *TcpCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TcpCheckResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the check
	manager := CreateTcpCheckManager(r.client)
	checks, err := manager.Read(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP check, got error: %s", err))
	}

	// Find the specific check by index
	var foundCheck *TcpCheckResourceModel
	for _, check := range checks {
		if check.Index.ValueInt64() == data.Index.ValueInt64() {
			foundCheck = &check
			break
		}
	}

	if foundCheck == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the data with the found check
	data = *foundCheck

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *TcpCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TcpCheckResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Individual TCP check resources should not be used - use haproxy_stack instead
	resp.Diagnostics.AddError("Invalid Usage", "TCP check resources should not be used directly. Use haproxy_stack resource instead.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *TcpCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TcpCheckResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Individual TCP check resources should not be used - use haproxy_stack instead
	resp.Diagnostics.AddError("Invalid Usage", "TCP check resources should not be used directly. Use haproxy_stack resource instead.")
}
