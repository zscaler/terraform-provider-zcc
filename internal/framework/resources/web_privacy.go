package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/web_privacy"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ resource.Resource                = &WebPrivacyResource{}
	_ resource.ResourceWithConfigure   = &WebPrivacyResource{}
	_ resource.ResourceWithImportState = &WebPrivacyResource{}
)

func NewWebPrivacyResource() resource.Resource {
	return &WebPrivacyResource{}
}

type WebPrivacyResource struct {
	client *client.Client
}

type WebPrivacyResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Active                        types.Bool   `tfsdk:"active"`
	CollectUserInfo               types.Bool   `tfsdk:"collect_user_info"`
	CollectMachineHostname        types.Bool   `tfsdk:"collect_machine_hostname"`
	CollectZdxLocation            types.Bool   `tfsdk:"collect_zdx_location"`
	EnablePacketCapture           types.Bool   `tfsdk:"enable_packet_capture"`
	DisableCrashlytics            types.Bool   `tfsdk:"disable_crashlytics"`
	OverrideT2ProtocolSetting     types.Bool   `tfsdk:"override_t2_protocol_setting"`
	RestrictRemotePacketCapture   types.Bool   `tfsdk:"restrict_remote_packet_capture"`
	GrantAccessToZscalerLogFolder types.Bool   `tfsdk:"grant_access_to_zscaler_log_folder"`
	ExportLogsForNonAdmin         types.Bool   `tfsdk:"export_logs_for_non_admin"`
	EnableAutoLogSnippet          types.Bool   `tfsdk:"enable_auto_log_snippet"`
	EnforceSecurePacUrls          types.Bool   `tfsdk:"enforce_secure_pac_urls"`
	EnableFQDNMatchForVpnBypasses types.Bool   `tfsdk:"enable_fqdn_match_for_vpn_bypasses"`
}

func (r *WebPrivacyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_privacy"
}

func webPrivacyBoolAttr(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: desc,
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	}
}

func (r *WebPrivacyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ZCC web privacy settings (singleton). The API only supports GET and PUT; create runs an initial PUT, update runs PUT, and delete only removes the resource from state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Settings record identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active":                             webPrivacyBoolAttr("Whether web privacy settings are active."),
			"collect_user_info":                  webPrivacyBoolAttr("Whether to collect user information."),
			"collect_machine_hostname":           webPrivacyBoolAttr("Whether to collect machine hostname."),
			"collect_zdx_location":               webPrivacyBoolAttr("Whether to collect ZDX location."),
			"enable_packet_capture":              webPrivacyBoolAttr("Whether packet capture is enabled."),
			"disable_crashlytics":                webPrivacyBoolAttr("Whether Crashlytics is disabled."),
			"override_t2_protocol_setting":       webPrivacyBoolAttr("Whether to override T2 protocol setting."),
			"restrict_remote_packet_capture":     webPrivacyBoolAttr("Whether remote packet capture is restricted."),
			"grant_access_to_zscaler_log_folder": webPrivacyBoolAttr("Whether to grant access to Zscaler log folder."),
			"export_logs_for_non_admin":          webPrivacyBoolAttr("Whether non-admin users can export logs."),
			"enable_auto_log_snippet":            webPrivacyBoolAttr("Whether automatic log snippet collection is enabled."),
			"enforce_secure_pac_urls":            webPrivacyBoolAttr("Whether secure PAC URLs are enforced."),
			"enable_fqdn_match_for_vpn_bypasses": webPrivacyBoolAttr("Whether FQDN matching for VPN bypasses is enabled."),
		},
	}
}

func (r *WebPrivacyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebPrivacyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan WebPrivacyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existing, err := web_privacy.GetWebPrivacyInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read web privacy settings: %v", err))
		return
	}

	payload := expandWebPrivacy(&plan, existing)
	tflog.Info(ctx, "Updating ZCC web privacy (singleton create)", map[string]any{"id": payload.ID})

	updated, err := web_privacy.UpdateWebPrivacyInfo(ctx, r.client.Service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update web privacy: %v", err))
		return
	}

	flattenWebPrivacyResource(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebPrivacyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state WebPrivacyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := web_privacy.GetWebPrivacyInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read web privacy settings: %v", err))
		return
	}

	flattenWebPrivacyResource(info, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebPrivacyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan WebPrivacyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existing, err := web_privacy.GetWebPrivacyInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read web privacy settings: %v", err))
		return
	}

	payload := expandWebPrivacy(&plan, existing)
	tflog.Info(ctx, "Updating ZCC web privacy", map[string]any{"id": payload.ID})

	updated, err := web_privacy.UpdateWebPrivacyInfo(ctx, r.client.Service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update web privacy: %v", err))
		return
	}

	flattenWebPrivacyResource(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebPrivacyResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *WebPrivacyResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	info, err := web_privacy.GetWebPrivacyInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read web privacy settings: %v", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(info.ID))...)
}

func boolToStr01(b types.Bool) string {
	if b.ValueBool() {
		return "1"
	}
	return "0"
}

