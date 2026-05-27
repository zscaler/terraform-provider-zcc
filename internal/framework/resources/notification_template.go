package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/notification_template"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/helpers"
)

var (
	_ resource.Resource                = &NotificationTemplateResource{}
	_ resource.ResourceWithConfigure   = &NotificationTemplateResource{}
	_ resource.ResourceWithImportState = &NotificationTemplateResource{}
)

func NewNotificationTemplateResource() resource.Resource {
	return &NotificationTemplateResource{}
}

type NotificationTemplateResource struct {
	client *client.Client
}

// NotificationTemplateResourceModel mirrors the user-configurable subset
// of notification_template.NotificationTemplate.
//
// Pure server-side metadata — createdBy, editedBy — is intentionally
// NOT carried on the resource so it can't show up as HCL-configurable in
// completions or in `terraform plan`. Consumers that need to read those
// fields should use the `zcc_notification_template` data source.
type NotificationTemplateResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	IsDefaultTemplate       types.Bool   `tfsdk:"is_default_template"`
	EnableClient            types.Bool   `tfsdk:"enable_client"`
	EnableZia               types.Bool   `tfsdk:"enable_zia"`
	EnableAppUpdates        types.Bool   `tfsdk:"enable_app_updates"`
	EnableServiceStatus     types.Bool   `tfsdk:"enable_service_status"`
	DurationInSeconds       types.Int64  `tfsdk:"duration_in_seconds"`
	EnablePersistent        types.Bool   `tfsdk:"enable_persistent"`
	EnableDoNotDisturb      types.Bool   `tfsdk:"enable_do_not_disturb"`
	ZIANotificationTemplate types.Object `tfsdk:"zia_notification_template"`
	ZPANotificationTemplate types.Object `tfsdk:"zpa_notification_template"`
}

func (r *NotificationTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_template"
}

func (r *NotificationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC notification template via /zcc/papi/public/v2/notification-templates. " +
			"Controls which end-user notifications the Client Connector raises (app updates, service status, " +
			"ZIA blocks, ZPA reauthentication, etc.). Per-service toggles live under the two nested blocks " +
			"`zia_notification_template` and `zpa_notification_template`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric identifier of the template, carried as a string per Terraform convention. API field: id (JSON number).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Operator-visible name. API field: name.",
			},
			"is_default_template": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Marks the template as the company default. Only one template should be the default at a time. API field: isDefaultTemplate.",
			},
			"enable_client": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable generic Client Connector notifications. API field: enableClient.",
			},
			"enable_zia": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Master switch for ZIA-driven notifications (per-channel toggles live under `zia_notification_template`). API field: enableZia.",
			},
			"enable_app_updates": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Notify the end user when the ZCC app receives an update. API field: enableAppUpdates.",
			},
			"enable_service_status": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Notify on service status changes (tunnel up/down, etc.). API field: enableServiceStatus.",
			},
			"duration_in_seconds": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "How long (in seconds) transient notifications stay on screen. API field: durationInSeconds.",
			},
			"enable_persistent": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Force notifications to stay on screen until dismissed. API field: enablePersistent.",
			},
			"enable_do_not_disturb": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Honour the OS \"do not disturb\" mode. API field: enableDoNotDisturb.",
			},
			"zia_notification_template": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Per-channel ZIA notification toggles. API field: ziaNotificationTemplate.",
				Attributes:  ziaNotificationTemplateAttributes(),
			},
			"zpa_notification_template": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Per-channel ZPA notification toggles. API field: zpaNotificationTemplate.",
				Attributes:  zpaNotificationTemplateAttributes(),
			},
		},
	}
}

