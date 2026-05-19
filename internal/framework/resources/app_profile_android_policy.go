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

const appProfileAndroidDeviceType = zccCommon.DeviceTypeAndroid

var (
	_ resource.Resource                = &AppProfileAndroidResource{}
	_ resource.ResourceWithConfigure   = &AppProfileAndroidResource{}
	_ resource.ResourceWithImportState = &AppProfileAndroidResource{}
)

func NewAppProfileAndroidResource() resource.Resource {
	return &AppProfileAndroidResource{}
}

type AppProfileAndroidResource struct {
	client *client.Client
}

type AppProfileAndroidResourceModel struct {
	common.WebPolicyBaseModel
	AndroidPolicy types.Object `tfsdk:"android_policy"`
}

func (r *AppProfileAndroidResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_profile_android"
}

func (r *AppProfileAndroidResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := common.WebPolicyBaseAttributes()
	attrs["android_policy"] = schema.SingleNestedAttribute{
		Description: "Android-specific app profile block (androidPolicy in the API).",
		Optional:    true,
		Computed:    true,
		Attributes:  androidPolicyAttributes(),
	}
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC Android app profile (web policy with deviceType=2). The /web/policy/edit endpoint creates or updates the record; the resource always re-reads via list-by-deviceType after a write so state reflects the server's authoritative view.",
		Attributes:  attrs,
	}
}

func (r *AppProfileAndroidResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppProfileAndroidResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileAndroidResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("")
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileAndroidDeviceType,
		func(p *web_policy.WebPolicy) { p.AndroidPolicy = expandAndroidPolicy(plan.AndroidPolicy) },
		func(p *web_policy.WebPolicy) { plan.AndroidPolicy = flattenAndroidPolicy(p.AndroidPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileAndroidResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileAndroidResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := common.RunRead(ctx, r.client.Service, &state.WebPolicyBaseModel, appProfileAndroidDeviceType,
		func(p *web_policy.WebPolicy) { state.AndroidPolicy = flattenAndroidPolicy(p.AndroidPolicy) },
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

func (r *AppProfileAndroidResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileAndroidResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AppProfileAndroidResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileAndroidDeviceType,
		func(p *web_policy.WebPolicy) { p.AndroidPolicy = expandAndroidPolicy(plan.AndroidPolicy) },
		func(p *web_policy.WebPolicy) { plan.AndroidPolicy = flattenAndroidPolicy(p.AndroidPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileAndroidResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileAndroidResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := common.RunDelete(ctx, r.client.Service, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

func (r *AppProfileAndroidResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// android_policy block
// =============================================================================

func androidPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allowed_apps":       types.StringType,
		"billing_day":        types.StringType,
		"bypass_android_app": types.StringType,
		"bypass_mms_apps":    types.StringType,
		"custom_text":        types.StringType,
		"disable_password":   types.StringType,
		"enable_verbose_log": types.StringType,
		"enforced":           types.StringType,
		"install_certs":      types.StringType,
		"limit":              types.StringType,
		"logout_password":    types.StringType,
		"quota_roaming":      types.StringType,
		"uninstall_password": types.StringType,
		"wifi_ssid":          types.StringType,
	}
}

func androidPolicyAttributes() map[string]schema.Attribute {
	stringOC := func() schema.Attribute { return schema.StringAttribute{Optional: true, Computed: true} }
	stringSensitive := func() schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Sensitive: true}
	}
	return map[string]schema.Attribute{
		"allowed_apps":       stringOC(),
		"billing_day":        stringOC(),
		"bypass_android_app": schema.StringAttribute{Optional: true, Computed: true, Description: "API field bypassAndroidApps."},
		"bypass_mms_apps":    stringOC(),
		"custom_text":        stringOC(),
		"disable_password":   stringSensitive(),
		"enable_verbose_log": stringOC(),
		"enforced":           stringOC(),
		"install_certs":      stringOC(),
		"limit":              stringOC(),
		"logout_password":    stringSensitive(),
		"quota_roaming":      stringOC(),
		"uninstall_password": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Description: "API field uninstallPassword (Go field UninstallPass)."},
		"wifi_ssid":          schema.StringAttribute{Optional: true, Computed: true, Description: "API field wifissid."},
	}
}

// expandAndroidPolicy converts the android_policy Terraform object into
// a pointer to a SDK AndroidPolicy. Returning a pointer (rather than a
// value) lets WebPolicy.AndroidPolicy use `omitempty`, so a payload for
// one device type never accidentally includes empty blocks for the
// other four.
func expandAndroidPolicy(obj types.Object) *web_policy.AndroidPolicy {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	a := obj.Attributes()
	return &web_policy.AndroidPolicy{
		AllowedApps:      stringFromAttr(a["allowed_apps"]),
		BillingDay:       stringFromAttr(a["billing_day"]),
		BypassAndroidApp: stringFromAttr(a["bypass_android_app"]),
		BypassMmsApps:    stringFromAttr(a["bypass_mms_apps"]),
		CustomText:       stringFromAttr(a["custom_text"]),
		DisablePassword:  stringFromAttr(a["disable_password"]),
		EnableVerboseLog: stringFromAttr(a["enable_verbose_log"]),
		Enforced:         stringFromAttr(a["enforced"]),
		InstallCerts:     stringFromAttr(a["install_certs"]),
		Limit:            stringFromAttr(a["limit"]),
		LogoutPassword:   stringFromAttr(a["logout_password"]),
		QuotaRoaming:     stringFromAttr(a["quota_roaming"]),
		UninstallPass:    stringFromAttr(a["uninstall_password"]),
		WifiSsid:         stringFromAttr(a["wifi_ssid"]),
	}
}

// flattenAndroidPolicy builds the Terraform android_policy object from
// the SDK AndroidPolicy pointer returned by the API. A nil pointer is
// flattened into a null Object so the framework treats it as absent
// rather than an empty object with zero-valued attributes.
func flattenAndroidPolicy(p *web_policy.AndroidPolicy) types.Object {
	if p == nil {
		return types.ObjectNull(androidPolicyAttrTypes())
	}
	obj, _ := types.ObjectValue(androidPolicyAttrTypes(), map[string]attr.Value{
		"allowed_apps":       types.StringValue(p.AllowedApps),
		"billing_day":        types.StringValue(p.BillingDay),
		"bypass_android_app": types.StringValue(p.BypassAndroidApp),
		"bypass_mms_apps":    types.StringValue(p.BypassMmsApps),
		"custom_text":        types.StringValue(p.CustomText),
		"disable_password":   types.StringValue(p.DisablePassword),
		"enable_verbose_log": types.StringValue(p.EnableVerboseLog),
		"enforced":           types.StringValue(p.Enforced),
		"install_certs":      types.StringValue(p.InstallCerts),
		"limit":              types.StringValue(p.Limit),
		"logout_password":    types.StringValue(p.LogoutPassword),
		"quota_roaming":      types.StringValue(p.QuotaRoaming),
		"uninstall_password": types.StringValue(p.UninstallPass),
		"wifi_ssid":          types.StringValue(p.WifiSsid),
	})
	return obj
}
