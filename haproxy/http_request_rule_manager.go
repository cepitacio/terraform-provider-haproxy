package haproxy

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

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
	if !rule.HdrMatch.IsNull() && !rule.HdrMatch.IsUnknown() {
		payload.HdrMatch = rule.HdrMatch.ValueString()
	}
	if !rule.RedirType.IsNull() && !rule.RedirType.IsUnknown() {
		payload.RedirType = rule.RedirType.ValueString()
	}
	if !rule.RedirValue.IsNull() && !rule.RedirValue.IsUnknown() {
		payload.RedirValue = rule.RedirValue.ValueString()
	}
	if !rule.RedirCode.IsNull() && !rule.RedirCode.IsUnknown() {
		payload.RedirCode = rule.RedirCode.ValueInt64()
	}
	if !rule.RedirOption.IsNull() && !rule.RedirOption.IsUnknown() {
		payload.RedirOption = rule.RedirOption.ValueString()
	}
	if !rule.PathMatch.IsNull() && !rule.PathMatch.IsUnknown() {
		payload.PathMatch = rule.PathMatch.ValueString()
	}
	if !rule.PathFmt.IsNull() && !rule.PathFmt.IsUnknown() {
		payload.PathFmt = rule.PathFmt.ValueString()
	}
	if !rule.UriMatch.IsNull() && !rule.UriMatch.IsUnknown() {
		payload.UriMatch = rule.UriMatch.ValueString()
	}
	if !rule.UriFmt.IsNull() && !rule.UriFmt.IsUnknown() {
		payload.UriFmt = rule.UriFmt.ValueString()
	}
	if !rule.QueryFmt.IsNull() && !rule.QueryFmt.IsUnknown() {
		payload.QueryFmt = rule.QueryFmt.ValueString()
	}
	if !rule.MethodFmt.IsNull() && !rule.MethodFmt.IsUnknown() {
		payload.MethodFmt = rule.MethodFmt.ValueString()
	}
	if !rule.VarName.IsNull() && !rule.VarName.IsUnknown() {
		payload.VarName = rule.VarName.ValueString()
	}
	if !rule.VarFormat.IsNull() && !rule.VarFormat.IsUnknown() {
		payload.VarFormat = rule.VarFormat.ValueString()
	}
	if !rule.VarExpr.IsNull() && !rule.VarExpr.IsUnknown() {
		payload.VarExpr = rule.VarExpr.ValueString()
	}
	if !rule.VarScope.IsNull() && !rule.VarScope.IsUnknown() {
		payload.VarScope = rule.VarScope.ValueString()
	}
	if !rule.CaptureID.IsNull() && !rule.CaptureID.IsUnknown() {
		payload.CaptureID = rule.CaptureID.ValueInt64()
	}
	if !rule.CaptureLen.IsNull() && !rule.CaptureLen.IsUnknown() {
		payload.CaptureLen = rule.CaptureLen.ValueInt64()
	}
	if !rule.CaptureSample.IsNull() && !rule.CaptureSample.IsUnknown() {
		payload.CaptureSample = rule.CaptureSample.ValueString()
	}
	if !rule.LogLevel.IsNull() && !rule.LogLevel.IsUnknown() {
		payload.LogLevel = rule.LogLevel.ValueString()
	}
	if !rule.Timeout.IsNull() && !rule.Timeout.IsUnknown() {
		payload.Timeout = rule.Timeout.ValueString()
	}
	if !rule.TimeoutType.IsNull() && !rule.TimeoutType.IsUnknown() {
		payload.TimeoutType = rule.TimeoutType.ValueString()
	}
	if !rule.StrictMode.IsNull() && !rule.StrictMode.IsUnknown() {
		payload.StrictMode = rule.StrictMode.ValueString()
	}
	if !rule.Normalizer.IsNull() && !rule.Normalizer.IsUnknown() {
		payload.Normalizer = rule.Normalizer.ValueString()
	}
	if !rule.NormalizerFull.IsNull() && !rule.NormalizerFull.IsUnknown() {
		payload.NormalizerFull = rule.NormalizerFull.ValueBool()
	}
	if !rule.NormalizerStrict.IsNull() && !rule.NormalizerStrict.IsUnknown() {
		payload.NormalizerStrict = rule.NormalizerStrict.ValueBool()
	}
	if !rule.NiceValue.IsNull() && !rule.NiceValue.IsUnknown() {
		payload.NiceValue = rule.NiceValue.ValueInt64()
	}
	if !rule.MarkValue.IsNull() && !rule.MarkValue.IsUnknown() {
		payload.MarkValue = rule.MarkValue.ValueString()
	}
	if !rule.TosValue.IsNull() && !rule.TosValue.IsUnknown() {
		payload.TosValue = rule.TosValue.ValueString()
	}
	if !rule.TrackScKey.IsNull() && !rule.TrackScKey.IsUnknown() {
		payload.TrackScKey = rule.TrackScKey.ValueString()
	}
	if !rule.TrackScTable.IsNull() && !rule.TrackScTable.IsUnknown() {
		payload.TrackScTable = rule.TrackScTable.ValueString()
	}
	if !rule.TrackScID.IsNull() && !rule.TrackScID.IsUnknown() {
		payload.TrackScID = rule.TrackScID.ValueInt64()
	}
	if !rule.TrackScIdx.IsNull() && !rule.TrackScIdx.IsUnknown() {
		payload.TrackScIdx = rule.TrackScIdx.ValueInt64()
	}
	if !rule.TrackScInt.IsNull() && !rule.TrackScInt.IsUnknown() {
		payload.TrackScInt = rule.TrackScInt.ValueInt64()
	}
	if !rule.ReturnStatusCode.IsNull() && !rule.ReturnStatusCode.IsUnknown() {
		payload.ReturnStatusCode = rule.ReturnStatusCode.ValueInt64()
	}
	if !rule.ReturnContent.IsNull() && !rule.ReturnContent.IsUnknown() {
		payload.ReturnContent = rule.ReturnContent.ValueString()
	}
	if !rule.ReturnContentType.IsNull() && !rule.ReturnContentType.IsUnknown() {
		payload.ReturnContentType = rule.ReturnContentType.ValueString()
	}
	if !rule.ReturnContentFormat.IsNull() && !rule.ReturnContentFormat.IsUnknown() {
		payload.ReturnContentFormat = rule.ReturnContentFormat.ValueString()
	}
	if !rule.DenyStatus.IsNull() && !rule.DenyStatus.IsUnknown() {
		payload.DenyStatus = rule.DenyStatus.ValueInt64()
	}
	if !rule.WaitTime.IsNull() && !rule.WaitTime.IsUnknown() {
		payload.WaitTime = rule.WaitTime.ValueInt64()
	}
	if !rule.WaitAtLeast.IsNull() && !rule.WaitAtLeast.IsUnknown() {
		payload.WaitAtLeast = rule.WaitAtLeast.ValueInt64()
	}
	if !rule.Expr.IsNull() && !rule.Expr.IsUnknown() {
		payload.Expr = rule.Expr.ValueString()
	}
	if !rule.LuaAction.IsNull() && !rule.LuaAction.IsUnknown() {
		payload.LuaAction = rule.LuaAction.ValueString()
	}
	if !rule.LuaParams.IsNull() && !rule.LuaParams.IsUnknown() {
		payload.LuaParams = rule.LuaParams.ValueString()
	}
	if !rule.SpoeEngine.IsNull() && !rule.SpoeEngine.IsUnknown() {
		payload.SpoeEngine = rule.SpoeEngine.ValueString()
	}
	if !rule.SpoeGroup.IsNull() && !rule.SpoeGroup.IsUnknown() {
		payload.SpoeGroup = rule.SpoeGroup.ValueString()
	}
	if !rule.ServiceName.IsNull() && !rule.ServiceName.IsUnknown() {
		payload.ServiceName = rule.ServiceName.ValueString()
	}
	if !rule.CacheName.IsNull() && !rule.CacheName.IsUnknown() {
		payload.CacheName = rule.CacheName.ValueString()
	}
	if !rule.Resolvers.IsNull() && !rule.Resolvers.IsUnknown() {
		payload.Resolvers = rule.Resolvers.ValueString()
	}
	if !rule.Protocol.IsNull() && !rule.Protocol.IsUnknown() {
		payload.Protocol = rule.Protocol.ValueString()
	}
	if !rule.BandwidthLimitName.IsNull() && !rule.BandwidthLimitName.IsUnknown() {
		payload.BandwidthLimitName = rule.BandwidthLimitName.ValueString()
	}
	if !rule.BandwidthLimitLimit.IsNull() && !rule.BandwidthLimitLimit.IsUnknown() {
		payload.BandwidthLimitLimit = rule.BandwidthLimitLimit.ValueString()
	}
	if !rule.BandwidthLimitPeriod.IsNull() && !rule.BandwidthLimitPeriod.IsUnknown() {
		payload.BandwidthLimitPeriod = rule.BandwidthLimitPeriod.ValueString()
	}
	if !rule.MapFile.IsNull() && !rule.MapFile.IsUnknown() {
		payload.MapFile = rule.MapFile.ValueString()
	}
	if !rule.MapKeyfmt.IsNull() && !rule.MapKeyfmt.IsUnknown() {
		payload.MapKeyfmt = rule.MapKeyfmt.ValueString()
	}
	if !rule.MapValuefmt.IsNull() && !rule.MapValuefmt.IsUnknown() {
		payload.MapValuefmt = rule.MapValuefmt.ValueString()
	}
	if !rule.AclFile.IsNull() && !rule.AclFile.IsUnknown() {
		payload.AclFile = rule.AclFile.ValueString()
	}
	if !rule.AclKeyfmt.IsNull() && !rule.AclKeyfmt.IsUnknown() {
		payload.AclKeyfmt = rule.AclKeyfmt.ValueString()
	}
	if !rule.AuthRealm.IsNull() && !rule.AuthRealm.IsUnknown() {
		payload.AuthRealm = rule.AuthRealm.ValueString()
	}
	if !rule.HintName.IsNull() && !rule.HintName.IsUnknown() {
		payload.HintName = rule.HintName.ValueString()
	}
	if !rule.HintFormat.IsNull() && !rule.HintFormat.IsUnknown() {
		payload.HintFormat = rule.HintFormat.ValueString()
	}
	if !rule.ScExpr.IsNull() && !rule.ScExpr.IsUnknown() {
		payload.ScExpr = rule.ScExpr.ValueString()
	}
	if !rule.ScID.IsNull() && !rule.ScID.IsUnknown() {
		payload.ScID = rule.ScID.ValueInt64()
	}
	if !rule.ScIdx.IsNull() && !rule.ScIdx.IsUnknown() {
		payload.ScIdx = rule.ScIdx.ValueInt64()
	}
	if !rule.ScInt.IsNull() && !rule.ScInt.IsUnknown() {
		payload.ScInt = rule.ScInt.ValueInt64()
	}
	if !rule.ScAddGpc.IsNull() && !rule.ScAddGpc.IsUnknown() {
		payload.ScAddGpc = rule.ScAddGpc.ValueString()
	}
	if !rule.ScIncGpc.IsNull() && !rule.ScIncGpc.IsUnknown() {
		payload.ScIncGpc = rule.ScIncGpc.ValueString()
	}
	if !rule.ScIncGpc0.IsNull() && !rule.ScIncGpc0.IsUnknown() {
		payload.ScIncGpc0 = rule.ScIncGpc0.ValueString()
	}
	if !rule.ScIncGpc1.IsNull() && !rule.ScIncGpc1.IsUnknown() {
		payload.ScIncGpc1 = rule.ScIncGpc1.ValueString()
	}
	if !rule.ScSetGpt.IsNull() && !rule.ScSetGpt.IsUnknown() {
		payload.ScSetGpt = rule.ScSetGpt.ValueString()
	}
	if !rule.ScSetGpt0.IsNull() && !rule.ScSetGpt0.IsUnknown() {
		payload.ScSetGpt0 = rule.ScSetGpt0.ValueString()
	}
	if !rule.SetPriorityClass.IsNull() && !rule.SetPriorityClass.IsUnknown() {
		payload.SetPriorityClass = rule.SetPriorityClass.ValueString()
	}
	if !rule.SetPriorityOffset.IsNull() && !rule.SetPriorityOffset.IsUnknown() {
		payload.SetPriorityOffset = rule.SetPriorityOffset.ValueString()
	}
	if !rule.SetRetries.IsNull() && !rule.SetRetries.IsUnknown() {
		payload.SetRetries = rule.SetRetries.ValueString()
	}
	if !rule.SetBcMark.IsNull() && !rule.SetBcMark.IsUnknown() {
		payload.SetBcMark = rule.SetBcMark.ValueString()
	}
	if !rule.SetBcTos.IsNull() && !rule.SetBcTos.IsUnknown() {
		payload.SetBcTos = rule.SetBcTos.ValueString()
	}
	if !rule.SetFcMark.IsNull() && !rule.SetFcMark.IsUnknown() {
		payload.SetFcMark = rule.SetFcMark.ValueString()
	}
	if !rule.SetFcTos.IsNull() && !rule.SetFcTos.IsUnknown() {
		payload.SetFcTos = rule.SetFcTos.ValueString()
	}
	if !rule.SetDst.IsNull() && !rule.SetDst.IsUnknown() {
		payload.SetDst = rule.SetDst.ValueString()
	}
	if !rule.SetDstPort.IsNull() && !rule.SetDstPort.IsUnknown() {
		// Convert string to int64 for the API payload
		if port, err := strconv.ParseInt(rule.SetDstPort.ValueString(), 10, 64); err == nil {
			payload.SetDstPort = port
		}
	}
	if !rule.SetSrc.IsNull() && !rule.SetSrc.IsUnknown() {
		payload.SetSrc = rule.SetSrc.ValueString()
	}
	if !rule.SetSrcPort.IsNull() && !rule.SetSrcPort.IsUnknown() {
		// Convert string to int64 for the API payload
		if port, err := strconv.ParseInt(rule.SetSrcPort.ValueString(), 10, 64); err == nil {
			payload.SetSrcPort = port
		}
	}
	if !rule.SetTimeout.IsNull() && !rule.SetTimeout.IsUnknown() {
		payload.SetTimeout = rule.SetTimeout.ValueString()
	}
	if !rule.SetTos.IsNull() && !rule.SetTos.IsUnknown() {
		payload.SetTos = rule.SetTos.ValueString()
	}
	if !rule.SetMark.IsNull() && !rule.SetMark.IsUnknown() {
		payload.SetMark = rule.SetMark.ValueString()
	}
	if !rule.SetVar.IsNull() && !rule.SetVar.IsUnknown() {
		payload.SetVar = rule.SetVar.ValueString()
	}
	if !rule.SetVarFmt.IsNull() && !rule.SetVarFmt.IsUnknown() {
		payload.SetVarFmt = rule.SetVarFmt.ValueString()
	}
	if !rule.UnsetVar.IsNull() && !rule.UnsetVar.IsUnknown() {
		payload.UnsetVar = rule.UnsetVar.ValueString()
	}
	if !rule.EarlyHint.IsNull() && !rule.EarlyHint.IsUnknown() {
		payload.EarlyHint = rule.EarlyHint.ValueString()
	}
	if !rule.UseService.IsNull() && !rule.UseService.IsUnknown() {
		payload.UseService = rule.UseService.ValueString()
	}
	if !rule.WaitForBody.IsNull() && !rule.WaitForBody.IsUnknown() {
		payload.WaitForBody = rule.WaitForBody.ValueString()
	}
	if !rule.WaitForHandshake.IsNull() && !rule.WaitForHandshake.IsUnknown() {
		payload.WaitForHandshake = rule.WaitForHandshake.ValueString()
	}
	if !rule.SilentDrop.IsNull() && !rule.SilentDrop.IsUnknown() {
		payload.SilentDrop = rule.SilentDrop.ValueString()
	}
	if !rule.Tarpit.IsNull() && !rule.Tarpit.IsUnknown() {
		payload.Tarpit = rule.Tarpit.ValueString()
	}
	if !rule.DisableL7Retry.IsNull() && !rule.DisableL7Retry.IsUnknown() {
		payload.DisableL7Retry = rule.DisableL7Retry.ValueString()
	}
	if !rule.DoResolve.IsNull() && !rule.DoResolve.IsUnknown() {
		payload.DoResolve = rule.DoResolve.ValueString()
	}
	if !rule.SendSpoeGroup.IsNull() && !rule.SendSpoeGroup.IsUnknown() {
		payload.SendSpoeGroup = rule.SendSpoeGroup.ValueString()
	}
	if !rule.ReplaceHeader.IsNull() && !rule.ReplaceHeader.IsUnknown() {
		payload.ReplaceHeader = rule.ReplaceHeader.ValueString()
	}
	if !rule.ReplacePath.IsNull() && !rule.ReplacePath.IsUnknown() {
		payload.ReplacePath = rule.ReplacePath.ValueString()
	}
	if !rule.ReplacePathq.IsNull() && !rule.ReplacePathq.IsUnknown() {
		payload.ReplacePathq = rule.ReplacePathq.ValueString()
	}
	if !rule.ReplaceUri.IsNull() && !rule.ReplaceUri.IsUnknown() {
		payload.ReplaceUri = rule.ReplaceUri.ValueString()
	}
	if !rule.ReplaceValue.IsNull() && !rule.ReplaceValue.IsUnknown() {
		payload.ReplaceValue = rule.ReplaceValue.ValueString()
	}
	if !rule.AddHeader.IsNull() && !rule.AddHeader.IsUnknown() {
		payload.AddHeader = rule.AddHeader.ValueString()
	}
	if !rule.DelHeader.IsNull() && !rule.DelHeader.IsUnknown() {
		payload.DelHeader = rule.DelHeader.ValueString()
	}
	if !rule.AddAcl.IsNull() && !rule.AddAcl.IsUnknown() {
		payload.AddAcl = rule.AddAcl.ValueString()
	}
	if !rule.DelAcl.IsNull() && !rule.DelAcl.IsUnknown() {
		payload.DelAcl = rule.DelAcl.ValueString()
	}
	if !rule.SetMap.IsNull() && !rule.SetMap.IsUnknown() {
		payload.SetMap = rule.SetMap.ValueString()
	}
	if !rule.DelMap.IsNull() && !rule.DelMap.IsUnknown() {
		payload.DelMap = rule.DelMap.ValueString()
	}
	if !rule.CacheUse.IsNull() && !rule.CacheUse.IsUnknown() {
		payload.CacheUse = rule.CacheUse.ValueString()
	}
	if !rule.Capture.IsNull() && !rule.Capture.IsUnknown() {
		payload.Capture = rule.Capture.ValueString()
	}
	if !rule.Auth.IsNull() && !rule.Auth.IsUnknown() {
		payload.Auth = rule.Auth.ValueString()
	}
	if !rule.Allow.IsNull() && !rule.Allow.IsUnknown() {
		payload.Allow = rule.Allow.ValueString()
	}
	if !rule.Deny.IsNull() && !rule.Deny.IsUnknown() {
		payload.Deny = rule.Deny.ValueString()
	}
	if !rule.Return.IsNull() && !rule.Return.IsUnknown() {
		payload.Return = rule.Return.ValueString()
	}
	if !rule.Reject.IsNull() && !rule.Reject.IsUnknown() {
		payload.Reject = rule.Reject.ValueString()
	}
	if !rule.Pause.IsNull() && !rule.Pause.IsUnknown() {
		payload.Pause = rule.Pause.ValueString()
	}
	if !rule.NormalizeUri.IsNull() && !rule.NormalizeUri.IsUnknown() {
		payload.NormalizeUri = rule.NormalizeUri.ValueString()
	}
	if !rule.SetMethod.IsNull() && !rule.SetMethod.IsUnknown() {
		payload.SetMethod = rule.SetMethod.ValueString()
	}
	if !rule.SetQuery.IsNull() && !rule.SetQuery.IsUnknown() {
		payload.SetQuery = rule.SetQuery.ValueString()
	}
	if !rule.SetUri.IsNull() && !rule.SetUri.IsUnknown() {
		payload.SetUri = rule.SetUri.ValueString()
	}
	if !rule.SetLogLevel.IsNull() && !rule.SetLogLevel.IsUnknown() {
		payload.SetLogLevel = rule.SetLogLevel.ValueString()
	}
	if !rule.SetBandwidthLimit.IsNull() && !rule.SetBandwidthLimit.IsUnknown() {
		payload.SetBandwidthLimit = rule.SetBandwidthLimit.ValueString()
	}
	if !rule.RstTtl.IsNull() && !rule.RstTtl.IsUnknown() {
		payload.RstTtl = rule.RstTtl.ValueInt64()
	}

	// Handle return headers if present
	if rule.ReturnHdrs != nil && len(rule.ReturnHdrs) > 0 {
		var returnHdrs []ReturnHdr
		for _, hdr := range rule.ReturnHdrs {
			returnHdrs = append(returnHdrs, ReturnHdr{
				Name: hdr.Name.ValueString(),
				Fmt:  hdr.Fmt.ValueString(),
			})
		}
		payload.ReturnHdrs = returnHdrs
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