func (r *NotificationTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan NotificationTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	payload := expandNotificationTemplate(&plan)

	tflog.Info(ctx, "Creating ZCC notification template", map[string]any{"name": payload.Name})

	created, _, err := notification_template.Create(ctx, service, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create notification template: %v", err))
		return
	}

	tflog.Info(ctx, "Created ZCC notification template", map[string]any{"id": created.ID})
	flattenNotificationTemplate(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state NotificationTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, convErr := strconv.Atoi(state.ID.ValueString())
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Notification template id %q is not a valid integer: %v", state.ID.ValueString(), convErr))
		return
	}

	service := r.client.Service
	tpl, err := notification_template.Get(ctx, service, id)
	if err != nil {
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "Removing notification template from state - no longer exists", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read notification template %d: %v", id, err))
		return
	}

	flattenNotificationTemplate(tpl, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan NotificationTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state NotificationTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, convErr := strconv.Atoi(state.ID.ValueString())
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Notification template id %q is not a valid integer: %v", state.ID.ValueString(), convErr))
		return
	}

	service := r.client.Service
	payload := expandNotificationTemplate(&plan)
	payload.ID = id

	tflog.Info(ctx, "Updating ZCC notification template", map[string]any{"id": id})

	updated, _, err := notification_template.Update(ctx, service, id, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update notification template %d: %v", id, err))
		return
	}

	flattenNotificationTemplate(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete uses the idempotent "Get-then-Delete" pattern. See the longer
// comment on TrustedNetworkResource.Delete for the rationale; both
// branches treat a 404 (`errorx.ErrorResponse.IsObjectNotFound()`) as
// success rather than a hard error so out-of-band UI deletes and
// concurrent sweeper runs do not surface confusing "Record not
// available" diagnostics to Terraform users.
func (r *NotificationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state NotificationTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, convErr := strconv.Atoi(state.ID.ValueString())
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Notification template id %q is not a valid integer: %v", state.ID.ValueString(), convErr))
		return
	}

	service := r.client.Service

	if _, err := notification_template.Get(ctx, service, id); err != nil {
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "Notification template already removed upstream; nothing to delete", map[string]any{"id": id})
			return
		}
		tflog.Warn(ctx, "Pre-delete GET failed; proceeding to DELETE anyway", map[string]any{"id": id, "error": err.Error()})
	}

	tflog.Info(ctx, "Deleting ZCC notification template", map[string]any{"id": id})
	if _, err := notification_template.Delete(ctx, service, id); err != nil {
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "Notification template was removed between GET and DELETE; treating as success", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to delete notification template %d: %v", id, err))
	}
}

// ImportState supports two shapes:
//   - `terraform import zcc_notification_template.this 12345` — numeric
//     id passed straight through; Read fills the rest.
//   - `terraform import zcc_notification_template.this Default-Template`
//     — looked up by case-insensitive Name through the SDK's GetByName.
func (r *NotificationTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	id := req.ID
	if _, parseErr := strconv.ParseInt(id, 10, 64); parseErr == nil {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}

	service := r.client.Service
	tpl, err := notification_template.GetByName(ctx, service, id)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import notification template %q: %v", id, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(strconv.Itoa(tpl.ID)))...)
}

// ---------------------------------------------------------------------------
// expand / flatten — top level
// ---------------------------------------------------------------------------

func expandNotificationTemplate(plan *NotificationTemplateResourceModel) notification_template.NotificationTemplate {
	return notification_template.NotificationTemplate{
		Name:                    plan.Name.ValueString(),
		IsDefaultTemplate:       plan.IsDefaultTemplate.ValueBool(),
		EnableClient:            plan.EnableClient.ValueBool(),
		EnableZia:               plan.EnableZia.ValueBool(),
		EnableAppUpdates:        plan.EnableAppUpdates.ValueBool(),
		EnableServiceStatus:     plan.EnableServiceStatus.ValueBool(),
		DurationInSeconds:       int(plan.DurationInSeconds.ValueInt64()),
		EnablePersistent:        plan.EnablePersistent.ValueBool(),
		EnableDoNotDisturb:      plan.EnableDoNotDisturb.ValueBool(),
		ZIANotificationTemplate: expandZIANotificationTemplate(plan.ZIANotificationTemplate),
		ZPANotificationTemplate: expandZPANotificationTemplate(plan.ZPANotificationTemplate),
	}
}

