// Package common contains shared infrastructure for resources backed by the
// Zscaler ZCC `web_policy` SDK service. Every per-OS app_profile resource
// (zcc_app_profile_windows / _macos / _linux / _ios / _android) embeds
// WebPolicyBaseModel, hardcodes its deviceType integer, declares its own
// OS-specific block (windows_policy, mac_policy, etc.), and delegates the
// full lifecycle to RunUpsert / RunRead / RunDelete.
//
// Lifecycle (singleton-style PUT-and-re-read):
//
//	Create  → PUT /web/policy/edit (id="") → API returns {success,id}
//	          → re-read via GetWebPolicyByID(id, deviceType) for the
//	          full server-authoritative state
//	Read    → GetWebPolicyByID(id, deviceType); 404 ⇒ remove from state
//	Update  → PUT /web/policy/edit (id=state.id) → re-read same way
//	Delete  → DELETE /web/policy/{id}/delete
//
// When the optional `activate` attribute is true (default), Create / Update
// also call ActivateWebPolicy so the change goes live.
//
// The package is intentionally domain-scoped: helpers that are not specific
// to web_policy belong elsewhere. Keeping the scope narrow prevents this
// package from drifting into a junk drawer of unrelated framework helpers.
package common

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zccCommon "github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/common"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/web_policy"

	"github.com/zscaler/terraform-provider-zcc/internal/framework/helpers"
)

// =============================================================================
// Shared base model
// =============================================================================

// WebPolicyBaseModel is embedded by every per-OS app_profile resource model.
// The Plugin Framework walks anonymous embedded structs as if their fields
// were declared directly on the parent, so each resource automatically gets
// these tfsdk fields and adds only its OS-specific block on top.
//
// The integer-valued attributes (rule_order, log_mode, log_level,
// log_file_size, reactivate_web_security_minutes, reauth_period,
// send_disable_service_reason, tunnel_zapp_traffic, group_all,
// highlight_active_control, enable_device_groups) are typed as Int64
// because the ZCC /web/policy/edit endpoint requires them on the wire as
// JSON numbers. Sending them as quoted strings makes the API respond
// HTTP 200 + {"success":"false","id":0} (silent failure). The Go SDK
// uses common.IntOrString for these fields, which marshals as a number
// and unmarshals from either a number or a quoted string (the GET
// response is inconsistent and quotes some of them).
type WebPolicyBaseModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Active                    types.String `tfsdk:"active"`
	Description               types.String `tfsdk:"description"`
	DeviceType                types.String `tfsdk:"device_type"`
	Activate                  types.Bool   `tfsdk:"activate"`
	AllowUnreachablePac       types.Bool   `tfsdk:"allow_unreachable_pac"`
	HighlightActiveControl    types.Int64  `tfsdk:"highlight_active_control"`
	LogFileSize               types.Int64  `tfsdk:"log_file_size"`
	LogLevel                  types.Int64  `tfsdk:"log_level"`
	LogMode                   types.Int64  `tfsdk:"log_mode"`
	PacURL                    types.String `tfsdk:"pac_url"`
	ReactivateWebSecurityMins types.Int64  `tfsdk:"reactivate_web_security_minutes"`
	ReauthPeriod              types.Int64  `tfsdk:"reauth_period"`
	RuleOrder                 types.Int64  `tfsdk:"rule_order"`
	SendDisableServiceReason  types.Int64  `tfsdk:"send_disable_service_reason"`
	TunnelZappTraffic         types.Int64  `tfsdk:"tunnel_zapp_traffic"`
	GroupAll                  types.Int64  `tfsdk:"group_all"`
	EnableDeviceGroups        types.Int64  `tfsdk:"enable_device_groups"`
	ForwardingProfileId       types.Int64  `tfsdk:"forwarding_profile_id"`
	ZiaPostureConfigId        types.Int64  `tfsdk:"zia_posture_config_id"`
	GroupIds                  types.List   `tfsdk:"group_ids"`
	GroupNames                types.List   `tfsdk:"group_names"`
	UserIds                   types.List   `tfsdk:"user_ids"`
	UserNames                 types.List   `tfsdk:"user_names"`
	AppServiceIds             types.List   `tfsdk:"app_service_ids"`
	AppServiceNames           types.List   `tfsdk:"app_service_names"`
	AppIdentityNames          types.List   `tfsdk:"app_identity_names"`
	BypassAppIds              types.List   `tfsdk:"bypass_app_ids"`
	BypassCustomAppIds        types.List   `tfsdk:"bypass_custom_app_ids"`
	DeviceGroupIds            types.List   `tfsdk:"device_group_ids"`
	DeviceGroupNames          types.List   `tfsdk:"device_group_names"`
	DisasterRecovery          types.Object `tfsdk:"disaster_recovery"`
	PolicyExtension           types.Object `tfsdk:"policy_extension"`
}

// =============================================================================
// Top-level schema
// =============================================================================

// WebPolicyBaseAttributes returns the schema attributes shared by all five
// per-OS app_profile resources. The OS-specific nested attribute (e.g.
// `windows_policy`) is added separately in each resource's Schema method
// after copying this map.
func WebPolicyBaseAttributes() map[string]schema.Attribute {
	stringOC := func(desc string) schema.Attribute {
		return schema.StringAttribute{Description: desc, Optional: true, Computed: true}
	}
	int64OC := func(desc string) schema.Attribute {
		return schema.Int64Attribute{Description: desc, Optional: true, Computed: true}
	}
	listInt64OC := func(desc string) schema.Attribute {
		return schema.ListAttribute{Description: desc, ElementType: types.Int64Type, Optional: true, Computed: true}
	}
	listStringOC := func(desc string) schema.Attribute {
		return schema.ListAttribute{Description: desc, ElementType: types.StringType, Optional: true, Computed: true}
	}

	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Server-assigned web policy identifier.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name":        schema.StringAttribute{Description: "Display name of the policy.", Required: true},
		"description": stringOC("Free-form description of the policy."),
		"device_type": schema.StringAttribute{
			Description: "Friendly device type label (\"iOS\", \"Android\", \"Windows\", \"macOS\", \"Linux\") corresponding to the numeric device_type the ZCC API uses. The value is HARD-CODED per resource (zcc_app_profile_macos = macOS, zcc_app_profile_ios = iOS, etc.) — operators do NOT configure this attribute, and the schema enforces that by exposing it as Computed-only.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"active": schema.StringAttribute{
			Description: "Whether the policy is active. The wire format is \"1\" / \"0\"; the provider also accepts \"true\" / \"false\" / \"yes\" / \"no\" / \"on\" / \"off\" and normalizes them to \"1\" / \"0\" before sending — the ZCC API rejects \"true\" / \"false\" with a 400.",
			Optional:    true,
			Computed:    true,
		},
		"activate": schema.BoolAttribute{
			Description: "When true (default), the resource calls /web/policy/activate after every successful create or update so the change goes live.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
		},
		"allow_unreachable_pac":           schema.BoolAttribute{Description: "Allow PAC fetching when the PAC URL is unreachable.", Optional: true, Computed: true},
		"highlight_active_control":        int64OC("Highlight the active control surface in the UI (0/1). Sent as a JSON number — the API rejects quoted strings here."),
		"log_file_size":                   int64OC("Maximum log file size, in megabytes. Sent as a JSON number."),
		"log_level":                       int64OC("Log verbosity level. Sent as a JSON number."),
		"log_mode":                        int64OC("Log mode flag (-1 to disable, otherwise the numeric mode id). Sent as a JSON number."),
		"pac_url":                         stringOC("PAC URL the agent uses for proxy resolution."),
		"reactivate_web_security_minutes": int64OC("Minutes after which web security is automatically reactivated. Sent as a JSON number."),
		"reauth_period":                   int64OC("Re-authentication period, in hours. Sent as a JSON number."),
		"rule_order":                      int64OC("Numeric rule order (1-based). Sent as a JSON number."),
		"send_disable_service_reason":     int64OC("Whether the agent must collect a reason when the service is disabled (0/1)."),
		"tunnel_zapp_traffic":             int64OC("Tunnel ZApp traffic flag (0/1)."),
		"group_all":                       int64OC("1 to apply the policy to all groups, otherwise scoped via group_ids/group_names (0)."),
		"enable_device_groups":            int64OC("Whether device-group scoping is enabled (0/1)."),
		"forwarding_profile_id":           int64OC("Forwarding profile id linked to this policy."),
		"zia_posture_config_id":           int64OC("ZIA posture configuration id linked to this policy."),
		"group_ids":                       listInt64OC("Group ids the policy targets."),
		"group_names":                     listStringOC("Group names the policy targets."),
		"user_ids":                        listInt64OC("User ids the policy targets."),
		"user_names":                      listStringOC("User names the policy targets."),
		"app_service_ids":                 listInt64OC("App service ids referenced by the policy."),
		"app_service_names":               listStringOC("App service names referenced by the policy."),
		"app_identity_names":              listStringOC("App identity names referenced by the policy."),
		"bypass_app_ids":                  listInt64OC("App ids to bypass."),
		"bypass_custom_app_ids":           listInt64OC("Custom app ids to bypass."),
		"device_group_ids":                listInt64OC("Device group ids the policy targets."),
		"device_group_names":              listStringOC("Device group names the policy targets."),
		"disaster_recovery": schema.SingleNestedAttribute{
			Description: "ZIA/ZPA disaster-recovery configuration carried with the policy.",
			Optional:    true,
			Computed:    true,
			Attributes:  DisasterRecoveryAttributes(),
		},
		"policy_extension": schema.SingleNestedAttribute{
			Description: "Extended/auxiliary policy fields. All values are Optional+Computed so users can set what they care about and let server defaults flow through for the rest.",
			Optional:    true,
			Computed:    true,
			Attributes:  PolicyExtensionAttributes(),
		},
	}
}

