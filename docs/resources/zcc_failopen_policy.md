---
page_title: "zcc_failopen_policy Resource - terraform-provider-zcc"
subcategory: "Policy"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-fail-open-policy
  Manages the singleton ZCC fail-open policy (web security behavior on tunnel / proxy failure and captive-portal detection).
---

# zcc_failopen_policy (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-fail-open-policy)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages the ZCC **fail-open policy**. This is a **singleton** resource — the policy always exists for the tenant and cannot be created or deleted. The provider implements:

- `create` → reads the existing policy and issues a `PUT` with the merged plan + remote payload.
- `update` → `PUT` (same merge semantics).
- `delete` → removes the resource from Terraform state; the upstream policy record is left untouched.

The fail-open policy controls how Client Connector behaves when the tunnel cannot be established or when proxies are unreachable: whether to enforce strict mode (block traffic) or fail open (allow direct traffic) and how captive-portal detection should pause web-security enforcement to allow login pages to render.

## Example Usage

```terraform
resource "zcc_failopen_policy" "this" {
  active = true

  enable_fail_open                 = true
  enable_strict_enforcement_prompt = true
  strict_enforcement_prompt_message = "Network restricted: contact IT for help."
  strict_enforcement_prompt_delay_minutes = 5

  enable_web_sec_on_proxy_unreachable = true
  enable_web_sec_on_tunnel_failure    = false
  tunnel_failure_retry_count          = 3

  enable_captive_portal_detection         = true
  captive_portal_web_sec_disable_minutes  = 10
}
```

## Schema

### Optional

- `active` (Boolean) — Whether the fail-open policy is active.
- `enable_fail_open` (Boolean) — When `true`, ZCC fails open (allows direct traffic) when it cannot enforce web security; when `false`, ZCC fails closed (blocks).
- `enable_strict_enforcement_prompt` (Boolean) — Show the user a notification when strict enforcement is preventing connectivity.
- `strict_enforcement_prompt_message` (String) — Notification text shown to the user under strict enforcement.
- `strict_enforcement_prompt_delay_minutes` (Number) — Delay (in minutes) before showing the strict-enforcement prompt to avoid noise during transient outages.
- `enable_web_sec_on_proxy_unreachable` (Boolean) — Keep web security active when the configured proxy / ZIA service edge is unreachable.
- `enable_web_sec_on_tunnel_failure` (Boolean) — Keep web security active when the tunnel cannot be established.
- `tunnel_failure_retry_count` (Number) — Number of times ZCC retries the tunnel before considering it failed.
- `enable_captive_portal_detection` (Boolean) — Run captive-portal detection to allow login pages to load before web security is enforced.
- `captive_portal_web_sec_disable_minutes` (Number) — When captive-portal detection trips, how many minutes web security is paused so the user can complete the portal login.

### Read-Only

- `id` (String) — The unique identifier of the fail-open policy record.

## Import

```shell
terraform import zcc_failopen_policy.this <id>
```

The handler reads the singleton record and stores the API-reported identifier.
