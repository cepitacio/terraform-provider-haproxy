package haproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetBindSchema returns the schema for the bind block
func GetBindSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		Optional:    true,
		Description: "Bind configuration blocks for frontend listening addresses and ports.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"address": schema.StringAttribute{
					Required:    true,
					Description: "The bind address (e.g., 0.0.0.0, ::, or specific IP).",
				},
				"port": schema.Int64Attribute{
					Required:    true,
					Description: "The bind port (1-65535).",
				},
				"port_range_end": schema.Int64Attribute{
					Optional:    true,
					Description: "The end of the port range (1-65535).",
				},
				"transparent": schema.BoolAttribute{
					Optional:    true,
					Description: "Whether the bind is transparent.",
				},
				"mode": schema.StringAttribute{
					Optional:    true,
					Description: "The mode of the bind (http, tcp).",
				},
				"maxconn": schema.Int64Attribute{
					Optional:    true,
					Description: "Maximum number of connections for this bind.",
				},
				"ssl": schema.BoolAttribute{
					Optional:    true,
					Description: "Enable SSL/TLS for this bind.",
				},
				"ssl_cafile": schema.StringAttribute{
					Optional:    true,
					Description: "SSL CA file for the bind.",
				},
				"ssl_certificate": schema.StringAttribute{
					Optional:    true,
					Description: "SSL certificate for the bind.",
				},
				"ssl_max_ver": schema.StringAttribute{
					Optional:    true,
					Description: "SSL maximum version (SSLv3, TLSv1.0, TLSv1.1, TLSv1.2, TLSv1.3).",
				},
				"ssl_min_ver": schema.StringAttribute{
					Optional:    true,
					Description: "SSL minimum version (SSLv3, TLSv1.0, TLSv1.1, TLSv1.2, TLSv1.3).",
				},
				"ciphers": schema.StringAttribute{
					Optional:    true,
					Description: "Ciphers for the bind.",
				},
				"ciphersuites": schema.StringAttribute{
					Optional:    true,
					Description: "Cipher suites for the bind.",
				},
				"verify": schema.StringAttribute{
					Optional:    true,
					Description: "SSL verification (none, optional, required).",
				},
				"accept_proxy": schema.BoolAttribute{
					Optional:    true,
					Description: "Accept PROXY protocol.",
				},
				"allow_0rtt": schema.BoolAttribute{
					Optional:    true,
					Description: "Allow 0-RTT for TLS 1.3.",
				},
				"alpn": schema.StringAttribute{
					Optional:    true,
					Description: "ALPN protocols for the bind.",
				},
				"backlog": schema.StringAttribute{
					Optional:    true,
					Description: "Backlog for the bind.",
				},
				"defer_accept": schema.BoolAttribute{
					Optional:    true,
					Description: "Defer accept for the bind.",
				},
				"generate_certificates": schema.BoolAttribute{
					Optional:    true,
					Description: "Generate certificates for the bind.",
				},
				"gid": schema.Int64Attribute{
					Optional:    true,
					Description: "Group ID for the bind.",
				},
				"group": schema.StringAttribute{
					Optional:    true,
					Description: "Group name for the bind.",
				},
				"id": schema.StringAttribute{
					Optional:    true,
					Description: "ID for the bind.",
				},
				"interface": schema.StringAttribute{
					Optional:    true,
					Description: "Interface for the bind.",
				},
				"level": schema.StringAttribute{
					Optional:    true,
					Description: "Level for the bind (user, operator, admin).",
				},
				"namespace": schema.StringAttribute{
					Optional:    true,
					Description: "Namespace for the bind.",
				},
				"nice": schema.Int64Attribute{
					Optional:    true,
					Description: "Nice value for the bind.",
				},
				"no_ca_names": schema.BoolAttribute{
					Optional:    true,
					Description: "No CA names for the bind.",
				},
				"npn": schema.StringAttribute{
					Optional:    true,
					Description: "NPN protocols for the bind.",
				},
				"prefer_client_ciphers": schema.BoolAttribute{
					Optional:    true,
					Description: "Prefer client ciphers for the bind.",
				},
				"process": schema.StringAttribute{
					Optional:    true,
					Description: "Process/thread binding (v2 only). Use 'thread' for thread binding.",
				},
				"proto": schema.StringAttribute{
					Optional:    true,
					Description: "Protocol for the bind.",
				},
				"severity_output": schema.StringAttribute{
					Optional:    true,
					Description: "Severity output for the bind (none, number, string).",
				},
				"strict_sni": schema.BoolAttribute{
					Optional:    true,
					Description: "Strict SNI for the bind.",
				},
				"tcp_user_timeout": schema.Int64Attribute{
					Optional:    true,
					Description: "TCP user timeout for the bind.",
				},
				"tfo": schema.BoolAttribute{
					Optional:    true,
					Description: "TCP Fast Open for the bind.",
				},
				"tls_ticket_keys": schema.StringAttribute{
					Optional:    true,
					Description: "TLS ticket keys for the bind.",
				},
				"uid": schema.StringAttribute{
					Optional:    true,
					Description: "User ID for the bind.",
				},
				"user": schema.StringAttribute{
					Optional:    true,
					Description: "User name for the bind.",
				},
				"v4v6": schema.BoolAttribute{
					Optional:    true,
					Description: "IPv4/IPv6 for the bind.",
				},
				"v6only": schema.BoolAttribute{
					Optional:    true,
					Description: "IPv6 only for the bind.",
				},
				// v3 fields (non-deprecated)
				"sslv3": schema.BoolAttribute{
					Optional:    true,
					Description: "SSLv3 support for the bind (Data Plane API v3 only).",
				},
				"tlsv10": schema.BoolAttribute{
					Optional:    true,
					Description: "TLSv1.0 support for the bind (Data Plane API v3 only).",
				},
				"tlsv11": schema.BoolAttribute{
					Optional:    true,
					Description: "TLSv1.1 support for the bind (Data Plane API v3 only).",
				},
				"tlsv12": schema.BoolAttribute{
					Optional:    true,
					Description: "TLSv1.2 support for the bind (Data Plane API v3 only).",
				},
				"tlsv13": schema.BoolAttribute{
					Optional:    true,
					Description: "TLSv1.3 support for the bind (Data Plane API v3 only).",
				},
				"tls_tickets": schema.StringAttribute{
					Optional:    true,
					Description: "TLS tickets for the bind (enabled, disabled) (Data Plane API v3 only).",
				},
				"force_strict_sni": schema.StringAttribute{
					Optional:    true,
					Description: "Force strict SNI for the bind (enabled, disabled) (Data Plane API v3 only).",
				},
				"no_strict_sni": schema.BoolAttribute{
					Optional:    true,
					Description: "No strict SNI for the bind (Data Plane API v3 only).",
				},
				"guid_prefix": schema.StringAttribute{
					Optional:    true,
					Description: "GUID prefix for the bind (Data Plane API v3 only).",
				},
				"idle_ping": schema.Int64Attribute{
					Optional:    true,
					Description: "Idle ping for the bind (Data Plane API v3 only).",
				},
				"quic_cc_algo": schema.StringAttribute{
					Optional:    true,
					Description: "QUIC congestion control algorithm (cubic, newreno, bbr, nocc) (Data Plane API v3 only).",
				},
				"quic_force_retry": schema.BoolAttribute{
					Optional:    true,
					Description: "QUIC force retry for the bind (Data Plane API v3 only).",
				},
				"quic_socket": schema.StringAttribute{
					Optional:    true,
					Description: "QUIC socket for the bind (connection, listener) (Data Plane API v3 only).",
				},
				"quic_cc_algo_burst_size": schema.Int64Attribute{
					Optional:    true,
					Description: "QUIC CC algo burst size for the bind (Data Plane API v3 only).",
				},
				"quic_cc_algo_max_window": schema.Int64Attribute{
					Optional:    true,
					Description: "QUIC CC algo max window for the bind (Data Plane API v3 only).",
				},
				"metadata": schema.StringAttribute{
					Optional:    true,
					Description: "Metadata for the bind (Data Plane API v3 only).",
				},
				// v2 fields (deprecated in v3)
				"no_sslv3": schema.BoolAttribute{
					Optional:    true,
					Description: "No SSLv3 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"force_sslv3": schema.BoolAttribute{
					Optional:    true,
					Description: "Force SSLv3 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"force_tlsv10": schema.BoolAttribute{
					Optional:    true,
					Description: "Force TLSv1.0 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"force_tlsv11": schema.BoolAttribute{
					Optional:    true,
					Description: "Force TLSv1.1 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"force_tlsv12": schema.BoolAttribute{
					Optional:    true,
					Description: "Force TLSv1.2 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"force_tlsv13": schema.BoolAttribute{
					Optional:    true,
					Description: "Force TLSv1.3 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"no_tlsv10": schema.BoolAttribute{
					Optional:    true,
					Description: "No TLSv1.0 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"no_tlsv11": schema.BoolAttribute{
					Optional:    true,
					Description: "No TLSv1.1 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"no_tlsv12": schema.BoolAttribute{
					Optional:    true,
					Description: "No TLSv1.2 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"no_tlsv13": schema.BoolAttribute{
					Optional:    true,
					Description: "No TLSv1.3 for the bind (Data Plane API v2 only, deprecated in v3).",
				},
				"no_tls_tickets": schema.BoolAttribute{
					Optional:    true,
					Description: "No TLS tickets for the bind (Data Plane API v2 only, deprecated in v3).",
				},
			},
		},
	}
}