// =============================================================================
// disaster_recovery nested object
// =============================================================================

// DisasterRecoveryAttrTypes returns the attr.Type map for the
// disaster_recovery nested object. It mirrors the field set built by
// DisasterRecoveryAttributes so the schema and the runtime object value
// always agree on shape.
//
// Tfsdk names follow the wire shape that the ZCC API actually uses
// (verified against a live /listByCompany response): ziaDRMethod (not
// ziaDRRecoveryMethod), ziaRSAPubKey* / zpaRSAPubKey* for the RSA
// material, and ziaCustomDbUrl for custom DR endpoints.
func DisasterRecoveryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allow_zia_test":       types.BoolType,
		"allow_zpa_test":       types.BoolType,
		"enable_zia_dr":        types.BoolType,
		"enable_zpa_dr":        types.BoolType,
		"policy_id":            types.StringType,
		"use_zia_global_db":    types.BoolType,
		"zia_dr_method":        types.Int64Type,
		"zia_custom_db_url":    types.StringType,
		"zia_domain_name":      types.StringType,
		"zia_global_db_url":    types.StringType,
		"zia_global_db_url_v2": types.StringType,
		"zia_pac_url":          types.StringType,
		"zia_rsa_pub_key":      types.StringType,
		"zia_rsa_pub_key_name": types.StringType,
		"zpa_domain_name":      types.StringType,
		"zpa_rsa_pub_key":      types.StringType,
		"zpa_rsa_pub_key_name": types.StringType,
	}
}

// DisasterRecoveryObjectType is a convenience wrapper that returns the
// types.ObjectType built from DisasterRecoveryAttrTypes. Useful when an
// outer object needs to declare the disaster_recovery field's type.
func DisasterRecoveryObjectType() attr.Type {
	return types.ObjectType{AttrTypes: DisasterRecoveryAttrTypes()}
}

// DisasterRecoveryAttributes returns the schema attributes for the
// disaster_recovery SingleNestedAttribute. Every entry is Optional+Computed
// so users can specify only what they care about and let the API supply
// defaults for the rest.
func DisasterRecoveryAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allow_zia_test":       schema.BoolAttribute{Optional: true, Computed: true},
		"allow_zpa_test":       schema.BoolAttribute{Optional: true, Computed: true},
		"enable_zia_dr":        schema.BoolAttribute{Optional: true, Computed: true},
		"enable_zpa_dr":        schema.BoolAttribute{Optional: true, Computed: true},
		"policy_id":            schema.StringAttribute{Optional: true, Computed: true},
		"use_zia_global_db":    schema.BoolAttribute{Optional: true, Computed: true},
		"zia_dr_method":        schema.Int64Attribute{Optional: true, Computed: true, Description: "ZIA disaster-recovery method id. API field: ziaDRMethod."},
		"zia_custom_db_url":    schema.StringAttribute{Optional: true, Computed: true},
		"zia_domain_name":      schema.StringAttribute{Optional: true, Computed: true},
		"zia_global_db_url":    schema.StringAttribute{Optional: true, Computed: true},
		"zia_global_db_url_v2": schema.StringAttribute{Optional: true, Computed: true},
		"zia_pac_url":          schema.StringAttribute{Optional: true, Computed: true},
		"zia_rsa_pub_key":      schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Description: "ZIA RSA public key (PEM). API field: ziaRSAPubKey."},
		"zia_rsa_pub_key_name": schema.StringAttribute{Optional: true, Computed: true, Description: "Friendly name for the ZIA RSA public key. API field: ziaRSAPubKeyName."},
		"zpa_domain_name":      schema.StringAttribute{Optional: true, Computed: true},
		"zpa_rsa_pub_key":      schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Description: "ZPA RSA public key (PEM). API field: zpaRSAPubKey."},
		"zpa_rsa_pub_key_name": schema.StringAttribute{Optional: true, Computed: true, Description: "Friendly name for the ZPA RSA public key. API field: zpaRSAPubKeyName."},
	}
}

// ExpandDisasterRecovery overlays user-set fields from the Terraform
// disaster_recovery object onto out. A null/unknown input leaves out
// untouched, and a known input only writes attributes the operator
// actually set — null/unknown attributes inside the object preserve the
// caller-seeded default. This is what makes the per-device-type
// DefaultMacosWebPolicy.DisasterRecovery values survive when the
// operator only supplies a partial `disaster_recovery = { ... }` block.
func ExpandDisasterRecovery(obj types.Object, out *web_policy.DisasterRecovery) {
	if obj.IsNull() || obj.IsUnknown() {
		return
	}
	a := obj.Attributes()
	overlayString(&out.PolicyId, a, "policy_id")
	overlayString(&out.ZiaCustomDbUrl, a, "zia_custom_db_url")
	overlayString(&out.ZiaDomainName, a, "zia_domain_name")
	overlayString(&out.ZiaGlobalDbURL, a, "zia_global_db_url")
	overlayString(&out.ZiaGlobalDbURLV2, a, "zia_global_db_url_v2")
	overlayString(&out.ZiaPacURL, a, "zia_pac_url")
	overlayString(&out.ZiaRSAPubKey, a, "zia_rsa_pub_key")
	overlayString(&out.ZiaRSAPubKeyName, a, "zia_rsa_pub_key_name")
	overlayString(&out.ZpaDomainName, a, "zpa_domain_name")
	overlayString(&out.ZpaRSAPubKey, a, "zpa_rsa_pub_key")
	overlayString(&out.ZpaRSAPubKeyName, a, "zpa_rsa_pub_key_name")
	overlayBool(&out.AllowZiaTest, a, "allow_zia_test")
	overlayBool(&out.AllowZpaTest, a, "allow_zpa_test")
	overlayBool(&out.EnableZiaDR, a, "enable_zia_dr")
	overlayBool(&out.EnableZpaDR, a, "enable_zpa_dr")
	overlayBool(&out.UseZiaGlobalDb, a, "use_zia_global_db")
	overlayInt(&out.ZiaDRMethod, a, "zia_dr_method")
}

// overlayString writes the string value at key into dst only if it is
// known and non-null. Used by every Expand* function to preserve
// caller-seeded defaults for attributes the operator did not set.
func overlayString(dst *string, a map[string]attr.Value, key string) {
	v, ok := a[key].(types.String)
	if !ok || v.IsNull() || v.IsUnknown() {
		return
	}
	*dst = v.ValueString()
}

// overlayBool is the bool sibling of overlayString.
func overlayBool(dst *bool, a map[string]attr.Value, key string) {
	v, ok := a[key].(types.Bool)
	if !ok || v.IsNull() || v.IsUnknown() {
		return
	}
	*dst = v.ValueBool()
}

// overlayInt is the int sibling of overlayString.
func overlayInt(dst *int, a map[string]attr.Value, key string) {
	v, ok := a[key].(types.Int64)
	if !ok || v.IsNull() || v.IsUnknown() {
		return
	}
	*dst = int(v.ValueInt64())
}

// overlayIntList is the int-list sibling of overlayString.
func overlayIntList(dst *[]int, a map[string]attr.Value, key string) {
	v, ok := a[key].(types.List)
	if !ok || v.IsNull() || v.IsUnknown() {
		return
	}
	*dst = intListFromList(v)
}

// overlayStringList is the string-list sibling of overlayString.
func overlayStringList(dst *[]string, a map[string]attr.Value, key string) {
	v, ok := a[key].(types.List)
	if !ok || v.IsNull() || v.IsUnknown() {
		return
	}
	*dst = stringListFromList(v)
}

// FlattenDisasterRecovery converts the SDK DisasterRecovery struct returned
// by the API into a Plugin Framework types.Object whose AttrTypes match
// DisasterRecoveryAttrTypes.
func FlattenDisasterRecovery(dr web_policy.DisasterRecovery) types.Object {
	obj, _ := types.ObjectValue(DisasterRecoveryAttrTypes(), map[string]attr.Value{
		"allow_zia_test":       types.BoolValue(dr.AllowZiaTest),
		"allow_zpa_test":       types.BoolValue(dr.AllowZpaTest),
		"enable_zia_dr":        types.BoolValue(dr.EnableZiaDR),
		"enable_zpa_dr":        types.BoolValue(dr.EnableZpaDR),
		"policy_id":            types.StringValue(dr.PolicyId),
		"use_zia_global_db":    types.BoolValue(dr.UseZiaGlobalDb),
		"zia_dr_method":        types.Int64Value(int64(dr.ZiaDRMethod)),
		"zia_custom_db_url":    types.StringValue(dr.ZiaCustomDbUrl),
		"zia_domain_name":      types.StringValue(dr.ZiaDomainName),
		"zia_global_db_url":    types.StringValue(dr.ZiaGlobalDbURL),
		"zia_global_db_url_v2": types.StringValue(dr.ZiaGlobalDbURLV2),
		"zia_pac_url":          types.StringValue(dr.ZiaPacURL),
		"zia_rsa_pub_key":      types.StringValue(dr.ZiaRSAPubKey),
		"zia_rsa_pub_key_name": types.StringValue(dr.ZiaRSAPubKeyName),
		"zpa_domain_name":      types.StringValue(dr.ZpaDomainName),
		"zpa_rsa_pub_key":      types.StringValue(dr.ZpaRSAPubKey),
		"zpa_rsa_pub_key_name": types.StringValue(dr.ZpaRSAPubKeyName),
	})
	return obj
}

