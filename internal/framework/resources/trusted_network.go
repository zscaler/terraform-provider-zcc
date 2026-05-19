package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ resource.Resource                = &TrustedNetworkResource{}
	_ resource.ResourceWithConfigure   = &TrustedNetworkResource{}
	_ resource.ResourceWithImportState = &TrustedNetworkResource{}
)

func NewTrustedNetworkResource() resource.Resource {
	return &TrustedNetworkResource{}
}

type TrustedNetworkResource struct {
	client *client.Client
}

type TrustedNetworkResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	NetworkName            types.String `tfsdk:"network_name"`
	Active                 types.Bool   `tfsdk:"active"`
	ConditionType          types.Int64  `tfsdk:"condition_type"`
	DnsSearchDomains       types.String `tfsdk:"dns_search_domains"`
	DnsServers             types.String `tfsdk:"dns_servers"`
	Guid                   types.String `tfsdk:"guid"`
	Hostnames              types.String `tfsdk:"hostnames"`
	ResolvedIpsForHostname types.String `tfsdk:"resolved_ips_for_hostname"`
	TrustedDhcpServers     types.String `tfsdk:"trusted_dhcp_servers"`
	TrustedGateways        types.String `tfsdk:"trusted_gateways"`
	TrustedSubnets         types.String `tfsdk:"trusted_subnets"`
}

func (r *TrustedNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trusted_network"
}

func (r *TrustedNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ZCC trusted networks. The ZCC API requires all criteria " +
			"fields to be present in every create/update request. Fields that are " +
			"not set default to an empty string.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the trusted network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the trusted network.",
			},
			"active": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the trusted network is active.",
			},
			"condition_type": schema.Int64Attribute{
				Required:    true,
				Description: "The condition type (0 = match all, 1 = match any).",
			},
			"dns_search_domains": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated DNS search domains.",
			},
			"dns_servers": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated DNS server addresses.",
			},
			"guid": schema.StringAttribute{
				Computed:    true,
				Description: "The GUID of the trusted network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostnames": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated hostnames.",
			},
			"resolved_ips_for_hostname": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated resolved IPs for hostname.",
			},
			"trusted_dhcp_servers": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated trusted DHCP servers.",
			},
			"trusted_gateways": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated trusted gateways.",
			},
			"trusted_subnets": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comma-separated trusted subnets.",
			},
		},
	}
}

func (r *TrustedNetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TrustedNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan TrustedNetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	payload := expandTrustedNetwork(&plan)

	tflog.Info(ctx, "Creating ZCC trusted network", map[string]any{"network_name": payload.NetworkName})

	created, _, err := trusted_network.CreateTrustedNetwork(ctx, service, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create trusted network: %v", err))
		return
	}

	tflog.Info(ctx, "Created ZCC trusted network", map[string]any{"id": created.ID})
	flattenTrustedNetwork(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrustedNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state TrustedNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	net, _, err := trusted_network.GetTrustedNetworkByID(ctx, service, state.ID.ValueString())
	if err != nil {
		if respErr, ok := err.(*errorx.ErrorResponse); ok && respErr.IsObjectNotFound() {
			tflog.Info(ctx, "Removing trusted network from state - no longer exists", map[string]any{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read trusted network: %v", err))
		return
	}

	flattenTrustedNetwork(net, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TrustedNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan TrustedNetworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	payload := expandTrustedNetwork(&plan)
	payload.ID = plan.ID.ValueString()

	if !plan.Guid.IsNull() && !plan.Guid.IsUnknown() {
		payload.Guid = plan.Guid.ValueString()
	}

	tflog.Info(ctx, "Updating ZCC trusted network", map[string]any{"id": payload.ID})

	updated, _, err := trusted_network.UpdateTrustedNetwork(ctx, service, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update trusted network: %v", err))
		return
	}

	flattenTrustedNetwork(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrustedNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state TrustedNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	tflog.Info(ctx, "Deleting ZCC trusted network", map[string]any{"id": state.ID.ValueString()})
	if _, err := trusted_network.DeleteTrustedNetwork(ctx, service, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to delete trusted network: %v", err))
	}
}

func (r *TrustedNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
	res, _, err := trusted_network.GetMultipleTrustedNetworks(ctx, service, id, "id", nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import trusted network: %v", err))
		return
	}
	if res != nil && len(res.TrustedNetworkContracts) > 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(res.TrustedNetworkContracts[0].ID))...)
		return
	}
	resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Trusted network with identifier '%s' not found", id))
}

// ---------------------------------------------------------------------------
// expand: TF model → SDK struct
// ---------------------------------------------------------------------------

func expandTrustedNetwork(plan *TrustedNetworkResourceModel) trusted_network.TrustedNetwork {
	return trusted_network.TrustedNetwork{
		NetworkName:            plan.NetworkName.ValueString(),
		Active:                 plan.Active.ValueBool(),
		ConditionType:          int(plan.ConditionType.ValueInt64()),
		DnsSearchDomains:       plan.DnsSearchDomains.ValueString(),
		DnsServers:             plan.DnsServers.ValueString(),
		Hostnames:              plan.Hostnames.ValueString(),
		ResolvedIpsForHostname: plan.ResolvedIpsForHostname.ValueString(),
		TrustedDhcpServers:     plan.TrustedDhcpServers.ValueString(),
		TrustedGateways:        plan.TrustedGateways.ValueString(),
		TrustedSubnets:         plan.TrustedSubnets.ValueString(),
	}
}

// ---------------------------------------------------------------------------
// flatten: SDK struct → TF model
// ---------------------------------------------------------------------------

func flattenTrustedNetwork(tn *trusted_network.TrustedNetwork, model *TrustedNetworkResourceModel) {
	model.ID = types.StringValue(tn.ID)
	model.NetworkName = types.StringValue(tn.NetworkName)
	model.Active = types.BoolValue(tn.Active)
	model.ConditionType = types.Int64Value(int64(tn.ConditionType))
	model.DnsSearchDomains = types.StringValue(tn.DnsSearchDomains)
	model.DnsServers = types.StringValue(tn.DnsServers)
	model.Guid = types.StringValue(tn.Guid)
	model.Hostnames = types.StringValue(tn.Hostnames)
	model.ResolvedIpsForHostname = types.StringValue(tn.ResolvedIpsForHostname)
	model.TrustedDhcpServers = types.StringValue(tn.TrustedDhcpServers)
	model.TrustedGateways = types.StringValue(tn.TrustedGateways)
	model.TrustedSubnets = types.StringValue(tn.TrustedSubnets)
}
