package haproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	// Create the HTTP request rule using single transaction approach
	transactionID, err := r.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return
	}
	defer r.client.RollbackTransaction(transactionID)

	if err := r.client.CreateHttpRequestRuleInTransaction(ctx, transactionID, plan.ParentType.ValueString(), plan.ParentName.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError(
			"Error creating HTTP request rule",
			fmt.Sprintf("Could not create HTTP request rule: %s", err),
		)
		return
	}

	if err := r.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
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
	for i, rule := range rules {
		if int64(i) == state.Index.ValueInt64() {
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

	// Update the HTTP request rule using single transaction approach
	transactionID, err := r.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return
	}
	defer r.client.RollbackTransaction(transactionID)

	// For HTTP request rules, we need to delete and recreate since there's no direct update
	if err := r.client.DeleteHttpRequestRuleInTransaction(ctx, transactionID, plan.Index.ValueInt64(), plan.ParentType.ValueString(), plan.ParentName.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting HTTP request rule for update",
			fmt.Sprintf("Could not delete HTTP request rule: %s", err),
		)
		return
	}

	if err := r.client.CreateHttpRequestRuleInTransaction(ctx, transactionID, plan.ParentType.ValueString(), plan.ParentName.ValueString(), payload); err != nil {
		resp.Diagnostics.AddError(
			"Error creating HTTP request rule for update",
			fmt.Sprintf("Could not create HTTP request rule: %s", err),
		)
		return
	}

	if err := r.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
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

	// Delete the HTTP request rule using single transaction approach
	transactionID, err := r.client.BeginTransaction()
	if err != nil {
		resp.Diagnostics.AddError("Error beginning transaction", err.Error())
		return
	}
	defer r.client.RollbackTransaction(transactionID)

	if err := r.client.DeleteHttpRequestRuleInTransaction(ctx, transactionID, state.Index.ValueInt64(), state.ParentType.ValueString(), state.ParentName.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting HTTP request rule",
			fmt.Sprintf("Could not delete HTTP request rule: %s", err),
		)
		return
	}

	if err := r.client.CommitTransaction(transactionID); err != nil {
		resp.Diagnostics.AddError("Error committing transaction", err.Error())
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
	Index                types.Int64  `tfsdk:"index"` // For backward compatibility with existing state
	Type                 types.String `tfsdk:"type"`
	RedirType            types.String `tfsdk:"redir_type"`
	RedirValue           types.String `tfsdk:"redir_value"`
	Cond                 types.String `tfsdk:"cond"`
	CondTest             types.String `tfsdk:"cond_test"`
	HdrName              types.String `tfsdk:"hdr_name"`
	HdrFormat            types.String `tfsdk:"hdr_format"`
	HdrMatch             types.String `tfsdk:"hdr_match"`
	HdrMethod            types.String `tfsdk:"hdr_method"`
	RedirCode            types.Int64  `tfsdk:"redir_code"`
	RedirOption          types.String `tfsdk:"redir_option"`
	BandwidthLimitName   types.String `tfsdk:"bandwidth_limit_name"`
	BandwidthLimitLimit  types.String `tfsdk:"bandwidth_limit_limit"`
	BandwidthLimitPeriod types.String `tfsdk:"bandwidth_limit_period"`
	AclFile              types.String `tfsdk:"acl_file"`
	AclKeyfmt            types.String `tfsdk:"acl_keyfmt"`
	AuthRealm            types.String `tfsdk:"auth_realm"`
	CacheName            types.String `tfsdk:"cache_name"`
	CaptureId            types.Int64  `tfsdk:"capture_id"`
	CaptureLen           types.Int64  `tfsdk:"capture_len"`
	CaptureSample        types.String `tfsdk:"capture_sample"`
	DenyStatus           types.Int64  `tfsdk:"deny_status"`
	Expr                 types.String `tfsdk:"expr"`
	HintFormat           types.String `tfsdk:"hint_format"`
	HintName             types.String `tfsdk:"hint_name"`
	LogLevel             types.String `tfsdk:"log_level"`
	LuaAction            types.String `tfsdk:"lua_action"`
	LuaParams            types.String `tfsdk:"lua_params"`
	MapFile              types.String `tfsdk:"map_file"`
	MapKeyfmt            types.String `tfsdk:"map_keyfmt"`
	MapValuefmt          types.String `tfsdk:"map_valuefmt"`
	MarkValue            types.String `tfsdk:"mark_value"`
	NiceValue            types.Int64  `tfsdk:"nice_value"`
	ReturnContent        types.String `tfsdk:"return_content"`
	ReturnContentFormat  types.String `tfsdk:"return_content_format"`
	ReturnContentType    types.String `tfsdk:"return_content_type"`
	ReturnStatusCode     types.Int64  `tfsdk:"return_status_code"`
	RstTtl               types.Int64  `tfsdk:"rst_ttl"`
	ScExpr               types.String `tfsdk:"sc_expr"`
	ScId                 types.Int64  `tfsdk:"sc_id"`
	ScIdx                types.Int64  `tfsdk:"sc_idx"`
	ScInt                types.Int64  `tfsdk:"sc_int"`
	SpoeEngine           types.String `tfsdk:"spoe_engine"`
	SpoeGroup            types.String `tfsdk:"spoe_group"`
	Status               types.Int64  `tfsdk:"status"`
	StatusReason         types.String `tfsdk:"status_reason"`
	StrictMode           types.String `tfsdk:"strict_mode"`
	Timeout              types.String `tfsdk:"timeout"`
	TimeoutType          types.String `tfsdk:"timeout_type"`
	TosValue             types.String `tfsdk:"tos_value"`
	TrackScKey           types.String `tfsdk:"track_sc_key"`
	TrackScStickCounter  types.Int64  `tfsdk:"track_sc_stick_counter"`
	TrackScTable         types.String `tfsdk:"track_sc_table"`
	VarExpr              types.String `tfsdk:"var_expr"`
	VarFormat            types.String `tfsdk:"var_format"`
	VarName              types.String `tfsdk:"var_name"`
	VarScope             types.String `tfsdk:"var_scope"`
	WaitAtLeast          types.Int64  `tfsdk:"wait_at_least"`
	WaitTime             types.Int64  `tfsdk:"wait_time"`
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
	for i, rule := range sortedRules {
		rulePayload := r.convertToHttpRequestRulePayload(&rule, i)

		if err := r.client.CreateHttpRequestRule(ctx, parentType, parentName, rulePayload); err != nil {
			return fmt.Errorf("failed to create HTTP request rule at index %d: %w", i, err)
		}

		log.Printf("Created HTTP request rule at index %d for %s %s", i, parentType, parentName)
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

	// Use the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
	// Convert all rules to payloads
	var allPayloads []HttpRequestRulePayload
	for i, rule := range sortedRules {
		rulePayload := r.convertToHttpRequestRulePayload(&rule, i)
		allPayloads = append(allPayloads, *rulePayload)
	}

	// Send all rules in one request (same for both v2 and v3)
	if err := r.client.CreateAllHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		return fmt.Errorf("failed to create all HTTP request rules for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Created all %d HTTP request rules for %s %s in transaction %s", len(allPayloads), parentType, parentName, transactionID)

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

// processHttpRequestRulesBlock processes HTTP request rules using array position for ordering and deduplication
func (r *HttpRequestRuleManager) processHttpRequestRulesBlock(rules []haproxyHttpRequestRuleModel) []haproxyHttpRequestRuleModel {
	if len(rules) == 0 {
		return rules
	}

	// Deduplicate rules by creating a map with unique keys
	// Similar to how ACLs are deduplicated by ACL name
	ruleMap := make(map[string]haproxyHttpRequestRuleModel)

	for i, rule := range rules {
		// Generate a unique key based on rule content
		key := r.generateRuleKey(&rule)

		// If a rule with the same key already exists, keep the last one (like ACLs)
		ruleMap[key] = rule
		log.Printf("HTTP request rule %d: key='%s', type='%s', cond='%s'", i, key, rule.Type.ValueString(), rule.Cond.ValueString())
	}

	// Convert map back to slice, maintaining the original order for the first occurrence of each unique rule
	var deduplicatedRules []haproxyHttpRequestRuleModel
	seenKeys := make(map[string]bool)

	for _, rule := range rules {
		key := r.generateRuleKey(&rule)
		if !seenKeys[key] {
			deduplicatedRules = append(deduplicatedRules, rule)
			seenKeys[key] = true
		}
	}

	log.Printf("Deduplicated HTTP request rules: %d original -> %d unique", len(rules), len(deduplicatedRules))
	return deduplicatedRules
}

// generateRuleKey creates a unique key for an HTTP request rule based on its content
func (r *HttpRequestRuleManager) generateRuleKey(rule *haproxyHttpRequestRuleModel) string {
	// Create a key based on the most important fields that would make rules duplicates
	// This is similar to how ACLs use the ACL name as the key
	// Use consistent field ordering and handle empty values
	key := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		rule.Type.ValueString(),
		rule.RedirType.ValueString(),
		rule.RedirValue.ValueString(),
		rule.Cond.ValueString(),
		rule.CondTest.ValueString(),
		rule.HdrName.ValueString(),
	)
	return key
}

// convertToHttpRequestRulePayload converts the Terraform model to HAProxy API payload
func (r *HttpRequestRuleManager) convertToHttpRequestRulePayload(rule *haproxyHttpRequestRuleModel, position int) *HttpRequestRulePayload {
	payload := &HttpRequestRulePayload{
		Index: int64(position), // Use array position instead of index field
		Type:  rule.Type.ValueString(),
	}

	// Set optional fields only if they have values and are not empty
	if !rule.Cond.IsNull() && !rule.Cond.IsUnknown() && rule.Cond.ValueString() != "" {
		payload.Cond = rule.Cond.ValueString()
	}
	if !rule.CondTest.IsNull() && !rule.CondTest.IsUnknown() && rule.CondTest.ValueString() != "" {
		payload.CondTest = rule.CondTest.ValueString()
	}
	if !rule.HdrName.IsNull() && !rule.HdrName.IsUnknown() && rule.HdrName.ValueString() != "" {
		payload.HdrName = rule.HdrName.ValueString()
	}
	if !rule.HdrFormat.IsNull() && !rule.HdrFormat.IsUnknown() && rule.HdrFormat.ValueString() != "" {
		payload.HdrFormat = rule.HdrFormat.ValueString()
	}
	if !rule.HdrMatch.IsNull() && !rule.HdrMatch.IsUnknown() && rule.HdrMatch.ValueString() != "" {
		payload.HdrMatch = rule.HdrMatch.ValueString()
	}
	if !rule.RedirType.IsNull() && !rule.RedirType.IsUnknown() && rule.RedirType.ValueString() != "" {
		payload.RedirType = rule.RedirType.ValueString()
	}
	if !rule.RedirValue.IsNull() && !rule.RedirValue.IsUnknown() && rule.RedirValue.ValueString() != "" {
		payload.RedirValue = rule.RedirValue.ValueString()
	}
	if !rule.RedirCode.IsNull() && !rule.RedirCode.IsUnknown() {
		payload.RedirCode = rule.RedirCode.ValueInt64()
	}
	if !rule.RedirOption.IsNull() && !rule.RedirOption.IsUnknown() && rule.RedirOption.ValueString() != "" {
		payload.RedirOption = rule.RedirOption.ValueString()
	}
	if !rule.BandwidthLimitName.IsNull() && !rule.BandwidthLimitName.IsUnknown() && rule.BandwidthLimitName.ValueString() != "" {
		payload.BandwidthLimitName = rule.BandwidthLimitName.ValueString()
	}
	if !rule.BandwidthLimitLimit.IsNull() && !rule.BandwidthLimitLimit.IsUnknown() && rule.BandwidthLimitLimit.ValueString() != "" {
		payload.BandwidthLimitLimit = rule.BandwidthLimitLimit.ValueString()
	}
	if !rule.BandwidthLimitPeriod.IsNull() && !rule.BandwidthLimitPeriod.IsUnknown() && rule.BandwidthLimitPeriod.ValueString() != "" {
		payload.BandwidthLimitPeriod = rule.BandwidthLimitPeriod.ValueString()
	}
	if !rule.AclFile.IsNull() && !rule.AclFile.IsUnknown() && rule.AclFile.ValueString() != "" {
		payload.AclFile = rule.AclFile.ValueString()
	}
	if !rule.AclKeyfmt.IsNull() && !rule.AclKeyfmt.IsUnknown() && rule.AclKeyfmt.ValueString() != "" {
		payload.AclKeyfmt = rule.AclKeyfmt.ValueString()
	}
	if !rule.AuthRealm.IsNull() && !rule.AuthRealm.IsUnknown() && rule.AuthRealm.ValueString() != "" {
		payload.AuthRealm = rule.AuthRealm.ValueString()
	}
	if !rule.CacheName.IsNull() && !rule.CacheName.IsUnknown() && rule.CacheName.ValueString() != "" {
		payload.CacheName = rule.CacheName.ValueString()
	}
	if !rule.CaptureId.IsNull() && !rule.CaptureId.IsUnknown() {
		payload.CaptureID = rule.CaptureId.ValueInt64()
	}
	if !rule.CaptureLen.IsNull() && !rule.CaptureLen.IsUnknown() {
		payload.CaptureLen = rule.CaptureLen.ValueInt64()
	}
	if !rule.CaptureSample.IsNull() && !rule.CaptureSample.IsUnknown() && rule.CaptureSample.ValueString() != "" {
		payload.CaptureSample = rule.CaptureSample.ValueString()
	}
	if !rule.DenyStatus.IsNull() && !rule.DenyStatus.IsUnknown() {
		payload.DenyStatus = rule.DenyStatus.ValueInt64()
	}
	if !rule.Expr.IsNull() && !rule.Expr.IsUnknown() && rule.Expr.ValueString() != "" {
		payload.Expr = rule.Expr.ValueString()
	}
	if !rule.HintFormat.IsNull() && !rule.HintFormat.IsUnknown() && rule.HintFormat.ValueString() != "" {
		payload.HintFormat = rule.HintFormat.ValueString()
	}
	if !rule.HintName.IsNull() && !rule.HintName.IsUnknown() && rule.HintName.ValueString() != "" {
		payload.HintName = rule.HintName.ValueString()
	}
	if !rule.LogLevel.IsNull() && !rule.LogLevel.IsUnknown() && rule.LogLevel.ValueString() != "" {
		payload.LogLevel = rule.LogLevel.ValueString()
	}
	if !rule.LuaAction.IsNull() && !rule.LuaAction.IsUnknown() && rule.LuaAction.ValueString() != "" {
		payload.LuaAction = rule.LuaAction.ValueString()
	}
	if !rule.LuaParams.IsNull() && !rule.LuaParams.IsUnknown() && rule.LuaParams.ValueString() != "" {
		payload.LuaParams = rule.LuaParams.ValueString()
	}
	if !rule.MapFile.IsNull() && !rule.MapFile.IsUnknown() && rule.MapFile.ValueString() != "" {
		payload.MapFile = rule.MapFile.ValueString()
	}
	if !rule.MapKeyfmt.IsNull() && !rule.MapKeyfmt.IsUnknown() && rule.MapKeyfmt.ValueString() != "" {
		payload.MapKeyfmt = rule.MapKeyfmt.ValueString()
	}
	if !rule.MapValuefmt.IsNull() && !rule.MapValuefmt.IsUnknown() && rule.MapValuefmt.ValueString() != "" {
		payload.MapValuefmt = rule.MapValuefmt.ValueString()
	}
	if !rule.MarkValue.IsNull() && !rule.MarkValue.IsUnknown() && rule.MarkValue.ValueString() != "" {
		payload.MarkValue = rule.MarkValue.ValueString()
	}
	if !rule.NiceValue.IsNull() && !rule.NiceValue.IsUnknown() {
		payload.NiceValue = rule.NiceValue.ValueInt64()
	}
	if !rule.ReturnContent.IsNull() && !rule.ReturnContent.IsUnknown() && rule.ReturnContent.ValueString() != "" {
		payload.ReturnContent = rule.ReturnContent.ValueString()
	}
	if !rule.ReturnContentFormat.IsNull() && !rule.ReturnContentFormat.IsUnknown() && rule.ReturnContentFormat.ValueString() != "" {
		payload.ReturnContentFormat = rule.ReturnContentFormat.ValueString()
	}
	if !rule.ReturnContentType.IsNull() && !rule.ReturnContentType.IsUnknown() && rule.ReturnContentType.ValueString() != "" {
		payload.ReturnContentType = rule.ReturnContentType.ValueString()
	}
	if !rule.ReturnStatusCode.IsNull() && !rule.ReturnStatusCode.IsUnknown() {
		payload.ReturnStatusCode = rule.ReturnStatusCode.ValueInt64()
	}
	if !rule.RstTtl.IsNull() && !rule.RstTtl.IsUnknown() {
		payload.RstTtl = rule.RstTtl.ValueInt64()
	}
	if !rule.SpoeEngine.IsNull() && !rule.SpoeEngine.IsUnknown() && rule.SpoeEngine.ValueString() != "" {
		payload.SpoeEngine = rule.SpoeEngine.ValueString()
	}
	if !rule.SpoeGroup.IsNull() && !rule.SpoeGroup.IsUnknown() && rule.SpoeGroup.ValueString() != "" {
		payload.SpoeGroup = rule.SpoeGroup.ValueString()
	}
	if !rule.Status.IsNull() && !rule.Status.IsUnknown() {
		payload.StatusCode = rule.Status.ValueInt64()
	}
	if !rule.StatusReason.IsNull() && !rule.StatusReason.IsUnknown() && rule.StatusReason.ValueString() != "" {
		payload.StatusReason = rule.StatusReason.ValueString()
	}
	if !rule.StrictMode.IsNull() && !rule.StrictMode.IsUnknown() && rule.StrictMode.ValueString() != "" {
		payload.StrictMode = rule.StrictMode.ValueString()
	}
	if !rule.Timeout.IsNull() && !rule.Timeout.IsUnknown() && rule.Timeout.ValueString() != "" {
		payload.Timeout = rule.Timeout.ValueString()
	}
	if !rule.TimeoutType.IsNull() && !rule.TimeoutType.IsUnknown() && rule.TimeoutType.ValueString() != "" {
		payload.TimeoutType = rule.TimeoutType.ValueString()
	}

	// Debug logging to see what's being sent
	log.Printf("DEBUG: HTTP request rule payload: Type=%s, Cond=%s, CondTest=%s, HdrName=%s, HdrFormat=%s, BandwidthLimitName=%s, BandwidthLimitLimit=%s, BandwidthLimitPeriod=%s",
		payload.Type, payload.Cond, payload.CondTest, payload.HdrName, payload.HdrFormat, payload.BandwidthLimitName, payload.BandwidthLimitLimit, payload.BandwidthLimitPeriod)

	// Validate that required fields are present for specific rule types
	if payload.Type == "set-header" && (payload.HdrName == "" || payload.HdrFormat == "") {
		log.Printf("WARNING: set-header rule at position %d is missing hdr_name or hdr_format, this may cause formatting issues", position)
	}
	if payload.Type == "set-bandwidth-limit" && (payload.BandwidthLimitName == "" || payload.BandwidthLimitLimit == "" || payload.BandwidthLimitPeriod == "") {
		log.Printf("WARNING: set-bandwidth-limit rule at position %d is missing bandwidth_limit_name, bandwidth_limit_limit, or bandwidth_limit_period, this may cause formatting issues", position)
	}

	return payload
}

// updateHttpRequestRulesWithIndexing handles the complex logic of updating HTTP request rules while maintaining order
func (r *HttpRequestRuleManager) updateHttpRequestRulesWithIndexing(ctx context.Context, parentType string, parentName string, existingRules []HttpRequestRulePayload, newRules []haproxyHttpRequestRuleModel) error {
	// Process new rules with proper indexing and deduplication
	sortedNewRules := r.processHttpRequestRulesBlock(newRules)

	// For consistency with create operations, use the same "create all at once" approach
	// This ensures consistent formatting from HAProxy API
	if r.client.apiVersion == "v3" {
		// Convert all rules to payloads
		var allPayloads []HttpRequestRulePayload
		for i, rule := range sortedNewRules {
			rulePayload := r.convertToHttpRequestRulePayload(&rule, i)
			allPayloads = append(allPayloads, *rulePayload)
		}

		// Use transaction to get consistent formatting
		_, err := r.client.Transaction(func(transactionID string) (*http.Response, error) {
			// Send all rules in one request (same as create)
			if err := r.client.CreateAllHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
				return nil, fmt.Errorf("failed to update all HTTP request rules for %s %s: %w", parentType, parentName, err)
			}
			return nil, nil
		})
		if err != nil {
			return err
		}

		log.Printf("Updated all %d HTTP request rules for %s %s", len(allPayloads), parentType, parentName)
		return nil
	}

	// Fallback to individual operations for v2
	// Convert desired rules to map for easier comparison (similar to ACLs)
	desiredMap := make(map[string]HttpRequestRulePayload)
	for i, rule := range sortedNewRules {
		ruleKey := r.generateRuleKey(&rule)
		log.Printf("DEBUG: Desired HTTP request rule: %s (array position: %d)", ruleKey, i)
		desiredMap[ruleKey] = HttpRequestRulePayload{
			Index:      int64(i), // Use array position instead of index field
			Type:       rule.Type.ValueString(),
			Cond:       rule.Cond.ValueString(),
			CondTest:   rule.CondTest.ValueString(),
			HdrName:    rule.HdrName.ValueString(),
			HdrFormat:  rule.HdrFormat.ValueString(),
			RedirType:  rule.RedirType.ValueString(),
			RedirValue: rule.RedirValue.ValueString(),
		}
	}

	// Convert existing rules to map for easier comparison
	existingMap := make(map[string]HttpRequestRulePayload)
	for i, rule := range existingRules {
		// Use array position instead of API index since HAProxy API returns wrong indices
		rule.Index = int64(i)
		ruleKey := r.generateRuleKeyFromPayload(&rule)
		log.Printf("DEBUG: Found existing HTTP request rule: %s (corrected index: %d)", ruleKey, rule.Index)
		existingMap[ruleKey] = rule
	}

	// Find rules to delete (exist in HAProxy but not in desired state)
	var rulesToDelete []HttpRequestRulePayload
	for key, existingRule := range existingMap {
		if _, exists := desiredMap[key]; !exists {
			rulesToDelete = append(rulesToDelete, existingRule)
		}
	}

	// Find rules to create (exist in desired state but not in HAProxy)
	var rulesToCreate []HttpRequestRulePayload
	for key, desiredRule := range desiredMap {
		if _, exists := existingMap[key]; !exists {
			rulesToCreate = append(rulesToCreate, desiredRule)
		}
	}

	// Find rules to update (exist in both but have different content or position)
	var rulesToUpdate []HttpRequestRulePayload
	for key, desiredRule := range desiredMap {
		if existingRule, exists := existingMap[key]; exists {
			if r.hasRuleChangedFromPayload(&existingRule, &desiredRule) {
				log.Printf("DEBUG: HTTP request rule '%s' content changed, will update", key)
				rulesToUpdate = append(rulesToUpdate, desiredRule)
			} else if existingRule.Index != desiredRule.Index {
				log.Printf("DEBUG: HTTP request rule '%s' position changed from %d to %d, will reorder", key, existingRule.Index, desiredRule.Index)
				rulesToUpdate = append(rulesToUpdate, desiredRule)
			}
		}
	}

	// Delete rules that are no longer needed
	for _, rule := range rulesToDelete {
		log.Printf("Deleting HTTP request rule '%s' at index %d", r.generateRuleKeyFromPayload(&rule), rule.Index)
		if err := r.client.DeleteHttpRequestRule(ctx, rule.Index, parentType, parentName); err != nil {
			return fmt.Errorf("failed to delete HTTP request rule: %w", err)
		}
	}

	// Update rules that have changed
	for _, rule := range rulesToUpdate {
		log.Printf("Updating HTTP request rule '%s' at index %d", r.generateRuleKeyFromPayload(&rule), rule.Index)
		if err := r.client.UpdateHttpRequestRule(ctx, rule.Index, parentType, parentName, &rule); err != nil {
			return fmt.Errorf("failed to update HTTP request rule: %w", err)
		}
	}

	// Create new rules
	for _, rule := range rulesToCreate {
		log.Printf("Creating HTTP request rule '%s' at index %d", r.generateRuleKeyFromPayload(&rule), rule.Index)
		if err := r.client.CreateHttpRequestRule(ctx, parentType, parentName, &rule); err != nil {
			return fmt.Errorf("failed to create HTTP request rule: %w", err)
		}
	}

	return nil
}

