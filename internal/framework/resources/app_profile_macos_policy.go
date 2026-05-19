package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zccCommon "github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/common"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/web_policy"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/common"
)

const appProfileMacosDeviceType = zccCommon.DeviceTypeMacOS

var (
	_ resource.Resource                = &AppProfileMacosResource{}
	_ resource.ResourceWithConfigure   = &AppProfileMacosResource{}
	_ resource.ResourceWithImportState = &AppProfileMacosResource{}
)

func NewAppProfileMacosResource() resource.Resource {
	return &AppProfileMacosResource{}
}

type AppProfileMacosResource struct {
	client *client.Client
}

type AppProfileMacosResourceModel struct {
	common.WebPolicyBaseModel
	MacPolicy types.Object `tfsdk:"mac_policy"`
}

func (r *AppProfileMacosResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_profile_macos"
}

func (r *AppProfileMacosResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := common.WebPolicyBaseAttributes()
	attrs["mac_policy"] = schema.SingleNestedAttribute{
		Description: "macOS-specific app profile block (macPolicy in the API).",
		Optional:    true,
		Computed:    true,
		Attributes:  macPolicyAttributes(),
	}
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC macOS app profile (web policy with deviceType=4). The /web/policy/edit endpoint creates or updates the record; the resource always re-reads via list-by-deviceType after a write so state reflects the server's authoritative view.",
		Attributes:  attrs,
	}
}

