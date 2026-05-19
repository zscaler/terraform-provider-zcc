---
page_title: "zcc_web_app_service Data Source - terraform-provider-zcc"
subcategory: "Applications"
description: |-
  Official documentation: https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/webAppService/listByCompany-get
  Looks up a ZCC web app service (bypass app) by id or name.
---

# zcc_web_app_service (Data Source)

* [Zscaler Client Connector product documentation](https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/webAppService/listByCompany-get)

## Example Usage

```terraform
data "zcc_web_app_service" "example" {
  name = "MyBypassApp"
}
```

## Schema

### Optional

- `id` (String)
- `name` (String)

One of `id` or `name` is required.

### Read-Only

`app_name`, `app_version`, `app_svc_id`, `active`, `uid`, blob lists, `version`, metadata fields.
