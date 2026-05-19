package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
)

var (
	_ datasource.DataSource              = &TrustedNetworkDataSource{}
	_ datasource.DataSourceWithConfigure = &TrustedNetworkDataSource{}
)

func NewTrustedNetworkDataSource() datasource.DataSource {
	return &TrustedNetworkDataSource{}
}

type TrustedNetworkDataSource struct {
	client *client.Client
}

type TrustedNetworkDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	NetworkName            types.String `tfsdk:"network_name"`
	Active                 types.Bool   `tfsdk:"active"`
	ConditionType          types.Int64  `tfsdk:"condition_type"`
	CreatedBy              types.String `tfsdk:"created_by"`
	DnsSearchDomains       types.String `tfsdk:"dns_search_domains"`
	DnsServers             types.String `tfsdk:"dns_servers"`
	EditBy                 types.String `tfsdk:"edit_by"`
	Guid                   types.String `tfsdk:"guid"`
	Hostnames              types.String `tfsdk:"hostnames"`
	ResolvedIpsForHostname types.String `tfsdk:"resolved_ips_for_hostname"`
	Ssid                   types.String `tfsdk:"ssid"`
	TrustedDhcpServers     types.String `tfsdk:"trusted_dhcp_servers"`
	TrustedEgressIps       types.String `tfsdk:"trusted_egress_ips"`
	TrustedGateways        types.String `tfsdk:"trusted_gateways"`
	TrustedSubnets         types.String `tfsdk:"trusted_subnets"`
}

func (d *TrustedNetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trusted_network"
}

func (d *TrustedNetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC trusted network by ID or network name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the trusted network.",
				Optional:    true,
				Computed:    true,
			},
			"network_name": schema.StringAttribute{
				Description: "The name of the trusted network.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the trusted network is active.",
				Computed:    true,
			},
			"condition_type": schema.Int64Attribute{
				Description: "Condition type code from the API (`0` and `1` are both valid).",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "User who created the trusted network.",
				Computed:    true,
			},
			"dns_search_domains": schema.StringAttribute{
				Description: "DNS search domains.",
				Optional:    true,
				Computed:    true,
			},
			"dns_servers": schema.StringAttribute{
				Description: "DNS servers.",
				Computed:    true,
			},
			"edit_by": schema.StringAttribute{
				Description: "User who last edited the trusted network.",
				Computed:    true,
			},
			"guid": schema.StringAttribute{
				Description: "GUID of the trusted network.",
				Computed:    true,
			},
			"hostnames": schema.StringAttribute{
				Description: "Hostnames associated with the trusted network.",
				Computed:    true,
			},
			"resolved_ips_for_hostname": schema.StringAttribute{
				Description: "Resolved IPs for hostname.",
				Computed:    true,
			},
			"ssid": schema.StringAttribute{
				Description: "SSIDs associated with the trusted network.",
				Computed:    true,
			},
			"trusted_dhcp_servers": schema.StringAttribute{
				Description: "Trusted DHCP servers.",
				Computed:    true,
			},
			"trusted_egress_ips": schema.StringAttribute{
				Description: "Trusted egress IPs.",
				Computed:    true,
			},
			"trusted_gateways": schema.StringAttribute{
				Description: "Trusted gateways.",
				Computed:    true,
			},
			"trusted_subnets": schema.StringAttribute{
				Description: "Trusted subnets.",
				Computed:    true,
			},
		},
	}
}

func (d *TrustedNetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TrustedNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TrustedNetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (data.ID.IsNull() || data.ID.ValueString() == "") && (data.NetworkName.IsNull() || data.NetworkName.ValueString() == "") {
		resp.Diagnostics.AddError("Missing Identifier", "Either id or network_name must be specified")
		return
	}

	service := d.client.Service

	var (
		net *trusted_network.TrustedNetwork
		err error
	)

	if !data.ID.IsNull() && data.ID.ValueString() != "" {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching trusted network", map[string]any{"id": id})
		net, _, err = trusted_network.GetTrustedNetworkByID(ctx, service, id)
	} else {
		name := data.NetworkName.ValueString()
		tflog.Info(ctx, "Fetching trusted network", map[string]any{"network_name": name})
		net, _, err = trusted_network.GetTrustedNetworkByName(ctx, service, name)
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read trusted network: %v", err))
		return
	}

	model := TrustedNetworkDataSourceModel{
		ID:                     types.StringValue(net.ID),
		NetworkName:            types.StringValue(net.NetworkName),
		Active:                 types.BoolValue(net.Active),
		ConditionType:          types.Int64Value(int64(net.ConditionType)),
		CreatedBy:              types.StringValue(net.CreatedBy),
		DnsSearchDomains:       types.StringValue(net.DnsSearchDomains),
		DnsServers:             types.StringValue(net.DnsServers),
		EditBy:                 types.StringValue(net.EditedBy),
		Guid:                   types.StringValue(net.Guid),
		Hostnames:              types.StringValue(net.Hostnames),
		ResolvedIpsForHostname: types.StringValue(net.ResolvedIpsForHostname),
		Ssid:                   types.StringValue(net.Ssids),
		TrustedDhcpServers:     types.StringValue(net.TrustedDhcpServers),
		TrustedEgressIps:       types.StringValue(net.TrustedEgressIps),
		TrustedGateways:        types.StringValue(net.TrustedGateways),
		TrustedSubnets:         types.StringValue(net.TrustedSubnets),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
