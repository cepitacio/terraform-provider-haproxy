// main.go
package main

import (
	"context"
	"log"
	"terraform-provider-haproxy/haproxy"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	version string = "dev"
)

func main() {
	err := providerserver.Serve(context.Background(), haproxy.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/haproxy/haproxy",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
