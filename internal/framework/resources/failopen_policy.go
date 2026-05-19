package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/failopen_policy"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ resource.Resource                = &FailOpenPolicyResource{}
	_ resource.ResourceWithConfigure   = &FailOpenPolicyResource{}
	_ resource.ResourceWithImportState = &FailOpenPolicyResource{}
)

func NewFailOpenPolicyResource() resource.Resource {
	return &FailOpenPolicyResource{}
}

type FailOpenPolicyResource struct {
	client *client.Client
}

type FailOpenPolicyResourceModel struct {
	ID                                  types.String `tfsdk:"id"`
	Active                              types.Bool   `tfsdk:"active"`
	CaptivePortalWebSecDisableMinutes   types.Int64  `tfsdk:"captive_portal_web_sec_disable_minutes"`
	EnableCaptivePortalDetection        types.Bool   `tfsdk:"enable_captive_portal_detection"`
	EnableFailOpen                      types.Bool   `tfsdk:"enable_fail_open"`
	EnableStrictEnforcementPrompt       types.Bool   `tfsdk:"enable_strict_enforcement_prompt"`
	EnableWebSecOnProxyUnreachable      types.Bool   `tfsdk:"enable_web_sec_on_proxy_unreachable"`
	EnableWebSecOnTunnelFailure         types.Bool   `tfsdk:"enable_web_sec_on_tunnel_failure"`
	StrictEnforcementPromptDelayMinutes types.Int64  `tfsdk:"strict_enforcement_prompt_delay_minutes"`
	StrictEnforcementPromptMessage      types.String `tfsdk:"strict_enforcement_prompt_message"`
	TunnelFailureRetryCount             types.Int64  `tfsdk:"tunnel_failure_retry_count"`
}

func (r *FailOpenPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_failopen_policy"
}

