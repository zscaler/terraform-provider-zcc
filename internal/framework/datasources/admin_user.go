package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/admin_users"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &AdminUserDataSource{}
	_ datasource.DataSourceWithConfigure = &AdminUserDataSource{}
)

func NewAdminUserDataSource() datasource.DataSource {
	return &AdminUserDataSource{}
}

type AdminUserDataSource struct {
	client *client.Client
}

type AdminUserDataSourceModel struct {
	ID                           types.Int64  `tfsdk:"id"`
	UserName                     types.String `tfsdk:"user_name"`
	AccountEnabled               types.String `tfsdk:"account_enabled"`
	CompanyID                    types.String `tfsdk:"company_id"`
	EditEnabled                  types.String `tfsdk:"edit_enabled"`
	IsDefaultAdmin               types.String `tfsdk:"is_default_admin"`
	ServiceType                  types.String `tfsdk:"service_type"`
	AdminManagement              types.String `tfsdk:"admin_management"`
	AdministratorGroup           types.String `tfsdk:"administrator_group"`
	AndroidProfile               types.String `tfsdk:"android_profile"`
	IOSProfile                   types.String `tfsdk:"ios_profile"`
	MacProfile                   types.String `tfsdk:"mac_profile"`
	WindowsProfile               types.String `tfsdk:"windows_profile"`
	LinuxProfile                 types.String `tfsdk:"linux_profile"`
	AppBypass                    types.String `tfsdk:"app_bypass"`
	AppProfileGroup              types.String `tfsdk:"app_profile_group"`
	AuditLogs                    types.String `tfsdk:"audit_logs"`
	AuthSetting                  types.String `tfsdk:"auth_setting"`
	ClientConnectorAppStore      types.String `tfsdk:"client_connector_app_store"`
	ClientConnectorIDP           types.String `tfsdk:"client_connector_idp"`
	ClientConnectorNotifications types.String `tfsdk:"client_connector_notifications"`
	ClientConnectorSupport       types.String `tfsdk:"client_connector_support"`
	Dashboard                    types.String `tfsdk:"dashboard"`
	DDILConfiguration            types.String `tfsdk:"ddil_configuration"`
	DedicatedProxyPorts          types.String `tfsdk:"dedicated_proxy_ports"`
	DeviceGroups                 types.String `tfsdk:"device_groups"`
	DeviceOverview               types.String `tfsdk:"device_overview"`
	DevicePosture                types.String `tfsdk:"device_posture"`
	EnrolledDevicesGroup         types.String `tfsdk:"enrolled_devices_group"`
	ForwardingProfile            types.String `tfsdk:"forwarding_profile"`
	IsEditable                   types.Bool   `tfsdk:"is_editable"`
	MachineTunnel                types.String `tfsdk:"machine_tunnel"`
	ObfuscateData                types.String `tfsdk:"obfuscate_data"`
	PartnerDeviceOverview        types.String `tfsdk:"partner_device_overview"`
	PublicAPI                    types.String `tfsdk:"public_api"`
	RoleName                     types.String `tfsdk:"role_name"`
	TrustedNetwork               types.String `tfsdk:"trusted_network"`
	UpdatedBy                    types.String `tfsdk:"updated_by"`
	UserAgent                    types.String `tfsdk:"user_agent"`
	ZPAPartnerLogin              types.String `tfsdk:"zpa_partner_login"`
	ZscalerDeception             types.String `tfsdk:"zscaler_deception"`
	ZscalerEntitlement           types.String `tfsdk:"zscaler_entitlement"`
}

func (d *AdminUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_user"
}

func (d *AdminUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC admin user by ID or user name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The unique identifier of the admin user.",
				Optional:    true,
			},
			"user_name": schema.StringAttribute{
				Description: "The user name of the admin user.",
				Optional:    true,
			},
			"account_enabled":                schema.StringAttribute{Computed: true},
			"company_id":                     schema.StringAttribute{Computed: true},
			"edit_enabled":                   schema.StringAttribute{Computed: true},
			"is_default_admin":               schema.StringAttribute{Computed: true},
			"service_type":                   schema.StringAttribute{Computed: true},
			"admin_management":               schema.StringAttribute{Computed: true},
			"administrator_group":            schema.StringAttribute{Computed: true},
			"android_profile":                schema.StringAttribute{Computed: true},
			"ios_profile":                    schema.StringAttribute{Computed: true},
			"mac_profile":                    schema.StringAttribute{Computed: true},
			"windows_profile":                schema.StringAttribute{Computed: true},
			"linux_profile":                  schema.StringAttribute{Computed: true},
			"app_bypass":                     schema.StringAttribute{Computed: true},
			"app_profile_group":              schema.StringAttribute{Computed: true},
			"audit_logs":                     schema.StringAttribute{Computed: true},
			"auth_setting":                   schema.StringAttribute{Computed: true},
			"client_connector_app_store":     schema.StringAttribute{Computed: true},
			"client_connector_idp":           schema.StringAttribute{Computed: true},
			"client_connector_notifications": schema.StringAttribute{Computed: true},
			"client_connector_support":       schema.StringAttribute{Computed: true},
			"dashboard":                      schema.StringAttribute{Computed: true},
			"ddil_configuration":             schema.StringAttribute{Computed: true},
			"dedicated_proxy_ports":          schema.StringAttribute{Computed: true},
			"device_groups":                  schema.StringAttribute{Computed: true},
			"device_overview":                schema.StringAttribute{Computed: true},
			"device_posture":                 schema.StringAttribute{Computed: true},
			"enrolled_devices_group":         schema.StringAttribute{Computed: true},
			"forwarding_profile":             schema.StringAttribute{Computed: true},
			"is_editable":                    schema.BoolAttribute{Computed: true},
			"machine_tunnel":                 schema.StringAttribute{Computed: true},
			"obfuscate_data":                 schema.StringAttribute{Computed: true},
			"partner_device_overview":        schema.StringAttribute{Computed: true},
			"public_api":                     schema.StringAttribute{Computed: true},
			"role_name":                      schema.StringAttribute{Computed: true},
			"trusted_network":                schema.StringAttribute{Computed: true},
			"updated_by":                     schema.StringAttribute{Computed: true},
			"user_agent":                     schema.StringAttribute{Computed: true},
			"zpa_partner_login":              schema.StringAttribute{Computed: true},
			"zscaler_deception":              schema.StringAttribute{Computed: true},
			"zscaler_entitlement":            schema.StringAttribute{Computed: true},
		},
	}
}

