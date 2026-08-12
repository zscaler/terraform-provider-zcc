---
page_title: "zcc_trusted_network Resource - terraform-provider-zcc"
subcategory: "Trusted Network"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-trusted-network-rule
  Manages a ZCC trusted network via the v2 trusted-networks endpoint (list-of-string IP and domain criteria).
---

# zcc_trusted_network (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-trusted-network-rule)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages a ZCC **trusted network**. A trusted network defines a set of criteria (DNS servers, search domains, hostname, SSID, DHCP / gateway / egress IPs, subnets) that the Client Connector evaluates to determine whether the endpoint is on a known corporate network. Matching networks are typically referenced by forwarding profiles (`zcc_forwarding_profile.trusted_network_ids`) to apply different forwarding behavior.

> The IP / domain criteria fields are **lists of strings** in HCL regardless of API version — the v1 comma-separated string surface is not exposed. Pass an empty list (`[]`) for any criterion you do not want to set.

## API Version Compatibility

The newer `/zcc/papi/public/v2/trusted-networks` endpoints are not yet enabled on every Zscaler tenant. This resource handles that transparently while keeping the **same HCL** on both tenant generations — there is nothing to configure:

* The first trusted-network operation probes the v2 list endpoint once per Terraform run. If the tenant serves it, all operations use v2; otherwise the provider falls back to the legacy `/zcc/papi/public/v1/webTrustedNetwork` endpoints, converting between the two wire formats internally (list criteria ↔ comma-separated strings, `ALL`/`ANY` ↔ numeric condition codes).
* When a tenant later gains the v2 endpoints, the next Terraform run picks them up automatically — state is written in the same shape on both versions, so no migration is needed.

Behavioral notes when a tenant is served by **v1**:

* `name` is **required**: the v1 create response carries no ID, so the provider resolves the new record by its network name. Names should be unique per tenant.
* Reads are paginated list scans (v1 has no GET-by-id), which can be slower on tenants with many trusted networks.
* `zpa_id` does not exist on v1 and stays empty on the data source.

## Example Usage

```terraform
resource "zcc_trusted_network" "corp_office" {
  name           = "corp-office"
  active         = true
  condition_type = "ALL" # match every criterion below; use "ANY" for OR

  hostname           = "corp.local"
  ssid               = "CorpWiFi"
  dns_search_domains = ["corp.local", "corp.internal"]
  dns_server_ips     = ["10.0.0.10", "10.0.0.11"]

  trusted_subnet_ips        = ["10.0.0.0/8", "192.168.1.0/24"]
  trusted_gateway_ips       = ["10.0.0.1"]
  trusted_dhcp_servers_ips  = ["10.0.0.5"]
  trusted_egress_ips        = ["203.0.113.10"]
  resolved_ips_for_hostname = ["10.0.0.100"]
}
```

## Schema

### Required

- `active` (Boolean) — Whether the trusted network is active. API field: `active`.
- `condition_type` (String) — Match policy across the criteria below. The API accepts `ALL` / `ANY` (or the numeric forms `"0"` / `"1"`). API field: `conditionType`.

### Optional

- `name` (String) — Display name of the trusted network. API field: `name`.
- `hostname` (String) — Hostname used to identify the network. API field: `hostname`.
- `ssid` (String) — Wi-Fi SSID the network is identified by. API field: `ssid`.
- `dns_search_domains` (List of String) — DNS search-domain suffixes to match. API field: `dnsSearchDomains`.
- `dns_server_ips` (List of String) — DNS server IPs to match. API field: `dnsServerIps`.
- `resolved_ips_for_hostname` (List of String) — IPs that the configured `hostname` resolves to (for DNS-pinning checks). API field: `resolvedIpsForHostname`.
- `trusted_dhcp_servers_ips` (List of String) — Trusted DHCP server IPs. API field: `trustedDhcpServersIps`.
- `trusted_egress_ips` (List of String) — Trusted egress IPs (NAT / public addresses observed from the network). API field: `trustedEgressIps`.
- `trusted_gateway_ips` (List of String) — Trusted default-gateway IPs. API field: `trustedGatewayIps`.
- `trusted_subnet_ips` (List of String) — Trusted CIDR ranges (for example `"192.0.2.0/24"`). API field: `trustedSubnetIps`.

### Read-Only

- `id` (String) — Numeric identifier of the trusted network (carried as a string per Terraform convention). API field: `id`.

> The API also returns a linked ZPA tenant identifier (`zpaId`) and server-side audit metadata (`companyId`, `createdBy`, `editedBy`, `guid`). These are **not** exposed on this resource — the API populates `zpaId` lazily (the create response omits it but later GETs include it), which makes it unsuitable for resource state. Read those fields from the matching `zcc_trusted_network` **data source** instead.

## Import

```shell
terraform import zcc_trusted_network.corp_office <id>
```

The handler accepts either a numeric ID or a network name that resolves through the trusted-network list API. An exact name match (case-insensitive) always wins; a partial name is accepted when it unambiguously matches a single network (for example `TrustedNetwork02` resolves `BD-TrustedNetwork02` if no other network contains that string), and an ambiguous partial name fails with the list of candidates.
