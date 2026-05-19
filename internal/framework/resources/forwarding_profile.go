package resources

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/forwarding_profile"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ resource.Resource                = &ForwardingProfileResource{}
	_ resource.ResourceWithConfigure   = &ForwardingProfileResource{}
	_ resource.ResourceWithImportState = &ForwardingProfileResource{}
)

func NewForwardingProfileResource() resource.Resource {
	return &ForwardingProfileResource{}
}

type ForwardingProfileResource struct {
	client *client.Client
}

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

type ForwardingProfileResourceModel struct {
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Active                      types.Bool   `tfsdk:"active"`
	ConditionType               types.Int64  `tfsdk:"condition_type"`
	DnsServers                  types.String `tfsdk:"dns_servers"`
	DnsSearchDomains            types.String `tfsdk:"dns_search_domains"`
	EnableLWFDriver             types.Bool   `tfsdk:"enable_lwf_driver"`
	Hostname                    types.String `tfsdk:"hostname"`
	ResolvedIpsForHostname      types.String `tfsdk:"resolved_ips_for_hostname"`
	TrustedSubnets              types.String `tfsdk:"trusted_subnets"`
	TrustedGateways             types.String `tfsdk:"trusted_gateways"`
	TrustedDhcpServers          types.String `tfsdk:"trusted_dhcp_servers"`
	TrustedEgressIps            types.String `tfsdk:"trusted_egress_ips"`
	PredefinedTrustedNetworks   types.Bool   `tfsdk:"predefined_trusted_networks"`
	PredefinedTnAll             types.Bool   `tfsdk:"predefined_tn_all"`
	EnableSplitVpnTN            types.Bool   `tfsdk:"enable_split_vpn_tn"`
	EnableUnifiedTunnel         types.Bool   `tfsdk:"enable_unified_tunnel"`
	EvaluateTrustedNetwork      types.Bool   `tfsdk:"evaluate_trusted_network"`
	SkipTrustedCriteriaMatch    types.Bool   `tfsdk:"skip_trusted_criteria_match"`
	EnableAllDefaultAdaptersTN  types.Bool   `tfsdk:"enable_all_default_adapters_tn"`
	TrustedNetworkIds           types.List   `tfsdk:"trusted_network_ids"`
	TrustedNetworks             types.List   `tfsdk:"trusted_networks"`
	TrustedNetworkIdsSelected   types.List   `tfsdk:"trusted_network_ids_selected"`
	ForwardingProfileActions    types.List   `tfsdk:"forwarding_profile_actions"`
	ForwardingProfileZpaActions types.List   `tfsdk:"forwarding_profile_zpa_actions"`
	UnifiedTunnel               types.List   `tfsdk:"unified_tunnel"`
}

type ForwardingProfileActionModel struct {
	NetworkType                        types.Int64  `tfsdk:"network_type"`
	ActionType                         types.Int64  `tfsdk:"action_type"`
	PrimaryTransport                   types.Int64  `tfsdk:"primary_transport"`
	DTLSTimeout                        types.Int64  `tfsdk:"dtls_timeout"`
	UDPTimeout                         types.Int64  `tfsdk:"udp_timeout"`
	TLSTimeout                         types.Int64  `tfsdk:"tls_timeout"`
	MtuForZadapter                     types.Int64  `tfsdk:"mtu_for_zadapter"`
	Tunnel2FallbackType                types.Int64  `tfsdk:"tunnel2_fallback_type"`
	ZenProbeInterval                   types.Int64  `tfsdk:"zen_probe_interval"`
	ZenProbeSampleSize                 types.Int64  `tfsdk:"zen_probe_sample_size"`
	ZenThresholdLimit                  types.Int64  `tfsdk:"zen_threshold_limit"`
	LbsProbeInterval                   types.Int64  `tfsdk:"lbs_probe_interval"`
	LbsProbeSampleSize                 types.Int64  `tfsdk:"lbs_probe_sample_size"`
	LbsThresholdLimit                  types.Int64  `tfsdk:"lbs_threshold_limit"`
	CustomPac                          types.String `tfsdk:"custom_pac"`
	SystemProxy                        types.Bool   `tfsdk:"system_proxy"`
	EnablePacketTunnel                 types.Bool   `tfsdk:"enable_packet_tunnel"`
	BlockUnreachableDomainsTraffic     types.Bool   `tfsdk:"block_unreachable_domains_traffic"`
	AllowTLSFallback                   types.Bool   `tfsdk:"allow_tls_fallback"`
	SendAllDNSToTrustedServer          types.Bool   `tfsdk:"send_all_dns_to_trusted_server"`
	DropIpv6Traffic                    types.Bool   `tfsdk:"drop_ipv6_traffic"`
	RedirectWebTraffic                 types.Bool   `tfsdk:"redirect_web_traffic"`
	DropIpv6IncludeTrafficInT2         types.Bool   `tfsdk:"drop_ipv6_include_traffic_in_t2"`
	UseTunnel2ForProxiedWebTraffic     types.Bool   `tfsdk:"use_tunnel2_for_proxied_web_traffic"`
	UseTunnel2ForUnencryptedWebTraffic types.Bool   `tfsdk:"use_tunnel2_for_unencrypted_web_traffic"`
	PathMtuDiscovery                   types.Bool   `tfsdk:"path_mtu_discovery"`
	LatencyBasedZenEnablement          types.Bool   `tfsdk:"latency_based_zen_enablement"`
	LatencyBasedServerEnablement       types.Bool   `tfsdk:"latency_based_server_enablement"`
	LatencyBasedServerMTEnablement     types.Bool   `tfsdk:"latency_based_server_mt_enablement"`
	DropIpv6TrafficInIpv6Network       types.Bool   `tfsdk:"drop_ipv6_traffic_in_ipv6_network"`
	OptimiseForUnstableConnections     types.Bool   `tfsdk:"optimise_for_unstable_connections"`
	IsSameAsOnTrustedNetwork           types.Bool   `tfsdk:"is_same_as_on_trusted_network"`
	SystemProxyData                    types.Object `tfsdk:"system_proxy_data"`
}

type SystemProxyDataModel struct {
	ProxyAction             types.Int64  `tfsdk:"proxy_action"`
	PacURL                  types.String `tfsdk:"pac_url"`
	ProxyServerAddress      types.String `tfsdk:"proxy_server_address"`
	ProxyServerPort         types.String `tfsdk:"proxy_server_port"`
	PacDataPath             types.String `tfsdk:"pac_data_path"`
	EnableAutoDetect        types.Bool   `tfsdk:"enable_auto_detect"`
	EnablePAC               types.Bool   `tfsdk:"enable_pac"`
	EnableProxyServer       types.Bool   `tfsdk:"enable_proxy_server"`
	BypassProxyForPrivateIP types.Bool   `tfsdk:"bypass_proxy_for_private_ip"`
	PerformGPUpdate         types.Bool   `tfsdk:"perform_gp_update"`
}

type ForwardingProfileZpaActionModel struct {
	NetworkType                    types.Int64  `tfsdk:"network_type"`
	ActionType                     types.Int64  `tfsdk:"action_type"`
	PrimaryTransport               types.Int64  `tfsdk:"primary_transport"`
	DTLSTimeout                    types.Int64  `tfsdk:"dtls_timeout"`
	TLSTimeout                     types.Int64  `tfsdk:"tls_timeout"`
	MtuForZadapter                 types.Int64  `tfsdk:"mtu_for_zadapter"`
	LbsProbeInterval               types.Int64  `tfsdk:"lbs_probe_interval"`
	LbsProbeSampleSize             types.Int64  `tfsdk:"lbs_probe_sample_size"`
	LbsThresholdLimit              types.Int64  `tfsdk:"lbs_threshold_limit"`
	SendTrustedNetworkResultToZpa  types.Bool   `tfsdk:"send_trusted_network_result_to_zpa"`
	LatencyBasedServerEnablement   types.Bool   `tfsdk:"latency_based_server_enablement"`
	LatencyBasedServerMTEnablement types.Bool   `tfsdk:"latency_based_server_mt_enablement"`
	IsSameAsOnTrustedNetwork       types.Bool   `tfsdk:"is_same_as_on_trusted_network"`
	PartnerInfo                    types.Object `tfsdk:"partner_info"`
}