func (d *AdminUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AdminUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AdminUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !data.ID.IsNull() && data.ID.ValueInt64() != 0
	hasUserName := !data.UserName.IsNull() && data.UserName.ValueString() != ""
	if !hasID && !hasUserName {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or user_name must be specified")
		return
	}

	service := d.client.Service

	tflog.Info(ctx, "Fetching admin users")
	users, err := admin_users.GetAdminUsers(ctx, service, "")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read admin users: %v", err))
		return
	}

	var user *admin_users.AdminUser
	if !data.ID.IsNull() && data.ID.ValueInt64() != 0 {
		id := int(data.ID.ValueInt64())
		for i := range users {
			if users[i].ID == id {
				user = &users[i]
				break
			}
		}
		if user == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Admin user with ID %d not found", id))
			return
		}
	} else {
		name := data.UserName.ValueString()
		for i := range users {
			if users[i].UserName == name {
				user = &users[i]
				break
			}
		}
		if user == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Admin user with user_name '%s' not found", name))
			return
		}
	}

	role := user.CompanyRole
	model := AdminUserDataSourceModel{
		ID:                           types.Int64Value(int64(user.ID)),
		UserName:                     types.StringValue(user.UserName),
		AccountEnabled:               types.StringValue(user.AccountEnabled),
		CompanyID:                    types.StringValue(user.CompanyID),
		EditEnabled:                  types.StringValue(user.EditEnabled),
		IsDefaultAdmin:               types.StringValue(user.IsDefaultAdmin),
		ServiceType:                  types.StringValue(user.ServiceType),
		AdminManagement:              types.StringValue(role.AdminManagement),
		AdministratorGroup:           types.StringValue(role.AdministratorGroup),
		AndroidProfile:               types.StringValue(role.AndroidProfile),
		IOSProfile:                   types.StringValue(role.IOSProfile),
		MacProfile:                   types.StringValue(role.MACProfile),
		WindowsProfile:               types.StringValue(role.WindowsProfile),
		LinuxProfile:                 types.StringValue(role.LinuxProfile),
		AppBypass:                    types.StringValue(role.AppBypass),
		AppProfileGroup:              types.StringValue(role.AppProfileGroup),
		AuditLogs:                    types.StringValue(role.AuditLogs),
		AuthSetting:                  types.StringValue(role.AuthSetting),
		ClientConnectorAppStore:      types.StringValue(role.ClientConnectorAppStore),
		ClientConnectorIDP:           types.StringValue(role.ClientConnectorIDP),
		ClientConnectorNotifications: types.StringValue(role.ClientConnectorNotifications),
		ClientConnectorSupport:       types.StringValue(role.ClientConnectorSupport),
		Dashboard:                    types.StringValue(role.Dashboard),
		DDILConfiguration:            types.StringValue(role.DDILConfiguration),
		DedicatedProxyPorts:          types.StringValue(role.DedicatedProxyPorts),
		DeviceGroups:                 types.StringValue(role.DeviceGroups),
		DeviceOverview:               types.StringValue(role.DeviceOverview),
		DevicePosture:                types.StringValue(role.DevicePosture),
		EnrolledDevicesGroup:         types.StringValue(role.EnrolledDevicesGroup),
		ForwardingProfile:            types.StringValue(role.ForwardingProfile),
		IsEditable:                   types.BoolValue(role.IsEditable),
		MachineTunnel:                types.StringValue(role.MachineTunnel),
		ObfuscateData:                types.StringValue(role.ObfuscateData),
		PartnerDeviceOverview:        types.StringValue(role.PartnerDeviceOverview),
		PublicAPI:                    types.StringValue(role.PublicAPI),
		RoleName:                     types.StringValue(role.RoleName),
		TrustedNetwork:               types.StringValue(role.TrustedNetwork),
		UpdatedBy:                    types.StringValue(role.UpdatedBy),
		UserAgent:                    types.StringValue(role.UserAgent),
		ZPAPartnerLogin:              types.StringValue(role.ZPAPartnerLogin),
		ZscalerDeception:             types.StringValue(role.ZscalerDeception),
		ZscalerEntitlement:           types.StringValue(role.ZscalerEntitlement),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
