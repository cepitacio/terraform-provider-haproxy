package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &haproxyStackResource{}
)

// NewHaproxyStackResource is a helper function to simplify the provider implementation.
func NewHaproxyStackResource() resource.Resource {
	return &haproxyStackResource{}
}

// haproxyStackResource is the resource implementation.
type haproxyStackResource struct {
	stackManager *StackManager
	apiVersion   string
}

// haproxyStackResourceModel maps the resource schema data.
type haproxyStackResourceModel struct {
	Name     types.String                  `tfsdk:"name"`
	Backend  *haproxyBackendModel          `tfsdk:"backend"`
	Servers  map[string]haproxyServerModel `tfsdk:"servers"` // Multiple servers
	Frontend *haproxyFrontendModel         `tfsdk:"frontend"`
	Acls     []haproxyAclModel             `tfsdk:"acls"`
}

// haproxyBackendModel maps the backend block schema data.
type haproxyBackendModel struct {
	Name               types.String                  `tfsdk:"name"`
	Mode               types.String                  `tfsdk:"mode"`
	AdvCheck           types.String                  `tfsdk:"adv_check"`
	HttpConnectionMode types.String                  `tfsdk:"http_connection_mode"`
	ServerTimeout      types.Int64                   `tfsdk:"server_timeout"`
	CheckTimeout       types.Int64                   `tfsdk:"check_timeout"`
	ConnectTimeout     types.Int64                   `tfsdk:"connect_timeout"`
	QueueTimeout       types.Int64                   `tfsdk:"queue_timeout"`
	TunnelTimeout      types.Int64                   `tfsdk:"tunnel_timeout"`
	TarpitTimeout      types.Int64                   `tfsdk:"tarpit_timeout"`
	Checkcache         types.String                  `tfsdk:"checkcache"`
	Retries            types.Int64                   `tfsdk:"retries"`
	Balance            []haproxyBalanceModel         `tfsdk:"balance"`
	HttpchkParams      []haproxyHttpchkParamsModel   `tfsdk:"httpchk_params"`
	Forwardfor         []haproxyForwardforModel      `tfsdk:"forwardfor"`
	Httpcheck          []haproxyHttpcheckModel       `tfsdk:"httpcheck"`
	TcpCheck           []haproxyTcpCheckModel        `tfsdk:"tcp_check"`
	Acls               []haproxyAclModel             `tfsdk:"acls"`
	HttpRequestRule    []haproxyHttpRequestRuleModel `tfsdk:"http_request_rule"`

	HttpResponseRule []haproxyHttpResponseRuleModel `tfsdk:"http_response_rule"`
	TcpRequestRule   []haproxyTcpRequestRuleModel   `tfsdk:"tcp_request_rule"`
	TcpResponseRule  []haproxyTcpResponseRuleModel  `tfsdk:"tcp_response_rule"`
	DefaultServer    *haproxyDefaultServerModel     `tfsdk:"default_server"`
	StickTable       *haproxyStickTableModel        `tfsdk:"stick_table"`
	StickRule        []haproxyStickRuleModel        `tfsdk:"stick_rule"`
	StatsOptions     []haproxyStatsOptionsModel     `tfsdk:"stats_options"`
}

// haproxyDefaultServerModel maps the default_server block schema data.
type haproxyDefaultServerModel struct {
	Ssl            types.String `tfsdk:"ssl"`
	SslCafile      types.String `tfsdk:"ssl_cafile"`
	SslCertificate types.String `tfsdk:"ssl_certificate"`
	SslMaxVer      types.String `tfsdk:"ssl_max_ver"`
	SslMinVer      types.String `tfsdk:"ssl_min_ver"`
	SslReuse       types.String `tfsdk:"ssl_reuse"`
	Ciphers        types.String `tfsdk:"ciphers"`
	Ciphersuites   types.String `tfsdk:"ciphersuites"`
	Verify         types.String `tfsdk:"verify"`
	Sslv3          types.String `tfsdk:"sslv3"`
	Tlsv10         types.String `tfsdk:"tlsv10"`
	Tlsv11         types.String `tfsdk:"tlsv11"`
	Tlsv12         types.String `tfsdk:"tlsv12"`
	Tlsv13         types.String `tfsdk:"tlsv13"`
	NoSslv3        types.String `tfsdk:"no_sslv3"`
	NoTlsv10       types.String `tfsdk:"no_tlsv10"`
	NoTlsv11       types.String `tfsdk:"no_tlsv11"`
	NoTlsv12       types.String `tfsdk:"no_tlsv12"`
	NoTlsv13       types.String `tfsdk:"no_tlsv13"`
	ForceSslv3     types.String `tfsdk:"force_sslv3"`
	ForceTlsv10    types.String `tfsdk:"force_tlsv10"`
	ForceTlsv11    types.String `tfsdk:"force_tlsv11"`
	ForceTlsv12    types.String `tfsdk:"force_tlsv12"`
	ForceTlsv13    types.String `tfsdk:"force_tlsv13"`
	ForceStrictSni types.String `tfsdk:"force_strict_sni"`
}

