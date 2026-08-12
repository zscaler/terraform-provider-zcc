package datasources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/helpers"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/tnbackend"
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

// TrustedNetworkDataSourceModel mirrors trusted_network_v2.TrustedNetworkV2
// in full — unlike the resource, the data source surfaces every field
// (including the server-side metadata company_id, created_by, edited_by,
// guid) so callers can introspect networks they did not create.
type TrustedNetworkDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	CompanyID              types.Int64  `tfsdk:"company_id"`
	ZPAID                  types.String `tfsdk:"zpa_id"`
	Active                 types.Bool   `tfsdk:"active"`
	ConditionType          types.String `tfsdk:"condition_type"`
	Name                   types.String `tfsdk:"name"`
	CreatedBy              types.String `tfsdk:"created_by"`
	EditedBy               types.String `tfsdk:"edited_by"`
	Guid                   types.String `tfsdk:"guid"`
	Hostname               types.String `tfsdk:"hostname"`
	SSID                   types.String `tfsdk:"ssid"`
	DNSSearchDomains       types.List   `tfsdk:"dns_search_domains"`
	DNSServerIPs           types.List   `tfsdk:"dns_server_ips"`
	ResolvedIPsForHostname types.List   `tfsdk:"resolved_ips_for_hostname"`
	TrustedDhcpServersIPs  types.List   `tfsdk:"trusted_dhcp_servers_ips"`
	TrustedEgressIPs       types.List   `tfsdk:"trusted_egress_ips"`
	TrustedGatewayIPs      types.List   `tfsdk:"trusted_gateway_ips"`
	TrustedSubnetIPs       types.List   `tfsdk:"trusted_subnet_ips"`
}

func (d *TrustedNetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trusted_network"
}

func (d *TrustedNetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	stringList := func(desc string) schema.Attribute {
		return schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: desc,
		}
	}
	resp.Schema = schema.Schema{
		Description: "Retrieves a ZCC trusted network by numeric id or by network name. The data source " +
			"automatically detects which API generation the tenant serves: it uses the " +
			"/zcc/papi/public/v2/trusted-networks endpoint where available and falls back to " +
			"/zcc/papi/public/v1/webTrustedNetwork otherwise. `zpa_id` is only populated by the v2 API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the trusted network (carried as a string). Either `id` or `name` must be set.",
				Optional:    true,
				Computed:    true,
			},
			"company_id": schema.Int64Attribute{
				Description: "Numeric company id the network is scoped to.",
				Computed:    true,
			},
			"zpa_id": schema.StringAttribute{
				Description: "Linked ZPA tenant identifier.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the trusted network is active.",
				Computed:    true,
			},
			"condition_type": schema.StringAttribute{
				Description: "Match policy applied across the criteria (`ALL`/`ANY` or the numeric forms `0`/`1`).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the trusted network. Either `id` or `name` must be set; the data source looks up by the supplied value via the SDK's `GetByName` when `id` is omitted.",
				Optional:    true,
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "User who created the trusted network.",
				Computed:    true,
			},
			"edited_by": schema.StringAttribute{
				Description: "User who last edited the trusted network.",
				Computed:    true,
			},
			"guid": schema.StringAttribute{
				Description: "Stable GUID of the trusted network.",
				Computed:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "Hostname used to identify the network.",
				Computed:    true,
			},
			"ssid": schema.StringAttribute{
				Description: "Wi-Fi SSID the network is identified by.",
				Computed:    true,
			},
			"dns_search_domains":        stringList("DNS search domains."),
			"dns_server_ips":            stringList("DNS server IPs."),
			"resolved_ips_for_hostname": stringList("Resolved IPs for the configured hostnames."),
			"trusted_dhcp_servers_ips":  stringList("Trusted DHCP server IPs."),
			"trusted_egress_ips":        stringList("Trusted egress IPs."),
			"trusted_gateway_ips":       stringList("Trusted default-gateway IPs."),
			"trusted_subnet_ips":        stringList("Trusted CIDR ranges."),
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

	hasID := !data.ID.IsNull() && !data.ID.IsUnknown() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && !data.Name.IsUnknown() && data.Name.ValueString() != ""
	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing Identifier", "Either `id` or `name` must be specified on the zcc_trusted_network data source.")
		return
	}

	backend, err := tnbackend.For(ctx, d.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	var net *trusted_network_v2.TrustedNetworkV2

	if hasID {
		id := data.ID.ValueString()
		tflog.Info(ctx, "Fetching trusted network", map[string]any{"id": id, "api_version": backend.Version()})
		net, err = backend.Get(ctx, id)
	} else {
		name := data.Name.ValueString()
		tflog.Info(ctx, "Fetching trusted network", map[string]any{"network_name": name, "api_version": backend.Version()})
		net, err = backend.GetByName(ctx, name)
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read trusted network: %v", err))
		return
	}

	model := TrustedNetworkDataSourceModel{
		ID:                     types.StringValue(strconv.Itoa(net.ID)),
		CompanyID:              types.Int64Value(int64(net.CompanyID)),
		ZPAID:                  types.StringValue(net.ZPAID),
		Active:                 types.BoolValue(net.Active),
		ConditionType:          types.StringValue(net.ConditionType),
		Name:                   types.StringValue(net.Name),
		CreatedBy:              types.StringValue(net.CreatedBy),
		EditedBy:               types.StringValue(net.EditedBy),
		Guid:                   types.StringValue(net.Guid),
		Hostname:               types.StringValue(net.Hostname),
		SSID:                   types.StringValue(net.SSID),
		DNSSearchDomains:       helpers.StringListValue(net.DNSSearchDomains),
		DNSServerIPs:           helpers.StringListValue(net.DNSServerIPs),
		ResolvedIPsForHostname: helpers.StringListValue(net.ResolvedIPsForHostname),
		TrustedDhcpServersIPs:  helpers.StringListValue(net.TrustedDhcpServersIPs),
		TrustedEgressIPs:       helpers.StringListValue(net.TrustedEgressIPs),
		TrustedGatewayIPs:      helpers.StringListValue(net.TrustedGatewayIPs),
		TrustedSubnetIPs:       helpers.StringListValue(net.TrustedSubnetIPs),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
