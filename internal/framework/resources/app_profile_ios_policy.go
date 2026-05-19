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

const appProfileIosDeviceType = zccCommon.DeviceTypeIOS

var (
	_ resource.Resource                = &AppProfileIosResource{}
	_ resource.ResourceWithConfigure   = &AppProfileIosResource{}
	_ resource.ResourceWithImportState = &AppProfileIosResource{}
)

func NewAppProfileIosResource() resource.Resource {
	return &AppProfileIosResource{}
}

type AppProfileIosResource struct {
	client *client.Client
}

type AppProfileIosResourceModel struct {
	common.WebPolicyBaseModel
	IosPolicy types.Object `tfsdk:"ios_policy"`
}

func (r *AppProfileIosResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_profile_ios"
}

func (r *AppProfileIosResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := common.WebPolicyBaseAttributes()
	attrs["ios_policy"] = schema.SingleNestedAttribute{
		Description: "iOS-specific app profile block (iosPolicy in the API).",
		Optional:    true,
		Computed:    true,
		Attributes:  iosPolicyAttributes(),
	}
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC iOS app profile (web policy with deviceType=1). The /web/policy/edit endpoint creates or updates the record; the resource always re-reads via list-by-deviceType after a write so state reflects the server's authoritative view.",
		Attributes:  attrs,
	}
}

func (r *AppProfileIosResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppProfileIosResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileIosResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("")
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileIosDeviceType,
		func(p *web_policy.WebPolicy) { p.IosPolicy = expandIosPolicy(plan.IosPolicy) },
		func(p *web_policy.WebPolicy) { plan.IosPolicy = flattenIosPolicy(p.IosPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileIosResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileIosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := common.RunRead(ctx, r.client.Service, &state.WebPolicyBaseModel, appProfileIosDeviceType,
		func(p *web_policy.WebPolicy) { state.IosPolicy = flattenIosPolicy(p.IosPolicy) },
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

func (r *AppProfileIosResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileIosResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AppProfileIosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileIosDeviceType,
		func(p *web_policy.WebPolicy) { p.IosPolicy = expandIosPolicy(plan.IosPolicy) },
		func(p *web_policy.WebPolicy) { plan.IosPolicy = flattenIosPolicy(p.IosPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileIosResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileIosResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := common.RunDelete(ctx, r.client.Service, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

func (r *AppProfileIosResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// ios_policy block
// =============================================================================

func iosPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"disable_password":          types.StringType,
		"ipv6_mode":                 types.StringType,
		"logout_password":           types.StringType,
		"passcode":                  types.StringType,
		"show_vpn_tun_notification": types.StringType,
		"uninstall_password":        types.StringType,
	}
}

func iosPolicyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"disable_password":          schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"ipv6_mode":                 schema.StringAttribute{Optional: true, Computed: true},
		"logout_password":           schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"passcode":                  schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"show_vpn_tun_notification": schema.StringAttribute{Optional: true, Computed: true},
		"uninstall_password":        schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
	}
}

// expandIosPolicy converts the ios_policy Terraform object into a
// pointer to a SDK IosPolicy. Returning a pointer (rather than a value)
// lets WebPolicy.IosPolicy use `omitempty`, so a payload for one device
// type never accidentally includes empty blocks for the other four.
func expandIosPolicy(obj types.Object) *web_policy.IosPolicy {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	a := obj.Attributes()
	return &web_policy.IosPolicy{
		DisablePassword:        stringFromAttr(a["disable_password"]),
		Ipv6Mode:               stringFromAttr(a["ipv6_mode"]),
		LogoutPassword:         stringFromAttr(a["logout_password"]),
		Passcode:               stringFromAttr(a["passcode"]),
		ShowVPNTunNotification: stringFromAttr(a["show_vpn_tun_notification"]),
		UninstallPassword:      stringFromAttr(a["uninstall_password"]),
	}
}

// flattenIosPolicy builds the Terraform ios_policy object from the SDK
// IosPolicy pointer returned by the API. A nil pointer is flattened
// into a null Object so the framework treats it as absent rather than
// an empty object with zero-valued attributes.
func flattenIosPolicy(p *web_policy.IosPolicy) types.Object {
	if p == nil {
		return types.ObjectNull(iosPolicyAttrTypes())
	}
	obj, _ := types.ObjectValue(iosPolicyAttrTypes(), map[string]attr.Value{
		"disable_password":          types.StringValue(p.DisablePassword),
		"ipv6_mode":                 types.StringValue(p.Ipv6Mode),
		"logout_password":           types.StringValue(p.LogoutPassword),
		"passcode":                  types.StringValue(p.Passcode),
		"show_vpn_tun_notification": types.StringValue(p.ShowVPNTunNotification),
		"uninstall_password":        types.StringValue(p.UninstallPassword),
	})
	return obj
}
