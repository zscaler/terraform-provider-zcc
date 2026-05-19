package datasources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/devices"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &DeviceCleanupDataSource{}
	_ datasource.DataSourceWithConfigure = &DeviceCleanupDataSource{}
)

func NewDeviceCleanupDataSource() datasource.DataSource {
	return &DeviceCleanupDataSource{}
}

type DeviceCleanupDataSource struct {
	client *client.Client
}

type DeviceCleanupDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Active            types.Bool   `tfsdk:"active"`
	ForceRemoveType   types.String `tfsdk:"force_remove_type"`
	DeviceExceedLimit types.Int64  `tfsdk:"device_exceed_limit"`
	AutoRemovalDays   types.Int64  `tfsdk:"auto_removal_days"`
	AutoPurgeDays     types.Int64  `tfsdk:"auto_purge_days"`
}

func (d *DeviceCleanupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_cleanup"
}

func (d *DeviceCleanupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the ZCC device cleanup settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Settings record identifier.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether device cleanup is active.",
				Computed:    true,
			},
			"force_remove_type": schema.StringAttribute{
				Description: "Force remove type code (e.g. \"0\").",
				Computed:    true,
			},
			"device_exceed_limit": schema.Int64Attribute{
				Description: "Device exceed limit threshold.",
				Computed:    true,
			},
			"auto_removal_days": schema.Int64Attribute{
				Description: "Auto removal period in days.",
				Computed:    true,
			},
			"auto_purge_days": schema.Int64Attribute{
				Description: "Auto purge period in days.",
				Computed:    true,
			},
		},
	}
}

func (d *DeviceCleanupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeviceCleanupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading ZCC device cleanup settings")

	info, err := devices.GetDeviceCleanupInfo(ctx, d.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read device cleanup settings: %v", err))
		return
	}

	m := DeviceCleanupDataSourceModel{
		ID:                types.StringValue(info.ID),
		Active:            types.BoolValue(info.Active == "1"),
		ForceRemoveType:   types.StringValue(info.ForceRemoveType),
		DeviceExceedLimit: types.Int64Value(parseI64(info.DeviceExceedLimit)),
		AutoRemovalDays:   types.Int64Value(parseI64(info.AutoRemovalDays)),
		AutoPurgeDays:     types.Int64Value(parseI64(info.AutoPurgeDays)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func parseI64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
