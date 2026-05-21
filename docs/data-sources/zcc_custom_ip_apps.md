---
page_title: "zcc_custom_ip_apps Data Source - terraform-provider-zcc"
subcategory: "App Bypass"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass
  Looks up a ZCC custom IP-based application bypass entry by id or by name.
---

# zcc_custom_ip_apps (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Retrieves a ZCC **custom IP-based app** (tenant-defined application bypass entry) by numeric id or by name. Use [`zcc_predefined_ip_apps`](zcc_predefined_ip_apps.md) for IP-based apps shipped by Zscaler.

## Example Usage

```terraform
data "zcc_custom_ip_apps" "internal_jira" {
  name = "Internal Jira"
}

output "internal_jira_blobs" {
  value = data.zcc_custom_ip_apps.internal_jira.app_data_blob
}
```

## Schema

### Optional

- `id` (String) — Numeric identifier (carried as a string). Either `id` or `name` must be set.
- `name` (String) — Name of the custom IP-based app. Either `id` or `name` must be set.

### Read-Only

- `app_name` (String) — Name returned by the API (mirrors `name`).
- `active` (Boolean) — Whether the entry is active.
- `uid` (String) — Stable string identifier.
- `app_data_blob` (List of Object) — IPv4 application entries with attributes `proto`, `port`, `ipaddr`, `fqdn`.
- `app_data_blob_v6` (List of Object) — IPv6 application entries with the same shape.
- `zapp_data_blob` (String) — Serialized IPv4 bypass payload.
- `zapp_data_blob_v6` (String) — Serialized IPv6 bypass payload.
- `created_by` (String) — Administrator who created the entry.
- `edited_by` (String) — Administrator who last edited the entry.
- `edited_timestamp` (String) — Timestamp of the last edit.