type PartnerInfoModel struct {
	PrimaryTransport types.Int64 `tfsdk:"primary_transport"`
	MtuForZadapter   types.Int64 `tfsdk:"mtu_for_zadapter"`
	AllowTlsFallback types.Bool  `tfsdk:"allow_tls_fallback"`
}

type UnifiedTunnelModel struct {
	NetworkType                    types.Int64  `tfsdk:"network_type"`
	ActionTypeZIA                  types.Int64  `tfsdk:"action_type_zia"`
	ActionTypeZPA                  types.Int64  `tfsdk:"action_type_zpa"`
	PrimaryTransport               types.Int64  `tfsdk:"primary_transport"`
	DTLSTimeout                    types.Int64  `tfsdk:"dtls_timeout"`
	TLSTimeout                     types.Int64  `tfsdk:"tls_timeout"`
	MtuForZadapter                 types.Int64  `tfsdk:"mtu_for_zadapter"`
	Tunnel2FallbackType            types.Int64  `tfsdk:"tunnel2_fallback_type"`
	AllowTLSFallback               types.Bool   `tfsdk:"allow_tls_fallback"`
	PathMtuDiscovery               types.Bool   `tfsdk:"path_mtu_discovery"`
	OptimiseForUnstableConnections types.Bool   `tfsdk:"optimise_for_unstable_connections"`
	RedirectWebTraffic             types.Bool   `tfsdk:"redirect_web_traffic"`
	DropIpv6Traffic                types.Bool   `tfsdk:"drop_ipv6_traffic"`
	DropIpv6TrafficInIpv6Network   types.Bool   `tfsdk:"drop_ipv6_traffic_in_ipv6_network"`
	BlockUnreachableDomainsTraffic types.Bool   `tfsdk:"block_unreachable_domains_traffic"`
	DropIpv6IncludeTrafficInT2     types.Bool   `tfsdk:"drop_ipv6_include_traffic_in_t2"`
	SendAllDNSToTrustedServer      types.Bool   `tfsdk:"send_all_dns_to_trusted_server"`
	SameAsOnTrusted                types.Bool   `tfsdk:"same_as_on_trusted"`
	SystemProxyData                types.Object `tfsdk:"system_proxy_data"`
}

// ---------------------------------------------------------------------------
// Boolean utility functions
// ---------------------------------------------------------------------------

func boolToInt(b types.Bool) int {
	if b.ValueBool() {
		return 1
	}
	return 0
}

func intToBool(i int) types.Bool {
	return types.BoolValue(i != 0)
}

func string01ToBool(s string) types.Bool {
	return types.BoolValue(s == "1")
}

func boolFromAttr(v attr.Value) bool {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return false
	}
	if b, ok := v.(types.Bool); ok {
		return b.ValueBool()
	}
	return false
}

func intFromAttr(v attr.Value) int {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return 0
	}
	if i, ok := v.(types.Int64); ok {
		return int(i.ValueInt64())
	}
	return 0
}

func stringFromAttr(v attr.Value) string {
	if v == nil || v.IsNull() || v.IsUnknown() {
		return ""
	}
	if s, ok := v.(types.String); ok {
		return s.ValueString()
	}
	return ""
}

// ---------------------------------------------------------------------------
// Resource interface
// ---------------------------------------------------------------------------

func (r *ForwardingProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forwarding_profile"
}

func systemProxyDataBlockSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"proxy_action":                schema.Int64Attribute{Optional: true, Computed: true},
			"enable_auto_detect":          schema.BoolAttribute{Optional: true, Computed: true},
			"enable_pac":                  schema.BoolAttribute{Optional: true, Computed: true},
			"pac_url":                     schema.StringAttribute{Optional: true, Computed: true},
			"enable_proxy_server":         schema.BoolAttribute{Optional: true, Computed: true},
			"proxy_server_address":        schema.StringAttribute{Optional: true, Computed: true},
			"proxy_server_port":           schema.StringAttribute{Optional: true, Computed: true},
			"bypass_proxy_for_private_ip": schema.BoolAttribute{Optional: true, Computed: true},
			"perform_gp_update":           schema.BoolAttribute{Optional: true, Computed: true},
			"pac_data_path":               schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}