func flattenNotificationTemplate(tpl *notification_template.NotificationTemplate, model *NotificationTemplateResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(tpl.ID))
	model.Name = types.StringValue(tpl.Name)
	model.IsDefaultTemplate = types.BoolValue(tpl.IsDefaultTemplate)
	model.EnableClient = types.BoolValue(tpl.EnableClient)
	model.EnableZia = types.BoolValue(tpl.EnableZia)
	model.EnableAppUpdates = types.BoolValue(tpl.EnableAppUpdates)
	model.EnableServiceStatus = types.BoolValue(tpl.EnableServiceStatus)
	model.DurationInSeconds = types.Int64Value(int64(tpl.DurationInSeconds))
	model.EnablePersistent = types.BoolValue(tpl.EnablePersistent)
	model.EnableDoNotDisturb = types.BoolValue(tpl.EnableDoNotDisturb)
	model.ZIANotificationTemplate = flattenZIANotificationTemplate(tpl.ZIANotificationTemplate)
	model.ZPANotificationTemplate = flattenZPANotificationTemplate(tpl.ZPANotificationTemplate)
}

// =============================================================================
// zia_notification_template block
// =============================================================================

// ziaNotificationTemplateAttrTypes returns the attr.Type map used both
// for ObjectValue/ObjectNull construction and for the data source
// nested-attribute schema.
func ziaNotificationTemplateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enable_zia_firewall":       types.BoolType,
		"enable_zia_firewall_popup": types.BoolType,
		"enable_zia_dns":            types.BoolType,
		"enable_zia_dns_popup":      types.BoolType,
		"enable_zia_ips":            types.BoolType,
		"enable_zia_ips_popup":      types.BoolType,
		"enable_zia_persistent":     types.BoolType,
	}
}

// ziaNotificationTemplateAttributes returns the schema.Attribute map for
// the nested zia_notification_template block on the resource. The data
// source reuses this through a Computed-only variant below.
func ziaNotificationTemplateAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enable_zia_firewall":       schema.BoolAttribute{Optional: true, Computed: true, Description: "Notify on ZIA firewall block events. API field: enableZiaFirewall."},
		"enable_zia_firewall_popup": schema.BoolAttribute{Optional: true, Computed: true, Description: "Raise a popup (not just a tray notification) for ZIA firewall blocks. API field: enableZiaFirewallPopup."},
		"enable_zia_dns":            schema.BoolAttribute{Optional: true, Computed: true, Description: "Notify on ZIA DNS block events. API field: enableZiaDNS."},
		"enable_zia_dns_popup":      schema.BoolAttribute{Optional: true, Computed: true, Description: "Raise a popup for ZIA DNS blocks. API field: enableZiaDNSPopup."},
		"enable_zia_ips":            schema.BoolAttribute{Optional: true, Computed: true, Description: "Notify on ZIA IPS block events. API field: enableZiaIPS."},
		"enable_zia_ips_popup":      schema.BoolAttribute{Optional: true, Computed: true, Description: "Raise a popup for ZIA IPS blocks. API field: enableZiaIPSPopup."},
		"enable_zia_persistent":     schema.BoolAttribute{Optional: true, Computed: true, Description: "Keep ZIA notifications persistent until dismissed. API field: enableZiaPersistent."},
	}
}

func expandZIANotificationTemplate(obj types.Object) notification_template.ZIANotificationTemplate {
	if obj.IsNull() || obj.IsUnknown() {
		return notification_template.ZIANotificationTemplate{}
	}
	a := obj.Attributes()
	return notification_template.ZIANotificationTemplate{
		EnableZiaFirewall:      helpers.BoolFromAttr(a["enable_zia_firewall"]),
		EnableZiaFirewallPopup: helpers.BoolFromAttr(a["enable_zia_firewall_popup"]),
		EnableZiaDNS:           helpers.BoolFromAttr(a["enable_zia_dns"]),
		EnableZiaDNSPopup:      helpers.BoolFromAttr(a["enable_zia_dns_popup"]),
		EnableZiaIPS:           helpers.BoolFromAttr(a["enable_zia_ips"]),
		EnableZiaIPSPopup:      helpers.BoolFromAttr(a["enable_zia_ips_popup"]),
		EnableZiaPersistent:    helpers.BoolFromAttr(a["enable_zia_persistent"]),
	}
}

