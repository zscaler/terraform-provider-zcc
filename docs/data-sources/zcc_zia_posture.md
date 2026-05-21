---
page_title: "zcc_zia_posture Data Source - terraform-provider-zcc"
subcategory: "Posture"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/about-device-posture-profile
  Looks up a ZIA posture profile by numeric id (name/platform lookup is temporarily disabled).
---

# zcc_zia_posture (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-device-posture-profile)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Retrieves a **ZIA posture profile** from `/zcc/papi/public/v2/zia-posture-profiles`.

> **Note**: lookup is currently restricted to the numeric `id`. The upstream `/zia-posture-profiles` **list** endpoint mishandles pagination and silently returns a truncated set, so name-based and platform-based lookup is temporarily disabled. Use the `zcc_zia_posture` resource's import handler (which calls a different list endpoint) to discover the id, then plug that id into this data source.

## Example Usage

```terraform
data "zcc_zia_posture" "by_id" {
  id = "12345"
}

output "profile_platform" {
  value = data.zcc_zia_posture.by_id.platform
}
```

## Schema

### Required

- `id` (String) — Numeric identifier of the posture profile (carried as a string). Currently the only supported lookup key.

### Read-Only

- `name` (String) — Operator-visible profile name.
- `platform` (String) — Target operating system surfaced as a name: `ios`, `android`, `windows`, `macos`, or `linux` (translated from the API's numeric platform code).
- `high_trust_criteria` (Object) — Criteria sets that promote a device to the HIGH trust tier. Shape mirrors the matching resource: `cs[].cn[].{ id, name, udid }`.
- `medium_trust_criteria` (Object) — Criteria sets that promote a device to the MEDIUM trust tier.
- `low_trust_criteria` (Object) — Criteria sets that promote a device to the LOW trust tier.

See the [zcc_zia_posture resource page](../resources/zcc_zia_posture.md) for a full description of the nested OR-of-AND structure.
