package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HttpcheckResource defines the resource implementation.
type HttpcheckResource struct {
	client *HAProxyClient
}

// HttpcheckResourceModel describes the resource data model.
type HttpcheckResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ParentType      types.String `tfsdk:"parent_type"`
	ParentName      types.String `tfsdk:"parent_name"`
	Index           types.Int64  `tfsdk:"index"`
	Type            types.String `tfsdk:"type"`
	Addr            types.String `tfsdk:"addr"`
	Alpn            types.String `tfsdk:"alpn"`
	Body            types.String `tfsdk:"body"`
	BodyLogFormat   types.String `tfsdk:"body_log_format"`
	CheckComment    types.String `tfsdk:"check_comment"`
	Default         types.Bool   `tfsdk:"default"`
	ErrorStatus     types.String `tfsdk:"error_status"`
	ExclamationMark types.Bool   `tfsdk:"exclamation_mark"`
	Headers         types.List   `tfsdk:"headers"`
	Linger          types.Bool   `tfsdk:"linger"`
	Match           types.String `tfsdk:"match"`
	Method          types.String `tfsdk:"method"`
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
	Uri             types.String `tfsdk:"uri"`
	UriLogFormat    types.String `tfsdk:"uri_log_format"`
	VarExpr         types.String `tfsdk:"var_expr"`
	VarFormat       types.String `tfsdk:"var_format"`
	VarName         types.String `tfsdk:"var_name"`
	VarScope        types.String `tfsdk:"var_scope"`
	Version         types.String `tfsdk:"version"`
	ViaSocks4       types.Bool   `tfsdk:"via_socks4"`
}

// HttpcheckManager manages HTTP checks
type HttpcheckManager struct {
	client *HAProxyClient
}

// NewHttpcheckManager creates a new HTTP check manager
func CreateHttpcheckManager(client *HAProxyClient) *HttpcheckManager {
	return &HttpcheckManager{
		client: client,
	}
}

// Create creates HTTP checks
func (r *HttpcheckManager) Create(ctx context.Context, transactionID, parentType, parentName string, checks []HttpcheckResourceModel) error {
	if len(checks) == 0 {
		return nil
	}

	log.Printf("Creating %d HTTP checks for %s %s", len(checks), parentType, parentName)

	// Sort checks by index to ensure proper ordering
	sortedChecks := r.processHttpcheckBlock(checks)

	// Convert all checks to payloads
	allPayloads := make([]HttpcheckPayload, 0, len(sortedChecks))
	for i, check := range sortedChecks {
		checkPayload := r.convertToHttpcheckPayload(&check, i)
		allPayloads = append(allPayloads, *checkPayload)
	}

	// Send all checks in one request (same for both v2 and v3)
	if err := r.client.CreateAllHttpchecksInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		// Return the original error to preserve error type for retry logic
		return err
	}

	log.Printf("Created all %d HTTP checks for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
	return nil
}

// Read reads HTTP checks
func (r *HttpcheckManager) Read(ctx context.Context, parentType, parentName string) ([]HttpcheckResourceModel, error) {
	log.Printf("Reading HTTP checks for %s %s", parentType, parentName)

	payloads, err := r.client.ReadHttpchecks(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP checks for %s %s: %w", parentType, parentName, err)
	}

	// Convert payloads to resource models
	checks := make([]HttpcheckResourceModel, 0, len(payloads))
	for _, payload := range payloads {
		check := r.convertFromHttpcheckPayload(payload, parentType, parentName)
		checks = append(checks, check)
	}

	log.Printf("Read %d HTTP checks for %s %s", len(checks), parentType, parentName)
	return checks, nil
}