// haproxyBindModel maps the bind block schema data.
type haproxyBindModel struct {
	Address              types.String `tfsdk:"address"`
	Port                 types.Int64  `tfsdk:"port"`
	PortRangeEnd         types.Int64  `tfsdk:"port_range_end"`
	Transparent          types.Bool   `tfsdk:"transparent"`
	Mode                 types.String `tfsdk:"mode"`
	Maxconn              types.Int64  `tfsdk:"maxconn"`
	Ssl                  types.Bool   `tfsdk:"ssl"`
	SslCafile            types.String `tfsdk:"ssl_cafile"`
	SslCertificate       types.String `tfsdk:"ssl_certificate"`
	SslMaxVer            types.String `tfsdk:"ssl_max_ver"`
	SslMinVer            types.String `tfsdk:"ssl_min_ver"`
	Ciphers              types.String `tfsdk:"ciphers"`
	Ciphersuites         types.String `tfsdk:"ciphersuites"`
	Verify               types.String `tfsdk:"verify"`
	AcceptProxy          types.Bool   `tfsdk:"accept_proxy"`
	Allow0rtt            types.Bool   `tfsdk:"allow_0rtt"`
	Alpn                 types.String `tfsdk:"alpn"`
	Backlog              types.String `tfsdk:"backlog"`
	DeferAccept          types.Bool   `tfsdk:"defer_accept"`
	GenerateCertificates types.Bool   `tfsdk:"generate_certificates"`
	Gid                  types.Int64  `tfsdk:"gid"`
	Group                types.String `tfsdk:"group"`
	Id                   types.String `tfsdk:"id"`
	Interface            types.String `tfsdk:"interface"`
	Level                types.String `tfsdk:"level"`
	Namespace            types.String `tfsdk:"namespace"`
	Nice                 types.Int64  `tfsdk:"nice"`
	NoCaNames            types.Bool   `tfsdk:"no_ca_names"`
	Npn                  types.String `tfsdk:"npn"`
	PreferClientCiphers  types.Bool   `tfsdk:"prefer_client_ciphers"`
	Process              types.String `tfsdk:"process"`
	Proto                types.String `tfsdk:"proto"`
	SeverityOutput       types.String `tfsdk:"severity_output"`
	StrictSni            types.Bool   `tfsdk:"strict_sni"`
	TcpUserTimeout       types.Int64  `tfsdk:"tcp_user_timeout"`
	Tfo                  types.Bool   `tfsdk:"tfo"`
	TlsTicketKeys        types.String `tfsdk:"tls_ticket_keys"`
	Uid                  types.String `tfsdk:"uid"`
	User                 types.String `tfsdk:"user"`
	V4v6                 types.Bool   `tfsdk:"v4v6"`
	V6only               types.Bool   `tfsdk:"v6only"`
	// v3 fields (non-deprecated)
	Sslv3               types.Bool   `tfsdk:"sslv3"`
	Tlsv10              types.Bool   `tfsdk:"tlsv10"`
	Tlsv11              types.Bool   `tfsdk:"tlsv11"`
	Tlsv12              types.Bool   `tfsdk:"tlsv12"`
	Tlsv13              types.Bool   `tfsdk:"tlsv13"`
	TlsTickets          types.String `tfsdk:"tls_tickets"`
	ForceStrictSni      types.String `tfsdk:"force_strict_sni"`
	NoStrictSni         types.Bool   `tfsdk:"no_strict_sni"`
	GuidPrefix          types.String `tfsdk:"guid_prefix"`
	IdlePing            types.Int64  `tfsdk:"idle_ping"`
	QuicCcAlgo          types.String `tfsdk:"quic_cc_algo"`
	QuicForceRetry      types.Bool   `tfsdk:"quic_force_retry"`
	QuicSocket          types.String `tfsdk:"quic_socket"`
	QuicCcAlgoBurstSize types.Int64  `tfsdk:"quic_cc_algo_burst_size"`
	QuicCcAlgoMaxWindow types.Int64  `tfsdk:"quic_cc_algo_max_window"`
	Metadata            types.String `tfsdk:"metadata"`
	// v2 fields (deprecated in v3)
	NoSslv3      types.Bool `tfsdk:"no_sslv3"`
	ForceSslv3   types.Bool `tfsdk:"force_sslv3"`
	ForceTlsv10  types.Bool `tfsdk:"force_tlsv10"`
	ForceTlsv11  types.Bool `tfsdk:"force_tlsv11"`
	ForceTlsv12  types.Bool `tfsdk:"force_tlsv12"`
	ForceTlsv13  types.Bool `tfsdk:"force_tlsv13"`
	NoTlsv10     types.Bool `tfsdk:"no_tlsv10"`
	NoTlsv11     types.Bool `tfsdk:"no_tlsv11"`
	NoTlsv12     types.Bool `tfsdk:"no_tlsv12"`
	NoTlsv13     types.Bool `tfsdk:"no_tlsv13"`
	NoTlsTickets types.Bool `tfsdk:"no_tls_tickets"`
}

