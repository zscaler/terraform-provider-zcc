---
page_title: "zcc_admin_user Data Source - terraform-provider-zcc"
subcategory: "Administration"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-admin-users-in-your-organization
  Looks up a ZCC admin user by id or by user name and surfaces the user's account flags and inherited role permissions.
---

# zcc_admin_user (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-admin-users-in-your-organization)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-admin-users)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-admin-users-in-your-organization)

Retrieves a ZCC admin user by **numeric id** or by **user name**. The data source returns both the user-level flags (`account_enabled`, `service_type`, etc.) and a flattened view of the user's company role — see [`zcc_admin_roles`](zcc_admin_roles.md) for the role-level meaning of each permission code.

## Example Usage

```terraform
data "zcc_admin_user" "alice" {
  user_name = "alice@corp.example"
}

output "alice_role" {
  value = data.zcc_admin_user.alice.role_name
}
```

## Schema

### Optional

- `id` (Number) — Numeric user identifier. Either `id` or `user_name` must be set.
- `user_name` (String) — Login of the admin user (for example an email). Either `id` or `user_name` must be set.

### Read-Only

#### Account flags

- `account_enabled` (String) — Whether the account is enabled (`"true"`/`"false"`).
- `company_id` (String) — Numeric tenant identifier.
- `edit_enabled` (String) — Whether the account can be edited.
- `is_default_admin` (String) — Whether the account is the tenant's default super admin.
- `service_type` (String) — Service type the admin is provisioned for.
- `updated_by` (String) — Login of the admin who last modified the record.
- `user_agent` (String) — User-agent string the API returned for the account.

#### Inherited role

- `role_name` (String) — Role name attached to the user (matches `zcc_admin_roles.role_name`).
- `is_editable` (Boolean) — Whether the role can be edited.
- `admin_management`, `administrator_group`, `audit_logs`, `auth_setting` (String) — Permission codes from the inherited role.
- `app_bypass`, `app_profile_group`, `device_groups`, `device_overview`, `device_posture`, `enrolled_devices_group`, `forwarding_profile`, `trusted_network` (String) — Object-level permission codes.
- `client_connector_app_store`, `client_connector_idp`, `client_connector_notifications`, `client_connector_support` (String) — Client Connector portal-area permissions.
- `dashboard`, `ddil_configuration`, `dedicated_proxy_ports`, `machine_tunnel`, `obfuscate_data`, `partner_device_overview`, `public_api`, `zpa_partner_login`, `zscaler_deception`, `zscaler_entitlement` (String) — Other portal-area permissions.
- `android_profile`, `ios_profile`, `mac_profile`, `windows_profile`, `linux_profile` (String) — Per-platform profile permission codes.
