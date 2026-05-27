package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/zia_posture"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/helpers"
)

var (
	_ resource.Resource                = &ZIAPostureResource{}
	_ resource.ResourceWithConfigure   = &ZIAPostureResource{}
	_ resource.ResourceWithImportState = &ZIAPostureResource{}
)

func NewZIAPostureResource() resource.Resource {
	return &ZIAPostureResource{}
}

type ZIAPostureResource struct {
	client *client.Client
}

type ZIAPostureResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Platform            types.String `tfsdk:"platform"`
	HighTrustCriteria   types.Object `tfsdk:"high_trust_criteria"`
	MediumTrustCriteria types.Object `tfsdk:"medium_trust_criteria"`
	LowTrustCriteria    types.Object `tfsdk:"low_trust_criteria"`
}

func (r *ZIAPostureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zia_posture"
}

func (r *ZIAPostureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZIA posture profile via /zcc/papi/public/v2/zia-posture-profiles. " +
			"A posture profile bundles the device-trust criteria that ZCC evaluates against the local " +
			"machine, classifying each endpoint into one of three trust tiers (high / medium / low). " +
			"Each tier carries an OR-of-AND set of criteria: every criterion inside a `cs[].cn` block " +
			"must match (AND), and matching any single set in `cs` (OR) promotes the device to that tier.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric identifier of the posture profile, carried as a string per Terraform convention. API field: id (JSON number).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Operator-visible name. API field: name.",
			},
			"platform": schema.StringAttribute{
				Required: true,
				Description: "Target operating system. One of " +
					"`ios`, `android`, `windows`, `macos`, `linux` (case-insensitive). " +
					"Translated to the API's numeric platform code (1..5) at the SDK boundary. " +
					"API field: platform.",
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive(helpers.PlatformNames()...),
				},
			},
			"high_trust_criteria": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Criteria that promote a device to the HIGH trust tier. API field: highTrustCriteria.",
				Attributes:  trustCriteriaAttributes(),
			},
			"medium_trust_criteria": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Criteria that promote a device to the MEDIUM trust tier. API field: mediumTrustCriteria.",
				Attributes:  trustCriteriaAttributes(),
			},
			"low_trust_criteria": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Criteria that promote a device to the LOW trust tier. API field: lowTrustCriteria.",
				Attributes:  trustCriteriaAttributes(),
			},
		},
	}
}

func (r *ZIAPostureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ZIAPostureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan ZIAPostureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	payload := expandZIAPosture(&plan)

	tflog.Info(ctx, "Creating ZCC ZIA posture profile", map[string]any{"name": payload.Name, "platform": payload.Platform})

	created, _, err := zia_posture.Create(ctx, service, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create zia posture profile: %v", err))
		return
	}

	tflog.Info(ctx, "Created ZCC ZIA posture profile", map[string]any{"id": created.ID})
	flattenZIAPosture(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ZIAPostureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state ZIAPostureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, convErr := strconv.Atoi(state.ID.ValueString())
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ZIA posture id %q is not a valid integer: %v", state.ID.ValueString(), convErr))
		return
	}

	service := r.client.Service
	posture, err := zia_posture.Get(ctx, service, id)
	if err != nil {
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "Removing zia posture profile from state - no longer exists", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read zia posture profile %d: %v", id, err))
		return
	}

	flattenZIAPosture(posture, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ZIAPostureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan ZIAPostureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state ZIAPostureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, convErr := strconv.Atoi(state.ID.ValueString())
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ZIA posture id %q is not a valid integer: %v", state.ID.ValueString(), convErr))
		return
	}

	service := r.client.Service
	payload := expandZIAPosture(&plan)
	payload.ID = id

	tflog.Info(ctx, "Updating ZCC ZIA posture profile", map[string]any{"id": id})

	updated, _, err := zia_posture.Update(ctx, service, id, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update zia posture profile %d: %v", id, err))
		return
	}

	flattenZIAPosture(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ZIAPostureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state ZIAPostureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, convErr := strconv.Atoi(state.ID.ValueString())
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ZIA posture id %q is not a valid integer: %v", state.ID.ValueString(), convErr))
		return
	}

	service := r.client.Service

	if _, err := zia_posture.Get(ctx, service, id); err != nil {
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "ZIA posture profile already removed upstream; nothing to delete", map[string]any{"id": id})
			return
		}
		tflog.Warn(ctx, "Pre-delete GET failed; proceeding to DELETE anyway", map[string]any{"id": id, "error": err.Error()})
	}

	tflog.Info(ctx, "Deleting ZCC ZIA posture profile", map[string]any{"id": id})
	if _, err := zia_posture.Delete(ctx, service, id); err != nil {
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "ZIA posture profile was removed between GET and DELETE; treating as success", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to delete zia posture profile %d: %v", id, err))
	}
}