// =============================================================================
// policy_extension nested object (and its nested generate_cli_password_contract)
// =============================================================================

// policyExtensionStringFields drives both the schema and the attr.Type map
// for the string-typed entries in policy_extension. Each entry maps a
// Terraform attribute name to a getter and setter that move the value
// between the SDK struct and the Plugin Framework value. Centralising the
// list keeps the schema, the attr.Type map and the runtime expand/flatten
// logic in lock-step — a new SDK field only needs a single new entry here.
// policyExtensionStringFields lists the policy_extension attributes that
// are genuine strings on the wire — passwords, comma-separated CIDR
// lists, JSON blobs, configuration paths, etc. The getter/setter
// closure bridges the Terraform StringAttribute with the SDK field's
// Go type (always `string` for these entries).
//
// Boolean-shaped fields ("0"/"1" toggles) live in
// policyExtensionBoolFields below so they can be surfaced to HCL as
// types.Bool rather than types.String — much friendlier for operators
// who otherwise had to remember "is it 0 or 1 for on?". The expand /
// flatten loops in this file iterate BOTH registries.
var policyExtensionStringFields = []struct {
	name string
	get  func(p *web_policy.PolicyExtension) string
	set  func(p *web_policy.PolicyExtension, v string)
}{
	{"custom_dns", func(p *web_policy.PolicyExtension) string { return p.CustomDNS }, func(p *web_policy.PolicyExtension, v string) { p.CustomDNS = v }},
	{"ddil_config", func(p *web_policy.PolicyExtension) string { return p.DdilConfig }, func(p *web_policy.PolicyExtension, v string) { p.DdilConfig = v }},
	// delete_dhcp_option121_routes is a JSON-encoded blob keyed by
	// trust posture (trusted / offTrusted / vpnTrusted / splitVpnTrusted),
	// NOT a "0"/"1" toggle — keep it as a StringAttribute.
	{"delete_dhcp_option121_routes", func(p *web_policy.PolicyExtension) string { return p.DeleteDHCPOption121Routes }, func(p *web_policy.PolicyExtension, v string) { p.DeleteDHCPOption121Routes = v }},
	{"exit_password", func(p *web_policy.PolicyExtension) string { return p.ExitPassword }, func(p *web_policy.PolicyExtension, v string) { p.ExitPassword = v }},
	{"nonce", func(p *web_policy.PolicyExtension) string { return p.Nonce }, func(p *web_policy.PolicyExtension, v string) { p.Nonce = v }},
	{"packet_tunnel_dns_exclude_list", func(p *web_policy.PolicyExtension) string { return p.PacketTunnelDnsExcludeList }, func(p *web_policy.PolicyExtension, v string) { p.PacketTunnelDnsExcludeList = v }},
	{"packet_tunnel_dns_include_list", func(p *web_policy.PolicyExtension) string { return p.PacketTunnelDnsIncludeList }, func(p *web_policy.PolicyExtension, v string) { p.PacketTunnelDnsIncludeList = v }},
	// pcAdditionalSpace is a quoted-number string on the wire (e.g. "512"),
	// not a true integer — the UI capture surfaces it as "512" / "1024".
	// Keep it as a StringAttribute so the operator can pass the exact
	// label-value pair the API expects.
	{"pc_additional_space", func(p *web_policy.PolicyExtension) string { return p.PcAdditionalSpace }, func(p *web_policy.PolicyExtension, v string) { p.PcAdditionalSpace = v }},
	{"packet_tunnel_exclude_list", func(p *web_policy.PolicyExtension) string { return p.PacketTunnelExcludeList }, func(p *web_policy.PolicyExtension, v string) { p.PacketTunnelExcludeList = v }},
	{"packet_tunnel_exclude_list_for_ipv6", func(p *web_policy.PolicyExtension) string { return p.PacketTunnelExcludeListForIPv6 }, func(p *web_policy.PolicyExtension, v string) { p.PacketTunnelExcludeListForIPv6 = v }},
	{"packet_tunnel_include_list", func(p *web_policy.PolicyExtension) string { return p.PacketTunnelIncludeList }, func(p *web_policy.PolicyExtension, v string) { p.PacketTunnelIncludeList = v }},
	{"packet_tunnel_include_list_for_ipv6", func(p *web_policy.PolicyExtension) string { return p.PacketTunnelIncludeListForIPv6 }, func(p *web_policy.PolicyExtension, v string) { p.PacketTunnelIncludeListForIPv6 = v }},
	{"partner_domains", func(p *web_policy.PolicyExtension) string { return p.PartnerDomains }, func(p *web_policy.PolicyExtension, v string) { p.PartnerDomains = v }},
	{"source_port_based_bypasses", func(p *web_policy.PolicyExtension) string { return p.SourcePortBasedBypasses }, func(p *web_policy.PolicyExtension, v string) { p.SourcePortBasedBypasses = v }},
	{"vpn_gateways", func(p *web_policy.PolicyExtension) string { return p.VpnGateways }, func(p *web_policy.PolicyExtension, v string) { p.VpnGateways = v }},
	{"zcc_fail_close_settings_exit_uninstall_password", func(p *web_policy.PolicyExtension) string { return p.ZccFailCloseSettingsExitUninstallPassword }, func(p *web_policy.PolicyExtension, v string) { p.ZccFailCloseSettingsExitUninstallPassword = v }},
	{"zcc_fail_close_settings_ip_bypasses", func(p *web_policy.PolicyExtension) string { return p.ZccFailCloseSettingsIpBypasses }, func(p *web_policy.PolicyExtension, v string) { p.ZccFailCloseSettingsIpBypasses = v }},
	{"zcc_fail_close_settings_thumb_print", func(p *web_policy.PolicyExtension) string { return p.ZccFailCloseSettingsThumbPrint }, func(p *web_policy.PolicyExtension, v string) { p.ZccFailCloseSettingsThumbPrint = v }},
	{"zcc_revert_password", func(p *web_policy.PolicyExtension) string { return p.ZccRevertPassword }, func(p *web_policy.PolicyExtension, v string) { p.ZccRevertPassword = v }},
	{"zd_disable_password", func(p *web_policy.PolicyExtension) string { return p.ZdDisablePassword }, func(p *web_policy.PolicyExtension, v string) { p.ZdDisablePassword = v }},
	{"zdp_disable_password", func(p *web_policy.PolicyExtension) string { return p.ZdpDisablePassword }, func(p *web_policy.PolicyExtension, v string) { p.ZdpDisablePassword = v }},
	{"zdx_disable_password", func(p *web_policy.PolicyExtension) string { return p.ZdxDisablePassword }, func(p *web_policy.PolicyExtension, v string) { p.ZdxDisablePassword = v }},
	{"zdx_lite_config_obj", func(p *web_policy.PolicyExtension) string { return p.ZdxLiteConfigObj }, func(p *web_policy.PolicyExtension, v string) { p.ZdxLiteConfigObj = v }},
	{"zpa_disable_password", func(p *web_policy.PolicyExtension) string { return p.ZpaDisablePassword }, func(p *web_policy.PolicyExtension, v string) { p.ZpaDisablePassword = v }},
}

