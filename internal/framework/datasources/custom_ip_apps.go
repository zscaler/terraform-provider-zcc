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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/custom_ip_apps"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &CustomIPAppsDataSource{}
	_ datasource.DataSourceWithConfigure = &CustomIPAppsDataSource{}
)

func NewCustomIPAppsDataSource() datasource.DataSource {
	return &CustomIPAppsDataSource{}
}

type CustomIPAppsDataSource struct {
	client *client.Client
}

type CustomIPAppsDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
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

func (d *CustomIPAppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_ip_apps"
}

func appDataBlobSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"proto":  schema.StringAttribute{Computed: true},
				"port":   schema.StringAttribute{Computed: true},
				"ipaddr": schema.StringAttribute{Computed: true},
				"fqdn":   schema.StringAttribute{Computed: true},
			},
		},
	}
}

var appDataBlobObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"proto":  types.StringType,
		"port":   types.StringType,
		"ipaddr": types.StringType,
		"fqdn":   types.StringType,
	},
}

func flattenAppDataBlobs(blobs []custom_ip_apps.AppDataBlob) types.List {
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

func (d *CustomIPAppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC custom IP-based app by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the custom IP-based app.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the custom IP-based app (used for lookup).",
				Optional:    true,
			},
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

func (d *CustomIPAppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CustomIPAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CustomIPAppsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.Name.IsNull() || data.Name.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	var app *custom_ip_apps.CustomIPApp

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching custom IP app by ID", map[string]any{"id": id})
		result, _, err := custom_ip_apps.GetByAppID(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read custom IP app: %v", err))
			return
		}
		app = result
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching custom IP app by name", map[string]any{"name": name})
		result, _, err := custom_ip_apps.GetByName(ctx, service, name)
		if err != nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Custom IP app with name '%s' not found: %v", name, err))
			return
		}
		app = result
	}

	model := CustomIPAppsDataSourceModel{
		ID:              types.StringValue(strconv.Itoa(app.ID)),
		Name:            types.StringValue(app.AppName),
		AppName:         types.StringValue(app.AppName),
		Active:          types.BoolValue(app.Active),
		UID:             types.StringValue(app.UID),
		AppDataBlob:     flattenAppDataBlobs(app.AppDataBlob),
		AppDataBlobV6:   flattenAppDataBlobs(app.AppDataBlobV6),
		CreatedBy:       types.StringValue(app.CreatedBy),
		EditedBy:        types.StringValue(app.EditedBy),
		EditedTimestamp: types.StringValue(app.EditedTimestamp),
		ZappDataBlob:    types.StringValue(app.ZappDataBlob),
		ZappDataBlobV6:  types.StringValue(app.ZappDataBlobV6),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