// BindManager handles all bind-related operations
type BindManager struct {
	client *HAProxyClient
}

// NewBindManager creates a new BindManager instance
func CreateBindManager(client *HAProxyClient) *BindManager {
	return &BindManager{
		client: client,
	}
}

// CreateBinds creates binds for a parent resource
func (r *BindManager) CreateBinds(ctx context.Context, parentType string, parentName string, binds map[string]haproxyBindModel) error {
	if len(binds) == 0 {
		return nil
	}

	// Create binds using map iteration
	for bindName, bind := range binds {
		bindPayload := r.convertToBindPayload(bindName, &bind)

		if err := r.client.CreateBind(ctx, parentType, parentName, bindPayload); err != nil {
			return fmt.Errorf("failed to create bind '%s': %w", bindName, err)
		}

		log.Printf("Created bind '%s' for %s %s", bindName, parentType, parentName)
	}

	return nil
}

// CreateBindsInTransaction creates binds using an existing transaction ID
func (r *BindManager) CreateBindsInTransaction(ctx context.Context, transactionID, parentType string, parentName string, binds map[string]haproxyBindModel) error {
	if len(binds) == 0 {
		return nil
	}

	// Create binds using map iteration
	for bindName, bind := range binds {
		bindPayload := r.convertToBindPayload(bindName, &bind)

		// Debug: Log what we're trying to create
		log.Printf("DEBUG: Creating bind '%s' with payload: %+v", bindName, bindPayload)

		// Debug: Log specific problematic fields
		log.Printf("DEBUG: Bind '%s' - Process: '%s', Metadata: '%s', Tlsv12: %t, Tlsv13: %t",
			bindName, bindPayload.Process, bindPayload.Metadata, bindPayload.Tlsv12, bindPayload.Tlsv13)

		// Debug: Log the JSON payload being sent
		jsonPayload, _ := json.Marshal(bindPayload)
		log.Printf("DEBUG: JSON payload being sent to HAProxy: %s", string(jsonPayload))

		resp, err := r.client.CreateBindInTransaction(ctx, transactionID, parentType, parentName, bindPayload)
		if err != nil {
			return fmt.Errorf("failed to create bind '%s': %w", bindName, err)
		}
		defer resp.Body.Close()

		// Check response status code
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create bind '%s': HTTP %d - %s", bindName, resp.StatusCode, string(body))
		}

		log.Printf("Created bind '%s' for %s %s in transaction %s", bindName, parentType, parentName, transactionID)

		// Debug: Log response body to see what HAProxy actually stored
		if resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("DEBUG: HAProxy response for bind '%s': %s", bindName, string(body))
		}
	}

	return nil
}

