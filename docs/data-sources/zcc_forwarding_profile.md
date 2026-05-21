---
page_title: "zcc_forwarding_profile Data Source - terraform-provider-zcc"
subcategory: "Forwarding"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/about-forwarding-profile
  Looks up a ZCC forwarding profile by id or by name.
---

# zcc_forwarding_profile (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-forwarding-profile)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Retrieves the **top-level** attributes of a ZCC forwarding profile by id or by name. The data source intentionally omits the deeply nested `forwarding_profile_actions` / `forwarding_profile_zpa_actions` / `unified_tunnel` blocks — they are only available through the matching [resource](../resources/zcc_forwarding_profile.md).

## Example Usage

```terraform
data "zcc_forwarding_profile" "road_warrior" {
  name = "road-warrior"
}

output "trusted_network_refs" {
  value = data.zcc_forwarding_profile.road_warrior.trusted_network_ids
}
```

## Schema

### Optional

- `id` (String) — Numeric identifier of the forwarding profile. Either `id` or `name` must be set.
- `name` (String) — Profile name. Either `id` or `name` must be set.

### Read-Only

- `active` (Boolean) — Whether the profile is active.
- `condition_type` (Number) — Match policy across the trusted-network criteria (`0` = ANY, `1` = ALL).
- `dns_search_domains` (String) — Comma-separated DNS search domains.
- `dns_servers` (String) — Comma-separated DNS server IPs.
- `enable_lwf_driver` (Boolean) — Whether the Windows LWF driver is enabled.
- `enable_split_vpn_tn` (Boolean) — Whether split-VPN behavior for trusted networks is enabled.
- `enable_unified_tunnel` (Boolean) — Whether unified-tunnel mode is enabled.
- `evaluate_trusted_network` (Boolean) — Whether trusted-network evaluation is enabled.
- `enable_all_default_adapters_tn` (Boolean) — Whether trusted-network evaluation runs across all default adapters.
- `hostname` (String) — Hostname used to identify the network.
- `predefined_tn_all` (Boolean) — Whether all predefined trusted networks are included.
- `predefined_trusted_networks` (Boolean) — Whether predefined trusted networks are used.
- `resolved_ips_for_hostname` (String) — Comma-separated list of IPs the hostname resolves to.
- `skip_trusted_criteria_match` (Boolean) — Whether trusted-criteria matching is skipped.
- `trusted_dhcp_servers` (String) — Comma-separated trusted DHCP server IPs.
- `trusted_egress_ips` (String) — Comma-separated trusted egress IPs.
- `trusted_gateways` (String) — Comma-separated trusted default-gateway IPs.
- `trusted_network_ids` (List of Number) — Numeric IDs of referenced trusted networks.
- `trusted_networks` (List of String) — Names of referenced trusted networks.
- `trusted_subnets` (String) — Comma-separated trusted CIDR ranges.
