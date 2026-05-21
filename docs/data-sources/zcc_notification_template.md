---
page_title: "zcc_notification_template Data Source - terraform-provider-zcc"
subcategory: "Notifications"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-end-user-notifications
  Looks up a ZCC notification template by numeric id or by name.
---

# zcc_notification_template (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-end-user-notifications)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Retrieves a ZCC notification template from `/zcc/papi/public/v2/notification-templates` either by **numeric id** or by **case-insensitive name**. Unlike the resource, this data source also surfaces server-side metadata (`created_by`, `edited_by`).

## Example Usage

```terraform
data "zcc_notification_template" "by_name" {
  name = "corp-default"
}

data "zcc_notification_template" "by_id" {
  id = "12345"
}

output "default_template_id" {
  value = data.zcc_notification_template.by_name.id
}
```

## Schema

### Optional

- `id` (String) — Numeric identifier (carried as a string). Either `id` or `name` must be set.
- `name` (String) — Operator-visible template name. Either `id` or `name` must be set.

### Read-Only

- `is_default_template` (Boolean) — Whether this template is the company default.
- `enable_client` (Boolean) — Generic Client Connector notifications enabled.
- `enable_zia` (Boolean) — Master ZIA-notifications switch.
- `enable_app_updates` (Boolean) — App-update notifications enabled.
- `enable_service_status` (Boolean) — Service-status notifications enabled.
- `duration_in_seconds` (Number) — Toast notification display duration in seconds.
- `enable_persistent` (Boolean) — Whether notifications stay until dismissed.
- `enable_do_not_disturb` (Boolean) — Whether the template honours OS Do-Not-Disturb.
- `created_by` (Number) — Numeric user id of the operator who created the template.
- `edited_by` (Number) — Numeric user id of the operator who last edited the template.
- `zia_notification_template` (Object) — Per-channel ZIA notification toggles (same fields as on the matching resource): `enable_zia_firewall`, `enable_zia_firewall_popup`, `enable_zia_dns`, `enable_zia_dns_popup`, `enable_zia_ips`, `enable_zia_ips_popup`, `enable_zia_persistent`.
- `zpa_notification_template` (Object) — Per-channel ZPA notification toggles: `enable_device_posture_failure`, `enable_zpa_reauth`, `zpa_reauth_interval_in_minutes`, `delay_posture_failure_seconds`.
