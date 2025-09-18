package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// GetServerAttributes returns the common server attributes (without name field)
func GetServerAttributes() map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{
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
		"ssl_certificate": schema.StringAttribute{
			Optional:    true,
			Description: "SSL certificate for the server.",
		},
		"ssl_cafile": schema.StringAttribute{
			Optional:    true,
			Description: "SSL CA file for the server.",
		},
		"ssl_max_ver": schema.StringAttribute{
			Optional:    true,
			Description: "Maximum SSL/TLS version for the server.",
		},
		"ssl_min_ver": schema.StringAttribute{
			Optional:    true,
			Description: "Minimum SSL/TLS version for the server.",
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

	// Force TLS fields (v2 only, deprecated in v3)
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

	return attributes
}

// GetServerSchema returns the schema for the server block
func GetServerSchema() schema.SingleNestedBlock {
	attributes := GetServerAttributes()

	// Add the name field for individual server resources
	attributes["name"] = schema.StringAttribute{
		Required:    true,
		Description: "The name of the server.",
	}

	return schema.SingleNestedBlock{
		Description: "Server configuration.",
		Attributes:  attributes,
	}
}

// GetServersSchema returns the schema for the servers block (multiple servers)
func GetServersSchema() schema.MapNestedAttribute {
	// Reuse the common server attributes (without name field)
	attributes := GetServerAttributes()

	return schema.MapNestedAttribute{
		Optional:    true,
		Description: "Multiple server configurations.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: attributes,
		},
	}
}
