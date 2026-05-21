---
page_title: "zcc_notification_template Resource - terraform-provider-zcc"
subcategory: "Notifications"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-end-user-notifications
  Manages a ZCC end-user notification template (per-channel toggles for ZIA and ZPA notifications).
---

# zcc_notification_template (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-end-user-notifications)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages a ZCC **notification template** through `/zcc/papi/public/v2/notification-templates`. A notification template controls which end-user notifications the Client Connector raises on the endpoint — application updates, service-status changes, ZIA firewall/DNS/IPS blocks, ZPA reauthentication prompts, and posture-failure messages. Per-service toggles live under the two nested blocks `zia_notification_template` and `zpa_notification_template`.

A tenant may host multiple templates and pick one as the company default through `is_default_template`.

## Example Usage

```terraform
resource "zcc_notification_template" "corp" {
  name                  = "corp-default"
  is_default_template   = true
  enable_client         = true
  enable_zia            = true
  enable_app_updates    = true
  enable_service_status = true
  enable_persistent     = false
  enable_do_not_disturb = true
  duration_in_seconds   = 8

  zia_notification_template = {
    enable_zia_firewall       = true
    enable_zia_firewall_popup = false
    enable_zia_dns            = true
    enable_zia_dns_popup      = false
    enable_zia_ips            = true
    enable_zia_ips_popup      = true
    enable_zia_persistent     = false
  }

  zpa_notification_template = {
    enable_device_posture_failure  = true
    enable_zpa_reauth              = true
    zpa_reauth_interval_in_minutes = 60
    delay_posture_failure_seconds  = 5
  }
}
```

## Schema

### Required

- `name` (String) — Operator-visible template name. API field: `name`.

### Optional

- `is_default_template` (Boolean) — Marks this template as the company default. Only one template should hold the default flag at a time. API field: `isDefaultTemplate`.
- `enable_client` (Boolean) — Master switch for generic Client Connector notifications. API field: `enableClient`.
- `enable_zia` (Boolean) — Master switch for ZIA-driven notifications. Per-channel toggles live under `zia_notification_template`. API field: `enableZia`.
- `enable_app_updates` (Boolean) — Notify the user when ZCC receives an app update. API field: `enableAppUpdates`.
- `enable_service_status` (Boolean) — Notify on service-status changes (tunnel up/down, posture failure, etc.). API field: `enableServiceStatus`.
- `duration_in_seconds` (Number) — Duration (in seconds) transient toast notifications stay on screen. API field: `durationInSeconds`.
- `enable_persistent` (Boolean) — When `true`, notifications remain until the user dismisses them. API field: `enablePersistent`.
- `enable_do_not_disturb` (Boolean) — When `true`, the template honours the OS Do-Not-Disturb mode. API field: `enableDoNotDisturb`.
- `zia_notification_template` (Object) — Per-channel ZIA notification toggles. See [zia_notification_template](#zia_notification_template) below.
- `zpa_notification_template` (Object) — Per-channel ZPA notification toggles. See [zpa_notification_template](#zpa_notification_template) below.

### Read-Only

- `id` (String) — Numeric identifier of the template, carried as a string per Terraform convention.

<a id="zia_notification_template"></a>
### Nested Schema for `zia_notification_template`

All attributes are Optional + Computed booleans.

- `enable_zia_firewall` — Notify on ZIA firewall block events. API field: `enableZiaFirewall`.
- `enable_zia_firewall_popup` — Raise a popup (not only a tray toast) for ZIA firewall blocks. API field: `enableZiaFirewallPopup`.
- `enable_zia_dns` — Notify on ZIA DNS block events. API field: `enableZiaDNS`.
- `enable_zia_dns_popup` — Raise a popup for ZIA DNS blocks. API field: `enableZiaDNSPopup`.
- `enable_zia_ips` — Notify on ZIA IPS block events. API field: `enableZiaIPS`.
- `enable_zia_ips_popup` — Raise a popup for ZIA IPS blocks. API field: `enableZiaIPSPopup`.
- `enable_zia_persistent` — Keep ZIA notifications persistent until dismissed. API field: `enableZiaPersistent`.

<a id="zpa_notification_template"></a>
### Nested Schema for `zpa_notification_template`

- `enable_device_posture_failure` (Boolean) — Notify when device posture evaluation fails. API field: `enableDevicePostureFailure`.
- `enable_zpa_reauth` (Boolean) — Notify the user when ZPA reauthentication is required. API field: `enableZpaReauth`.
- `zpa_reauth_interval_in_minutes` (Number) — How often (in minutes) to remind the user to reauthenticate. API field: `zpaReauthIntervalInMinutes`.
- `delay_posture_failure_seconds` (Number) — Delay (in seconds) before reporting a posture failure to the user. API field: `delayPostureFailureSeconds`.

## Import

The import handler accepts either a numeric template ID or a case-insensitive template name (resolved through the SDK's `GetByName` helper):

```shell
# By numeric id
terraform import zcc_notification_template.corp 12345

# By name
terraform import zcc_notification_template.corp corp-default
```