func (r *ForwardingProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages ZCC forwarding profiles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the forwarding profile.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the forwarding profile.",
				Required:    true,
			},
			"active":                         schema.BoolAttribute{Optional: true, Computed: true},
			"condition_type":                 schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"dns_servers":                    schema.StringAttribute{Optional: true, Computed: true},
			"dns_search_domains":             schema.StringAttribute{Optional: true, Computed: true},
			"enable_lwf_driver":              schema.BoolAttribute{Optional: true, Computed: true},
			"hostname":                       schema.StringAttribute{Optional: true, Computed: true},
			"resolved_ips_for_hostname":      schema.StringAttribute{Optional: true, Computed: true},
			"trusted_subnets":                schema.StringAttribute{Optional: true, Computed: true},
			"trusted_gateways":               schema.StringAttribute{Optional: true, Computed: true},
			"trusted_dhcp_servers":           schema.StringAttribute{Optional: true, Computed: true},
			"trusted_egress_ips":             schema.StringAttribute{Optional: true, Computed: true},
			"predefined_trusted_networks":    schema.BoolAttribute{Optional: true, Computed: true},
			"predefined_tn_all":              schema.BoolAttribute{Optional: true, Computed: true},
			"enable_split_vpn_tn":            schema.BoolAttribute{Optional: true, Computed: true},
			"enable_unified_tunnel":          schema.BoolAttribute{Optional: true, Computed: true},
			"evaluate_trusted_network":       schema.BoolAttribute{Optional: true, Computed: true},
			"skip_trusted_criteria_match":    schema.BoolAttribute{Optional: true, Computed: true},
			"enable_all_default_adapters_tn": schema.BoolAttribute{Optional: true, Computed: true},
			"trusted_network_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
			},
			"trusted_networks": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"trusted_network_ids_selected": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Computed:    true,
			},
		},

		Blocks: map[string]schema.Block{
			"forwarding_profile_actions": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"network_type":                            schema.Int64Attribute{Optional: true, Computed: true},
						"action_type":                             schema.Int64Attribute{Optional: true, Computed: true},
						"primary_transport":                       schema.Int64Attribute{Optional: true, Computed: true},
						"dtls_timeout":                            schema.Int64Attribute{Optional: true, Computed: true},
						"udp_timeout":                             schema.Int64Attribute{Optional: true, Computed: true},
						"tls_timeout":                             schema.Int64Attribute{Optional: true, Computed: true},
						"mtu_for_zadapter":                        schema.Int64Attribute{Optional: true, Computed: true},
						"tunnel2_fallback_type":                   schema.Int64Attribute{Optional: true, Computed: true},
						"zen_probe_interval":                      schema.Int64Attribute{Optional: true, Computed: true},
						"zen_probe_sample_size":                   schema.Int64Attribute{Optional: true, Computed: true},
						"zen_threshold_limit":                     schema.Int64Attribute{Optional: true, Computed: true},
						"lbs_probe_interval":                      schema.Int64Attribute{Optional: true, Computed: true},
						"lbs_probe_sample_size":                   schema.Int64Attribute{Optional: true, Computed: true},
						"lbs_threshold_limit":                     schema.Int64Attribute{Optional: true, Computed: true},
						"custom_pac":                              schema.StringAttribute{Optional: true, Computed: true},
						"system_proxy":                            schema.BoolAttribute{Optional: true, Computed: true},
						"enable_packet_tunnel":                    schema.BoolAttribute{Optional: true, Computed: true},
						"block_unreachable_domains_traffic":       schema.BoolAttribute{Optional: true, Computed: true},
						"allow_tls_fallback":                      schema.BoolAttribute{Optional: true, Computed: true},
						"send_all_dns_to_trusted_server":          schema.BoolAttribute{Optional: true, Computed: true},
						"drop_ipv6_traffic":                       schema.BoolAttribute{Optional: true, Computed: true},
						"redirect_web_traffic":                    schema.BoolAttribute{Optional: true, Computed: true},
						"drop_ipv6_include_traffic_in_t2":         schema.BoolAttribute{Optional: true, Computed: true},
						"use_tunnel2_for_proxied_web_traffic":     schema.BoolAttribute{Optional: true, Computed: true},
						"use_tunnel2_for_unencrypted_web_traffic": schema.BoolAttribute{Optional: true, Computed: true},
						"path_mtu_discovery":                      schema.BoolAttribute{Optional: true, Computed: true},
						"latency_based_zen_enablement":            schema.BoolAttribute{Optional: true, Computed: true},
						"latency_based_server_enablement":         schema.BoolAttribute{Optional: true, Computed: true},
						"latency_based_server_mt_enablement":      schema.BoolAttribute{Optional: true, Computed: true},
						"drop_ipv6_traffic_in_ipv6_network":       schema.BoolAttribute{Optional: true, Computed: true},
						"optimise_for_unstable_connections":       schema.BoolAttribute{Optional: true, Computed: true},
						"is_same_as_on_trusted_network":           schema.BoolAttribute{Optional: true, Computed: true},
					},
					Blocks: map[string]schema.Block{
						"system_proxy_data": systemProxyDataBlockSchema(),
					},
				},
			},

			"forwarding_profile_zpa_actions": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"network_type":                       schema.Int64Attribute{Optional: true, Computed: true},
						"action_type":                        schema.Int64Attribute{Optional: true, Computed: true},
						"primary_transport":                  schema.Int64Attribute{Optional: true, Computed: true},
						"dtls_timeout":                       schema.Int64Attribute{Optional: true, Computed: true},
						"tls_timeout":                        schema.Int64Attribute{Optional: true, Computed: true},
						"mtu_for_zadapter":                   schema.Int64Attribute{Optional: true, Computed: true},
						"lbs_probe_interval":                 schema.Int64Attribute{Optional: true, Computed: true},
						"lbs_probe_sample_size":              schema.Int64Attribute{Optional: true, Computed: true},
						"lbs_threshold_limit":                schema.Int64Attribute{Optional: true, Computed: true},
						"send_trusted_network_result_to_zpa": schema.BoolAttribute{Optional: true, Computed: true},
						"latency_based_server_enablement":    schema.BoolAttribute{Optional: true, Computed: true},
						"latency_based_server_mt_enablement": schema.BoolAttribute{Optional: true, Computed: true},
						"is_same_as_on_trusted_network":      schema.BoolAttribute{Optional: true, Computed: true},
					},
					Blocks: map[string]schema.Block{
						"partner_info": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"primary_transport":  schema.Int64Attribute{Optional: true, Computed: true},
								"mtu_for_zadapter":   schema.Int64Attribute{Optional: true, Computed: true},
								"allow_tls_fallback": schema.BoolAttribute{Optional: true, Computed: true},
							},
						},
					},
				},
			},

			"unified_tunnel": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"network_type":                      schema.Int64Attribute{Optional: true, Computed: true},
						"action_type_zia":                   schema.Int64Attribute{Optional: true, Computed: true},
						"action_type_zpa":                   schema.Int64Attribute{Optional: true, Computed: true},
						"primary_transport":                 schema.Int64Attribute{Optional: true, Computed: true},
						"dtls_timeout":                      schema.Int64Attribute{Optional: true, Computed: true},
						"tls_timeout":                       schema.Int64Attribute{Optional: true, Computed: true},
						"mtu_for_zadapter":                  schema.Int64Attribute{Optional: true, Computed: true},
						"tunnel2_fallback_type":             schema.Int64Attribute{Optional: true, Computed: true},
						"allow_tls_fallback":                schema.BoolAttribute{Optional: true, Computed: true},
						"path_mtu_discovery":                schema.BoolAttribute{Optional: true, Computed: true},
						"optimise_for_unstable_connections": schema.BoolAttribute{Optional: true, Computed: true},
						"redirect_web_traffic":              schema.BoolAttribute{Optional: true, Computed: true},
						"drop_ipv6_traffic":                 schema.BoolAttribute{Optional: true, Computed: true},
						"drop_ipv6_traffic_in_ipv6_network": schema.BoolAttribute{Optional: true, Computed: true},
						"block_unreachable_domains_traffic": schema.BoolAttribute{Optional: true, Computed: true},
						"drop_ipv6_include_traffic_in_t2":   schema.BoolAttribute{Optional: true, Computed: true},
						"send_all_dns_to_trusted_server":    schema.BoolAttribute{Optional: true, Computed: true},
						"same_as_on_trusted":                schema.BoolAttribute{Optional: true, Computed: true},
					},
					Blocks: map[string]schema.Block{
						"system_proxy_data": systemProxyDataBlockSchema(),
					},
				},
			},
		},
	}
}

func (r *ForwardingProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ForwardingProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan ForwardingProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	payload := expandForwardingProfileRequest(ctx, &plan, "-1")

	tflog.Info(ctx, "Creating ZCC forwarding profile", map[string]any{"name": payload.Name})

	createResp, err := forwarding_profile.CreateForwardingProfile(ctx, service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to create forwarding profile: %v", err))
		return
	}

	tflog.Info(ctx, "Forwarding profile POST response", map[string]any{"id": createResp.ID, "success": createResp.Success})

	profile, err := findProfileByID(ctx, service, createResp.ID)
	if err != nil {
		profile, err = findProfileByName(ctx, service, payload.Name)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Forwarding profile created (id=%d), but failed to re-read: %v", createResp.ID, err))
			return
		}
	}

	tflog.Info(ctx, "Created ZCC forwarding profile", map[string]any{"id": profile.ID, "name": profile.Name})

	flattenForwardingProfile(ctx, profile, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ForwardingProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state ForwardingProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idStr := state.ID.ValueString()
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Forwarding profile ID must be numeric: %v", err))
		return
	}

	service := r.client.Service

	profiles, err := forwarding_profile.GetForwardingProfileByCompanyID(ctx, service, "", nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read forwarding profile: %v", err))
		return
	}

	var profile *forwarding_profile.ForwardingProfile
	for i := range profiles {
		if int(profiles[i].ID) == idInt {
			profile = &profiles[i]
			break
		}
	}

	if profile == nil {
		tflog.Info(ctx, "Removing forwarding profile from state - no longer exists", map[string]any{"id": idStr})
		resp.State.RemoveResource(ctx)
		return
	}

	flattenForwardingProfile(ctx, profile, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ForwardingProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var plan ForwardingProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := r.client.Service
	existingID := plan.ID.ValueString()
	payload := expandForwardingProfileRequest(ctx, &plan, existingID)

	tflog.Info(ctx, "Updating ZCC forwarding profile", map[string]any{"id": existingID})

	updateResp, err := forwarding_profile.CreateForwardingProfile(ctx, service, payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update forwarding profile: %v", err))
		return
	}

	readID := updateResp.ID
	if readID == 0 {
		readID, _ = strconv.Atoi(existingID)
	}

	profile, err := findProfileByID(ctx, service, readID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Forwarding profile updated, but failed to re-read by ID '%d': %v", readID, err))
		return
	}

	flattenForwardingProfile(ctx, profile, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ForwardingProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before managing resources.")
		return
	}

	var state ForwardingProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	idStr := state.ID.ValueString()
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Forwarding profile ID must be numeric: %v", err))
		return
	}

	service := r.client.Service
	tflog.Info(ctx, "Deleting ZCC forwarding profile", map[string]any{"id": idStr})

	if _, err := forwarding_profile.DeleteForwardingProfile(ctx, service, idInt); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to delete forwarding profile: %v", err))
		return
	}
}

