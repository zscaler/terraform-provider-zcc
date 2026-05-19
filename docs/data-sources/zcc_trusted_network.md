---
page_title: "zcc_trusted_network Data Source - terraform-provider-zcc"
subcategory: "Trusted Network"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-trusted-networks-by-company
  Looks up a ZCC trusted network by id or network_name.
---

# zcc_trusted_network (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-trusted-networks-by-company)

## Example Usage

```terraform
data "zcc_trusted_network" "example" {
  network_name = "corp-office"
}
```

## Schema

### Optional

- `id` (String) — Trusted network ID.
- `network_name` (String) — Trusted network name.

One of `id` or `network_name` is required.

### Read-Only

Attributes mirror the **zcc_trusted_network** resource (`active`, `condition_type` as a number, DNS/SSID/trusted fields, `guid`, etc.).
