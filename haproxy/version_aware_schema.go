package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// VersionAwareSchemaBuilder provides methods to build schemas based on API version
type VersionAwareSchemaBuilder struct {
	apiVersion string
}

// NewVersionAwareSchemaBuilder creates a new schema builder for the specified API version
func CreateVersionAwareSchemaBuilder(apiVersion string) *VersionAwareSchemaBuilder {
	if apiVersion == "" {
		apiVersion = "v3" // Default to v3 (latest version)
	}
	return &VersionAwareSchemaBuilder{apiVersion: apiVersion}
}

// IsV3 returns true if the API version is v3
func (b *VersionAwareSchemaBuilder) IsV3() bool {
	return b.apiVersion == "v3"
}

// IsV2 returns true if the API version is v2
func (b *VersionAwareSchemaBuilder) IsV2() bool {
	return b.apiVersion == "v2"
}

// BuildDefaultServerSchema builds the default_server schema based on API version
func (b *VersionAwareSchemaBuilder) BuildDefaultServerSchema() schema.ListNestedBlock {
	attributes := map[string]schema.Attribute{
		"ssl": schema.StringAttribute{
			Optional:    true,
			Description: "SSL configuration for the default server.",
		},
		"ssl_cafile": schema.StringAttribute{
			Optional:    true,
			Description: "SSL CA file for the default server.",
		},
		"ssl_certificate": schema.StringAttribute{
			Optional:    true,
			Description: "SSL certificate for the default server.",
		},
		"ssl_max_ver": schema.StringAttribute{
			Optional:    true,
			Description: "SSL maximum version for the default server.",
		},
		"ssl_min_ver": schema.StringAttribute{
			Optional:    true,
			Description: "SSL minimum version for the default server.",
		},
		"ssl_reuse": schema.StringAttribute{
			Optional:    true,
			Description: "SSL reuse configuration for the default server.",
		},
		"ciphers": schema.StringAttribute{
			Optional:    true,
			Description: "Ciphers for the default server.",
		},
		"ciphersuites": schema.StringAttribute{
			Optional:    true,
			Description: "Cipher suites for the default server.",
		},
		"verify": schema.StringAttribute{
			Optional:    true,
			Description: "SSL verification for the default server.",
		},
	}

	// Add v3-specific fields
	if b.IsV3() {
		attributes["sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "SSLv3 support for the default server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv10"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.0 support for the default server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv11"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.1 support for the default server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv12"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.2 support for the default server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv13"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.3 support for the default server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
	}

	// Add v2-specific fields (deprecated in v3)
	if b.IsV2() {
		attributes["sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "SSLv3 support for the default server (v2 only, deprecated in v3).",
		}
		attributes["no_sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable SSLv3 for the default server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv10"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.0 for the default server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv11"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.1 for the default server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv12"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.2 for the default server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv13"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.3 for the default server (v2 only, deprecated in v3).",
		}
		attributes["force_sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force SSLv3 for the default server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv10"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.0 for the default server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv11"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.1 for the default server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv12"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.2 for the default server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv13"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.3 for the default server (v2 only, deprecated in v3).",
		}
		attributes["force_strict_sni"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force strict SNI for the default server (v2 only, deprecated in v3).",
		}
	}

	return schema.ListNestedBlock{
		Description: "Default server configuration for the backend.",
		NestedObject: schema.NestedBlockObject{
			Attributes: attributes,
		},
	}
}

// BuildServerSchema builds the server schema based on API version
func (b *VersionAwareSchemaBuilder) BuildServerSchema() schema.ListNestedBlock {
	attributes := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the server.",
		},
		"address": schema.StringAttribute{
			Required:    true,
			Description: "The IP address of the server.",
		},
		"port": schema.Int64Attribute{
			Required:    true,
			Description: "The port number of the server.",
		},
		"check": schema.StringAttribute{
			Optional:    true,
			Description: "The health check configuration for the server.",
		},
		"backup": schema.StringAttribute{
			Optional:    true,
			Description: "Whether the server is a backup server.",
		},
		"maxconn": schema.Int64Attribute{
			Optional:    true,
			Description: "The maximum number of connections for the server.",
		},
		"weight": schema.Int64Attribute{
			Optional:    true,
			Description: "The weight of the server for load balancing.",
		},
		"rise": schema.Int64Attribute{
			Optional:    true,
			Description: "The number of successful health checks before marking the server as up.",
		},
		"fall": schema.Int64Attribute{
			Optional:    true,
			Description: "The number of failed health checks before marking the server as down.",
		},
		"inter": schema.Int64Attribute{
			Optional:    true,
			Description: "The interval between health checks.",
		},
		"fastinter": schema.Int64Attribute{
			Optional:    true,
			Description: "The fast interval between health checks.",
		},
		"downinter": schema.Int64Attribute{
			Optional:    true,
			Description: "The interval between health checks when the server is down.",
		},
		"ssl": schema.StringAttribute{
			Optional:    true,
			Description: "SSL configuration for the server.",
		},
		"verify": schema.StringAttribute{
			Optional:    true,
			Description: "SSL verification for the server.",
		},
		"cookie": schema.StringAttribute{
			Optional:    true,
			Description: "Cookie configuration for the server.",
		},
		"disabled": schema.BoolAttribute{
			Optional:    true,
			Description: "Whether the server is disabled.",
		},
	}

	// Add v3-specific fields
	if b.IsV3() {
		attributes["sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "SSLv3 support for the server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv10"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.0 support for the server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv11"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.1 support for the server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv12"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.2 support for the server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
		attributes["tlsv13"] = schema.StringAttribute{
			Optional:    true,
			Description: "TLSv1.3 support for the server (v3 only).",
			Validators: []validator.String{
				stringvalidator.OneOf("enabled", "disabled"),
			},
		}
	}

	// Add v2-specific fields (deprecated in v3)
	if b.IsV2() {
		attributes["force_sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force SSLv3 for the server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv10"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.0 for the server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv11"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.1 for the server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv12"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.2 for the server (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv13"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force TLSv1.3 for the server (v2 only, deprecated in v3).",
		}
		attributes["force_strict_sni"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force strict SNI for the server (v2 only, deprecated in v3).",
		}
		attributes["no_sslv3"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable SSLv3 for the server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv10"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.0 for the server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv11"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.1 for the server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv12"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.2 for the server (v2 only, deprecated in v3).",
		}
		attributes["no_tlsv13"] = schema.StringAttribute{
			Optional:    true,
			Description: "Disable TLSv1.3 for the server (v2 only, deprecated in v3).",
		}
	}

	return schema.ListNestedBlock{
		Description: "Server configuration for the backend.",
		NestedObject: schema.NestedBlockObject{
			Attributes: attributes,
		},
	}
}

// BuildBindSchema builds the bind schema based on API version
func (b *VersionAwareSchemaBuilder) BuildBindSchema() schema.ListNestedBlock {
	attributes := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the bind.",
		},
		"address": schema.StringAttribute{
			Required:    true,
			Description: "The IP address for the bind.",
		},
		"port": schema.Int64Attribute{
			Optional:    true,
			Description: "The port number for the bind.",
		},
		"maxconn": schema.Int64Attribute{
			Optional:    true,
			Description: "The maximum number of connections for the bind.",
		},
		"user": schema.StringAttribute{
			Optional:    true,
			Description: "The user for the bind.",
		},
		"group": schema.StringAttribute{
			Optional:    true,
			Description: "The group for the bind.",
		},
		"mode": schema.StringAttribute{
			Optional:    true,
			Description: "The mode for the bind.",
		},
		"ssl": schema.BoolAttribute{
			Optional:    true,
			Description: "Whether SSL is enabled for the bind.",
		},
		"ssl_cafile": schema.StringAttribute{
			Optional:    true,
			Description: "The SSL CA file for the bind.",
		},
		"ssl_max_ver": schema.StringAttribute{
			Optional:    true,
			Description: "The SSL maximum version for the bind.",
		},
		"ssl_min_ver": schema.StringAttribute{
			Optional:    true,
			Description: "The SSL minimum version for the bind.",
		},
		"ssl_certificate": schema.StringAttribute{
			Optional:    true,
			Description: "The SSL certificate for the bind.",
		},
		"ciphers": schema.StringAttribute{
			Optional:    true,
			Description: "The ciphers for the bind.",
		},
		"ciphersuites": schema.StringAttribute{
			Optional:    true,
			Description: "The cipher suites for the bind.",
		},
		"transparent": schema.BoolAttribute{
			Optional:    true,
			Description: "Whether the bind is transparent.",
		},
	}

	// Add v3-specific fields
	if b.IsV3() {
		attributes["sslv3"] = schema.BoolAttribute{
			Optional:    true,
			Description: "SSLv3 support for the bind (v3 only).",
		}
		attributes["tlsv10"] = schema.BoolAttribute{
			Optional:    true,
			Description: "TLSv1.0 support for the bind (v3 only).",
		}
		attributes["tlsv11"] = schema.BoolAttribute{
			Optional:    true,
			Description: "TLSv1.1 support for the bind (v3 only).",
		}
		attributes["tlsv12"] = schema.BoolAttribute{
			Optional:    true,
			Description: "TLSv1.2 support for the bind (v3 only).",
		}
		attributes["tlsv13"] = schema.BoolAttribute{
			Optional:    true,
			Description: "TLSv1.3 support for the bind (v3 only).",
		}
	}

	// Add v2-specific fields (deprecated in v3)
	if b.IsV2() {
		attributes["force_sslv3"] = schema.BoolAttribute{
			Optional:    true,
			Description: "Force SSLv3 for the bind (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv10"] = schema.BoolAttribute{
			Optional:    true,
			Description: "Force TLSv1.0 for the bind (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv11"] = schema.BoolAttribute{
			Optional:    true,
			Description: "Force TLSv1.1 for the bind (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv12"] = schema.BoolAttribute{
			Optional:    true,
			Description: "Force TLSv1.2 for the bind (v2 only, deprecated in v3).",
		}
		attributes["force_tlsv13"] = schema.BoolAttribute{
			Optional:    true,
			Description: "Force TLSv1.3 for the bind (v2 only, deprecated in v3).",
		}
		attributes["force_strict_sni"] = schema.StringAttribute{
			Optional:    true,
			Description: "Force strict SNI for the bind (v2 only, deprecated in v3).",
		}
	}

	return schema.ListNestedBlock{
		Description: "Bind configuration for the frontend.",
		NestedObject: schema.NestedBlockObject{
			Attributes: attributes,
		},
	}
}