func (r *ForwardingProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Provider", "The provider must be configured before importing resources.")
		return
	}

	id := req.ID
	if _, err := strconv.Atoi(id); err != nil {
		resp.Diagnostics.AddError("Import Error", "Forwarding profile import requires a numeric ID")
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helpers to re-read after create/update
// ---------------------------------------------------------------------------

func findProfileByName(ctx context.Context, service *zscaler.Service, name string) (*forwarding_profile.ForwardingProfile, error) {
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			tflog.Info(ctx, "Retrying profile lookup by name", map[string]any{"name": name, "attempt": attempt + 1})
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		profiles, err := forwarding_profile.GetForwardingProfileByCompanyID(ctx, service, "", nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list forwarding profiles: %w", err)
		}
		for i := range profiles {
			if profiles[i].Name == name {
				return &profiles[i], nil
			}
		}
	}
	return nil, fmt.Errorf("forwarding profile with name '%s' not found after retries", name)
}

func findProfileByID(ctx context.Context, service *zscaler.Service, id int) (*forwarding_profile.ForwardingProfile, error) {
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			tflog.Info(ctx, "Retrying profile lookup by ID", map[string]any{"id": id, "attempt": attempt + 1})
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		profiles, err := forwarding_profile.GetForwardingProfileByCompanyID(ctx, service, "", nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list forwarding profiles: %w", err)
		}
		for i := range profiles {
			if int(profiles[i].ID) == id {
				return &profiles[i], nil
			}
		}
	}
	return nil, fmt.Errorf("forwarding profile with ID '%d' not found after retries", id)
}

// ---------------------------------------------------------------------------
// Expand (plan → SDK struct)
// ---------------------------------------------------------------------------

func expandForwardingProfileRequest(ctx context.Context, plan *ForwardingProfileResourceModel, id string) *forwarding_profile.ForwardingProfileRequest {
	req := &forwarding_profile.ForwardingProfileRequest{
		ID:                         id,
		Active:                     boolToInt(plan.Active),
		Name:                       plan.Name.ValueString(),
		ConditionType:              int(plan.ConditionType.ValueInt64()),
		DnsServers:                 plan.DnsServers.ValueString(),
		DnsSearchDomains:           plan.DnsSearchDomains.ValueString(),
		EnableLWFDriver:            boolToInt(plan.EnableLWFDriver),
		Hostname:                   plan.Hostname.ValueString(),
		ResolvedIpsForHostname:     plan.ResolvedIpsForHostname.ValueString(),
		TrustedSubnets:             plan.TrustedSubnets.ValueString(),
		TrustedGateways:            plan.TrustedGateways.ValueString(),
		TrustedDhcpServers:         plan.TrustedDhcpServers.ValueString(),
		TrustedEgressIps:           plan.TrustedEgressIps.ValueString(),
		EnableUnifiedTunnel:        boolToInt(plan.EnableUnifiedTunnel),
		EnableAllDefaultAdaptersTN: boolToInt(plan.EnableAllDefaultAdaptersTN),
		EvaluateTrustedNetwork:     boolToInt(plan.EvaluateTrustedNetwork),
		EnableSplitVpnTN:           boolToInt(plan.EnableSplitVpnTN),
		SkipTrustedCriteriaMatch:   boolToInt(plan.SkipTrustedCriteriaMatch),
		PredefinedTrustedNetworks:  plan.PredefinedTrustedNetworks.ValueBool(),
		PredefinedTnAll:            plan.PredefinedTnAll.ValueBool(),
	}

	if !plan.TrustedNetworkIds.IsNull() {
		elems := plan.TrustedNetworkIds.Elements()
		req.TrustedNetworkIds = make([]int, 0, len(elems))
		for _, e := range elems {
			if v, ok := e.(types.Int64); ok {
				req.TrustedNetworkIds = append(req.TrustedNetworkIds, int(v.ValueInt64()))
			}
		}
	} else {
		req.TrustedNetworkIds = []int{}
	}

	if !plan.ForwardingProfileActions.IsNull() && !plan.ForwardingProfileActions.IsUnknown() {
		var actions []ForwardingProfileActionModel
		plan.ForwardingProfileActions.ElementsAs(ctx, &actions, false)
		req.ForwardingProfileActions = expandProfileActionRequests(ctx, actions)
	}

	if !plan.ForwardingProfileZpaActions.IsNull() && !plan.ForwardingProfileZpaActions.IsUnknown() {
		var zpaActions []ForwardingProfileZpaActionModel
		plan.ForwardingProfileZpaActions.ElementsAs(ctx, &zpaActions, false)
		req.ForwardingProfileZpaActions = expandZpaActionRequests(ctx, zpaActions)
	}

	if !plan.UnifiedTunnel.IsNull() && !plan.UnifiedTunnel.IsUnknown() {
		var tunnels []UnifiedTunnelModel
		plan.UnifiedTunnel.ElementsAs(ctx, &tunnels, false)
		req.UnifiedTunnel = expandUnifiedTunnelRequests(ctx, tunnels)
	}

	return req
}

func expandSystemProxyDataRequest(_ context.Context, obj types.Object) forwarding_profile.SystemProxyDataRequest {
	if obj.IsNull() || obj.IsUnknown() {
		return forwarding_profile.SystemProxyDataRequest{}
	}
	a := obj.Attributes()
	boolAttrToInt := func(v attr.Value) int {
		if boolFromAttr(v) {
			return 1
		}
		return 0
	}
	return forwarding_profile.SystemProxyDataRequest{
		ProxyAction:             intFromAttr(a["proxy_action"]),
		EnableAutoDetect:        boolAttrToInt(a["enable_auto_detect"]),
		EnablePAC:               boolAttrToInt(a["enable_pac"]),
		PacURL:                  stringFromAttr(a["pac_url"]),
		EnableProxyServer:       boolAttrToInt(a["enable_proxy_server"]),
		ProxyServerAddress:      stringFromAttr(a["proxy_server_address"]),
		ProxyServerPort:         stringFromAttr(a["proxy_server_port"]),
		BypassProxyForPrivateIP: boolAttrToInt(a["bypass_proxy_for_private_ip"]),
		PerformGPUpdate:         boolAttrToInt(a["perform_gp_update"]),
		PacDataPath:             stringFromAttr(a["pac_data_path"]),
	}
}

