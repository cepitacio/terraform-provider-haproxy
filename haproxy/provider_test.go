package haproxy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server with the given provider
// configuration.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"haproxy": providerserver.NewProtocol6WithError(New()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the environment that are necessary for the test to run.
}

func TestProvider(t *testing.T) {
	t.Parallel()

	// Test that the provider can be created
	provider := New()
	if provider == nil {
		t.Fatal("Expected provider to be created, got nil")
	}
}

func TestProviderFrameworkIntegration(t *testing.T) {
	t.Parallel()

	// This test ensures the provider works with the Terraform Plugin Framework
	provider := New()

	// Test that the provider can be used with providerserver
	server, err := providerserver.NewProtocol6WithError(provider)()
	if err != nil {
		t.Fatalf("Failed to create provider server: %v", err)
	}

	var _ tfprotov6.ProviderServer = server
}