// policyExtensionBoolFields lists the policy_extension attributes that
// look like genuine on/off toggles on the wire — the ZCC API echoes
// them back as "0" / "1" (or as a JSON number 0/1 for the
// common.IntOrString-typed entries). The getter / setter on each
// entry continues to deal in the SDK's native shape — string or
// IntOrString through intOrStringFieldToString/FromString — and the
// expand/flatten loops on these entries bridge to types.Bool at the
// framework boundary via helpers.BoolToString01 / helpers.String01ToBool.
//
// Moving a field out of policyExtensionStringFields into this list is
// the only change required to switch its HCL surface from "1" / "0"
// string to true / false bool; the SDK struct and the wire shape are
// untouched.
var policyExtensionBoolFields = []struct {
	name string
	get  func(p *web_policy.PolicyExtension) string
	set  func(p *web_policy.PolicyExtension, v string)
}{
	{"disable_dns_route_exclusion", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.DisableDNSRouteExclusion)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.DisableDNSRouteExclusion = intOrStringFieldFromString(v)
	}},
	{"drop_quic_traffic", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.DropQuicTraffic) }, func(p *web_policy.PolicyExtension, v string) { p.DropQuicTraffic = intOrStringFieldFromString(v) }},
	{"enable_anti_tampering", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.EnableAntiTampering) }, func(p *web_policy.PolicyExtension, v string) { p.EnableAntiTampering = intOrStringFieldFromString(v) }},
	{"enable_set_proxy_on_vpn_adapters", func(p *web_policy.PolicyExtension) string { return p.EnableSetProxyOnVPNAdapters }, func(p *web_policy.PolicyExtension, v string) { p.EnableSetProxyOnVPNAdapters = v }},
	{"enable_zcc_revert", func(p *web_policy.PolicyExtension) string { return p.EnableZCCRevert }, func(p *web_policy.PolicyExtension, v string) { p.EnableZCCRevert = v }},
	{"enable_zdp_service", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.EnableZdpService) }, func(p *web_policy.PolicyExtension, v string) { p.EnableZdpService = intOrStringFieldFromString(v) }},
	{"enforce_split_dns", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.EnforceSplitDNS) }, func(p *web_policy.PolicyExtension, v string) { p.EnforceSplitDNS = intOrStringFieldFromString(v) }},
	{"fallback_to_gateway_domain", func(p *web_policy.PolicyExtension) string { return p.FallbackToGatewayDomain }, func(p *web_policy.PolicyExtension, v string) { p.FallbackToGatewayDomain = v }},
	{"follow_global_for_partner_login", func(p *web_policy.PolicyExtension) string { return p.FollowGlobalForPartnerLogin }, func(p *web_policy.PolicyExtension, v string) { p.FollowGlobalForPartnerLogin = v }},
	{"follow_routing_table", func(p *web_policy.PolicyExtension) string { return p.FollowRoutingTable }, func(p *web_policy.PolicyExtension, v string) { p.FollowRoutingTable = v }},
	{"intercept_zia_traffic_all_adapters", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.InterceptZIATrafficAllAdapters)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.InterceptZIATrafficAllAdapters = intOrStringFieldFromString(v)
	}},
	{"override_at_cmd_by_policy", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.OverrideATCmdByPolicy) }, func(p *web_policy.PolicyExtension, v string) {
		p.OverrideATCmdByPolicy = intOrStringFieldFromString(v)
	}},
	{"prioritize_dns_exclusions", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.PrioritizeDnsExclusions) }, func(p *web_policy.PolicyExtension, v string) {
		p.PrioritizeDnsExclusions = intOrStringFieldFromString(v)
	}},
	{"purge_kerberos_preferred_dc_cache", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.PurgeKerberosPreferredDCCache)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.PurgeKerberosPreferredDCCache = intOrStringFieldFromString(v)
	}},
	{"truncate_large_udp_dns_response", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.TruncateLargeUDPDNSResponse)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.TruncateLargeUDPDNSResponse = intOrStringFieldFromString(v)
	}},
	{"update_dns_search_order", func(p *web_policy.PolicyExtension) string { return p.UpdateDnsSearchOrder }, func(p *web_policy.PolicyExtension, v string) { p.UpdateDnsSearchOrder = v }},
	{"use_default_adapter_for_dns", func(p *web_policy.PolicyExtension) string { return p.UseDefaultAdapterForDNS }, func(p *web_policy.PolicyExtension, v string) { p.UseDefaultAdapterForDNS = v }},
	{"use_proxy_port_for_t1", func(p *web_policy.PolicyExtension) string { return p.UseProxyPortForT1 }, func(p *web_policy.PolicyExtension, v string) { p.UseProxyPortForT1 = v }},
	{"use_proxy_port_for_t2", func(p *web_policy.PolicyExtension) string { return p.UseProxyPortForT2 }, func(p *web_policy.PolicyExtension, v string) { p.UseProxyPortForT2 = v }},
	{"use_v8_js_engine", func(p *web_policy.PolicyExtension) string { return p.UseV8JsEngine }, func(p *web_policy.PolicyExtension, v string) { p.UseV8JsEngine = v }},
	{"use_wsa_poll_for_zpa", func(p *web_policy.PolicyExtension) string { return p.UseWsaPollForZpa }, func(p *web_policy.PolicyExtension, v string) { p.UseWsaPollForZpa = v }},
	{"use_zscaler_notification_framework", func(p *web_policy.PolicyExtension) string { return p.UseZscalerNotificationFramework }, func(p *web_policy.PolicyExtension, v string) { p.UseZscalerNotificationFramework = v }},
	{"user_allowed_to_add_partner", func(p *web_policy.PolicyExtension) string { return p.UserAllowedToAddPartner }, func(p *web_policy.PolicyExtension, v string) { p.UserAllowedToAddPartner = v }},
	{"zcc_app_fail_open_policy", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.ZccAppFailOpenPolicy) }, func(p *web_policy.PolicyExtension, v string) {
		p.ZccAppFailOpenPolicy = intOrStringFieldFromString(v)
	}},
	{"zcc_fail_close_settings_lockdown_on_tunnel_process_exit", func(p *web_policy.PolicyExtension) string { return p.ZccFailCloseSettingsLockdownOnTunnelProcessExit }, func(p *web_policy.PolicyExtension, v string) {
		p.ZccFailCloseSettingsLockdownOnTunnelProcessExit = v
	}},
	{"zcc_tunnel_fail_policy", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.ZccTunnelFailPolicy) }, func(p *web_policy.PolicyExtension, v string) {
		p.ZccTunnelFailPolicy = intOrStringFieldFromString(v)
	}},
	{"zpa_auth_exp_on_net_ip_change", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.ZpaAuthExpOnNetIpChange) }, func(p *web_policy.PolicyExtension, v string) {
		p.ZpaAuthExpOnNetIpChange = intOrStringFieldFromString(v)
	}},
	{"zpa_auth_exp_on_sleep", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.ZpaAuthExpOnSleep) }, func(p *web_policy.PolicyExtension, v string) {
		p.ZpaAuthExpOnSleep = intOrStringFieldFromString(v)
	}},
	{"zpa_auth_exp_on_sys_restart", func(p *web_policy.PolicyExtension) string { return intOrStringFieldToString(p.ZpaAuthExpOnSysRestart) }, func(p *web_policy.PolicyExtension, v string) {
		p.ZpaAuthExpOnSysRestart = intOrStringFieldFromString(v)
	}},
	{"zpa_auth_exp_on_win_logon_session", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.ZpaAuthExpOnWinLogonSession)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.ZpaAuthExpOnWinLogonSession = intOrStringFieldFromString(v)
	}},
	{"zpa_auth_exp_on_win_session_lock", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.ZpaAuthExpOnWinSessionLock)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.ZpaAuthExpOnWinSessionLock = intOrStringFieldFromString(v)
	}},

	// --- Additional 0/1 toggles surfaced from the macOS /web/policy/edit payload ---
	// All entries below were already present on the SDK PolicyExtension
	// struct (and the per-OS default constructor seeded sensible values
	// for them); they just weren't exposed to HCL. They follow the same
	// "true → "1" on the wire, false → "0"" convention as the rest of the
	// boolean registry. Fields whose SDK type is common.IntOrString go
	// through intOrStringField{To,From}String so the wire shape is the
	// JSON number the API expects on writes.
	{"follow_global_for_zpa_reauth", func(p *web_policy.PolicyExtension) string { return p.FollowGlobalForZpaReauth }, func(p *web_policy.PolicyExtension, v string) { p.FollowGlobalForZpaReauth = v }},
	{"follow_global_for_packet_capture", func(p *web_policy.PolicyExtension) string { return p.FollowGlobalForPacketCapture }, func(p *web_policy.PolicyExtension, v string) { p.FollowGlobalForPacketCapture = v }},
	{"enable_local_packet_capture", func(p *web_policy.PolicyExtension) string { return p.EnableLocalPacketCapture }, func(p *web_policy.PolicyExtension, v string) { p.EnableLocalPacketCapture = v }},
	{"switch_focus_to_notification", func(p *web_policy.PolicyExtension) string { return p.SwitchFocusToNotification }, func(p *web_policy.PolicyExtension, v string) { p.SwitchFocusToNotification = v }},
	{"allow_pac_exclusions_only", func(p *web_policy.PolicyExtension) string { return p.AllowPacExclusionsOnly }, func(p *web_policy.PolicyExtension, v string) { p.AllowPacExclusionsOnly = v }},
	{"instant_force_zpa_reauth_state_update", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.InstantForceZPAReauthStateUpdate)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.InstantForceZPAReauthStateUpdate = intOrStringFieldFromString(v)
	}},
	{"enable_flow_based_tunnel", func(p *web_policy.PolicyExtension) string { return p.EnableFlowBasedTunnel }, func(p *web_policy.PolicyExtension, v string) { p.EnableFlowBasedTunnel = v }},
	{"zcc_fail_close_settings_lockdown_on_firewall_error", func(p *web_policy.PolicyExtension) string { return p.ZccFailCloseSettingsLockdownOnFirewallError }, func(p *web_policy.PolicyExtension, v string) {
		p.ZccFailCloseSettingsLockdownOnFirewallError = v
	}},
	{"zcc_fail_close_settings_lockdown_on_driver_error", func(p *web_policy.PolicyExtension) string { return p.ZccFailCloseSettingsLockdownOnDriverError }, func(p *web_policy.PolicyExtension, v string) {
		p.ZccFailCloseSettingsLockdownOnDriverError = v
	}},
	{"allow_client_cert_caching_for_web_view2", func(p *web_policy.PolicyExtension) string { return p.AllowClientCertCachingForWebView2 }, func(p *web_policy.PolicyExtension, v string) { p.AllowClientCertCachingForWebView2 = v }},
	{"show_confirmation_dialog_for_cached_cert", func(p *web_policy.PolicyExtension) string { return p.ShowConfirmationDialogForCachedCert }, func(p *web_policy.PolicyExtension, v string) { p.ShowConfirmationDialogForCachedCert = v }},
	{"one_id_mt_device_auth_enabled", func(p *web_policy.PolicyExtension) string { return p.OneIdMTDeviceAuthEnabled }, func(p *web_policy.PolicyExtension, v string) { p.OneIdMTDeviceAuthEnabled = v }},
	{"prevent_auto_reauth_during_device_lock", func(p *web_policy.PolicyExtension) string { return p.PreventAutoReauthDuringDeviceLock }, func(p *web_policy.PolicyExtension, v string) { p.PreventAutoReauthDuringDeviceLock = v }},
	{"enable_network_traffic_process_mapping", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.EnableNetworkTrafficProcessMapping)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.EnableNetworkTrafficProcessMapping = intOrStringFieldFromString(v)
	}},
	{"use_end_point_location_for_dc_selection", func(p *web_policy.PolicyExtension) string { return p.UseEndPointLocationForDCSelection }, func(p *web_policy.PolicyExtension, v string) { p.UseEndPointLocationForDCSelection = v }},
	{"recache_system_proxy", func(p *web_policy.PolicyExtension) string { return p.RecacheSystemProxy }, func(p *web_policy.PolicyExtension, v string) { p.RecacheSystemProxy = v }},
	{"enable_location_policy_override", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.EnableLocationPolicyOverride)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.EnableLocationPolicyOverride = intOrStringFieldFromString(v)
	}},
	{"block_private_relay", func(p *web_policy.PolicyExtension) string { return p.BlockPrivateRelay }, func(p *web_policy.PolicyExtension, v string) { p.BlockPrivateRelay = v }},
	{"enable_automatic_packet_capture", func(p *web_policy.PolicyExtension) string { return p.EnableAutomaticPacketCapture }, func(p *web_policy.PolicyExtension, v string) { p.EnableAutomaticPacketCapture = v }},
	{"enable_apc_for_critical_sections", func(p *web_policy.PolicyExtension) string { return p.EnableAPCforCriticalSections }, func(p *web_policy.PolicyExtension, v string) { p.EnableAPCforCriticalSections = v }},
	{"enable_apc_for_other_sections", func(p *web_policy.PolicyExtension) string { return p.EnableAPCforOtherSections }, func(p *web_policy.PolicyExtension, v string) { p.EnableAPCforOtherSections = v }},
	{"enable_pc_additional_space", func(p *web_policy.PolicyExtension) string { return p.EnablePCAdditionalSpace }, func(p *web_policy.PolicyExtension, v string) { p.EnablePCAdditionalSpace = v }},
	{"enable_custom_proxy_detection", func(p *web_policy.PolicyExtension) string { return p.EnableCustomProxyDetection }, func(p *web_policy.PolicyExtension, v string) { p.EnableCustomProxyDetection = v }},
	{"enable_crash_reporting", func(p *web_policy.PolicyExtension) string { return p.EnableCrashReporting }, func(p *web_policy.PolicyExtension, v string) { p.EnableCrashReporting = v }},
	{"bypass_dns_traffic_using_udp_proxy", func(p *web_policy.PolicyExtension) string { return p.BypassDNSTrafficUsingUDPProxy }, func(p *web_policy.PolicyExtension, v string) { p.BypassDNSTrafficUsingUDPProxy = v }},
	{"reconnect_tun_on_wakeup", func(p *web_policy.PolicyExtension) string { return p.ReconnectTunOnWakeup }, func(p *web_policy.PolicyExtension, v string) { p.ReconnectTunOnWakeup = v }},
	{"rsc_mode_on_all_adapters", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.RscModeOnAllAdapters)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.RscModeOnAllAdapters = intOrStringFieldFromString(v)
	}},
	{"enable_adapter_hardware_offloading", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.EnableAdapterHardwareOffloading)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.EnableAdapterHardwareOffloading = intOrStringFieldFromString(v)
	}},
	{"support_zpa_search_domains_in_trp", func(p *web_policy.PolicyExtension) string {
		return intOrStringFieldToString(p.SupportZPASearchDomainsInTRP)
	}, func(p *web_policy.PolicyExtension, v string) {
		p.SupportZPASearchDomainsInTRP = intOrStringFieldFromString(v)
	}},
}