func expandProfileActionRequests(ctx context.Context, actions []ForwardingProfileActionModel) []forwarding_profile.ForwardingProfileActionRequest {
	result := make([]forwarding_profile.ForwardingProfileActionRequest, 0, len(actions))
	for _, a := range actions {
		act := forwarding_profile.ForwardingProfileActionRequest{
			NetworkType:                        int(a.NetworkType.ValueInt64()),
			ActionType:                         int(a.ActionType.ValueInt64()),
			PrimaryTransport:                   int(a.PrimaryTransport.ValueInt64()),
			DTLSTimeout:                        int(a.DTLSTimeout.ValueInt64()),
			UDPTimeout:                         int(a.UDPTimeout.ValueInt64()),
			TLSTimeout:                         int(a.TLSTimeout.ValueInt64()),
			MtuForZadapter:                     strconv.Itoa(int(a.MtuForZadapter.ValueInt64())),
			Tunnel2FallbackType:                int(a.Tunnel2FallbackType.ValueInt64()),
			CustomPac:                          a.CustomPac.ValueString(),
			EnablePacketTunnel:                 boolToInt(a.EnablePacketTunnel),
			BlockUnreachableDomainsTraffic:     strconv.Itoa(boolToInt(a.BlockUnreachableDomainsTraffic)),
			DropIpv6Traffic:                    boolToInt(a.DropIpv6Traffic),
			AllowTLSFallback:                   boolToInt(a.AllowTLSFallback),
			PathMtuDiscovery:                   boolToInt(a.PathMtuDiscovery),
			UseTunnel2ForProxiedWebTraffic:     boolToInt(a.UseTunnel2ForProxiedWebTraffic),
			UseTunnel2ForUnencryptedWebTraffic: boolToInt(a.UseTunnel2ForUnencryptedWebTraffic),
			RedirectWebTraffic:                 boolToInt(a.RedirectWebTraffic),
			DropIpv6IncludeTrafficInT2:         boolToInt(a.DropIpv6IncludeTrafficInT2),
			LatencyBasedZenEnablement:          boolToInt(a.LatencyBasedZenEnablement),
			ZenProbeInterval:                   int(a.ZenProbeInterval.ValueInt64()),
			ZenProbeSampleSize:                 int(a.ZenProbeSampleSize.ValueInt64()),
			ZenThresholdLimit:                  int(a.ZenThresholdLimit.ValueInt64()),
			LatencyBasedServerEnablement:       boolToInt(a.LatencyBasedServerEnablement),
			LbsProbeInterval:                   int(a.LbsProbeInterval.ValueInt64()),
			LbsProbeSampleSize:                 int(a.LbsProbeSampleSize.ValueInt64()),
			LbsThresholdLimit:                  int(a.LbsThresholdLimit.ValueInt64()),
			LatencyBasedServerMTEnablement:     boolToInt(a.LatencyBasedServerMTEnablement),
			SystemProxyData:                    expandSystemProxyDataRequest(ctx, a.SystemProxyData),
		}

		if !a.IsSameAsOnTrustedNetwork.IsNull() && !a.IsSameAsOnTrustedNetwork.IsUnknown() {
			v := a.IsSameAsOnTrustedNetwork.ValueBool()
			act.IsSameAsOnTrustedNetwork = &v
		}

		result = append(result, act)
	}
	return result
}

func expandZpaActionRequests(_ context.Context, actions []ForwardingProfileZpaActionModel) []forwarding_profile.ForwardingProfileZpaActionRequest {
	result := make([]forwarding_profile.ForwardingProfileZpaActionRequest, 0, len(actions))
	for _, a := range actions {
		act := forwarding_profile.ForwardingProfileZpaActionRequest{
			NetworkType:                    int(a.NetworkType.ValueInt64()),
			ActionType:                     int(a.ActionType.ValueInt64()),
			PrimaryTransport:               int(a.PrimaryTransport.ValueInt64()),
			DTLSTimeout:                    int(a.DTLSTimeout.ValueInt64()),
			TLSTimeout:                     int(a.TLSTimeout.ValueInt64()),
			MtuForZadapter:                 strconv.Itoa(int(a.MtuForZadapter.ValueInt64())),
			LatencyBasedServerEnablement:   boolToInt(a.LatencyBasedServerEnablement),
			LbsProbeInterval:               int(a.LbsProbeInterval.ValueInt64()),
			LbsProbeSampleSize:             int(a.LbsProbeSampleSize.ValueInt64()),
			LbsThresholdLimit:              int(a.LbsThresholdLimit.ValueInt64()),
			LatencyBasedServerMTEnablement: boolToInt(a.LatencyBasedServerMTEnablement),
			SendTrustedNetworkResultToZpa:  boolToInt(a.SendTrustedNetworkResultToZpa),
		}

		if !a.PartnerInfo.IsNull() && !a.PartnerInfo.IsUnknown() {
			pa := a.PartnerInfo.Attributes()
			allowFallback := 0
			if boolFromAttr(pa["allow_tls_fallback"]) {
				allowFallback = 1
			}
			act.PartnerInfo = forwarding_profile.PartnerInfoRequest{
				PrimaryTransport: intFromAttr(pa["primary_transport"]),
				AllowTlsFallback: allowFallback,
				MtuForZadapter:   intFromAttr(pa["mtu_for_zadapter"]),
			}
		}

		if !a.IsSameAsOnTrustedNetwork.IsNull() && !a.IsSameAsOnTrustedNetwork.IsUnknown() {
			v := a.IsSameAsOnTrustedNetwork.ValueBool()
			act.IsSameAsOnTrustedNetwork = &v
		}

		result = append(result, act)
	}
	return result
}

func expandUnifiedTunnelRequests(ctx context.Context, tunnels []UnifiedTunnelModel) []forwarding_profile.UnifiedTunnelRequest {
	result := make([]forwarding_profile.UnifiedTunnelRequest, 0, len(tunnels))
	for _, t := range tunnels {
		result = append(result, forwarding_profile.UnifiedTunnelRequest{
			NetworkType:                    int(t.NetworkType.ValueInt64()),
			ActionTypeZIA:                  int(t.ActionTypeZIA.ValueInt64()),
			ActionTypeZPA:                  int(t.ActionTypeZPA.ValueInt64()),
			PrimaryTransport:               int(t.PrimaryTransport.ValueInt64()),
			DTLSTimeout:                    int(t.DTLSTimeout.ValueInt64()),
			TLSTimeout:                     int(t.TLSTimeout.ValueInt64()),
			MtuForZadapter:                 strconv.Itoa(int(t.MtuForZadapter.ValueInt64())),
			AllowTLSFallback:               boolToInt(t.AllowTLSFallback),
			PathMtuDiscovery:               boolToInt(t.PathMtuDiscovery),
			Tunnel2FallbackType:            int(t.Tunnel2FallbackType.ValueInt64()),
			RedirectWebTraffic:             boolToInt(t.RedirectWebTraffic),
			DropIpv6Traffic:                boolToInt(t.DropIpv6Traffic),
			DropIpv6IncludeTrafficInT2:     boolToInt(t.DropIpv6IncludeTrafficInT2),
			BlockUnreachableDomainsTraffic: strconv.Itoa(boolToInt(t.BlockUnreachableDomainsTraffic)),
			SameAsOnTrusted:                boolToInt(t.SameAsOnTrusted),
			SystemProxyData:                expandSystemProxyDataRequest(ctx, t.SystemProxyData),
		})
	}
	return result
}

// ---------------------------------------------------------------------------
// Flatten (SDK struct → model)
// ---------------------------------------------------------------------------