// generateRuleKeyFromPayload creates a unique key for an HTTP request rule payload based on its content
func (r *HttpRequestRuleManager) generateRuleKeyFromPayload(rule *HttpRequestRulePayload) string {
	// Create a key based on the most important fields that would make rules duplicates
	// Use consistent field ordering and handle empty values - must match generateRuleKey exactly
	key := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		rule.Type,
		rule.RedirType,
		rule.RedirValue,
		rule.Cond,
		rule.CondTest,
		rule.HdrName,
	)
	return key
}

// hasRuleChangedFromPayload compares two HTTP request rule payloads to determine if they have different content
func (r *HttpRequestRuleManager) hasRuleChangedFromPayload(existing, desired *HttpRequestRulePayload) bool {
	return existing.Type != desired.Type ||
		existing.Cond != desired.Cond ||
		existing.CondTest != desired.CondTest ||
		existing.HdrName != desired.HdrName ||
		existing.HdrFormat != desired.HdrFormat ||
		existing.RedirType != desired.RedirType ||
		existing.RedirValue != desired.RedirValue
}

// hasRuleChanged checks if an HTTP request rule has changed between existing and new configurations
func (r *HttpRequestRuleManager) hasRuleChanged(existing *HttpRequestRulePayload, new *haproxyHttpRequestRuleModel) bool {
	return existing.Type != new.Type.ValueString() ||
		existing.Cond != new.Cond.ValueString() ||
		existing.CondTest != new.CondTest.ValueString() ||
		existing.HdrName != new.HdrName.ValueString() ||
		existing.HdrFormat != new.HdrFormat.ValueString() ||
		existing.RedirType != new.RedirType.ValueString() ||
		existing.RedirValue != new.RedirValue.ValueString()
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
			return fmt.Errorf("failed to delete HTTP request rule: %w", err)
		}
	}

	return nil
}