// ReadBinds reads binds for a parent resource
func (r *BindManager) ReadBinds(ctx context.Context, parentType string, parentName string) ([]BindPayload, error) {
	binds, err := r.client.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		return nil, fmt.Errorf("failed to read binds for %s %s: %w", parentType, parentName, err)
	}

	// Debug: Log what we're reading back
	log.Printf("DEBUG: Read %d binds for %s %s (original order):", len(binds), parentType, parentName)
	for i, bind := range binds {
		log.Printf("DEBUG: Bind %d: %+v", i, bind)
	}

	return binds, nil
}

// UpdateBinds updates binds for a parent resource
func (r *BindManager) UpdateBinds(ctx context.Context, parentType string, parentName string, newBinds map[string]haproxyBindModel) error {
	if len(newBinds) == 0 {
		// Delete all existing binds
		return r.deleteAllBinds(ctx, parentType, parentName)
	}

	// Read existing binds
	existingBinds, err := r.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing binds: %w", err)
	}

	// Process updates with proper handling
	return r.updateBindsWithHandling(ctx, parentType, parentName, existingBinds, newBinds)
}

// DeleteBinds deletes binds for a parent resource
func (r *BindManager) DeleteBinds(ctx context.Context, parentType string, parentName string) error {
	return r.deleteAllBinds(ctx, parentType, parentName)
}

// UpdateBindsInTransaction updates binds using an existing transaction ID
func (r *BindManager) UpdateBindsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, binds map[string]haproxyBindModel) error {
	if len(binds) == 0 {
		// Delete all existing binds
		return r.deleteAllBindsInTransaction(ctx, transactionID, parentType, parentName)
	}

	// Read existing binds
	existingBinds, err := r.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read existing binds: %w", err)
	}

	// Use smart update logic with transaction support
	return r.updateBindsWithHandlingInTransaction(ctx, transactionID, parentType, parentName, existingBinds, binds)
}

// DeleteBindsInTransaction deletes all binds for a parent resource using an existing transaction ID
func (r *BindManager) DeleteBindsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string) error {
	binds, err := r.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read binds for deletion: %w", err)
	}

	// Delete in reverse order to avoid issues
	sort.Slice(binds, func(i, j int) bool {
		return binds[i].Name > binds[j].Name
	})

	for _, bind := range binds {
		log.Printf("Deleting bind '%s' in transaction %s", bind.Name, transactionID)
		_, err := r.client.DeleteBindInTransaction(ctx, transactionID, bind.Name, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete bind '%s': %w", bind.Name, err)
		}
	}

	return nil
}

// Helper methods
// processBindsBlock is no longer needed with map-based binds

