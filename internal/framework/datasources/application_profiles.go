package datasources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/application_profiles"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &ApplicationProfilesDataSource{}
	_ datasource.DataSourceWithConfigure = &ApplicationProfilesDataSource{}
)

func NewApplicationProfilesDataSource() datasource.DataSource {
	return &ApplicationProfilesDataSource{}
}

type ApplicationProfilesDataSource struct {
	client *client.Client
}

type ApplicationProfilesDataSourceModel struct {
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	DeviceType                   types.String `tfsdk:"device_type"`
	Description                  types.String `tfsdk:"description"`
	PacURL                       types.String `tfsdk:"pac_url"`
	Active                       types.Int64  `tfsdk:"active"`
	RuleOrder                    types.Int64  `tfsdk:"rule_order"`
	LogMode                      types.Int64  `tfsdk:"log_mode"`
	LogLevel                     types.Int64  `tfsdk:"log_level"`
	LogFileSize                  types.Int64  `tfsdk:"log_file_size"`
	ReauthPeriod                 types.String `tfsdk:"reauth_period"`
	ReactivateWebSecurityMinutes types.String `tfsdk:"reactivate_web_security_minutes"`
	HighlightActiveControl       types.Int64  `tfsdk:"highlight_active_control"`
	SendDisableServiceReason     types.Int64  `tfsdk:"send_disable_service_reason"`
	RefreshKerberosToken         types.Int64  `tfsdk:"refresh_kerberos_token"`
	EnableDeviceGroups           types.Int64  `tfsdk:"enable_device_groups"`
	Groups                       types.List   `tfsdk:"groups"`
	DeviceGroups                 types.List   `tfsdk:"device_groups"`
	NotificationTemplateId       types.Int64  `tfsdk:"notification_template_id"`
	ForwardingProfileId          types.Int64  `tfsdk:"forwarding_profile_id"`
	ZiaPostureConfigId           types.Int64  `tfsdk:"zia_posture_config_id"`
	PolicyToken                  types.String `tfsdk:"policy_token"`
	TunnelZappTraffic            types.Int64  `tfsdk:"tunnel_zapp_traffic"`
	GroupAll                     types.Int64  `tfsdk:"group_all"`
	Users                        types.List   `tfsdk:"users"`
	GroupIds                     types.List   `tfsdk:"group_ids"`
	DeviceGroupIds               types.List   `tfsdk:"device_group_ids"`
	UserIds                      types.List   `tfsdk:"user_ids"`
	BypassAppIds                 types.List   `tfsdk:"bypass_app_ids"`
	AppServiceIds                types.List   `tfsdk:"app_service_ids"`
	BypassCustomAppIds           types.List   `tfsdk:"bypass_custom_app_ids"`
	Passcode                     types.String `tfsdk:"passcode"`
	LogoutPassword               types.String `tfsdk:"logout_password"`
	DisablePassword              types.String `tfsdk:"disable_password"`
	UninstallPassword            types.String `tfsdk:"uninstall_password"`
	ShowVPNTunNotification       types.Int64  `tfsdk:"show_vpn_tun_notification"`
	Ipv6Mode                     types.Int64  `tfsdk:"ipv6_mode"`
	DisasterRecovery             types.List   `tfsdk:"disaster_recovery"`
	PolicyExtension              types.List   `tfsdk:"policy_extension"`
}

func (d *ApplicationProfilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_profiles"
}

