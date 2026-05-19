---
page_title: "zcc_forwarding_profile Data Source - terraform-provider-zcc"
subcategory: "Forwarding"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-forwarding-profiles-by-company
  Looks up a ZCC forwarding profile by id or name.
---

# zcc_forwarding_profile (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-forwarding-profiles-by-company)

## Example Usage

```terraform
data "zcc_forwarding_profile" "example" {
  name = "road-warrior"
}
```

## Schema

### Optional

- `id` (String)
- `name` (String)

One of `id` or `name` is required.

### Read-Only

Forwarding profile attributes (active, condition_type, DNS, trusted network lists, flags, etc.).
