package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/helpers"
	"github.com/zscaler/terraform-provider-zcc/internal/framework/tnbackend"
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

// TrustedNetworkResourceModel mirrors the user-configurable subset of
// trusted_network_v2.TrustedNetworkV2.
//
// Pure server-side metadata — companyId, createdBy, editedBy, guid,
// zpaId — is intentionally NOT carried on the resource so it can't show
// up as HCL-configurable in completions or in `terraform plan`.
// Consumers that need to read those fields should use the
// `zcc_trusted_network` data source.
//
// `zpaId` is a particularly important exclusion: the ZCC API populates
// it **lazily** — the POST response on create omits it, but subsequent
// GETs return it. Carrying it on the resource breaks
// `ImportStateVerify` (post-create state has `""`; post-import state
// has the API-assigned GUID).
//
// Type notes:
//   - id is a string at the TF boundary (Terraform convention) even
//     though the SDK uses int; conversion happens via strconv.
//   - The "...IPs" / "...Domains" fields are List(String) — that's what
//     the v2 API exchanges on the wire.
//   - hostname and ssid are single strings (the v2 SDK types them as
//     scalars, not arrays).
//   - condition_type is a string ("ALL"/"ANY" or "0"/"1" per the API).
type TrustedNetworkResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Active                 types.Bool   `tfsdk:"active"`
	ConditionType          types.String `tfsdk:"condition_type"`
	Name                   types.String `tfsdk:"name"`
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

func (r *TrustedNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trusted_network"
}