// Update updates HTTP checks
func (r *HttpcheckManager) Update(ctx context.Context, transactionID, parentType, parentName string, checks []HttpcheckResourceModel) error {
	if len(checks) == 0 {
		// If no checks, delete all existing checks
		return r.Delete(ctx, transactionID, parentType, parentName)
	}

	log.Printf("Updating %d HTTP checks for %s %s", len(checks), parentType, parentName)

	// Sort new checks by index to ensure proper ordering
	sortedChecks := r.processHttpcheckBlock(checks)

	// Convert new checks to payloads
	desiredPayloads := make([]HttpcheckPayload, 0, len(sortedChecks))
	for i, check := range sortedChecks {
		checkPayload := r.convertToHttpcheckPayload(&check, i)
		desiredPayloads = append(desiredPayloads, *checkPayload)
	}

	// Use delete-all-then-create-all pattern (same as http_request_rules)
	// First, delete all existing HTTP checks to avoid duplicates
	if err := r.deleteAllHttpchecksInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing HTTP checks for %s %s: %w", parentType, parentName, err)
	}

	// Then create all desired checks using the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
	if err := r.client.CreateAllHttpchecksInTransaction(ctx, transactionID, parentType, parentName, desiredPayloads); err != nil {
		return fmt.Errorf("failed to create new HTTP checks for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Updated %d HTTP checks for %s %s in transaction %s (delete-then-create)", len(desiredPayloads), parentType, parentName, transactionID)
	return nil
}

// Delete deletes HTTP checks
func (r *HttpcheckManager) Delete(ctx context.Context, transactionID, parentType, parentName string) error {
	log.Printf("Deleting all HTTP checks for %s %s", parentType, parentName)

	// Read existing checks to get their indices
	existingChecks, err := r.client.ReadHttpchecks(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing HTTP checks for deletion: %w", err)
	}

	// Delete each check by index in reverse order (highest index first) to avoid shifting issues
	sort.Slice(existingChecks, func(i, j int) bool {
		return existingChecks[i].Index > existingChecks[j].Index
	})

	for _, check := range existingChecks {
		if err := r.client.DeleteHttpcheckInTransaction(ctx, transactionID, check.Index, parentType, parentName); err != nil {
			return fmt.Errorf("failed to delete HTTP check at index %d: %w", check.Index, err)
		}
	}

	log.Printf("Deleted %d HTTP checks for %s %s", len(existingChecks), parentType, parentName)
	return nil
}

