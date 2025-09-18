package haproxy

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &haproxyProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &haproxyProvider{}
}

// haproxyProvider is the provider implementation.
type haproxyProvider struct{}

// haproxyProviderModel maps provider schema data to a Go type.
type haproxyProviderModel struct {
	URL        types.String `tfsdk:"url"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Insecure   types.Bool   `tfsdk:"insecure"`
	APIVersion types.String `tfsdk:"api_version"`
}

// ProviderData contains data that resources and data sources can access
type ProviderData struct {
	Client     *HAProxyClient
	APIVersion string
}

// Metadata returns the provider type name.
func (p *haproxyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "haproxy"
}

// Schema defines the provider-level schema for configuration data.
func (p *haproxyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The URL of the HAProxy Data Plane API.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for the HAProxy Data Plane API.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for the HAProxy Data Plane API.",
				Required:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Whether to skip SSL certificate verification.",
				Optional:    true,
			},
			"api_version": schema.StringAttribute{
				Description: "The version of the HAProxy Data Plane API to use.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a HAProxy API client for data source and resource calls.
func (p *haproxyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config haproxyProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure the client
	httpClient := &http.Client{}
	if config.Insecure.ValueBool() {
		// Add transport to skip verification
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	apiVersion := config.APIVersion.ValueString()
	if apiVersion == "" {
		apiVersion = "v3"
	}

	client := NewHAProxyClient(httpClient, config.URL.ValueString(), config.Username.ValueString(), config.Password.ValueString(), apiVersion)

	// Create a provider data structure that includes both client and API version
	providerData := &ProviderData{
		Client:     client,
		APIVersion: apiVersion,
	}

	// Make the provider data available to resources and data sources
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// DataSources defines the data sources implemented in the provider.
func (p *haproxyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAclDataSource,
		NewBackendsDataSource,
		NewFrontendsDataSource,
		NewHttpcheckDataSource,
		NewHttpRequestRuleDataSource,
		NewHttpResponseRuleDataSource,
		NewServerDataSource,
		NewTcpCheckDataSource,
		NewTcpCheckSingleDataSource,
		NewTcpRequestRuleDataSource,
		NewTcpRequestRuleSingleDataSource,
		NewTcpResponseRuleDataSource,
		NewTcpResponseRuleSingleDataSource,
		NewBackendSingleDataSource,
		NewFrontendSingleDataSource,
		NewServerSingleDataSource,
		NewAclSingleDataSource,
		NewHttpcheckSingleDataSource,
		NewHttpRequestRuleSingleDataSource,
		NewHttpResponseRuleSingleDataSource,
		NewBindDataSource,
		NewBindSingleDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *haproxyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHaproxyStackResource,
	}
}