func (r *TrustedNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	stringListOC := func(desc string) schema.Attribute {
		return schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			Description: desc,
		}
	}
	resp.Schema = schema.Schema{
		Description: "Manages a ZCC trusted network. The resource automatically detects which API generation the " +
			"tenant serves: it speaks the /zcc/papi/public/v2/trusted-networks endpoint where available and " +
			"transparently falls back to /zcc/papi/public/v1/webTrustedNetwork otherwise — no configuration " +
			"required. The HCL surface is identical on both: the IP/domain criteria fields are List(String) " +
			"(the v1 comma-separated string surface is not exposed); pass `[]` for criteria you do not want to set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Numeric identifier of the trusted network, carried as a string per Terraform convention. API field: id (JSON number).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the trusted network is active. API field: active.",
			},
			"condition_type": schema.StringAttribute{
				Required:    true,
				Description: "Match policy applied across the criteria below. The API accepts `ALL`/`ANY` (or the numeric forms `0`/`1`). API field: conditionType.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Server-side name. Most operators only set `network_name`; the API echoes `name` separately. API field: name.",
			},

			"hostname": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Hostname used to identify the network. API field: hostname.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Wi-Fi SSID the network is identified by. API field: ssid.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns_search_domains":        stringListOC("DNS search domains. API field: dnsSearchDomains."),
			"dns_server_ips":            stringListOC("DNS server IPs. API field: dnsServerIps."),
			"resolved_ips_for_hostname": stringListOC("Resolved IPs for the configured hostnames. API field: resolvedIpsForHostname."),
			"trusted_dhcp_servers_ips":  stringListOC("Trusted DHCP server IPs. API field: trustedDhcpServersIps."),
			"trusted_egress_ips":        stringListOC("Trusted egress IPs (NAT/public addresses observed from the network). API field: trustedEgressIps."),
			"trusted_gateway_ips":       stringListOC("Trusted default-gateway IPs. API field: trustedGatewayIps."),
			"trusted_subnet_ips":        stringListOC("Trusted CIDR ranges (e.g. \"192.0.2.0/24\"). API field: trustedSubnetIps."),
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

	backend, err := tnbackend.For(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	payload := expandTrustedNetwork(&plan)

	tflog.Info(ctx, "Creating ZCC trusted network", map[string]any{"name": payload.Name, "api_version": backend.Version()})

	created, err := backend.Create(ctx, &payload)
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

	id := state.ID.ValueString()

	backend, err := tnbackend.For(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	net, err := backend.Get(ctx, id)
	if err != nil {
		if tnbackend.IsNotFound(err) {
			tflog.Info(ctx, "Removing trusted network from state - no longer exists", map[string]any{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to read trusted network %s: %v", id, err))
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
	var state TrustedNetworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	backend, err := tnbackend.For(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}
	payload := expandTrustedNetwork(&plan)

	tflog.Info(ctx, "Updating ZCC trusted network", map[string]any{"id": id, "api_version": backend.Version()})

	updated, err := backend.Update(ctx, id, &payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to update trusted network %s: %v", id, err))
		return
	}

	flattenTrustedNetwork(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete uses the idempotent "Get-then-Delete" pattern (mirroring
// terraform-provider-zia/zia/resource_zia_email_profile.go::resourceEmailProfileDelete):
//
//  1. GET first. If the API has already removed the record (out-of-band
//     UI delete, prior sweeper run, race with a concurrent operator),
//     treat that as success and exit cleanly instead of surfacing a
//     "Record not available" / 404 error to the user.
//  2. Call DELETE. If the upstream raced us between the GET and the
//     DELETE, the resulting 404 is also treated as success — the goal
//     state ("record does not exist") has been reached either way.
//
// Both branches go through tnbackend.IsNotFound — which wraps the
// structured errorx.ErrorResponse + IsObjectNotFound() helper — rather
// than substring-matching the error message, because the ZCC endpoints
// return varied 404 bodies (e.g. `{"code":3199,"message":"Record not
// available"}`).
//
// On the v1 backend the pre-delete GET is skipped: v1 has no GET-by-id
// (reads are paginated list scans), so per the list-based-GET exception
// only the 404 on the DELETE call is tolerated.
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

	id := state.ID.ValueString()

	backend, err := tnbackend.For(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", err.Error())
		return
	}

	if backend.SupportsGetByID() {
		if _, err := backend.Get(ctx, id); err != nil {
			if tnbackend.IsNotFound(err) {
				tflog.Info(ctx, "Trusted network already removed upstream; nothing to delete", map[string]any{"id": id})
				return
			}
			tflog.Warn(ctx, "Pre-delete GET failed; proceeding to DELETE anyway", map[string]any{"id": id, "error": err.Error()})
		}
	}

	tflog.Info(ctx, "Deleting ZCC trusted network", map[string]any{"id": id, "api_version": backend.Version()})
	if err := backend.Delete(ctx, id); err != nil {
		if tnbackend.IsNotFound(err) {
			tflog.Info(ctx, "Trusted network was removed between GET and DELETE; treating as success", map[string]any{"id": id})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Failed to delete trusted network %s: %v", id, err))
	}
}

// ImportState supports two shapes:
//   - `terraform import zcc_trusted_network.this 12345` — numeric id is
//     written straight into state and the next Read fills the rest.
//   - `terraform import zcc_trusted_network.this Corp-WiFi` — looked up
//     by name through the active backend's GetByName, then the resolved
//     id is written into state. An exact (case-insensitive) match wins;
//     a partial name resolves only when it matches exactly one network,
//     and an ambiguous partial name fails listing the candidates.
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

	backend, err := tnbackend.For(ctx, r.client.Service)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	net, err := backend.GetByName(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Unable to import trusted network %q: %v", id, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(strconv.Itoa(net.ID)))...)
}

// ---------------------------------------------------------------------------
// expand: TF model → SDK struct
// ---------------------------------------------------------------------------

// expandTrustedNetwork builds a TrustedNetworkV2 from the plan. Lists
// are materialised via the shared helpers so null/unknown values cleanly
// become empty []string slices that the SDK will then omit on the wire
// via the `,omitempty` JSON tag.
func expandTrustedNetwork(plan *TrustedNetworkResourceModel) trusted_network_v2.TrustedNetworkV2 {
	return trusted_network_v2.TrustedNetworkV2{
		Active:                 plan.Active.ValueBool(),
		ConditionType:          plan.ConditionType.ValueString(),
		Name:                   plan.Name.ValueString(),
		Hostname:               plan.Hostname.ValueString(),
		SSID:                   plan.SSID.ValueString(),
		DNSSearchDomains:       helpers.StringListFromList(plan.DNSSearchDomains),
		DNSServerIPs:           helpers.StringListFromList(plan.DNSServerIPs),
		ResolvedIPsForHostname: helpers.StringListFromList(plan.ResolvedIPsForHostname),
		TrustedDhcpServersIPs:  helpers.StringListFromList(plan.TrustedDhcpServersIPs),
		TrustedEgressIPs:       helpers.StringListFromList(plan.TrustedEgressIPs),
		TrustedGatewayIPs:      helpers.StringListFromList(plan.TrustedGatewayIPs),
		TrustedSubnetIPs:       helpers.StringListFromList(plan.TrustedSubnetIPs),
	}
}

// ---------------------------------------------------------------------------
// flatten: SDK struct → TF model
// ---------------------------------------------------------------------------

// flattenTrustedNetwork copies the server's authoritative view back into
// the Terraform model. `zpaId` (and other server-only metadata) is
// deliberately not touched here — the resource model does not expose it;
// consumers read it via the matching data source.
func flattenTrustedNetwork(tn *trusted_network_v2.TrustedNetworkV2, model *TrustedNetworkResourceModel) {
	model.ID = types.StringValue(strconv.Itoa(tn.ID))
	model.Active = types.BoolValue(tn.Active)
	model.ConditionType = types.StringValue(tn.ConditionType)
	model.Name = types.StringValue(tn.Name)
	model.Hostname = types.StringValue(tn.Hostname)
	model.SSID = types.StringValue(tn.SSID)
	model.DNSSearchDomains = helpers.StringListValue(tn.DNSSearchDomains)
	model.DNSServerIPs = helpers.StringListValue(tn.DNSServerIPs)
	model.ResolvedIPsForHostname = helpers.StringListValue(tn.ResolvedIPsForHostname)
	model.TrustedDhcpServersIPs = helpers.StringListValue(tn.TrustedDhcpServersIPs)
	model.TrustedEgressIPs = helpers.StringListValue(tn.TrustedEgressIPs)
	model.TrustedGatewayIPs = helpers.StringListValue(tn.TrustedGatewayIPs)
	model.TrustedSubnetIPs = helpers.StringListValue(tn.TrustedSubnetIPs)
}