func (r *BindManager) convertToBindPayload(bindName string, bind *haproxyBindModel) *BindPayload {
	payload := &BindPayload{
		Name:    bindName,
		Address: bind.Address.ValueString(),
		Port:    &[]int64{bind.Port.ValueInt64()}[0],
	}

	// Get API version to determine which fields to send
	apiVersion := r.client.GetAPIVersion()

	// Debug: Log API version
	log.Printf("DEBUG: Using API version: %s for bind '%s'", apiVersion, bindName)

	// Set optional fields only if they have values
	if !bind.PortRangeEnd.IsNull() && !bind.PortRangeEnd.IsUnknown() {
		payload.PortRangeEnd = &[]int64{bind.PortRangeEnd.ValueInt64()}[0]
	}
	if !bind.Transparent.IsNull() && !bind.Transparent.IsUnknown() {
		payload.Transparent = bind.Transparent.ValueBool()
	}
	if !bind.Mode.IsNull() && !bind.Mode.IsUnknown() {
		payload.Mode = bind.Mode.ValueString()
	}
	if !bind.Maxconn.IsNull() && !bind.Maxconn.IsUnknown() {
		payload.Maxconn = bind.Maxconn.ValueInt64()
	}
	if !bind.Ssl.IsNull() && !bind.Ssl.IsUnknown() {
		payload.Ssl = bind.Ssl.ValueBool()
	}
	if !bind.SslCafile.IsNull() && !bind.SslCafile.IsUnknown() {
		payload.SslCafile = bind.SslCafile.ValueString()
	}
	if !bind.SslCertificate.IsNull() && !bind.SslCertificate.IsUnknown() {
		payload.SslCertificate = bind.SslCertificate.ValueString()
	}
	if !bind.SslMaxVer.IsNull() && !bind.SslMaxVer.IsUnknown() {
		payload.SslMaxVer = bind.SslMaxVer.ValueString()
	}
	if !bind.SslMinVer.IsNull() && !bind.SslMinVer.IsUnknown() {
		payload.SslMinVer = bind.SslMinVer.ValueString()
	}
	if !bind.Ciphers.IsNull() && !bind.Ciphers.IsUnknown() {
		payload.Ciphers = bind.Ciphers.ValueString()
	}
	if !bind.Ciphersuites.IsNull() && !bind.Ciphersuites.IsUnknown() {
		payload.Ciphersuites = bind.Ciphersuites.ValueString()
	}
	if !bind.Verify.IsNull() && !bind.Verify.IsUnknown() {
		payload.Verify = bind.Verify.ValueString()
	}
	if !bind.AcceptProxy.IsNull() && !bind.AcceptProxy.IsUnknown() {
		payload.AcceptProxy = bind.AcceptProxy.ValueBool()
	}
	if !bind.Allow0rtt.IsNull() && !bind.Allow0rtt.IsUnknown() {
		payload.Allow0rtt = bind.Allow0rtt.ValueBool()
	}
	if !bind.Alpn.IsNull() && !bind.Alpn.IsUnknown() {
		payload.Alpn = bind.Alpn.ValueString()
	}
	if !bind.Backlog.IsNull() && !bind.Backlog.IsUnknown() {
		payload.Backlog = bind.Backlog.ValueString()
	}
	if !bind.DeferAccept.IsNull() && !bind.DeferAccept.IsUnknown() {
		payload.DeferAccept = bind.DeferAccept.ValueBool()
	}
	if !bind.GenerateCertificates.IsNull() && !bind.GenerateCertificates.IsUnknown() {
		payload.GenerateCertificates = bind.GenerateCertificates.ValueBool()
	}
	if !bind.Gid.IsNull() && !bind.Gid.IsUnknown() {
		payload.Gid = bind.Gid.ValueInt64()
	}
	if !bind.Group.IsNull() && !bind.Group.IsUnknown() {
		payload.Group = bind.Group.ValueString()
	}
	if !bind.Id.IsNull() && !bind.Id.IsUnknown() {
		payload.Id = bind.Id.ValueString()
	}
	if !bind.Interface.IsNull() && !bind.Interface.IsUnknown() {
		payload.Interface = bind.Interface.ValueString()
	}
	if !bind.Level.IsNull() && !bind.Level.IsUnknown() {
		payload.Level = bind.Level.ValueString()
	}
	if !bind.Namespace.IsNull() && !bind.Namespace.IsUnknown() {
		payload.Namespace = bind.Namespace.ValueString()
	}
	if !bind.Nice.IsNull() && !bind.Nice.IsUnknown() {
		payload.Nice = bind.Nice.ValueInt64()
	}
	if !bind.NoCaNames.IsNull() && !bind.NoCaNames.IsUnknown() {
		payload.NoCaNames = bind.NoCaNames.ValueBool()
	}
	if !bind.Npn.IsNull() && !bind.Npn.IsUnknown() {
		payload.Npn = bind.Npn.ValueString()
	}
	if !bind.PreferClientCiphers.IsNull() && !bind.PreferClientCiphers.IsUnknown() {
		payload.PreferClientCiphers = bind.PreferClientCiphers.ValueBool()
	}
	// Process field - only supported in v2, not v3
	if apiVersion == "v2" && !bind.Process.IsNull() && !bind.Process.IsUnknown() {
		payload.Process = bind.Process.ValueString()
	}
	if !bind.Proto.IsNull() && !bind.Proto.IsUnknown() {
		payload.Proto = bind.Proto.ValueString()
	}
	if !bind.SeverityOutput.IsNull() && !bind.SeverityOutput.IsUnknown() {
		payload.SeverityOutput = bind.SeverityOutput.ValueString()
	}
	if !bind.StrictSni.IsNull() && !bind.StrictSni.IsUnknown() {
		payload.StrictSni = bind.StrictSni.ValueBool()
	}
	if !bind.TcpUserTimeout.IsNull() && !bind.TcpUserTimeout.IsUnknown() {
		payload.TcpUserTimeout = bind.TcpUserTimeout.ValueInt64()
	}
	if !bind.Tfo.IsNull() && !bind.Tfo.IsUnknown() {
		payload.Tfo = bind.Tfo.ValueBool()
	}
	if !bind.TlsTicketKeys.IsNull() && !bind.TlsTicketKeys.IsUnknown() {
		payload.TlsTicketKeys = bind.TlsTicketKeys.ValueString()
	}
	if !bind.Uid.IsNull() && !bind.Uid.IsUnknown() {
		payload.Uid = bind.Uid.ValueString()
	}
	if !bind.User.IsNull() && !bind.User.IsUnknown() {
		payload.User = bind.User.ValueString()
	}
	if !bind.V4v6.IsNull() && !bind.V4v6.IsUnknown() {
		payload.V4v6 = bind.V4v6.ValueBool()
	}
	if !bind.V6only.IsNull() && !bind.V6only.IsUnknown() {
		payload.V6only = bind.V6only.ValueBool()
	}

	// v3 fields (non-deprecated)
	if !bind.Sslv3.IsNull() && !bind.Sslv3.IsUnknown() {
		payload.Sslv3 = bind.Sslv3.ValueBool()
	}
	if !bind.Tlsv10.IsNull() && !bind.Tlsv10.IsUnknown() {
		payload.Tlsv10 = bind.Tlsv10.ValueBool()
	}
	if !bind.Tlsv11.IsNull() && !bind.Tlsv11.IsUnknown() {
		payload.Tlsv11 = bind.Tlsv11.ValueBool()
	}
	// TLS version fields - not supported in either v2 or v3 for binds
	// (HAProxy doesn't store these fields even though API docs show them)
	if !bind.TlsTickets.IsNull() && !bind.TlsTickets.IsUnknown() {
		payload.TlsTickets = bind.TlsTickets.ValueString()
	}
	if !bind.ForceStrictSni.IsNull() && !bind.ForceStrictSni.IsUnknown() {
		payload.ForceStrictSni = bind.ForceStrictSni.ValueString()
	}
	if !bind.NoStrictSni.IsNull() && !bind.NoStrictSni.IsUnknown() {
		payload.NoStrictSni = bind.NoStrictSni.ValueBool()
	}
	if !bind.GuidPrefix.IsNull() && !bind.GuidPrefix.IsUnknown() {
		payload.GuidPrefix = bind.GuidPrefix.ValueString()
	}
	if !bind.IdlePing.IsNull() && !bind.IdlePing.IsUnknown() {
		payload.IdlePing = &[]int64{bind.IdlePing.ValueInt64()}[0]
	}
	if !bind.QuicCcAlgo.IsNull() && !bind.QuicCcAlgo.IsUnknown() {
		payload.QuicCcAlgo = bind.QuicCcAlgo.ValueString()
	}
	if !bind.QuicForceRetry.IsNull() && !bind.QuicForceRetry.IsUnknown() {
		payload.QuicForceRetry = bind.QuicForceRetry.ValueBool()
	}
	if !bind.QuicSocket.IsNull() && !bind.QuicSocket.IsUnknown() {
		payload.QuicSocket = bind.QuicSocket.ValueString()
	}
	if !bind.QuicCcAlgoBurstSize.IsNull() && !bind.QuicCcAlgoBurstSize.IsUnknown() {
		payload.QuicCcAlgoBurstSize = &[]int64{bind.QuicCcAlgoBurstSize.ValueInt64()}[0]
	}
	if !bind.QuicCcAlgoMaxWindow.IsNull() && !bind.QuicCcAlgoMaxWindow.IsUnknown() {
		payload.QuicCcAlgoMaxWindow = &[]int64{bind.QuicCcAlgoMaxWindow.ValueInt64()}[0]
	}
	// Metadata field - not supported in either v2 or v3 for binds
	// (HAProxy doesn't store this field even though API docs show it)

	// v2 fields (deprecated in v3)
	if !bind.NoSslv3.IsNull() && !bind.NoSslv3.IsUnknown() {
		payload.NoSslv3 = bind.NoSslv3.ValueBool()
	}
	if !bind.ForceSslv3.IsNull() && !bind.ForceSslv3.IsUnknown() {
		payload.ForceSslv3 = bind.ForceSslv3.ValueBool()
	}
	if !bind.ForceTlsv10.IsNull() && !bind.ForceTlsv10.IsUnknown() {
		payload.ForceTlsv10 = bind.ForceTlsv10.ValueBool()
	}
	if !bind.ForceTlsv11.IsNull() && !bind.ForceTlsv11.IsUnknown() {
		payload.ForceTlsv11 = bind.ForceTlsv11.ValueBool()
	}
	if !bind.ForceTlsv12.IsNull() && !bind.ForceTlsv12.IsUnknown() {
		payload.ForceTlsv12 = bind.ForceTlsv12.ValueBool()
	}
	if !bind.ForceTlsv13.IsNull() && !bind.ForceTlsv13.IsUnknown() {
		payload.ForceTlsv13 = bind.ForceTlsv13.ValueBool()
	}
	if !bind.NoTlsv10.IsNull() && !bind.NoTlsv10.IsUnknown() {
		payload.NoTlsv10 = bind.NoTlsv10.ValueBool()
	}
	if !bind.NoTlsv11.IsNull() && !bind.NoTlsv11.IsUnknown() {
		payload.NoTlsv11 = bind.NoTlsv11.ValueBool()
	}
	if !bind.NoTlsv12.IsNull() && !bind.NoTlsv12.IsUnknown() {
		payload.NoTlsv12 = bind.NoTlsv12.ValueBool()
	}
	if !bind.NoTlsv13.IsNull() && !bind.NoTlsv13.IsUnknown() {
		payload.NoTlsv13 = bind.NoTlsv13.ValueBool()
	}
	if !bind.NoTlsTickets.IsNull() && !bind.NoTlsTickets.IsUnknown() {
		payload.NoTlsTickets = bind.NoTlsTickets.ValueBool()
	}

	return payload
}