func expandWebPrivacy(plan *WebPrivacyResourceModel, existing *web_privacy.WebPrivacyInfo) *web_privacy.WebPrivacyInfo {
	p := *existing

	if !plan.Active.IsNull() && !plan.Active.IsUnknown() {
		p.Active = boolToStr01(plan.Active)
	}
	if !plan.CollectUserInfo.IsNull() && !plan.CollectUserInfo.IsUnknown() {
		p.CollectUserInfo = boolToStr01(plan.CollectUserInfo)
	}
	if !plan.CollectMachineHostname.IsNull() && !plan.CollectMachineHostname.IsUnknown() {
		p.CollectMachineHostname = boolToStr01(plan.CollectMachineHostname)
	}
	if !plan.CollectZdxLocation.IsNull() && !plan.CollectZdxLocation.IsUnknown() {
		p.CollectZdxLocation = boolToStr01(plan.CollectZdxLocation)
	}
	if !plan.EnablePacketCapture.IsNull() && !plan.EnablePacketCapture.IsUnknown() {
		p.EnablePacketCapture = boolToStr01(plan.EnablePacketCapture)
	}
	if !plan.DisableCrashlytics.IsNull() && !plan.DisableCrashlytics.IsUnknown() {
		p.DisableCrashlytics = boolToStr01(plan.DisableCrashlytics)
	}
	if !plan.OverrideT2ProtocolSetting.IsNull() && !plan.OverrideT2ProtocolSetting.IsUnknown() {
		p.OverrideT2ProtocolSetting = boolToStr01(plan.OverrideT2ProtocolSetting)
	}
	if !plan.RestrictRemotePacketCapture.IsNull() && !plan.RestrictRemotePacketCapture.IsUnknown() {
		p.RestrictRemotePacketCapture = boolToStr01(plan.RestrictRemotePacketCapture)
	}
	if !plan.GrantAccessToZscalerLogFolder.IsNull() && !plan.GrantAccessToZscalerLogFolder.IsUnknown() {
		p.GrantAccessToZscalerLogFolder = boolToStr01(plan.GrantAccessToZscalerLogFolder)
	}
	if !plan.ExportLogsForNonAdmin.IsNull() && !plan.ExportLogsForNonAdmin.IsUnknown() {
		p.ExportLogsForNonAdmin = boolToStr01(plan.ExportLogsForNonAdmin)
	}
	if !plan.EnableAutoLogSnippet.IsNull() && !plan.EnableAutoLogSnippet.IsUnknown() {
		p.EnableAutoLogSnippet = boolToStr01(plan.EnableAutoLogSnippet)
	}
	if !plan.EnforceSecurePacUrls.IsNull() && !plan.EnforceSecurePacUrls.IsUnknown() {
		p.EnforceSecurePacUrls = boolToStr01(plan.EnforceSecurePacUrls)
	}
	if !plan.EnableFQDNMatchForVpnBypasses.IsNull() && !plan.EnableFQDNMatchForVpnBypasses.IsUnknown() {
		p.EnableFQDNMatchForVpnBypasses = boolToStr01(plan.EnableFQDNMatchForVpnBypasses)
	}

	return &p
}

func flattenWebPrivacyResource(p *web_privacy.WebPrivacyInfo, m *WebPrivacyResourceModel) {
	m.ID = types.StringValue(p.ID)
	m.Active = types.BoolValue(p.Active == "1")
	m.CollectUserInfo = types.BoolValue(p.CollectUserInfo == "1")
	m.CollectMachineHostname = types.BoolValue(p.CollectMachineHostname == "1")
	m.CollectZdxLocation = types.BoolValue(p.CollectZdxLocation == "1")
	m.EnablePacketCapture = types.BoolValue(p.EnablePacketCapture == "1")
	m.DisableCrashlytics = types.BoolValue(p.DisableCrashlytics == "1")
	m.OverrideT2ProtocolSetting = types.BoolValue(p.OverrideT2ProtocolSetting == "1")
	m.RestrictRemotePacketCapture = types.BoolValue(p.RestrictRemotePacketCapture == "1")
	m.GrantAccessToZscalerLogFolder = types.BoolValue(p.GrantAccessToZscalerLogFolder == "1")
	m.ExportLogsForNonAdmin = types.BoolValue(p.ExportLogsForNonAdmin == "1")
	m.EnableAutoLogSnippet = types.BoolValue(p.EnableAutoLogSnippet == "1")

	// API GET omits these fields; preserve existing plan/state value to avoid drift.
	if m.EnforceSecurePacUrls.IsNull() || m.EnforceSecurePacUrls.IsUnknown() {
		m.EnforceSecurePacUrls = types.BoolValue(p.EnforceSecurePacUrls == "1")
	}
	if m.EnableFQDNMatchForVpnBypasses.IsNull() || m.EnableFQDNMatchForVpnBypasses.IsUnknown() {
		m.EnableFQDNMatchForVpnBypasses = types.BoolValue(p.EnableFQDNMatchForVpnBypasses == "1")
	}
}
