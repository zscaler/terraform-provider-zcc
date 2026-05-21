---
page_title: "zcc_web_privacy Data Source - terraform-provider-zcc"
subcategory: "Privacy"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-client-connector-privacy-options
  Reads the singleton ZCC web privacy settings.
---

# zcc_web_privacy (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-client-connector-privacy-options)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Reads the **singleton** ZCC web privacy settings record. There is exactly one record per tenant, so no filters are required.

## Example Usage

```terraform
data "zcc_web_privacy" "this" {}

output "packet_capture_enabled" {
  value = data.zcc_web_privacy.this.enable_packet_capture
}
```

## Schema

### Read-Only

- `id` (String) — Settings record identifier.
- `active` (Boolean) — Master switch for the privacy settings record.
- `collect_user_info` (Boolean) — Whether end-user information is collected.
- `collect_machine_hostname` (Boolean) — Whether the machine hostname is collected.
- `collect_zdx_location` (Boolean) — Whether ZDX location metadata is collected.
- `enable_packet_capture` (Boolean) — Whether packet capture is allowed.
- `disable_crashlytics` (Boolean) — Whether Crashlytics-style telemetry is disabled.
- `override_t2_protocol_setting` (Boolean) — Whether endpoints may override the company T2 protocol setting.
- `restrict_remote_packet_capture` (Boolean) — Whether remote packet capture is restricted.
- `grant_access_to_zscaler_log_folder` (Boolean) — Whether the local user can access the Zscaler log folder.
- `export_logs_for_non_admin` (Boolean) — Whether non-admins can export Client Connector logs.
- `enable_auto_log_snippet` (Boolean) — Whether auto log snippet capture is enabled.
- `enforce_secure_pac_urls` (Boolean) — Whether secure (HTTPS) PAC URLs are enforced.
- `enable_fqdn_match_for_vpn_bypasses` (Boolean) — Whether FQDN matching for VPN bypasses is enabled.
