package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/company"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &CompanyInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &CompanyInfoDataSource{}
)

func NewCompanyInfoDataSource() datasource.DataSource {
	return &CompanyInfoDataSource{}
}

type CompanyInfoDataSource struct {
	client *client.Client
}

type CompanyInfoDataSourceModel struct {
	ID                                       types.String `tfsdk:"id"`
	OrgID                                    types.String `tfsdk:"org_id"`
	MasterCustomerID                         types.String `tfsdk:"master_customer_id"`
	Name                                     types.String `tfsdk:"name"`
	BusinessName                             types.String `tfsdk:"business_name"`
	BusinessContactNumber                    types.String `tfsdk:"business_contact_number"`
	ActivationRecipient                      types.String `tfsdk:"activation_recipient"`
	ActivationCopy                           types.String `tfsdk:"activation_copy"`
	MdmStatus                                types.String `tfsdk:"mdm_status"`
	SendEmail                                types.String `tfsdk:"send_email"`
	ProxyEnabled                             types.String `tfsdk:"proxy_enabled"`
	ZpnEnabled                               types.String `tfsdk:"zpn_enabled"`
	UpmEnabled                               types.String `tfsdk:"upm_enabled"`
	ZadEnabled                               types.String `tfsdk:"zad_enabled"`
	EnableDeceptionForAll                    types.String `tfsdk:"enable_deception_for_all"`
	DlpEnabled                               types.String `tfsdk:"dlp_enabled"`
	TunnelProtocolType                       types.String `tfsdk:"tunnel_protocol_type"`
	SecureAgentBasic                         types.String `tfsdk:"secure_agent_basic"`
	SecureAgentAdvanced                      types.String `tfsdk:"secure_agent_advanced"`
	SupportAdminEmail                        types.String `tfsdk:"support_admin_email"`
	SupportEnabled                           types.Int64  `tfsdk:"support_enabled"`
	FetchLogsForAdminsEnabled                types.Int64  `tfsdk:"fetch_logs_for_admins_enabled"`
	EnableRectifyUtils                       types.Int64  `tfsdk:"enable_rectify_utils"`
	SupportTicketEnabled                     types.Int64  `tfsdk:"support_ticket_enabled"`
	DisableLoggingControls                   types.Int64  `tfsdk:"disable_logging_controls"`
	DefaultAuthType                          types.Int64  `tfsdk:"default_auth_type"`
	Version                                  types.String `tfsdk:"version"`
	PolicyActivationRequired                 types.Int64  `tfsdk:"policy_activation_required"`
	EnableAutofillUsername                   types.Int64  `tfsdk:"enable_autofill_username"`
	AutoFillUsingLoginHint                   types.Int64  `tfsdk:"auto_fill_using_login_hint"`
	DcServiceReadOnly                        types.Int64  `tfsdk:"dc_service_read_only"`
	EnableTunnelZappTrafficToggle            types.String `tfsdk:"enable_tunnel_zapp_traffic_toggle"`
	MachineIdpAuth                           types.String `tfsdk:"machine_idp_auth"`
	LinuxVisibility                          types.String `tfsdk:"linux_visibility"`
	RegistryPathForPac                       types.String `tfsdk:"registry_path_for_pac"`
	UsePollsetForSocketReactor               types.String `tfsdk:"use_pollset_for_socket_reactor"`
	EnableDtlsForZpa                         types.String `tfsdk:"enable_dtls_for_zpa"`
	UseV8JsEngine                            types.String `tfsdk:"use_v8_js_engine"`
	DisableParallelIpv4AndIPv6               types.String `tfsdk:"disable_parallel_ipv4_and_ipv6"`
	Send64BitBuild                           types.String `tfsdk:"send_64bit_build"`
	UseAddIfscopeRoute                       types.String `tfsdk:"use_add_ifscope_route"`
	UseClearArpCache                         types.String `tfsdk:"use_clear_arp_cache"`
	UseDnsPriorityOrdering                   types.String `tfsdk:"use_dns_priority_ordering"`
	EnableBrowserAuth                        types.String `tfsdk:"enable_browser_auth"`
	EnablePublicAPI                          types.String `tfsdk:"enable_public_api"`
	DisableReasonVisibility                  types.String `tfsdk:"disable_reason_visibility"`
	FollowRoutingTable                       types.String `tfsdk:"follow_routing_table"`
	UseDefaultAdapterForDNS                  types.String `tfsdk:"use_default_adapter_for_dns"`
	EnableMinimumDeviceCleanupAsOne          types.String `tfsdk:"enable_minimum_device_cleanup_as_one"`
	DnsPriorityOrderingForTrustedDnsCriteria types.String `tfsdk:"dns_priority_ordering_for_trusted_dns_criteria"`
	MachineTunnelPosture                     types.String `tfsdk:"machine_tunnel_posture"`
	ZpaPartnerLogin                          types.String `tfsdk:"zpa_partner_login"`
	ProxyPort                                types.Int64  `tfsdk:"proxy_port"`
	DnsCacheTtlWindows                       types.Int64  `tfsdk:"dns_cache_ttl_windows"`
	DnsCacheTtlMac                           types.Int64  `tfsdk:"dns_cache_ttl_mac"`
	DnsCacheTtlAndroid                       types.Int64  `tfsdk:"dns_cache_ttl_android"`
	DnsCacheTtlIos                           types.Int64  `tfsdk:"dns_cache_ttl_ios"`
	DnsCacheTtlLinux                         types.Int64  `tfsdk:"dns_cache_ttl_linux"`
	ZpaClientCertExpInDays                   types.Int64  `tfsdk:"zpa_client_cert_exp_in_days"`
	EnableFlowLogger                         types.String `tfsdk:"enable_flow_logger"`
	FlowLoggingBufferLimit                   types.Int64  `tfsdk:"flow_logging_buffer_limit"`
	FlowLoggingTimeInterval                  types.Int64  `tfsdk:"flow_logging_time_interval"`
	PostureBasedService                      types.String `tfsdk:"posture_based_service"`
	EnablePostureBasedProfile                types.String `tfsdk:"enable_posture_based_profile"`
	DisasterRecovery                         types.String `tfsdk:"disaster_recovery"`
	ZiaGlobalDbUrlForDR                      types.String `tfsdk:"zia_global_db_url_for_dr"`
	EnableReactUI                            types.String `tfsdk:"enable_react_ui"`
	LaunchReactUIbyDefault                   types.String `tfsdk:"launch_react_ui_by_default"`
	DlpNotification                          types.String `tfsdk:"dlp_notification"`
	VpnGatewayCharLimit                      types.Int64  `tfsdk:"vpn_gateway_char_limit"`
	DeviceGroupsCount                        types.Int64  `tfsdk:"device_groups_count"`
	VpnBypassRefreshInterval                 types.Int64  `tfsdk:"vpn_bypass_refresh_interval"`
	DestIncludeExcludeCharLimit              types.Int64  `tfsdk:"dest_include_exclude_char_limit"`
	IpV6SupportForTunnel2                    types.String `tfsdk:"ipv6_support_for_tunnel2"`
	DestIncludeExcludeCharLimitForIpv6       types.Int64  `tfsdk:"dest_include_exclude_char_limit_for_ipv6"`
	EnableSetProxyOnVPNAdapters              types.String `tfsdk:"enable_set_proxy_on_vpn_adapters"`
	DisableDNSRouteExclusion                 types.String `tfsdk:"disable_dns_route_exclusion"`
	ShowVPNTunNotification                   types.String `tfsdk:"show_vpn_tun_notification"`
	AddAppBypassToVPNGateway                 types.String `tfsdk:"add_app_bypass_to_vpn_gateway"`
	EnableZscalerFirewall                    types.String `tfsdk:"enable_zscaler_firewall"`
	PersistentZscalerFirewall                types.String `tfsdk:"persistent_zscaler_firewall"`
	ClearMupCache                            types.String `tfsdk:"clear_mup_cache"`
	ExecuteGpoUpdate                         types.String `tfsdk:"execute_gpo_update"`
	EnablePortBasedZPAFilter                 types.String `tfsdk:"enable_port_based_zpa_filter"`
	EnableAntiTampering                      types.String `tfsdk:"enable_anti_tampering"`
	ZpaReauthEnabled                         types.Int64  `tfsdk:"zpa_reauth_enabled"`
	ZpaAutoReauthTimeout                     types.Int64  `tfsdk:"zpa_auto_reauth_timeout"`
	EnableZpaAuthUserName                    types.Int64  `tfsdk:"enable_zpa_auth_user_name"`
	EnableGlobalZCCTelemetry                 types.Int64  `tfsdk:"enable_global_zcc_telemetry"`
	ConfigureTunnel2fallbackForZia           types.String `tfsdk:"configure_tunnel2_fallback_for_zia"`
	WebAppConfig                             types.Object `tfsdk:"web_app_config"`
	EnableInstallWebView2                    types.String `tfsdk:"enable_install_webview2"`
	EnableCustomProxyPorts                   types.String `tfsdk:"enable_custom_proxy_ports"`
	InterceptZIATrafficAllAdapters           types.String `tfsdk:"intercept_zia_traffic_all_adapters"`
	SwaggerLink                              types.String `tfsdk:"swagger_link"`
	EnableOneIdAdmin                         types.String `tfsdk:"enable_one_id_admin"`
	EnableOneIdUser                          types.String `tfsdk:"enable_one_id_user"`
	RestrictAdminAccess                      types.String `tfsdk:"restrict_admin_access"`
	EnableZiaUserDepartmentSync              types.String `tfsdk:"enable_zia_user_department_sync"`
	EnableUDPTransportSelection              types.String `tfsdk:"enable_udp_transport_selection"`
	ComputeDeviceGroupsForZIA                types.String `tfsdk:"compute_device_groups_for_zia"`
	ComputeDeviceGroupsForZPA                types.String `tfsdk:"compute_device_groups_for_zpa"`
	ComputeDeviceGroupsForZDX                types.String `tfsdk:"compute_device_groups_for_zdx"`
	ComputeDeviceGroupsForZAD                types.String `tfsdk:"compute_device_groups_for_zad"`
	UseTunnel2SmeForTunnel1                  types.String `tfsdk:"use_tunnel2_sme_for_tunnel1"`
	MaCloudName                              types.String `tfsdk:"ma_cloud_name"`
	ZiaCloudName                             types.String `tfsdk:"zia_cloud_name"`
	Zt2HealthProbeInterval                   types.Int64  `tfsdk:"zt2_health_probe_interval"`
	DevicePostureFrequency                   types.List   `tfsdk:"device_posture_frequency"`
	ZdxManualRollout                         types.String `tfsdk:"zdx_manual_rollout"`
	WinZdxLiteEnabled                        types.String `tfsdk:"win_zdx_lite_enabled"`
	TelemetryDefault                         types.Int64  `tfsdk:"telemetry_default"`
}

