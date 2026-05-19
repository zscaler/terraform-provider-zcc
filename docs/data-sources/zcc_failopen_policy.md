---
page_title: "zcc_failopen_policy Data Source - terraform-provider-zcc"
subcategory: "Policy"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-fail-open-policies-for-the-company
  Reads the ZCC fail-open policy by optional id, or the company singleton when id is omitted.
---

# zcc_failopen_policy (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/gets-the-list-of-fail-open-policies-for-the-company)

## Example Usage

```terraform
data "zcc_failopen_policy" "example" {}
```

## Schema

### Optional

- `id` (String) — When set, read that policy by ID; when omitted, the first company policy is returned.

### Read-Only

Policy fields: `active`, captive portal and fail-open toggles, strict enforcement settings, `company_id`, `created_by`, `edited_by`, etc.
