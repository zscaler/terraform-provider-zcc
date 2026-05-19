---
page_title: "zcc_web_app_service Resource - terraform-provider-zcc"
subcategory: "Applications"
description: |-
  Official documentation: https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/webAppService/listByCompany-get
  Manages a ZCC web app service (bypass app) that already exists in the tenant.
---

# zcc_web_app_service (Resource)

* [Zscaler Client Connector product documentation](https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/webAppService/listByCompany-get)

The **zcc_web_app_service** resource updates an existing web app service (bypass application). **Create** locates the service by `app_name` and applies changes; **delete** removes the resource from state only (no API delete).

## Example Usage

```terraform
resource "zcc_web_app_service" "example" {
  app_name = "MyBypassApp"
}
```

## Schema

### Required

- `app_name` (String) — Name of the existing web app service.

### Optional

- `active` (Boolean)
- `app_data_blob`, `app_data_blob_v6` — Nested blocks: `proto`, `port`, `ipaddr`, `fqdn`.
- `zapp_data_blob`, `zapp_data_blob_v6` (String)

### Read-Only

- `id`, `app_version`, `app_svc_id`, `uid`, `created_by`, `edited_by`, `edited_timestamp`, `version`

## Import

```shell
terraform import zcc_web_app_service.example <app_id_or_name>
```