// PolicyExtensionAttrTypes returns the attr.Type map for the
// policy_extension nested object. It aggregates three registries:
//
//   - policyExtensionStringFields  → types.StringType
//   - policyExtensionBoolFields    → types.BoolType (0/1 toggles)
//
// plus the small handful of typed scalars, list fields, and the nested
// generate_cli_password_contract object.
func PolicyExtensionAttrTypes() map[string]attr.Type {
	out := make(map[string]attr.Type, len(policyExtensionStringFields)+len(policyExtensionBoolFields)+8)
	for _, f := range policyExtensionStringFields {
		out[f.name] = types.StringType
	}
	for _, f := range policyExtensionBoolFields {
		out[f.name] = types.BoolType
	}
	out["advance_zpa_reauth"] = types.BoolType
	out["advance_zpa_reauth_time"] = types.Int64Type
	out["machine_idp_auth"] = types.BoolType
	out["reactivate_anti_tampering_time"] = types.Int64Type
	out["zpa_auth_exp_session_lock_state_min_time_in_second"] = types.Int64Type
	// Additional scalars surfaced from the macOS payload: numeric timeouts
	// (IntOrString on the wire), a UI-language picker integer, and the
	// pair of plain-int 0/1 toggles that don't belong in the string-or-
	// bool registries above.
	out["zpa_auto_reauth_timeout"] = types.Int64Type
	out["client_connector_ui_language"] = types.Int64Type
	out["enable_local_packet_capture_v2"] = types.BoolType
	out["enable_custom_theme"] = types.BoolType
	out["zcc_fail_close_settings_app_by_pass_ids"] = types.ListType{ElemType: types.Int64Type}
	out["zcc_fail_close_settings_app_by_pass_names"] = types.ListType{ElemType: types.StringType}
	out["generate_cli_password_contract"] = types.ObjectType{AttrTypes: generateCliPasswordContractAttrTypes()}
	return out
}

// PolicyExtensionObjectType returns the types.ObjectType built from
// PolicyExtensionAttrTypes — convenient when an outer attribute needs to
// reference the policy_extension type by value.
func PolicyExtensionObjectType() attr.Type {
	return types.ObjectType{AttrTypes: PolicyExtensionAttrTypes()}
}

// PolicyExtensionAttributes returns the schema attributes for the
// policy_extension SingleNestedAttribute. Genuine string fields come
// from policyExtensionStringFields; 0/1 toggle fields come from
// policyExtensionBoolFields and are surfaced as BoolAttribute so HCL
// stays readable (`use_v8_js_engine = true` instead of `= "1"`).
func PolicyExtensionAttributes() map[string]schema.Attribute {
	out := make(map[string]schema.Attribute, len(policyExtensionStringFields)+len(policyExtensionBoolFields)+8)
	for _, f := range policyExtensionStringFields {
		out[f.name] = schema.StringAttribute{Optional: true, Computed: true}
	}
	for _, f := range policyExtensionBoolFields {
		out[f.name] = schema.BoolAttribute{Optional: true, Computed: true}
	}
	out["advance_zpa_reauth"] = schema.BoolAttribute{Optional: true, Computed: true}
	out["advance_zpa_reauth_time"] = schema.Int64Attribute{Optional: true, Computed: true}
	out["machine_idp_auth"] = schema.BoolAttribute{Optional: true, Computed: true}
	out["reactivate_anti_tampering_time"] = schema.Int64Attribute{Optional: true, Computed: true}
	out["zpa_auth_exp_session_lock_state_min_time_in_second"] = schema.Int64Attribute{Optional: true, Computed: true}
	out["zpa_auto_reauth_timeout"] = schema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Timeout in minutes after which ZPA reauth is forced. API field: zpaAutoReauthTimeout.",
	}
	out["client_connector_ui_language"] = schema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Client Connector UI language code (0 = Use System Language). API field: clientConnectorUiLanguage.",
	}
	out["enable_local_packet_capture_v2"] = schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Enable the v2 local packet capture pipeline. API field: enableLocalPacketCaptureV2 (wire shape: JSON number 0/1).",
	}
	out["enable_custom_theme"] = schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Enable the custom ZCC client UI theme. API field: enableCustomTheme (wire shape: JSON number 0/1).",
	}
	out["zcc_fail_close_settings_app_by_pass_ids"] = schema.ListAttribute{Optional: true, Computed: true, ElementType: types.Int64Type}
	out["zcc_fail_close_settings_app_by_pass_names"] = schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType}
	out["generate_cli_password_contract"] = schema.SingleNestedAttribute{
		Optional:   true,
		Computed:   true,
		Attributes: generateCliPasswordContractAttributes(),
	}
	return out
}