func (r *BindManager) updateBindsWithHandlingInTransaction(ctx context.Context, transactionID string, parentType string, parentName string, existingBinds []BindPayload, newBinds map[string]haproxyBindModel) error {
	// Create maps for efficient lookup
	existingBindMap := make(map[string]*BindPayload)
	for i := range existingBinds {
		existingBindMap[existingBinds[i].Name] = &existingBinds[i]
	}

	// Track which binds we've processed to avoid duplicates
	processedBinds := make(map[string]bool)

	// Track binds that need to be recreated due to changes
	var bindsToRecreate []string

	// First pass: identify binds that need changes and mark them for update or recreation
	for bindName, newBind := range newBinds {

		// Check if this bind exists by name
		if existingBind, exists := existingBindMap[bindName]; exists {
			// Bind exists, check if content has changed
			if r.hasBindChanged(existingBind, &newBind) {
				// Content has changed, check if it's a simple update or needs recreation
				if r.canUpdateBind(existingBind, &newBind) {
					// Simple field change, can update in place
					log.Printf("Bind '%s' has changed, will update in place", bindName)
					// Update the bind immediately
					bindPayload := r.convertToBindPayload(bindName, &newBind)
					_, err := r.client.UpdateBindInTransaction(ctx, transactionID, existingBind.Name, parentType, parentName, bindPayload)
					if err != nil {
						return fmt.Errorf("failed to update bind '%s': %w", bindName, err)
					}
					log.Printf("Bind '%s' updated successfully", bindName)
				} else {
					// Complex change, needs recreation
					log.Printf("Bind '%s' has complex changes, will recreate", bindName)
					bindsToRecreate = append(bindsToRecreate, bindName)
				}
			} else {
				// Bind is identical, no changes needed
				log.Printf("Bind '%s' is unchanged", bindName)
			}
			// Mark this bind as processed
			processedBinds[bindName] = true
		} else {
			// This is a new bind, mark for creation
			log.Printf("Bind '%s' is new, will create", bindName)
		}
	}

	// Second pass: delete all binds that need to be recreated (due to content changes)
	for _, bindName := range bindsToRecreate {
		if existingBind, exists := existingBindMap[bindName]; exists {
			log.Printf("Deleting bind '%s' for recreation", bindName)
			_, err := r.client.DeleteBindInTransaction(ctx, transactionID, existingBind.Name, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete bind '%s': %w", bindName, err)
			}
		}
	}

	// Third pass: create all binds that need to be recreated at their positions
	for _, bindName := range bindsToRecreate {
		newBind := newBinds[bindName]

		log.Printf("Creating bind '%s'", bindName)
		bindPayload := r.convertToBindPayload(bindName, &newBind)

		_, err := r.client.CreateBindInTransaction(ctx, transactionID, parentType, parentName, bindPayload)
		if err != nil {
			return fmt.Errorf("failed to create bind '%s': %w", bindName, err)
		}
	}

	// Delete binds that are no longer needed (not in the new configuration)
	var bindsToDelete []BindPayload
	for _, existingBind := range existingBinds {
		if !processedBinds[existingBind.Name] {
			bindsToDelete = append(bindsToDelete, existingBind)
		}
	}

	// Sort by name in descending order (reverse alphabetical)
	sort.Slice(bindsToDelete, func(i, j int) bool {
		return bindsToDelete[i].Name > bindsToDelete[j].Name
	})

	// Delete binds in reverse order
	for _, bindToDelete := range bindsToDelete {
		log.Printf("Deleting bind '%s' (no longer needed)", bindToDelete.Name)
		_, err := r.client.DeleteBindInTransaction(ctx, transactionID, bindToDelete.Name, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete bind '%s': %w", bindToDelete.Name, err)
		}
	}

	// Create new binds that don't exist yet
	for bindName, newBind := range newBinds {
		if !processedBinds[bindName] {
			// This is a new bind, create it
			log.Printf("Creating new bind '%s'", bindName)
			bindPayload := r.convertToBindPayload(bindName, &newBind)

			_, err := r.client.CreateBindInTransaction(ctx, transactionID, parentType, parentName, bindPayload)
			if err != nil {
				return fmt.Errorf("failed to create bind '%s': %w", bindName, err)
			}
		}
	}

	return nil
}