func (r *FailOpenPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the ZCC fail open policy. This is a singleton resource — the policy always exists and cannot be created or deleted, only updated.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the fail open policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the fail open policy is active.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"captive_portal_web_sec_disable_minutes": schema.Int64Attribute{
				Description: "Number of minutes to disable web security for captive portal detection.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"enable_captive_portal_detection": schema.BoolAttribute{
				Description: "Enable captive portal detection.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_fail_open": schema.BoolAttribute{
				Description: "Enable fail open behavior.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_strict_enforcement_prompt": schema.BoolAttribute{
				Description: "Enable strict enforcement prompt.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_web_sec_on_proxy_unreachable": schema.BoolAttribute{
				Description: "Enable web security when proxy is unreachable.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_web_sec_on_tunnel_failure": schema.BoolAttribute{
				Description: "Enable web security on tunnel failure.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"strict_enforcement_prompt_delay_minutes": schema.Int64Attribute{
				Description: "Delay in minutes for strict enforcement prompt.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"strict_enforcement_prompt_message": schema.StringAttribute{
				Description: "Message displayed for strict enforcement prompt.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tunnel_failure_retry_count": schema.Int64Attribute{
				Description: "Number of retries on tunnel failure.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FailOpenPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FailOpenPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan FailOpenPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service

	// Read the existing singleton policy first to get its ID
	policies, err := failopen_policy.GetFailOpenPolicy(ctx, service, 1000)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read existing fail open policy: %v", err))
		return
	}
	if len(policies) == 0 {
		resp.Diagnostics.AddError("Not Found", "No fail open policy exists to manage")
		return
	}

	existing := &policies[0]
	payload := r.expandFailOpenPolicy(&plan, existing)

	tflog.Info(ctx, "Updating ZCC fail open policy (singleton create)", map[string]any{"id": payload.ID})

	updated, err := failopen_policy.UpdateFailOpenPolicy(ctx, service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update fail open policy: %v", err))
		return
	}

	r.flattenFailOpenPolicy(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FailOpenPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state FailOpenPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	id := state.ID.ValueString()

	tflog.Info(ctx, "Reading ZCC fail open policy", map[string]any{"id": id})

	policy, err := failopen_policy.GetFailOpenPolicyByID(ctx, service, id)
	if err != nil {
		// For a singleton, if not found, try fetching the first available
		policies, listErr := failopen_policy.GetFailOpenPolicy(ctx, service, 1000)
		if listErr != nil || len(policies) == 0 {
			tflog.Info(ctx, "Removing fail open policy from state - no longer exists", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		policy = &policies[0]
	}

	r.flattenFailOpenPolicy(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FailOpenPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan FailOpenPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FailOpenPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service

	// Read the current state from the API to merge with plan
	existing, err := failopen_policy.GetFailOpenPolicyByID(ctx, service, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read fail open policy: %v", err))
		return
	}

	payload := r.expandFailOpenPolicy(&plan, existing)

	tflog.Info(ctx, "Updating ZCC fail open policy", map[string]any{"id": payload.ID})

	updated, err := failopen_policy.UpdateFailOpenPolicy(ctx, service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update fail open policy: %v", err))
		return
	}

	r.flattenFailOpenPolicy(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op for a singleton resource — the policy always exists.
func (r *FailOpenPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Delete called on singleton fail open policy — removing from state only (no API delete)")
}

func (r *FailOpenPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	service := r.client.Service

	// Accept any ID; for convenience also accept "failopen_policy" as a sentinel
	id := req.ID

	var policy *failopen_policy.WebFailOpenPolicy
	result, err := failopen_policy.GetFailOpenPolicyByID(ctx, service, id)
	if err != nil {
		// Fall back to fetching the first (singleton) policy
		policies, listErr := failopen_policy.GetFailOpenPolicy(ctx, service, 1000)
		if listErr != nil || len(policies) == 0 {
			resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import fail open policy: %v", err))
			return
		}
		policy = &policies[0]
	} else {
		policy = result
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(policy.ID))...)
}

func (r *FailOpenPolicyResource) expandFailOpenPolicy(plan *FailOpenPolicyResourceModel, existing *failopen_policy.WebFailOpenPolicy) *failopen_policy.WebFailOpenPolicy {
	payload := *existing

	payload.CompanyID = ""
	payload.CreatedBy = ""
	payload.EditedBy = ""

	if !plan.Active.IsNull() && !plan.Active.IsUnknown() {
		payload.Active = boolToInvertedStr(plan.Active.ValueBool())
	}
	if !plan.CaptivePortalWebSecDisableMinutes.IsNull() && !plan.CaptivePortalWebSecDisableMinutes.IsUnknown() {
		payload.CaptivePortalWebSecDisableMinutes = int(plan.CaptivePortalWebSecDisableMinutes.ValueInt64())
	}
	if !plan.EnableCaptivePortalDetection.IsNull() && !plan.EnableCaptivePortalDetection.IsUnknown() {
		payload.EnableCaptivePortalDetection = boolToInvertedInt(plan.EnableCaptivePortalDetection.ValueBool())
	}
	if !plan.EnableFailOpen.IsNull() && !plan.EnableFailOpen.IsUnknown() {
		payload.EnableFailOpen = boolToInvertedInt(plan.EnableFailOpen.ValueBool())
	}
	if !plan.EnableStrictEnforcementPrompt.IsNull() && !plan.EnableStrictEnforcementPrompt.IsUnknown() {
		payload.EnableStrictEnforcementPrompt = boolToInt(plan.EnableStrictEnforcementPrompt)
	}
	if !plan.EnableWebSecOnProxyUnreachable.IsNull() && !plan.EnableWebSecOnProxyUnreachable.IsUnknown() {
		payload.EnableWebSecOnProxyUnreachable = boolToInvertedStr(plan.EnableWebSecOnProxyUnreachable.ValueBool())
	}
	if !plan.EnableWebSecOnTunnelFailure.IsNull() && !plan.EnableWebSecOnTunnelFailure.IsUnknown() {
		payload.EnableWebSecOnTunnelFailure = boolToInvertedStr(plan.EnableWebSecOnTunnelFailure.ValueBool())
	}
	if !plan.StrictEnforcementPromptDelayMinutes.IsNull() && !plan.StrictEnforcementPromptDelayMinutes.IsUnknown() {
		payload.StrictEnforcementPromptDelayMins = int(plan.StrictEnforcementPromptDelayMinutes.ValueInt64())
	}
	if !plan.StrictEnforcementPromptMessage.IsNull() && !plan.StrictEnforcementPromptMessage.IsUnknown() {
		payload.StrictEnforcementPromptMessage = plan.StrictEnforcementPromptMessage.ValueString()
	}
	if !plan.TunnelFailureRetryCount.IsNull() && !plan.TunnelFailureRetryCount.IsUnknown() {
		payload.TunnelFailureRetryCount = int(plan.TunnelFailureRetryCount.ValueInt64())
	}

	return &payload
}

func (r *FailOpenPolicyResource) flattenFailOpenPolicy(p *failopen_policy.WebFailOpenPolicy, plan *FailOpenPolicyResourceModel) {
	plan.ID = types.StringValue(p.ID)
	plan.Active = invertedStrToBool(p.Active)
	plan.CaptivePortalWebSecDisableMinutes = types.Int64Value(int64(p.CaptivePortalWebSecDisableMinutes))
	plan.EnableCaptivePortalDetection = invertedIntToBool(p.EnableCaptivePortalDetection)
	plan.EnableFailOpen = invertedIntToBool(p.EnableFailOpen)
	plan.EnableStrictEnforcementPrompt = types.BoolValue(p.EnableStrictEnforcementPrompt == 1)
	plan.EnableWebSecOnProxyUnreachable = invertedStrToBool(p.EnableWebSecOnProxyUnreachable)
	plan.EnableWebSecOnTunnelFailure = invertedStrToBool(p.EnableWebSecOnTunnelFailure)
	plan.StrictEnforcementPromptDelayMinutes = types.Int64Value(int64(p.StrictEnforcementPromptDelayMins))
	plan.StrictEnforcementPromptMessage = types.StringValue(p.StrictEnforcementPromptMessage)
	plan.TunnelFailureRetryCount = types.Int64Value(int64(p.TunnelFailureRetryCount))
}

// The ZCC API uses inverted boolean logic: 0 = enabled/true, 1 = disabled/false.

func boolToInvertedStr(v bool) string {
	if v {
		return "0"
	}
	return "1"
}

func boolToInvertedInt(v bool) int {
	if v {
		return 0
	}
	return 1
}

func invertedStrToBool(v string) types.Bool {
	return types.BoolValue(v == "0")
}

func invertedIntToBool(v int) types.Bool {
	return types.BoolValue(v == 0)
}