// haproxyServerModel maps the server block schema data.
type haproxyServerModel struct {
	// Note: Name is now the map key, not a field
	Address   types.String `tfsdk:"address"`
	Port      types.Int64  `tfsdk:"port"`
	Check     types.String `tfsdk:"check"`
	Backup    types.String `tfsdk:"backup"`
	Maxconn   types.Int64  `tfsdk:"maxconn"`
	Weight    types.Int64  `tfsdk:"weight"`
	Rise      types.Int64  `tfsdk:"rise"`
	Fall      types.Int64  `tfsdk:"fall"`
	Inter     types.Int64  `tfsdk:"inter"`
	Fastinter types.Int64  `tfsdk:"fastinter"`
	Downinter types.Int64  `tfsdk:"downinter"`
	Ssl       types.String `tfsdk:"ssl"`
	Verify    types.String `tfsdk:"verify"`
	Cookie    types.String `tfsdk:"cookie"`

	// SSL/TLS Protocol Control (v3 fields)
	Sslv3  types.String `tfsdk:"sslv3"`
	Tlsv10 types.String `tfsdk:"tlsv10"`
	Tlsv11 types.String `tfsdk:"tlsv11"`
	Tlsv12 types.String `tfsdk:"tlsv12"`
	Tlsv13 types.String `tfsdk:"tlsv13"`
	// SSL/TLS Protocol Control (deprecated v2 fields)
	NoSslv3        types.String `tfsdk:"no_sslv3"`
	NoTlsv10       types.String `tfsdk:"no_tlsv10"`
	NoTlsv11       types.String `tfsdk:"no_tlsv11"`
	NoTlsv12       types.String `tfsdk:"no_tlsv12"`
	NoTlsv13       types.String `tfsdk:"no_tlsv13"`
	ForceSslv3     types.String `tfsdk:"force_sslv3"`
	ForceTlsv10    types.String `tfsdk:"force_tlsv10"`
	ForceTlsv11    types.String `tfsdk:"force_tlsv11"`
	ForceTlsv12    types.String `tfsdk:"force_tlsv12"`
	ForceTlsv13    types.String `tfsdk:"force_tlsv13"`
	ForceStrictSni types.String `tfsdk:"force_strict_sni"`
}

// haproxyFrontendModel maps the frontend block schema data.
type haproxyFrontendModel struct {
	Name             types.String                  `tfsdk:"name"`
	Mode             types.String                  `tfsdk:"mode"`
	DefaultBackend   types.String                  `tfsdk:"default_backend"`
	Maxconn          types.Int64                   `tfsdk:"maxconn"`
	Backlog          types.Int64                   `tfsdk:"backlog"`
	Ssl              types.Bool                    `tfsdk:"ssl"`
	SslCertificate   types.String                  `tfsdk:"ssl_certificate"`
	SslCafile        types.String                  `tfsdk:"ssl_cafile"`
	SslMaxVer        types.String                  `tfsdk:"ssl_max_ver"`
	SslMinVer        types.String                  `tfsdk:"ssl_min_ver"`
	Ciphers          types.String                  `tfsdk:"ciphers"`
	Ciphersuites     types.String                  `tfsdk:"ciphersuites"`
	Verify           types.String                  `tfsdk:"verify"`
	AcceptProxy      types.Bool                    `tfsdk:"accept_proxy"`
	DeferAccept      types.Bool                    `tfsdk:"defer_accept"`
	TcpUserTimeout   types.Int64                   `tfsdk:"tcp_user_timeout"`
	Tfo              types.Bool                    `tfsdk:"tfo"`
	V4v6             types.Bool                    `tfsdk:"v4v6"`
	V6only           types.Bool                    `tfsdk:"v6only"`
	Binds            map[string]haproxyBindModel   `tfsdk:"binds"`
	Acls             []haproxyAclModel             `tfsdk:"acls"`
	HttpRequestRules []haproxyHttpRequestRuleModel `tfsdk:"http_request_rules"`
	StatsOptions     []haproxyStatsOptionsModel    `tfsdk:"stats_options"`
}

