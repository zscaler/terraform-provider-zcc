---
page_title: "zcc_devices Data Source - terraform-provider-zcc"
subcategory: "Devices"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/downloadDevices-get
  Lists ZCC enrolled devices and optionally filters by user, OS, or UDID.
---

# zcc_devices (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/downloadDevices-get)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-zcc-enrolled-devices)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/downloadDevices-get)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller#/papi/public/v1/downloadDevices-get)

Lists ZCC enrolled devices. Use the optional filters to narrow the result set; pass `udid` to fetch a single device's detail record (which goes through the dedicated device-details endpoint rather than the list endpoint).

## Example Usage

```terraform
# All devices for a user
data "zcc_devices" "alice" {
  username = "alice@corp.example"
}

# Filter by OS
data "zcc_devices" "windows_devices" {
  os_type = "windows"
}

# Fetch a single device's details by UDID
data "zcc_devices" "by_udid" {
  udid = "ABCD-1234-EFGH-5678"
}
```

## Schema

### Optional

- `username` (String) — Filter devices by enrolled user login.
- `os_type` (String) — Filter devices by OS type. Accepts the typical ZCC strings: `windows`, `mac`, `linux`, `ios`, `android`.
- `udid` (String) — Fetch the device-details record for a specific device UDID. When set, only that device is returned (and the list filters are ignored).

### Read-Only

- `id` (String) — Synthetic identifier for the data source (`"zcc_devices"`).
- `devices` (List of Object) — One entry per device. Each entry exposes:
  - `agent_version` (String) — Installed ZCC version.
  - `company_name` (String) — Tenant display name.
  - `config_download_time` (String) — Timestamp of the last successful config download.
  - `deregistration_timestamp` (String) — Timestamp of deregistration, if applicable.
  - `detail` (String) — Free-form detail string from the API.
  - `download_count` (Number) — Number of policy/config downloads.
  - `hardware_fingerprint` (String) — Hardware fingerprint computed by ZCC.
  - `keep_alive_time` (String) — Timestamp of the last keep-alive heartbeat.
  - `last_seen_time` (String) — Timestamp of the last time the device contacted ZCC.
  - `mac_address` (String) — MAC address.
  - `machine_hostname` (String) — Machine hostname.
  - `manufacturer` (String) — Hardware manufacturer.
  - `os_version` (String) — Operating-system version string.
  - `owner` (String) — Owner / enrolled user identifier.
  - `policy_name` (String) — Name of the policy / profile applied to the device.
  - `registration_state` (String) — Current registration state.
  - `registration_time` (String) — Timestamp the device was first registered.
  - `state` (String) — Current device state (e.g. active, deregistered).
  - `tunnel_version` (String) — Tunnel version reported by the agent.
  - `type` (String) — Numeric OS type code reported by the API (as a string).
  - `udid` (String) — Unique device identifier.
  - `upm_version` (String) — UPM agent version (if applicable).
  - `user` (String) — Enrolled user login.
  - `vpn_state` (String) — Current VPN state code.
  - `zapp_arch` (String) — Reported app architecture (e.g. `x64`, `arm64`).
