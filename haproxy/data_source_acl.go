package haproxy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AclSingleDataSource defines the single data source implementation.
type AclSingleDataSource struct {
	client *HAProxyClient
}

// AclSingleDataSourceModel describes the single data source data model.
type AclSingleDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
	Index      types.Int64  `tfsdk:"index"`
	Acl        types.String `tfsdk:"acl"`
}

// Metadata returns the single data source type name.
func (d *AclSingleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_single"
}

// Schema defines the schema for the single data source.
func (d *AclSingleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Single ACL data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ACL identifier",
				Computed:            true,
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
				MarkdownDescription: "ACL index",
				Required:            true,
			},
			"acl": schema.StringAttribute{
				MarkdownDescription: "Complete ACL data from HAProxy API as JSON string",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the single data source.
func (d *AclSingleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
}

// Read refreshes the Terraform state with the latest data for single data source.
func (d *AclSingleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AclSingleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the ACLs
	acls, err := d.client.ReadAcls(ctx, data.ParentType.ValueString(), data.ParentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ACLs, got error: %s", err))
		return
	}

	// Find the specific ACL by array position (more predictable than API index)
	var foundAcl *AclPayload
	if data.Index.ValueInt64() < int64(len(acls)) {
		foundAcl = &acls[data.Index.ValueInt64()]
	}

	if foundAcl == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ACL at position %d not found", data.Index.ValueInt64()))
		return
	}

	// Convert to data source model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", data.ParentType.ValueString(), data.ParentName.ValueString(), data.Index.ValueInt64()))

	// Fix the index to use array position instead of API index
	foundAcl.Index = data.Index.ValueInt64()

	// Convert ACL to JSON for dynamic output
	jsonData, err := json.Marshal(foundAcl)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal ACL to JSON, got error: %s", err))
		return
	}
	data.Acl = types.StringValue(string(jsonData))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func NewAclDataSource() datasource.DataSource {
	return &aclDataSource{}
}

// NewAclSingleDataSource creates a new single ACL data source
func NewAclSingleDataSource() datasource.DataSource {
	return &AclSingleDataSource{}
}

type aclDataSource struct {
	client *HAProxyClient
}

type aclDataSourceModel struct {
	Acls       types.String `tfsdk:"acls"`
	ParentType types.String `tfsdk:"parent_type"`
	ParentName types.String `tfsdk:"parent_name"`
}

func (d *aclDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl"
}

func (d *aclDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"acls": schema.StringAttribute{
				Computed:    true,
				Description: "Complete ACLs data from HAProxy API as JSON string",
			},
			"parent_type": schema.StringAttribute{
				Required:    true,
				Description: "The parent type (frontend or backend).",
			},
			"parent_name": schema.StringAttribute{
				Required:    true,
				Description: "The parent name to get ACLs for.",
			},
		},
	}
}

func (d *aclDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
}

func (d *aclDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state aclDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType := state.ParentType.ValueString()
	parentName := state.ParentName.ValueString()

	acls, err := d.client.ReadACLs(ctx, parentType, parentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HAProxy ACLs",
			"Could not read HAProxy ACLs, unexpected error: "+err.Error(),
		)
		return
	}

	// Fix index field - use array position if API returns 0 for all rules
	for i := range acls {
		acls[i].Index = int64(i)
	}

	// Convert ACLs to JSON for dynamic output
	jsonData, err := json.Marshal(acls)
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Unable to marshal ACLs to JSON, got error: %s", err))
		return
	}

	// Set JSON string directly
	state.Acls = types.StringValue(string(jsonData))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