func (r *AppProfileMacosResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *AppProfileMacosResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileMacosResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("")
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileMacosDeviceType,
		func(p *web_policy.WebPolicy) { p.MacPolicy = expandMacPolicy(plan.MacPolicy) },
		func(p *web_policy.WebPolicy) { plan.MacPolicy = flattenMacPolicy(p.MacPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileMacosResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileMacosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := common.RunRead(ctx, r.client.Service, &state.WebPolicyBaseModel, appProfileMacosDeviceType,
		func(p *web_policy.WebPolicy) { state.MacPolicy = flattenMacPolicy(p.MacPolicy) },
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AppProfileMacosResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileMacosResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AppProfileMacosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileMacosDeviceType,
		func(p *web_policy.WebPolicy) { p.MacPolicy = expandMacPolicy(plan.MacPolicy) },
		func(p *web_policy.WebPolicy) { plan.MacPolicy = flattenMacPolicy(p.MacPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileMacosResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileMacosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := common.RunDelete(ctx, r.client.Service, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

func (r *AppProfileMacosResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// mac_policy block
// =============================================================================

// macPolicyAttrTypes mirrors the schema produced by macPolicyAttributes.
// Field naming follows the wire shape captured from a working
// UI-generated request body: install_ssl_certs is a JSON number (not a
// string, and not the older installCerts spelling), browser_auth_type /
// use_default_browser / captive_portal_config round out the macPolicy
// object the API actually expects.
func macPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"add_ifscope_route":                              types.StringType,
		"browser_auth_type":                              types.Int64Type,
		"cache_system_proxy":                             types.StringType,
		"captive_portal_config":                          types.StringType,
		"clear_arp_cache":                                types.StringType,
		"disable_password":                               types.StringType,
		"dns_priority_ordering":                          types.StringType,
		"dns_priority_ordering_for_trusted_dns_criteria": types.StringType,
		"enable_app_based_bypass":                        types.StringType,
		"enable_zscaler_firewall":                        types.StringType,
		"install_ssl_certs":                              types.Int64Type,
		"logout_password":                                types.StringType,
		"persistent_zscaler_firewall":                    types.StringType,
		"uninstall_password":                             types.StringType,
		"use_default_browser":                            types.Int64Type,
	}
}

func macPolicyAttributes() map[string]schema.Attribute {
	stringOC := func() schema.Attribute { return schema.StringAttribute{Optional: true, Computed: true} }
	stringSensitive := func() schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Sensitive: true}
	}
	int64OC := func(desc string) schema.Attribute {
		return schema.Int64Attribute{Optional: true, Computed: true, Description: desc}
	}
	return map[string]schema.Attribute{
		"add_ifscope_route":  stringOC(),
		"browser_auth_type":  int64OC("Browser authentication mode for macOS (-1 follows the global config). API field: browserAuthType."),
		"cache_system_proxy": stringOC(),
		"captive_portal_config": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "JSON-encoded macOS captive portal config (e.g. `{\"automaticCapture\":1,\"enableCaptivePortalDetection\":1,...}`). API field: captivePortalConfig.",
		},
		"clear_arp_cache":       stringOC(),
		"disable_password":      stringSensitive(),
		"dns_priority_ordering": stringOC(),
		"dns_priority_ordering_for_trusted_dns_criteria": stringOC(),
		"enable_app_based_bypass":                        schema.StringAttribute{Optional: true, Computed: true, Description: "API field enableApplicationBasedBypass."},
		"enable_zscaler_firewall":                        stringOC(),
		"install_ssl_certs":                              int64OC("Install SSL certificates on the endpoint (0/1). Sent as a JSON number — earlier releases used `install_certs` as a string, which the API silently ignored. API field: install_ssl_certs."),
		"logout_password":                                stringSensitive(),
		"persistent_zscaler_firewall":                    stringOC(),
		"uninstall_password":                             stringSensitive(),
		"use_default_browser":                            int64OC("Use the OS default browser instead of the embedded WebView (0/1). API field: useDefaultBrowser."),
	}
}

// expandMacPolicy converts the mac_policy Terraform object into a
// pointer to a SDK MacPolicy. It seeds the result with DefaultMacosMacPolicy()
// (which mirrors a fresh UI-generated macPolicy block byte-for-byte) and
// overlays only the attributes the user explicitly set in HCL — null /
// unknown attributes leave the default intact. A nil block from the user
// is treated as "all defaults", so the macPolicy section is always
// present in the wire payload (the API rejects bodies that omit it).
func expandMacPolicy(obj types.Object) *web_policy.MacPolicy {
	out := web_policy.DefaultMacosMacPolicy()
	if obj.IsNull() || obj.IsUnknown() {
		return out
	}
	a := obj.Attributes()
	setStr := func(dst *string, key string) {
		v, ok := a[key].(types.String)
		if !ok || v.IsNull() || v.IsUnknown() {
			return
		}
		*dst = v.ValueString()
	}
	setInt := func(dst *int, key string) {
		v, ok := a[key].(types.Int64)
		if !ok || v.IsNull() || v.IsUnknown() {
			return
		}
		*dst = int(v.ValueInt64())
	}
	setIntOrString := func(dst *zccCommon.IntOrString, key string) {
		v, ok := a[key].(types.Int64)
		if !ok || v.IsNull() || v.IsUnknown() {
			return
		}
		*dst = zccCommon.IntOrString(v.ValueInt64())
	}
	setStr(&out.AddIfscopeRoute, "add_ifscope_route")
	setInt(&out.BrowserAuthType, "browser_auth_type")
	setStr(&out.CacheSystemProxy, "cache_system_proxy")
	setStr(&out.CaptivePortalConfig, "captive_portal_config")
	setStr(&out.ClearArpCache, "clear_arp_cache")
	setStr(&out.DisablePassword, "disable_password")
	setStr(&out.DnsPriorityOrdering, "dns_priority_ordering")
	setStr(&out.DnsPriorityOrderingForTrustedDnsCrit, "dns_priority_ordering_for_trusted_dns_criteria")
	setStr(&out.EnableAppBasedBypass, "enable_app_based_bypass")
	setStr(&out.EnableZscalerFirewall, "enable_zscaler_firewall")
	setIntOrString(&out.InstallSslCerts, "install_ssl_certs")
	setStr(&out.LogoutPassword, "logout_password")
	setStr(&out.PersistentZscalerFirewall, "persistent_zscaler_firewall")
	setStr(&out.UninstallPassword, "uninstall_password")
	setInt(&out.UseDefaultBrowser, "use_default_browser")
	return out
}

// flattenMacPolicy builds the Terraform mac_policy object from the SDK
// MacPolicy pointer returned by the API. A nil pointer (which is what
// the listByCompany endpoint returns when the policy isn't a macOS one)
// is flattened into a null Object so the framework treats it as absent
// rather than an empty object with zero-valued attributes.
func flattenMacPolicy(p *web_policy.MacPolicy) types.Object {
	if p == nil {
		return types.ObjectNull(macPolicyAttrTypes())
	}
	obj, _ := types.ObjectValue(macPolicyAttrTypes(), map[string]attr.Value{
		"add_ifscope_route":                              types.StringValue(p.AddIfscopeRoute),
		"browser_auth_type":                              types.Int64Value(int64(p.BrowserAuthType)),
		"cache_system_proxy":                             types.StringValue(p.CacheSystemProxy),
		"captive_portal_config":                          types.StringValue(p.CaptivePortalConfig),
		"clear_arp_cache":                                types.StringValue(p.ClearArpCache),
		"disable_password":                               types.StringValue(p.DisablePassword),
		"dns_priority_ordering":                          types.StringValue(p.DnsPriorityOrdering),
		"dns_priority_ordering_for_trusted_dns_criteria": types.StringValue(p.DnsPriorityOrderingForTrustedDnsCrit),
		"enable_app_based_bypass":                        types.StringValue(p.EnableAppBasedBypass),
		"enable_zscaler_firewall":                        types.StringValue(p.EnableZscalerFirewall),
		"install_ssl_certs":                              types.Int64Value(int64(p.InstallSslCerts)),
		"logout_password":                                types.StringValue(p.LogoutPassword),
		"persistent_zscaler_firewall":                    types.StringValue(p.PersistentZscalerFirewall),
		"uninstall_password":                             types.StringValue(p.UninstallPassword),
		"use_default_browser":                            types.Int64Value(int64(p.UseDefaultBrowser)),
	})
	return obj
}