// haproxyBalanceModel maps the balance block schema data.
type haproxyBalanceModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	UrlParam  types.String `tfsdk:"url_param"`
}

// haproxyHttpchkParamsModel maps the httpchk_params block schema data.
type haproxyHttpchkParamsModel struct {
	Method  types.String `tfsdk:"method"`
	Uri     types.String `tfsdk:"uri"`
	Version types.String `tfsdk:"version"`
}

// haproxyForwardforModel maps the forwardfor block schema data.
type haproxyForwardforModel struct {
	Enabled types.String `tfsdk:"enabled"`
}

// haproxyHttpcheckModel maps the httpcheck block schema data.
type haproxyHttpcheckModel struct {
	Index           types.Int64  `tfsdk:"index"`
	Type            types.String `tfsdk:"type"`
	Method          types.String `tfsdk:"method"`
	Uri             types.String `tfsdk:"uri"`
	Version         types.String `tfsdk:"version"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	Match           types.String `tfsdk:"match"`
	Pattern         types.String `tfsdk:"pattern"`
	Addr            types.String `tfsdk:"addr"`
	Port            types.Int64  `tfsdk:"port"`
	ExclamationMark types.String `tfsdk:"exclamation_mark"`
	LogLevel        types.String `tfsdk:"log_level"`
	SendProxy       types.String `tfsdk:"send_proxy"`
	ViaSocks4       types.String `tfsdk:"via_socks4"`
	CheckComment    types.String `tfsdk:"check_comment"`
}

// haproxyTcpCheckModel maps the tcp_check block schema data.
type haproxyTcpCheckModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// haproxyReturnHdrModel maps the return_hdrs block schema data.
type haproxyReturnHdrModel struct {
	Name types.String `tfsdk:"name"`
	Fmt  types.String `tfsdk:"fmt"`
}

// haproxyHttpResponseRuleModel maps the http_response_rule block schema data.
type haproxyHttpResponseRuleModel struct {
	Index     types.Int64  `tfsdk:"index"`
	Type      types.String `tfsdk:"type"`
	Cond      types.String `tfsdk:"cond"`
	CondTest  types.String `tfsdk:"cond_test"`
	HdrName   types.String `tfsdk:"hdr_name"`
	HdrFormat types.String `tfsdk:"hdr_format"`
}

// haproxyTcpRequestRuleModel maps the tcp_request_rule block schema data.
type haproxyTcpRequestRuleModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// haproxyTcpResponseRuleModel maps the tcp_response_rule block schema data.
type haproxyTcpResponseRuleModel struct {
	Index    types.Int64  `tfsdk:"index"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Cond     types.String `tfsdk:"cond"`
	CondTest types.String `tfsdk:"cond_test"`
}

// haproxyStickTableModel maps the stick_table block schema data.
type haproxyStickTableModel struct {
	Type    types.String `tfsdk:"type"`
	Size    types.Int64  `tfsdk:"size"`
	Expire  types.Int64  `tfsdk:"expire"`
	Nopurge types.Bool   `tfsdk:"nopurge"`
	Peers   types.String `tfsdk:"peers"`
}

// haproxyStickRuleModel maps the stick_rule block schema data.
type haproxyStickRuleModel struct {
	Index   types.Int64  `tfsdk:"index"`
	Type    types.String `tfsdk:"type"`
	Table   types.String `tfsdk:"table"`
	Pattern types.String `tfsdk:"pattern"`
}

// haproxyStatsOptionsModel maps the stats_options block schema data.
type haproxyStatsOptionsModel struct {
	StatsEnable types.Bool   `tfsdk:"stats_enable"`
	StatsUri    types.String `tfsdk:"stats_uri"`
	StatsRealm  types.String `tfsdk:"stats_realm"`
	StatsAuth   types.String `tfsdk:"stats_auth"`
}