// ExpandPolicyExtension overlays user-set fields from the Terraform
// policy_extension object onto out. The function is the overlay
// equivalent of the older "rebuild from scratch" expand: anything the
// operator did not set in HCL is left at its caller-seeded default
// (from DefaultMacosWebPolicy or the equivalent per-device-type
// constructor), so partial `policy_extension = { ... }` blocks no
// longer wipe the ~50 UI-shape defaults the API expects to see.
func ExpandPolicyExtension(obj types.Object, out *web_policy.PolicyExtension) {
	if obj.IsNull() || obj.IsUnknown() {
		return
	}
	a := obj.Attributes()

	// String / IntOrString fields driven by policyExtensionStringFields.
	// Each entry already routes through the right adapter; we just have
	// to skip null/unknown so defaults survive.
	for _, f := range policyExtensionStringFields {
		v, ok := a[f.name].(types.String)
		if !ok || v.IsNull() || v.IsUnknown() {
			continue
		}
		f.set(out, v.ValueString())
	}

	// 0/1 toggle fields surfaced to HCL as types.Bool — convert
	// true/false to the literal "1" / "0" strings the SDK setter
	// expects (each registry entry routes that string through the
	// correct String / IntOrString adapter for its backing field).
	for _, f := range policyExtensionBoolFields {
		v, ok := a[f.name].(types.Bool)
		if !ok || v.IsNull() || v.IsUnknown() {
			continue
		}
		f.set(out, helpers.BoolToString01(v))
	}

	overlayBool(&out.AdvanceZpaReauth, a, "advance_zpa_reauth")
	overlayInt(&out.AdvanceZpaReauthTime, a, "advance_zpa_reauth_time")
	overlayBool(&out.MachineIdpAuth, a, "machine_idp_auth")
	overlayInt(&out.ReactivateAntiTamperingTime, a, "reactivate_anti_tampering_time")

	// zpa_auth_exp_session_lock_state_min_time_in_second is surfaced to HCL
	// as an Int64Attribute but the API serialises it as a quoted-string
	// number. Only overwrite the default when the operator actually sets
	// the attribute.
	if v, ok := a["zpa_auth_exp_session_lock_state_min_time_in_second"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.ZpaAuthExpSessionLockStateMinTime = strconv.Itoa(int(v.ValueInt64()))
	}

	// Additional scalars overlaid only when the operator set them in HCL.
	// IntOrString-backed timeouts wrap the int64 in the dual-shape wire
	// type; the two plain-int 0/1 toggles round-trip via helpers.
	if v, ok := a["zpa_auto_reauth_timeout"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.ZpaAutoReauthTimeout = zccCommon.IntOrString(v.ValueInt64())
	}
	if v, ok := a["client_connector_ui_language"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.ClientConnectorUiLanguage = zccCommon.IntOrString(v.ValueInt64())
	}
	if v, ok := a["enable_local_packet_capture_v2"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.EnableLocalPacketCaptureV2 = helpers.BoolToInt(v)
	}
	if v, ok := a["enable_custom_theme"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.EnableCustomTheme = helpers.BoolToInt(v)
	}

	overlayIntList(&out.ZccFailCloseSettingsAppByPassIds, a, "zcc_fail_close_settings_app_by_pass_ids")
	overlayStringList(&out.ZccFailCloseSettingsAppByPassNames, a, "zcc_fail_close_settings_app_by_pass_names")

	if cli, ok := a["generate_cli_password_contract"].(types.Object); ok && !cli.IsNull() && !cli.IsUnknown() {
		overlayCliPasswordContract(cli, &out.GenerateCliPasswordContract)
	}
}

// FlattenPolicyExtension converts the SDK PolicyExtension struct returned
// by the API into a Plugin Framework types.Object whose AttrTypes match
// PolicyExtensionAttrTypes.
func FlattenPolicyExtension(p web_policy.PolicyExtension) types.Object {
	values := make(map[string]attr.Value, len(policyExtensionStringFields)+len(policyExtensionBoolFields)+8)
	for _, f := range policyExtensionStringFields {
		values[f.name] = types.StringValue(f.get(&p))
	}
	for _, f := range policyExtensionBoolFields {
		values[f.name] = helpers.String01ToBool(f.get(&p))
	}
	values["advance_zpa_reauth"] = types.BoolValue(p.AdvanceZpaReauth)
	values["advance_zpa_reauth_time"] = types.Int64Value(int64(p.AdvanceZpaReauthTime))
	values["machine_idp_auth"] = types.BoolValue(p.MachineIdpAuth)
	values["reactivate_anti_tampering_time"] = types.Int64Value(int64(p.ReactivateAntiTamperingTime))
	zpaAuthExpSessionLockN, _ := strconv.Atoi(p.ZpaAuthExpSessionLockStateMinTime)
	values["zpa_auth_exp_session_lock_state_min_time_in_second"] = types.Int64Value(int64(zpaAuthExpSessionLockN))
	values["zpa_auto_reauth_timeout"] = types.Int64Value(int64(p.ZpaAutoReauthTimeout))
	values["client_connector_ui_language"] = types.Int64Value(int64(p.ClientConnectorUiLanguage))
	values["enable_local_packet_capture_v2"] = helpers.IntToBool(p.EnableLocalPacketCaptureV2)
	values["enable_custom_theme"] = helpers.IntToBool(p.EnableCustomTheme)
	values["zcc_fail_close_settings_app_by_pass_ids"] = intListValue(p.ZccFailCloseSettingsAppByPassIds)
	values["zcc_fail_close_settings_app_by_pass_names"] = stringListValue(p.ZccFailCloseSettingsAppByPassNames)
	values["generate_cli_password_contract"] = flattenCliPasswordContract(p.GenerateCliPasswordContract)
	obj, _ := types.ObjectValue(PolicyExtensionAttrTypes(), values)
	return obj
}

// generateCliPasswordContractAttrTypes returns the attr.Type map for the
// nested generate_cli_password_contract object inside policy_extension.
// Kept unexported because callers only ever interact with policy_extension
// as a whole.
func generateCliPasswordContractAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allow_zpa_disable_without_password": types.BoolType,
		"allow_zia_disable_without_password": types.BoolType,
		"allow_zdx_disable_without_password": types.BoolType,
		"enable_cli":                         types.BoolType,
		"policy_id":                          types.Int64Type,
	}
}

// generateCliPasswordContractAttributes returns the schema attributes for
// the nested generate_cli_password_contract object. Unexported for the
// same reason as the AttrTypes helper.
func generateCliPasswordContractAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allow_zpa_disable_without_password": schema.BoolAttribute{Optional: true, Computed: true},
		"allow_zia_disable_without_password": schema.BoolAttribute{Optional: true, Computed: true},
		"allow_zdx_disable_without_password": schema.BoolAttribute{Optional: true, Computed: true},
		"enable_cli":                         schema.BoolAttribute{Optional: true, Computed: true},
		"policy_id":                          schema.Int64Attribute{Optional: true, Computed: true},
	}
}

// expandCliPasswordContract converts the Plugin Framework object into the
// SDK GenerateCliPasswordContract struct used inside PolicyExtension.
// overlayCliPasswordContract overlays user-set fields from the
// Terraform generate_cli_password_contract object onto out. The
// overlay (not rebuild) semantics let the per-OS default keep its
// allow_*_disable_without_password trio set to true even when the
// operator omits the nested block (or only sets a subset).
func overlayCliPasswordContract(obj types.Object, out *web_policy.GenerateCliPasswordContract) {
	if obj.IsNull() || obj.IsUnknown() {
		return
	}
	a := obj.Attributes()
	overlayBool(&out.AllowZpaDisableWithoutPassword, a, "allow_zpa_disable_without_password")
	overlayBool(&out.AllowZiaDisableWithoutPassword, a, "allow_zia_disable_without_password")
	overlayBool(&out.AllowZdxDisableWithoutPassword, a, "allow_zdx_disable_without_password")
	overlayBool(&out.EnableCli, a, "enable_cli")
	overlayInt(&out.PolicyId, a, "policy_id")
}

// flattenCliPasswordContract converts the SDK GenerateCliPasswordContract
// returned by the API into a Plugin Framework types.Object.
func flattenCliPasswordContract(c web_policy.GenerateCliPasswordContract) types.Object {
	obj, _ := types.ObjectValue(generateCliPasswordContractAttrTypes(), map[string]attr.Value{
		"allow_zpa_disable_without_password": types.BoolValue(c.AllowZpaDisableWithoutPassword),
		"allow_zia_disable_without_password": types.BoolValue(c.AllowZiaDisableWithoutPassword),
		"allow_zdx_disable_without_password": types.BoolValue(c.AllowZdxDisableWithoutPassword),
		"enable_cli":                         types.BoolValue(c.EnableCli),
		"policy_id":                          types.Int64Value(int64(c.PolicyId)),
	})
	return obj
}

// =============================================================================
// Top-level expand/flatten
// =============================================================================

// normalizeActive maps any of the boolean-ish spellings a Terraform user
// might type for the `active` attribute into the exact "1" / "0" form
// the ZCC API actually accepts. The Python reference implementation
// always emits "1" / "0", and the API rejects "true" / "false" with a
// 400. Empty input passes through unchanged so the API can decide.
func normalizeActive(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return "1"
	case "false", "0", "no", "off":
		return "0"
	default:
		return s
	}
}