func (d *CompanyInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_company_info"
}

func (d *CompanyInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the full ZCC company configuration returned by the GET /zcc/papi/public/v1/getCompanyInfo endpoint. This is a singleton data source: there is exactly one company info record per tenant, so no filters are required.",
		Attributes: map[string]schema.Attribute{
			"id":                                   schema.StringAttribute{Description: "Synthetic identifier for the data source (mirrors org_id when present).", Computed: true},
			"org_id":                               schema.StringAttribute{Description: "Organization identifier returned by the API.", Computed: true},
			"master_customer_id":                   schema.StringAttribute{Description: "Master customer identifier.", Computed: true},
			"name":                                 schema.StringAttribute{Description: "Internal name of the organization.", Computed: true},
			"business_name":                        schema.StringAttribute{Description: "Business display name.", Computed: true},
			"business_contact_number":              schema.StringAttribute{Description: "Primary business contact phone number.", Computed: true},
			"activation_recipient":                 schema.StringAttribute{Description: "Email recipient for activation notifications.", Computed: true},
			"activation_copy":                      schema.StringAttribute{Description: "Email cc list for activation notifications.", Computed: true},
			"mdm_status":                           schema.StringAttribute{Description: "MDM status flag.", Computed: true},
			"send_email":                           schema.StringAttribute{Description: "Whether activation emails are sent.", Computed: true},
			"proxy_enabled":                        schema.StringAttribute{Description: "Whether proxy mode is enabled.", Computed: true},
			"zpn_enabled":                          schema.StringAttribute{Description: "Whether ZPA is enabled for the tenant.", Computed: true},
			"upm_enabled":                          schema.StringAttribute{Description: "Whether UPM is enabled.", Computed: true},
			"zad_enabled":                          schema.StringAttribute{Description: "Whether Zscaler Active Defense is enabled.", Computed: true},
			"enable_deception_for_all":             schema.StringAttribute{Description: "Whether deception is enabled for all users.", Computed: true},
			"dlp_enabled":                          schema.StringAttribute{Description: "Whether DLP is enabled.", Computed: true},
			"tunnel_protocol_type":                 schema.StringAttribute{Description: "Tunnel protocol type.", Computed: true},
			"secure_agent_basic":                   schema.StringAttribute{Description: "Secure agent basic flag.", Computed: true},
			"secure_agent_advanced":                schema.StringAttribute{Description: "Secure agent advanced flag.", Computed: true},
			"support_admin_email":                  schema.StringAttribute{Description: "Support administrator email address.", Computed: true},
			"support_enabled":                      schema.Int64Attribute{Description: "Whether support access is enabled.", Computed: true},
			"fetch_logs_for_admins_enabled":        schema.Int64Attribute{Description: "Whether admins can fetch logs.", Computed: true},
			"enable_rectify_utils":                 schema.Int64Attribute{Description: "Whether rectify utilities are enabled.", Computed: true},
			"support_ticket_enabled":               schema.Int64Attribute{Description: "Whether support ticket creation is enabled.", Computed: true},
			"disable_logging_controls":             schema.Int64Attribute{Description: "Whether logging controls are disabled.", Computed: true},
			"default_auth_type":                    schema.Int64Attribute{Description: "Default authentication type.", Computed: true},
			"version":                              schema.StringAttribute{Description: "Configuration version.", Computed: true},
			"policy_activation_required":           schema.Int64Attribute{Description: "Whether policy activation is required.", Computed: true},
			"enable_autofill_username":             schema.Int64Attribute{Description: "Whether autofill username is enabled.", Computed: true},
			"auto_fill_using_login_hint":           schema.Int64Attribute{Description: "Whether autofill uses the login hint.", Computed: true},
			"dc_service_read_only":                 schema.Int64Attribute{Description: "Whether the DC service is read-only.", Computed: true},
			"enable_tunnel_zapp_traffic_toggle":    schema.StringAttribute{Description: "Whether tunneling of ZApp traffic toggle is enabled.", Computed: true},
			"machine_idp_auth":                     schema.StringAttribute{Description: "Machine IdP authentication setting.", Computed: true},
			"linux_visibility":                     schema.StringAttribute{Description: "Linux visibility setting.", Computed: true},
			"registry_path_for_pac":                schema.StringAttribute{Description: "Windows registry path used to source PAC URLs.", Computed: true},
			"use_pollset_for_socket_reactor":       schema.StringAttribute{Description: "Whether to use pollset for the socket reactor.", Computed: true},
			"enable_dtls_for_zpa":                  schema.StringAttribute{Description: "Whether DTLS is enabled for ZPA.", Computed: true},
			"use_v8_js_engine":                     schema.StringAttribute{Description: "Whether the V8 JS engine is enabled for PAC evaluation.", Computed: true},
			"disable_parallel_ipv4_and_ipv6":       schema.StringAttribute{Description: "Whether parallel IPv4/IPv6 resolution is disabled.", Computed: true},
			"send_64bit_build":                     schema.StringAttribute{Description: "Whether to send the 64-bit build.", Computed: true},
			"use_add_ifscope_route":                schema.StringAttribute{Description: "Whether to use 'add ifscope' route on macOS.", Computed: true},
			"use_clear_arp_cache":                  schema.StringAttribute{Description: "Whether to clear the ARP cache on network changes.", Computed: true},
			"use_dns_priority_ordering":            schema.StringAttribute{Description: "Whether to use DNS priority ordering.", Computed: true},
			"enable_browser_auth":                  schema.StringAttribute{Description: "Whether browser-based authentication is enabled.", Computed: true},
			"enable_public_api":                    schema.StringAttribute{Description: "Whether the public API is enabled.", Computed: true},
			"disable_reason_visibility":            schema.StringAttribute{Description: "Whether the disable-reason field is visible to end users.", Computed: true},
			"follow_routing_table":                 schema.StringAttribute{Description: "Whether to follow the system routing table.", Computed: true},
			"use_default_adapter_for_dns":          schema.StringAttribute{Description: "Whether to use the default adapter for DNS.", Computed: true},
			"enable_minimum_device_cleanup_as_one": schema.StringAttribute{Description: "Whether to enforce a one-day minimum device cleanup window.", Computed: true},
			"dns_priority_ordering_for_trusted_dns_criteria": schema.StringAttribute{Description: "DNS priority ordering for trusted DNS criteria evaluation.", Computed: true},
			"machine_tunnel_posture":                         schema.StringAttribute{Description: "Machine tunnel posture setting.", Computed: true},
			"zpa_partner_login":                              schema.StringAttribute{Description: "Whether ZPA partner login is allowed.", Computed: true},
			"proxy_port":                                     schema.Int64Attribute{Description: "Local proxy listening port.", Computed: true},
			"dns_cache_ttl_windows":                          schema.Int64Attribute{Description: "DNS cache TTL on Windows in seconds.", Computed: true},
			"dns_cache_ttl_mac":                              schema.Int64Attribute{Description: "DNS cache TTL on macOS in seconds.", Computed: true},
			"dns_cache_ttl_android":                          schema.Int64Attribute{Description: "DNS cache TTL on Android in seconds.", Computed: true},
			"dns_cache_ttl_ios":                              schema.Int64Attribute{Description: "DNS cache TTL on iOS in seconds.", Computed: true},
			"dns_cache_ttl_linux":                            schema.Int64Attribute{Description: "DNS cache TTL on Linux in seconds.", Computed: true},
			"zpa_client_cert_exp_in_days":                    schema.Int64Attribute{Description: "ZPA client certificate expiration in days.", Computed: true},
			"enable_flow_logger":                             schema.StringAttribute{Description: "Whether flow logging is enabled.", Computed: true},
			"flow_logging_buffer_limit":                      schema.Int64Attribute{Description: "Flow logging buffer size limit.", Computed: true},
			"flow_logging_time_interval":                     schema.Int64Attribute{Description: "Flow logging time interval in seconds.", Computed: true},
			"posture_based_service":                          schema.StringAttribute{Description: "Whether posture-based service is enabled.", Computed: true},
			"enable_posture_based_profile":                   schema.StringAttribute{Description: "Whether posture-based profiles are enabled.", Computed: true},
			"disaster_recovery":                              schema.StringAttribute{Description: "Disaster recovery setting.", Computed: true},
			"zia_global_db_url_for_dr":                       schema.StringAttribute{Description: "ZIA global database URL for disaster recovery.", Computed: true},
			"enable_react_ui":                                schema.StringAttribute{Description: "Whether the React UI is enabled.", Computed: true},
			"launch_react_ui_by_default":                     schema.StringAttribute{Description: "Whether the React UI launches by default.", Computed: true},
			"dlp_notification":                               schema.StringAttribute{Description: "DLP notification setting.", Computed: true},
			"vpn_gateway_char_limit":                         schema.Int64Attribute{Description: "Maximum length of the VPN gateway field.", Computed: true},
			"device_groups_count":                            schema.Int64Attribute{Description: "Number of configured device groups.", Computed: true},
			"vpn_bypass_refresh_interval":                    schema.Int64Attribute{Description: "VPN bypass refresh interval in seconds.", Computed: true},
			"dest_include_exclude_char_limit":                schema.Int64Attribute{Description: "Character limit for destination include/exclude lists.", Computed: true},
			"ipv6_support_for_tunnel2":                       schema.StringAttribute{Description: "IPv6 support flag for Tunnel 2.0.", Computed: true},
			"dest_include_exclude_char_limit_for_ipv6":       schema.Int64Attribute{Description: "Character limit for destination include/exclude lists for IPv6.", Computed: true},
			"enable_set_proxy_on_vpn_adapters":               schema.StringAttribute{Description: "Whether to set proxy on VPN adapters.", Computed: true},
			"disable_dns_route_exclusion":                    schema.StringAttribute{Description: "Whether DNS route exclusion is disabled.", Computed: true},
			"show_vpn_tun_notification":                      schema.StringAttribute{Description: "Whether to show VPN tunnel notifications.", Computed: true},
			"add_app_bypass_to_vpn_gateway":                  schema.StringAttribute{Description: "Whether to add app bypass entries to the VPN gateway.", Computed: true},
			"enable_zscaler_firewall":                        schema.StringAttribute{Description: "Whether the Zscaler Firewall is enabled.", Computed: true},
			"persistent_zscaler_firewall":                    schema.StringAttribute{Description: "Whether the Zscaler Firewall is persistent.", Computed: true},
			"clear_mup_cache":                                schema.StringAttribute{Description: "Whether to clear the MUP cache.", Computed: true},
			"execute_gpo_update":                             schema.StringAttribute{Description: "Whether to execute Group Policy update.", Computed: true},
			"enable_port_based_zpa_filter":                   schema.StringAttribute{Description: "Whether to enable port-based ZPA filtering.", Computed: true},
			"enable_anti_tampering":                          schema.StringAttribute{Description: "Whether anti-tampering is enabled.", Computed: true},
			"zpa_reauth_enabled":                             schema.Int64Attribute{Description: "Whether ZPA re-authentication is enabled.", Computed: true},
			"zpa_auto_reauth_timeout":                        schema.Int64Attribute{Description: "ZPA automatic re-authentication timeout in minutes.", Computed: true},
			"enable_zpa_auth_user_name":                      schema.Int64Attribute{Description: "Whether to use the ZPA auth user name.", Computed: true},
			"enable_global_zcc_telemetry":                    schema.Int64Attribute{Description: "Whether global ZCC telemetry is enabled.", Computed: true},
			"configure_tunnel2_fallback_for_zia":             schema.StringAttribute{Description: "Whether to configure Tunnel 2.0 fallback for ZIA.", Computed: true},
			"web_app_config": schema.SingleNestedAttribute{
				Description: "Web/UI feature visibility configuration. All values are returned by the API as strings (typically \"0\"/\"1\" or feature codes).",
				Computed:    true,
				Attributes:  webAppConfigSchemaAttributes(),
			},
			"enable_install_webview2":            schema.StringAttribute{Description: "Whether WebView2 installation is enabled.", Computed: true},
			"enable_custom_proxy_ports":          schema.StringAttribute{Description: "Whether custom proxy ports are enabled.", Computed: true},
			"intercept_zia_traffic_all_adapters": schema.StringAttribute{Description: "Whether to intercept ZIA traffic across all adapters.", Computed: true},
			"swagger_link":                       schema.StringAttribute{Description: "Swagger documentation link.", Computed: true},
			"enable_one_id_admin":                schema.StringAttribute{Description: "Whether OneID is enabled for admins.", Computed: true},
			"enable_one_id_user":                 schema.StringAttribute{Description: "Whether OneID is enabled for users.", Computed: true},
			"restrict_admin_access":              schema.StringAttribute{Description: "Whether admin access is restricted.", Computed: true},
			"enable_zia_user_department_sync":    schema.StringAttribute{Description: "Whether ZIA user/department sync is enabled.", Computed: true},
			"enable_udp_transport_selection":     schema.StringAttribute{Description: "Whether UDP transport selection is enabled.", Computed: true},
			"compute_device_groups_for_zia":      schema.StringAttribute{Description: "Whether device groups are computed for ZIA.", Computed: true},
			"compute_device_groups_for_zpa":      schema.StringAttribute{Description: "Whether device groups are computed for ZPA.", Computed: true},
			"compute_device_groups_for_zdx":      schema.StringAttribute{Description: "Whether device groups are computed for ZDX.", Computed: true},
			"compute_device_groups_for_zad":      schema.StringAttribute{Description: "Whether device groups are computed for ZAD.", Computed: true},
			"use_tunnel2_sme_for_tunnel1":        schema.StringAttribute{Description: "Whether to use Tunnel 2.0 SME for Tunnel 1.0.", Computed: true},
			"ma_cloud_name":                      schema.StringAttribute{Description: "Mobile Admin cloud name.", Computed: true},
			"zia_cloud_name":                     schema.StringAttribute{Description: "ZIA cloud name.", Computed: true},
			"zt2_health_probe_interval":          schema.Int64Attribute{Description: "Tunnel 2.0 health probe interval in seconds.", Computed: true},
			"device_posture_frequency": schema.ListNestedAttribute{
				Description: "Per-platform posture evaluation frequency overrides.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"posture_id":    schema.Int64Attribute{Description: "Posture identifier.", Computed: true},
						"posture_name":  schema.StringAttribute{Description: "Posture name.", Computed: true},
						"ios_value":     schema.Int64Attribute{Description: "iOS evaluation interval.", Computed: true},
						"android_value": schema.Int64Attribute{Description: "Android evaluation interval.", Computed: true},
						"windows_value": schema.Int64Attribute{Description: "Windows evaluation interval.", Computed: true},
						"mac_value":     schema.Int64Attribute{Description: "macOS evaluation interval.", Computed: true},
						"linux_value":   schema.Int64Attribute{Description: "Linux evaluation interval.", Computed: true},
						"default_value": schema.Int64Attribute{Description: "Default evaluation interval.", Computed: true},
					},
				},
			},
			"zdx_manual_rollout":   schema.StringAttribute{Description: "Whether ZDX manual rollout is enabled.", Computed: true},
			"win_zdx_lite_enabled": schema.StringAttribute{Description: "Whether Windows ZDX Lite is enabled.", Computed: true},
			"telemetry_default":    schema.Int64Attribute{Description: "Default telemetry setting.", Computed: true},
		},
	}
}