func (r *ZIAPostureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	id := req.ID
	if _, parseErr := strconv.ParseInt(id, 10, 64); parseErr == nil {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		return
	}

	service := r.client.Service
	posture, err := zia_posture.GetByName(ctx, service, id)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import zia posture profile %q: %v", id, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(strconv.Itoa(posture.ID)))...)
}

// ---------------------------------------------------------------------------
// expand / flatten — top level
// ---------------------------------------------------------------------------

func expandZIAPosture(plan *ZIAPostureResourceModel) zia_posture.ZIAPosture {
	return zia_posture.ZIAPosture{
		Name:                plan.Name.ValueString(),
		Platform:            helpers.PlatformNameToInt(plan.Platform.ValueString()),
		HighTrustCriteria:   zia_posture.HighTrustCriteria{Cs: expandTrustCriteriaSets(plan.HighTrustCriteria)},
		MediumTrustCriteria: zia_posture.MediumTrustCriteria{Cs: expandTrustCriteriaSets(plan.MediumTrustCriteria)},
		LowTrustCriteria:    zia_posture.LowTrustCriteria{Cs: expandTrustCriteriaSets(plan.LowTrustCriteria)},
	}
}

func flattenZIAPosture(p *zia_posture.ZIAPosture, model *ZIAPostureResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(p.ID))
	model.Name = types.StringValue(p.Name)
	model.Platform = types.StringValue(helpers.PlatformIntToName(p.Platform))
	model.HighTrustCriteria = flattenTrustCriteriaSets(p.HighTrustCriteria.Cs)
	model.MediumTrustCriteria = flattenTrustCriteriaSets(p.MediumTrustCriteria.Cs)
	model.LowTrustCriteria = flattenTrustCriteriaSets(p.LowTrustCriteria.Cs)
}

// =============================================================================
// nested-block helpers
//
// The three top-level trust-tier blocks (high / medium / low) all share
// the exact same nested shape:
//
//	{ cs: [ { cn: [ { id, name, udid } ] } ] }
//
// So we factor the schema + attr.Type maps + expand/flatten helpers once
// here and reuse them across the three top-level blocks (and across the
// matching data source).
// =============================================================================

// trustCriterionAttrTypes is the attr.Type map for one TrustCriterion
// (the leaf object in `cs[].cn[]`).
func trustCriterionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"udid": types.StringType,
	}
}

// trustCriteriaSetAttrTypes is the attr.Type map for one TrustCriteriaSet
// (the `cs[]` object, which wraps a list of TrustCriterion entries).
func trustCriteriaSetAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cn": types.ListType{ElemType: types.ObjectType{AttrTypes: trustCriterionAttrTypes()}},
	}
}

// trustCriteriaAttrTypes is the attr.Type map for the whole
// `*_trust_criteria` block — a single nested object whose only field is
// the `cs` list of TrustCriteriaSet entries.
func trustCriteriaAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cs": types.ListType{ElemType: types.ObjectType{AttrTypes: trustCriteriaSetAttrTypes()}},
	}
}

// trustCriteriaAttributes returns the schema.Attribute map used by all
// three top-level trust-tier blocks on the resource. The data source
// reuses an equivalent Computed-only variant.
func trustCriteriaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cs": schema.ListNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: "OR-list of criteria sets. Each set is itself an AND-list of criteria in `cn`. API field: cs.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"cn": schema.ListNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "AND-list of criteria — every criterion in this list must match for the parent set to be considered satisfied. API field: cn.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required:    true,
									Description: "Criterion identifier. API field: id.",
								},
								"name": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Operator-visible label. API field: name.",
								},
								"udid": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Optional device UDID the criterion is scoped to. API field: udid.",
								},
							},
						},
					},
				},
			},
		},
	}
}