// ExpandWebPolicyBase overlays user-set fields from the resource model
// onto payload. The per-OS resource is expected to seed `payload` with a
// per-device-type default (see NewDefaultWebPolicyForDeviceType) before
// calling this — that way any attribute the user left unset stays at
// the API-correct default rather than being clobbered with a Go zero
// value.
//
// `active` is normalized to "1" / "0" because the API rejects the more
// natural-looking "true" / "false" form with a 400.
func ExpandWebPolicyBase(base *WebPolicyBaseModel, payload *web_policy.WebPolicy) {
	setString := func(dst *string, src types.String) {
		if src.IsNull() || src.IsUnknown() {
			return
		}
		*dst = src.ValueString()
	}
	setInt := func(dst *int, src types.Int64) {
		if src.IsNull() || src.IsUnknown() {
			return
		}
		*dst = int(src.ValueInt64())
	}
	setIntOrString := func(dst *zccCommon.IntOrString, src types.Int64) {
		if src.IsNull() || src.IsUnknown() {
			return
		}
		*dst = zccCommon.IntOrString(src.ValueInt64())
	}
	setBool := func(dst *bool, src types.Bool) {
		if src.IsNull() || src.IsUnknown() {
			return
		}
		*dst = src.ValueBool()
	}
	setIntList := func(dst *[]int, src types.List) {
		if src.IsNull() || src.IsUnknown() {
			return
		}
		*dst = intListFromList(src)
	}
	setStringList := func(dst *[]string, src types.List) {
		if src.IsNull() || src.IsUnknown() {
			return
		}
		*dst = stringListFromList(src)
	}

	setString(&payload.ID, base.ID)
	setString(&payload.Name, base.Name)
	if !base.Active.IsNull() && !base.Active.IsUnknown() {
		payload.Active = normalizeActive(base.Active.ValueString())
	}
	setString(&payload.Description, base.Description)
	setBool(&payload.AllowUnreachablePac, base.AllowUnreachablePac)
	setIntOrString(&payload.HighlightActiveControl, base.HighlightActiveControl)
	setIntOrString(&payload.LogFileSize, base.LogFileSize)
	setIntOrString(&payload.LogLevel, base.LogLevel)
	setIntOrString(&payload.LogMode, base.LogMode)
	setString(&payload.PacURL, base.PacURL)
	setIntOrString(&payload.ReactivateWebSecurityMins, base.ReactivateWebSecurityMins)
	setIntOrString(&payload.ReauthPeriod, base.ReauthPeriod)
	setIntOrString(&payload.RuleOrder, base.RuleOrder)
	setIntOrString(&payload.SendDisableServiceReason, base.SendDisableServiceReason)
	setIntOrString(&payload.TunnelZappTraffic, base.TunnelZappTraffic)
	setIntOrString(&payload.GroupAll, base.GroupAll)
	setIntOrString(&payload.EnableDeviceGroups, base.EnableDeviceGroups)
	setInt(&payload.ForwardingProfileId, base.ForwardingProfileId)
	setInt(&payload.ZiaPostureConfigId, base.ZiaPostureConfigId)
	setIntList(&payload.GroupIds, base.GroupIds)
	setStringList(&payload.GroupNames, base.GroupNames)
	setIntList(&payload.UserIds, base.UserIds)
	setStringList(&payload.UserNames, base.UserNames)
	setIntList(&payload.AppServiceIds, base.AppServiceIds)
	setStringList(&payload.AppServiceNames, base.AppServiceNames)
	setStringList(&payload.AppIdentityNames, base.AppIdentityNames)
	setIntList(&payload.BypassAppIds, base.BypassAppIds)
	setIntList(&payload.BypassCustomAppIds, base.BypassCustomAppIds)
	setIntList(&payload.DeviceGroupIds, base.DeviceGroupIds)
	setStringList(&payload.DeviceGroupNames, base.DeviceGroupNames)
	ExpandDisasterRecovery(base.DisasterRecovery, &payload.DisasterRecovery)
	ExpandPolicyExtension(base.PolicyExtension, &payload.PolicyExtension)
}

/*
// NewDefaultWebPolicyForDeviceType returns the per-device-type baseline
// WebPolicy the provider should seed an /edit request with. Each return
// value mirrors the byte-for-byte shape a fresh UI save produces (see
// web_policy.DefaultMacosWebPolicy and the docs/local_dev fixtures).
// Device types we have not yet captured fall back to a zero-value
// WebPolicy.
func NewDefaultWebPolicyForDeviceType(deviceType int) web_policy.WebPolicy {
	switch deviceType {
	case zccCommon.DeviceTypeMacOS:
		return web_policy.DefaultMacosWebPolicy()
	case zccCommon.DeviceTypeIOS:
		return web_policy.DefaultIosWebPolicy()
	default:
		return web_policy.WebPolicy{}
	}
}
*/
// FlattenWebPolicyBase populates every shared field from the GET response
// into the embedded base model. OS-specific blocks are flattened in the
// calling resource using its own helper.
//
// Note on device_type: the API returns it as a JSON number that this
// function maps to a friendly label ("iOS", "macOS", ...) for the
// Terraform state. Each per-OS resource is single-purpose (
// zcc_app_profile_macos always means deviceType=4, zcc_app_profile_ios
// always means deviceType=1, etc.), so device_type is NEVER something
// the operator configures — the schema attribute is Computed-only. To
// guarantee the state value matches the resource's intent even if the
// API ever echoes back an unexpected 0 (e.g. via a partial-decode
// fallback), the shared RunUpsert / RunRead helpers explicitly
// overwrite base.DeviceType with the friendly label derived from the
// deviceType passed in by the resource — see the corresponding override
// at the bottom of those two functions.
func FlattenWebPolicyBase(p *web_policy.WebPolicy, base *WebPolicyBaseModel) {
	base.ID = types.StringValue(p.ID)
	base.Name = types.StringValue(p.Name)
	base.Active = types.StringValue(p.Active)
	base.Description = types.StringValue(p.Description)
	base.DeviceType = types.StringValue(zccCommon.GetDeviceTypeName(p.DeviceType))
	base.AllowUnreachablePac = types.BoolValue(p.AllowUnreachablePac)
	base.HighlightActiveControl = types.Int64Value(int64(p.HighlightActiveControl))
	base.LogFileSize = types.Int64Value(int64(p.LogFileSize))
	base.LogLevel = types.Int64Value(int64(p.LogLevel))
	base.LogMode = types.Int64Value(int64(p.LogMode))
	base.PacURL = types.StringValue(p.PacURL)
	base.ReactivateWebSecurityMins = types.Int64Value(int64(p.ReactivateWebSecurityMins))
	base.ReauthPeriod = types.Int64Value(int64(p.ReauthPeriod))
	base.RuleOrder = types.Int64Value(int64(p.RuleOrder))
	base.SendDisableServiceReason = types.Int64Value(int64(p.SendDisableServiceReason))
	base.TunnelZappTraffic = types.Int64Value(int64(p.TunnelZappTraffic))
	base.GroupAll = types.Int64Value(int64(p.GroupAll))
	base.EnableDeviceGroups = types.Int64Value(int64(p.EnableDeviceGroups))
	base.ForwardingProfileId = types.Int64Value(int64(p.ForwardingProfileId))
	base.ZiaPostureConfigId = types.Int64Value(int64(p.ZiaPostureConfigId))
	base.GroupIds = intListValue(p.GroupIds)
	base.GroupNames = stringListValue(p.GroupNames)
	base.UserIds = intListValue(p.UserIds)
	base.UserNames = stringListValue(p.UserNames)
	base.AppServiceIds = intListValue(p.AppServiceIds)
	base.AppServiceNames = stringListValue(p.AppServiceNames)
	base.AppIdentityNames = stringListValue(p.AppIdentityNames)
	base.BypassAppIds = intListValue(p.BypassAppIds)
	base.BypassCustomAppIds = intListValue(p.BypassCustomAppIds)
	base.DeviceGroupIds = intListValue(p.DeviceGroupIds)
	base.DeviceGroupNames = stringListValue(p.DeviceGroupNames)
	base.DisasterRecovery = FlattenDisasterRecovery(p.DisasterRecovery)
	base.PolicyExtension = FlattenPolicyExtension(p.PolicyExtension)
}

// =============================================================================
// Small attribute helpers (unexported — internal to this package)
//
// Most generic attr.Value/list-of-attr extractors have moved to the
// shared internal/framework/helpers package. What stays here is only
// the IntOrString-bridge logic that the policyExtensionStringFields
// machinery depends on (parked alongside the per-OS app_profile_*
// resources under local_dev/Backup_Config_Future/).
// =============================================================================

// intOrStringFieldToString is the bridge between the SDK's
// common.IntOrString fields (numeric on the wire, but historically
// surfaced as schema.StringAttribute in this provider) and a Go string
// the policyExtensionStringFields plumbing expects. It accepts the
// IntOrString value and renders the underlying int as a decimal string
// — so a wire `"0"` HCL value flattens consistently for users.
func intOrStringFieldToString(v zccCommon.IntOrString) string {
	return strconv.Itoa(int(v))
}

// intOrStringFieldFromString converts a Go string (sourced from an
// attr.Value via stringFromAttr) into a common.IntOrString that the SDK
// will marshal as a JSON number. Empty / non-numeric strings collapse
// to 0; the policyExtensionStringFields entries that wrap IntOrString
// SDK fields use this on the setter side.
func intOrStringFieldFromString(s string) zccCommon.IntOrString {
	if s == "" {
		return zccCommon.IntOrString(0)
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return zccCommon.IntOrString(0)
	}
	return zccCommon.IntOrString(n)
}

// intListFromList materialises a types.List of Int64 values into a Go
// []int. Null/unknown lists become an empty slice so the SDK payload
// always has a defined value.
func intListFromList(l types.List) []int {
	if l.IsNull() || l.IsUnknown() {
		return []int{}
	}
	elems := l.Elements()
	out := make([]int, 0, len(elems))
	for _, e := range elems {
		if v, ok := e.(types.Int64); ok {
			out = append(out, int(v.ValueInt64()))
		}
	}
	return out
}

