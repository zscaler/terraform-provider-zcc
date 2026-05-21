---
page_title: "zcc_device_cleanup Data Source - terraform-provider-zcc"
subcategory: "Device Cleanup"
description: |-
  Official documentation: https://help.zscaler.com/legacy-apis/public-api-controller
  Reads the singleton ZCC device cleanup settings.
---

# zcc_device_cleanup (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Reads the **singleton** ZCC device cleanup settings record. There is exactly one record per tenant, so no filters are required.

## Example Usage

```terraform
data "zcc_device_cleanup" "this" {}

output "auto_removal_days" {
  value = data.zcc_device_cleanup.this.auto_removal_days
}
```

## Schema

### Read-Only

- `id` (String) — Settings record identifier returned by the API.
- `active` (Boolean) — Whether device cleanup is active for the tenant.
- `force_remove_type` (String) — Force-remove behavior code (for example `"0"` for Restrict).
- `device_exceed_limit` (Number) — Threshold at which the per-user enrolled-device limit is considered exceeded.
- `auto_removal_days` (Number) — Auto-removal window in days for inactive devices.
- `auto_purge_days` (Number) — Auto-purge window in days for removed device records.
