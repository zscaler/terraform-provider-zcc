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

const appProfileWindowsDeviceType = zccCommon.DeviceTypeWindows

var (
	_ resource.Resource                = &AppProfileWindowsResource{}
	_ resource.ResourceWithConfigure   = &AppProfileWindowsResource{}
	_ resource.ResourceWithImportState = &AppProfileWindowsResource{}
)

func NewAppProfileWindowsResource() resource.Resource {
	return &AppProfileWindowsResource{}
}

type AppProfileWindowsResource struct {
	client *client.Client
}

type AppProfileWindowsResourceModel struct {
	common.WebPolicyBaseModel
	WindowsPolicy types.Object `tfsdk:"windows_policy"`
}

func (r *AppProfileWindowsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_profile_windows"
}

func (r *AppProfileWindowsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := common.WebPolicyBaseAttributes()
	attrs["windows_policy"] = schema.SingleNestedAttribute{
		Description: "Windows-specific app profile block (windowsPolicy in the API).",
		Optional:    true,
		Computed:    true,
		Attributes:  windowsPolicyAttributes(),
	}
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC Windows app profile (web policy with deviceType=3). The /web/policy/edit endpoint creates or updates the record; the resource always re-reads via list-by-deviceType after a write so state reflects the server's authoritative view.",
		Attributes:  attrs,
	}
}

func (r *AppProfileWindowsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppProfileWindowsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileWindowsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue("")
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileWindowsDeviceType,
		func(p *web_policy.WebPolicy) { p.WindowsPolicy = expandWindowsPolicy(plan.WindowsPolicy) },
		func(p *web_policy.WebPolicy) { plan.WindowsPolicy = flattenWindowsPolicy(p.WindowsPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileWindowsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, err := common.RunRead(ctx, r.client.Service, &state.WebPolicyBaseModel, appProfileWindowsDeviceType,
		func(p *web_policy.WebPolicy) { state.WindowsPolicy = flattenWindowsPolicy(p.WindowsPolicy) },
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

func (r *AppProfileWindowsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var plan AppProfileWindowsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state AppProfileWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	if err := common.RunUpsert(ctx, r.client.Service, &plan.WebPolicyBaseModel, appProfileWindowsDeviceType,
		func(p *web_policy.WebPolicy) { p.WindowsPolicy = expandWindowsPolicy(plan.WindowsPolicy) },
		func(p *web_policy.WebPolicy) { plan.WindowsPolicy = flattenWindowsPolicy(p.WindowsPolicy) },
	); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AppProfileWindowsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}
	var state AppProfileWindowsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := common.RunDelete(ctx, r.client.Service, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
	}
}

func (r *AppProfileWindowsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// windows_policy block
// =============================================================================

func windowsPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cache_system_proxy":               types.Int64Type,
		"captive_portal_config":            types.StringType,
		"disable_loop_back_restriction":    types.Int64Type,
		"disable_parallel_ipv4_and_ipv6":   types.StringType,
		"disable_password":                 types.StringType,
		"flow_logger_config":               types.StringType,
		"force_location_refresh_sccm":      types.Int64Type,
		"install_certs":                    types.StringType,
		"install_windows_firewall_inbound": types.Int64Type,
		"logout_password":                  types.StringType,
		"override_wpad":                    types.Int64Type,
		"pac_data_path":                    types.StringType,
		"pac_type":                         types.Int64Type,
		"prioritize_ipv4":                  types.Int64Type,
		"remove_exempted_containers":       types.Int64Type,
		"restart_win_http_svc":             types.Int64Type,
		"trigger_domain_profile_detection": types.Int64Type,
		"uninstall_password":               types.StringType,
		"wfp_driver":                       types.Int64Type,
	}
}

func windowsPolicyAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cache_system_proxy":               schema.Int64Attribute{Optional: true, Computed: true},
		"captive_portal_config":            schema.StringAttribute{Optional: true, Computed: true},
		"disable_loop_back_restriction":    schema.Int64Attribute{Optional: true, Computed: true},
		"disable_parallel_ipv4_and_ipv6":   schema.StringAttribute{Optional: true, Computed: true},
		"disable_password":                 schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"flow_logger_config":               schema.StringAttribute{Optional: true, Computed: true},
		"force_location_refresh_sccm":      schema.Int64Attribute{Optional: true, Computed: true},
		"install_certs":                    schema.StringAttribute{Optional: true, Computed: true},
		"install_windows_firewall_inbound": schema.Int64Attribute{Optional: true, Computed: true, Description: "API field installWindowsFirewallInboundRule."},
		"logout_password":                  schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"override_wpad":                    schema.Int64Attribute{Optional: true, Computed: true},
		"pac_data_path":                    schema.StringAttribute{Optional: true, Computed: true},
		"pac_type":                         schema.Int64Attribute{Optional: true, Computed: true},
		"prioritize_ipv4":                  schema.Int64Attribute{Optional: true, Computed: true},
		"remove_exempted_containers":       schema.Int64Attribute{Optional: true, Computed: true},
		"restart_win_http_svc":             schema.Int64Attribute{Optional: true, Computed: true},
		"trigger_domain_profile_detection": schema.Int64Attribute{Optional: true, Computed: true, Description: "API field triggerDomainProfleDetection (sic)."},
		"uninstall_password":               schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
		"wfp_driver":                       schema.Int64Attribute{Optional: true, Computed: true},
	}
}