func flattenForwardingProfile(ctx context.Context, p *forwarding_profile.ForwardingProfile, model *ForwardingProfileResourceModel) {
	// Snapshot plan/state values for fields that the API silently drops or
	// returns under a different key — we restore them below to avoid the
	// "provider produced inconsistent result after apply" diagnostic.
	priorEvaluateTrustedNetwork := model.EvaluateTrustedNetwork

	model.ID = types.StringValue(strconv.Itoa(int(p.ID)))
	model.Name = types.StringValue(p.Name)
	model.Active = string01ToBool(p.Active)
	model.ConditionType = types.Int64Value(int64(p.ConditionType))
	model.DnsServers = types.StringValue(p.DnsServers)
	model.DnsSearchDomains = types.StringValue(p.DnsSearchDomains)
	model.EnableLWFDriver = string01ToBool(p.EnableLWFDriver)
	model.Hostname = types.StringValue(p.Hostname)
	model.ResolvedIpsForHostname = types.StringValue(p.ResolvedIpsForHostname)
	model.TrustedSubnets = types.StringValue(p.TrustedSubnets)
	model.TrustedGateways = types.StringValue(p.TrustedGateways)
	model.TrustedDhcpServers = types.StringValue(p.TrustedDhcpServers)
	model.TrustedEgressIps = types.StringValue(p.TrustedEgressIps)
	model.PredefinedTrustedNetworks = types.BoolValue(p.PredefinedTrustedNetworks)
	model.PredefinedTnAll = types.BoolValue(p.PredefinedTnAll)
	model.EnableSplitVpnTN = intToBool(p.EnableSplitVpnTN)
	model.EnableUnifiedTunnel = intToBool(p.EnableUnifiedTunnel)
	model.EvaluateTrustedNetwork = intToBool(p.EvaluateTrustedNetwork)
	model.SkipTrustedCriteriaMatch = intToBool(p.SkipTrustedCriteriaMatch)
	if p.EnableAllDefaultAdaptersTN != 0 {
		model.EnableAllDefaultAdaptersTN = intToBool(p.EnableAllDefaultAdaptersTN)
	}

	// The webForwardingProfile/listByCompany GET response does not echo
	// `evaluateTrustedNetwork`, so the SDK always unmarshals it as 0 (false).
	// Preserve the plan/state value when the user set one to avoid spurious
	// drift after every apply.
	if !priorEvaluateTrustedNetwork.IsNull() && !priorEvaluateTrustedNetwork.IsUnknown() {
		model.EvaluateTrustedNetwork = priorEvaluateTrustedNetwork
	}

	trustedNetworkIdVals := make([]attr.Value, 0, len(p.TrustedNetworkIds))
	for _, v := range p.TrustedNetworkIds {
		trustedNetworkIdVals = append(trustedNetworkIdVals, types.Int64Value(int64(v)))
	}
	model.TrustedNetworkIds, _ = types.ListValue(types.Int64Type, trustedNetworkIdVals)

	trustedNetworkVals := make([]attr.Value, 0, len(p.TrustedNetworks))
	for _, v := range p.TrustedNetworks {
		trustedNetworkVals = append(trustedNetworkVals, types.StringValue(v))
	}
	model.TrustedNetworks, _ = types.ListValue(types.StringType, trustedNetworkVals)

	trustedNetworkIdSelectedVals := make([]attr.Value, 0, len(p.TrustedNetworkIdsSelected))
	for _, v := range p.TrustedNetworkIdsSelected {
		trustedNetworkIdSelectedVals = append(trustedNetworkIdSelectedVals, types.Int64Value(int64(v)))
	}
	model.TrustedNetworkIdsSelected, _ = types.ListValue(types.Int64Type, trustedNetworkIdSelectedVals)

	var planActionOrder []int
	planActionModels := map[int]ForwardingProfileActionModel{}
	if !model.ForwardingProfileActions.IsNull() && !model.ForwardingProfileActions.IsUnknown() {
		var existingActions []ForwardingProfileActionModel
		model.ForwardingProfileActions.ElementsAs(ctx, &existingActions, false)
		for _, e := range existingActions {
			nt := int(e.NetworkType.ValueInt64())
			planActionOrder = append(planActionOrder, nt)
			planActionModels[nt] = e
		}
	}
	model.ForwardingProfileActions = flattenProfileActions(p.ForwardingProfileActions, planActionModels, planActionOrder)

	var planZpaOrder []int
	planZpaModels := map[int]ForwardingProfileZpaActionModel{}
	if !model.ForwardingProfileZpaActions.IsNull() && !model.ForwardingProfileZpaActions.IsUnknown() {
		var existing []ForwardingProfileZpaActionModel
		model.ForwardingProfileZpaActions.ElementsAs(ctx, &existing, false)
		for _, e := range existing {
			nt := int(e.NetworkType.ValueInt64())
			planZpaOrder = append(planZpaOrder, nt)
			planZpaModels[nt] = e
		}
	}
	model.ForwardingProfileZpaActions = flattenZpaActions(p.ForwardingProfileZpaActions, planZpaModels, planZpaOrder)

	var planTunnelOrder []int
	planTunnelModels := map[int]UnifiedTunnelModel{}
	if !model.UnifiedTunnel.IsNull() && !model.UnifiedTunnel.IsUnknown() {
		var existingTunnels []UnifiedTunnelModel
		model.UnifiedTunnel.ElementsAs(ctx, &existingTunnels, false)
		for _, e := range existingTunnels {
			nt := int(e.NetworkType.ValueInt64())
			planTunnelOrder = append(planTunnelOrder, nt)
			planTunnelModels[nt] = e
		}
	}
	model.UnifiedTunnel = flattenUnifiedTunnels(p.UnifiedTunnel, planTunnelModels, planTunnelOrder)
}

func flattenSystemProxyDataObject(spd forwarding_profile.SystemProxyData) types.Object {
	obj, _ := types.ObjectValue(
		systemProxyDataAttrTypes(),
		map[string]attr.Value{
			"proxy_action":                types.Int64Value(int64(spd.ProxyAction)),
			"enable_auto_detect":          types.BoolValue(spd.EnableAutoDetect != 0),
			"enable_pac":                  types.BoolValue(spd.EnablePAC != 0),
			"pac_url":                     types.StringValue(spd.PacURL),
			"enable_proxy_server":         types.BoolValue(spd.EnableProxyServer != 0),
			"proxy_server_address":        types.StringValue(spd.ProxyServerAddress),
			"proxy_server_port":           types.StringValue(spd.ProxyServerPort),
			"bypass_proxy_for_private_ip": types.BoolValue(spd.BypassProxyForPrivateIP != 0),
			"perform_gp_update":           types.BoolValue(spd.PerformGPUpdate != 0),
			"pac_data_path":               types.StringValue(spd.PacDataPath),
		},
	)
	return obj
}

func sortByPlanOrder[T any](items []T, getNetworkType func(T) int, planOrder []int) []T {
	if len(planOrder) == 0 {
		return items
	}
	orderIndex := make(map[int]int, len(planOrder))
	for i, nt := range planOrder {
		orderIndex[nt] = i
	}
	sorted := make([]T, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool {
		idxI, okI := orderIndex[getNetworkType(sorted[i])]
		idxJ, okJ := orderIndex[getNetworkType(sorted[j])]
		if !okI {
			idxI = len(planOrder)
		}
		if !okJ {
			idxJ = len(planOrder)
		}
		return idxI < idxJ
	})
	return sorted
}

func flattenProfileActions(actions []forwarding_profile.ForwardingProfileAction, planModels map[int]ForwardingProfileActionModel, planOrder []int) types.List {
	if len(actions) == 0 {
		return types.ListNull(forwardingProfileActionObjectType())
	}

	actions = sortByPlanOrder(actions, func(a forwarding_profile.ForwardingProfileAction) int { return a.NetworkType }, planOrder)

	elems := make([]attr.Value, 0, len(actions))
	for _, a := range actions {
		// The API omits isSameAsOnTrustedNetwork from the GET response,
		// so Go defaults it to false. Use the plan value to avoid drift.
		isSame := types.BoolValue(a.IsSameAsOnTrustedNetwork)
		if pm, ok := planModels[a.NetworkType]; ok {
			if !pm.IsSameAsOnTrustedNetwork.IsNull() && !pm.IsSameAsOnTrustedNetwork.IsUnknown() {
				isSame = pm.IsSameAsOnTrustedNetwork
			}
		}

		obj, _ := types.ObjectValue(
			forwardingProfileActionAttrTypes(),
			map[string]attr.Value{
				"network_type":                            types.Int64Value(int64(a.NetworkType)),
				"action_type":                             types.Int64Value(int64(a.ActionType)),
				"primary_transport":                       types.Int64Value(int64(a.PrimaryTransport)),
				"dtls_timeout":                            types.Int64Value(int64(a.DTLSTimeout)),
				"udp_timeout":                             types.Int64Value(int64(a.UDPTimeout)),
				"tls_timeout":                             types.Int64Value(int64(a.TLSTimeout)),
				"mtu_for_zadapter":                        types.Int64Value(int64(a.MtuForZadapter)),
				"tunnel2_fallback_type":                   types.Int64Value(int64(a.Tunnel2FallbackType)),
				"zen_probe_interval":                      types.Int64Value(int64(a.ZenProbeInterval)),
				"zen_probe_sample_size":                   types.Int64Value(int64(a.ZenProbeSampleSize)),
				"zen_threshold_limit":                     types.Int64Value(int64(a.ZenThresholdLimit)),
				"lbs_probe_interval":                      types.Int64Value(int64(a.LbsProbeInterval)),
				"lbs_probe_sample_size":                   types.Int64Value(int64(a.LbsProbeSampleSize)),
				"lbs_threshold_limit":                     types.Int64Value(int64(a.LbsThresholdLimit)),
				"custom_pac":                              types.StringValue(a.CustomPac),
				"system_proxy":                            intToBool(a.SystemProxy),
				"enable_packet_tunnel":                    intToBool(a.EnablePacketTunnel),
				"block_unreachable_domains_traffic":       intToBool(int(a.BlockUnreachableDomainsTraffic)),
				"allow_tls_fallback":                      intToBool(a.AllowTLSFallback),
				"send_all_dns_to_trusted_server":          intToBool(a.SendAllDNSToTrustedServer),
				"drop_ipv6_traffic":                       intToBool(int(a.DropIpv6Traffic)),
				"redirect_web_traffic":                    intToBool(int(a.RedirectWebTraffic)),
				"drop_ipv6_include_traffic_in_t2":         intToBool(int(a.DropIpv6IncludeTrafficInT2)),
				"use_tunnel2_for_proxied_web_traffic":     intToBool(a.UseTunnel2ForProxiedWebTraffic),
				"use_tunnel2_for_unencrypted_web_traffic": intToBool(a.UseTunnel2ForUnencryptedWebTraffic),
				"path_mtu_discovery":                      intToBool(a.PathMtuDiscovery),
				"latency_based_zen_enablement":            intToBool(int(a.LatencyBasedZenEnablement)),
				"latency_based_server_enablement":         intToBool(a.LatencyBasedServerEnablement),
				"latency_based_server_mt_enablement":      intToBool(a.LatencyBasedServerMTEnablement),
				"drop_ipv6_traffic_in_ipv6_network":       intToBool(int(a.DropIpv6TrafficInIpv6Network)),
				"optimise_for_unstable_connections":       intToBool(a.OptimiseForUnstableConnections),
				"is_same_as_on_trusted_network":           isSame,
				"system_proxy_data":                       flattenSystemProxyDataObject(a.SystemProxyData),
			},
		)
		elems = append(elems, obj)
	}
	list, _ := types.ListValue(forwardingProfileActionObjectType(), elems)
	return list
}

