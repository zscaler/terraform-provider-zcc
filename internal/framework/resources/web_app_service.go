package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/web_app_service"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ resource.Resource                = &WebAppServiceResource{}
	_ resource.ResourceWithConfigure   = &WebAppServiceResource{}
	_ resource.ResourceWithImportState = &WebAppServiceResource{}
)

func NewWebAppServiceResource() resource.Resource {
	return &WebAppServiceResource{}
}

type WebAppServiceResource struct {
	client *client.Client
}

type WebAppServiceResourceModel struct {
	ID              types.String `tfsdk:"id"`
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

var webAppBlobObjType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"proto":  types.StringType,
		"port":   types.StringType,
		"ipaddr": types.StringType,
		"fqdn":   types.StringType,
	},
}

func (r *WebAppServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_app_service"
}

func (r *WebAppServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC web app service (bypass app). This is a singleton-style resource — the service always exists and cannot be created or deleted, only updated.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the web app service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_name": schema.StringAttribute{
				Description: "The name of the web app service. Used to identify the service during initial creation (import).",
				Required:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the web app service is active.",
				Optional:    true,
				Computed:    true,
			},
			"app_version": schema.Int64Attribute{
				Description: "The application version.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"app_svc_id": schema.Int64Attribute{
				Description: "The application service identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"uid": schema.StringAttribute{
				Description: "The unique identifier string.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_data_blob": schema.ListNestedAttribute{
				Description: "IPv4 application data entries.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"proto":  schema.StringAttribute{Optional: true, Computed: true},
						"port":   schema.StringAttribute{Optional: true, Computed: true},
						"ipaddr": schema.StringAttribute{Optional: true, Computed: true},
						"fqdn":   schema.StringAttribute{Optional: true, Computed: true},
					},
				},
			},
			"app_data_blob_v6": schema.ListNestedAttribute{
				Description: "IPv6 application data entries.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"proto":  schema.StringAttribute{Optional: true, Computed: true},
						"port":   schema.StringAttribute{Optional: true, Computed: true},
						"ipaddr": schema.StringAttribute{Optional: true, Computed: true},
						"fqdn":   schema.StringAttribute{Optional: true, Computed: true},
					},
				},
			},
			"created_by": schema.StringAttribute{
				Description: "User who created the service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"edited_by": schema.StringAttribute{
				Description: "User who last edited the service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"edited_timestamp": schema.StringAttribute{
				Description: "Timestamp of the last edit.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zapp_data_blob": schema.StringAttribute{
				Description: "Zapp data blob.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zapp_data_blob_v6": schema.StringAttribute{
				Description: "Zapp data blob (IPv6).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "The version number.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *WebAppServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *WebAppServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan WebAppServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	name := plan.AppName.ValueString()

	tflog.Info(ctx, "Looking up existing web app service by name for singleton create", map[string]any{"name": name})

	existing, err := web_app_service.GetByName(ctx, service, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to find web app service with name '%s': %v", name, err))
		return
	}

	payload := r.expandWebAppService(&plan, existing)

	tflog.Info(ctx, "Updating ZCC web app service (singleton create)", map[string]any{"id": payload.ID})

	updated, err := web_app_service.UpdateWebAppService(ctx, service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update web app service: %v", err))
		return
	}

	r.flattenWebAppService(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebAppServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state WebAppServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	id := state.ID.ValueString()

	tflog.Info(ctx, "Reading ZCC web app service", map[string]any{"id": id})

	app, err := web_app_service.GetByAppID(ctx, service, id)
	if err != nil {
		tflog.Info(ctx, "Removing web app service from state - no longer exists", map[string]any{"id": id})
		resp.State.RemoveResource(ctx)
		return
	}

	r.flattenWebAppService(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebAppServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan WebAppServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WebAppServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service

	existing, err := web_app_service.GetByAppID(ctx, service, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read web app service: %v", err))
		return
	}

	payload := r.expandWebAppService(&plan, existing)

	tflog.Info(ctx, "Updating ZCC web app service", map[string]any{"id": payload.ID})

	updated, err := web_app_service.UpdateWebAppService(ctx, service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update web app service: %v", err))
		return
	}

	r.flattenWebAppService(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebAppServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete called on singleton web app service — removing from state only (no API delete)")
}

func (r *WebAppServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	service := r.client.Service
	id := req.ID

	app, err := web_app_service.GetByAppID(ctx, service, id)
	if err != nil {
		app, err = web_app_service.GetByName(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import web app service with identifier '%s': %v", id, err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(strconv.Itoa(app.ID)))...)
}

func (r *WebAppServiceResource) expandWebAppService(plan *WebAppServiceResourceModel, existing *web_app_service.WebAppService) *web_app_service.WebAppService {
	payload := *existing

	if !plan.AppName.IsNull() && !plan.AppName.IsUnknown() {
		payload.AppName = plan.AppName.ValueString()
	}
	if !plan.Active.IsNull() && !plan.Active.IsUnknown() {
		payload.Active = plan.Active.ValueBool()
	}
	if !plan.ZappDataBlob.IsNull() && !plan.ZappDataBlob.IsUnknown() {
		payload.ZappDataBlob = plan.ZappDataBlob.ValueString()
	}
	if !plan.ZappDataBlobV6.IsNull() && !plan.ZappDataBlobV6.IsUnknown() {
		payload.ZappDataBlobV6 = plan.ZappDataBlobV6.ValueString()
	}
	if !plan.AppDataBlob.IsNull() && !plan.AppDataBlob.IsUnknown() {
		payload.AppDataBlob = r.expandAppDataBlobs(plan.AppDataBlob)
	}
	if !plan.AppDataBlobV6.IsNull() && !plan.AppDataBlobV6.IsUnknown() {
		payload.AppDataBlobV6 = r.expandAppDataBlobs(plan.AppDataBlobV6)
	}

	return &payload
}

func (r *WebAppServiceResource) expandAppDataBlobs(list types.List) []web_app_service.AppDataBlob {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	var blobs []web_app_service.AppDataBlob
	for _, elem := range list.Elements() {
		obj := elem.(types.Object)
		attrs := obj.Attributes()
		blob := web_app_service.AppDataBlob{}
		if v, ok := attrs["proto"]; ok && !v.(types.String).IsNull() {
			blob.Proto = v.(types.String).ValueString()
		}
		if v, ok := attrs["port"]; ok && !v.(types.String).IsNull() {
			blob.Port = v.(types.String).ValueString()
		}
		if v, ok := attrs["ipaddr"]; ok && !v.(types.String).IsNull() {
			blob.Ipaddr = v.(types.String).ValueString()
		}
		if v, ok := attrs["fqdn"]; ok && !v.(types.String).IsNull() {
			blob.Fqdn = v.(types.String).ValueString()
		}
		blobs = append(blobs, blob)
	}
	return blobs
}

func (r *WebAppServiceResource) flattenWebAppService(app *web_app_service.WebAppService, model *WebAppServiceResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(app.ID))
	model.AppName = types.StringValue(app.AppName)
	model.Active = types.BoolValue(app.Active)
	model.AppVersion = types.Int64Value(int64(app.AppVersion))
	model.AppSvcId = types.Int64Value(int64(app.AppSvcId))
	model.UID = types.StringValue(app.UID)
	model.CreatedBy = types.StringValue(app.CreatedBy)
	model.EditedBy = types.StringValue(app.EditedBy)
	model.EditedTimestamp = types.StringValue(app.EditedTimestamp)
	model.ZappDataBlob = types.StringValue(app.ZappDataBlob)
	model.ZappDataBlobV6 = types.StringValue(app.ZappDataBlobV6)
	model.Version = types.Int64Value(int64(app.Version))
	model.AppDataBlob = r.flattenAppDataBlobs(app.AppDataBlob)
	model.AppDataBlobV6 = r.flattenAppDataBlobs(app.AppDataBlobV6)
}

func (r *WebAppServiceResource) flattenAppDataBlobs(blobs []web_app_service.AppDataBlob) types.List {
	if len(blobs) == 0 {
		return types.ListNull(webAppBlobObjType)
	}
	elements := make([]attr.Value, 0, len(blobs))
	for _, b := range blobs {
		obj, _ := types.ObjectValue(webAppBlobObjType.AttrTypes, map[string]attr.Value{
			"proto":  types.StringValue(b.Proto),
			"port":   types.StringValue(b.Port),
			"ipaddr": types.StringValue(b.Ipaddr),
			"fqdn":   types.StringValue(b.Fqdn),
		})
		elements = append(elements, obj)
	}
	list, _ := types.ListValue(webAppBlobObjType, elements)
	return list
}
