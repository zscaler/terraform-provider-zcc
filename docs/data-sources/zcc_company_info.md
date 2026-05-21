---
page_title: "zcc_company_info Data Source - terraform-provider-zcc"
subcategory: "Tenant"
description: |-
  Official documentation: https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/getCompanyInfo-get
  Reads the full ZCC company / tenant configuration record returned by GET /zcc/papi/public/v1/getCompanyInfo.
---

# zcc_company_info (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-company-information)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/getCompanyInfo-get)

Reads the full ZCC company / tenant configuration record returned by `GET /zcc/papi/public/v1/getCompanyInfo`. This is a **singleton** data source — there is exactly one record per tenant — so no filters are required.

Most of the surfaced fields are returned by the API as opaque string codes (`"0"` / `"1"` or feature codes); the data source preserves them as strings or integers, matching the wire format, so consumers can compare values without re-decoding them.

## Example Usage

```terraform
data "zcc_company_info" "this" {}

output "tenant_name" {
  value = data.zcc_company_info.this.name
}

output "is_zpa_enabled" {
  value = data.zcc_company_info.this.zpn_enabled
}
```

## Schema

### Read-Only

#### Identity

- `id` (String) — Synthetic identifier (mirrors `org_id` when present).
- `org_id` (String) — Organization identifier returned by the API.
- `master_customer_id` (String) — Master customer identifier.
- `name` (String) — Internal organization name.
- `business_name` (String) — Business display name.
- `business_contact_number` (String) — Primary business contact phone number.
- `support_admin_email` (String) — Support administrator email address.
- `version` (String) — Configuration version returned by the API.

#### Tenant feature flags

Most of these fields are returned as `"0"`/`"1"` string-encoded booleans:

- `proxy_enabled`, `zpn_enabled`, `upm_enabled`, `zad_enabled`, `dlp_enabled` — Per-service feature flags.
- `enable_deception_for_all` — Whether deception is enabled for every user.
- `send_email`, `activation_recipient`, `activation_copy` — Activation-email destinations.
- `mdm_status` — MDM enrollment flag.
- `default_auth_type` (Number) — Default authentication type code.
- `tunnel_protocol_type` — Tunnel protocol code.
- `secure_agent_basic`, `secure_agent_advanced` — Secure agent feature codes.
- `support_enabled`, `fetch_logs_for_admins_enabled`, `enable_rectify_utils`, `support_ticket_enabled`, `disable_logging_controls` (Number) — Support / logging feature toggles.

#### Tunnel & DNS configuration

- `proxy_port` (Number) — Local proxy listening port.
- `dns_cache_ttl_windows`, `dns_cache_ttl_mac`, `dns_cache_ttl_android`, `dns_cache_ttl_ios`, `dns_cache_ttl_linux` (Number) — Per-OS DNS cache TTLs in seconds.
- `zpa_client_cert_exp_in_days` (Number) — ZPA client certificate expiry in days.
- `vpn_gateway_char_limit`, `vpn_bypass_refresh_interval`, `dest_include_exclude_char_limit`, `dest_include_exclude_char_limit_for_ipv6`, `zt2_health_probe_interval` (Number) — Operational character / interval limits.
- `flow_logging_buffer_limit`, `flow_logging_time_interval` (Number) — Flow-logging tuning knobs.
- `zpa_reauth_enabled`, `zpa_auto_reauth_timeout`, `enable_zpa_auth_user_name` (Number) — ZPA reauthentication tuning.
- `enable_global_zcc_telemetry`, `telemetry_default` (Number) — Telemetry / opt-in defaults.
- `device_groups_count` (Number) — Number of configured device groups.

#### String-encoded toggles

The following fields are returned by the API as strings (`"true"`/`"false"`, `"0"`/`"1"`, or feature codes). Refer to the ZCC admin portal for the labeled meaning of each code:

- `enable_tunnel_zapp_traffic_toggle`, `machine_idp_auth`, `linux_visibility`, `registry_path_for_pac`, `use_pollset_for_socket_reactor`
- `enable_dtls_for_zpa`, `use_v8_js_engine`, `disable_parallel_ipv4_and_ipv6`, `send_64bit_build`, `use_add_ifscope_route`, `use_clear_arp_cache`, `use_dns_priority_ordering`
- `enable_browser_auth`, `enable_public_api`, `disable_reason_visibility`, `follow_routing_table`, `use_default_adapter_for_dns`
- `enable_minimum_device_cleanup_as_one`, `dns_priority_ordering_for_trusted_dns_criteria`, `machine_tunnel_posture`, `zpa_partner_login`
- `enable_flow_logger`, `posture_based_service`, `enable_posture_based_profile`, `disaster_recovery`, `zia_global_db_url_for_dr`, `enable_react_ui`, `launch_react_ui_by_default`, `dlp_notification`
- `ipv6_support_for_tunnel2`, `enable_set_proxy_on_vpn_adapters`, `disable_dns_route_exclusion`, `show_vpn_tun_notification`, `add_app_bypass_to_vpn_gateway`, `enable_zscaler_firewall`, `persistent_zscaler_firewall`, `clear_mup_cache`, `execute_gpo_update`, `enable_port_based_zpa_filter`, `enable_anti_tampering`, `configure_tunnel2_fallback_for_zia`
- `enable_install_webview2`, `enable_custom_proxy_ports`, `intercept_zia_traffic_all_adapters`, `swagger_link`
- `enable_one_id_admin`, `enable_one_id_user`, `restrict_admin_access`, `enable_zia_user_department_sync`, `enable_udp_transport_selection`
- `compute_device_groups_for_zia`, `compute_device_groups_for_zpa`, `compute_device_groups_for_zdx`, `compute_device_groups_for_zad`, `use_tunnel2_sme_for_tunnel1`
- `ma_cloud_name`, `zia_cloud_name`, `zdx_manual_rollout`, `win_zdx_lite_enabled`

#### Policy activation / autofill

- `policy_activation_required`, `enable_autofill_username`, `auto_fill_using_login_hint`, `dc_service_read_only` (Number)

#### Posture frequency

- `device_posture_frequency` (List of Object) — Per-platform posture-evaluation interval overrides. Each entry contains:
  - `posture_id` (Number)
  - `posture_name` (String)
  - `ios_value`, `android_value`, `windows_value`, `mac_value`, `linux_value`, `default_value` (Number)

#### Web app config (UI visibility)

- `web_app_config` (Object) — Roughly 185 nested **Computed** string fields, each a `"0"` / `"1"` feature-visibility toggle backing the ZCC admin portal's UI gates. The shape of this block mirrors the upstream `webAppConfig` payload. Fields cover device-cleanup options, app-bypass visibility, posture-related UI elements, flow-logger options, partner-device handling, and a long tail of feature-flag toggles. Refer to the ZCC admin portal source-of-truth or the SDK definition (`zscaler-sdk-go/v3/zscaler/zcc/services/company`) for the complete attribute list and their portal labels.