func (r *BindManager) updateBindsWithHandling(ctx context.Context, parentType string, parentName string, existingBinds []BindPayload, newBinds map[string]haproxyBindModel) error {
	// Create maps for efficient lookup
	existingBindMap := make(map[string]*BindPayload)
	for i := range existingBinds {
		existingBindMap[existingBinds[i].Name] = &existingBinds[i]
	}

	// Track which binds we've processed to avoid duplicates
	processedBinds := make(map[string]bool)

	// Track binds that need to be recreated due to changes
	var bindsToRecreate []string

	// First pass: identify binds that need changes and mark them for recreation
	for bindName, newBind := range newBinds {

		// Check if this bind exists by name
		if existingBind, exists := existingBindMap[bindName]; exists {
			// Bind exists, check if content has changed
			if r.hasBindChanged(existingBind, &newBind) {
				// Content has changed, mark for recreation
				log.Printf("Bind '%s' has changed, will recreate", bindName)
				bindsToRecreate = append(bindsToRecreate, bindName)
			} else {
				// Bind is identical, no changes needed
				log.Printf("Bind '%s' is unchanged", bindName)
			}
			// Mark this bind as processed
			processedBinds[bindName] = true
		} else {
			// This is a new bind, mark for creation
			log.Printf("Bind '%s' is new, will create", bindName)
		}
	}

	// Second pass: delete all binds that need to be recreated (due to content changes)
	for _, bindName := range bindsToRecreate {
		if existingBind, exists := existingBindMap[bindName]; exists {
			log.Printf("Deleting bind '%s' for recreation", bindName)
			err := r.client.DeleteBind(ctx, existingBind.Name, parentType, parentName)
			if err != nil {
				return fmt.Errorf("failed to delete bind '%s': %w", bindName, err)
			}
		}
	}

	// Third pass: create all binds that need to be recreated at their positions
	for _, bindName := range bindsToRecreate {
		newBind := newBinds[bindName]

		log.Printf("Creating bind '%s'", bindName)
		bindPayload := r.convertToBindPayload(bindName, &newBind)

		err := r.client.CreateBind(ctx, parentType, parentName, bindPayload)
		if err != nil {
			return fmt.Errorf("failed to create bind '%s': %w", bindName, err)
		}
	}

	// Delete binds that are no longer needed (not in the new configuration)
	var bindsToDelete []BindPayload
	for _, existingBind := range existingBinds {
		if !processedBinds[existingBind.Name] {
			bindsToDelete = append(bindsToDelete, existingBind)
		}
	}

	// Sort by name in descending order (reverse alphabetical)
	sort.Slice(bindsToDelete, func(i, j int) bool {
		return bindsToDelete[i].Name > bindsToDelete[j].Name
	})

	// Delete binds in reverse order
	for _, bindToDelete := range bindsToDelete {
		log.Printf("Deleting bind '%s' (no longer needed)", bindToDelete.Name)
		err := r.client.DeleteBind(ctx, bindToDelete.Name, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete bind '%s': %w", bindToDelete.Name, err)
		}
	}

	// Create new binds that don't exist yet
	for bindName, newBind := range newBinds {
		if !processedBinds[bindName] {
			// This is a new bind, create it
			log.Printf("Creating new bind '%s'", bindName)
			bindPayload := r.convertToBindPayload(bindName, &newBind)

			err := r.client.CreateBind(ctx, parentType, parentName, bindPayload)
			if err != nil {
				return fmt.Errorf("failed to create bind '%s': %w", bindName, err)
			}
		}
	}

	return nil
}

