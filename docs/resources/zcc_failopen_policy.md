---
page_title: "zcc_failopen_policy Resource - terraform-provider-zcc"
subcategory: "Policy"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-fail-open-policies-for-the-company
  Reads the ZCC fail-open policy by optional id, or the company singleton when id i
  Manages the ZCC fail-open (web fail open) policy singleton for the company.
---

# zcc_failopen_policy (Resource)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-fail-open-policies-for-the-company)

The **zcc_failopen_policy** resource updates the company fail-open policy. The object always exists in ZCC; Terraform **create** applies desired settings to the existing singleton, and **delete** only removes the resource from state (no API delete).

## Example Usage

```terraform
resource "zcc_failopen_policy" "example" {}
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
