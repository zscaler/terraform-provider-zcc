---
page_title: "zcc_application_profiles Data Source - terraform-provider-zcc"
subcategory: "Profiles"
description: |-
  Official documentation: https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-list-of-application-profile-policies
  Looks up a ZCC application profile by id or name.
---

# zcc_application_profiles (Data Source)

* [Zscaler Client Connector product documentation](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller/retrieves-the-list-of-application-profile-policies)

## Example Usage

```terraform
data "zcc_application_profiles" "example" {
  name = "Default Windows Profile"
}
```

## Schema

### Optional

- `id` (String)
- `name` (String)

One of `id` or `name` is required.

### Read-Only

Large profile surface: device type, PAC URL, logging, forwarding profile ID, posture IDs, group and bypass app ID lists, passwords/notification flags, IPv6 mode, nested `disaster_recovery` and `policy_extension` objects, and related ZCC fields. See the provider schema for the full attribute list.
