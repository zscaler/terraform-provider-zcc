---
page_title: "zcc_app_profile_ios Resource - terraform-provider-zcc"
subcategory: "Policy"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/adds-or-updates-a-policy-or-app-profile-for-the-company-by-platform
  Adds or updates a policy or app profile for the company by platform (iOS, Android, Windows, macOS, and Linux).
---

# zcc_app_profile_ios (Resource)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-fail-open-policies-for-the-company)

The **zcc_app_profile_ios** resource Adds or updates a policy or app profile for the company by platform (iOS).

## Example Usage

```terraform
resource "zcc_app_profile_ios" "this" {}
```

## Schema

### Optional

- `active` (String)
- `captive_portal_web_sec_disable_minutes` (Number)
- `enable_captive_portal_detection`, `enable_fail_open`, `enable_strict_enforcement_prompt` (Number)
- `enable_web_sec_on_proxy_unreachable`, `enable_web_sec_on_tunnel_failure` (String)
- `strict_enforcement_prompt_delay_minutes` (Number)
- `strict_enforcement_prompt_message` (String)
- `tunnel_failure_retry_count` (Number)

### Read-Only

- `id`, `company_id`, `created_by`, `edited_by`

## Import

```shell
terraform import zcc_failopen_policy.example <id>
```

You may also use the sentinel `failopen_policy` when importing if the provider resolves the singleton automatically.
