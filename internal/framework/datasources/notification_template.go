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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/notification_template"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &NotificationTemplateDataSource{}
	_ datasource.DataSourceWithConfigure = &NotificationTemplateDataSource{}
)

func NewNotificationTemplateDataSource() datasource.DataSource {
	return &NotificationTemplateDataSource{}
}

type NotificationTemplateDataSource struct {
	client *client.Client
}

// NotificationTemplateDataSourceModel mirrors
// notification_template.NotificationTemplate in full — unlike the
// resource, the data source surfaces the server-side metadata
// (created_by, edited_by) so callers can introspect templates they did
// not create.
type NotificationTemplateDataSourceModel struct {
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
	CreatedBy               types.Int64  `tfsdk:"created_by"`
	EditedBy                types.Int64  `tfsdk:"edited_by"`
	ZIANotificationTemplate types.Object `tfsdk:"zia_notification_template"`
	ZPANotificationTemplate types.Object `tfsdk:"zpa_notification_template"`
}

func (d *NotificationTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_template"
}

func (d *NotificationTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC notification template from /zcc/papi/public/v2/notification-templates, by numeric id or by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the notification template (carried as a string). Either id or name must be set.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Operator-visible name. Either id or name must be set.",
				Optional:    true,
				Computed:    true,
			},
			"is_default_template":   schema.BoolAttribute{Description: "Whether this template is the company default.", Computed: true},
			"enable_client":         schema.BoolAttribute{Description: "Whether generic Client Connector notifications are enabled.", Computed: true},
			"enable_zia":            schema.BoolAttribute{Description: "Master switch for ZIA-driven notifications.", Computed: true},
			"enable_app_updates":    schema.BoolAttribute{Description: "Whether app-update notifications are enabled.", Computed: true},
			"enable_service_status": schema.BoolAttribute{Description: "Whether service-status notifications are enabled.", Computed: true},
			"duration_in_seconds":   schema.Int64Attribute{Description: "Transient notification display duration in seconds.", Computed: true},
			"enable_persistent":     schema.BoolAttribute{Description: "Whether notifications stay until dismissed.", Computed: true},
			"enable_do_not_disturb": schema.BoolAttribute{Description: "Whether the template honours the OS Do-Not-Disturb mode.", Computed: true},
			"created_by":            schema.Int64Attribute{Description: "Numeric user id of the operator who created the template.", Computed: true},
			"edited_by":             schema.Int64Attribute{Description: "Numeric user id of the operator who last edited the template.", Computed: true},
			"zia_notification_template": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Per-channel ZIA notification toggles.",
				Attributes:  ziaNotificationTemplateDataSourceAttributes(),
			},
			"zpa_notification_template": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Per-channel ZPA notification toggles.",
				Attributes:  zpaNotificationTemplateDataSourceAttributes(),
			},
		},
	}
}

