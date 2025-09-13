package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// GetServerSchema returns the schema for the server block
func GetServerSchema(schemaBuilder *VersionAwareSchemaBuilder) schema.SingleNestedBlock {
	// If no schema builder is provided, include all fields for backward compatibility
	if schemaBuilder == nil {
		schemaBuilder = CreateVersionAwareSchemaBuilder("v3") // Default to v3
	}
	attributes := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the server.",
		},
		"address": schema.StringAttribute{
			Required:    true,
			Description: "The address of the server.",
		},
		"port": schema.Int64Attribute{
			Required:    true,
			Description: "The port of the server.",
		},
		"check": schema.StringAttribute{
			Optional:    true,
			Description: "Whether to enable health checks for the server.",
		},
		"backup": schema.StringAttribute{
			Optional:    true,
			Description: "Whether the server is a backup server.",
		},
		"maxconn": schema.Int64Attribute{
			Optional:    true,
			Description: "Maximum number of connections for the server.",
		},
		"weight": schema.Int64Attribute{
			Optional:    true,
			Description: "Load balancing weight for the server.",
		},
		"rise": schema.Int64Attribute{
			Optional:    true,
			Description: "Number of successful health checks to mark server as up.",
		},
		"fall": schema.Int64Attribute{
			Optional:    true,
			Description: "Number of failed health checks to mark server as down.",
		},
		"inter": schema.Int64Attribute{
			Optional:    true,
			Description: "Interval between health checks in milliseconds.",
		},
		"fastinter": schema.Int64Attribute{
			Optional:    true,
			Description: "Fast interval between health checks in milliseconds.",
		},
		"downinter": schema.Int64Attribute{
			Optional:    true,
			Description: "Down interval between health checks in milliseconds.",
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
			Description: "Cookie value for the server.",
		},
	}

	// Add all SSL/TLS fields with version information in descriptions
	// v3 fields
	attributes["sslv3"] = schema.StringAttribute{
		Optional:    true,
		Description: "SSLv3 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv10"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.0 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv11"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.1 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv12"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.2 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv13"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.3 support for the server (Data Plane API v3 only).",
	}

	// v2 fields (deprecated in v3)
	attributes["force_sslv3"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force SSLv3 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv10"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.0 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv11"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.1 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv12"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.2 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv13"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.3 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_strict_sni"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force strict SNI for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_sslv3"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable SSLv3 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv10"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.0 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv11"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.1 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv12"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.2 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv13"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.3 for the server (Data Plane API v2 only, deprecated in v3).",
	}

	return schema.SingleNestedBlock{
		Description: "Server configuration.",
		Attributes:  attributes,
	}
}

// GetServersSchema returns the schema for the servers block (multiple servers)
func GetServersSchema(schemaBuilder *VersionAwareSchemaBuilder) schema.MapNestedAttribute {
	// If no schema builder is provided, include all fields for backward compatibility
	if schemaBuilder == nil {
		schemaBuilder = CreateVersionAwareSchemaBuilder("v3") // Default to v3
	}
	attributes := map[string]schema.Attribute{
		// Note: "name" is now the map key, not a field
		"address": schema.StringAttribute{
			Required:    true,
			Description: "The address of the server.",
		},
		"port": schema.Int64Attribute{
			Required:    true,
			Description: "The port of the server.",
		},
		"check": schema.StringAttribute{
			Optional:    true,
			Description: "Whether to enable health checks for the server.",
		},
		"backup": schema.StringAttribute{
			Optional:    true,
			Description: "Whether the server is a backup server.",
		},
		"maxconn": schema.Int64Attribute{
			Optional:    true,
			Description: "Maximum number of connections for the server.",
		},
		"weight": schema.Int64Attribute{
			Optional:    true,
			Description: "Load balancing weight for the server.",
		},
		"rise": schema.Int64Attribute{
			Optional:    true,
			Description: "Number of successful health checks to mark server as up.",
		},
		"fall": schema.Int64Attribute{
			Optional:    true,
			Description: "Number of failed health checks to mark server as down.",
		},
		"inter": schema.Int64Attribute{
			Optional:    true,
			Description: "Interval between health checks in milliseconds.",
		},
		"fastinter": schema.Int64Attribute{
			Optional:    true,
			Description: "Fast interval between health checks in milliseconds.",
		},
		"downinter": schema.Int64Attribute{
			Optional:    true,
			Description: "Down interval between health checks in milliseconds.",
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
			Description: "Cookie value for the server.",
		},
	}

	// Add all SSL/TLS fields with version information in descriptions
	// v3 fields
	attributes["sslv3"] = schema.StringAttribute{
		Optional:    true,
		Description: "SSLv3 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv10"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.0 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv11"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.1 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv12"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.2 support for the server (Data Plane API v3 only).",
	}
	attributes["tlsv13"] = schema.StringAttribute{
		Optional:    true,
		Description: "TLSv1.3 support for the server (Data Plane API v3 only).",
	}

	// v2 fields (deprecated in v3)
	attributes["force_sslv3"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force SSLv3 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv10"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.0 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv11"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.1 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv12"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.2 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_tlsv13"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force TLSv1.3 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["force_strict_sni"] = schema.StringAttribute{
		Optional:    true,
		Description: "Force strict SNI for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_sslv3"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable SSLv3 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv10"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.0 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv11"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.1 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv12"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.2 for the server (Data Plane API v2 only, deprecated in v3).",
	}
	attributes["no_tlsv13"] = schema.StringAttribute{
		Optional:    true,
		Description: "Disable TLSv1.3 for the server (Data Plane API v2 only, deprecated in v3).",
	}

	return schema.MapNestedAttribute{
		Optional:    true,
		Description: "Multiple server configurations.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: attributes,
		},
	}
}
