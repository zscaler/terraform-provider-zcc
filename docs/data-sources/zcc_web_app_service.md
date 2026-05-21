---
page_title: "zcc_web_app_service Data Source - terraform-provider-zcc"
subcategory: "Web App Service"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass
  Looks up a ZCC web app service (application bypass entry) by id or by name.
---

# zcc_web_app_service (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Retrieves a ZCC web app service (application bypass entry) by numeric id or by name.

## Example Usage

```terraform
data "zcc_web_app_service" "office365" {
  name = "Office 365"
}

output "office365_active" {
  value = data.zcc_web_app_service.office365.active
}
```

## Schema

### Optional

- `id` (String) — Numeric identifier of the bypass entry. Either `id` or `name` must be set.
- `name` (String) — Name of the bypass entry (`app_name` on the underlying API). Either `id` or `name` must be set.

### Read-Only

- `app_name` (String) — Name returned by the API (mirrors `name`).
- `active` (Boolean) — Whether the bypass entry is active.
- `uid` (String) — Stable string identifier returned by the API.
- `app_version` (Number) — Application version.
- `app_svc_id` (Number) — Internal application-service identifier.
- `version` (Number) — Record version.
- `app_data_blob` (List of Object) — IPv4 bypass entries with attributes `proto`, `port`, `ipaddr`, `fqdn`.
- `app_data_blob_v6` (List of Object) — IPv6 bypass entries with the same shape.
- `zapp_data_blob` (String) — Serialized IPv4 bypass payload.
- `zapp_data_blob_v6` (String) — Serialized IPv6 bypass payload.
- `created_by` (String) — Administrator who created the entry.
- `edited_by` (String) — Administrator who last edited the entry.
- `edited_timestamp` (String) — Timestamp of the last edit.