func (d *ApplicationProfilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC application profile by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application profile.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the application profile.",
				Optional:    true,
				Computed:    true,
			},
			"device_type":                     schema.StringAttribute{Computed: true},
			"description":                     schema.StringAttribute{Computed: true},
			"pac_url":                         schema.StringAttribute{Computed: true},
			"active":                          schema.Int64Attribute{Computed: true},
			"rule_order":                      schema.Int64Attribute{Computed: true},
			"log_mode":                        schema.Int64Attribute{Computed: true},
			"log_level":                       schema.Int64Attribute{Computed: true},
			"log_file_size":                   schema.Int64Attribute{Computed: true},
			"reauth_period":                   schema.StringAttribute{Computed: true},
			"reactivate_web_security_minutes": schema.StringAttribute{Computed: true},
			"highlight_active_control":        schema.Int64Attribute{Computed: true},
			"send_disable_service_reason":     schema.Int64Attribute{Computed: true},
			"refresh_kerberos_token":          schema.Int64Attribute{Computed: true},
			"enable_device_groups":            schema.Int64Attribute{Computed: true},
			"groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.Int64Attribute{Computed: true},
						"name":              schema.StringAttribute{Computed: true},
						"auth_type":         schema.StringAttribute{Computed: true},
						"active":            schema.Int64Attribute{Computed: true},
						"last_modification": schema.StringAttribute{Computed: true},
					},
				},
			},
			"device_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.Int64Attribute{Computed: true},
						"name":              schema.StringAttribute{Computed: true},
						"auth_type":         schema.StringAttribute{Computed: true},
						"active":            schema.Int64Attribute{Computed: true},
						"last_modification": schema.StringAttribute{Computed: true},
					},
				},
			},
			"notification_template_id": schema.Int64Attribute{Computed: true},
			"forwarding_profile_id":    schema.Int64Attribute{Computed: true},
			"zia_posture_config_id":    schema.Int64Attribute{Computed: true},
			"policy_token":             schema.StringAttribute{Computed: true},
			"tunnel_zapp_traffic":      schema.Int64Attribute{Computed: true},
			"group_all":                schema.Int64Attribute{Computed: true},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true},
						"login_name":        schema.StringAttribute{Computed: true},
						"last_modification": schema.StringAttribute{Computed: true},
						"active":            schema.Int64Attribute{Computed: true},
						"company_id":        schema.StringAttribute{Computed: true},
					},
				},
			},
			"group_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"device_group_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"user_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"bypass_app_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"app_service_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"bypass_custom_app_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"passcode":                  schema.StringAttribute{Computed: true},
			"logout_password":           schema.StringAttribute{Computed: true},
			"disable_password":          schema.StringAttribute{Computed: true},
			"uninstall_password":        schema.StringAttribute{Computed: true},
			"show_vpn_tun_notification": schema.Int64Attribute{Computed: true},
			"ipv6_mode":                 schema.Int64Attribute{Computed: true},
			"disaster_recovery": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"enable_zia_dr":        schema.BoolAttribute{Computed: true},
						"enable_zpa_dr":        schema.BoolAttribute{Computed: true},
						"zia_dr_method":        schema.Int64Attribute{Computed: true},
						"zia_custom_db_url":    schema.StringAttribute{Computed: true},
						"use_zia_global_db":    schema.BoolAttribute{Computed: true},
						"zia_global_db_url":    schema.StringAttribute{Computed: true},
						"zia_global_db_url_v2": schema.StringAttribute{Computed: true},
						"zia_domain_name":      schema.StringAttribute{Computed: true},
						"zia_rsa_pub_key_name": schema.StringAttribute{Computed: true},
						"zia_rsa_pub_key":      schema.StringAttribute{Computed: true},
						"zpa_domain_name":      schema.StringAttribute{Computed: true},
						"zpa_rsa_pub_key_name": schema.StringAttribute{Computed: true},
						"zpa_rsa_pub_key":      schema.StringAttribute{Computed: true},
						"allow_zia_test":       schema.BoolAttribute{Computed: true},
						"allow_zpa_test":       schema.BoolAttribute{Computed: true},
					},
				},
			},
			"policy_extension": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source_port_based_bypasses":                      schema.StringAttribute{Computed: true},
						"packet_tunnel_exclude_list":                      schema.StringAttribute{Computed: true},
						"packet_tunnel_include_list":                      schema.StringAttribute{Computed: true},
						"custom_dns":                                      schema.StringAttribute{Computed: true},
						"exit_password":                                   schema.StringAttribute{Computed: true},
						"use_v8_js_engine":                                schema.StringAttribute{Computed: true},
						"zdx_disable_password":                            schema.StringAttribute{Computed: true},
						"zd_disable_password":                             schema.StringAttribute{Computed: true},
						"zpa_disable_password":                            schema.StringAttribute{Computed: true},
						"zdp_disable_password":                            schema.StringAttribute{Computed: true},
						"follow_routing_table":                            schema.StringAttribute{Computed: true},
						"use_wsa_poll_for_zpa":                            schema.StringAttribute{Computed: true},
						"use_default_adapter_for_dns":                     schema.StringAttribute{Computed: true},
						"use_zscaler_notification_framework":              schema.StringAttribute{Computed: true},
						"switch_focus_to_notification":                    schema.StringAttribute{Computed: true},
						"fallback_to_gateway_domain":                      schema.StringAttribute{Computed: true},
						"enable_zcc_revert":                               schema.StringAttribute{Computed: true},
						"zcc_revert_password":                             schema.StringAttribute{Computed: true},
						"zpa_auth_exp_on_sleep":                           schema.Int64Attribute{Computed: true},
						"zpa_auth_exp_on_sys_restart":                     schema.Int64Attribute{Computed: true},
						"zpa_auth_exp_on_net_ip_change":                   schema.Int64Attribute{Computed: true},
						"instant_force_zpa_reauth_state_update":           schema.Int64Attribute{Computed: true},
						"zpa_auth_exp_on_win_logon_session":               schema.Int64Attribute{Computed: true},
						"zpa_auth_exp_on_win_session_lock":                schema.Int64Attribute{Computed: true},
						"zpa_auth_exp_session_lock_min_time":              schema.Int64Attribute{Computed: true},
						"packet_tunnel_exclude_list_for_ipv6":             schema.StringAttribute{Computed: true},
						"packet_tunnel_include_list_for_ipv6":             schema.StringAttribute{Computed: true},
						"enable_set_proxy_on_vpn_adapters":                schema.Int64Attribute{Computed: true},
						"disable_dns_route_exclusion":                     schema.Int64Attribute{Computed: true},
						"advance_zpa_reauth":                              schema.BoolAttribute{Computed: true},
						"use_proxy_port_for_t1":                           schema.StringAttribute{Computed: true},
						"use_proxy_port_for_t2":                           schema.StringAttribute{Computed: true},
						"allow_pac_exclusions_only":                       schema.StringAttribute{Computed: true},
						"intercept_zia_traffic_all_adapters":              schema.StringAttribute{Computed: true},
						"enable_anti_tampering":                           schema.StringAttribute{Computed: true},
						"override_at_cmd_by_policy":                       schema.StringAttribute{Computed: true},
						"reactivate_anti_tampering_time":                  schema.Int64Attribute{Computed: true},
						"enforce_split_dns":                               schema.Int64Attribute{Computed: true},
						"drop_quic_traffic":                               schema.Int64Attribute{Computed: true},
						"enable_zdp_service":                              schema.StringAttribute{Computed: true},
						"update_dns_search_order":                         schema.Int64Attribute{Computed: true},
						"truncate_large_udp_dns_response":                 schema.Int64Attribute{Computed: true},
						"prioritize_dns_exclusions":                       schema.Int64Attribute{Computed: true},
						"purge_kerberos_preferred_dc_cache":               schema.StringAttribute{Computed: true},
						"delete_dhcp_option_121_routes":                   schema.StringAttribute{Computed: true},
						"enable_location_policy_override":                 schema.Int64Attribute{Computed: true},
						"enable_custom_theme":                             schema.Int64Attribute{Computed: true},
						"zdx_lite_config_obj":                             schema.StringAttribute{Computed: true},
						"ddil_config":                                     schema.StringAttribute{Computed: true},
						"zcc_fail_close_settings_ip_bypasses":             schema.StringAttribute{Computed: true},
						"zcc_fail_close_settings_exit_uninstall_password": schema.StringAttribute{Computed: true},
						"zcc_fail_close_lockdown_tunnel_proc_exit":        schema.Int64Attribute{Computed: true},
						"zcc_fail_close_lockdown_firewall_error":          schema.Int64Attribute{Computed: true},
						"zcc_fail_close_lockdown_driver_error":            schema.Int64Attribute{Computed: true},
						"zcc_fail_close_settings_thumb_print":             schema.StringAttribute{Computed: true},
						"zcc_app_fail_open_policy":                        schema.Int64Attribute{Computed: true},
						"zcc_tunnel_fail_policy":                          schema.Int64Attribute{Computed: true},
						"follow_global_for_partner_login":                 schema.StringAttribute{Computed: true},
						"user_allowed_to_add_partner":                     schema.StringAttribute{Computed: true},
						"allow_client_cert_caching_for_webview2":          schema.StringAttribute{Computed: true},
						"show_confirmation_dialog_for_cached_cert":        schema.StringAttribute{Computed: true},
						"enable_flow_based_tunnel":                        schema.Int64Attribute{Computed: true},
						"enable_network_traffic_process_mapping":          schema.Int64Attribute{Computed: true},
						"enable_local_packet_capture":                     schema.StringAttribute{Computed: true},
						"one_id_mt_device_auth_enabled":                   schema.StringAttribute{Computed: true},
						"enable_custom_proxy_detection":                   schema.StringAttribute{Computed: true},
						"prevent_auto_reauth_during_device_lock":          schema.StringAttribute{Computed: true},
						"use_endpoint_location_for_dc_selection":          schema.StringAttribute{Computed: true},
						"enable_crash_reporting":                          schema.Int64Attribute{Computed: true},
						"recache_system_proxy":                            schema.StringAttribute{Computed: true},
						"enable_automatic_packet_capture":                 schema.Int64Attribute{Computed: true},
						"enable_apc_for_critical_sections":                schema.Int64Attribute{Computed: true},
						"enable_apc_for_other_sections":                   schema.Int64Attribute{Computed: true},
						"enable_pc_additional_space":                      schema.Int64Attribute{Computed: true},
						"pc_additional_space":                             schema.Int64Attribute{Computed: true},
						"client_connector_ui_language":                    schema.Int64Attribute{Computed: true},
						"block_private_relay":                             schema.StringAttribute{Computed: true},
						"bypass_dns_traffic_using_udp_proxy":              schema.Int64Attribute{Computed: true},
						"reconnect_tun_on_wakeup":                         schema.Int64Attribute{Computed: true},
						"browser_auth_type":                               schema.StringAttribute{Computed: true},
						"use_default_browser":                             schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ApplicationProfilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *ApplicationProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.Name.IsNull() || data.Name.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	var profile *application_profiles.ApplicationProfile

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching application profile by ID", map[string]any{"id": id})
		result, _, err := application_profiles.GetByProfileID(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read application profile: %v", err))
			return
		}
		profile = result
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching application profile by name", map[string]any{"name": name})
		result, _, err := application_profiles.GetByName(ctx, service, name)
		if err != nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Application profile with name '%s' not found: %v", name, err))
			return
		}
		profile = result
	}

	model := flattenApplicationProfile(profile)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func ptrStringValue(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func ptrInt64Value(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

var policyGroupObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":                types.Int64Type,
		"name":              types.StringType,
		"auth_type":         types.StringType,
		"active":            types.Int64Type,
		"last_modification": types.StringType,
	},
}