// Metadata returns the resource type name.
func (r *haproxyStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Schema defines the schema for the resource.
func (r *haproxyStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Use default v3 if apiVersion is not set (latest version)
	apiVersion := r.apiVersion
	if apiVersion == "" {
		apiVersion = "v3"
	}

	schemaBuilder := NewVersionAwareSchemaBuilder(apiVersion)

	resp.Schema = schema.Schema{
		Description: "Manages a complete HAProxy stack including backend, server, frontend, and ACLs.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the HAProxy stack.",
			},
			"servers": GetServersSchema(schemaBuilder), // Multiple servers
		},
		Blocks: map[string]schema.Block{
			"backend":  GetBackendSchema(schemaBuilder),
			"frontend": GetFrontendSchema(schemaBuilder),
			"acls":     GetACLSchema(),
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *haproxyStackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developer.", req.ProviderData),
		)
		return
	}

	// Initialize the stack manager with all required components
	aclManager := NewACLManager(providerData.Client)
	frontendManager := NewFrontendManager(providerData.Client)
	backendManager := NewBackendManager(providerData.Client)
	r.stackManager = NewStackManager(providerData.Client, aclManager, frontendManager, backendManager)

	// Store the API version for schema generation
	r.apiVersion = providerData.APIVersion
}

// Create resource.
func (r *haproxyStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Validate configuration based on API version before creating
	if err := r.validateConfigForAPIVersion(ctx, req, resp); err != nil {
		return // Validation failed, don't proceed with creation
	}

	if err := r.stackManager.Create(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error creating HAProxy stack", err.Error())
	}
}

// Read resource.
func (r *haproxyStackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if err := r.stackManager.Read(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy stack", err.Error())
	}
}

// Update resource.
func (r *haproxyStackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if err := r.stackManager.Update(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error updating HAProxy stack", err.Error())
	}
}

// Delete resource.
func (r *haproxyStackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.stackManager.Delete(ctx, req, resp); err != nil {
		resp.Diagnostics.AddError("Error deleting HAProxy stack", err.Error())
	}
}

// validateConfigForAPIVersion validates the configuration based on API version
func (r *haproxyStackResource) validateConfigForAPIVersion(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) error {
	// Get the API version from the provider configuration
	apiVersion := r.apiVersion
	if apiVersion == "" {
		apiVersion = "v3" // Default to v3
	}

	// Get the configuration data
	var config haproxyStackResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return fmt.Errorf("failed to get configuration")
	}

	// Validate based on API version
	if apiVersion == "v2" {
		// Check for v3 fields that are not supported in v2
		if config.Backend != nil && config.Backend.DefaultServer != nil {
			validateDefaultServerV2ForCreate(ctx, &resp.Diagnostics, config.Backend.DefaultServer, "backend.default_server")
		}

		if len(config.Servers) > 0 {
			for i, server := range config.Servers {
				validateServerV2ForCreate(ctx, &resp.Diagnostics, server, fmt.Sprintf("servers[%d]", i))
			}
		}

		if config.Frontend != nil {
			for bindName, bind := range config.Frontend.Binds {
				validateBindV2ForCreate(ctx, &resp.Diagnostics, bind, fmt.Sprintf("frontend.binds[%s]", bindName))
			}
		}
	} else if apiVersion == "v3" {
		// Check for v2 fields that are deprecated in v3
		if config.Backend != nil && config.Backend.DefaultServer != nil {
			validateDefaultServerV3ForCreate(ctx, &resp.Diagnostics, config.Backend.DefaultServer, "backend.default_server")
		}

		if len(config.Servers) > 0 {
			for i, server := range config.Servers {
				validateServerV3ForCreate(ctx, &resp.Diagnostics, server, fmt.Sprintf("servers[%d]", i))
			}
		}

		if config.Frontend != nil {
			for bindName, bind := range config.Frontend.Binds {
				validateBindV3ForCreate(ctx, &resp.Diagnostics, bind, fmt.Sprintf("frontend.binds[%s]", bindName))
			}
		}
	}

	// Check if validation produced any errors
	if resp.Diagnostics.HasError() {
		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

// Create-specific validation functions that work with diag.Diagnostics
func validateDefaultServerV2ForCreate(ctx context.Context, diags *diag.Diagnostics, defaultServer *haproxyDefaultServerModel, pathPrefix string) {
	if !defaultServer.Sslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("sslv3"),
			"Unsupported field in v2",
			"Field 'sslv3' is not supported in Data Plane API v2. Use 'no_sslv3' or 'force_sslv3' instead.",
		)
	}
	if !defaultServer.Tlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv10"),
			"Unsupported field in v2",
			"Field 'tlsv10' is not supported in Data Plane API v2. Use 'no_tlsv10' or 'force_tlsv10' instead.",
		)
	}
	if !defaultServer.Tlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv11"),
			"Unsupported field in v2",
			"Field 'tlsv11' is not supported in Data Plane API v2. Use 'no_tlsv11' or 'force_tlsv11' instead.",
		)
	}
	if !defaultServer.Tlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv12"),
			"Unsupported field in v2",
			"Field 'tlsv12' is not supported in Data Plane API v2. Use 'no_tlsv12' or 'force_tlsv12' instead.",
		)
	}
	if !defaultServer.Tlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv13"),
			"Unsupported field in v2",
			"Field 'tlsv13' is not supported in Data Plane API v2. Use 'no_tlsv13' or 'force_tlsv13' instead.",
		)
	}
}

