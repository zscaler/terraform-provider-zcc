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

const appProfileLinuxDeviceType = zccCommon.DeviceTypeLinux

var (
	_ resource.Resource                = &AppProfileLinuxResource{}
	_ resource.ResourceWithConfigure   = &AppProfileLinuxResource{}
	_ resource.ResourceWithImportState = &AppProfileLinuxResource{}
)

func NewAppProfileLinuxResource() resource.Resource {
	return &AppProfileLinuxResource{}
}

type AppProfileLinuxResource struct {
	client *client.Client
}

type AppProfileLinuxResourceModel struct {
	common.WebPolicyBaseModel
	LinuxPolicy types.Object `tfsdk:"linux_policy"`
}

func (r *AppProfileLinuxResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_profile_linux"
}

func (r *AppProfileLinuxResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := common.WebPolicyBaseAttributes()
	attrs["linux_policy"] = schema.SingleNestedAttribute{
		Description: "Linux-specific app profile block (linuxPolicy in the API).",
		Optional:    true,
		Computed:    true,
		Attributes:  linuxPolicyAttributes(),
	}
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC Linux app profile (web policy with deviceType=5). The /web/policy/edit endpoint creates or updates the record; the resource always re-reads via list-by-deviceType after a write so state reflects the server's authoritative view.",
		Attributes:  attrs,
	}
}

func (r *AppProfileLinuxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppProfileLinuxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileLinuxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("")
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileLinuxDeviceType,
		func(p *web_policy.WebPolicy) { p.LinuxPolicy = expandLinuxPolicy(plan.LinuxPolicy) },
		func(p *web_policy.WebPolicy) { plan.LinuxPolicy = flattenLinuxPolicy(p.LinuxPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileLinuxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileLinuxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := common.RunRead(ctx, r.client.Service, &state.WebPolicyBaseModel, appProfileLinuxDeviceType,
		func(p *web_policy.WebPolicy) { state.LinuxPolicy = flattenLinuxPolicy(p.LinuxPolicy) },
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

func (r *AppProfileLinuxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileLinuxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AppProfileLinuxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileLinuxDeviceType,
		func(p *web_policy.WebPolicy) { p.LinuxPolicy = expandLinuxPolicy(plan.LinuxPolicy) },
		func(p *web_policy.WebPolicy) { plan.LinuxPolicy = flattenLinuxPolicy(p.LinuxPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileLinuxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileLinuxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := common.RunDelete(ctx, r.client.Service, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

func (r *AppProfileLinuxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// linux_policy block
// =============================================================================

func linuxPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"disable_password":   types.StringType,
		"install_certs":      types.StringType,
		"logout_password":    types.StringType,
		"uninstall_password": types.StringType,
	}
}

func linuxPolicyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"disable_password":   schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"install_certs":      schema.StringAttribute{Optional: true, Computed: true},
		"logout_password":    schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"uninstall_password": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
	}
}

// expandLinuxPolicy converts the linux_policy Terraform object into a
// pointer to a SDK LinuxPolicy. Returning a pointer (rather than a
// value) lets WebPolicy.LinuxPolicy use `omitempty`, so a payload for
// one device type never accidentally includes empty blocks for the
// other four.
func expandLinuxPolicy(obj types.Object) *web_policy.LinuxPolicy {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	a := obj.Attributes()
	return &web_policy.LinuxPolicy{
		DisablePassword:   stringFromAttr(a["disable_password"]),
		InstallCerts:      stringFromAttr(a["install_certs"]),
		LogoutPassword:    stringFromAttr(a["logout_password"]),
		UninstallPassword: stringFromAttr(a["uninstall_password"]),
	}
}

// flattenLinuxPolicy builds the Terraform linux_policy object from the
// SDK LinuxPolicy pointer returned by the API. A nil pointer is
// flattened into a null Object so the framework treats it as absent
// rather than an empty object with zero-valued attributes.
func flattenLinuxPolicy(p *web_policy.LinuxPolicy) types.Object {
	if p == nil {
		return types.ObjectNull(linuxPolicyAttrTypes())
	}
	obj, _ := types.ObjectValue(linuxPolicyAttrTypes(), map[string]attr.Value{
		"disable_password":   types.StringValue(p.DisablePassword),
		"install_certs":      types.StringValue(p.InstallCerts),
		"logout_password":    types.StringValue(p.LogoutPassword),
		"uninstall_password": types.StringValue(p.UninstallPassword),
	})
	return obj
}
