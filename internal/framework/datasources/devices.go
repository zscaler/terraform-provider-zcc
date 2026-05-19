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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/devices"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &DevicesDataSource{}
	_ datasource.DataSourceWithConfigure = &DevicesDataSource{}
)

func NewDevicesDataSource() datasource.DataSource {
	return &DevicesDataSource{}
}

type DevicesDataSource struct {
	client *client.Client
}

type DevicesDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	OsType   types.String `tfsdk:"os_type"`
	Udid     types.String `tfsdk:"udid"`
	Devices  types.List   `tfsdk:"devices"`
}

type DeviceModel struct {
	AgentVersion            types.String `tfsdk:"agent_version"`
	CompanyName             types.String `tfsdk:"company_name"`
	ConfigDownloadTime      types.String `tfsdk:"config_download_time"`
	DeregistrationTimestamp types.String `tfsdk:"deregistration_timestamp"`
	Detail                  types.String `tfsdk:"detail"`
	DownloadCount           types.Int64  `tfsdk:"download_count"`
	HardwareFingerprint     types.String `tfsdk:"hardware_fingerprint"`
	KeepAliveTime           types.String `tfsdk:"keep_alive_time"`
	LastSeenTime            types.String `tfsdk:"last_seen_time"`
	MacAddress              types.String `tfsdk:"mac_address"`
	MachineHostname         types.String `tfsdk:"machine_hostname"`
	Manufacturer            types.String `tfsdk:"manufacturer"`
	OsVersion               types.String `tfsdk:"os_version"`
	Owner                   types.String `tfsdk:"owner"`
	PolicyName              types.String `tfsdk:"policy_name"`
	RegistrationState       types.String `tfsdk:"registration_state"`
	RegistrationTime        types.String `tfsdk:"registration_time"`
	State                   types.String `tfsdk:"state"`
	TunnelVersion           types.String `tfsdk:"tunnel_version"`
	Type                    types.String `tfsdk:"type"`
	Udid                    types.String `tfsdk:"udid"`
	UpmVersion              types.String `tfsdk:"upm_version"`
	User                    types.String `tfsdk:"user"`
	VpnState                types.String `tfsdk:"vpn_state"`
	ZappArch                types.String `tfsdk:"zapp_arch"`
}