// expandTrustCriteriaSets walks a *_trust_criteria nested object and
// returns the SDK's []TrustCriteriaSet. Null / unknown input produces an
// empty slice so the SDK can `omitempty` it on the wire.
func expandTrustCriteriaSets(obj types.Object) []zia_posture.TrustCriteriaSet {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	attrs := obj.Attributes()
	csVal, ok := attrs["cs"]
	if !ok {
		return nil
	}
	csList, ok := csVal.(types.List)
	if !ok || csList.IsNull() || csList.IsUnknown() {
		return nil
	}
	elems := csList.Elements()
	out := make([]zia_posture.TrustCriteriaSet, 0, len(elems))
	for _, e := range elems {
		setObj, ok := e.(types.Object)
		if !ok {
			continue
		}
		out = append(out, zia_posture.TrustCriteriaSet{
			Cn: expandTrustCriteria(setObj),
		})
	}
	return out
}

// expandTrustCriteria walks a single `cs[]` element and returns the
// SDK's []TrustCriterion (the inner `cn` list).
func expandTrustCriteria(setObj types.Object) []zia_posture.TrustCriterion {
	attrs := setObj.Attributes()
	cnVal, ok := attrs["cn"]
	if !ok {
		return nil
	}
	cnList, ok := cnVal.(types.List)
	if !ok || cnList.IsNull() || cnList.IsUnknown() {
		return nil
	}
	elems := cnList.Elements()
	out := make([]zia_posture.TrustCriterion, 0, len(elems))
	for _, e := range elems {
		critObj, ok := e.(types.Object)
		if !ok {
			continue
		}
		a := critObj.Attributes()
		out = append(out, zia_posture.TrustCriterion{
			ID:   stringFromAttrs(a, "id"),
			Name: stringFromAttrs(a, "name"),
			UDID: stringFromAttrs(a, "udid"),
		})
	}
	return out
}

// flattenTrustCriteriaSets rebuilds a *_trust_criteria types.Object from
// the SDK's []TrustCriteriaSet, producing the exact attr.Type shape
// declared in trustCriteriaAttrTypes so the framework accepts it without
// type-mismatch errors.
func flattenTrustCriteriaSets(sets []zia_posture.TrustCriteriaSet) types.Object {
	csElems := make([]attr.Value, 0, len(sets))
	for _, s := range sets {
		cnElems := make([]attr.Value, 0, len(s.Cn))
		for _, c := range s.Cn {
			critObj, _ := types.ObjectValue(trustCriterionAttrTypes(), map[string]attr.Value{
				"id":   types.StringValue(c.ID),
				"name": types.StringValue(c.Name),
				"udid": types.StringValue(c.UDID),
			})
			cnElems = append(cnElems, critObj)
		}
		cnList, _ := types.ListValue(types.ObjectType{AttrTypes: trustCriterionAttrTypes()}, cnElems)
		setObj, _ := types.ObjectValue(trustCriteriaSetAttrTypes(), map[string]attr.Value{
			"cn": cnList,
		})
		csElems = append(csElems, setObj)
	}
	csList, _ := types.ListValue(types.ObjectType{AttrTypes: trustCriteriaSetAttrTypes()}, csElems)
	obj, _ := types.ObjectValue(trustCriteriaAttrTypes(), map[string]attr.Value{
		"cs": csList,
	})
	return obj
}

// stringFromAttrs reads a `types.String` out of an Attributes() map by
// name, returning the empty string if the key is missing, null, unknown,
// or holds a non-String value. Local to the zia_posture file so we don't
// leak this very narrow helper into the broader helpers package.
func stringFromAttrs(attrs map[string]attr.Value, key string) string {
	v, ok := attrs[key]
	if !ok || v == nil || v.IsNull() || v.IsUnknown() {
		return ""
	}
	if s, ok := v.(types.String); ok {
		return s.ValueString()
	}
	return ""
}