func flattenPolicyGroups(groups []application_profiles.ApplicationPolicyGroup) types.List {
	if len(groups) == 0 {
		return types.ListNull(policyGroupObjType)
	}
	elements := make([]attr.Value, 0, len(groups))
	for _, g := range groups {
		obj, _ := types.ObjectValue(policyGroupObjType.AttrTypes, map[string]attr.Value{
			"id":                types.Int64Value(g.ID),
			"name":              types.StringValue(g.Name),
			"auth_type":         types.StringValue(g.AuthType),
			"active":            types.Int64Value(int64(g.Active)),
			"last_modification": types.StringValue(g.LastModification),
		})
		elements = append(elements, obj)
	}
	list, _ := types.ListValue(policyGroupObjType, elements)
	return list
}

var policyUserObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":                types.StringType,
		"login_name":        types.StringType,
		"last_modification": types.StringType,
		"active":            types.Int64Type,
		"company_id":        types.StringType,
	},
}

func flattenPolicyUsers(users []application_profiles.ApplicationPolicyUser) types.List {
	if len(users) == 0 {
		return types.ListNull(policyUserObjType)
	}
	elements := make([]attr.Value, 0, len(users))
	for _, u := range users {
		obj, _ := types.ObjectValue(policyUserObjType.AttrTypes, map[string]attr.Value{
			"id":                types.StringValue(u.ID),
			"login_name":        types.StringValue(u.LoginName),
			"last_modification": types.StringValue(u.LastModification),
			"active":            types.Int64Value(int64(u.Active)),
			"company_id":        types.StringValue(u.CompanyID),
		})
		elements = append(elements, obj)
	}
	list, _ := types.ListValue(policyUserObjType, elements)
	return list
}

