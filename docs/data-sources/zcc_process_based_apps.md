---
page_title: "zcc_process_based_apps Data Source - terraform-provider-zcc"
subcategory: "App Bypass"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass
  Looks up a ZCC process-based application bypass entry by id or by name.
---

# zcc_process_based_apps (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-application-bypass)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Retrieves a ZCC **process-based app** — an application-bypass entry that matches on local process attributes (file name, path, code signature, certificate) instead of destination IPs.

## Example Usage

```terraform
data "zcc_process_based_apps" "teams" {
  name = "Microsoft Teams"
}

output "teams_match_criteria" {
  value = data.zcc_process_based_apps.teams.matching_criteria
}
```

## Schema

### Optional

- `id` (String) — Numeric identifier (carried as a string). Either `id` or `name` must be set.
- `name` (String) — Name of the process-based app. Either `id` or `name` must be set.

### Read-Only

- `app_name` (String) — Name returned by the API.
- `file_names` (List of String) — Process file names that match this entry (for example `teams.exe`, `Microsoft Teams`).
- `file_paths` (List of String) — Absolute file paths to match.
- `matching_criteria` (Number) — Bitfield / code describing which fields above must match for the entry to apply.
- `signature_payload` (String) — Code-signature payload used to validate the process binary.
- `certificate_payload` (String) — Certificate payload (Authenticode / codesign) the process binary must present.
- `created_by` (String) — Administrator who created the entry.
- `edited_by` (String) — Administrator who last edited the entry.
- `edited_timestamp` (String) — Timestamp of the last edit.
