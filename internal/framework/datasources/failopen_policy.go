package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/failopen_policy"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &FailOpenPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &FailOpenPolicyDataSource{}
)

func NewFailOpenPolicyDataSource() datasource.DataSource {
	return &FailOpenPolicyDataSource{}
}

type FailOpenPolicyDataSource struct {
	client *client.Client
}

type FailOpenPolicyDataSourceModel struct {
	ID                                  types.String `tfsdk:"id"`
	Active                              types.String `tfsdk:"active"`
	CaptivePortalWebSecDisableMinutes   types.Int64  `tfsdk:"captive_portal_web_sec_disable_minutes"`
	CompanyID                           types.String `tfsdk:"company_id"`
	CreatedBy                           types.String `tfsdk:"created_by"`
	EditedBy                            types.String `tfsdk:"edited_by"`
	EnableCaptivePortalDetection        types.Int64  `tfsdk:"enable_captive_portal_detection"`
	EnableFailOpen                      types.Int64  `tfsdk:"enable_fail_open"`
	EnableStrictEnforcementPrompt       types.Int64  `tfsdk:"enable_strict_enforcement_prompt"`
	EnableWebSecOnProxyUnreachable      types.String `tfsdk:"enable_web_sec_on_proxy_unreachable"`
	EnableWebSecOnTunnelFailure         types.String `tfsdk:"enable_web_sec_on_tunnel_failure"`
	StrictEnforcementPromptDelayMinutes types.Int64  `tfsdk:"strict_enforcement_prompt_delay_minutes"`
	StrictEnforcementPromptMessage      types.String `tfsdk:"strict_enforcement_prompt_message"`
	TunnelFailureRetryCount             types.Int64  `tfsdk:"tunnel_failure_retry_count"`
}

func (d *FailOpenPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failopen_policy"
}

func failOpenPolicySchemaAttrs(forResource bool) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique identifier of the fail open policy.",
			Optional:    true,
			Computed:    true,
		},
		"active": schema.StringAttribute{
			Description: "Whether the fail open policy is active.",
			Optional:    forResource,
			Computed:    true,
		},
		"captive_portal_web_sec_disable_minutes": schema.Int64Attribute{
			Description: "Number of minutes to disable web security for captive portal detection.",
			Optional:    forResource,
			Computed:    true,
		},
		"company_id": schema.StringAttribute{
			Description: "The company identifier.",
			Computed:    true,
		},
		"created_by": schema.StringAttribute{
			Description: "User who created the policy.",
			Computed:    true,
		},
		"edited_by": schema.StringAttribute{
			Description: "User who last edited the policy.",
			Computed:    true,
		},
		"enable_captive_portal_detection": schema.Int64Attribute{
			Description: "Enable captive portal detection.",
			Optional:    forResource,
			Computed:    true,
		},
		"enable_fail_open": schema.Int64Attribute{
			Description: "Enable fail open behavior.",
			Optional:    forResource,
			Computed:    true,
		},
		"enable_strict_enforcement_prompt": schema.Int64Attribute{
			Description: "Enable strict enforcement prompt.",
			Optional:    forResource,
			Computed:    true,
		},
		"enable_web_sec_on_proxy_unreachable": schema.StringAttribute{
			Description: "Enable web security when proxy is unreachable.",
			Optional:    forResource,
			Computed:    true,
		},
		"enable_web_sec_on_tunnel_failure": schema.StringAttribute{
			Description: "Enable web security on tunnel failure.",
			Optional:    forResource,
			Computed:    true,
		},
		"strict_enforcement_prompt_delay_minutes": schema.Int64Attribute{
			Description: "Delay in minutes for strict enforcement prompt.",
			Optional:    forResource,
			Computed:    true,
		},
		"strict_enforcement_prompt_message": schema.StringAttribute{
			Description: "Message displayed for strict enforcement prompt.",
			Optional:    forResource,
			Computed:    true,
		},
		"tunnel_failure_retry_count": schema.Int64Attribute{
			Description: "Number of retries on tunnel failure.",
			Optional:    forResource,
			Computed:    true,
		},
	}
}

func (d *FailOpenPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC fail open policy by ID, or returns the company's fail open policy.",
		Attributes:  failOpenPolicySchemaAttrs(false),
	}
}

func (d *FailOpenPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FailOpenPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FailOpenPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := d.client.Service

	var policy *failopen_policy.WebFailOpenPolicy

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching fail open policy by ID", map[string]any{"id": id})
		result, err := failopen_policy.GetFailOpenPolicyByID(ctx, service, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fail open policy: %v", err))
			return
		}
		policy = result
	} else {
		tflog.Info(ctx, "Fetching fail open policy")
		policies, err := failopen_policy.GetFailOpenPolicy(ctx, service, 1000)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fail open policies: %v", err))
			return
		}
		if len(policies) == 0 {
			resp.Diagnostics.AddError("Not Found", "No fail open policy found")
			return
		}
		policy = &policies[0]
	}

	model := flattenFailOpenPolicy(policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenFailOpenPolicy(p *failopen_policy.WebFailOpenPolicy) FailOpenPolicyDataSourceModel {
	return FailOpenPolicyDataSourceModel{
		ID:                                  types.StringValue(p.ID),
		Active:                              types.StringValue(p.Active),
		CaptivePortalWebSecDisableMinutes:   types.Int64Value(int64(p.CaptivePortalWebSecDisableMinutes)),
		CompanyID:                           types.StringValue(p.CompanyID),
		CreatedBy:                           types.StringValue(p.CreatedBy),
		EditedBy:                            types.StringValue(p.EditedBy),
		EnableCaptivePortalDetection:        types.Int64Value(int64(p.EnableCaptivePortalDetection)),
		EnableFailOpen:                      types.Int64Value(int64(p.EnableFailOpen)),
		EnableStrictEnforcementPrompt:       types.Int64Value(int64(p.EnableStrictEnforcementPrompt)),
		EnableWebSecOnProxyUnreachable:      types.StringValue(p.EnableWebSecOnProxyUnreachable),
		EnableWebSecOnTunnelFailure:         types.StringValue(p.EnableWebSecOnTunnelFailure),
		StrictEnforcementPromptDelayMinutes: types.Int64Value(int64(p.StrictEnforcementPromptDelayMins)),
		StrictEnforcementPromptMessage:      types.StringValue(p.StrictEnforcementPromptMessage),
		TunnelFailureRetryCount:             types.Int64Value(int64(p.TunnelFailureRetryCount)),
	}
}
