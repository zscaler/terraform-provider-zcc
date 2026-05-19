---
page_title: "zcc_process_based_apps Data Source - terraform-provider-zcc"
subcategory: "Applications"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-list-of-process-based-applications
  Looks up a ZCC process-based application by id or name.
---

# zcc_process_based_apps (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-list-of-process-based-applications)

## Example Usage

```terraform
data "zcc_process_based_apps" "example" {
  name = "ExampleProcessApp"
}
```

## Schema

### Optional

- `id` (String)
- `name` (String)

One of `id` or `name` is required.

### Read-Only

`app_name`, `file_names`, `file_paths`, `matching_criteria`, signature/certificate payloads, metadata.
