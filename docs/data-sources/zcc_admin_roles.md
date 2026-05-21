---
page_title: "zcc_admin_roles Data Source - terraform-provider-zcc"
subcategory: "Administration"
description: |-
  Official documentation: https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/getAdminRoles-get
  Looks up a ZCC admin role by id or by role name and surfaces the per-feature visibility codes.
---

# zcc_admin_roles (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-admin-roles)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/getAdminRoles-get)

Retrieves a ZCC admin role by **numeric id** or by **role name**. The role record carries the per-feature visibility / permission codes used by the ZCC admin portal (each `*_profile` and feature field is returned as a string code: typically `NONE`, `VIEW_ONLY`, or `FULL` — refer to the legacy API documentation for the full code table).

## Example Usage

```terraform
data "zcc_admin_roles" "super_admin" {
  role_name = "Super Admin"
}

output "super_admin_id" {
  value = data.zcc_admin_roles.super_admin.id
}
```

## Schema

### Optional

- `id` (String) — Numeric admin role identifier. Either `id` or `role_name` must be set.
- `role_name` (String) — Display name of the admin role (for example `Super Admin`). Either `id` or `role_name` must be set.

### Read-Only

#### Identity / metadata

- `company_id` (String) — Numeric tenant identifier.
- `created_by` / `updated_by` (String) — Login of the administrator who created / last modified the role.
- `is_editable` (Boolean) — Whether the role definition can be modified through the admin UI.

#### Per-platform Client Connector profile codes

- `android_profile` (String)
- `ios_profile` (String)
- `mac_profile` (String)
- `windows_profile` (String)
- `linux_profile` (String)

#### Per-feature permission codes

Each of the following is a string permission code:

- `admin_management`, `administrator_group`
- `app_bypass`, `app_profile_group`
- `audit_logs`, `auth_setting`
- `client_connector_app_store`, `client_connector_idp`, `client_connector_notifications`, `client_connector_support`
- `dashboard`, `ddil_configuration`, `dedicated_proxy_ports`
- `device_groups`, `device_overview`, `device_posture`
- `enrolled_devices_group`, `forwarding_profile`
- `machine_tunnel`, `obfuscate_data`, `partner_device_overview`
- `public_api`, `trusted_network`, `user_agent`
- `zpa_partner_login`, `zscaler_deception`, `zscaler_entitlement`
