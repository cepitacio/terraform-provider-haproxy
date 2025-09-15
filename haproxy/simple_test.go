package haproxy

import (
	"testing"
)

// Test that the provider can be created
func TestProviderCreation(t *testing.T) {
	t.Parallel()

	provider := New()
	if provider == nil {
		t.Fatal("Expected provider to be created, got nil")
	}
}

// Test that the haproxy_stack resource can be created
func TestHaproxyStackResourceCreation(t *testing.T) {
	t.Parallel()

	resource := NewHaproxyStackResource()
	if resource == nil {
		t.Fatal("Expected resource to be created, got nil")
	}
}

// Test that data sources can be created
func TestDataSourcesCreation(t *testing.T) {
	t.Parallel()

	dataSources := []func() interface{}{
		func() interface{} { return NewBackendsDataSource() },
		func() interface{} { return NewBackendSingleDataSource() },
		func() interface{} { return NewFrontendsDataSource() },
		func() interface{} { return NewFrontendSingleDataSource() },
		func() interface{} { return NewAclDataSource() },
		func() interface{} { return NewAclSingleDataSource() },
		func() interface{} { return NewBindDataSource() },
		func() interface{} { return NewBindSingleDataSource() },
		func() interface{} { return NewServerDataSource() },
		func() interface{} { return NewServerSingleDataSource() },
		func() interface{} { return NewHttpRequestRuleDataSource() },
		func() interface{} { return NewHttpRequestRuleSingleDataSource() },
		func() interface{} { return NewHttpResponseRuleDataSource() },
		func() interface{} { return NewHttpResponseRuleSingleDataSource() },
		func() interface{} { return NewHttpcheckDataSource() },
		func() interface{} { return NewHttpcheckSingleDataSource() },
		func() interface{} { return NewTcpCheckDataSource() },
		func() interface{} { return NewTcpCheckSingleDataSource() },
		func() interface{} { return NewTcpRequestRuleDataSource() },
		func() interface{} { return NewTcpRequestRuleSingleDataSource() },
		func() interface{} { return NewTcpResponseRuleDataSource() },
		func() interface{} { return NewTcpResponseRuleSingleDataSource() },
	}

	for i, dsFunc := range dataSources {
		ds := dsFunc()
		if ds == nil {
			t.Errorf("Expected data source %d to be created, got nil", i)
		}
	}
}

// Test that the provider builds without errors
func TestProviderBuilds(t *testing.T) {
	t.Parallel()

	// This test ensures the provider compiles correctly
	// If we get here, the build was successful
	t.Log("Provider builds successfully")
}

// Benchmark test for provider creation
func BenchmarkProviderCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider := New()
		if provider == nil {
			b.Fatal("Expected provider to be created, got nil")
		}
	}
}

// Benchmark test for resource creation
func BenchmarkResourceCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resource := NewHaproxyStackResource()
		if resource == nil {
			b.Fatal("Expected resource to be created, got nil")
		}
	}
}