// deleteAllHttpchecksInTransaction deletes all HTTP checks for a parent resource using an existing transaction ID
func (r *HttpcheckManager) deleteAllHttpchecksInTransaction(ctx context.Context, transactionID string, parentType, parentName string) error {
	checks, err := r.client.ReadHttpchecks(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read HTTP checks for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(checks, func(i, j int) bool {
		return checks[i].Index > checks[j].Index
	})

	for _, check := range checks {
		log.Printf("Deleting HTTP check at index %d in transaction %s", check.Index, transactionID)
		err := r.client.DeleteHttpcheckInTransaction(ctx, transactionID, check.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete HTTP check at index %d: %w", check.Index, err)
		}
	}

	return nil
}

// processHttpcheckBlock processes and sorts HTTP checks
func (r *HttpcheckManager) processHttpcheckBlock(checks []HttpcheckResourceModel) []HttpcheckResourceModel {
	// Sort checks by index to ensure proper ordering
	sortedChecks := make([]HttpcheckResourceModel, len(checks))
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

// convertToHttpcheckPayload converts a resource model to a payload
func (r *HttpcheckManager) convertToHttpcheckPayload(check *HttpcheckResourceModel, index int) *HttpcheckPayload {
	payload := &HttpcheckPayload{
		Index: int64(index),
		Type:  check.Type.ValueString(),
	}

	// Set optional fields if they have values
	if !check.Addr.IsNull() && !check.Addr.IsUnknown() {
		payload.Addr = check.Addr.ValueString()
	}
	if !check.Alpn.IsNull() && !check.Alpn.IsUnknown() {
		payload.Alpn = check.Alpn.ValueString()
	}
	if !check.Body.IsNull() && !check.Body.IsUnknown() {
		payload.Body = check.Body.ValueString()
	}
	if !check.BodyLogFormat.IsNull() && !check.BodyLogFormat.IsUnknown() {
		payload.BodyLogFormat = check.BodyLogFormat.ValueString()
	}
	if !check.CheckComment.IsNull() && !check.CheckComment.IsUnknown() {
		payload.CheckComment = check.CheckComment.ValueString()
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
	if !check.Headers.IsNull() && !check.Headers.IsUnknown() {
		// Convert headers list to CheckHdr slice
		headers := make([]CheckHdr, 0, len(check.Headers.Elements()))
		for _, header := range check.Headers.Elements() {
			if !header.IsNull() && !header.IsUnknown() {
				headers = append(headers, CheckHdr{Name: header.(types.String).ValueString()})
			}
		}
		payload.Headers = headers
	}
	if !check.Linger.IsNull() && !check.Linger.IsUnknown() {
		payload.Linger = check.Linger.ValueBool()
	}
	if !check.Match.IsNull() && !check.Match.IsUnknown() {
		payload.Match = check.Match.ValueString()
	}
	if !check.Method.IsNull() && !check.Method.IsUnknown() {
		payload.Method = check.Method.ValueString()
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
		payload.Port = check.Port.ValueInt64()
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
	if !check.Uri.IsNull() && !check.Uri.IsUnknown() {
		payload.Uri = check.Uri.ValueString()
	}
	if !check.UriLogFormat.IsNull() && !check.UriLogFormat.IsUnknown() {
		payload.UriLogFormat = check.UriLogFormat.ValueString()
	}
	if !check.VarExpr.IsNull() && !check.VarExpr.IsUnknown() {
		payload.VarExpr = check.VarExpr.ValueString()
	}
	if !check.VarFormat.IsNull() && !check.VarFormat.IsUnknown() {
		payload.VarFormat = check.VarFormat.ValueString()
	}
	if !check.VarName.IsNull() && !check.VarName.IsUnknown() {
		payload.VarName = check.VarName.ValueString()
	}
	if !check.VarScope.IsNull() && !check.VarScope.IsUnknown() {
		payload.VarScope = check.VarScope.ValueString()
	}
	if !check.Version.IsNull() && !check.Version.IsUnknown() {
		payload.Version = check.Version.ValueString()
	}
	if !check.ViaSocks4.IsNull() && !check.ViaSocks4.IsUnknown() {
		payload.ViaSocks4 = check.ViaSocks4.ValueBool()
	}

	// TODO: Handle Headers field (types.List) - this would require more complex conversion

	return payload
}

// convertFromHttpcheckPayload converts a payload to a resource model
func (r *HttpcheckManager) convertFromHttpcheckPayload(payload HttpcheckPayload, parentType, parentName string) HttpcheckResourceModel {
	check := HttpcheckResourceModel{
		ID:         types.StringValue(fmt.Sprintf("%s/%s/httpcheck/%d", parentType, parentName, payload.Index)),
		ParentType: types.StringValue(parentType),
		ParentName: types.StringValue(parentName),
		Index:      types.Int64Value(payload.Index),
		Type:       types.StringValue(payload.Type),
	}

	// Set optional fields
	if payload.Addr != "" {
		check.Addr = types.StringValue(payload.Addr)
	}
	if payload.Alpn != "" {
		check.Alpn = types.StringValue(payload.Alpn)
	}
	if payload.Body != "" {
		check.Body = types.StringValue(payload.Body)
	}
	if payload.BodyLogFormat != "" {
		check.BodyLogFormat = types.StringValue(payload.BodyLogFormat)
	}
	if payload.CheckComment != "" {
		check.CheckComment = types.StringValue(payload.CheckComment)
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
	if len(payload.Headers) > 0 {
		// Convert CheckHdr slice to headers list
		headerStrings := make([]attr.Value, 0, len(payload.Headers))
		for _, header := range payload.Headers {
			headerStrings = append(headerStrings, types.StringValue(header.Name))
		}
		check.Headers = types.ListValueMust(types.StringType, headerStrings)
	}
	if payload.Linger {
		check.Linger = types.BoolValue(payload.Linger)
	}
	if payload.Match != "" {
		check.Match = types.StringValue(payload.Match)
	}
	if payload.Method != "" {
		check.Method = types.StringValue(payload.Method)
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
	if payload.Uri != "" {
		check.Uri = types.StringValue(payload.Uri)
	}
	if payload.UriLogFormat != "" {
		check.UriLogFormat = types.StringValue(payload.UriLogFormat)
	}
	if payload.VarExpr != "" {
		check.VarExpr = types.StringValue(payload.VarExpr)
	}
	if payload.VarFormat != "" {
		check.VarFormat = types.StringValue(payload.VarFormat)
	}
	if payload.VarName != "" {
		check.VarName = types.StringValue(payload.VarName)
	}
	if payload.VarScope != "" {
		check.VarScope = types.StringValue(payload.VarScope)
	}
	if payload.Version != "" {
		check.Version = types.StringValue(payload.Version)
	}
	if payload.ViaSocks4 {
		check.ViaSocks4 = types.BoolValue(payload.ViaSocks4)
	}

	// TODO: Handle Headers field - this would require more complex conversion

	return check
}

// Metadata returns the resource type name.
func (r *HttpcheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_httpcheck"
}

// GetHttpcheckSchema returns the schema for the httpcheck block
func GetHttpcheckSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "HTTP check configuration.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Required:    true,
					Description: "Check type.",
				},
				"addr": schema.StringAttribute{
					Optional:    true,
					Description: "Address.",
				},
				"alpn": schema.StringAttribute{
					Optional:    true,
					Description: "ALPN.",
				},
				"body": schema.StringAttribute{
					Optional:    true,
					Description: "Body.",
				},
				"body_log_format": schema.StringAttribute{
					Optional:    true,
					Description: "Body log format.",
				},
				"check_comment": schema.StringAttribute{
					Optional:    true,
					Description: "Check comment.",
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
				"headers": schema.ListAttribute{
					Optional:    true,
					ElementType: types.StringType,
					Description: "Headers.",
				},
				"linger": schema.BoolAttribute{
					Optional:    true,
					Description: "Linger.",
				},
				"match": schema.StringAttribute{
					Optional:    true,
					Description: "Match.",
				},
				"method": schema.StringAttribute{
					Optional:    true,
					Description: "Method.",
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
				"uri": schema.StringAttribute{
					Optional:    true,
					Description: "URI.",
				},
				"uri_log_format": schema.StringAttribute{
					Optional:    true,
					Description: "URI log format.",
				},
				"var_expr": schema.StringAttribute{
					Optional:    true,
					Description: "Variable expression.",
				},
				"var_format": schema.StringAttribute{
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
				"version": schema.StringAttribute{
					Optional:    true,
					Description: "Version.",
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
func (r *HttpcheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *HttpcheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HttpcheckResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Individual HTTP check resources should only be used within haproxy_stack context
	// This resource is not registered and should not be used standalone
	resp.Diagnostics.AddError("Invalid Usage", "HTTP check resources should only be used within haproxy_stack context. Use haproxy_stack resource instead.")
}

// Read refreshes the Terraform state with the latest data.
func (r *HttpcheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HttpcheckResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the check
	manager := CreateHttpcheckManager(r.client)
	checks, err := manager.Read(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HTTP check, got error: %s", err))
	}

	// Find the specific check by index
	var foundCheck *HttpcheckResourceModel
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
func (r *HttpcheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HttpcheckResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Individual HTTP check resources should only be used within haproxy_stack context
	// This resource is not registered and should not be used standalone
	resp.Diagnostics.AddError("Invalid Usage", "HTTP check resources should only be used within haproxy_stack context. Use haproxy_stack resource instead.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *HttpcheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HttpcheckResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Individual HTTP check resources should only be used within haproxy_stack context
	// This resource is not registered and should not be used standalone
	resp.Diagnostics.AddError("Invalid Usage", "HTTP check resources should only be used within haproxy_stack context. Use haproxy_stack resource instead.")
}
