---
page_title: "zcc_trusted_network Resource - terraform-provider-zcc"
subcategory: "Trusted Network"
description: |-
  Official documentation: hhttps://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-trusted-networks-by-company
  Manages a ZCC trusted network (web trusted network contract).
---

# zcc_trusted_network (Resource)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-trusted-networks-by-company)

The **zcc_trusted_network** resource creates, updates, and deletes ZCC trusted networks used for Client Connector trusted network evaluation.

## Example Usage

```terraform
resource "zcc_trusted_network" "example" {
  network_name    = "corp-office"
  active          = true
  # Use the same value as GET listByCompany returns (`0` and `1` are both valid in the API).
  condition_type  = 0
  trusted_subnets = "10.0.0.0/8"
}
```

## Schema

### Required

- `network_name` (String) — Trusted network display name.

### Optional

- `active` (Boolean) — Whether the trusted network is active.
- `condition_type` (Number) — Condition type from the API (`0` and `1` are both valid; see GET `listByCompany`). Omit on update to leave the remote value unchanged.
- `dns_search_domains`, `dns_servers`, `hostnames`, `resolved_ips_for_hostname`, `ssid` — Match fields. Adding or changing these in HCL triggers an in-place update.
- `trusted_dhcp_servers`, `trusted_egress_ips`, `trusted_gateways`, `trusted_subnets` — Trusted criteria strings.

### Read-Only

- `id` (String) — Trusted network identifier (set after create).
- `guid` (String) — GUID assigned by the API; sent automatically on PUT updates.

## Import

Import accepts a numeric/string ID or a lookup key resolved via the trusted network list API. After import, set **`network_name`** (and other fields as needed) in configuration to match the remote object before the next apply.

```shell
terraform import zcc_trusted_network.example <id>
```
