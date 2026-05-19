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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/predefined_ip_apps"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &PredefinedIPAppsDataSource{}
	_ datasource.DataSourceWithConfigure = &PredefinedIPAppsDataSource{}
)

func NewPredefinedIPAppsDataSource() datasource.DataSource {
	return &PredefinedIPAppsDataSource{}
}

type PredefinedIPAppsDataSource struct {
	client *client.Client
}

type PredefinedIPAppsDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AppVersion      types.Int64  `tfsdk:"app_version"`
	AppSvcId        types.Int64  `tfsdk:"app_svc_id"`
	AppName         types.String `tfsdk:"app_name"`
	Active          types.Bool   `tfsdk:"active"`
	UID             types.String `tfsdk:"uid"`
	AppDataBlob     types.List   `tfsdk:"app_data_blob"`
	AppDataBlobV6   types.List   `tfsdk:"app_data_blob_v6"`
	CreatedBy       types.String `tfsdk:"created_by"`
	EditedBy        types.String `tfsdk:"edited_by"`
	EditedTimestamp types.String `tfsdk:"edited_timestamp"`
	ZappDataBlob    types.String `tfsdk:"zapp_data_blob"`
	ZappDataBlobV6  types.String `tfsdk:"zapp_data_blob_v6"`
}

func (d *PredefinedIPAppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_predefined_ip_apps"
}

func flattenPredefinedAppDataBlobs(blobs []predefined_ip_apps.AppDataBlob) types.List {
	if len(blobs) == 0 {
		return types.ListNull(appDataBlobObjType)
	}
	elements := make([]attr.Value, 0, len(blobs))
	for _, b := range blobs {
		obj, _ := types.ObjectValue(appDataBlobObjType.AttrTypes, map[string]attr.Value{
			"proto":  types.StringValue(b.Proto),
			"port":   types.StringValue(b.Port),
			"ipaddr": types.StringValue(b.Ipaddr),
			"fqdn":   types.StringValue(b.Fqdn),
		})
		elements = append(elements, obj)
	}
	list, _ := types.ListValue(appDataBlobObjType, elements)
	return list
}

func (d *PredefinedIPAppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC predefined IP-based app by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the predefined IP-based app.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the predefined IP-based app (used for lookup).",
				Optional:    true,
			},
			"app_version":       schema.Int64Attribute{Computed: true},
			"app_svc_id":        schema.Int64Attribute{Computed: true},
			"app_name":          schema.StringAttribute{Computed: true},
			"active":            schema.BoolAttribute{Computed: true},
			"uid":               schema.StringAttribute{Computed: true},
			"app_data_blob":     appDataBlobSchema(),
			"app_data_blob_v6":  appDataBlobSchema(),
			"created_by":        schema.StringAttribute{Computed: true},
			"edited_by":         schema.StringAttribute{Computed: true},
			"edited_timestamp":  schema.StringAttribute{Computed: true},
			"zapp_data_blob":    schema.StringAttribute{Computed: true},
			"zapp_data_blob_v6": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *PredefinedIPAppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PredefinedIPAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PredefinedIPAppsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.Name.IsNull() || data.Name.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	var app *predefined_ip_apps.PredefinedIPApp

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching predefined IP app by ID", map[string]any{"id": id})
		result, _, err := predefined_ip_apps.GetByAppID(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read predefined IP app: %v", err))
			return
		}
		app = result
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching predefined IP app by name", map[string]any{"name": name})
		result, _, err := predefined_ip_apps.GetByName(ctx, service, name)
		if err != nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Predefined IP app with name '%s' not found: %v", name, err))
			return
		}
		app = result
	}

	model := PredefinedIPAppsDataSourceModel{
		ID:              types.StringValue(strconv.Itoa(app.ID)),
		Name:            types.StringValue(app.AppName),
		AppVersion:      types.Int64Value(int64(app.AppVersion)),
		AppSvcId:        types.Int64Value(int64(app.AppSvcId)),
		AppName:         types.StringValue(app.AppName),
		Active:          types.BoolValue(app.Active),
		UID:             types.StringValue(app.UID),
		AppDataBlob:     flattenPredefinedAppDataBlobs(app.AppDataBlob),
		AppDataBlobV6:   flattenPredefinedAppDataBlobs(app.AppDataBlobV6),
		CreatedBy:       types.StringValue(app.CreatedBy),
		EditedBy:        types.StringValue(app.EditedBy),
		EditedTimestamp: types.StringValue(app.EditedTimestamp),
		ZappDataBlob:    types.StringValue(app.ZappDataBlob),
		ZappDataBlobV6:  types.StringValue(app.ZappDataBlobV6),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