func (d *NotificationTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationTemplateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""
	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	var (
		tpl *notification_template.NotificationTemplate
		err error
	)

	if hasID {
		idStr := data.ID.ValueString()
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Notification template id %q is not a valid integer: %v", idStr, convErr))
			return
		}
		tflog.Info(ctx, "Fetching notification template", map[string]any{"id": id})
		tpl, err = notification_template.Get(ctx, service, id)
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching notification template", map[string]any{"name": name})
		tpl, err = notification_template.GetByName(ctx, service, name)
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification template: %v", err))
		return
	}

	model := NotificationTemplateDataSourceModel{
		ID:                      types.StringValue(strconv.Itoa(tpl.ID)),
		Name:                    types.StringValue(tpl.Name),
		IsDefaultTemplate:       types.BoolValue(tpl.IsDefaultTemplate),
		EnableClient:            types.BoolValue(tpl.EnableClient),
		EnableZia:               types.BoolValue(tpl.EnableZia),
		EnableAppUpdates:        types.BoolValue(tpl.EnableAppUpdates),
		EnableServiceStatus:     types.BoolValue(tpl.EnableServiceStatus),
		DurationInSeconds:       types.Int64Value(int64(tpl.DurationInSeconds)),
		EnablePersistent:        types.BoolValue(tpl.EnablePersistent),
		EnableDoNotDisturb:      types.BoolValue(tpl.EnableDoNotDisturb),
		CreatedBy:               types.Int64Value(int64(tpl.CreatedBy)),
		EditedBy:                types.Int64Value(int64(tpl.EditedBy)),
		ZIANotificationTemplate: flattenZIANotificationTemplateDataSource(tpl.ZIANotificationTemplate),
		ZPANotificationTemplate: flattenZPANotificationTemplateDataSource(tpl.ZPANotificationTemplate),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// =============================================================================
// nested-block helpers (data source / Computed-only)
// =============================================================================

// ziaNotificationTemplateDataSourceAttrTypes mirrors the resource's
// nested attr.Type map; kept in this file so the data source does not
// import the resources package.
func ziaNotificationTemplateDataSourceAttrTypes() map[string]attr.Type {
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

func ziaNotificationTemplateDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enable_zia_firewall":       schema.BoolAttribute{Description: "Whether ZIA firewall block notifications are enabled.", Computed: true},
		"enable_zia_firewall_popup": schema.BoolAttribute{Description: "Whether ZIA firewall blocks raise a popup.", Computed: true},
		"enable_zia_dns":            schema.BoolAttribute{Description: "Whether ZIA DNS block notifications are enabled.", Computed: true},
		"enable_zia_dns_popup":      schema.BoolAttribute{Description: "Whether ZIA DNS blocks raise a popup.", Computed: true},
		"enable_zia_ips":            schema.BoolAttribute{Description: "Whether ZIA IPS block notifications are enabled.", Computed: true},
		"enable_zia_ips_popup":      schema.BoolAttribute{Description: "Whether ZIA IPS blocks raise a popup.", Computed: true},
		"enable_zia_persistent":     schema.BoolAttribute{Description: "Whether ZIA notifications stay until dismissed.", Computed: true},
	}
}

func flattenZIANotificationTemplateDataSource(t notification_template.ZIANotificationTemplate) types.Object {
	obj, _ := types.ObjectValue(ziaNotificationTemplateDataSourceAttrTypes(), map[string]attr.Value{
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

func zpaNotificationTemplateDataSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enable_device_posture_failure":  types.BoolType,
		"enable_zpa_reauth":              types.BoolType,
		"zpa_reauth_interval_in_minutes": types.Int64Type,
		"delay_posture_failure_seconds":  types.Int64Type,
	}
}

func zpaNotificationTemplateDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enable_device_posture_failure":  schema.BoolAttribute{Description: "Whether posture-failure notifications are enabled.", Computed: true},
		"enable_zpa_reauth":              schema.BoolAttribute{Description: "Whether ZPA reauthentication prompts are enabled.", Computed: true},
		"zpa_reauth_interval_in_minutes": schema.Int64Attribute{Description: "Reauthentication reminder interval in minutes.", Computed: true},
		"delay_posture_failure_seconds":  schema.Int64Attribute{Description: "Delay (in seconds) before reporting a posture failure.", Computed: true},
	}
}

func flattenZPANotificationTemplateDataSource(t notification_template.ZPANotificationTemplate) types.Object {
	obj, _ := types.ObjectValue(zpaNotificationTemplateDataSourceAttrTypes(), map[string]attr.Value{
		"enable_device_posture_failure":  types.BoolValue(t.EnableDevicePostureFailure),
		"enable_zpa_reauth":              types.BoolValue(t.EnableZpaReauth),
		"zpa_reauth_interval_in_minutes": types.Int64Value(int64(t.ZpaReauthIntervalInMinutes)),
		"delay_posture_failure_seconds":  types.Int64Value(int64(t.DelayPostureFailureSeconds)),
	})
	return obj
}
