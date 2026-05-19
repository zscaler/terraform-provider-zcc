package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/devices"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ resource.Resource                = &DeviceCleanupResource{}
	_ resource.ResourceWithConfigure   = &DeviceCleanupResource{}
	_ resource.ResourceWithImportState = &DeviceCleanupResource{}
)

func NewDeviceCleanupResource() resource.Resource {
	return &DeviceCleanupResource{}
}

type DeviceCleanupResource struct {
	client *client.Client
}

type DeviceCleanupResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Active            types.Bool   `tfsdk:"active"`
	ForceRemoveType   types.String `tfsdk:"force_remove_type"`
	DeviceExceedLimit types.Int64  `tfsdk:"device_exceed_limit"`
	AutoRemovalDays   types.Int64  `tfsdk:"auto_removal_days"`
	AutoPurgeDays     types.Int64  `tfsdk:"auto_purge_days"`
}

func (r *DeviceCleanupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_cleanup"
}

func (r *DeviceCleanupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ZCC device cleanup settings (singleton). The API only supports GET and PUT; create runs an initial PUT, update runs PUT, and delete only removes the resource from state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Settings record identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether device cleanup is active.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"force_remove_type": schema.StringAttribute{
				Description: "Force remove type code. Supported values: `0` (Restrict), `8`, `9`, `10`, `11`, `12`, `13`, `14`, `15`, `16`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("0", "8", "9", "10", "11", "12", "13", "14", "15", "16"),
				},
			},
			"device_exceed_limit": schema.Int64Attribute{
				Description: "Device exceed limit threshold.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"auto_removal_days": schema.Int64Attribute{
				Description: "Auto removal period in days. Supported values: `0` (Never), `30`, `60`, `90`, `120`, `150`, `180`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(0, 30, 60, 90, 120, 150, 180),
				},
			},
			"auto_purge_days": schema.Int64Attribute{
				Description: "Auto purge period in days. Supported values: `0`, `30`, `60`, `90`, `120`, `150`, `180`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.OneOf(0, 30, 60, 90, 120, 150, 180),
				},
			},
		},
	}
}

func (r *DeviceCleanupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeviceCleanupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan DeviceCleanupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existing, err := devices.GetDeviceCleanupInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read device cleanup settings: %v", err))
		return
	}

	payload := expandDeviceCleanup(&plan, existing)
	tflog.Info(ctx, "Updating ZCC device cleanup (singleton create)", map[string]any{"id": payload.ID})

	updated, err := devices.SetDeviceCleanupInfo(ctx, r.client.Service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update device cleanup: %v", err))
		return
	}

	flattenDeviceCleanupResource(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceCleanupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state DeviceCleanupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := devices.GetDeviceCleanupInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read device cleanup settings: %v", err))
		return
	}

	flattenDeviceCleanupResource(info, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceCleanupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan DeviceCleanupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existing, err := devices.GetDeviceCleanupInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read device cleanup settings: %v", err))
		return
	}

	payload := expandDeviceCleanup(&plan, existing)
	tflog.Info(ctx, "Updating ZCC device cleanup", map[string]any{"id": payload.ID})

	updated, err := devices.SetDeviceCleanupInfo(ctx, r.client.Service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update device cleanup: %v", err))
		return
	}

	flattenDeviceCleanupResource(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceCleanupResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *DeviceCleanupResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	info, err := devices.GetDeviceCleanupInfo(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to read device cleanup settings: %v", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(info.ID))...)
}

func expandDeviceCleanup(plan *DeviceCleanupResourceModel, existing *devices.DeviceCleanupInfo) *devices.DeviceCleanupInfo {
	p := *existing

	if !plan.Active.IsNull() && !plan.Active.IsUnknown() {
		if plan.Active.ValueBool() {
			p.Active = "1"
		} else {
			p.Active = "0"
		}
	}
	if !plan.ForceRemoveType.IsNull() && !plan.ForceRemoveType.IsUnknown() {
		p.ForceRemoveType = plan.ForceRemoveType.ValueString()
	}
	if !plan.DeviceExceedLimit.IsNull() && !plan.DeviceExceedLimit.IsUnknown() {
		p.DeviceExceedLimit = strconv.FormatInt(plan.DeviceExceedLimit.ValueInt64(), 10)
	}
	if !plan.AutoRemovalDays.IsNull() && !plan.AutoRemovalDays.IsUnknown() {
		p.AutoRemovalDays = strconv.FormatInt(plan.AutoRemovalDays.ValueInt64(), 10)
	}
	if !plan.AutoPurgeDays.IsNull() && !plan.AutoPurgeDays.IsUnknown() {
		p.AutoPurgeDays = strconv.FormatInt(plan.AutoPurgeDays.ValueInt64(), 10)
	}

	return &p
}

func flattenDeviceCleanupResource(p *devices.DeviceCleanupInfo, m *DeviceCleanupResourceModel) {
	m.ID = types.StringValue(p.ID)
	m.Active = types.BoolValue(p.Active == "1")
	m.ForceRemoveType = types.StringValue(p.ForceRemoveType)
	m.DeviceExceedLimit = types.Int64Value(mustParseInt64Str(p.DeviceExceedLimit))
	m.AutoRemovalDays = types.Int64Value(mustParseInt64Str(p.AutoRemovalDays))
	m.AutoPurgeDays = types.Int64Value(mustParseInt64Str(p.AutoPurgeDays))
}

func mustParseInt64Str(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
