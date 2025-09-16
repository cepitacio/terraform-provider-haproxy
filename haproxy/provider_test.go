package haproxy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

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