func flattenZIANotificationTemplate(t notification_template.ZIANotificationTemplate) types.Object {
	obj, _ := types.ObjectValue(ziaNotificationTemplateAttrTypes(), map[string]attr.Value{
		"enable_zia_firewall":       types.BoolValue(t.EnableZiaFirewall),
		"enable_zia_firewall_popup": types.BoolValue(t.EnableZiaFirewallPopup),
		"enable_zia_dns":            types.BoolValue(t.EnableZiaDNS),
		"enable_zia_dns_popup":      types.BoolValue(t.EnableZiaDNSPopup),
		"enable_zia_ips":            types.BoolValue(t.EnableZiaIPS),
		"enable_zia_ips_popup":      types.BoolValue(t.EnableZiaIPSPopup),
		"enable_zia_persistent":     types.BoolValue(t.EnableZiaPersistent),
	})
	return obj
}

// =============================================================================
// zpa_notification_template block
// =============================================================================

func zpaNotificationTemplateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enable_device_posture_failure":  types.BoolType,
		"enable_zpa_reauth":              types.BoolType,
		"zpa_reauth_interval_in_minutes": types.Int64Type,
		"delay_posture_failure_seconds":  types.Int64Type,
	}
}

func zpaNotificationTemplateAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enable_device_posture_failure":  schema.BoolAttribute{Optional: true, Computed: true, Description: "Notify when device posture evaluation fails. API field: enableDevicePostureFailure."},
		"enable_zpa_reauth":              schema.BoolAttribute{Optional: true, Computed: true, Description: "Notify the user when ZPA reauthentication is required. API field: enableZpaReauth."},
		"zpa_reauth_interval_in_minutes": schema.Int64Attribute{Optional: true, Computed: true, Description: "How often (in minutes) to remind the user to reauthenticate. API field: zpaReauthIntervalInMinutes."},
		"delay_posture_failure_seconds":  schema.Int64Attribute{Optional: true, Computed: true, Description: "Delay (in seconds) before reporting a posture failure to the user. API field: delayPostureFailureSeconds."},
	}
}

func expandZPANotificationTemplate(obj types.Object) notification_template.ZPANotificationTemplate {
	if obj.IsNull() || obj.IsUnknown() {
		return notification_template.ZPANotificationTemplate{}
	}
	a := obj.Attributes()
	return notification_template.ZPANotificationTemplate{
		EnableDevicePostureFailure: helpers.BoolFromAttr(a["enable_device_posture_failure"]),
		EnableZpaReauth:            helpers.BoolFromAttr(a["enable_zpa_reauth"]),
		ZpaReauthIntervalInMinutes: helpers.IntFromAttr(a["zpa_reauth_interval_in_minutes"]),
		DelayPostureFailureSeconds: helpers.IntFromAttr(a["delay_posture_failure_seconds"]),
	}
}

func flattenZPANotificationTemplate(t notification_template.ZPANotificationTemplate) types.Object {
	obj, _ := types.ObjectValue(zpaNotificationTemplateAttrTypes(), map[string]attr.Value{
		"enable_device_posture_failure":  types.BoolValue(t.EnableDevicePostureFailure),
		"enable_zpa_reauth":              types.BoolValue(t.EnableZpaReauth),
		"zpa_reauth_interval_in_minutes": types.Int64Value(int64(t.ZpaReauthIntervalInMinutes)),
		"delay_posture_failure_seconds":  types.Int64Value(int64(t.DelayPostureFailureSeconds)),
	})
	return obj
}