func (r *BindManager) hasBindChanged(existing *BindPayload, new *haproxyBindModel) bool {
	// Compare the most important fields
	if existing.Address != new.Address.ValueString() {
		return true
	}
	if existing.Port != nil && *existing.Port != new.Port.ValueInt64() {
		return true
	}
	if existing.Mode != new.Mode.ValueString() {
		return true
	}
	if existing.Ssl != new.Ssl.ValueBool() {
		return true
	}

	// Check Process field (important for HAProxy v2)
	if existing.Process != new.Process.ValueString() {
		return true
	}

	// Check other important fields that could cause drift
	if existing.Maxconn > 0 && new.Maxconn.ValueInt64() > 0 && existing.Maxconn != new.Maxconn.ValueInt64() {
		return true
	}
	if existing.User != new.User.ValueString() {
		return true
	}
	if existing.Group != new.Group.ValueString() {
		return true
	}
	if existing.Interface != new.Interface.ValueString() {
		return true
	}
	if existing.Level != new.Level.ValueString() {
		return true
	}

	// Add more field comparisons as needed
	return false
}

// canUpdateBind determines if a bind can be updated in place or needs recreation
func (r *BindManager) canUpdateBind(existing *BindPayload, new *haproxyBindModel) bool {
	// Simple field changes that can be updated in place
	// Port changes are safe to update
	if existing.Port != nil && *existing.Port != new.Port.ValueInt64() {
		return true
	}

	// Maxconn changes are safe to update
	if existing.Maxconn != new.Maxconn.ValueInt64() {
		return true
	}

	// User/Group changes are safe to update
	if existing.User != new.User.ValueString() || existing.Group != new.Group.ValueString() {
		return true
	}

	// Level changes are safe to update
	if existing.Level != new.Level.ValueString() {
		return true
	}

	// For complex changes (SSL, ciphers, etc.), we need recreation
	// This is a conservative approach - only allow simple field updates
	return false
}

func (r *BindManager) deleteAllBindsInTransaction(ctx context.Context, transactionID string, parentType string, parentName string) error {
	binds, err := r.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read binds for deletion: %w", err)
	}

	// Delete in reverse order to avoid issues
	sort.Slice(binds, func(i, j int) bool {
		return binds[i].Name > binds[j].Name
	})

	for _, bind := range binds {
		log.Printf("Deleting bind '%s' in transaction %s", bind.Name, transactionID)
		_, err := r.client.DeleteBindInTransaction(ctx, transactionID, bind.Name, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete bind '%s': %w", bind.Name, err)
		}
	}

	return nil
}

func (r *BindManager) deleteAllBinds(ctx context.Context, parentType string, parentName string) error {
	binds, err := r.ReadBinds(ctx, parentType, parentName)
	if err != nil {
		return fmt.Errorf("failed to read binds for deletion: %w", err)
	}

	// Delete in reverse order to avoid issues
	sort.Slice(binds, func(i, j int) bool {
		return binds[i].Name > binds[j].Name
	})

	for _, bind := range binds {
		log.Printf("Deleting bind '%s'", bind.Name)
		err := r.client.DeleteBind(ctx, bind.Name, parentType, parentName)
		if err != nil {
			return fmt.Errorf("failed to delete bind '%s': %w", bind.Name, err)
		}
	}

	return nil
}
