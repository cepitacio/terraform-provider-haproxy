package haproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TcpRequestRuleResource defines the resource implementation.
type TcpRequestRuleResource struct {
	client *HAProxyClient
}

// TcpRequestRuleResourceModel describes the resource data model.
type TcpRequestRuleResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ParentType           types.String `tfsdk:"parent_type"`
	ParentName           types.String `tfsdk:"parent_name"`
	Index                types.Int64  `tfsdk:"index"`
	Type                 types.String `tfsdk:"type"`
	Action               types.String `tfsdk:"action"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	Expr                 types.String `tfsdk:"expr"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	LogLevel             types.String `tfsdk:"log_level"`
	MarkValue            types.String `tfsdk:"mark_value"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	TosValue             types.String `tfsdk:"tos_value"`
	CaptureLen           types.Int64  `tfsdk:"capture_len"`
	CaptureSample        types.String `tfsdk:"capture_sample"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	ResolveProtocol      types.String `tfsdk:"resolve_protocol"`
	ResolveResolvers     types.String `tfsdk:"resolve_resolvers"`
	ResolveVar           types.String `tfsdk:"resolve_var"`
	RstTtl               types.Int64  `tfsdk:"rst_ttl"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScIncId              types.String `tfsdk:"sc_inc_id"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	ServerName           types.String `tfsdk:"server_name"`
	ServiceName          types.String `tfsdk:"service_name"`
	SpoeEngineName       types.String `tfsdk:"spoe_engine_name"`
	SpoeGroupName        types.String `tfsdk:"spoe_group_name"`
	SwitchModeProto      types.String `tfsdk:"switch_mode_proto"`
	TrackKey             types.String `tfsdk:"track_key"`
	TrackStickCounter    types.Int64  `tfsdk:"track_stick_counter"`
	TrackTable           types.String `tfsdk:"track_table"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	VarExpr              types.String `tfsdk:"var_expr"`
	GptValue             types.String `tfsdk:"gpt_value"`
}

// TcpRequestRuleManager manages TCP request rules
type TcpRequestRuleManager struct {
	client *HAProxyClient
}

// NewTcpRequestRuleManager creates a new TCP request rule manager
func NewTcpRequestRuleManager(client *HAProxyClient) *TcpRequestRuleManager {
	return &TcpRequestRuleManager{
		client: client,
	}
}

// Create creates TCP request rules
func (r *TcpRequestRuleManager) Create(ctx context.Context, transactionID, parentType, parentName string, rules []TcpRequestRuleResourceModel) error {
	if len(rules) == 0 {
		return nil
	}

	log.Printf("Creating %d TCP request rules for %s %s", len(rules), parentType, parentName)

	// Sort rules by index to ensure proper ordering
	sortedRules := r.processTcpRequestRulesBlock(rules)

	// Convert all rules to payloads
	allPayloads := make([]TcpRequestRulePayload, 0, len(sortedRules))
	for i, rule := range sortedRules {
		rulePayload := r.convertToTcpRequestRulePayload(&rule, i)
		allPayloads = append(allPayloads, *rulePayload)
	}

	// Send all rules in one request (same for both v2 and v3)
	if err := r.client.CreateAllTcpRequestRulesInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		return fmt.Errorf("failed to create all TCP request rules for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Created all %d TCP request rules for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
	return nil
}

// Read reads TCP request rules
func (r *TcpRequestRuleManager) Read(ctx context.Context, parentType, parentName string) ([]TcpRequestRuleResourceModel, error) {
	log.Printf("Reading TCP request rules for %s %s", parentType, parentName)

	payloads, err := r.client.ReadTcpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read TCP request rules for %s %s: %w", parentType, parentName, err)
	}

	// Convert payloads to resource models
	rules := make([]TcpRequestRuleResourceModel, 0, len(payloads))
	for _, payload := range payloads {
		rule := r.convertFromTcpRequestRulePayload(payload, parentType, parentName)
		rules = append(rules, rule)
	}

	log.Printf("Read %d TCP request rules for %s %s", len(rules), parentType, parentName)
	return rules, nil
}

// Update updates TCP request rules
func (r *TcpRequestRuleManager) Update(ctx context.Context, transactionID, parentType, parentName string, rules []TcpRequestRuleResourceModel) error {
	if len(rules) == 0 {
		// If no rules, delete all existing rules
		return r.Delete(ctx, transactionID, parentType, parentName)
	}

	log.Printf("Updating %d TCP request rules for %s %s", len(rules), parentType, parentName)

	// Get existing rules from the API
	existingRules, err := r.client.ReadTcpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing TCP request rules: %w", err)
	}

	// Sort new rules by index to ensure proper ordering
	sortedRules := r.processTcpRequestRulesBlock(rules)

	// Convert new rules to payloads
	desiredPayloads := make([]TcpRequestRulePayload, 0, len(sortedRules))
	for i, rule := range sortedRules {
		rulePayload := r.convertToTcpRequestRulePayload(&rule, i)
		desiredPayloads = append(desiredPayloads, *rulePayload)
	}

	// Create maps for comparison
	existingMap := make(map[string]TcpRequestRulePayload)
	desiredMap := make(map[string]TcpRequestRulePayload)

	// Populate existing rules map
	for _, rule := range existingRules {
		key := r.generateRuleKeyFromPayload(&rule)
		existingMap[key] = rule
	}

	// Populate desired rules map
	for _, rule := range desiredPayloads {
		key := r.generateRuleKeyFromPayload(&rule)
		desiredMap[key] = rule
	}

	// Find rules to delete, update, and create
	var rulesToDelete, rulesToUpdate, rulesToCreate []TcpRequestRulePayload

	// Rules to delete: exist in state but not in plan
	for key, existingRule := range existingMap {
		if _, exists := desiredMap[key]; !exists {
			rulesToDelete = append(rulesToDelete, existingRule)
		}
	}

	// Rules to update: exist in both but have changed
	for key, desiredRule := range desiredMap {
		if existingRule, exists := existingMap[key]; exists {
			if r.hasRuleChangedFromPayload(&existingRule, &desiredRule) {
				log.Printf("DEBUG: TCP request rule '%s' content changed, will update", key)
				rulesToUpdate = append(rulesToUpdate, desiredRule)
			} else if existingRule.Index != desiredRule.Index {
				log.Printf("DEBUG: TCP request rule '%s' position changed from %d to %d, will reorder", key, existingRule.Index, desiredRule.Index)
				rulesToUpdate = append(rulesToUpdate, desiredRule)
			}
		}
	}

	// Rules to create: exist in plan but not in state
	for key, desiredRule := range desiredMap {
		if _, exists := existingMap[key]; !exists {
			rulesToCreate = append(rulesToCreate, desiredRule)
		}
	}

	// Delete rules that are no longer needed
	for _, rule := range rulesToDelete {
		log.Printf("Deleting TCP request rule '%s' at index %d", r.generateRuleKeyFromPayload(&rule), rule.Index)
		if err := r.client.DeleteTcpRequestRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName); err != nil {
			return fmt.Errorf("failed to delete TCP request rule: %w", err)
		}
	}

	// Update rules that have changed
	for _, rule := range rulesToUpdate {
		log.Printf("Updating TCP request rule '%s' at index %d", r.generateRuleKeyFromPayload(&rule), rule.Index)
		if err := r.client.UpdateTcpRequestRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName, &rule); err != nil {
			return fmt.Errorf("failed to update TCP request rule: %w", err)
		}
	}

	// Create new rules
	for _, rule := range rulesToCreate {
		log.Printf("Creating TCP request rule '%s' at index %d", r.generateRuleKeyFromPayload(&rule), rule.Index)
		if err := r.client.CreateTcpRequestRuleInTransaction(ctx, transactionID, parentType, parentName, &rule); err != nil {
			return fmt.Errorf("failed to create TCP request rule: %w", err)
		}
	}

	log.Printf("Updated %d TCP request rules for %s %s (deleted: %d, updated: %d, created: %d)",
		len(desiredPayloads), parentType, parentName, len(rulesToDelete), len(rulesToUpdate), len(rulesToCreate))
	return nil
}

// generateRuleKeyFromPayload creates a unique key for a TCP request rule payload based on its content
func (r *TcpRequestRuleManager) generateRuleKeyFromPayload(payload *TcpRequestRulePayload) string {
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
func (r *TcpRequestRuleManager) hasRuleChangedFromPayload(existing, desired *TcpRequestRulePayload) bool {
	// Compare all fields except Index
	return existing.Type != desired.Type ||
		existing.Action != desired.Action ||
		existing.Expr != desired.Expr ||
		existing.VarName != desired.VarName ||
		existing.VarScope != desired.VarScope ||
		existing.NiceValue != desired.NiceValue ||
		existing.MarkValue != desired.MarkValue
}

// Delete deletes TCP request rules
func (r *TcpRequestRuleManager) Delete(ctx context.Context, transactionID, parentType, parentName string) error {
	log.Printf("Deleting all TCP request rules for %s %s", parentType, parentName)

	// Read existing rules to get their indices
	existingRules, err := r.client.ReadTcpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing TCP request rules for deletion: %w", err)
	}

	// Delete each rule by index in reverse order (highest index first) to avoid shifting issues
	sort.Slice(existingRules, func(i, j int) bool {
		return existingRules[i].Index > existingRules[j].Index
	})

	for _, rule := range existingRules {
		if err := r.client.DeleteTcpRequestRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName); err != nil {
			return fmt.Errorf("failed to delete TCP request rule at index %d: %w", rule.Index, err)
		}
	}

	log.Printf("Deleted %d TCP request rules for %s %s", len(existingRules), parentType, parentName)
	return nil
}

// processTcpRequestRulesBlock processes and sorts TCP request rules
func (r *TcpRequestRuleManager) processTcpRequestRulesBlock(rules []TcpRequestRuleResourceModel) []TcpRequestRuleResourceModel {
	// Sort rules by index to ensure proper ordering
	sortedRules := make([]TcpRequestRuleResourceModel, len(rules))
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

// convertToTcpRequestRulePayload converts a resource model to a payload
func (r *TcpRequestRuleManager) convertToTcpRequestRulePayload(rule *TcpRequestRuleResourceModel, index int) *TcpRequestRulePayload {
	payload := &TcpRequestRulePayload{
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
	if !rule.Timeout.IsNull() && !rule.Timeout.IsUnknown() {
		payload.Timeout = rule.Timeout.ValueInt64()
	}
	if !rule.LuaAction.IsNull() && !rule.LuaAction.IsUnknown() {
		payload.LuaAction = rule.LuaAction.ValueString()
	}
	if !rule.LuaParams.IsNull() && !rule.LuaParams.IsUnknown() {
		payload.LuaParams = rule.LuaParams.ValueString()
	}
	if !rule.LogLevel.IsNull() && !rule.LogLevel.IsUnknown() {
		payload.LogLevel = rule.LogLevel.ValueString()
	}
	if !rule.MarkValue.IsNull() && !rule.MarkValue.IsUnknown() {
		payload.MarkValue = rule.MarkValue.ValueString()
	}
	if !rule.NiceValue.IsNull() && !rule.NiceValue.IsUnknown() {
		payload.NiceValue = rule.NiceValue.ValueInt64()
	}
	if !rule.TosValue.IsNull() && !rule.TosValue.IsUnknown() {
		payload.TosValue = rule.TosValue.ValueString()
	}
	if !rule.CaptureLen.IsNull() && !rule.CaptureLen.IsUnknown() {
		payload.CaptureLen = rule.CaptureLen.ValueInt64()
	}
	if !rule.CaptureSample.IsNull() && !rule.CaptureSample.IsUnknown() {
		payload.CaptureSample = rule.CaptureSample.ValueString()
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
	if !rule.ResolveProtocol.IsNull() && !rule.ResolveProtocol.IsUnknown() {
		payload.ResolveProtocol = rule.ResolveProtocol.ValueString()
	}
	if !rule.ResolveResolvers.IsNull() && !rule.ResolveResolvers.IsUnknown() {
		payload.ResolveResolvers = rule.ResolveResolvers.ValueString()
	}
	if !rule.ResolveVar.IsNull() && !rule.ResolveVar.IsUnknown() {
		payload.ResolveVar = rule.ResolveVar.ValueString()
	}
	if !rule.RstTtl.IsNull() && !rule.RstTtl.IsUnknown() {
		payload.RstTtl = rule.RstTtl.ValueInt64()
	}
	if !rule.ScIdx.IsNull() && !rule.ScIdx.IsUnknown() {
		payload.ScIdx = rule.ScIdx.ValueInt64()
	}
	if !rule.ScIncId.IsNull() && !rule.ScIncId.IsUnknown() {
		payload.ScIncId = rule.ScIncId.ValueString()
	}
	if !rule.ScInt.IsNull() && !rule.ScInt.IsUnknown() {
		payload.ScInt = rule.ScInt.ValueInt64()
	}
	if !rule.ServerName.IsNull() && !rule.ServerName.IsUnknown() {
		payload.ServerName = rule.ServerName.ValueString()
	}
	if !rule.ServiceName.IsNull() && !rule.ServiceName.IsUnknown() {
		payload.ServiceName = rule.ServiceName.ValueString()
	}
	if !rule.SpoeEngineName.IsNull() && !rule.SpoeEngineName.IsUnknown() {
		payload.SpoeEngineName = rule.SpoeEngineName.ValueString()
	}
	if !rule.SpoeGroupName.IsNull() && !rule.SpoeGroupName.IsUnknown() {
		payload.SpoeGroupName = rule.SpoeGroupName.ValueString()
	}
	if !rule.SwitchModeProto.IsNull() && !rule.SwitchModeProto.IsUnknown() {
		payload.SwitchModeProto = rule.SwitchModeProto.ValueString()
	}
	if !rule.TrackKey.IsNull() && !rule.TrackKey.IsUnknown() {
		payload.TrackKey = rule.TrackKey.ValueString()
	}
	if !rule.TrackStickCounter.IsNull() && !rule.TrackStickCounter.IsUnknown() {
		payload.TrackStickCounter = rule.TrackStickCounter.ValueInt64()
	}
	if !rule.TrackTable.IsNull() && !rule.TrackTable.IsUnknown() {
		payload.TrackTable = rule.TrackTable.ValueString()
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
	if !rule.GptValue.IsNull() && !rule.GptValue.IsUnknown() {
		payload.GptValue = rule.GptValue.ValueString()
	}

	return payload
}

// convertFromTcpRequestRulePayload converts a payload to a resource model
func (r *TcpRequestRuleManager) convertFromTcpRequestRulePayload(payload TcpRequestRulePayload, parentType, parentName string) TcpRequestRuleResourceModel {
	rule := TcpRequestRuleResourceModel{
		ID:         types.StringValue(fmt.Sprintf("%s/%s/tcp_request_rule/%d", parentType, parentName, payload.Index)),
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
	if payload.Timeout != 0 {
		rule.Timeout = types.Int64Value(payload.Timeout)
	}
	if payload.LuaAction != "" {
		rule.LuaAction = types.StringValue(payload.LuaAction)
	}
	if payload.LuaParams != "" {
		rule.LuaParams = types.StringValue(payload.LuaParams)
	}
	if payload.LogLevel != "" {
		rule.LogLevel = types.StringValue(payload.LogLevel)
	}
	if payload.MarkValue != "" {
		rule.MarkValue = types.StringValue(payload.MarkValue)
	}
	if payload.NiceValue != 0 {
		rule.NiceValue = types.Int64Value(payload.NiceValue)
	}
	if payload.TosValue != "" {
		rule.TosValue = types.StringValue(payload.TosValue)
	}
	if payload.CaptureLen != 0 {
		rule.CaptureLen = types.Int64Value(payload.CaptureLen)
	}
	if payload.CaptureSample != "" {
		rule.CaptureSample = types.StringValue(payload.CaptureSample)
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
	if payload.ResolveProtocol != "" {
		rule.ResolveProtocol = types.StringValue(payload.ResolveProtocol)
	}
	if payload.ResolveResolvers != "" {
		rule.ResolveResolvers = types.StringValue(payload.ResolveResolvers)
	}
	if payload.ResolveVar != "" {
		rule.ResolveVar = types.StringValue(payload.ResolveVar)
	}
	if payload.RstTtl != 0 {
		rule.RstTtl = types.Int64Value(payload.RstTtl)
	}
	if payload.ScIdx != 0 {
		rule.ScIdx = types.Int64Value(payload.ScIdx)
	}
	if payload.ScIncId != "" {
		rule.ScIncId = types.StringValue(payload.ScIncId)
	}
	if payload.ScInt != 0 {
		rule.ScInt = types.Int64Value(payload.ScInt)
	}
	if payload.ServerName != "" {
		rule.ServerName = types.StringValue(payload.ServerName)
	}
	if payload.ServiceName != "" {
		rule.ServiceName = types.StringValue(payload.ServiceName)
	}
	if payload.SpoeEngineName != "" {
		rule.SpoeEngineName = types.StringValue(payload.SpoeEngineName)
	}
	if payload.SpoeGroupName != "" {
		rule.SpoeGroupName = types.StringValue(payload.SpoeGroupName)
	}
	if payload.SwitchModeProto != "" {
		rule.SwitchModeProto = types.StringValue(payload.SwitchModeProto)
	}
	if payload.TrackKey != "" {
		rule.TrackKey = types.StringValue(payload.TrackKey)
	}
	if payload.TrackStickCounter != 0 {
		rule.TrackStickCounter = types.Int64Value(payload.TrackStickCounter)
	}
	if payload.TrackTable != "" {
		rule.TrackTable = types.StringValue(payload.TrackTable)
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
	if payload.GptValue != "" {
		rule.GptValue = types.StringValue(payload.GptValue)
	}

	return rule
}

// Metadata returns the resource type name.
func (r *TcpRequestRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tcp_request_rule"
}

// Schema defines the schema for the resource.
func (r *TcpRequestRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "TCP Request Rule resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "TCP request rule identifier",
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
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout",
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
			"log_level": schema.StringAttribute{
				MarkdownDescription: "Log level",
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
			"tos_value": schema.StringAttribute{
				MarkdownDescription: "TOS value",
				Optional:            true,
			},
			"capture_len": schema.Int64Attribute{
				MarkdownDescription: "Capture length",
				Optional:            true,
			},
			"capture_sample": schema.StringAttribute{
				MarkdownDescription: "Capture sample",
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
			"resolve_protocol": schema.StringAttribute{
				MarkdownDescription: "Resolve protocol",
				Optional:            true,
			},
			"resolve_resolvers": schema.StringAttribute{
				MarkdownDescription: "Resolve resolvers",
				Optional:            true,
			},
			"resolve_var": schema.StringAttribute{
				MarkdownDescription: "Resolve variable",
				Optional:            true,
			},
			"rst_ttl": schema.Int64Attribute{
				MarkdownDescription: "RST TTL",
				Optional:            true,
			},
			"sc_idx": schema.StringAttribute{
				MarkdownDescription: "SC index",
				Optional:            true,
			},
			"sc_inc_id": schema.StringAttribute{
				MarkdownDescription: "SC increment ID",
				Optional:            true,
			},
			"sc_int": schema.Int64Attribute{
				MarkdownDescription: "SC integer",
				Optional:            true,
			},
			"server_name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Optional:            true,
			},
			"service_name": schema.StringAttribute{
				MarkdownDescription: "Service name",
				Optional:            true,
			},
			"spoe_engine_name": schema.StringAttribute{
				MarkdownDescription: "SPOE engine name",
				Optional:            true,
			},
			"spoe_group_name": schema.StringAttribute{
				MarkdownDescription: "SPOE group name",
				Optional:            true,
			},
			"switch_mode_proto": schema.StringAttribute{
				MarkdownDescription: "Switch mode protocol",
				Optional:            true,
			},
			"track_key": schema.StringAttribute{
				MarkdownDescription: "Track key",
				Optional:            true,
			},
			"track_stick_counter": schema.Int64Attribute{
				MarkdownDescription: "Track stick counter",
				Optional:            true,
			},
			"track_table": schema.StringAttribute{
				MarkdownDescription: "Track table",
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
			"gpt_value": schema.StringAttribute{
				MarkdownDescription: "GPT value",
				Optional:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *TcpRequestRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *TcpRequestRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TcpRequestRuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the rule using transaction
	manager := NewTcpRequestRuleManager(r.client)
	_, err := r.client.Transaction(func(transactionID string) (*http.Response, error) {
		if err := manager.Create(ctx, transactionID, data.ParentType.ValueString(), data.ParentName.ValueString(), []TcpRequestRuleResourceModel{data}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create TCP request rule, got error: %s", err))
		return
	}

	// Set ID
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/tcp_request_rule/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a TCP request rule resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *TcpRequestRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TcpRequestRuleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the rule
	manager := NewTcpRequestRuleManager(r.client)
	rules, err := manager.Read(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read TCP request rule, got error: %s", err))
		return
	}

	// Find the specific rule by index
	var foundRule *TcpRequestRuleResourceModel
	for _, rule := range rules {
		if rule.Index.ValueInt64() == data.Index.ValueInt64() {
			foundRule = &rule
			break
		}
	}

	if foundRule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the data with the found rule
	data = *foundRule

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *TcpRequestRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TcpRequestRuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule using transaction
	manager := NewTcpRequestRuleManager(r.client)
	_, err := r.client.Transaction(func(transactionID string) (*http.Response, error) {
		if err := manager.Update(ctx, transactionID, data.ParentType.ValueString(), data.ParentName.ValueString(), []TcpRequestRuleResourceModel{data}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update TCP request rule, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *TcpRequestRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TcpRequestRuleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the rule using transaction
	_, err := r.client.Transaction(func(transactionID string) (*http.Response, error) {
		if err := r.client.DeleteTcpRequestRuleInTransaction(ctx, transactionID, data.Index.ValueInt64(), data.ParentType.ValueString(), data.ParentName.ValueString()); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete TCP request rule, got error: %s", err))
		return
	}
}
