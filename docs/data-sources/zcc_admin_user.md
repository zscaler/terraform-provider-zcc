---
page_title: "zcc_admin_user Data Source - terraform-provider-zcc"
subcategory: "Administration"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-admin-users-in-your-organization
  Looks up a ZCC admin user by id or user_name.
---

# zcc_admin_user (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-admin-users-in-your-organization)

## Example Usage

```terraform
data "zcc_admin_user" "example" {
  user_name = "admin@example.com"
}
```

## Schema

### Optional

- `id` (Number)
- `user_name` (String)

One of `id` or `user_name` is required.

### Read-Only

Account flags, role-related permissions, `company_id`, `role_name`, and related ZCC admin fields.
