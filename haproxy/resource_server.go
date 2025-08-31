package haproxy

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// GetServerSchema returns the schema for the server block
func GetServerSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Server configuration.",
		Attributes: map[string]schema.Attribute{
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
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the server is disabled.",
			},
		},
	}
}