// UpdateHttpRequestRulesInTransaction updates HTTP request rules using an existing transaction ID with smart comparison
func (r *HttpRequestRuleManager) UpdateHttpRequestRulesInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, rules []haproxyHttpRequestRuleModel) error {
	return r.updateHttpRequestRulesWithIndexingInTransaction(ctx, transactionID, parentType, parentName, rules)
}

// updateHttpRequestRulesWithIndexingInTransaction performs smart HTTP request rule updates by comparing existing vs desired
func (r *HttpRequestRuleManager) updateHttpRequestRulesWithIndexingInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, desiredRules []haproxyHttpRequestRuleModel) error {
	log.Printf("DEBUG: Starting HTTP request rule update for %s '%s' with %d desired rules", parentType, parentName, len(desiredRules))

	// Read existing HTTP request rules to compare with desired ones
	existingRules, err := r.client.ReadHttpRequestRules(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing HTTP request rules for %s %s: %w", parentType, parentName, err)
	}

	// Use the smart comparison logic to only update what changed
	return r.updateHttpRequestRulesWithIndexingInTransactionSmart(ctx, transactionID, parentType, parentName, existingRules, desiredRules)
}

// updateHttpRequestRulesWithIndexingInTransactionSmart performs smart HTTP request rule updates by comparing existing vs desired
func (r *HttpRequestRuleManager) updateHttpRequestRulesWithIndexingInTransactionSmart(ctx context.Context, transactionID string, parentType string, parentName string, existingRules []HttpRequestRulePayload, desiredRules []haproxyHttpRequestRuleModel) error {
	// Process desired rules with proper indexing and deduplication
	sortedDesiredRules := r.processHttpRequestRulesBlock(desiredRules)

	// Create a map of existing rules by content key for quick lookup (since HAProxy API v3 returns all with index 0)
	existingRuleMap := make(map[string]*HttpRequestRulePayload)
	for i := range existingRules {
		ruleKey := r.generateRuleKeyFromPayload(&existingRules[i])
		log.Printf("DEBUG: Existing HTTP request rule %d - Type='%s', Cond='%s', CondTest='%s', HdrName='%s', RedirType='%s', RedirValue='%s'",
			i, existingRules[i].Type, existingRules[i].Cond, existingRules[i].CondTest, existingRules[i].HdrName, existingRules[i].RedirType, existingRules[i].RedirValue)
		log.Printf("DEBUG: Existing HTTP request rule %d - generated key: '%s'", i, ruleKey)
		existingRuleMap[ruleKey] = &existingRules[i]
	}

	// Check if any rules actually changed (content OR order)
	hasChanges := false
	var rulesToRecreate []haproxyHttpRequestRuleModel

	log.Printf("DEBUG: Comparing %d existing HTTP request rules with %d desired rules for %s %s", len(existingRules), len(sortedDesiredRules), parentType, parentName)

	// First, check if the number of rules changed
	if len(existingRules) != len(sortedDesiredRules) {
		log.Printf("DEBUG: Rule count changed from %d to %d, marking for recreation", len(existingRules), len(sortedDesiredRules))
		hasChanges = true
		rulesToRecreate = sortedDesiredRules
	} else {
		// Check if rules have changed content OR order
		for i, desiredRule := range sortedDesiredRules {
			desiredKey := r.generateRuleKey(&desiredRule)
			existingRule, exists := existingRuleMap[desiredKey]

			log.Printf("DEBUG: Desired HTTP request rule %d - Type='%s', Cond='%s', CondTest='%s', HdrName='%s', RedirType='%s', RedirValue='%s'",
				i, desiredRule.Type.ValueString(), desiredRule.Cond.ValueString(), desiredRule.CondTest.ValueString(), desiredRule.HdrName.ValueString(), desiredRule.RedirType.ValueString(), desiredRule.RedirValue.ValueString())
			log.Printf("DEBUG: HTTP request rule %d - desired key: '%s'", i, desiredKey)

			if !exists {
				log.Printf("DEBUG: HTTP request rule %d - no existing rule with key '%s', checking if it's a content change", i, desiredKey)
				// Check if this might be a content change by looking for a rule at the same position
				if i < len(existingRules) {
					existingKey := r.generateRuleKeyFromPayload(&existingRules[i])
					log.Printf("DEBUG: HTTP request rule %d - existing rule at position %d has key '%s'", i, i, existingKey)
					if desiredKey != existingKey {
						log.Printf("DEBUG: HTTP request rule %d - content change detected at position %d - desired: '%s', existing: '%s'", i, i, desiredKey, existingKey)
						hasChanges = true
						rulesToRecreate = sortedDesiredRules
						break
					}
				} else {
					log.Printf("DEBUG: HTTP request rule %d - truly new rule, marking for recreation", i)
					hasChanges = true
					rulesToRecreate = sortedDesiredRules
					break
				}
			} else {
				log.Printf("DEBUG: HTTP request rule %d - found existing rule with key '%s'", i, desiredKey)
				changed := hasHttpRequestRuleChanged(*existingRule, *r.convertToHttpRequestRulePayload(&desiredRule, i))
				log.Printf("DEBUG: HTTP request rule %d - hasHttpRequestRuleChanged returned: %t", i, changed)
				if changed {
					log.Printf("DEBUG: HTTP request rule %d - marked for recreation due to content changes", i)
					hasChanges = true
					rulesToRecreate = sortedDesiredRules
					break
				}
			}
		}

		// If no content changes detected, check for order changes by comparing the sequence
		if !hasChanges {
			log.Printf("DEBUG: No content changes detected, checking for order changes...")
			for i, desiredRule := range sortedDesiredRules {
				desiredKey := r.generateRuleKey(&desiredRule)
				// Check if the rule at position i has the same key as the existing rule at position i
				if i < len(existingRules) {
					existingKey := r.generateRuleKeyFromPayload(&existingRules[i])
					if desiredKey != existingKey {
						log.Printf("DEBUG: Order change detected at position %d - desired: '%s', existing: '%s'", i, desiredKey, existingKey)
						hasChanges = true
						rulesToRecreate = sortedDesiredRules
						break
					}
				}
			}
		}
	}

	// Also check if any existing rules need to be removed (not in desired list)
	if !hasChanges {
		for existingKey := range existingRuleMap {
			found := false
			for _, desiredRule := range sortedDesiredRules {
				if r.generateRuleKey(&desiredRule) == existingKey {
					found = true
					break
				}
			}
			if !found {
				log.Printf("DEBUG: Existing HTTP request rule '%s' not in desired list, marking for removal", existingKey)
				hasChanges = true
				rulesToRecreate = sortedDesiredRules
				break
			}
		}
	}

	log.Printf("DEBUG: Final hasChanges result: %t, HTTP request rules to recreate: %d", hasChanges, len(rulesToRecreate))

	// If no changes detected, skip the update
	if !hasChanges {
		log.Printf("No HTTP request rule changes detected for %s %s, skipping update", parentType, parentName)
		return nil
	}

	// First, delete all existing HTTP request rules to avoid duplicates
	if err := r.deleteAllHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName); err != nil {
		return fmt.Errorf("failed to delete existing HTTP request rules for %s %s: %w", parentType, parentName, err)
	}

	// Then create all desired rules using the same "create all at once" approach for both v2 and v3
	// This ensures consistent formatting from HAProxy API
	// Process rules with proper indexing and deduplication
	sortedRules := r.processHttpRequestRulesBlock(desiredRules)

	// Convert all rules to payloads
	var allPayloads []HttpRequestRulePayload
	for i, rule := range sortedRules {
		rulePayload := r.convertToHttpRequestRulePayload(&rule, i)
		allPayloads = append(allPayloads, *rulePayload)
	}

	// Send all rules in one request (same for both v2 and v3)
	if err := r.client.CreateAllHttpRequestRulesInTransaction(ctx, transactionID, parentType, parentName, allPayloads); err != nil {
		return fmt.Errorf("failed to create new HTTP request rules for %s %s: %w", parentType, parentName, err)
	}

	log.Printf("Updated %d HTTP request rules for %s %s in transaction %s (delete-then-create)", len(allPayloads), parentType, parentName, transactionID)
	return nil
}

// hasHttpRequestRuleChanged compares two HTTP request rules to determine if they have different content
func hasHttpRequestRuleChanged(existing, desired HttpRequestRulePayload) bool {
	return existing.Type != desired.Type || existing.Cond != desired.Cond || existing.CondTest != desired.CondTest
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
		order = append(order, rule.Type.ValueString())
	}
	return fmt.Sprintf("[%s]", strings.Join(order, ", "))
}
