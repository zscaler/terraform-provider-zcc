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

// Ensure ZCCProvider satisfies the provider interface.
var _ provider.Provider = &ZCCProvider{}

// ZCCProvider defines the provider implementation.
type ZCCProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ZCCProviderModel describes the provider data model. Authentication
// is OneAPI (Zidentity) only — the legacy ZCC V2 client
// (zcc_client_id / zcc_client_secret / zcc_cloud /
// use_legacy_client) has been removed; existing configurations
// referencing those attributes must migrate to ZSCALER_CLIENT_ID /
// ZSCALER_CLIENT_SECRET (or ZSCALER_PRIVATE_KEY) + ZSCALER_VANITY_DOMAIN.
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
		// Singleton: ZCC device cleanup (GET getDeviceCleanupInfo / PUT setDeviceCleanupInfo).
		resources.NewDeviceCleanupResource,
		resources.NewWebAppServiceResource,
		resources.NewWebPrivacyResource,
		// Per-OS app profiles (zcc_app_profile_macos, _ios, _windows,
		// _linux, _android) backed by /web/policy/edit are intentionally
		// deregistered. The underlying singleton API is unstable —
		// success responses depend on undocumented field/type
		// combinations that vary per OS and per UI capture — so the
		// resources are parked under local_dev/Backup_Config_Future
		// until the API contract is stabilised upstream.
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
		// datasources.NewApplicationProfilesDataSource — deregistered with the per-OS app_profile resources; the underlying /web/policy APIs are still being stabilised upstream.
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