func flattenApplicationProfile(p *application_profiles.ApplicationProfile) ApplicationProfilesDataSourceModel {
	return ApplicationProfilesDataSourceModel{
		ID:                           types.StringValue(strconv.Itoa(p.ID)),
		Name:                         types.StringValue(p.Name),
		DeviceType:                   types.StringValue(p.DeviceType),
		Description:                  types.StringValue(p.Description),
		PacURL:                       types.StringValue(p.PacURL),
		Active:                       types.Int64Value(int64(p.Active)),
		RuleOrder:                    types.Int64Value(int64(p.RuleOrder)),
		LogMode:                      types.Int64Value(int64(p.LogMode)),
		LogLevel:                     types.Int64Value(int64(p.LogLevel)),
		LogFileSize:                  types.Int64Value(int64(p.LogFileSize)),
		ReauthPeriod:                 ptrStringValue(p.ReauthPeriod),
		ReactivateWebSecurityMinutes: types.StringValue(p.ReactivateWebSecurityMinutes),
		HighlightActiveControl:       types.Int64Value(int64(p.HighlightActiveControl)),
		SendDisableServiceReason:     types.Int64Value(int64(p.SendDisableServiceReason)),
		RefreshKerberosToken:         types.Int64Value(int64(p.RefreshKerberosToken)),
		EnableDeviceGroups:           types.Int64Value(int64(p.EnableDeviceGroups)),
		Groups:                       flattenPolicyGroups(p.Groups),
		DeviceGroups:                 flattenPolicyGroups(p.DeviceGroups),
		NotificationTemplateId:       types.Int64Value(int64(p.NotificationTemplateId)),
		ForwardingProfileId:          ptrInt64Value(p.ForwardingProfileId),
		ZiaPostureConfigId:           ptrInt64Value(p.ZiaPostureConfigId),
		PolicyToken:                  types.StringValue(p.PolicyToken),
		TunnelZappTraffic:            types.Int64Value(int64(p.TunnelZappTraffic)),
		GroupAll:                     types.Int64Value(int64(p.GroupAll)),
		Users:                        flattenPolicyUsers(p.Users),
		GroupIds:                     flattenStringSlice(p.GroupIds),
		DeviceGroupIds:               flattenStringSlice(p.DeviceGroupIds),
		UserIds:                      flattenStringSlice(p.UserIds),
		BypassAppIds:                 flattenStringSlice(p.BypassAppIds),
		AppServiceIds:                flattenStringSlice(p.AppServiceIds),
		BypassCustomAppIds:           flattenStringSlice(p.BypassCustomAppIds),
		Passcode:                     types.StringValue(p.Passcode),
		LogoutPassword:               types.StringValue(p.LogoutPassword),
		DisablePassword:              types.StringValue(p.DisablePassword),
		UninstallPassword:            ptrStringValue(p.UninstallPassword),
		ShowVPNTunNotification:       types.Int64Value(int64(p.ShowVPNTunNotification)),
		Ipv6Mode:                     types.Int64Value(int64(p.Ipv6Mode)),
		DisasterRecovery:             flattenDisasterRecovery(&p.DisasterRecovery),
		PolicyExtension:              flattenPolicyExtension(&p.PolicyExtension),
	}
}

var disasterRecoveryObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"enable_zia_dr":        types.BoolType,
		"enable_zpa_dr":        types.BoolType,
		"zia_dr_method":        types.Int64Type,
		"zia_custom_db_url":    types.StringType,
		"use_zia_global_db":    types.BoolType,
		"zia_global_db_url":    types.StringType,
		"zia_global_db_url_v2": types.StringType,
		"zia_domain_name":      types.StringType,
		"zia_rsa_pub_key_name": types.StringType,
		"zia_rsa_pub_key":      types.StringType,
		"zpa_domain_name":      types.StringType,
		"zpa_rsa_pub_key_name": types.StringType,
		"zpa_rsa_pub_key":      types.StringType,
		"allow_zia_test":       types.BoolType,
		"allow_zpa_test":       types.BoolType,
	},
}

func flattenDisasterRecovery(dr *application_profiles.DisasterRecovery) types.List {
	obj, _ := types.ObjectValue(disasterRecoveryObjType.AttrTypes, map[string]attr.Value{
		"enable_zia_dr":        types.BoolValue(dr.EnableZiaDR),
		"enable_zpa_dr":        types.BoolValue(dr.EnableZpaDR),
		"zia_dr_method":        types.Int64Value(int64(dr.ZiaDRMethod)),
		"zia_custom_db_url":    types.StringValue(dr.ZiaCustomDbUrl),
		"use_zia_global_db":    types.BoolValue(dr.UseZiaGlobalDb),
		"zia_global_db_url":    types.StringValue(dr.ZiaGlobalDbUrl),
		"zia_global_db_url_v2": types.StringValue(dr.ZiaGlobalDbUrlv2),
		"zia_domain_name":      types.StringValue(dr.ZiaDomainName),
		"zia_rsa_pub_key_name": types.StringValue(dr.ZiaRSAPubKeyName),
		"zia_rsa_pub_key":      types.StringValue(dr.ZiaRSAPubKey),
		"zpa_domain_name":      types.StringValue(dr.ZpaDomainName),
		"zpa_rsa_pub_key_name": types.StringValue(dr.ZpaRSAPubKeyName),
		"zpa_rsa_pub_key":      types.StringValue(dr.ZpaRSAPubKey),
		"allow_zia_test":       types.BoolValue(dr.AllowZiaTest),
		"allow_zpa_test":       types.BoolValue(dr.AllowZpaTest),
	})
	list, _ := types.ListValue(disasterRecoveryObjType, []attr.Value{obj})
	return list
}

var policyExtensionObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"source_port_based_bypasses":                      types.StringType,
		"packet_tunnel_exclude_list":                      types.StringType,
		"packet_tunnel_include_list":                      types.StringType,
		"custom_dns":                                      types.StringType,
		"exit_password":                                   types.StringType,
		"use_v8_js_engine":                                types.StringType,
		"zdx_disable_password":                            types.StringType,
		"zd_disable_password":                             types.StringType,
		"zpa_disable_password":                            types.StringType,
		"zdp_disable_password":                            types.StringType,
		"follow_routing_table":                            types.StringType,
		"use_wsa_poll_for_zpa":                            types.StringType,
		"use_default_adapter_for_dns":                     types.StringType,
		"use_zscaler_notification_framework":              types.StringType,
		"switch_focus_to_notification":                    types.StringType,
		"fallback_to_gateway_domain":                      types.StringType,
		"enable_zcc_revert":                               types.StringType,
		"zcc_revert_password":                             types.StringType,
		"zpa_auth_exp_on_sleep":                           types.Int64Type,
		"zpa_auth_exp_on_sys_restart":                     types.Int64Type,
		"zpa_auth_exp_on_net_ip_change":                   types.Int64Type,
		"instant_force_zpa_reauth_state_update":           types.Int64Type,
		"zpa_auth_exp_on_win_logon_session":               types.Int64Type,
		"zpa_auth_exp_on_win_session_lock":                types.Int64Type,
		"zpa_auth_exp_session_lock_min_time":              types.Int64Type,
		"packet_tunnel_exclude_list_for_ipv6":             types.StringType,
		"packet_tunnel_include_list_for_ipv6":             types.StringType,
		"enable_set_proxy_on_vpn_adapters":                types.Int64Type,
		"disable_dns_route_exclusion":                     types.Int64Type,
		"advance_zpa_reauth":                              types.BoolType,
		"use_proxy_port_for_t1":                           types.StringType,
		"use_proxy_port_for_t2":                           types.StringType,
		"allow_pac_exclusions_only":                       types.StringType,
		"intercept_zia_traffic_all_adapters":              types.StringType,
		"enable_anti_tampering":                           types.StringType,
		"override_at_cmd_by_policy":                       types.StringType,
		"reactivate_anti_tampering_time":                  types.Int64Type,
		"enforce_split_dns":                               types.Int64Type,
		"drop_quic_traffic":                               types.Int64Type,
		"enable_zdp_service":                              types.StringType,
		"update_dns_search_order":                         types.Int64Type,
		"truncate_large_udp_dns_response":                 types.Int64Type,
		"prioritize_dns_exclusions":                       types.Int64Type,
		"purge_kerberos_preferred_dc_cache":               types.StringType,
		"delete_dhcp_option_121_routes":                   types.StringType,
		"enable_location_policy_override":                 types.Int64Type,
		"enable_custom_theme":                             types.Int64Type,
		"zdx_lite_config_obj":                             types.StringType,
		"ddil_config":                                     types.StringType,
		"zcc_fail_close_settings_ip_bypasses":             types.StringType,
		"zcc_fail_close_settings_exit_uninstall_password": types.StringType,
		"zcc_fail_close_lockdown_tunnel_proc_exit":        types.Int64Type,
		"zcc_fail_close_lockdown_firewall_error":          types.Int64Type,
		"zcc_fail_close_lockdown_driver_error":            types.Int64Type,
		"zcc_fail_close_settings_thumb_print":             types.StringType,
		"zcc_app_fail_open_policy":                        types.Int64Type,
		"zcc_tunnel_fail_policy":                          types.Int64Type,
		"follow_global_for_partner_login":                 types.StringType,
		"user_allowed_to_add_partner":                     types.StringType,
		"allow_client_cert_caching_for_webview2":          types.StringType,
		"show_confirmation_dialog_for_cached_cert":        types.StringType,
		"enable_flow_based_tunnel":                        types.Int64Type,
		"enable_network_traffic_process_mapping":          types.Int64Type,
		"enable_local_packet_capture":                     types.StringType,
		"one_id_mt_device_auth_enabled":                   types.StringType,
		"enable_custom_proxy_detection":                   types.StringType,
		"prevent_auto_reauth_during_device_lock":          types.StringType,
		"use_endpoint_location_for_dc_selection":          types.StringType,
		"enable_crash_reporting":                          types.Int64Type,
		"recache_system_proxy":                            types.StringType,
		"enable_automatic_packet_capture":                 types.Int64Type,
		"enable_apc_for_critical_sections":                types.Int64Type,
		"enable_apc_for_other_sections":                   types.Int64Type,
		"enable_pc_additional_space":                      types.Int64Type,
		"pc_additional_space":                             types.Int64Type,
		"client_connector_ui_language":                    types.Int64Type,
		"block_private_relay":                             types.StringType,
		"bypass_dns_traffic_using_udp_proxy":              types.Int64Type,
		"reconnect_tun_on_wakeup":                         types.Int64Type,
		"browser_auth_type":                               types.StringType,
		"use_default_browser":                             types.StringType,
	},
}

