---
page_title: "zcc_predefined_ip_apps Data Source - terraform-provider-zcc"
subcategory: "Applications"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-list-of-predefined-ip-based-applications
  Looks up a ZCC predefined IP-based application by id or name.
---

# zcc_predefined_ip_apps (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-list-of-predefined-ip-based-applications)

## Example Usage

```terraform
data "zcc_predefined_ip_apps" "example" {
  name = "ExamplePredefinedApp"
}
```

## Schema

### Optional

- `id` (String)
- `name` (String)

One of `id` or `name` is required.

### Read-Only

`app_name`, `app_version`, `app_svc_id`, `active`, `uid`, IPv4/IPv6 app data blobs, zapp blobs, metadata.