func flattenZpaActions(actions []forwarding_profile.ForwardingProfileZpaAction, planModels map[int]ForwardingProfileZpaActionModel, planOrder []int) types.List {
	if len(actions) == 0 {
		return types.ListNull(zpaActionObjectType())
	}

	actions = sortByPlanOrder(actions, func(a forwarding_profile.ForwardingProfileZpaAction) int { return a.NetworkType }, planOrder)

	elems := make([]attr.Value, 0, len(actions))
	for _, a := range actions {
		piObj, _ := types.ObjectValue(
			partnerInfoAttrTypes(),
			map[string]attr.Value{
				"primary_transport":  types.Int64Value(int64(a.PartnerInfo.PrimaryTransport)),
				"mtu_for_zadapter":   types.Int64Value(int64(a.PartnerInfo.MtuForZadapter)),
				"allow_tls_fallback": intToBool(a.PartnerInfo.AllowTlsFallback),
			},
		)

		// The SDK's ForwardingProfileZpaAction struct currently declares the
		// JSON tags latencyBasedZpaServerEnablement, lbsZpaProbeInterval,
		// lbsZpaProbeSampleSize, and lbsZpaThresholdLimit, but the API
		// actually returns the un-prefixed forms (latencyBasedServerEnablement,
		// lbsProbeInterval, etc.). The SDK therefore leaves those fields at
		// zero after unmarshal. Until the SDK is corrected we preserve the
		// plan values here. The same trick handles isSameAsOnTrustedNetwork,
		// which the API omits entirely from the GET response.
		lbsEnablement := intToBool(a.LatencyBasedServerEnablement)
		lbsProbeInterval := types.Int64Value(int64(a.LbsProbeInterval))
		lbsProbeSampleSize := types.Int64Value(int64(a.LbsProbeSampleSize))
		lbsThresholdLimit := types.Int64Value(int64(a.LbsThresholdLimit))
		isSame := types.BoolValue(a.IsSameAsOnTrustedNetwork)
		if pm, ok := planModels[a.NetworkType]; ok {
			if !pm.LatencyBasedServerEnablement.IsNull() && !pm.LatencyBasedServerEnablement.IsUnknown() {
				lbsEnablement = pm.LatencyBasedServerEnablement
			}
			if !pm.LbsProbeInterval.IsNull() && !pm.LbsProbeInterval.IsUnknown() {
				lbsProbeInterval = pm.LbsProbeInterval
			}
			if !pm.LbsProbeSampleSize.IsNull() && !pm.LbsProbeSampleSize.IsUnknown() {
				lbsProbeSampleSize = pm.LbsProbeSampleSize
			}
			if !pm.LbsThresholdLimit.IsNull() && !pm.LbsThresholdLimit.IsUnknown() {
				lbsThresholdLimit = pm.LbsThresholdLimit
			}
			if !pm.IsSameAsOnTrustedNetwork.IsNull() && !pm.IsSameAsOnTrustedNetwork.IsUnknown() {
				isSame = pm.IsSameAsOnTrustedNetwork
			}
		}

		obj, _ := types.ObjectValue(
			zpaActionAttrTypes(),
			map[string]attr.Value{
				"network_type":                       types.Int64Value(int64(a.NetworkType)),
				"action_type":                        types.Int64Value(int64(a.ActionType)),
				"primary_transport":                  types.Int64Value(int64(a.PrimaryTransport)),
				"dtls_timeout":                       types.Int64Value(int64(a.DTLSTimeout)),
				"tls_timeout":                        types.Int64Value(int64(a.TLSTimeout)),
				"mtu_for_zadapter":                   types.Int64Value(int64(a.MtuForZadapter)),
				"lbs_probe_interval":                 lbsProbeInterval,
				"lbs_probe_sample_size":              lbsProbeSampleSize,
				"lbs_threshold_limit":                lbsThresholdLimit,
				"send_trusted_network_result_to_zpa": intToBool(a.SendTrustedNetworkResultToZpa),
				"latency_based_server_enablement":    lbsEnablement,
				"latency_based_server_mt_enablement": intToBool(a.LatencyBasedServerMTEnablement),
				"is_same_as_on_trusted_network":      isSame,
				"partner_info":                       piObj,
			},
		)
		elems = append(elems, obj)
	}
	list, _ := types.ListValue(zpaActionObjectType(), elems)
	return list
}

func flattenUnifiedTunnels(tunnels []forwarding_profile.UnifiedTunnel, planModels map[int]UnifiedTunnelModel, planOrder []int) types.List {
	if len(tunnels) == 0 {
		return types.ListNull(unifiedTunnelObjectType())
	}

	tunnels = sortByPlanOrder(tunnels, func(t forwarding_profile.UnifiedTunnel) int { return t.NetworkType }, planOrder)

	elems := make([]attr.Value, 0, len(tunnels))
	for _, t := range tunnels {
		// When actionTypeZIA and actionTypeZPA are both 0 the API silently
		// resets nested systemProxyData fields (most notably proxyAction) to
		// zero on the GET response, even when the POST body explicitly sent
		// non-zero values. Overlay the plan's known systemProxyData attributes
		// on top of the API result so explicit user values survive the round
		// trip; unspecified (unknown) plan attributes fall back to the API
		// value so we never propagate an unknown into state.
		spdObj := flattenSystemProxyDataObject(t.SystemProxyData)
		if pm, ok := planModels[t.NetworkType]; ok {
			spdObj = overlayKnownPlanAttributes(spdObj, pm.SystemProxyData, systemProxyDataAttrTypes())
		}

		obj, _ := types.ObjectValue(
			unifiedTunnelAttrTypes(),
			map[string]attr.Value{
				"network_type":                      types.Int64Value(int64(t.NetworkType)),
				"action_type_zia":                   types.Int64Value(int64(t.ActionTypeZIA)),
				"action_type_zpa":                   types.Int64Value(int64(t.ActionTypeZPA)),
				"primary_transport":                 types.Int64Value(int64(t.PrimaryTransport)),
				"dtls_timeout":                      types.Int64Value(int64(t.DTLSTimeout)),
				"tls_timeout":                       types.Int64Value(int64(t.TLSTimeout)),
				"mtu_for_zadapter":                  types.Int64Value(int64(t.MtuForZadapter)),
				"tunnel2_fallback_type":             types.Int64Value(int64(t.Tunnel2FallbackType)),
				"allow_tls_fallback":                intToBool(t.AllowTLSFallback),
				"path_mtu_discovery":                intToBool(t.PathMtuDiscovery),
				"optimise_for_unstable_connections": intToBool(t.OptimiseForUnstableConnections),
				"redirect_web_traffic":              intToBool(t.RedirectWebTraffic),
				"drop_ipv6_traffic":                 intToBool(t.DropIpv6Traffic),
				"drop_ipv6_traffic_in_ipv6_network": intToBool(t.DropIpv6TrafficInIpv6Network),
				"block_unreachable_domains_traffic": intToBool(t.BlockUnreachableDomainsTraffic),
				"drop_ipv6_include_traffic_in_t2":   intToBool(t.DropIpv6IncludeTrafficInT2),
				"send_all_dns_to_trusted_server":    intToBool(t.SendAllDNSToTrustedServer),
				"same_as_on_trusted":                intToBool(t.SameAsOnTrusted),
				"system_proxy_data":                 spdObj,
			},
		)
		elems = append(elems, obj)
	}
	list, _ := types.ListValue(unifiedTunnelObjectType(), elems)
	return list
}