// expandWindowsPolicy converts the windows_policy Terraform object into
// a pointer to a SDK WindowsPolicy. Returning a pointer (rather than a
// value) lets WebPolicy.WindowsPolicy use `omitempty`, so a payload for
// one device type never accidentally includes empty blocks for the
// other four.
func expandWindowsPolicy(obj types.Object) *web_policy.WindowsPolicy {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	a := obj.Attributes()
	return &web_policy.WindowsPolicy{
		CacheSystemProxy:              intFromAttr(a["cache_system_proxy"]),
		CaptivePortalConfig:           stringFromAttr(a["captive_portal_config"]),
		DisableLoopBackRestriction:    intFromAttr(a["disable_loop_back_restriction"]),
		DisableParallelIpv4andIpv6:    stringFromAttr(a["disable_parallel_ipv4_and_ipv6"]),
		DisablePassword:               stringFromAttr(a["disable_password"]),
		FlowLoggerConfig:              stringFromAttr(a["flow_logger_config"]),
		ForceLocationRefreshSccm:      intFromAttr(a["force_location_refresh_sccm"]),
		InstallCerts:                  stringFromAttr(a["install_certs"]),
		InstallWindowsFirewallInbound: intFromAttr(a["install_windows_firewall_inbound"]),
		LogoutPassword:                stringFromAttr(a["logout_password"]),
		OverrideWPAD:                  intFromAttr(a["override_wpad"]),
		PacDataPath:                   stringFromAttr(a["pac_data_path"]),
		PacType:                       intFromAttr(a["pac_type"]),
		PrioritizeIPv4:                intFromAttr(a["prioritize_ipv4"]),
		RemoveExemptedContainers:      intFromAttr(a["remove_exempted_containers"]),
		RestartWinHttpSvc:             intFromAttr(a["restart_win_http_svc"]),
		TriggerDomainProfleDetection:  intFromAttr(a["trigger_domain_profile_detection"]),
		UninstallPassword:             stringFromAttr(a["uninstall_password"]),
		WfpDriver:                     intFromAttr(a["wfp_driver"]),
	}
}

// flattenWindowsPolicy builds the Terraform windows_policy object from
// the SDK WindowsPolicy pointer returned by the API. A nil pointer is
// flattened into a null Object so the framework treats it as absent
// rather than an empty object with zero-valued attributes.
func flattenWindowsPolicy(p *web_policy.WindowsPolicy) types.Object {
	if p == nil {
		return types.ObjectNull(windowsPolicyAttrTypes())
	}
	obj, _ := types.ObjectValue(windowsPolicyAttrTypes(), map[string]attr.Value{
		"cache_system_proxy":               types.Int64Value(int64(p.CacheSystemProxy)),
		"captive_portal_config":            types.StringValue(p.CaptivePortalConfig),
		"disable_loop_back_restriction":    types.Int64Value(int64(p.DisableLoopBackRestriction)),
		"disable_parallel_ipv4_and_ipv6":   types.StringValue(p.DisableParallelIpv4andIpv6),
		"disable_password":                 types.StringValue(p.DisablePassword),
		"flow_logger_config":               types.StringValue(p.FlowLoggerConfig),
		"force_location_refresh_sccm":      types.Int64Value(int64(p.ForceLocationRefreshSccm)),
		"install_certs":                    types.StringValue(p.InstallCerts),
		"install_windows_firewall_inbound": types.Int64Value(int64(p.InstallWindowsFirewallInbound)),
		"logout_password":                  types.StringValue(p.LogoutPassword),
		"override_wpad":                    types.Int64Value(int64(p.OverrideWPAD)),
		"pac_data_path":                    types.StringValue(p.PacDataPath),
		"pac_type":                         types.Int64Value(int64(p.PacType)),
		"prioritize_ipv4":                  types.Int64Value(int64(p.PrioritizeIPv4)),
		"remove_exempted_containers":       types.Int64Value(int64(p.RemoveExemptedContainers)),
		"restart_win_http_svc":             types.Int64Value(int64(p.RestartWinHttpSvc)),
		"trigger_domain_profile_detection": types.Int64Value(int64(p.TriggerDomainProfleDetection)),
		"uninstall_password":               types.StringValue(p.UninstallPassword),
		"wfp_driver":                       types.Int64Value(int64(p.WfpDriver)),
	})
	return obj
}