// validateServerV2ForCreate validates that v3 fields are not used in v2 mode
func validateServerV2ForCreate(ctx context.Context, diags *diag.Diagnostics, server haproxyServerModel, pathPrefix string) {
	if !server.Sslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("sslv3"),
			"Unsupported field in v2",
			"Field 'sslv3' is not supported in Data Plane API v2. Use 'no_sslv3' or 'force_sslv3' instead.",
		)
	}
	if !server.Tlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv10"),
			"Unsupported field in v2",
			"Field 'tlsv10' is not supported in Data Plane API v2. Use 'no_tlsv10' or 'force_tlsv10' instead.",
		)
	}
	if !server.Tlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv11"),
			"Unsupported field in v2",
			"Field 'tlsv11' is not supported in Data Plane API v2. Use 'no_tlsv11' or 'force_tlsv11' instead.",
		)
	}
	if !server.Tlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv12"),
			"Unsupported field in v2",
			"Field 'tlsv12' is not supported in Data Plane API v2. Use 'no_tlsv12' or 'force_tlsv12' instead.",
		)
	}
	if !server.Tlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv13"),
			"Unsupported field in v2",
			"Field 'tlsv13' is not supported in Data Plane API v2. Use 'no_tlsv13' or 'force_tlsv13' instead.",
		)
	}
}

// validateBindV2ForCreate validates that v3 fields are not used in v2 mode
func validateBindV2ForCreate(ctx context.Context, diags *diag.Diagnostics, bind haproxyBindModel, pathPrefix string) {
	if !bind.Sslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("sslv3"),
			"Unsupported field in v2",
			"Field 'sslv3' is not supported in Data Plane API v2. Use 'force_sslv3' instead.",
		)
	}
	if !bind.Tlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv10"),
			"Unsupported field in v2",
			"Field 'tlsv10' is not supported in Data Plane API v2. Use 'force_tlsv10' instead.",
		)
	}
	if !bind.Tlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv11"),
			"Unsupported field in v2",
			"Field 'tlsv11' is not supported in Data Plane API v2. Use 'force_tlsv11' instead.",
		)
	}
	if !bind.Tlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv12"),
			"Unsupported field in v2",
			"Field 'tlsv12' is not supported in Data Plane API v2. Use 'force_tlsv12' instead.",
		)
	}
	if !bind.Tlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("tlsv13"),
			"Unsupported field in v2",
			"Field 'tlsv13' is not supported in Data Plane API v2. Use 'force_tlsv13' instead.",
		)
	}
}

