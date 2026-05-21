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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/zia_posture"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/helpers"
)

var (
	_ datasource.DataSource              = &ZIAPostureDataSource{}
	_ datasource.DataSourceWithConfigure = &ZIAPostureDataSource{}
)

func NewZIAPostureDataSource() datasource.DataSource {
	return &ZIAPostureDataSource{}
}

type ZIAPostureDataSource struct {
	client *client.Client
}

// ZIAPostureDataSourceModel mirrors zia_posture.ZIAPosture. Unlike the
// resource it has no user-configurable / server-only split — every
// field is Computed so callers can introspect profiles managed
// elsewhere.
type ZIAPostureDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Platform            types.String `tfsdk:"platform"`
	HighTrustCriteria   types.Object `tfsdk:"high_trust_criteria"`
	MediumTrustCriteria types.Object `tfsdk:"medium_trust_criteria"`
	LowTrustCriteria    types.Object `tfsdk:"low_trust_criteria"`
}

func (d *ZIAPostureDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zia_posture"
}

func (d *ZIAPostureDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZIA posture profile from /zcc/papi/public/v2/zia-posture-profiles. " +
			"NOTE: lookup is currently restricted to the numeric `id` — the upstream " +
			"`/zia-posture-profiles` list endpoint does not paginate correctly and " +
			"therefore cannot be used to resolve a profile by name or by platform. " +
			"Once that API is fixed, name-based and platform-based lookup will be " +
			"reinstated; until then `id` is required.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the posture profile (carried as a string). Currently the only supported lookup key.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Operator-visible name. Returned by the API and surfaced as Computed; cannot currently be used as a lookup key (see resource description for the temporary API limitation).",
				Computed:    true,
			},
			"platform": schema.StringAttribute{
				Description: "Target operating system, surfaced as a name: `ios`, `android`, `windows`, `macos`, or `linux` (translated from the API's numeric platform code).",
				Computed:    true,
			},
			"high_trust_criteria": schema.SingleNestedAttribute{
				Description: "Criteria that promote a device to the HIGH trust tier.",
				Computed:    true,
				Attributes:  trustCriteriaDataSourceAttributes(),
			},
			"medium_trust_criteria": schema.SingleNestedAttribute{
				Description: "Criteria that promote a device to the MEDIUM trust tier.",
				Computed:    true,
				Attributes:  trustCriteriaDataSourceAttributes(),
			},
			"low_trust_criteria": schema.SingleNestedAttribute{
				Description: "Criteria that promote a device to the LOW trust tier.",
				Computed:    true,
				Attributes:  trustCriteriaDataSourceAttributes(),
			},
		},
	}
}

func (d *ZIAPostureDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ZIAPostureDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ZIAPostureDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TEMPORARY: only id-based lookup is supported. The upstream
	// `/zia-posture-profiles` list endpoint mishandles pagination and
	// silently returns a truncated set, so GetByName / platform-based
	// filtering cannot guarantee a correct match. Required+id-only
	// keeps the surface honest until the API is fixed; at that point
	// the name / platform branches can be reinstated.
	idStr := data.ID.ValueString()
	if data.ID.IsNull() || idStr == "" {
		resp.Diagnostics.AddError("Missing Identifier", "id is required (name-based lookup is temporarily disabled due to an upstream pagination bug)")
		return
	}

	id, convErr := strconv.Atoi(idStr)
	if convErr != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ZIA posture id %q is not a valid integer: %v", idStr, convErr))
		return
	}

	service := d.client.Service
	tflog.Info(ctx, "Fetching zia posture profile", map[string]any{"id": id})
	posture, err := zia_posture.Get(ctx, service, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read zia posture profile %d: %v", id, err))
		return
	}

	model := ZIAPostureDataSourceModel{
		ID:                  types.StringValue(strconv.Itoa(posture.ID)),
		Name:                types.StringValue(posture.Name),
		Platform:            types.StringValue(helpers.PlatformIntToName(posture.Platform)),
		HighTrustCriteria:   flattenTrustCriteriaSetsDataSource(posture.HighTrustCriteria.Cs),
		MediumTrustCriteria: flattenTrustCriteriaSetsDataSource(posture.MediumTrustCriteria.Cs),
		LowTrustCriteria:    flattenTrustCriteriaSetsDataSource(posture.LowTrustCriteria.Cs),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// =============================================================================
// nested-block helpers (data source / Computed-only)
//
// Mirrors the resource helpers in resources/zia_posture.go, but defined
// here so the datasources package does not import the resources package.
// =============================================================================

func trustCriterionDataSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"udid": types.StringType,
	}
}

func trustCriteriaSetDataSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cn": types.ListType{ElemType: types.ObjectType{AttrTypes: trustCriterionDataSourceAttrTypes()}},
	}
}

func trustCriteriaDataSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cs": types.ListType{ElemType: types.ObjectType{AttrTypes: trustCriteriaSetDataSourceAttrTypes()}},
	}
}

func trustCriteriaDataSourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cs": schema.ListNestedAttribute{
			Description: "OR-list of criteria sets. Each set is itself an AND-list of criteria in `cn`.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"cn": schema.ListNestedAttribute{
						Description: "AND-list of criteria — every criterion in this list must match.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id":   schema.StringAttribute{Description: "Criterion identifier.", Computed: true},
								"name": schema.StringAttribute{Description: "Operator-visible label.", Computed: true},
								"udid": schema.StringAttribute{Description: "Optional device UDID the criterion is scoped to.", Computed: true},
							},
						},
					},
				},
			},
		},
	}
}

func flattenTrustCriteriaSetsDataSource(sets []zia_posture.TrustCriteriaSet) types.Object {
	csElems := make([]attr.Value, 0, len(sets))
	for _, s := range sets {
		cnElems := make([]attr.Value, 0, len(s.Cn))
		for _, c := range s.Cn {
			critObj, _ := types.ObjectValue(trustCriterionDataSourceAttrTypes(), map[string]attr.Value{
				"id":   types.StringValue(c.ID),
				"name": types.StringValue(c.Name),
				"udid": types.StringValue(c.UDID),
			})
			cnElems = append(cnElems, critObj)
		}
		cnList, _ := types.ListValue(types.ObjectType{AttrTypes: trustCriterionDataSourceAttrTypes()}, cnElems)
		setObj, _ := types.ObjectValue(trustCriteriaSetDataSourceAttrTypes(), map[string]attr.Value{
			"cn": cnList,
		})
		csElems = append(csElems, setObj)
	}
	csList, _ := types.ListValue(types.ObjectType{AttrTypes: trustCriteriaSetDataSourceAttrTypes()}, csElems)
	obj, _ := types.ObjectValue(trustCriteriaDataSourceAttrTypes(), map[string]attr.Value{
		"cs": csList,
	})
	return obj
}