func (d *DevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *DevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves ZCC enrolled devices. Can filter by username and/or os_type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier for the datasource.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "Filter devices by username.",
				Optional:    true,
			},
			"os_type": schema.StringAttribute{
				Description: "Filter devices by OS type (e.g., windows, mac, linux, ios, android).",
				Optional:    true,
			},
			"udid": schema.StringAttribute{
				Description: "Filter for a specific device by UDID. When set, returns device details.",
				Optional:    true,
			},
			"devices": schema.ListNestedAttribute{
				Description: "List of enrolled devices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"agent_version":            schema.StringAttribute{Computed: true},
						"company_name":             schema.StringAttribute{Computed: true},
						"config_download_time":     schema.StringAttribute{Computed: true},
						"deregistration_timestamp": schema.StringAttribute{Computed: true},
						"detail":                   schema.StringAttribute{Computed: true},
						"download_count":           schema.Int64Attribute{Computed: true},
						"hardware_fingerprint":     schema.StringAttribute{Computed: true},
						"keep_alive_time":          schema.StringAttribute{Computed: true},
						"last_seen_time":           schema.StringAttribute{Computed: true},
						"mac_address":              schema.StringAttribute{Computed: true},
						"machine_hostname":         schema.StringAttribute{Computed: true},
						"manufacturer":             schema.StringAttribute{Computed: true},
						"os_version":               schema.StringAttribute{Computed: true},
						"owner":                    schema.StringAttribute{Computed: true},
						"policy_name":              schema.StringAttribute{Computed: true},
						"registration_state":       schema.StringAttribute{Computed: true},
						"registration_time":        schema.StringAttribute{Computed: true},
						"state":                    schema.StringAttribute{Computed: true},
						"tunnel_version":           schema.StringAttribute{Computed: true},
						"type":                     schema.StringAttribute{Computed: true},
						"udid":                     schema.StringAttribute{Computed: true},
						"upm_version":              schema.StringAttribute{Computed: true},
						"user":                     schema.StringAttribute{Computed: true},
						"vpn_state":                schema.StringAttribute{Computed: true},
						"zapp_arch":                schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *DevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DevicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := d.client.Service
	username := data.Username.ValueString()
	osType := data.OsType.ValueString()
	udid := data.Udid.ValueString()

	var deviceList []DeviceModel

	if udid != "" {
		tflog.Info(ctx, "Fetching device details by udid", map[string]any{"udid": udid})
		details, err := devices.GetDeviceDetails(ctx, service, username, udid)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device details: %v", err))
			return
		}
		for _, dd := range details {
			deviceList = append(deviceList, deviceDetailsToModel(dd))
		}
	} else {
		tflog.Info(ctx, "Fetching devices", map[string]any{"username": username, "os_type": osType})
		devs, err := devices.GetAll(ctx, service, username, osType)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read devices: %v", err))
			return
		}
		for _, dev := range devs {
			deviceList = append(deviceList, getDevicesToModel(dev))
		}
	}

	deviceObjType := map[string]attr.Type{
		"agent_version":            types.StringType,
		"company_name":             types.StringType,
		"config_download_time":     types.StringType,
		"deregistration_timestamp": types.StringType,
		"detail":                   types.StringType,
		"download_count":           types.Int64Type,
		"hardware_fingerprint":     types.StringType,
		"keep_alive_time":          types.StringType,
		"last_seen_time":           types.StringType,
		"mac_address":              types.StringType,
		"machine_hostname":         types.StringType,
		"manufacturer":             types.StringType,
		"os_version":               types.StringType,
		"owner":                    types.StringType,
		"policy_name":              types.StringType,
		"registration_state":       types.StringType,
		"registration_time":        types.StringType,
		"state":                    types.StringType,
		"tunnel_version":           types.StringType,
		"type":                     types.StringType,
		"udid":                     types.StringType,
		"upm_version":              types.StringType,
		"user":                     types.StringType,
		"vpn_state":                types.StringType,
		"zapp_arch":                types.StringType,
	}

	elements := make([]attr.Value, 0, len(deviceList))
	for _, dm := range deviceList {
		obj, diags := types.ObjectValue(deviceObjType, map[string]attr.Value{
			"agent_version":            dm.AgentVersion,
			"company_name":             dm.CompanyName,
			"config_download_time":     dm.ConfigDownloadTime,
			"deregistration_timestamp": dm.DeregistrationTimestamp,
			"detail":                   dm.Detail,
			"download_count":           dm.DownloadCount,
			"hardware_fingerprint":     dm.HardwareFingerprint,
			"keep_alive_time":          dm.KeepAliveTime,
			"last_seen_time":           dm.LastSeenTime,
			"mac_address":              dm.MacAddress,
			"machine_hostname":         dm.MachineHostname,
			"manufacturer":             dm.Manufacturer,
			"os_version":               dm.OsVersion,
			"owner":                    dm.Owner,
			"policy_name":              dm.PolicyName,
			"registration_state":       dm.RegistrationState,
			"registration_time":        dm.RegistrationTime,
			"state":                    dm.State,
			"tunnel_version":           dm.TunnelVersion,
			"type":                     dm.Type,
			"udid":                     dm.Udid,
			"upm_version":              dm.UpmVersion,
			"user":                     dm.User,
			"vpn_state":                dm.VpnState,
			"zapp_arch":                dm.ZappArch,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elements = append(elements, obj)
	}

	listVal, err := types.ListValue(types.ObjectType{AttrTypes: deviceObjType}, elements)
	if err != nil {
		resp.Diagnostics.AddError("Serialization Error", fmt.Sprintf("Failed to build devices list: %v", err))
		return
	}

	data.ID = types.StringValue("zcc_devices")
	if username != "" {
		data.Username = types.StringValue(username)
	}
	if osType != "" {
		data.OsType = types.StringValue(osType)
	}
	if udid != "" {
		data.Udid = types.StringValue(udid)
	}
	data.Devices = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getDevicesToModel(dev devices.GetDevices) DeviceModel {
	return DeviceModel{
		AgentVersion:            types.StringValue(dev.AgentVersion),
		CompanyName:             types.StringValue(dev.CompanyName),
		ConfigDownloadTime:      types.StringValue(dev.ConfigDownloadTime),
		DeregistrationTimestamp: types.StringValue(dev.DeregistrationTimestamp),
		Detail:                  types.StringValue(dev.Detail),
		DownloadCount:           types.Int64Value(int64(dev.DownloadCount)),
		HardwareFingerprint:     types.StringValue(dev.HardwareFingerprint),
		KeepAliveTime:           types.StringValue(dev.KeepAliveTime),
		LastSeenTime:            types.StringValue(dev.LastSeenTime),
		MacAddress:              types.StringValue(dev.MacAddress),
		MachineHostname:         types.StringValue(dev.MachineHostname),
		Manufacturer:            types.StringValue(dev.Manufacturer),
		OsVersion:               types.StringValue(dev.OsVersion),
		Owner:                   types.StringValue(dev.Owner),
		PolicyName:              types.StringValue(dev.PolicyName),
		RegistrationState:       types.StringValue(dev.RegistrationState),
		RegistrationTime:        types.StringValue(dev.RegistrationTime),
		State:                   types.StringValue(strconv.Itoa(dev.State)),
		TunnelVersion:           types.StringValue(ptrToString(dev.TunnelVersion)),
		Type:                    types.StringValue(strconv.Itoa(dev.Type)),
		Udid:                    types.StringValue(dev.Udid),
		UpmVersion:              types.StringValue(dev.UpmVersion),
		User:                    types.StringValue(dev.User),
		VpnState:                types.StringValue(strconv.Itoa(dev.VpnState)),
		ZappArch:                types.StringValue(ptrToString(dev.ZappArch)),
	}
}

func deviceDetailsToModel(dd devices.DeviceDetails) DeviceModel {
	return DeviceModel{
		AgentVersion:            types.StringValue(dd.AgentVersion),
		CompanyName:             types.StringValue(""),
		ConfigDownloadTime:      types.StringValue(dd.ConfigDownloadTime),
		DeregistrationTimestamp: types.StringValue(dd.DeregistrationTime),
		Detail:                  types.StringValue(""),
		DownloadCount:           types.Int64Value(int64(dd.DownloadCount)),
		HardwareFingerprint:     types.StringValue(dd.HardwareFingerprint),
		KeepAliveTime:           types.StringValue(dd.KeepAliveTime),
		LastSeenTime:            types.StringValue(dd.LastSeenTime),
		MacAddress:              types.StringValue(dd.MacAddress),
		MachineHostname:         types.StringValue(dd.MachineHostname),
		Manufacturer:            types.StringValue(dd.Manufacturer),
		OsVersion:               types.StringValue(dd.OSVersion),
		Owner:                   types.StringValue(dd.Owner),
		PolicyName:              types.StringValue(dd.DevicePolicyName),
		RegistrationState:       types.StringValue(dd.State),
		RegistrationTime:        types.StringValue(dd.RegistrationTime),
		State:                   types.StringValue(dd.State),
		TunnelVersion:           types.StringValue(dd.TunnelVersion),
		Type:                    types.StringValue(dd.Type),
		Udid:                    types.StringValue(dd.UniqueID),
		UpmVersion:              types.StringValue(dd.UpmVersion),
		User:                    types.StringValue(dd.UserName),
		VpnState:                types.StringValue(""),
		ZappArch:                types.StringValue(dd.ZappArch),
	}
}