// overlayKnownPlanAttributes returns a copy of apiObj where any attribute that
// the user explicitly set in planObj (i.e. is non-null and known) is replaced
// by the plan value. Unknown / null plan attributes are left untouched so
// Optional+Computed fields that the user did not write in HCL still come from
// the API response. This is the safe shape for fields where the API
// occasionally normalises user-supplied values out (e.g. proxyAction inside
// unified_tunnel.system_proxy_data when both action types are 0) without
// leaking unknown values into post-apply state.
func overlayKnownPlanAttributes(apiObj, planObj types.Object, attrTypes map[string]attr.Type) types.Object {
	if planObj.IsNull() || planObj.IsUnknown() {
		return apiObj
	}
	apiAttrs := apiObj.Attributes()
	planAttrs := planObj.Attributes()
	merged := make(map[string]attr.Value, len(attrTypes))
	for name := range attrTypes {
		if pv, ok := planAttrs[name]; ok && pv != nil && !pv.IsNull() && !pv.IsUnknown() {
			merged[name] = pv
			continue
		}
		if av, ok := apiAttrs[name]; ok {
			merged[name] = av
			continue
		}
		merged[name] = types.StringNull()
	}
	obj, _ := types.ObjectValue(attrTypes, merged)
	return obj
}

// ---------------------------------------------------------------------------
// Attr-type helpers
// ---------------------------------------------------------------------------

func systemProxyDataAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"proxy_action":                types.Int64Type,
		"enable_auto_detect":          types.BoolType,
		"enable_pac":                  types.BoolType,
		"pac_url":                     types.StringType,
		"enable_proxy_server":         types.BoolType,
		"proxy_server_address":        types.StringType,
		"proxy_server_port":           types.StringType,
		"bypass_proxy_for_private_ip": types.BoolType,
		"perform_gp_update":           types.BoolType,
		"pac_data_path":               types.StringType,
	}
}

func systemProxyDataObjectType() attr.Type {
	return types.ObjectType{AttrTypes: systemProxyDataAttrTypes()}
}

func forwardingProfileActionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"network_type":                            types.Int64Type,
		"action_type":                             types.Int64Type,
		"primary_transport":                       types.Int64Type,
		"dtls_timeout":                            types.Int64Type,
		"udp_timeout":                             types.Int64Type,
		"tls_timeout":                             types.Int64Type,
		"mtu_for_zadapter":                        types.Int64Type,
		"tunnel2_fallback_type":                   types.Int64Type,
		"zen_probe_interval":                      types.Int64Type,
		"zen_probe_sample_size":                   types.Int64Type,
		"zen_threshold_limit":                     types.Int64Type,
		"lbs_probe_interval":                      types.Int64Type,
		"lbs_probe_sample_size":                   types.Int64Type,
		"lbs_threshold_limit":                     types.Int64Type,
		"custom_pac":                              types.StringType,
		"system_proxy":                            types.BoolType,
		"enable_packet_tunnel":                    types.BoolType,
		"block_unreachable_domains_traffic":       types.BoolType,
		"allow_tls_fallback":                      types.BoolType,
		"send_all_dns_to_trusted_server":          types.BoolType,
		"drop_ipv6_traffic":                       types.BoolType,
		"redirect_web_traffic":                    types.BoolType,
		"drop_ipv6_include_traffic_in_t2":         types.BoolType,
		"use_tunnel2_for_proxied_web_traffic":     types.BoolType,
		"use_tunnel2_for_unencrypted_web_traffic": types.BoolType,
		"path_mtu_discovery":                      types.BoolType,
		"latency_based_zen_enablement":            types.BoolType,
		"latency_based_server_enablement":         types.BoolType,
		"latency_based_server_mt_enablement":      types.BoolType,
		"drop_ipv6_traffic_in_ipv6_network":       types.BoolType,
		"optimise_for_unstable_connections":       types.BoolType,
		"is_same_as_on_trusted_network":           types.BoolType,
		"system_proxy_data":                       systemProxyDataObjectType(),
	}
}

func forwardingProfileActionObjectType() attr.Type {
	return types.ObjectType{AttrTypes: forwardingProfileActionAttrTypes()}
}

func partnerInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"primary_transport":  types.Int64Type,
		"mtu_for_zadapter":   types.Int64Type,
		"allow_tls_fallback": types.BoolType,
	}
}

func partnerInfoObjectType() attr.Type {
	return types.ObjectType{AttrTypes: partnerInfoAttrTypes()}
}

func zpaActionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"network_type":                       types.Int64Type,
		"action_type":                        types.Int64Type,
		"primary_transport":                  types.Int64Type,
		"dtls_timeout":                       types.Int64Type,
		"tls_timeout":                        types.Int64Type,
		"mtu_for_zadapter":                   types.Int64Type,
		"lbs_probe_interval":                 types.Int64Type,
		"lbs_probe_sample_size":              types.Int64Type,
		"lbs_threshold_limit":                types.Int64Type,
		"send_trusted_network_result_to_zpa": types.BoolType,
		"latency_based_server_enablement":    types.BoolType,
		"latency_based_server_mt_enablement": types.BoolType,
		"is_same_as_on_trusted_network":      types.BoolType,
		"partner_info":                       partnerInfoObjectType(),
	}
}

func zpaActionObjectType() attr.Type {
	return types.ObjectType{AttrTypes: zpaActionAttrTypes()}
}

func unifiedTunnelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"network_type":                      types.Int64Type,
		"action_type_zia":                   types.Int64Type,
		"action_type_zpa":                   types.Int64Type,
		"primary_transport":                 types.Int64Type,
		"dtls_timeout":                      types.Int64Type,
		"tls_timeout":                       types.Int64Type,
		"mtu_for_zadapter":                  types.Int64Type,
		"tunnel2_fallback_type":             types.Int64Type,
		"allow_tls_fallback":                types.BoolType,
		"path_mtu_discovery":                types.BoolType,
		"optimise_for_unstable_connections": types.BoolType,
		"redirect_web_traffic":              types.BoolType,
		"drop_ipv6_traffic":                 types.BoolType,
		"drop_ipv6_traffic_in_ipv6_network": types.BoolType,
		"block_unreachable_domains_traffic": types.BoolType,
		"drop_ipv6_include_traffic_in_t2":   types.BoolType,
		"send_all_dns_to_trusted_server":    types.BoolType,
		"same_as_on_trusted":                types.BoolType,
		"system_proxy_data":                 systemProxyDataObjectType(),
	}
}

func unifiedTunnelObjectType() attr.Type {
	return types.ObjectType{AttrTypes: unifiedTunnelAttrTypes()}
}
