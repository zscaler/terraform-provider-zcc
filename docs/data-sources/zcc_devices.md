---
page_title: "zcc_devices Data Source - terraform-provider-zcc"
subcategory: "Devices"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/lists-device-details-of-enrolled-devices-of-your-organization
  Lists ZCC enrolled devices with optional filters (username, os_type, udid).
---

# zcc_devices (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/lists-device-details-of-enrolled-devices-of-your-organization)

## Example Usage

```terraform
data "zcc_devices" "windows" {
  os_type = "windows"
}
```

## Schema

### Optional

- `username` (String) — Filter by user.
- `os_type` (String) — e.g. `windows`, `mac`, `linux`, `ios`, `android`.
- `udid` (String) — When set, returns details for that device.

### Read-Only

- `id` (String) — Static placeholder `zcc_devices`.
- `devices` (List of Object) — Device attributes: agent/version, hostname, registration state, policy, tunnel, etc.
