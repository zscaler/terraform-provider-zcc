---
page_title: "zcc_custom_ip_apps Data Source - terraform-provider-zcc"
subcategory: "Applications"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-custom-ip-based-application-using-app-id
  Looks up a ZCC custom IP-based application by id or name.
---

# zcc_custom_ip_apps (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-custom-ip-based-application-using-app-id)

## Example Usage

```terraform
data "zcc_custom_ip_apps" "example" {
  name = "MyCustomApp"
}
```

## Schema

### Optional

- `id` (String)
- `name` (String)

One of `id` or `name` is required.

### Read-Only

`app_name`, `active`, `uid`, `app_data_blob` / `app_data_blob_v6` (nested proto/port/ipaddr/fqdn), metadata.
