---
page_title: "zcc_trusted_network Data Source - terraform-provider-zcc"
subcategory: "Trusted Network"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-trusted-network-rule
  Looks up a ZCC trusted network by id or by name.
---

# zcc_trusted_network (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-trusted-network-rule)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Reads a ZCC trusted network by numeric id or by name. The API generation is detected automatically: the data source uses `/zcc/papi/public/v2/trusted-networks` where available and transparently falls back to `/zcc/papi/public/v1/webTrustedNetwork` on tenants where v2 is not yet enabled.

## Example Usage

```terraform
data "zcc_trusted_network" "corp_office" {
  name = "corp-office"
}

data "zcc_trusted_network" "by_id" {
  id = "12345"
}

output "corp_subnets" {
  value = data.zcc_trusted_network.corp_office.trusted_subnet_ips
}
```

## Schema

### Optional

- `id` (String) — Numeric identifier of the trusted network. Either `id` or `name` must be set.
- `name` (String) — Display name of the trusted network. Either `id` or `name` must be set.

### Read-Only

- `active` (Boolean) — Whether the trusted network is active.
- `condition_type` (String) — Match policy applied across the criteria below (`ALL` / `ANY`).
- `company_id` (Number) — Numeric company id the network is scoped to.
- `zpa_id` (String) — Linked ZPA tenant identifier. Only populated by the v2 API; empty on tenants served by v1.
- `guid` (String) — Stable GUID of the trusted network.
- `created_by` (String) — Administrator who created the trusted network.
- `edited_by` (String) — Administrator who last edited the trusted network.
- `hostname` (String) — Hostname used to identify the network.
- `ssid` (String) — Wi-Fi SSID the network is identified by.
- `dns_search_domains` (List of String) — DNS search-domain suffixes.
- `dns_server_ips` (List of String) — DNS server IPs.
- `resolved_ips_for_hostname` (List of String) — IPs that the configured hostname resolves to.
- `trusted_dhcp_servers_ips` (List of String) — Trusted DHCP server IPs.
- `trusted_egress_ips` (List of String) — Trusted egress IPs.
- `trusted_gateway_ips` (List of String) — Trusted default-gateway IPs.
- `trusted_subnet_ips` (List of String) — Trusted CIDR ranges.
