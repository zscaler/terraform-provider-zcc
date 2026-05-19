---
page_title: "zcc_admin_roles Data Source - terraform-provider-zcc"
subcategory: "Administration"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/list-of-admin-roles-in-your-organization
  Looks up a ZCC admin role by id or role_name.
---

# zcc_admin_roles (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/list-of-admin-roles-in-your-organization)

## Example Usage

```terraform
data "zcc_admin_roles" "example" {
  role_name = "Super Admin"
}
```

## Schema

### Optional

- `id` (String)
- `role_name` (String)

One of `id` or `role_name` is required.

### Read-Only

RBAC capability flags (`admin_management`, profile permissions, `company_id`, `is_editable`, etc.).
