---
page_title: "zcc_failopen_policy Data Source - terraform-provider-zcc"
subcategory: "Policy"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-fail-open-policy
  Reads the ZCC fail-open policy.
---

# zcc_failopen_policy (Data Source)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-fail-open-policy)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Reads the ZCC fail-open policy. When `id` is omitted the first company policy is returned (the API hosts a single policy per tenant).

## Example Usage

```terraform
data "zcc_failopen_policy" "this" {}

output "captive_portal_disable_minutes" {
  value = data.zcc_failopen_policy.this.captive_portal_web_sec_disable_minutes
}
```

## Schema

### Optional

- `id` (String) — When set, reads that policy by id. When omitted the company singleton is returned.

### Read-Only

> Note: a number of upstream fields are surfaced as strings or numeric integers (`0`/`1`) because that is how the API encodes them in the `GET` payload. The matching **resource** exposes the boolean toggles as `bool` values for readability.

- `active` (String) — Whether the fail-open policy is active (`"0"` / `"1"`).
- `company_id` (String) — Numeric company identifier the policy is scoped to.
- `created_by` (String) — Administrator who created the policy.
- `edited_by` (String) — Administrator who last edited the policy.
- `enable_fail_open` (Number) — Whether fail-open behaviour is enabled (`0` / `1`).
- `enable_captive_portal_detection` (Number) — Whether captive-portal detection is enabled (`0` / `1`).
- `captive_portal_web_sec_disable_minutes` (Number) — How many minutes web security is paused after captive-portal detection trips.
- `enable_strict_enforcement_prompt` (Number) — Whether the strict-enforcement prompt is shown to the end user (`0` / `1`).
- `strict_enforcement_prompt_message` (String) — Notification text shown to the user under strict enforcement.
- `strict_enforcement_prompt_delay_minutes` (Number) — Delay before showing the strict-enforcement prompt.
- `enable_web_sec_on_proxy_unreachable` (String) — Whether to keep web security active when the proxy is unreachable (`"0"` / `"1"`).
- `enable_web_sec_on_tunnel_failure` (String) — Whether to keep web security active on tunnel failure (`"0"` / `"1"`).
- `tunnel_failure_retry_count` (Number) — Number of times ZCC retries the tunnel before considering it failed.
