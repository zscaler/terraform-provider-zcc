---
page_title: "zcc_web_privacy Resource - terraform-provider-zcc"
subcategory: "Privacy"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-client-connector-privacy-options
  Manages the singleton ZCC web privacy settings (logging, packet capture, and information-collection toggles).
---

# zcc_web_privacy (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-client-connector-privacy-options)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages the ZCC **web privacy** singleton settings record. The upstream API only supports `GET` and `PUT`, so create and update both issue a `PUT` (after pre-reading the singleton record so omitted fields keep their server-side value) and delete only removes the resource from state — the record on the API is left intact.

Every toggle below maps to one of the API's `"0"`/`"1"` string-encoded booleans and is surfaced in HCL as a regular `true`/`false`.

## Example Usage

```terraform
resource "zcc_web_privacy" "this" {
  active                             = true
  collect_user_info                  = true
  collect_machine_hostname           = true
  collect_zdx_location               = false
  enable_packet_capture              = true
  disable_crashlytics                = false
  override_t2_protocol_setting       = false
  restrict_remote_packet_capture     = true
  grant_access_to_zscaler_log_folder = true
  export_logs_for_non_admin          = false
  enable_auto_log_snippet            = true
  enforce_secure_pac_urls            = true
  enable_fqdn_match_for_vpn_bypasses = true
}
```

## Schema

### Optional

- `active` (Boolean) — Master switch for the privacy settings record.
- `collect_user_info` (Boolean) — Whether ZCC may collect end-user information (login name, group memberships).
- `collect_machine_hostname` (Boolean) — Whether ZCC may collect the machine's hostname.
- `collect_zdx_location` (Boolean) — Whether ZCC may collect ZDX location metadata.
- `enable_packet_capture` (Boolean) — Whether end users (or admins) can capture packets through the Client Connector UI.
- `disable_crashlytics` (Boolean) — Disable Crashlytics-style crash telemetry from ZCC.
- `override_t2_protocol_setting` (Boolean) — Whether the device can override the company-level T2 (Tunnel 2.0) protocol setting.
- `restrict_remote_packet_capture` (Boolean) — Restrict remotely-initiated packet capture flows.
- `grant_access_to_zscaler_log_folder` (Boolean) — Grant the local user filesystem access to the Zscaler log folder.
- `export_logs_for_non_admin` (Boolean) — Allow non-admin users to export Client Connector logs.
- `enable_auto_log_snippet` (Boolean) — Enable automatic capture of relevant log snippets when an issue is detected.
- `enforce_secure_pac_urls` (Boolean) — Require PAC URLs to use TLS.
- `enable_fqdn_match_for_vpn_bypasses` (Boolean) — Enable FQDN matching for VPN bypass rules (as opposed to IP-only matching).

### Read-Only

- `id` (String) — Settings record identifier (singleton).

## Import

```shell
terraform import zcc_web_privacy.this <id>
```

The provided ID is ignored — the import handler always re-reads the singleton record and stores the API-reported identifier.