// validateDefaultServerV3ForCreate validates that deprecated v2 fields are not used in v3 mode
func validateDefaultServerV3ForCreate(ctx context.Context, diags *diag.Diagnostics, defaultServer *haproxyDefaultServerModel, pathPrefix string) {
	if !defaultServer.NoSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_sslv3"),
			"Invalid field in v3 default-server",
			"Field 'no_sslv3' is not accepted in default-server sections in Data Plane API v3. Use 'sslv3' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv10"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv10' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv10' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv11"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv11' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv11' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv12"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv12' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv12' in individual server sections instead.",
		)
	}
	if !defaultServer.NoTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv13"),
			"Invalid field in v3 default-server",
			"Field 'no_tlsv13' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv13' in individual server sections instead.",
		)
	}
	if !defaultServer.ForceSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_sslv3"),
			"Invalid field in v3 default-server",
			"Field 'force_sslv3' is not accepted in default-server sections in Data Plane API v3. Use 'sslv3' in individual server sections instead.",
		)
	}
	if !defaultServer.ForceTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv10"),
			"Invalid field in v3 default-server",
			"Field 'force_tlsv10' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv10' in individual server sections instead.",
		)
	}
	if !defaultServer.ForceTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv11"),
			"Invalid field in v3 default-server",
			"Field 'force_tlsv11' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv11' in individual server sections instead.",
		)
	}
	if !defaultServer.ForceTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv12"),
			"Invalid field in v3 default-server",
			"Field 'force_tlsv12' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv12' in individual server sections instead.",
		)
	}
	if !defaultServer.ForceTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv13"),
			"Invalid field in v3 default-server",
			"Field 'force_tlsv13' is not accepted in default-server sections in Data Plane API v3. Use 'tlsv13' in individual server sections instead.",
		)
	}
}

// validateServerV3ForCreate validates that deprecated v2 fields are not used in v3 mode
func validateServerV3ForCreate(ctx context.Context, diags *diag.Diagnostics, server haproxyServerModel, pathPrefix string) {
	if !server.NoSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_sslv3"),
			"Deprecated field in v3",
			"Field 'no_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !server.NoTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv10"),
			"Deprecated field in v3",
			"Field 'no_tlsv10' is deprecated in Data Plane API v3. Use 'tlsv10' instead.",
		)
	}
	if !server.NoTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv11"),
			"Deprecated field in v3",
			"Field 'no_tlsv11' is deprecated in Data Plane API v3. Use 'tlsv11' instead.",
		)
	}
	if !server.NoTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv12"),
			"Deprecated field in v3",
			"Field 'no_tlsv12' is deprecated in Data Plane API v3. Use 'tlsv12' instead.",
		)
	}
	if !server.NoTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_tlsv13"),
			"Deprecated field in v3",
			"Field 'no_tlsv13' is deprecated in Data Plane API v3. Use 'tlsv13' instead.",
		)
	}
	if !server.ForceSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_sslv3"),
			"Deprecated field in v3",
			"Field 'force_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !server.ForceTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv10"),
			"Deprecated field in v3",
			"Field 'force_tlsv10' is deprecated in Data Plane API v3. Use 'tlsv10' instead.",
		)
	}
	if !server.ForceTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv11"),
			"Deprecated field in v3",
			"Field 'force_tlsv11' is deprecated in Data Plane API v3. Use 'tlsv11' instead.",
		)
	}
	if !server.ForceTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv12"),
			"Deprecated field in v3",
			"Field 'force_tlsv12' is deprecated in Data Plane API v3. Use 'tlsv12' instead.",
		)
	}
	if !server.ForceTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv13"),
			"Deprecated field in v3",
			"Field 'force_tlsv13' is deprecated in Data Plane API v3. Use 'tlsv13' instead.",
		)
	}
}

// validateBindV3ForCreate validates that deprecated v2 fields are not used in v3 mode
func validateBindV3ForCreate(ctx context.Context, diags *diag.Diagnostics, bind haproxyBindModel, pathPrefix string) {
	if !bind.NoSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("no_sslv3"),
			"Deprecated field in v3",
			"Field 'no_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !bind.ForceSslv3.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_sslv3"),
			"Deprecated field in v3",
			"Field 'force_sslv3' is deprecated in Data Plane API v3. Use 'sslv3' instead.",
		)
	}
	if !bind.ForceTlsv10.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv10"),
			"Deprecated field in v3",
			"Field 'force_tlsv10' is deprecated in Data Plane API v3. Use 'tlsv10' instead.",
		)
	}
	if !bind.ForceTlsv11.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv11"),
			"Deprecated field in v3",
			"Field 'force_tlsv11' is deprecated in Data Plane API v3. Use 'tlsv11' instead.",
		)
	}
	if !bind.ForceTlsv12.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv12"),
			"Deprecated field in v3",
			"Field 'force_tlsv12' is deprecated in Data Plane API v3. Use 'tlsv12' instead.",
		)
	}
	if !bind.ForceTlsv13.IsNull() {
		diags.AddAttributeError(
			path.Root(pathPrefix).AtName("force_tlsv13"),
			"Deprecated field in v3",
			"Field 'force_tlsv13' is deprecated in Data Plane API v3. Use 'tlsv13' instead.",
		)
	}
}