func flattenPolicyExtension(pe *application_profiles.PolicyExtension) types.List {
	obj, _ := types.ObjectValue(policyExtensionObjType.AttrTypes, map[string]attr.Value{
		"source_port_based_bypasses":                      types.StringValue(pe.SourcePortBasedBypasses),
		"packet_tunnel_exclude_list":                      types.StringValue(pe.PacketTunnelExcludeList),
		"packet_tunnel_include_list":                      types.StringValue(pe.PacketTunnelIncludeList),
		"custom_dns":                                      types.StringValue(pe.CustomDNS),
		"exit_password":                                   types.StringValue(pe.ExitPassword),
		"use_v8_js_engine":                                types.StringValue(pe.UseV8JsEngine),
		"zdx_disable_password":                            types.StringValue(pe.ZdxDisablePassword),
		"zd_disable_password":                             types.StringValue(pe.ZdDisablePassword),
		"zpa_disable_password":                            types.StringValue(pe.ZpaDisablePassword),
		"zdp_disable_password":                            types.StringValue(pe.ZdpDisablePassword),
		"follow_routing_table":                            types.StringValue(pe.FollowRoutingTable),
		"use_wsa_poll_for_zpa":                            types.StringValue(pe.UseWsaPollForZpa),
		"use_default_adapter_for_dns":                     types.StringValue(pe.UseDefaultAdapterForDNS),
		"use_zscaler_notification_framework":              types.StringValue(pe.UseZscalerNotificationFramework),
		"switch_focus_to_notification":                    types.StringValue(pe.SwitchFocusToNotification),
		"fallback_to_gateway_domain":                      types.StringValue(pe.FallbackToGatewayDomain),
		"enable_zcc_revert":                               types.StringValue(pe.EnableZCCRevert),
		"zcc_revert_password":                             types.StringValue(pe.ZccRevertPassword),
		"zpa_auth_exp_on_sleep":                           types.Int64Value(int64(pe.ZpaAuthExpOnSleep)),
		"zpa_auth_exp_on_sys_restart":                     types.Int64Value(int64(pe.ZpaAuthExpOnSysRestart)),
		"zpa_auth_exp_on_net_ip_change":                   types.Int64Value(int64(pe.ZpaAuthExpOnNetIpChange)),
		"instant_force_zpa_reauth_state_update":           types.Int64Value(int64(pe.InstantForceZPAReauthStateUpdate)),
		"zpa_auth_exp_on_win_logon_session":               types.Int64Value(int64(pe.ZpaAuthExpOnWinLogonSession)),
		"zpa_auth_exp_on_win_session_lock":                types.Int64Value(int64(pe.ZpaAuthExpOnWinSessionLock)),
		"zpa_auth_exp_session_lock_min_time":              types.Int64Value(int64(pe.ZpaAuthExpSessionLockStateMinTimeInSecond)),
		"packet_tunnel_exclude_list_for_ipv6":             types.StringValue(pe.PacketTunnelExcludeListForIPv6),
		"packet_tunnel_include_list_for_ipv6":             types.StringValue(pe.PacketTunnelIncludeListForIPv6),
		"enable_set_proxy_on_vpn_adapters":                types.Int64Value(int64(pe.EnableSetProxyOnVPNAdapters)),
		"disable_dns_route_exclusion":                     types.Int64Value(int64(pe.DisableDNSRouteExclusion)),
		"advance_zpa_reauth":                              types.BoolValue(pe.AdvanceZpaReauth),
		"use_proxy_port_for_t1":                           types.StringValue(pe.UseProxyPortForT1),
		"use_proxy_port_for_t2":                           types.StringValue(pe.UseProxyPortForT2),
		"allow_pac_exclusions_only":                       types.StringValue(pe.AllowPacExclusionsOnly),
		"intercept_zia_traffic_all_adapters":              types.StringValue(pe.InterceptZIATrafficAllAdapters),
		"enable_anti_tampering":                           types.StringValue(pe.EnableAntiTampering),
		"override_at_cmd_by_policy":                       types.StringValue(pe.OverrideATCmdByPolicy),
		"reactivate_anti_tampering_time":                  types.Int64Value(int64(pe.ReactivateAntiTamperingTime)),
		"enforce_split_dns":                               types.Int64Value(int64(pe.EnforceSplitDNS)),
		"drop_quic_traffic":                               types.Int64Value(int64(pe.DropQuicTraffic)),
		"enable_zdp_service":                              types.StringValue(pe.EnableZdpService),
		"update_dns_search_order":                         types.Int64Value(int64(pe.UpdateDnsSearchOrder)),
		"truncate_large_udp_dns_response":                 types.Int64Value(int64(pe.TruncateLargeUDPDNSResponse)),
		"prioritize_dns_exclusions":                       types.Int64Value(int64(pe.PrioritizeDnsExclusions)),
		"purge_kerberos_preferred_dc_cache":               types.StringValue(pe.PurgeKerberosPreferredDCCache),
		"delete_dhcp_option_121_routes":                   types.StringValue(pe.DeleteDHCPOption121Routes),
		"enable_location_policy_override":                 types.Int64Value(int64(pe.EnableLocationPolicyOverride)),
		"enable_custom_theme":                             types.Int64Value(int64(pe.EnableCustomTheme)),
		"zdx_lite_config_obj":                             types.StringValue(pe.ZdxLiteConfigObj),
		"ddil_config":                                     types.StringValue(pe.DdilConfig),
		"zcc_fail_close_settings_ip_bypasses":             types.StringValue(pe.ZccFailCloseSettingsIpBypasses),
		"zcc_fail_close_settings_exit_uninstall_password": types.StringValue(pe.ZccFailCloseSettingsExitUninstallPassword),
		"zcc_fail_close_lockdown_tunnel_proc_exit":        types.Int64Value(int64(pe.ZccFailCloseSettingsLockdownOnTunnelProcExit)),
		"zcc_fail_close_lockdown_firewall_error":          types.Int64Value(int64(pe.ZccFailCloseSettingsLockdownOnFirewallError)),
		"zcc_fail_close_lockdown_driver_error":            types.Int64Value(int64(pe.ZccFailCloseSettingsLockdownOnDriverError)),
		"zcc_fail_close_settings_thumb_print":             types.StringValue(pe.ZccFailCloseSettingsThumbPrint),
		"zcc_app_fail_open_policy":                        types.Int64Value(int64(pe.ZccAppFailOpenPolicy)),
		"zcc_tunnel_fail_policy":                          types.Int64Value(int64(pe.ZccTunnelFailPolicy)),
		"follow_global_for_partner_login":                 types.StringValue(pe.FollowGlobalForPartnerLogin),
		"user_allowed_to_add_partner":                     types.StringValue(pe.UserAllowedToAddPartner),
		"allow_client_cert_caching_for_webview2":          types.StringValue(pe.AllowClientCertCachingForWebView2),
		"show_confirmation_dialog_for_cached_cert":        types.StringValue(pe.ShowConfirmationDialogForCachedCert),
		"enable_flow_based_tunnel":                        types.Int64Value(int64(pe.EnableFlowBasedTunnel)),
		"enable_network_traffic_process_mapping":          types.Int64Value(int64(pe.EnableNetworkTrafficProcessMapping)),
		"enable_local_packet_capture":                     types.StringValue(pe.EnableLocalPacketCapture),
		"one_id_mt_device_auth_enabled":                   types.StringValue(pe.OneIdMTDeviceAuthEnabled),
		"enable_custom_proxy_detection":                   types.StringValue(pe.EnableCustomProxyDetection),
		"prevent_auto_reauth_during_device_lock":          types.StringValue(pe.PreventAutoReauthDuringDeviceLock),
		"use_endpoint_location_for_dc_selection":          types.StringValue(pe.UseEndPointLocationForDCSelection),
		"enable_crash_reporting":                          types.Int64Value(int64(pe.EnableCrashReporting)),
		"recache_system_proxy":                            types.StringValue(pe.RecacheSystemProxy),
		"enable_automatic_packet_capture":                 types.Int64Value(int64(pe.EnableAutomaticPacketCapture)),
		"enable_apc_for_critical_sections":                types.Int64Value(int64(pe.EnableAPCforCriticalSections)),
		"enable_apc_for_other_sections":                   types.Int64Value(int64(pe.EnableAPCforOtherSections)),
		"enable_pc_additional_space":                      types.Int64Value(int64(pe.EnablePCAdditionalSpace)),
		"pc_additional_space":                             types.Int64Value(int64(pe.PcAdditionalSpace)),
		"client_connector_ui_language":                    types.Int64Value(int64(pe.ClientConnectorUiLanguage)),
		"block_private_relay":                             types.StringValue(pe.BlockPrivateRelay),
		"bypass_dns_traffic_using_udp_proxy":              types.Int64Value(int64(pe.BypassDNSTrafficUsingUDPProxy)),
		"reconnect_tun_on_wakeup":                         types.Int64Value(int64(pe.ReconnectTunOnWakeup)),
		"browser_auth_type":                               types.StringValue(pe.BrowserAuthType),
		"use_default_browser":                             types.StringValue(pe.UseDefaultBrowser),
	})
	list, _ := types.ListValue(policyExtensionObjType, []attr.Value{obj})
	return list
}
