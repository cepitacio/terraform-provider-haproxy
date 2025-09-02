package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &HttpRequestRuleResource{}
	_ resource.ResourceWithConfigure   = &HttpRequestRuleResource{}
	_ resource.ResourceWithImportState = &HttpRequestRuleResource{}
)

// HttpRequestRuleResource is the resource implementation.
type HttpRequestRuleResource struct {
	client *HAProxyClient
}

// HttpRequestRuleResourceModel maps the resource schema data.
type HttpRequestRuleResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ParentType           types.String `tfsdk:"parent_type"`
	ParentName           types.String `tfsdk:"parent_name"`
	Index                types.Int64  `tfsdk:"index"`
	Type                 types.String `tfsdk:"type"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	HdrName              types.String `tfsdk:"hdr_name"`
	HdrFormat            types.String `tfsdk:"hdr_format"`
	HdrMatch             types.String `tfsdk:"hdr_match"`
	RedirType            types.String `tfsdk:"redir_type"`
	RedirValue           types.String `tfsdk:"redir_value"`
	RedirCode            types.Int64  `tfsdk:"redir_code"`
	RedirOption          types.String `tfsdk:"redir_option"`
	PathMatch            types.String `tfsdk:"path_match"`
	PathFmt              types.String `tfsdk:"path_fmt"`
	UriMatch             types.String `tfsdk:"uri_match"`
	UriFmt               types.String `tfsdk:"uri_fmt"`
	QueryFmt             types.String `tfsdk:"query_fmt"`
	MethodFmt            types.String `tfsdk:"method_fmt"`
	VarName              types.String `tfsdk:"var_name"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarExpr              types.String `tfsdk:"var_expr"`
	VarScope             types.String `tfsdk:"var_scope"`
	CaptureID            types.Int64  `tfsdk:"capture_id"`
	CaptureLen           types.Int64  `tfsdk:"capture_len"`
	CaptureSample        types.String `tfsdk:"capture_sample"`
	LogLevel             types.String `tfsdk:"log_level"`
	Timeout              types.String `tfsdk:"timeout"`
	TimeoutType          types.String `tfsdk:"timeout_type"`
	StrictMode           types.String `tfsdk:"strict_mode"`
	Normalizer           types.String `tfsdk:"normalizer"`
	NormalizerFull       types.Bool   `tfsdk:"normalizer_full"`
	NormalizerStrict     types.Bool   `tfsdk:"normalizer_strict"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	MarkValue            types.String `tfsdk:"mark_value"`
	TosValue             types.String `tfsdk:"tos_value"`
	TrackScKey           types.String `tfsdk:"track_sc_key"`
	TrackScTable         types.String `tfsdk:"track_sc_table"`
	TrackScID            types.Int64  `tfsdk:"track_sc_id"`
	TrackScIdx           types.Int64  `tfsdk:"track_sc_idx"`
	TrackScInt           types.Int64  `tfsdk:"track_sc_int"`
	ReturnStatusCode     types.Int64  `tfsdk:"return_status_code"`
	ReturnContent        types.String `tfsdk:"return_content"`
	ReturnContentType    types.String `tfsdk:"return_content_type"`
	ReturnContentFormat  types.String `tfsdk:"return_content_format"`
	DenyStatus           types.Int64  `tfsdk:"return_deny_status"`
	WaitTime             types.Int64  `tfsdk:"wait_time"`
	WaitAtLeast          types.Int64  `tfsdk:"wait_at_least"`
	Expr                 types.String `tfsdk:"expr"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	ServiceName          types.String `tfsdk:"service_name"`
	CacheName            types.String `tfsdk:"cache_name"`
	Resolvers            types.String `tfsdk:"resolvers"`
	Protocol             types.String `tfsdk:"protocol"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	MapFile              types.String `tfsdk:"map_file"`
	MapKeyfmt            types.String `tfsdk:"map_keyfmt"`
	MapValuefmt          types.String `tfsdk:"map_valuefmt"`
	AclFile              types.String `tfsdk:"acl_file"`
	AclKeyfmt            types.String `tfsdk:"acl_keyfmt"`
	AuthRealm            types.String `tfsdk:"auth_realm"`
	HintName             types.String `tfsdk:"hint_name"`
	HintFormat           types.String `tfsdk:"hint_format"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScID                 types.Int64  `tfsdk:"sc_id"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	ScAddGpc             types.String `tfsdk:"sc_add_gpc"`
	ScIncGpc             types.String `tfsdk:"sc_inc_gpc"`
	ScIncGpc0            types.String `tfsdk:"sc_inc_gpc0"`
	ScIncGpc1            types.String `tfsdk:"sc_inc_gpc1"`
	ScSetGpt             types.String `tfsdk:"sc_set_gpt"`
	ScSetGpt0            types.String `tfsdk:"sc_set_gpt0"`
	SetPriorityClass     types.String `tfsdk:"set_priority_class"`
	SetPriorityOffset    types.String `tfsdk:"set_priority_offset"`
	SetRetries           types.String `tfsdk:"set_retries"`
	SetBcMark            types.String `tfsdk:"set_bc_mark"`
	SetBcTos             types.String `tfsdk:"set_bc_tos"`
	SetFcMark            types.String `tfsdk:"set_fc_mark"`
	SetFcTos             types.String `tfsdk:"set_fc_tos"`
	SetSrc               types.String `tfsdk:"set_src"`
	SetSrcPort           types.Int64  `tfsdk:"set_src_port"`
	SetDst               types.String `tfsdk:"set_dst"`
	SetDstPort           types.Int64  `tfsdk:"set_dst_port"`
	SetMethod            types.String `tfsdk:"set_method"`
	SetPath              types.String `tfsdk:"set_path"`
	SetPathq             types.String `tfsdk:"set_pathq"`
	SetQuery             types.String `tfsdk:"set_query"`
	SetUri               types.String `tfsdk:"set_uri"`
	SetVar               types.String `tfsdk:"set_var"`
	SetVarFmt            types.String `tfsdk:"set_var_fmt"`
	UnsetVar             types.String `tfsdk:"unset_var"`
	SilentDrop           types.Bool   `tfsdk:"silent_drop"`
	DoLog                types.Bool   `tfsdk:"do_log"`
}

// NewHttpRequestRuleResource is a helper function to simplify the resource implementation.
func NewHttpRequestRuleResource() resource.Resource {
	return &HttpRequestRuleResource{}
}

// Metadata returns the resource type name.
func (r *HttpRequestRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_request_rule"
}

// Schema defines the schema for the resource.
func (r *HttpRequestRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an HAProxy HTTP Request Rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this HTTP request rule.",
				Computed:    true,
			},
			"parent_type": schema.StringAttribute{
				Description: "The type of parent resource (frontend or backend).",
				Required:    true,
			},
			"parent_name": schema.StringAttribute{
				Description: "The name of the parent resource.",
				Required:    true,
			},
			"index": schema.Int64Attribute{
				Description: "The index of the HTTP request rule.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of HTTP request rule.",
				Required:    true,
			},
			"cond": schema.StringAttribute{
				Description: "The condition for the rule (if/unless).",
				Optional:    true,
			},
			"cond_test": schema.StringAttribute{
				Description: "The condition test for the rule.",
				Optional:    true,
			},
			"hdr_name": schema.StringAttribute{
				Description: "The header name for header-related rules.",
				Optional:    true,
			},
			"hdr_format": schema.StringAttribute{
				Description: "The header format for header-related rules.",
				Optional:    true,
			},
			"hdr_match": schema.StringAttribute{
				Description: "The header match for header-related rules.",
				Optional:    true,
			},
			"redir_type": schema.StringAttribute{
				Description: "The redirect type for redirect rules.",
				Optional:    true,
			},
			"redir_value": schema.StringAttribute{
				Description: "The redirect value for redirect rules.",
				Optional:    true,
			},
			"status_code": schema.Int64Attribute{
				Description: "The status code for response rules.",
				Optional:    true,
			},
			"status_reason": schema.StringAttribute{
				Description: "The status reason for response rules.",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *HttpRequestRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*HAProxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *HAProxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *HttpRequestRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HttpRequestRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to payload
	payload := &HttpRequestRulePayload{
		Index:        plan.Index.ValueInt64(),
		Type:         plan.Type.ValueString(),
		Cond:         plan.Cond.ValueString(),
		CondTest:     plan.CondTest.ValueString(),
		HdrName:      plan.HdrName.ValueString(),
		HdrFormat:    plan.HdrFormat.ValueString(),
		HdrMatch:     plan.HdrMatch.ValueString(),
		RedirType:    plan.RedirType.ValueString(),
		RedirValue:   plan.RedirValue.ValueString(),
		StatusCode:   plan.ReturnStatusCode.ValueInt64(),
		StatusReason: plan.ReturnContent.ValueString(),
	}

	// Create the HTTP request rule
	if err := r.client.CreateHttpRequestRule(ctx, plan.ParentType.ValueString(), plan.ParentName.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError(
			"Error creating HTTP request rule",
			fmt.Sprintf("Could not create HTTP request rule: %s", err),
		)
		return
	}

	// Set the ID
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", plan.ParentType.ValueString(), plan.ParentName.ValueString(), plan.Index.ValueInt64()))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *HttpRequestRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HttpRequestRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the HTTP request rule
	rules, err := r.client.ReadHttpRequestRules(ctx, state.ParentType.ValueString(), state.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HTTP request rules",
			fmt.Sprintf("Could not read HTTP request rules: %s", err),
		)
		return
	}

	// Find the specific rule by index
	var foundRule *HttpRequestRulePayload
	for _, rule := range rules {
		if rule.Index == state.Index.ValueInt64() {
			foundRule = &rule
			break
		}
	}

	if foundRule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state
	state.Type = types.StringValue(foundRule.Type)
	state.Cond = types.StringValue(foundRule.Cond)
	state.CondTest = types.StringValue(foundRule.CondTest)
	state.HdrName = types.StringValue(foundRule.HdrName)
	state.HdrFormat = types.StringValue(foundRule.HdrFormat)
	state.HdrMatch = types.StringValue(foundRule.HdrMatch)
	state.RedirType = types.StringValue(foundRule.RedirType)
	state.RedirValue = types.StringValue(foundRule.RedirValue)
	state.ReturnStatusCode = types.Int64Value(foundRule.StatusCode)
	state.ReturnContent = types.StringValue(foundRule.StatusReason)

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *HttpRequestRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan HttpRequestRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to payload
	payload := &HttpRequestRulePayload{
		Index:        plan.Index.ValueInt64(),
		Type:         plan.Type.ValueString(),
		Cond:         plan.Cond.ValueString(),
		CondTest:     plan.CondTest.ValueString(),
		HdrName:      plan.HdrName.ValueString(),
		HdrFormat:    plan.HdrFormat.ValueString(),
		HdrMatch:     plan.HdrMatch.ValueString(),
		RedirType:    plan.RedirType.ValueString(),
		RedirValue:   plan.RedirValue.ValueString(),
		StatusCode:   plan.ReturnStatusCode.ValueInt64(),
		StatusReason: plan.ReturnContent.ValueString(),
	}

	// Update the HTTP request rule
	if err := r.client.UpdateHttpRequestRule(ctx, plan.Index.ValueInt64(), plan.ParentType.ValueString(), plan.ParentName.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError(
			"Error updating HTTP request rule",
			fmt.Sprintf("Could not update HTTP request rule: %s", err),
		)
		return
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *HttpRequestRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HttpRequestRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the HTTP request rule
	if err := r.client.DeleteHttpRequestRule(ctx, state.Index.ValueInt64(), state.ParentType.ValueString(), state.ParentName.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting HTTP request rule",
			fmt.Sprintf("Could not delete HTTP request rule: %s", err),
		)
		return
	}
}

// ImportState configures the resource for import.
func (r *HttpRequestRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the import ID format: parent_type/parent_name/index
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format: parent_type/parent_name/index",
		)
		return
	}

	parentType := parts[0]
	parentName := parts[1]
	indexStr := parts[2]

	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid index in import ID",
			fmt.Sprintf("Index must be a number: %s", err),
		)
		return
	}

	// Set the imported values
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_type"), parentType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_name"), parentName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("index"), index)...)
}

// haproxyHttpRequestRuleModel maps the http_request_rule block schema data.
type haproxyHttpRequestRuleModel struct {
	Index      types.Int64  `tfsdk:"index"`
	Type       types.String `tfsdk:"type"`
	RedirType  types.String `tfsdk:"redir_type"`
	RedirValue types.String `tfsdk:"redir_value"`
	Cond       types.String `tfsdk:"cond"`
	CondTest   types.String `tfsdk:"cond_test"`
	HdrName    types.String `tfsdk:"hdr_name"`
	HdrFormat  types.String `tfsdk:"hdr_format"`
}

// HttpRequestRuleManager handles all HTTP request rule-related operations
type HttpRequestRuleManager struct {
	client *HAProxyClient
}

// NewHttpRequestRuleManager creates a new HttpRequestRuleManager instance
func NewHttpRequestRuleManager(client *HAProxyClient) *HttpRequestRuleManager {
	return &HttpRequestRuleManager{
		client: client,
	}
}

// CreateHttpRequestRules creates HTTP request rules for a parent resource
func (r *HttpRequestRuleManager) CreateHttpRequestRules(ctx context.Context, parentType string, parentName string, rules []haproxyHttpRequestRuleModel) error {
	if len(rules) == 0 {
		return nil
	}

	// Sort rules by index to ensure proper order
	sortedRules := r.processHttpRequestRulesBlock(rules)

	// Create rules in order
	for _, rule := range sortedRules {
		rulePayload := r.convertToHttpRequestRulePayload(&rule)

		if err := r.client.CreateHttpRequestRule(ctx, parentType, parentName, rulePayload); err != nil {
			return fmt.Errorf("failed to create HTTP request rule at index %d: %w", rule.Index.ValueInt64(), err)
		}

		log.Printf("Created HTTP request rule at index %d for %s %s", rule.Index.ValueInt64(), parentType, parentName)
	}

	return nil
}

// CreateHttpRequestRulesInTransaction creates HTTP request rules using an existing transaction ID
func (r *HttpRequestRuleManager) CreateHttpRequestRulesInTransaction(ctx context.Context, transactionID, parentType string, parentName string, rules []haproxyHttpRequestRuleModel) error {
	if len(rules) == 0 {
		return nil
	}

	// Sort rules by index to ensure proper order
	sortedRules := r.processHttpRequestRulesBlock(rules)

	// For v3, we need to send all rules at once due to API limitations
	if r.client.apiVersion == "v3" {
		// Convert all rules to payloads
		var allPayloads []HttpRequestRulePayload
		for _, rule := range sortedRules {
			rulePayload := r.convertToHttpRequestRulePayload(&rule)
			allPayloads = append(allPayloads, *rulePayload)
		}

		// Send all rules in one request
		if err := r.client.CreateAllHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
			return fmt.Errorf("failed to create all HTTP request rules for %s %s: %w", parentType, parentName, err)
		}

		log.Printf("Created all %d HTTP request rules for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)
		return nil
	}

	// v2: Create rules one by one (original logic)
	for _, rule := range sortedRules {
		rulePayload := r.convertToHttpRequestRulePayload(&rule)

		if err := r.client.CreateHttpRequestRuleInTransaction(ctx, transactionID, parentType, parentName, rulePayload); err != nil {
			return fmt.Errorf("failed to create HTTP request rule at index %d: %w", rule.Index.ValueInt64(), err)
		}

		log.Printf("Created HTTP request rule at index %d for %s %s in transaction %s", rule.Index.ValueInt64(), parentType, parentName, transactionID)
	}

	return nil
}

// ReadHttpRequestRules reads HTTP request rules for a parent resource
func (r *HttpRequestRuleManager) ReadHttpRequestRules(ctx context.Context, parentType string, parentName string) ([]HttpRequestRulePayload, error) {
	rules, err := r.client.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP request rules for %s %s: %w", parentType, parentName, err)
	}
	return rules, nil
}

// UpdateHttpRequestRules updates HTTP request rules for a parent resource
func (r *HttpRequestRuleManager) UpdateHttpRequestRules(ctx context.Context, parentType string, parentName string, newRules []haproxyHttpRequestRuleModel) error {
	if len(newRules) == 0 {
		// Delete all existing rules
		return r.deleteAllHttpRequestRules(ctx, parentType, parentName)
	}

	// Read existing rules
	existingRules, err := r.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing HTTP request rules: %w", err)
	}

	// Process updates with proper indexing
	return r.updateHttpRequestRulesWithIndexing(ctx, parentType, parentName, existingRules, newRules)
}

// DeleteHttpRequestRules deletes HTTP request rules for a parent resource
func (r *HttpRequestRuleManager) DeleteHttpRequestRules(ctx context.Context, parentType string, parentName string) error {
	return r.deleteAllHttpRequestRules(ctx, parentType, parentName)
}

// processHttpRequestRulesBlock processes and sorts HTTP request rules by index
func (r *HttpRequestRuleManager) processHttpRequestRulesBlock(rules []haproxyHttpRequestRuleModel) []haproxyHttpRequestRuleModel {
	if len(rules) == 0 {
		return rules
	}

	// Sort by index to ensure proper order
	sortedRules := make([]haproxyHttpRequestRuleModel, len(rules))
	copy(sortedRules, rules)

	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Index.ValueInt64() < sortedRules[j].Index.ValueInt64()
	})

	return sortedRules
}

// convertToHttpRequestRulePayload converts the Terraform model to HAProxy API payload
func (r *HttpRequestRuleManager) convertToHttpRequestRulePayload(rule *haproxyHttpRequestRuleModel) *HttpRequestRulePayload {
	payload := &HttpRequestRulePayload{
		Index: rule.Index.ValueInt64(),
		Type:  rule.Type.ValueString(),
	}

	// Set optional fields only if they have values
	if !rule.Cond.IsNull() && !rule.Cond.IsUnknown() {
		payload.Cond = rule.Cond.ValueString()
	}
	if !rule.CondTest.IsNull() && !rule.CondTest.IsUnknown() {
		payload.CondTest = rule.CondTest.ValueString()
	}
	if !rule.HdrName.IsNull() && !rule.HdrName.IsUnknown() {
		payload.HdrName = rule.HdrName.ValueString()
	}
	if !rule.HdrFormat.IsNull() && !rule.HdrFormat.IsUnknown() {
		payload.HdrFormat = rule.HdrFormat.ValueString()
	}
	if !rule.RedirType.IsNull() && !rule.RedirType.IsUnknown() {
		payload.RedirType = rule.RedirType.ValueString()
	}
	if !rule.RedirValue.IsNull() && !rule.RedirValue.IsUnknown() {
		payload.RedirValue = rule.RedirValue.ValueString()
	}

	return payload
}

// updateHttpRequestRulesWithIndexing handles the complex logic of updating HTTP request rules while maintaining order
func (r *HttpRequestRuleManager) updateHttpRequestRulesWithIndexing(ctx context.Context, parentType string, parentName string, existingRules []HttpRequestRulePayload, newRules []haproxyHttpRequestRuleModel) error {
	// Process new rules with proper indexing
	sortedNewRules := r.processHttpRequestRulesBlock(newRules)

	// Create maps for efficient lookup
	existingRuleMap := make(map[int64]*HttpRequestRulePayload)
	for i := range existingRules {
		existingRuleMap[existingRules[i].Index] = &existingRules[i]
	}

	// Track which rules we've processed to avoid duplicates
	processedRules := make(map[int64]bool)

	// Track rules that need to be recreated due to index changes
	var rulesToRecreate []haproxyHttpRequestRuleModel

	// First pass: identify rules that need index changes and mark them for recreation
	for _, newRule := range sortedNewRules {
		newRuleIndex := newRule.Index.ValueInt64()

		// Check if this rule exists by index
		if existingRule, exists := existingRuleMap[newRuleIndex]; exists {
			// Index exists, check if content has changed
			if r.hasRuleChanged(existingRule, &newRule) {
				// Content has changed, mark for recreation
				log.Printf("HTTP request rule at index %d has changed, will recreate", newRuleIndex)
				rulesToRecreate = append(rulesToRecreate, newRule)
			} else {
				// Rule is identical, no changes needed
				log.Printf("HTTP request rule at index %d is unchanged", newRuleIndex)
			}
			// Mark this rule as processed
			processedRules[newRuleIndex] = true
		} else {
			// This is a new rule, mark for creation
			log.Printf("HTTP request rule at index %d is new, will create", newRuleIndex)
		}
	}

	// Second pass: delete all rules that need to be recreated (due to content changes)
	// Delete in reverse order (highest index first) to avoid shifting issues
	for _, newRule := range rulesToRecreate {
		newRuleIndex := newRule.Index.ValueInt64()
		if existingRule, exists := existingRuleMap[newRuleIndex]; exists {
			log.Printf("Deleting HTTP request rule at index %d for recreation", newRuleIndex)
			err := r.client.DeleteHttpRequestRule(ctx, existingRule.Index, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete HTTP request rule at index %d: %w", newRuleIndex, err)
			}
		}
	}

	// Third pass: create all rules that need to be recreated at their positions
	for _, newRule := range rulesToRecreate {
		newRuleIndex := newRule.Index.ValueInt64()

		log.Printf("Creating HTTP request rule at index %d", newRuleIndex)
		rulePayload := r.convertToHttpRequestRulePayload(&newRule)

		err := r.client.CreateHttpRequestRule(ctx, parentType, parentName, rulePayload)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request rule at index %d: %w", newRuleIndex, err)
		}
	}

	// Delete rules that are no longer needed (not in the new configuration)
	// Delete in reverse order (highest index first) to avoid shifting issues
	var rulesToDelete []HttpRequestRulePayload
	for _, existingRule := range existingRules {
		if !processedRules[existingRule.Index] {
			rulesToDelete = append(rulesToDelete, existingRule)
		}
	}

	// Sort by index in descending order (highest first)
	sort.Slice(rulesToDelete, func(i, j int) bool {
		return rulesToDelete[i].Index > rulesToDelete[j].Index
	})

	// Delete rules in reverse order
	for _, ruleToDelete := range rulesToDelete {
		log.Printf("Deleting HTTP request rule at index %d (no longer needed)", ruleToDelete.Index)
		err := r.client.DeleteHttpRequestRule(ctx, ruleToDelete.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete HTTP request rule: %w", err)
		}
	}

	// Create new rules that don't exist yet
	for _, newRule := range sortedNewRules {
		newRuleIndex := newRule.Index.ValueInt64()
		if !processedRules[newRuleIndex] {
			// This is a new rule, create it with the user-specified index
			log.Printf("Creating new HTTP request rule at index %d", newRuleIndex)
			rulePayload := r.convertToHttpRequestRulePayload(&newRule)

			err := r.client.CreateHttpRequestRule(ctx, parentType, parentName, rulePayload)
			if err != nil {
				return fmt.Errorf("failed to create HTTP request rule: %w", err)
			}
		}
	}

	return nil
}

// UpdateHttpRequestRulesInTransaction updates HTTP request rules using an existing transaction ID
func (r *HttpRequestRuleManager) UpdateHttpRequestRulesInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, rules []haproxyHttpRequestRuleModel) error {
	// For now, we'll use the existing UpdateHttpRequestRules logic but with transaction support
	// This is a simplified version that creates a new transaction for rule updates
	// In a more sophisticated implementation, we could reuse the existing transaction

	// Delete existing rules first
	if err := r.deleteAllHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing HTTP request rules: %w", err)
	}

	// Create new rules with the transaction
	if err := r.CreateHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName, rules); err != nil {
		return fmt.Errorf("failed to create new HTTP request rules: %w", err)
	}

	return nil
}

// hasRuleChanged checks if an existing rule has changed compared to a new rule
func (r *HttpRequestRuleManager) hasRuleChanged(existing *HttpRequestRulePayload, new *haproxyHttpRequestRuleModel) bool {
	// Compare the most important fields
	if existing.Type != new.Type.ValueString() {
		return true
	}
	if existing.Cond != new.Cond.ValueString() {
		return true
	}
	if existing.CondTest != new.CondTest.ValueString() {
		return true
	}
	if existing.HdrName != new.HdrName.ValueString() {
		return true
	}
	if existing.HdrFormat != new.HdrFormat.ValueString() {
		return true
	}
	if existing.RedirType != new.RedirType.ValueString() {
		return true
	}
	if existing.RedirValue != new.RedirValue.ValueString() {
		return true
	}
	// Add more field comparisons as needed
	return false
}

// deleteAllHttpRequestRulesInTransaction deletes all HTTP request rules for a parent resource using an existing transaction ID
func (r *HttpRequestRuleManager) deleteAllHttpRequestRulesInTransaction(ctx context.Context, transactionID string, parentType string, parentName string) error {
	rules, err := r.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read HTTP request rules for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Index > rules[j].Index
	})

	for _, rule := range rules {
		log.Printf("Deleting HTTP request rule at index %d in transaction %s", rule.Index, transactionID)
		err := r.client.DeleteHttpRequestRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete HTTP request rule at index %d: %w", rule.Index, err)
		}
	}

	return nil
}

// deleteAllHttpRequestRules deletes all HTTP request rules for a parent resource
func (r *HttpRequestRuleManager) deleteAllHttpRequestRules(ctx context.Context, parentType string, parentName string) error {
	rules, err := r.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read HTTP request rules for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Index > rules[j].Index
	})

	for _, rule := range rules {
		log.Printf("Deleting HTTP request rule at index %d", rule.Index)
		err := r.client.DeleteHttpRequestRule(ctx, rule.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete HTTP request rule at index %d: %w", rule.Index, err)
		}
	}

	return nil
}

// DeleteHttpRequestRulesInTransaction deletes all HTTP request rules for a parent resource using an existing transaction ID
func (r *HttpRequestRuleManager) DeleteHttpRequestRulesInTransaction(ctx context.Context, transactionID string, parentType string, parentName string) error {
	rules, err := r.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read HTTP request rules for deletion: %w", err)
	}

	// Delete in reverse order (highest index first) to avoid shifting issues
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Index > rules[j].Index
	})

	for _, rule := range rules {
		log.Printf("Deleting HTTP request rule at index %d in transaction %s", rule.Index, transactionID)
		err := r.client.DeleteHttpRequestRuleInTransaction(ctx, transactionID, rule.Index, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete HTTP request rule at index %d: %w", rule.Index, err)
		}
	}

	return nil
}

// formatHttpRequestRuleOrder formats the order of HTTP request rules for logging
func (r *HttpRequestRuleManager) formatHttpRequestRuleOrder(rules []haproxyHttpRequestRuleModel) string {
	if len(rules) == 0 {
		return "[]"
	}

	var order []string
	for _, rule := range rules {
		order = append(order, fmt.Sprintf("%d", rule.Index.ValueInt64()))
	}
	return fmt.Sprintf("[%s]", strings.Join(order, ", "))
}