func (d *CompanyInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *CompanyInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading ZCC company info")

	info, err := company.GetCompanyInfo(ctx, d.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read company info: %v", err))
		return
	}

	webAppCfg, diags := webAppConfigToObject(ctx, info.WebAppConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	postureList, diags := devicePostureFrequencyToList(ctx, info.DevicePostureFrequency)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := info.OrgID
	if id == "" {
		id = info.MasterCustomerID
	}
	if id == "" {
		id = "company_info"
	}

	m := CompanyInfoDataSourceModel{
		ID:                                       types.StringValue(id),
		OrgID:                                    types.StringValue(info.OrgID),
		MasterCustomerID:                         types.StringValue(info.MasterCustomerID),
		Name:                                     types.StringValue(info.Name),
		BusinessName:                             types.StringValue(info.BusinessName),
		BusinessContactNumber:                    types.StringValue(info.BusinessContactNumber),
		ActivationRecipient:                      types.StringValue(info.ActivationRecipient),
		ActivationCopy:                           types.StringValue(info.ActivationCopy),
		MdmStatus:                                types.StringValue(info.MdmStatus),
		SendEmail:                                types.StringValue(info.SendEmail),
		ProxyEnabled:                             types.StringValue(info.ProxyEnabled),
		ZpnEnabled:                               types.StringValue(info.ZpnEnabled),
		UpmEnabled:                               types.StringValue(info.UpmEnabled),
		ZadEnabled:                               types.StringValue(info.ZadEnabled),
		EnableDeceptionForAll:                    types.StringValue(info.EnableDeceptionForAll),
		DlpEnabled:                               types.StringValue(info.DlpEnabled),
		TunnelProtocolType:                       types.StringValue(info.TunnelProtocolType),
		SecureAgentBasic:                         types.StringValue(info.SecureAgentBasic),
		SecureAgentAdvanced:                      types.StringValue(info.SecureAgentAdvanced),
		SupportAdminEmail:                        types.StringValue(info.SupportAdminEmail),
		SupportEnabled:                           types.Int64Value(int64(info.SupportEnabled)),
		FetchLogsForAdminsEnabled:                types.Int64Value(int64(info.FetchLogsForAdminsEnabled)),
		EnableRectifyUtils:                       types.Int64Value(int64(info.EnableRectifyUtils)),
		SupportTicketEnabled:                     types.Int64Value(int64(info.SupportTicketEnabled)),
		DisableLoggingControls:                   types.Int64Value(int64(info.DisableLoggingControls)),
		DefaultAuthType:                          types.Int64Value(int64(info.DefaultAuthType)),
		Version:                                  types.StringValue(info.Version),
		PolicyActivationRequired:                 types.Int64Value(int64(info.PolicyActivationRequired)),
		EnableAutofillUsername:                   types.Int64Value(int64(info.EnableAutofillUsername)),
		AutoFillUsingLoginHint:                   types.Int64Value(int64(info.AutoFillUsingLoginHint)),
		DcServiceReadOnly:                        types.Int64Value(int64(info.DcServiceReadOnly)),
		EnableTunnelZappTrafficToggle:            types.StringValue(info.EnableTunnelZappTrafficToggle),
		MachineIdpAuth:                           types.StringValue(info.MachineIdpAuth),
		LinuxVisibility:                          types.StringValue(info.LinuxVisibility),
		RegistryPathForPac:                       types.StringValue(info.RegistryPathForPac),
		UsePollsetForSocketReactor:               types.StringValue(info.UsePollsetForSocketReactor),
		EnableDtlsForZpa:                         types.StringValue(info.EnableDtlsForZpa),
		UseV8JsEngine:                            types.StringValue(info.UseV8JsEngine),
		DisableParallelIpv4AndIPv6:               types.StringValue(info.DisableParallelIpv4AndIPv6),
		Send64BitBuild:                           types.StringValue(info.Send64BitBuild),
		UseAddIfscopeRoute:                       types.StringValue(info.UseAddIfscopeRoute),
		UseClearArpCache:                         types.StringValue(info.UseClearArpCache),
		UseDnsPriorityOrdering:                   types.StringValue(info.UseDnsPriorityOrdering),
		EnableBrowserAuth:                        types.StringValue(info.EnableBrowserAuth),
		EnablePublicAPI:                          types.StringValue(info.EnablePublicAPI),
		DisableReasonVisibility:                  types.StringValue(info.DisableReasonVisibility),
		FollowRoutingTable:                       types.StringValue(info.FollowRoutingTable),
		UseDefaultAdapterForDNS:                  types.StringValue(info.UseDefaultAdapterForDNS),
		EnableMinimumDeviceCleanupAsOne:          types.StringValue(info.EnableMinimumDeviceCleanupAsOne),
		DnsPriorityOrderingForTrustedDnsCriteria: types.StringValue(info.DnsPriorityOrderingForTrustedDnsCriteria),
		MachineTunnelPosture:                     types.StringValue(info.MachineTunnelPosture),
		ZpaPartnerLogin:                          types.StringValue(info.ZpaPartnerLogin),
		ProxyPort:                                types.Int64Value(int64(info.ProxyPort)),
		DnsCacheTtlWindows:                       types.Int64Value(int64(info.DnsCacheTtlWindows)),
		DnsCacheTtlMac:                           types.Int64Value(int64(info.DnsCacheTtlMac)),
		DnsCacheTtlAndroid:                       types.Int64Value(int64(info.DnsCacheTtlAndroid)),
		DnsCacheTtlIos:                           types.Int64Value(int64(info.DnsCacheTtlIos)),
		DnsCacheTtlLinux:                         types.Int64Value(int64(info.DnsCacheTtlLinux)),
		ZpaClientCertExpInDays:                   types.Int64Value(int64(info.ZpaClientCertExpInDays)),
		EnableFlowLogger:                         types.StringValue(info.EnableFlowLogger),
		FlowLoggingBufferLimit:                   types.Int64Value(int64(info.FlowLoggingBufferLimit)),
		FlowLoggingTimeInterval:                  types.Int64Value(int64(info.FlowLoggingTimeInterval)),
		PostureBasedService:                      types.StringValue(info.PostureBasedService),
		EnablePostureBasedProfile:                types.StringValue(info.EnablePostureBasedProfile),
		DisasterRecovery:                         types.StringValue(info.DisasterRecovery),
		ZiaGlobalDbUrlForDR:                      types.StringValue(info.ZiaGlobalDbUrlForDR),
		EnableReactUI:                            types.StringValue(info.EnableReactUI),
		LaunchReactUIbyDefault:                   types.StringValue(info.LaunchReactUIbyDefault),
		DlpNotification:                          types.StringValue(info.DlpNotification),
		VpnGatewayCharLimit:                      types.Int64Value(int64(info.VpnGatewayCharLimit)),
		DeviceGroupsCount:                        types.Int64Value(int64(info.DeviceGroupsCount)),
		VpnBypassRefreshInterval:                 types.Int64Value(int64(info.VpnBypassRefreshInterval)),
		DestIncludeExcludeCharLimit:              types.Int64Value(int64(info.DestIncludeExcludeCharLimit)),
		IpV6SupportForTunnel2:                    types.StringValue(info.IpV6SupportForTunnel2),
		DestIncludeExcludeCharLimitForIpv6:       types.Int64Value(int64(info.DestIncludeExcludeCharLimitForIpv6)),
		EnableSetProxyOnVPNAdapters:              types.StringValue(info.EnableSetProxyOnVPNAdapters),
		DisableDNSRouteExclusion:                 types.StringValue(info.DisableDNSRouteExclusion),
		ShowVPNTunNotification:                   types.StringValue(info.ShowVPNTunNotification),
		AddAppBypassToVPNGateway:                 types.StringValue(info.AddAppBypassToVPNGateway),
		EnableZscalerFirewall:                    types.StringValue(info.EnableZscalerFirewall),
		PersistentZscalerFirewall:                types.StringValue(info.PersistentZscalerFirewall),
		ClearMupCache:                            types.StringValue(info.ClearMupCache),
		ExecuteGpoUpdate:                         types.StringValue(info.ExecuteGpoUpdate),
		EnablePortBasedZPAFilter:                 types.StringValue(info.EnablePortBasedZPAFilter),
		EnableAntiTampering:                      types.StringValue(info.EnableAntiTampering),
		ZpaReauthEnabled:                         types.Int64Value(int64(info.ZpaReauthEnabled)),
		ZpaAutoReauthTimeout:                     types.Int64Value(int64(info.ZpaAutoReauthTimeout)),
		EnableZpaAuthUserName:                    types.Int64Value(int64(info.EnableZpaAuthUserName)),
		EnableGlobalZCCTelemetry:                 types.Int64Value(int64(info.EnableGlobalZCCTelemetry)),
		ConfigureTunnel2fallbackForZia:           types.StringValue(info.ConfigureTunnel2fallbackForZia),
		WebAppConfig:                             webAppCfg,
		EnableInstallWebView2:                    types.StringValue(info.EnableInstallWebView2),
		EnableCustomProxyPorts:                   types.StringValue(info.EnableCustomProxyPorts),
		InterceptZIATrafficAllAdapters:           types.StringValue(info.InterceptZIATrafficAllAdapters),
		SwaggerLink:                              types.StringValue(info.SwaggerLink),
		EnableOneIdAdmin:                         types.StringValue(info.EnableOneIdAdmin),
		EnableOneIdUser:                          types.StringValue(info.EnableOneIdUser),
		RestrictAdminAccess:                      types.StringValue(info.RestrictAdminAccess),
		EnableZiaUserDepartmentSync:              types.StringValue(info.EnableZiaUserDepartmentSync),
		EnableUDPTransportSelection:              types.StringValue(info.EnableUDPTransportSelection),
		ComputeDeviceGroupsForZIA:                types.StringValue(info.ComputeDeviceGroupsForZIA),
		ComputeDeviceGroupsForZPA:                types.StringValue(info.ComputeDeviceGroupsForZPA),
		ComputeDeviceGroupsForZDX:                types.StringValue(info.ComputeDeviceGroupsForZDX),
		ComputeDeviceGroupsForZAD:                types.StringValue(info.ComputeDeviceGroupsForZAD),
		UseTunnel2SmeForTunnel1:                  types.StringValue(info.UseTunnel2SmeForTunnel1),
		MaCloudName:                              types.StringValue(info.MaCloudName),
		ZiaCloudName:                             types.StringValue(info.ZiaCloudName),
		Zt2HealthProbeInterval:                   types.Int64Value(int64(info.Zt2HealthProbeInterval)),
		DevicePostureFrequency:                   postureList,
		ZdxManualRollout:                         types.StringValue(info.ZdxManualRollout),
		WinZdxLiteEnabled:                        types.StringValue(info.WinZdxLiteEnabled),
		TelemetryDefault:                         types.Int64Value(int64(info.TelemetryDefault)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// devicePostureFrequencyAttrTypes describes the schema of a single posture
// frequency entry; kept separate so it can be reused when constructing the
// list value during Read.
func devicePostureFrequencyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"posture_id":    types.Int64Type,
		"posture_name":  types.StringType,
		"ios_value":     types.Int64Type,
		"android_value": types.Int64Type,
		"windows_value": types.Int64Type,
		"mac_value":     types.Int64Type,
		"linux_value":   types.Int64Type,
		"default_value": types.Int64Type,
	}
}

func devicePostureFrequencyToList(_ context.Context, items []company.DevicePostureFrequency) (types.List, diag.Diagnostics) {
	objType := types.ObjectType{AttrTypes: devicePostureFrequencyAttrTypes()}
	if len(items) == 0 {
		return types.ListValue(objType, []attr.Value{})
	}
	values := make([]attr.Value, 0, len(items))
	for _, p := range items {
		obj, diags := types.ObjectValue(devicePostureFrequencyAttrTypes(), map[string]attr.Value{
			"posture_id":    types.Int64Value(int64(p.PostureID)),
			"posture_name":  types.StringValue(p.PostureName),
			"ios_value":     types.Int64Value(int64(p.IosValue)),
			"android_value": types.Int64Value(int64(p.AndroidValue)),
			"windows_value": types.Int64Value(int64(p.WindowsValue)),
			"mac_value":     types.Int64Value(int64(p.MacValue)),
			"linux_value":   types.Int64Value(int64(p.LinuxValue)),
			"default_value": types.Int64Value(int64(p.DefaultValue)),
		})
		if diags.HasError() {
			return types.ListNull(objType), diags
		}
		values = append(values, obj)
	}
	return types.ListValue(objType, values)
}

// =============================================================================
// web_app_config nested object helpers
// =============================================================================
//
// The `web_app_config` SingleNestedAttribute is part of this data source's
// schema (see Schema above). The 185-entry table below is the single source
// of truth for every field exposed under that block: webAppConfigAttrTypes
// and webAppConfigSchemaAttributes both derive their output from it, and
// webAppConfigToObject flattens the SDK's *company.WebAppConfig pointer
// into the corresponding types.Object during Read. Adding or removing a
// field on the SDK only requires a single new entry here.

// webAppConfigStringFields maps each Terraform attribute name (snake_case)
// to a getter that pulls the matching string field out of
// *company.WebAppConfig. The data source is read-only so only a getter is
// needed.
var webAppConfigStringFields = []struct {
	name string
	get  func(*company.WebAppConfig) string
}{
	{"enable_fips_mode", func(p *company.WebAppConfig) string { return p.EnableFipsMode }},
	{"device_cleanup", func(p *company.WebAppConfig) string { return p.DeviceCleanup }},
	{"sync_time_hours", func(p *company.WebAppConfig) string { return p.SyncTimeHours }},
	{"hide_non_fed_settings", func(p *company.WebAppConfig) string { return p.HideNonFedSettings }},
	{"hide_audit_logs", func(p *company.WebAppConfig) string { return p.HideAuditLogs }},
	{"activate_policy", func(p *company.WebAppConfig) string { return p.ActivatePolicy }},
	{"trusted_network", func(p *company.WebAppConfig) string { return p.TrustedNetwork }},
	{"process_postures", func(p *company.WebAppConfig) string { return p.ProcessPostures }},
	{"zpa_reauth", func(p *company.WebAppConfig) string { return p.ZpaReauth }},
	{"inactive_device_cleanup", func(p *company.WebAppConfig) string { return p.InactiveDeviceCleanup }},
	{"zpa_auth_username", func(p *company.WebAppConfig) string { return p.ZpaAuthUsername }},
	{"machine_tunnel", func(p *company.WebAppConfig) string { return p.MachineTunnel }},
	{"cache_system_proxy", func(p *company.WebAppConfig) string { return p.CacheSystemProxy }},
	{"hide_dtls_support_settings", func(p *company.WebAppConfig) string { return p.HideDTLSSupportSettings }},
	{"machine_token", func(p *company.WebAppConfig) string { return p.MachineToken }},
	{"application_bypass_info", func(p *company.WebAppConfig) string { return p.ApplicationBypassInfo }},
	{"tunnel_two_for_android_devices", func(p *company.WebAppConfig) string { return p.TunnelTwoForAndroidDevices }},
	{"tunnel_two_for_ios_devices", func(p *company.WebAppConfig) string { return p.TunnelTwoForiOSDevices }},
	{"ownership_variable_posture", func(p *company.WebAppConfig) string { return p.OwnershipVariablePosture }},
	{"block_unreachable_domains_traffic_flag", func(p *company.WebAppConfig) string { return p.BlockUnreachableDomainsTrafficFlag }},
	{"prioritize_ipv4_over_ipv6", func(p *company.WebAppConfig) string { return p.PrioritizeIPv4OverIpv6 }},
	{"crowd_strike_zta_score_visibility", func(p *company.WebAppConfig) string { return p.CrowdStrikeZTAScoreVisibility }},
	{"notification_for_zpa_reauth_visibility", func(p *company.WebAppConfig) string { return p.NotificationForZPAReauthVisibility }},
	{"crl_check_visibility_flag", func(p *company.WebAppConfig) string { return p.CrlCheckVisibilityFlag }},
	{"dedicated_proxy_ports_visibility", func(p *company.WebAppConfig) string { return p.DedicatedProxyPortsVisibility }},
	{"remote_fetch_logs", func(p *company.WebAppConfig) string { return p.RemoteFetchLogs }},
	{"ms_defender_posture_visibility", func(p *company.WebAppConfig) string { return p.MsDefenderPostureVisibility }},
	{"exit_password_visibility", func(p *company.WebAppConfig) string { return p.ExitPasswordVisibility }},
	{"collect_zdx_location_visibility", func(p *company.WebAppConfig) string { return p.CollectZdxLocationVisibility }},
	{"use_v8_js_engine_visibility", func(p *company.WebAppConfig) string { return p.UseV8JsEngineVisibility }},
	{"zdx_disable_password_visibility", func(p *company.WebAppConfig) string { return p.ZdxDisablePasswordVisibility }},
	{"zad_disable_password_visibility", func(p *company.WebAppConfig) string { return p.ZadDisablePasswordVisibility }},
	{"zpa_disable_password_visibility", func(p *company.WebAppConfig) string { return p.ZpaDisablePasswordVisibility }},
	{"default_protocol_for_zpa", func(p *company.WebAppConfig) string { return p.DefaultProtocolForZPA }},
	{"drop_ipv6_traffic_visibility", func(p *company.WebAppConfig) string { return p.DropIpv6TrafficVisibility }},
	{"mac_cache_system_proxy_visibility", func(p *company.WebAppConfig) string { return p.MacCacheSystemProxyVisibility }},
	{"use_wsa_poll_for_zpa", func(p *company.WebAppConfig) string { return p.UseWsaPollForZpa }},
	{"enable_64bit_feature", func(p *company.WebAppConfig) string { return p.Enable64BitFeature }},
	{"antivirus_posture_visibility", func(p *company.WebAppConfig) string { return p.AntivirusPostureVisibility }},
	{"system_proxy_on_any_network_change_visibility", func(p *company.WebAppConfig) string { return p.SystemProxyOnAnyNetworkChangeVisibility }},
	{"device_posture_os_version_visibility", func(p *company.WebAppConfig) string { return p.DevicePostureOsVersionVisibility }},
	{"sccm_config_visibility", func(p *company.WebAppConfig) string { return p.SccmConfigVisibility }},
	{"browser_auth_flag_visibility", func(p *company.WebAppConfig) string { return p.BrowserAuthFlagVisibility }},
	{"install_webview2_flag_visibility", func(p *company.WebAppConfig) string { return p.InstallWebView2FlagVisibility }},
	{"allow_webview2_to_follow_sp_visibility", func(p *company.WebAppConfig) string { return p.AllowWebView2ToFollowSPVisibility }},
	{"enable_ipv6_resolution_for_zscaler_domains_visibility", func(p *company.WebAppConfig) string { return p.EnableIpv6ResolutionForZscalerDomainsVisibility }},
	{"disable_reason_visibility", func(p *company.WebAppConfig) string { return p.DisableReasonVisibility }},
	{"follow_routing_table_visibility", func(p *company.WebAppConfig) string { return p.FollowRoutingTableVisibility }},
	{"zia_device_posture_visibility", func(p *company.WebAppConfig) string { return p.ZiaDevicePostureVisibility }},
	{"use_custom_dns", func(p *company.WebAppConfig) string { return p.UseCustomDNS }},
	{"use_default_adapter_for_dns_visibility", func(p *company.WebAppConfig) string { return p.UseDefaultAdapterForDNSVisibility }},
	{"t2_fallback_block_all_traffic_and_tls_fallback", func(p *company.WebAppConfig) string { return p.T2FallbackBlockAllTrafficAndTlsFallback }},
	{"override_t2_protocol_setting", func(p *company.WebAppConfig) string { return p.OverrideT2ProtocolSetting }},
	{"grant_access_to_zscaler_log_folder_visibility", func(p *company.WebAppConfig) string { return p.GrantAccessToZscalerLogFolderVisibility }},
	{"admin_management_visibility", func(p *company.WebAppConfig) string { return p.AdminManagementVisibility }},
	{"redirect_web_traffic_to_zcc_listening_proxy_visibility", func(p *company.WebAppConfig) string { return p.RedirectWebTrafficToZccListeningProxyVisibility }},
	{"use_ztunnel_2_0_for_proxied_web_traffic_visibility", func(p *company.WebAppConfig) string { return p.UseZtunnel2_0ForProxiedWebTrafficVisibility }},
	{"split_vpn_visibility", func(p *company.WebAppConfig) string { return p.SplitVpnVisibility }},
	{"evaluate_trusted_network_visibility", func(p *company.WebAppConfig) string { return p.EvaluateTrustedNetworkVisibility }},
	{"vpn_adapters_configuration_visibility", func(p *company.WebAppConfig) string { return p.VpnAdaptersConfigurationVisibility }},
	{"vpn_services_visibility", func(p *company.WebAppConfig) string { return p.VpnServicesVisibility }},
	{"skip_trusted_criteria_match_visibility", func(p *company.WebAppConfig) string { return p.SkipTrustedCriteriaMatchVisibility }},
	{"external_device_id_visibility", func(p *company.WebAppConfig) string { return p.ExternalDeviceIdVisibility }},
	{"flow_logger_loopback_type_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerLoopbackTypeVisibility }},
	{"flow_logger_zpa_type_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerZPATypeVisibility }},
	{"flow_logger_vpn_type_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerVPNTypeVisibility }},
	{"flow_logger_vpn_tunnel_type_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerVPNTunnelTypeVisibility }},
	{"flow_logger_direct_type_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerDirectTypeVisibility }},
	{"use_zscaler_notification_framework", func(p *company.WebAppConfig) string { return p.UseZscalerNotificationFramework }},
	{"fallback_to_gateway_domain", func(p *company.WebAppConfig) string { return p.FallbackToGatewayDomain }},
	{"zcc_revert_visibility", func(p *company.WebAppConfig) string { return p.ZccRevertVisibility }},
	{"force_zcc_revert_visibility", func(p *company.WebAppConfig) string { return p.ForceZccRevertVisibility }},
	{"disaster_recovery_visibility", func(p *company.WebAppConfig) string { return p.DisasterRecoveryVisibility }},
	{"device_group_visibility", func(p *company.WebAppConfig) string { return p.DeviceGroupVisibility }},
	{"ipv6_support_for_tunnel2", func(p *company.WebAppConfig) string { return p.IpV6SupportForTunnel2 }},
	{"path_mtu_discovery", func(p *company.WebAppConfig) string { return p.PathMtuDiscovery }},
	{"posture_disc_encryption_visibility_for_linux", func(p *company.WebAppConfig) string { return p.PostureDiscEncryptionVisibilityForLinux }},
	{"posture_ms_defender_visibility_for_linux", func(p *company.WebAppConfig) string { return p.PostureMsDefenderVisibilityForLinux }},
	{"posture_os_version_visibility_for_linux", func(p *company.WebAppConfig) string { return p.PostureOsVersionVisibilityForLinux }},
	{"posture_crowd_strike_zta_score_visibility_for_linux", func(p *company.WebAppConfig) string { return p.PostureCrowdStrikeZTAScoreVisibilityForLinux }},
	{"flow_logger_zcc_blocked_traffic_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerZCCBlockedTrafficVisibility }},
	{"flow_logger_intranet_traffic_visibility", func(p *company.WebAppConfig) string { return p.FlowLoggerIntranetTrafficVisibility }},
	{"custom_mtu_for_zpa_visibility", func(p *company.WebAppConfig) string { return p.CustomMTUForZpaVisibility }},
	{"zpa_auto_reauth_timeout_visibility", func(p *company.WebAppConfig) string { return p.ZpaAutoReauthTimeoutVisibility }},
	{"force_zpa_auth_expire_visibility", func(p *company.WebAppConfig) string { return p.ForceZpaAuthExpireVisibility }},
	{"enable_set_proxy_on_vpn_adapters_visibility", func(p *company.WebAppConfig) string { return p.EnableSetProxyOnVPNAdaptersVisibility }},
	{"dns_server_route_exclusion_visibility", func(p *company.WebAppConfig) string { return p.DnsServerRouteExclusionVisibility }},
	{"enable_separate_otp_for_device", func(p *company.WebAppConfig) string { return p.EnableSeparateOtpForDevice }},
	{"uninstall_password_for_profile_visibility", func(p *company.WebAppConfig) string { return p.UninstallPasswordForProfileVisibility }},
	{"zpa_advance_reauth_visibility", func(p *company.WebAppConfig) string { return p.ZpaAdvanceReauthVisibility }},
	{"latency_based_zen_enablement_visibility", func(p *company.WebAppConfig) string { return p.LatencyBasedZenEnablementVisibility }},
	{"dynamic_zpa_service_edge_assignmentt_visibility", func(p *company.WebAppConfig) string { return p.DynamicZPAServiceEdgeAssignmenttVisibility }},
	{"custom_proxy_ports_visibility", func(p *company.WebAppConfig) string { return p.CustomProxyPortsVisibility }},
	{"domain_inclusion_exclusion_for_dns_request_visibility", func(p *company.WebAppConfig) string { return p.DomainInclusionExclusionForDNSRequestVisibility }},
	{"app_notification_config_visibility", func(p *company.WebAppConfig) string { return p.AppNotificationConfigVisibility }},
	{"enable_anti_tampering_visibility", func(p *company.WebAppConfig) string { return p.EnableAntiTamperingVisibility }},
	{"strict_enforcement_status_visibility", func(p *company.WebAppConfig) string { return p.StrictEnforcementStatusVisibility }},
	{"anti_tampering_otp_support_visibility", func(p *company.WebAppConfig) string { return p.AntiTamperingOtpSupportVisibility }},
	{"override_at_cmd_by_policy_visibility", func(p *company.WebAppConfig) string { return p.OverrideATCmdByPolicyVisibility }},
	{"device_trust_level_visibility", func(p *company.WebAppConfig) string { return p.DeviceTrustLevelVisibility }},
	{"source_port_based_bypasses_visibility", func(p *company.WebAppConfig) string { return p.SourcePortBasedBypassesVisibility }},
	{"process_based_application_bypass_visibility", func(p *company.WebAppConfig) string { return p.ProcessBasedApplicationBypassVisibility }},
	{"custom_based_application_bypass_visibility", func(p *company.WebAppConfig) string { return p.CustomBasedApplicationBypassVisibility }},
	{"client_certificate_template_visibility", func(p *company.WebAppConfig) string { return p.ClientCertificateTemplateVisibility }},
	{"supported_zcc_version_chart_visibility", func(p *company.WebAppConfig) string { return p.SupportedZccVersionChartVisibility }},
	{"ios_ipv6_mode_visibility", func(p *company.WebAppConfig) string { return p.IosIpv6ModeVisibility }},
	{"device_group_multiple_postures_visibility", func(p *company.WebAppConfig) string { return p.DeviceGroupMultiplePosturesVisibility }},
	{"drop_non_zscaler_packets_visibility", func(p *company.WebAppConfig) string { return p.DropNonZscalerPacketsVisibility }},
	{"zcc_synthetic_ip_range_visibility", func(p *company.WebAppConfig) string { return p.ZccSyntheticIPRangeVisibility }},
	{"device_posture_frequency_visibility", func(p *company.WebAppConfig) string { return p.DevicePostureFrequencyVisibility }},
	{"enforce_split_dns_visibility", func(p *company.WebAppConfig) string { return p.EnforceSplitDNSVisibility }},
	{"data_protection_visibility", func(p *company.WebAppConfig) string { return p.DataProtectionVisibility }},
	{"drop_quic_traffic_visibility", func(p *company.WebAppConfig) string { return p.DropQuicTrafficVisibility }},
	{"truncate_large_udp_dns_response_visibility", func(p *company.WebAppConfig) string { return p.TruncateLargeUDPDNSResponseVisibility }},
	{"prioritize_dns_exclusions_visibility", func(p *company.WebAppConfig) string { return p.PrioritizeDnsExclusionsVisibility }},
	{"fetch_log_configuration_option_visibility", func(p *company.WebAppConfig) string { return p.FetchLogConfigurationOptionVisibility }},
	{"enable_serial_number_visibility", func(p *company.WebAppConfig) string { return p.EnableSerialNumberVisibility }},
	{"support_multiple_pwl_postures", func(p *company.WebAppConfig) string { return p.SupportMultiplePWLPostures }},
	{"restrict_remote_packet_capture_visibility", func(p *company.WebAppConfig) string { return p.RestrictRemotePacketCaptureVisibility }},
	{"enable_application_based_bypass_for_mac_visibility", func(p *company.WebAppConfig) string { return p.EnableApplicationBasedBypassForMacVisibility }},
	{"remove_exempted_containers_visibility", func(p *company.WebAppConfig) string { return p.RemoveExemptedContainersVisibility }},
	{"captive_portal_detection_visibility", func(p *company.WebAppConfig) string { return p.CaptivePortalDetectionVisibility }},
	{"device_group_in_profile_visibility", func(p *company.WebAppConfig) string { return p.DeviceGroupInProfileVisibility }},
	{"update_dns_search_order", func(p *company.WebAppConfig) string { return p.UpdateDnsSearchOrder }},
	{"install_activity_based_monitoring_driver_visibility", func(p *company.WebAppConfig) string { return p.InstallActivityBasedMonitoringDriverVisibility }},
	{"slow_rollout_zcc", func(p *company.WebAppConfig) string { return p.SlowRolloutZCC }},
	{"zcc_tunnel_version_visibility", func(p *company.WebAppConfig) string { return p.ZccTunnelVersionVisibility }},
	{"anti_tampering_status_visibility", func(p *company.WebAppConfig) string { return p.AntiTamperingStatusVisibility }},
	{"lbb_threshold_rank_to_percent_mapping", func(p *company.WebAppConfig) string { return p.LbbThresholdRankToPercentMapping }},
	{"remove_zscaler_ssl_cert_url", func(p *company.WebAppConfig) string { return p.RemoveZscalerSslCertUrl }},
	{"lbz_threshold_rank_to_percent_mapping", func(p *company.WebAppConfig) string { return p.LbzThresholdRankToPercentMapping }},
	{"splash_screen_url", func(p *company.WebAppConfig) string { return p.SplashScreenUrl }},
	{"splash_screen_visibility", func(p *company.WebAppConfig) string { return p.SplashScreenVisibility }},
	{"trusted_network_range_criteria_visibility", func(p *company.WebAppConfig) string { return p.TrustedNetworkRangeCriteriaVisibility }},
	{"trusted_egress_ips_visibility", func(p *company.WebAppConfig) string { return p.TrustedEgressIpsVisibility }},
	{"domain_profile_detection_visibility", func(p *company.WebAppConfig) string { return p.DomainProfileDetectionVisibility }},
	{"all_inbound_traffic_visibility", func(p *company.WebAppConfig) string { return p.AllInboundTrafficVisibility }},
	{"export_logs_for_non_admin_visibility", func(p *company.WebAppConfig) string { return p.ExportLogsForNonAdminVisibility }},
	{"enable_auto_log_snippet_visibility", func(p *company.WebAppConfig) string { return p.EnableAutoLogSnippetVisibility }},
	{"enable_cli_visibility", func(p *company.WebAppConfig) string { return p.EnableCliVisibility }},
	{"zcc_user_type_visibility", func(p *company.WebAppConfig) string { return p.ZccUserTypeVisibility }},
	{"install_windows_firewall_inbound_rule", func(p *company.WebAppConfig) string { return p.InstallWindowsFirewallInboundRule }},
	{"retry_after_in_seconds", func(p *company.WebAppConfig) string { return p.RetryAfterInSeconds }},
	{"azure_ad_posture_visibility", func(p *company.WebAppConfig) string { return p.AzureADPostureVisibility }},
	{"server_cert_posture_visibility", func(p *company.WebAppConfig) string { return p.ServerCertPostureVisibility }},
	{"perform_crl_check_server_posture_visibility", func(p *company.WebAppConfig) string { return p.PerformCRLCheckServerPostureVisibility }},
	{"auto_fill_using_login_hint_visibility", func(p *company.WebAppConfig) string { return p.AutoFillUsingLoginHintVisibility }},
	{"send_default_policy_for_invalid_policy_token", func(p *company.WebAppConfig) string { return p.SendDefaultPolicyForInvalidPolicyToken }},
	{"enable_zcc_password_settings", func(p *company.WebAppConfig) string { return p.EnableZccPasswordSettings }},
	{"cli_password_expiry_minutes", func(p *company.WebAppConfig) string { return p.CliPasswordExpiryMinutes }},
	{"sso_using_windows_primary_account", func(p *company.WebAppConfig) string { return p.SsoUsingWindowsPrimaryAccount }},
	{"enable_verbose_log", func(p *company.WebAppConfig) string { return p.EnableVerboseLog }},
	{"zpa_auth_exp_on_win_logon_session", func(p *company.WebAppConfig) string { return p.ZpaAuthExpOnWinLogonSession }},
	{"zpa_auth_exp_on_win_session_lock_visibility", func(p *company.WebAppConfig) string { return p.ZpaAuthExpOnWinSessionLockVisibility }},
	{"enable_zcc_slow_rollout_by_default", func(p *company.WebAppConfig) string { return p.EnableZccSlowRolloutByDefault }},
	{"purge_kerberos_preferred_dc_cache_visibility", func(p *company.WebAppConfig) string { return p.PurgeKerberosPreferredDCCacheVisibility }},
	{"posture_jamf_detection_visibility", func(p *company.WebAppConfig) string { return p.PostureJamfDetectionVisibility }},
	{"posture_jamf_device_risk_visibility", func(p *company.WebAppConfig) string { return p.PostureJamfDeviceRiskVisibility }},
	{"windows_ap_captive_portal_detection_visibility", func(p *company.WebAppConfig) string { return p.WindowsAPCaptivePortalDetectionVisibility }},
	{"windows_ap_enable_fail_open_visibility", func(p *company.WebAppConfig) string { return p.WindowsAPEnableFailOpenVisibility }},
	{"automatic_capture_duration", func(p *company.WebAppConfig) string { return p.AutomaticCaptureDuration }},
	{"force_location_refresh_sccm", func(p *company.WebAppConfig) string { return p.ForceLocationRefreshSccm }},
	{"enable_posture_failure_dashboard", func(p *company.WebAppConfig) string { return p.EnablePostureFailureDashboard }},
	{"enable_one_id_phase_2_changes", func(p *company.WebAppConfig) string { return p.EnableOneIDPhase2Changes }},
	{"drop_ipv6_traffic_in_ipv6_network_visibility", func(p *company.WebAppConfig) string { return p.DropIpv6TrafficInIpv6NetworkVisibility }},
	{"enable_postures_for_partner", func(p *company.WebAppConfig) string { return p.EnablePosturesForPartner }},
	{"enable_partner_config_in_primary_policy", func(p *company.WebAppConfig) string { return p.EnablePartnerConfigInPrimaryPolicy }},
	{"enable_one_id_admin_migration_changes", func(p *company.WebAppConfig) string { return p.EnableOneIDAdminMigrationChanges }},
	{"ddil_config_visibility", func(p *company.WebAppConfig) string { return p.DdilConfigVisibility }},
	{"add_zdx_service_entitlement", func(p *company.WebAppConfig) string { return p.AddZDXServiceEntitlement }},
	{"use_zcdn", func(p *company.WebAppConfig) string { return p.UseZcdn }},
	{"delete_dhcp_option_121_routes_visibility", func(p *company.WebAppConfig) string { return p.DeleteDHCPOption121RoutesVisibility }},
	{"zdx_rollout_control_visibility", func(p *company.WebAppConfig) string { return p.ZdxRolloutControlVisibility }},
	{"show_m365_services_in_app_bypasses", func(p *company.WebAppConfig) string { return p.ShowM365ServicesInAppBypasses }},
	{"allow_webview2_ignore_client_cert_errors", func(p *company.WebAppConfig) string { return p.AllowWebView2IgnoreClientCertErrors }},
	{"linux_rpm_build_visibility", func(p *company.WebAppConfig) string { return p.LinuxRPMBuildVisibility }},
	{"help_banner_data_visibility", func(p *company.WebAppConfig) string { return p.HelpBannerDataVisibility }},
	{"zpa_only_device_cleanup_visibility", func(p *company.WebAppConfig) string { return p.ZpaOnlyDeviceCleanupVisibility }},
	{"app_profile_fail_open_policy_visibility", func(p *company.WebAppConfig) string { return p.AppProfileFailOpenPolicyVisibility }},
	{"show_registry_option_in_enforce_and_none", func(p *company.WebAppConfig) string { return p.ShowRegistryOptionInEnforceAndNone }},
	{"strict_enforcement_notification_visibility", func(p *company.WebAppConfig) string { return p.StrictEnforcementNotificationVisibility }},
	{"crowd_strike_zta_os_score_visibility", func(p *company.WebAppConfig) string { return p.CrowdStrikeZTAOsScoreVisibility }},
	{"crowd_strike_zta_sensor_config_score_visibility", func(p *company.WebAppConfig) string { return p.CrowdStrikeZTASensorConfigScoreVisibility }},
	{"resize_window_to_fit_to_page_visibility", func(p *company.WebAppConfig) string { return p.ResizeWindowToFitToPageVisibility }},
	{"enable_zcc_fail_close_settings_for_se_mode", func(p *company.WebAppConfig) string { return p.EnableZCCFailCloseSettingsForSEMode }},
}

// webAppConfigAttrTypes is the attr.Type map matching the schema attributes
// produced by webAppConfigSchemaAttributes. Both are derived from
// webAppConfigStringFields so they cannot drift apart.
func webAppConfigAttrTypes() map[string]attr.Type {
	out := make(map[string]attr.Type, len(webAppConfigStringFields))
	for _, f := range webAppConfigStringFields {
		out[f.name] = types.StringType
	}
	return out
}

// webAppConfigSchemaAttributes returns the schema attribute map for the
// `web_app_config` SingleNestedAttribute. Every entry is a Computed string
// because the API returns these values as opaque flags ("0"/"1" or feature
// codes) and the data source is read-only.
func webAppConfigSchemaAttributes() map[string]schema.Attribute {
	out := make(map[string]schema.Attribute, len(webAppConfigStringFields))
	for _, f := range webAppConfigStringFields {
		out[f.name] = schema.StringAttribute{Computed: true}
	}
	return out
}

// webAppConfigToObject flattens *company.WebAppConfig into a types.Object
// whose AttrTypes match webAppConfigAttrTypes. When the API omits the
// `webAppConfig` field entirely the SDK pointer is nil, in which case we
// return a typed null Object so the framework still has the right schema
// shape.
func webAppConfigToObject(_ context.Context, w *company.WebAppConfig) (types.Object, diag.Diagnostics) {
	objType := webAppConfigAttrTypes()
	if w == nil {
		return types.ObjectNull(objType), nil
	}
	values := make(map[string]attr.Value, len(webAppConfigStringFields))
	for _, f := range webAppConfigStringFields {
		values[f.name] = types.StringValue(f.get(w))
	}
	return types.ObjectValue(objType, values)
}
