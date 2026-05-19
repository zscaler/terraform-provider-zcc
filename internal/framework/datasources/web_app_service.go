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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/web_app_service"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &WebAppServiceDataSource{}
	_ datasource.DataSourceWithConfigure = &WebAppServiceDataSource{}
)

func NewWebAppServiceDataSource() datasource.DataSource {
	return &WebAppServiceDataSource{}
}

type WebAppServiceDataSource struct {
	client *client.Client
}

type WebAppServiceDataSourceModel struct {
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
	Version         types.Int64  `tfsdk:"version"`
}

func (d *WebAppServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_app_service"
}

func (d *WebAppServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC web app service (bypass app) by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the web app service.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the web app service (used for lookup).",
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
			"version":           schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *WebAppServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WebAppServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WebAppServiceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.Name.IsNull() || data.Name.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	var app *web_app_service.WebAppService

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching web app service by ID", map[string]any{"id": id})
		result, err := web_app_service.GetByAppID(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read web app service: %v", err))
			return
		}
		app = result
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching web app service by name", map[string]any{"name": name})
		result, err := web_app_service.GetByName(ctx, service, name)
		if err != nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Web app service with name '%s' not found: %v", name, err))
			return
		}
		app = result
	}

	model := WebAppServiceDataSourceModel{
		ID:              types.StringValue(strconv.Itoa(app.ID)),
		Name:            types.StringValue(app.AppName),
		AppVersion:      types.Int64Value(int64(app.AppVersion)),
		AppSvcId:        types.Int64Value(int64(app.AppSvcId)),
		AppName:         types.StringValue(app.AppName),
		Active:          types.BoolValue(app.Active),
		UID:             types.StringValue(app.UID),
		AppDataBlob:     flattenWebAppServiceDataBlobs(app.AppDataBlob),
		AppDataBlobV6:   flattenWebAppServiceDataBlobs(app.AppDataBlobV6),
		CreatedBy:       types.StringValue(app.CreatedBy),
		EditedBy:        types.StringValue(app.EditedBy),
		EditedTimestamp: types.StringValue(app.EditedTimestamp),
		ZappDataBlob:    types.StringValue(app.ZappDataBlob),
		ZappDataBlobV6:  types.StringValue(app.ZappDataBlobV6),
		Version:         types.Int64Value(int64(app.Version)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenWebAppServiceDataBlobs(blobs []web_app_service.AppDataBlob) types.List {
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
