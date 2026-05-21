---
page_title: "zcc_web_app_service Resource - terraform-provider-zcc"
subcategory: "Web App Service"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass
  Manages an existing ZCC web app service (application bypass entry) — supports update only, no create / delete.
---

# zcc_web_app_service (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages a ZCC **web app service** (also called an application bypass entry). This is an **update-only / singleton-style** resource — the underlying API does not expose create or delete verbs, so the provider:

- `create` looks up the existing record by `app_name` and issues a `PUT` with the merged plan + remote payload.
- `update` issues another `PUT` with the merged payload.
- `delete` removes the resource from Terraform state only; the upstream record is left untouched.

Use this resource to tune IPv4 / IPv6 bypass entries (`app_data_blob`, `app_data_blob_v6`) and the `active` toggle for a named bypass app that already exists in the tenant.

## Example Usage

```terraform
resource "zcc_web_app_service" "office365" {
  app_name = "Office 365"
  active   = true

  app_data_blob = [
    {
      proto  = "TCP"
      port   = "443"
      ipaddr = "13.107.6.152/31"
      fqdn   = "*.office.com"
    },
  ]
}
```

## Schema

### Required

- `app_name` (String) — Name of the bypass app to manage. Must already exist in the tenant — used as the lookup key during initial create.

### Optional

- `active` (Boolean) — Whether the bypass entry is active.
- `app_data_blob` (List of Object) — IPv4 application data entries (see [Nested Schema](#nested-app_data_blob) below).
- `app_data_blob_v6` (List of Object) — IPv6 application data entries (same shape).
- `zapp_data_blob` (String) — Serialized IPv4 bypass payload. Computed by the API; you can override it but most users let it default to the API's serialization of `app_data_blob`.
- `zapp_data_blob_v6` (String) — Serialized IPv6 bypass payload (same notes as above).

### Read-Only

- `id` (String) — Numeric identifier of the web app service (carried as a string).
- `uid` (String) — Stable string identifier returned by the API.
- `app_svc_id` (Number) — Internal application-service identifier.
- `app_version` (Number) — Application version returned by the API.
- `version` (Number) — Record version (incremented on each update).
- `created_by` (String) — Administrator who created the entry.
- `edited_by` (String) — Administrator who last edited the entry.
- `edited_timestamp` (String) — Timestamp of the last edit.

<a id="nested-app_data_blob"></a>
### Nested Schema for `app_data_blob` / `app_data_blob_v6`

Each entry is an object with the following Optional / Computed attributes:

- `proto` (String) — Transport protocol (for example `TCP`, `UDP`).
- `port` (String) — Port or range (the API stores ports as strings).
- `ipaddr` (String) — IP address or CIDR.
- `fqdn` (String) — Fully-qualified domain name or wildcard pattern.

## Import

The import handler accepts either a numeric ID or the app's name:

```shell
terraform import zcc_web_app_service.office365 12345
# or
terraform import zcc_web_app_service.office365 "Office 365"
```
