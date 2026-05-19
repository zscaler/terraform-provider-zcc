package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/web_privacy"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &WebPrivacyDataSource{}
	_ datasource.DataSourceWithConfigure = &WebPrivacyDataSource{}
)

func NewWebPrivacyDataSource() datasource.DataSource {
	return &WebPrivacyDataSource{}
}

type WebPrivacyDataSource struct {
	client *client.Client
}

type WebPrivacyDataSourceModel struct {
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

func (d *WebPrivacyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_privacy"
}

func (d *WebPrivacyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the ZCC web privacy settings.",
		Attributes: map[string]schema.Attribute{
			"id":                                 schema.StringAttribute{Description: "Settings record identifier.", Computed: true},
			"active":                             schema.BoolAttribute{Description: "Whether web privacy settings are active.", Computed: true},
			"collect_user_info":                  schema.BoolAttribute{Description: "Whether to collect user information.", Computed: true},
			"collect_machine_hostname":           schema.BoolAttribute{Description: "Whether to collect machine hostname.", Computed: true},
			"collect_zdx_location":               schema.BoolAttribute{Description: "Whether to collect ZDX location.", Computed: true},
			"enable_packet_capture":              schema.BoolAttribute{Description: "Whether packet capture is enabled.", Computed: true},
			"disable_crashlytics":                schema.BoolAttribute{Description: "Whether Crashlytics is disabled.", Computed: true},
			"override_t2_protocol_setting":       schema.BoolAttribute{Description: "Whether to override T2 protocol setting.", Computed: true},
			"restrict_remote_packet_capture":     schema.BoolAttribute{Description: "Whether remote packet capture is restricted.", Computed: true},
			"grant_access_to_zscaler_log_folder": schema.BoolAttribute{Description: "Whether to grant access to Zscaler log folder.", Computed: true},
			"export_logs_for_non_admin":          schema.BoolAttribute{Description: "Whether non-admin users can export logs.", Computed: true},
			"enable_auto_log_snippet":            schema.BoolAttribute{Description: "Whether automatic log snippet collection is enabled.", Computed: true},
			"enforce_secure_pac_urls":            schema.BoolAttribute{Description: "Whether secure PAC URLs are enforced.", Computed: true},
			"enable_fqdn_match_for_vpn_bypasses": schema.BoolAttribute{Description: "Whether FQDN matching for VPN bypasses is enabled.", Computed: true},
		},
	}
}

func (d *WebPrivacyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *WebPrivacyDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading ZCC web privacy settings")

	info, err := web_privacy.GetWebPrivacyInfo(ctx, d.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read web privacy settings: %v", err))
		return
	}

	m := WebPrivacyDataSourceModel{
		ID:                            types.StringValue(info.ID),
		Active:                        types.BoolValue(info.Active == "1"),
		CollectUserInfo:               types.BoolValue(info.CollectUserInfo == "1"),
		CollectMachineHostname:        types.BoolValue(info.CollectMachineHostname == "1"),
		CollectZdxLocation:            types.BoolValue(info.CollectZdxLocation == "1"),
		EnablePacketCapture:           types.BoolValue(info.EnablePacketCapture == "1"),
		DisableCrashlytics:            types.BoolValue(info.DisableCrashlytics == "1"),
		OverrideT2ProtocolSetting:     types.BoolValue(info.OverrideT2ProtocolSetting == "1"),
		RestrictRemotePacketCapture:   types.BoolValue(info.RestrictRemotePacketCapture == "1"),
		GrantAccessToZscalerLogFolder: types.BoolValue(info.GrantAccessToZscalerLogFolder == "1"),
		ExportLogsForNonAdmin:         types.BoolValue(info.ExportLogsForNonAdmin == "1"),
		EnableAutoLogSnippet:          types.BoolValue(info.EnableAutoLogSnippet == "1"),
		EnforceSecurePacUrls:          types.BoolValue(info.EnforceSecurePacUrls == "1"),
		EnableFQDNMatchForVpnBypasses: types.BoolValue(info.EnableFQDNMatchForVpnBypasses == "1"),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}
