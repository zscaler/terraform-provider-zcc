---
page_title: "zcc_forwarding_profile Resource - terraform-provider-zcc"
subcategory: "Forwarding"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-forwarding-profiles-by-company
  Manages a ZCC web forwarding profile.
---

# zcc_forwarding_profile (Resource)

* [Zscaler Client Connector product documentation](https://help.zscaler.com/zscaler-client-connector)

The **zcc_forwarding_profile** resource manages ZCC forwarding profiles that control how Client Connector forwards traffic and evaluates trusted networks.

## Example Usage

```terraform
resource "zcc_forwarding_profile" "example" {
  name = "road-warrior"
}
```

## Schema

### Required

- `name` (String) — Forwarding profile name.

### Optional

- `active`, `condition_type`, `dns_search_domains`, `dns_servers`, `enable_lwf_driver`, `enable_split_vpn_tn`, `evaluate_trusted_network`, `hostname`, `predefined_tn_all`, `predefined_trusted_networks`, `resolved_ips_for_hostname`, `skip_trusted_criteria_match`, `trusted_dhcp_servers`, `trusted_egress_ips`, `trusted_gateways`, `trusted_subnets` — Profile fields aligned with the ZCC API.
- `trusted_network_ids` (List of Number) — Referenced trusted network IDs.
- `trusted_networks` (List of String) — Referenced trusted network names or identifiers as strings.

### Read-Only

- `id` (String) — Forwarding profile ID (numeric string).

## Import

```shell
terraform import zcc_forwarding_profile.example <id>
```