// stringListFromList materialises a types.List of String values into a Go
// []string. Null/unknown lists become an empty slice.
func stringListFromList(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return []string{}
	}
	elems := l.Elements()
	out := make([]string, 0, len(elems))
	for _, e := range elems {
		if v, ok := e.(types.String); ok {
			out = append(out, v.ValueString())
		}
	}
	return out
}

// intListValue wraps a Go []int into a types.List of Int64 for state
// flattening.
func intListValue(in []int) types.List {
	vals := make([]attr.Value, 0, len(in))
	for _, v := range in {
		vals = append(vals, types.Int64Value(int64(v)))
	}
	out, _ := types.ListValue(types.Int64Type, vals)
	return out
}

// stringListValue wraps a Go []string into a types.List of String for
// state flattening.
func stringListValue(in []string) types.List {
	vals := make([]attr.Value, 0, len(in))
	for _, v := range in {
		vals = append(vals, types.StringValue(v))
	}
	out, _ := types.ListValue(types.StringType, vals)
	return out
}

/*
// =============================================================================
// Lifecycle helpers
// =============================================================================

// RunUpsert is the shared body of Create and Update for every per-OS
// app_profile resource. It expands the model into the SDK payload, calls
// UpdateWebPolicy (the /edit endpoint that creates or updates depending on
// whether `id` is empty), then re-reads the policy by id+deviceType to
// populate state with the server's authoritative view, and finally
// optionally activates it.
//
// The caller supplies expandOSBlock and flattenOSBlock callbacks because
// each per-OS resource owns its own OS-specific schema and SDK block.
func RunUpsert(
	ctx context.Context,
	service *zscaler.Service,
	base *WebPolicyBaseModel,
	deviceType int,
	expandOSBlock func(*web_policy.WebPolicy),
	flattenOSBlock func(*web_policy.WebPolicy),
) error {
	// Seed the payload from the per-device-type baseline so every field
	// the user did not override still travels on the wire with the
	// API-expected default — the API silently rejects payloads that omit
	// any of the ~130 fields the UI emits, so we mirror the UI default
	// (DefaultMacosWebPolicy, etc.) and then overlay the resource model.
	payload := NewDefaultWebPolicyForDeviceType(deviceType)
	ExpandWebPolicyBase(base, &payload)
	expandOSBlock(&payload)
	payload.DeviceType = deviceType

	tflog.Info(ctx, "Submitting web policy /edit", map[string]any{
		"existing_id": payload.ID,
		"name":        payload.Name,
		"device_type": payload.DeviceType,
	})

	resp, err := web_policy.UpdateWebPolicy(ctx, service, &payload)
	if err != nil {
		return fmt.Errorf("write web policy: %w", err)
	}

	newID := resp.ID.String()
	if newID == "" {
		newID = payload.ID
	}
	if newID == "" {
		return fmt.Errorf("API accepted the policy but did not return an id")
	}

	tflog.Info(ctx, "Re-reading web policy", map[string]any{"id": newID, "device_type": deviceType})
	policy, err := web_policy.GetWebPolicyByID(ctx, service, newID, deviceType)
	// The PUT /edit endpoint already echoed back success:true plus the
	// new id, so the policy is on the server regardless of what happens
	// on the follow-up re-read. There are three outcomes to handle:
	//
	//   1. err == nil and the policy strict-decoded cleanly. Use the
	//      server's view as authoritative.
	//   2. err is ErrWebPolicyPartialDecode — the entry was found, but
	//      the strict-typed WebPolicy could not represent every field
	//      the API echoed back (typically a JSON-number-where-string
	//      mismatch on one of the ~130 fields). Use the just-sent
	//      payload as state since that is exactly what we wrote and the
	//      API confirmed.
	//   3. Any other err — list call failed, true 404, transient API
	//      issue. Same fallback: use the payload. Next terraform
	//      refresh will reconcile against the live server view.
	if err != nil {
		if errors.Is(err, web_policy.ErrWebPolicyPartialDecode) {
			tflog.Warn(ctx, "Re-read decoded partially; using request body as authoritative state", map[string]any{
				"id":          newID,
				"device_type": deviceType,
				"error":       err.Error(),
			})
		} else {
			tflog.Warn(ctx, "Re-read after write failed; using request body as authoritative state", map[string]any{
				"id":          newID,
				"device_type": deviceType,
				"error":       err.Error(),
			})
		}
		payload.ID = newID
		payload.DeviceType = deviceType
		policy = &payload
	}

	FlattenWebPolicyBase(policy, base)
	flattenOSBlock(policy)

	// device_type is hard-coded per resource (NewAppProfileMacosResource
	// always operates on deviceType=4, NewAppProfileIosResource on
	// deviceType=1, and so on). The API echoes the integer back on reads
	// so FlattenWebPolicyBase derives a friendly label from `p.DeviceType`,
	// but a partial-decode fallback or a server hiccup can leave that
	// integer at 0 — which would flip the state's `device_type` to "".
	// The resource itself is the authoritative source of truth, so we
	// always overwrite the flattened value with the friendly name
	// derived from the per-OS const passed in by the resource.
	base.DeviceType = types.StringValue(zccCommon.GetDeviceTypeName(deviceType))

	if base.Activate.ValueBool() {
		idInt, convErr := strconv.Atoi(newID)
		if convErr != nil {
			return fmt.Errorf("activate: parse id %q: %w", newID, convErr)
		}
		tflog.Info(ctx, "Activating web policy", map[string]any{"id": newID, "device_type": deviceType})
		if _, err := web_policy.ActivateWebPolicy(ctx, service, &web_policy.WebPolicyActivation{
			PolicyId:   idInt,
			DeviceType: deviceType,
		}); err != nil {
			return fmt.Errorf("activate web policy: %w", err)
		}
	}

	return nil
}

// RunRead fetches the policy by its stored id and refreshes state. It
// returns (true, nil) if the policy was found, (false, nil) if it has
// been deleted out-of-band (signal to the caller to drop it from state),
// or (false, err) for any other failure — callers must surface that
// error so a transient API issue is not mistaken for a missing resource.
func RunRead(
	ctx context.Context,
	service *zscaler.Service,
	base *WebPolicyBaseModel,
	deviceType int,
	flattenOSBlock func(*web_policy.WebPolicy),
) (bool, error) {
	id := base.ID.ValueString()
	if id == "" {
		return false, nil
	}
	policy, err := web_policy.GetWebPolicyByID(ctx, service, id, deviceType)
	if err != nil {
		if isWebPolicyNotFound(err) {
			tflog.Info(ctx, "Web policy not found, removing from state", map[string]any{"id": id, "device_type": deviceType, "error": err.Error()})
			return false, nil
		}
		// Strict-decode mismatch — the policy IS on the server (the id
		// was found in the list), but the response carries field shapes
		// the typed struct cannot represent. Keep the existing state
		// intact rather than surfacing a transient error to Terraform
		// or dropping the resource; drift detection on those fields is
		// the only thing temporarily reduced.
		if errors.Is(err, web_policy.ErrWebPolicyPartialDecode) {
			tflog.Warn(ctx, "Web policy read decoded partially; keeping existing state", map[string]any{
				"id":          id,
				"device_type": deviceType,
				"error":       err.Error(),
			})
			return true, nil
		}
		return false, fmt.Errorf("read web policy %s (deviceType=%d): %w", id, deviceType, err)
	}
	FlattenWebPolicyBase(policy, base)
	flattenOSBlock(policy)

	// See the same-named override in RunUpsert: device_type is
	// hard-coded per resource and is the authoritative source of truth
	// for the state value, not whatever the API echoed back.
	base.DeviceType = types.StringValue(zccCommon.GetDeviceTypeName(deviceType))
	return true, nil
}

// isWebPolicyNotFound reports whether err from GetWebPolicyByID indicates
// the policy is genuinely missing on the server (HTTP 404 from the
// underlying list call, or the id was simply absent from the
// listByCompany result). Any other error — auth, network, rate-limit —
// must NOT be treated as a missing resource because Terraform would then
// drop it from state and recreate it on the next apply, deactivating the
// live policy mid-cycle.
func isWebPolicyNotFound(err error) bool {
	if err == nil {
		return false
	}
	var respErr *errorx.ErrorResponse
	if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
		return true
	}
	// GetWebPolicyByID returns a plain fmt.Errorf("web policy with id %q
	// (deviceType=%d) not found", ...) when the id is absent from the
	// list response. Match the exact SDK sentinel — ") not found" — so a
	// decode-time error like "...: cannot unmarshal number into ... not
	// found in registry" is NOT mistaken for a missing resource (which
	// would cause Terraform to drop the live policy from state).
	return strings.HasSuffix(err.Error(), ") not found")
}

// RunDelete deletes the policy by its stored id. The API rejects deletion
// of the default per-OS policy with a typical Zscaler error; the message
// is surfaced to the user verbatim so they can use `terraform state rm`
// to unmanage the resource without deleting it server-side.
func RunDelete(ctx context.Context, service *zscaler.Service, idStr string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("delete: parse id %q: %w", idStr, err)
	}
	tflog.Info(ctx, "Deleting web policy", map[string]any{"id": idStr})
	if _, err := web_policy.DeleteWebPolicy(ctx, service, id); err != nil {
		return fmt.Errorf("delete web policy %d: %w", id, err)
	}
	return nil
}
*/
