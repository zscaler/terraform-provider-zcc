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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/forwarding_profile"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &ForwardingProfileDataSource{}
	_ datasource.DataSourceWithConfigure = &ForwardingProfileDataSource{}
)

func NewForwardingProfileDataSource() datasource.DataSource {
	return &ForwardingProfileDataSource{}
}

type ForwardingProfileDataSource struct {
	client *client.Client
}

type ForwardingProfileDataSourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Active                     types.Bool   `tfsdk:"active"`
	ConditionType              types.Int64  `tfsdk:"condition_type"`
	DnsSearchDomains           types.String `tfsdk:"dns_search_domains"`
	DnsServers                 types.String `tfsdk:"dns_servers"`
	EnableLWFDriver            types.Bool   `tfsdk:"enable_lwf_driver"`
	EnableSplitVpnTN           types.Bool   `tfsdk:"enable_split_vpn_tn"`
	EnableUnifiedTunnel        types.Bool   `tfsdk:"enable_unified_tunnel"`
	EvaluateTrustedNetwork     types.Bool   `tfsdk:"evaluate_trusted_network"`
	EnableAllDefaultAdaptersTN types.Bool   `tfsdk:"enable_all_default_adapters_tn"`
	Hostname                   types.String `tfsdk:"hostname"`
	PredefinedTnAll            types.Bool   `tfsdk:"predefined_tn_all"`
	PredefinedTrustedNetworks  types.Bool   `tfsdk:"predefined_trusted_networks"`
	ResolvedIpsForHostname     types.String `tfsdk:"resolved_ips_for_hostname"`
	SkipTrustedCriteriaMatch   types.Bool   `tfsdk:"skip_trusted_criteria_match"`
	TrustedDhcpServers         types.String `tfsdk:"trusted_dhcp_servers"`
	TrustedEgressIps           types.String `tfsdk:"trusted_egress_ips"`
	TrustedGateways            types.String `tfsdk:"trusted_gateways"`
	TrustedNetworkIds          types.List   `tfsdk:"trusted_network_ids"`
	TrustedNetworks            types.List   `tfsdk:"trusted_networks"`
	TrustedSubnets             types.String `tfsdk:"trusted_subnets"`
}

func (d *ForwardingProfileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forwarding_profile"
}

func (d *ForwardingProfileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC forwarding profile by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the forwarding profile.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the forwarding profile.",
				Optional:    true,
				Computed:    true,
			},
			"active":                         schema.BoolAttribute{Computed: true},
			"condition_type":                 schema.Int64Attribute{Computed: true},
			"dns_search_domains":             schema.StringAttribute{Computed: true},
			"dns_servers":                    schema.StringAttribute{Computed: true},
			"enable_lwf_driver":              schema.BoolAttribute{Computed: true},
			"enable_split_vpn_tn":            schema.BoolAttribute{Computed: true},
			"enable_unified_tunnel":          schema.BoolAttribute{Computed: true},
			"evaluate_trusted_network":       schema.BoolAttribute{Computed: true},
			"enable_all_default_adapters_tn": schema.BoolAttribute{Computed: true},
			"hostname":                       schema.StringAttribute{Computed: true},
			"predefined_tn_all":              schema.BoolAttribute{Computed: true},
			"predefined_trusted_networks":    schema.BoolAttribute{Computed: true},
			"resolved_ips_for_hostname":      schema.StringAttribute{Computed: true},
			"skip_trusted_criteria_match":    schema.BoolAttribute{Computed: true},
			"trusted_dhcp_servers":           schema.StringAttribute{Computed: true},
			"trusted_egress_ips":             schema.StringAttribute{Computed: true},
			"trusted_gateways":               schema.StringAttribute{Computed: true},
			"trusted_network_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Computed:    true,
			},
			"trusted_networks": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"trusted_subnets": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *ForwardingProfileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ForwardingProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ForwardingProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.Name.IsNull() || data.Name.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or name must be specified")
		return
	}

	service := d.client.Service

	tflog.Info(ctx, "Fetching forwarding profiles")
	profiles, err := forwarding_profile.GetForwardingProfileByCompanyID(ctx, service, "", nil, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read forwarding profiles: %v", err))
		return
	}

	var profile *forwarding_profile.ForwardingProfile
	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		idStr := data.ID.ValueString()
		idInt, parseErr := strconv.Atoi(idStr)
		if parseErr == nil {
			for i := range profiles {
				if int(profiles[i].ID) == idInt {
					profile = &profiles[i]
					break
				}
			}
		}
		if profile == nil {
			for i := range profiles {
				if strconv.Itoa(int(profiles[i].ID)) == idStr {
					profile = &profiles[i]
					break
				}
			}
		}
		if profile == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Forwarding profile with ID '%s' not found", idStr))
			return
		}
	} else {
		name := data.Name.ValueString()
		for i := range profiles {
			if profiles[i].Name == name {
				profile = &profiles[i]
				break
			}
		}
		if profile == nil {
			resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Forwarding profile with name '%s' not found", name))
			return
		}
	}

	trustedNetworkIdVals := make([]attr.Value, 0, len(profile.TrustedNetworkIds))
	for _, v := range profile.TrustedNetworkIds {
		trustedNetworkIdVals = append(trustedNetworkIdVals, types.Int64Value(int64(v)))
	}

	trustedNetworkVals := make([]attr.Value, 0, len(profile.TrustedNetworks))
	for _, v := range profile.TrustedNetworks {
		trustedNetworkVals = append(trustedNetworkVals, types.StringValue(v))
	}

	model := ForwardingProfileDataSourceModel{
		ID:                         types.StringValue(strconv.Itoa(int(profile.ID))),
		Name:                       types.StringValue(profile.Name),
		Active:                     types.BoolValue(profile.Active == "1"),
		ConditionType:              types.Int64Value(int64(profile.ConditionType)),
		DnsSearchDomains:           types.StringValue(profile.DnsSearchDomains),
		DnsServers:                 types.StringValue(profile.DnsServers),
		EnableLWFDriver:            types.BoolValue(profile.EnableLWFDriver == "1"),
		EnableSplitVpnTN:           types.BoolValue(profile.EnableSplitVpnTN != 0),
		EnableUnifiedTunnel:        types.BoolValue(profile.EnableUnifiedTunnel != 0),
		EvaluateTrustedNetwork:     types.BoolValue(profile.EvaluateTrustedNetwork != 0),
		EnableAllDefaultAdaptersTN: types.BoolValue(profile.EnableAllDefaultAdaptersTN != 0),
		Hostname:                   types.StringValue(profile.Hostname),
		PredefinedTnAll:            types.BoolValue(profile.PredefinedTnAll),
		PredefinedTrustedNetworks:  types.BoolValue(profile.PredefinedTrustedNetworks),
		ResolvedIpsForHostname:     types.StringValue(profile.ResolvedIpsForHostname),
		SkipTrustedCriteriaMatch:   types.BoolValue(profile.SkipTrustedCriteriaMatch != 0),
		TrustedDhcpServers:         types.StringValue(profile.TrustedDhcpServers),
		TrustedEgressIps:           types.StringValue(profile.TrustedEgressIps),
		TrustedGateways:            types.StringValue(profile.TrustedGateways),
		TrustedSubnets:             types.StringValue(profile.TrustedSubnets),
	}
	model.TrustedNetworkIds, _ = types.ListValue(types.Int64Type, trustedNetworkIdVals)
	model.TrustedNetworks, _ = types.ListValue(types.StringType, trustedNetworkVals)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
