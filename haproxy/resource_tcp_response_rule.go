package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TcpResponseRuleResource defines the resource implementation.
type TcpResponseRuleResource struct {
	client *HAProxyClient
}

// TcpResponseRuleResourceModel describes the resource data model.
type TcpResponseRuleResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ParentType           types.String `tfsdk:"parent_type"`
	ParentName           types.String `tfsdk:"parent_name"`
	Index                types.Int64  `tfsdk:"index"`
	Type                 types.String `tfsdk:"type"`
	Action               types.String `tfsdk:"action"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	Expr                 types.String `tfsdk:"expr"`
	LogLevel             types.String `tfsdk:"log_level"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	MarkValue            types.String `tfsdk:"mark_value"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	RstTtl               types.Int64  `tfsdk:"rst_ttl"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScId                 types.Int64  `tfsdk:"sc_id"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	TosValue             types.String `tfsdk:"tos_value"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	VarExpr              types.String `tfsdk:"var_expr"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
}

// TcpResponseRuleManager manages TCP response rules
type TcpResponseRuleManager struct {
	client *HAProxyClient
}

// NewTcpResponseRuleManager creates a new TCP response rule manager
func CreateTcpResponseRuleManager(client *HAProxyClient) *TcpResponseRuleManager {
	return &TcpResponseRuleManager{
		client: client,
	}
}

// Create creates TCP response rules
func (r *TcpResponseRuleManager) Create(ctx context.Context, transactionID, parentType, parentName string, rules []TcpResponseRuleResourceModel) error {
	if len(rules) == 0 {
		return nil
	}

	log.Printf("Creating %d TCP response rules for %s %s", len(rules), parentType, parentName)

	// Sort rules by index to ensure proper ordering
	sortedRules := r.processTcpResponseRulesBlock(rules)

	// Convert all rules to payloads
	allPayloads := make([]TcpResponseRulePayload, 0, len(sortedRules))
	for i, rule := range sortedRules {
		rulePayload := r.convertToTcpResponseRulePayload(&rule, i)
		allPayloads = append(allPayloads, *rulePayload)
	}

	// Send all rules in one request (same for both v2 and v3)
	if err := r.client.CreateAllTcpResponseRulesInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		return fmt.Errorf("failed to create all TCP response rules for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Created all %d TCP response rules for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
	return nil
}

// Read reads TCP response rules
func (r *TcpResponseRuleManager) Read(ctx context.Context, parentType, parentName string) ([]TcpResponseRuleResourceModel, error) {
	log.Printf("Reading TCP response rules for %s %s", parentType, parentName)

	payloads, err := r.client.ReadTcpResponseRules(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read TCP response rules for %s %s: %w", parentType, parentName, err)
	}

	// Convert payloads to resource models
	rules := make([]TcpResponseRuleResourceModel, 0, len(payloads))
	for _, payload := range payloads {
		rule := r.convertFromTcpResponseRulePayload(payload, parentType, parentName)
		rules = append(rules, rule)
	}

	log.Printf("Read %d TCP response rules for %s %s", len(rules), parentType, parentName)
	return rules, nil
}

// Update updates TCP response rules
func (r *TcpResponseRuleManager) Update(ctx context.Context, transactionID, parentType, parentName string, rules []TcpResponseRuleResourceModel) error {
	if len(rules) == 0 {
		// If no rules, delete all existing rules
		return r.Delete(ctx, transactionID, parentType, parentName)
	}

	log.Printf("Updating %d TCP response rules for %s %s", len(rules), parentType, parentName)

	// Sort new rules by index to ensure proper ordering
	sortedRules := r.processTcpResponseRulesBlock(rules)

	// Convert new rules to payloads
	desiredPayloads := make([]TcpResponseRulePayload, 0, len(sortedRules))
	for i, rule := range sortedRules {
		rulePayload := r.convertToTcpResponseRulePayload(&rule, i)
		desiredPayloads = append(desiredPayloads, *rulePayload)
	}

	// Use delete-all-then-create-all pattern (same as http_request_rules)
	// First, delete all existing TCP response rules to avoid duplicates
	if err := r.deleteAllTcpResponseRulesInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing TCP response rules for %s %s: %w", parentType, parentName, err)
	}

	// Then create all desired rules using the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
	if err := r.client.CreateAllTcpResponseRulesInTransaction(ctx, transactionID, parentType, parentName, desiredPayloads); err != nil {
		return fmt.Errorf("failed to create new TCP response rules for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Updated %d TCP response rules for %s %s in transaction %s (delete-then-create)", len(desiredPayloads), parentType, parentName, transactionID)
	return nil
}

// generateRuleKeyFromPayload creates a unique key for a TCP response rule payload based on its content
func (r *TcpResponseRuleManager) generateRuleKeyFromPayload(payload *TcpResponseRulePayload) string {
	// Create a unique key based on the rule's content (excluding index)
	key := fmt.Sprintf("%s-%s-%s", payload.Type, payload.Action, payload.Expr)
	if payload.VarName != "" {
		key += "-" + payload.VarName
	}
	if payload.VarScope != "" {
		key += "-" + payload.VarScope
	}
	if payload.NiceValue != 0 {
		key += fmt.Sprintf("-nice%d", payload.NiceValue)
	}
	if payload.MarkValue != "" {
		key += "-mark" + payload.MarkValue
	}
	return key
}

// hasRuleChangedFromPayload checks if a rule has changed by comparing two payloads
func (r *TcpResponseRuleManager) hasRuleChangedFromPayload(existing, desired *TcpResponseRulePayload) bool {
	// Compare all fields except Index
	return existing.Type != desired.Type ||
		existing.Action != desired.Action ||
		existing.Expr != desired.Expr ||
		existing.VarName != desired.VarName ||
		existing.VarScope != desired.VarScope ||
		existing.NiceValue != desired.NiceValue ||
		existing.MarkValue != desired.MarkValue
}

// Delete deletes TCP response rules
func (r *TcpResponseRuleManager) Delete(ctx context.Context, transactionID, parentType, parentName string) error {
	log.Printf("Deleting all TCP response rules for %s %s", parentType, parentName)

	// Read existing rules to get their indices
	existingRules, err := r.client.ReadTcpResponseRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing TCP response rules for deletion: %w", err)
	}

	// Delete each rule by index in reverse order (highest index first) to avoid shifting issues
	sort.Slice(existingRules, func(i, j int) bool {
		return existingRules[i].Index > existingRules[j].Index
	})

	for _, rule := range existingRules {
		if err := r.client.DeleteTcpResponseRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName); err != nil {
			return fmt.Errorf("failed to delete TCP response rule at index %d: %w", rule.Index, err)
		}
	}

	log.Printf("Deleted %d TCP response rules for %s %s", len(existingRules), parentType, parentName)
	return nil
}

// deleteAllTcpResponseRulesInTransaction deletes all TCP response rules for a parent resource using an existing transaction ID
func (r *TcpResponseRuleManager) deleteAllTcpResponseRulesInTransaction(ctx context.Context, transactionID string, parentType string, parentName string) error {
	rules, err := r.client.ReadTcpResponseRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read TCP response rules for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Index > rules[j].Index
	})

	for _, rule := range rules {
		log.Printf("Deleting TCP response rule at index %d in transaction %s", rule.Index, transactionID)
		err := r.client.DeleteTcpResponseRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete TCP response rule at index %d: %w", rule.Index, err)
		}
	}

	return nil
}

// processTcpResponseRulesBlock processes and sorts TCP response rules
func (r *TcpResponseRuleManager) processTcpResponseRulesBlock(rules []TcpResponseRuleResourceModel) []TcpResponseRuleResourceModel {
	// Sort rules by index to ensure proper ordering
	sortedRules := make([]TcpResponseRuleResourceModel, len(rules))
	copy(sortedRules, rules)

	// Sort by index
	for i := 0; i < len(sortedRules)-1; i++ {
		for j := i + 1; j < len(sortedRules); j++ {
			if sortedRules[i].Index.ValueInt64() > sortedRules[j].Index.ValueInt64() {
				sortedRules[i], sortedRules[j] = sortedRules[j], sortedRules[i]
			}
		}
	}

	return sortedRules
}

// convertToTcpResponseRulePayload converts a resource model to a payload
func (r *TcpResponseRuleManager) convertToTcpResponseRulePayload(rule *TcpResponseRuleResourceModel, index int) *TcpResponseRulePayload {
	payload := &TcpResponseRulePayload{
		Index: int64(index),
		Type:  rule.Type.ValueString(),
	}

	// Set optional fields if they have values
	if !rule.Action.IsNull() && !rule.Action.IsUnknown() {
		payload.Action = rule.Action.ValueString()
	}
	if !rule.Cond.IsNull() && !rule.Cond.IsUnknown() {
		payload.Cond = rule.Cond.ValueString()
	}
	if !rule.CondTest.IsNull() && !rule.CondTest.IsUnknown() {
		payload.CondTest = rule.CondTest.ValueString()
	}
	if !rule.Expr.IsNull() && !rule.Expr.IsUnknown() {
		payload.Expr = rule.Expr.ValueString()
	}
	if !rule.LogLevel.IsNull() && !rule.LogLevel.IsUnknown() {
		payload.LogLevel = rule.LogLevel.ValueString()
	}
	if !rule.LuaAction.IsNull() && !rule.LuaAction.IsUnknown() {
		payload.LuaAction = rule.LuaAction.ValueString()
	}
	if !rule.LuaParams.IsNull() && !rule.LuaParams.IsUnknown() {
		payload.LuaParams = rule.LuaParams.ValueString()
	}
	if !rule.MarkValue.IsNull() && !rule.MarkValue.IsUnknown() {
		payload.MarkValue = rule.MarkValue.ValueString()
	}
	if !rule.NiceValue.IsNull() && !rule.NiceValue.IsUnknown() {
		payload.NiceValue = rule.NiceValue.ValueInt64()
	}
	if !rule.RstTtl.IsNull() && !rule.RstTtl.IsUnknown() {
		payload.RstTtl = rule.RstTtl.ValueInt64()
	}
	if !rule.ScExpr.IsNull() && !rule.ScExpr.IsUnknown() {
		payload.ScExpr = rule.ScExpr.ValueString()
	}
	if !rule.ScId.IsNull() && !rule.ScId.IsUnknown() {
		payload.ScId = rule.ScId.ValueInt64()
	}
	if !rule.ScIdx.IsNull() && !rule.ScIdx.IsUnknown() {
		payload.ScIdx = rule.ScIdx.ValueInt64()
	}
	if !rule.ScInt.IsNull() && !rule.ScInt.IsUnknown() {
		payload.ScInt = rule.ScInt.ValueInt64()
	}
	if !rule.SpoeEngine.IsNull() && !rule.SpoeEngine.IsUnknown() {
		payload.SpoeEngine = rule.SpoeEngine.ValueString()
	}
	if !rule.SpoeGroup.IsNull() && !rule.SpoeGroup.IsUnknown() {
		payload.SpoeGroup = rule.SpoeGroup.ValueString()
	}
	if !rule.Timeout.IsNull() && !rule.Timeout.IsUnknown() {
		payload.Timeout = rule.Timeout.ValueInt64()
	}
	if !rule.TosValue.IsNull() && !rule.TosValue.IsUnknown() {
		payload.TosValue = rule.TosValue.ValueString()
	}
	if !rule.VarFormat.IsNull() && !rule.VarFormat.IsUnknown() {
		payload.VarFormat = rule.VarFormat.ValueString()
	}
	if !rule.VarName.IsNull() && !rule.VarName.IsUnknown() {
		payload.VarName = rule.VarName.ValueString()
	}
	if !rule.VarScope.IsNull() && !rule.VarScope.IsUnknown() {
		payload.VarScope = rule.VarScope.ValueString()
	}
	if !rule.VarExpr.IsNull() && !rule.VarExpr.IsUnknown() {
		payload.VarExpr = rule.VarExpr.ValueString()
	}
	if !rule.BandwidthLimitLimit.IsNull() && !rule.BandwidthLimitLimit.IsUnknown() {
		payload.BandwidthLimitLimit = rule.BandwidthLimitLimit.ValueString()
	}
	if !rule.BandwidthLimitName.IsNull() && !rule.BandwidthLimitName.IsUnknown() {
		payload.BandwidthLimitName = rule.BandwidthLimitName.ValueString()
	}
	if !rule.BandwidthLimitPeriod.IsNull() && !rule.BandwidthLimitPeriod.IsUnknown() {
		payload.BandwidthLimitPeriod = rule.BandwidthLimitPeriod.ValueString()
	}
	if !rule.Timeout.IsNull() && !rule.Timeout.IsUnknown() {
		payload.Timeout = rule.Timeout.ValueInt64()
	}
	if !rule.TosValue.IsNull() && !rule.TosValue.IsUnknown() {
		payload.TosValue = rule.TosValue.ValueString()
	}
	if !rule.VarFormat.IsNull() && !rule.VarFormat.IsUnknown() {
		payload.VarFormat = rule.VarFormat.ValueString()
	}
	if !rule.VarName.IsNull() && !rule.VarName.IsUnknown() {
		payload.VarName = rule.VarName.ValueString()
	}
	if !rule.VarScope.IsNull() && !rule.VarScope.IsUnknown() {
		payload.VarScope = rule.VarScope.ValueString()
	}

	return payload
}

// convertFromTcpResponseRulePayload converts a payload to a resource model
func (r *TcpResponseRuleManager) convertFromTcpResponseRulePayload(payload TcpResponseRulePayload, parentType, parentName string) TcpResponseRuleResourceModel {
	rule := TcpResponseRuleResourceModel{
		ID:         types.StringValue(fmt.Sprintf("%s/%s/tcp_response_rule/%d", parentType, parentName, payload.Index)),
		ParentType: types.StringValue(parentType),
		ParentName: types.StringValue(parentName),
		Index:      types.Int64Value(payload.Index),
		Type:       types.StringValue(payload.Type),
	}

	// Set optional fields
	if payload.Action != "" {
		rule.Action = types.StringValue(payload.Action)
	}
	if payload.Cond != "" {
		rule.Cond = types.StringValue(payload.Cond)
	}
	if payload.CondTest != "" {
		rule.CondTest = types.StringValue(payload.CondTest)
	}
	if payload.Expr != "" {
		rule.Expr = types.StringValue(payload.Expr)
	}
	if payload.LogLevel != "" {
		rule.LogLevel = types.StringValue(payload.LogLevel)
	}
	if payload.LuaAction != "" {
		rule.LuaAction = types.StringValue(payload.LuaAction)
	}
	if payload.LuaParams != "" {
		rule.LuaParams = types.StringValue(payload.LuaParams)
	}
	if payload.MarkValue != "" {
		rule.MarkValue = types.StringValue(payload.MarkValue)
	}
	if payload.NiceValue != 0 {
		rule.NiceValue = types.Int64Value(payload.NiceValue)
	}
	if payload.RstTtl != 0 {
		rule.RstTtl = types.Int64Value(payload.RstTtl)
	}
	if payload.ScExpr != "" {
		rule.ScExpr = types.StringValue(payload.ScExpr)
	}
	if payload.ScId != 0 {
		rule.ScId = types.Int64Value(payload.ScId)
	}
	if payload.ScIdx != 0 {
		rule.ScIdx = types.Int64Value(payload.ScIdx)
	}
	if payload.ScInt != 0 {
		rule.ScInt = types.Int64Value(payload.ScInt)
	}
	if payload.SpoeEngine != "" {
		rule.SpoeEngine = types.StringValue(payload.SpoeEngine)
	}
	if payload.SpoeGroup != "" {
		rule.SpoeGroup = types.StringValue(payload.SpoeGroup)
	}
	if payload.Timeout != 0 {
		rule.Timeout = types.Int64Value(payload.Timeout)
	}
	if payload.TosValue != "" {
		rule.TosValue = types.StringValue(payload.TosValue)
	}
	if payload.VarFormat != "" {
		rule.VarFormat = types.StringValue(payload.VarFormat)
	}
	if payload.VarName != "" {
		rule.VarName = types.StringValue(payload.VarName)
	}
	if payload.VarScope != "" {
		rule.VarScope = types.StringValue(payload.VarScope)
	}
	if payload.VarExpr != "" {
		rule.VarExpr = types.StringValue(payload.VarExpr)
	}
	if payload.BandwidthLimitLimit != "" {
		rule.BandwidthLimitLimit = types.StringValue(payload.BandwidthLimitLimit)
	}
	if payload.BandwidthLimitName != "" {
		rule.BandwidthLimitName = types.StringValue(payload.BandwidthLimitName)
	}
	if payload.BandwidthLimitPeriod != "" {
		rule.BandwidthLimitPeriod = types.StringValue(payload.BandwidthLimitPeriod)
	}

	return rule
}

// Metadata returns the resource type name.
func (r *TcpResponseRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_response_rule"
}

// Schema defines the schema for the resource.
func (r *TcpResponseRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TCP Response Rule resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "TCP response rule identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_type": schema.StringAttribute{
				MarkdownDescription: "Parent type (frontend or backend)",
				Required:            true,
			},
			"parent_name": schema.StringAttribute{
				MarkdownDescription: "Parent name",
				Required:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Rule index",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Rule type",
				Required:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Rule action",
				Optional:            true,
			},
			"cond": schema.StringAttribute{
				MarkdownDescription: "Condition",
				Optional:            true,
			},
			"cond_test": schema.StringAttribute{
				MarkdownDescription: "Condition test",
				Optional:            true,
			},
			"expr": schema.StringAttribute{
				MarkdownDescription: "Expression",
				Optional:            true,
			},
			"log_level": schema.StringAttribute{
				MarkdownDescription: "Log level",
				Optional:            true,
			},
			"lua_action": schema.StringAttribute{
				MarkdownDescription: "Lua action",
				Optional:            true,
			},
			"lua_params": schema.StringAttribute{
				MarkdownDescription: "Lua parameters",
				Optional:            true,
			},
			"mark_value": schema.StringAttribute{
				MarkdownDescription: "Mark value",
				Optional:            true,
			},
			"nice_value": schema.Int64Attribute{
				MarkdownDescription: "Nice value",
				Optional:            true,
			},
			"rst_ttl": schema.Int64Attribute{
				MarkdownDescription: "RST TTL",
				Optional:            true,
			},
			"sc_expr": schema.StringAttribute{
				MarkdownDescription: "SC expression",
				Optional:            true,
			},
			"sc_id": schema.Int64Attribute{
				MarkdownDescription: "SC ID",
				Optional:            true,
			},
			"sc_idx": schema.Int64Attribute{
				MarkdownDescription: "SC index",
				Optional:            true,
			},
			"sc_int": schema.Int64Attribute{
				MarkdownDescription: "SC integer",
				Optional:            true,
			},
			"spoe_engine": schema.StringAttribute{
				MarkdownDescription: "SPOE engine",
				Optional:            true,
			},
			"spoe_group": schema.StringAttribute{
				MarkdownDescription: "SPOE group",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout",
				Optional:            true,
			},
			"tos_value": schema.StringAttribute{
				MarkdownDescription: "TOS value",
				Optional:            true,
			},
			"var_format": schema.StringAttribute{
				MarkdownDescription: "Variable format",
				Optional:            true,
			},
			"var_name": schema.StringAttribute{
				MarkdownDescription: "Variable name",
				Optional:            true,
			},
			"var_scope": schema.StringAttribute{
				MarkdownDescription: "Variable scope",
				Optional:            true,
			},
			"var_expr": schema.StringAttribute{
				MarkdownDescription: "Variable expression",
				Optional:            true,
			},
			"bandwidth_limit_limit": schema.StringAttribute{
				MarkdownDescription: "Bandwidth limit",
				Optional:            true,
			},
			"bandwidth_limit_name": schema.StringAttribute{
				MarkdownDescription: "Bandwidth limit name",
				Optional:            true,
			},
			"bandwidth_limit_period": schema.StringAttribute{
				MarkdownDescription: "Bandwidth limit period",
				Optional:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *TcpResponseRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
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
func (r *TcpResponseRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TcpResponseRuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
	}

	// Individual TCP response rule resources should only be used within haproxy_stack context
	// This resource is not registered and should not be used standalone
	resp.Diagnostics.AddError("Invalid Usage", "TCP response rule resources should only be used within haproxy_stack context. Use haproxy_stack resource instead.")
}

// Read refreshes the Terraform state with the latest data.
func (r *TcpResponseRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TcpResponseRuleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
	}

	// Read the rule
	manager := CreateTcpResponseRuleManager(r.client)
	rules, err := manager.Read(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP response rule, got error: %s", err))
	}

	// Find the specific rule by index
	var foundRule *TcpResponseRuleResourceModel
	for _, rule := range rules {
		if rule.Index.ValueInt64() == data.Index.ValueInt64() {
			foundRule = &rule
			break
		}
	}

	if foundRule == nil {
		resp.State.RemoveResource(ctx)
	}

	// Update the data with the found rule
	data = *foundRule

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *TcpResponseRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TcpResponseRuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
	}

	// Individual TCP response rule resources should only be used within haproxy_stack context
	// This resource is not registered and should not be used standalone
	resp.Diagnostics.AddError("Invalid Usage", "TCP response rule resources should only be used within haproxy_stack context. Use haproxy_stack resource instead.")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *TcpResponseRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TcpResponseRuleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
	}

	// Individual TCP response rule resources should only be used within haproxy_stack context
	// This resource is not registered and should not be used standalone
	resp.Diagnostics.AddError("Invalid Usage", "TCP response rule resources should only be used within haproxy_stack context. Use haproxy_stack resource instead.")
}
