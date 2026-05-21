---
page_title: "zcc_zia_posture Resource - terraform-provider-zcc"
subcategory: "Posture"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/about-device-posture-profile
  Manages a ZIA device-posture profile (per-platform trust-tier criteria) used by Zscaler Client Connector.
---

# zcc_zia_posture (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-device-posture-profile)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages a **ZIA posture profile** through `/zcc/papi/public/v2/zia-posture-profiles`. A posture profile bundles the device-trust criteria that Client Connector evaluates against the local machine and classifies the endpoint into one of three trust tiers — `high`, `medium`, or `low`.

Every tier uses the same nested shape:

```hcl
high_trust_criteria = {
  cs = [
    { cn = [{ id = "<criterion-id>", name = "Antivirus enabled" }] }, # criteria set A
    { cn = [{ id = "<criterion-id>", name = "OS patched" }] },         # criteria set B
  ]
}
```

The outer `cs` list is **OR**-ed: matching **any one** of the sets in `cs` promotes the device to that tier. The inner `cn` list inside each set is **AND**-ed: **every** criterion in `cn` must match for the parent set to satisfy.

## Example Usage

```terraform
resource "zcc_zia_posture" "corp_high_trust" {
  name     = "Corp High Trust"
  platform = "macos" # one of ios | android | windows | macos | linux

  high_trust_criteria = {
    cs = [
      {
        cn = [
          { id = "1" }, # e.g. "Antivirus enabled"
          { id = "5" }, # e.g. "Full disk encryption"
        ]
      }
    ]
  }

  medium_trust_criteria = {
    cs = [
      { cn = [{ id = "1" }] }, # antivirus only
    ]
  }

  low_trust_criteria = {
    cs = []
  }
}
```

## Schema

### Required

- `name` (String) — Operator-visible profile name. API field: `name`.
- `platform` (String) — Target operating system. One of `ios`, `android`, `windows`, `macos`, `linux` (case-insensitive). Translated to the API's numeric platform code (`1`=iOS, `2`=Android, `3`=Windows, `4`=macOS, `5`=Linux) at the SDK boundary. API field: `platform`.

### Optional

- `high_trust_criteria` (Object) — Criteria sets that promote a device to the **HIGH** trust tier. API field: `highTrustCriteria`. See [trust_criteria](#trust_criteria) below.
- `medium_trust_criteria` (Object) — Criteria sets that promote a device to the **MEDIUM** trust tier. API field: `mediumTrustCriteria`.
- `low_trust_criteria` (Object) — Criteria sets that promote a device to the **LOW** trust tier. API field: `lowTrustCriteria`.

### Read-Only

- `id` (String) — Numeric identifier of the posture profile (carried as a string per Terraform convention).

<a id="trust_criteria"></a>
### Nested Schema for the three `*_trust_criteria` blocks

Each block has a single attribute:

- `cs` (List of Object) — **OR-list** of criteria sets. Matching any one set promotes the device to that trust tier. Each entry contains:
  - `cn` (List of Object) — **AND-list** of criteria. Every entry in `cn` must match for the parent set to be considered satisfied. Each criterion contains:
    - `id` (String, **Required**) — Criterion identifier as returned by ZIA's posture-criteria catalog. API field: `id`.
    - `name` (String, Optional/Computed) — Operator-visible label. API field: `name`.
    - `udid` (String, Optional/Computed) — Optional device UDID the criterion is scoped to. API field: `udid`.

## Import

The import handler accepts either a numeric profile id or a case-insensitive profile name (resolved via the SDK's `GetByName` helper):

```shell
# By numeric id
terraform import zcc_zia_posture.corp_high_trust 12345

# By name
terraform import zcc_zia_posture.corp_high_trust "Corp High Trust"
```
