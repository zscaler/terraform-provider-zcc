package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/datasources"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/resources"
)

var _ provider.Provider = &ZCCProvider{}

type ZCCProvider struct {
	version string
}

type ZCCProviderModel struct {
	ClientID       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	PrivateKey     types.String `tfsdk:"private_key"`
	VanityDomain   types.String `tfsdk:"vanity_domain"`
	ZscalerCloud   types.String `tfsdk:"zscaler_cloud"`
	HTTPProxy      types.String `tfsdk:"http_proxy"`
	MaxRetries     types.Int64  `tfsdk:"max_retries"`
	Parallelism    types.Int64  `tfsdk:"parallelism"`
	RequestTimeout types.Int64  `tfsdk:"request_timeout"`
	MinWaitSeconds types.Int64  `tfsdk:"min_wait_seconds"`
	MaxWaitSeconds types.Int64  `tfsdk:"max_wait_seconds"`
}

func (p *ZCCProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zcc"
	resp.Version = p.version
}

func (p *ZCCProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "ZCC client ID (OAuth2)",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "ZCC client secret (OAuth2)",
				Optional:    true,
				Sensitive:   true,
			},
			"private_key": schema.StringAttribute{
				Description: "ZCC private key (OAuth2)",
				Optional:    true,
				Sensitive:   true,
			},
			"vanity_domain": schema.StringAttribute{
				Description: "Zscaler Vanity Domain",
				Optional:    true,
				Sensitive:   true,
			},
			"zscaler_cloud": schema.StringAttribute{
				Description: "Zscaler Cloud Name",
				Optional:    true,
				Sensitive:   true,
			},
			"http_proxy": schema.StringAttribute{
				Description: "Alternate HTTP proxy of scheme://hostname or scheme://hostname:port format",
				Optional:    true,
			},
			"max_retries": schema.Int64Attribute{
				Description: "Maximum number of retries to attempt before erroring out",
				Optional:    true,
			},
			"parallelism": schema.Int64Attribute{
				Description: "Number of concurrent requests to make within a resource where bulk operations are not possible",
				Optional:    true,
			},
			"request_timeout": schema.Int64Attribute{
				Description: "Timeout for single request (in seconds) which is made to Zscaler",
				Optional:    true,
			},
			"min_wait_seconds": schema.Int64Attribute{
				Description: "Minimum wait in seconds between retry attempts when rate limited",
				Optional:    true,
			},
			"max_wait_seconds": schema.Int64Attribute{
				Description: "Maximum wait in seconds between retry attempts when rate limited",
				Optional:    true,
			},
		},
	}
}

func (p *ZCCProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewTrustedNetworkResource,
		resources.NewNotificationTemplateResource,
		resources.NewZIAPostureResource,
		resources.NewForwardingProfileResource,
		resources.NewFailOpenPolicyResource,
		resources.NewDeviceCleanupResource,
		resources.NewWebPrivacyResource,
	}
}

func (p *ZCCProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewTrustedNetworkDataSource,
		datasources.NewNotificationTemplateDataSource,
		datasources.NewZIAPostureDataSource,
		datasources.NewAdminRolesDataSource,
		datasources.NewForwardingProfileDataSource,
		datasources.NewAdminUserDataSource,
		datasources.NewDevicesDataSource,
		datasources.NewCustomIPAppsDataSource,
		datasources.NewPredefinedIPAppsDataSource,
		datasources.NewProcessBasedAppsDataSource,
		datasources.NewWebAppServiceDataSource,
		datasources.NewFailOpenPolicyDataSource,
		datasources.NewDeviceCleanupDataSource,
		datasources.NewWebPrivacyDataSource,
		datasources.NewCompanyInfoDataSource,
	}
}

func New(version string) provider.Provider {
	return &ZCCProvider{
		version: version,
	}
}
