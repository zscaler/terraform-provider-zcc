---
page_title: "zcc_device_cleanup Resource - terraform-provider-zcc"
subcategory: "Device Cleanup"
description: |-
  Official documentation: https://help.zscaler.com/legacy-apis/public-api-controller
  Manages the singleton ZCC device cleanup settings (force removal, exceed-limit threshold, auto-removal / auto-purge windows).
---

# zcc_device_cleanup (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

The **`zcc_device_cleanup`** resource manages how the ZCC service ages-out and removes dormant or oversized client devices from a tenant. The underlying API is a **singleton** — exactly one record exists per company — and the upstream service only exposes `GET` and `PUT`. As a consequence:

- `create` is implemented as an initial `PUT` against the existing singleton record.
- `update` issues a `PUT`.
- `delete` removes the resource from Terraform state only; the upstream record is left untouched (the API does not support deletion).

Import the singleton with `terraform import` if you want to bring an already-configured tenant under management.

## Example Usage

```terraform
resource "zcc_device_cleanup" "this" {
  active              = true
  force_remove_type   = "0"   # Restrict
  device_exceed_limit = 5
  auto_removal_days   = 90
  auto_purge_days     = 180
}
```

## Schema

### Optional

- `active` (Boolean) — Whether device cleanup is active for the tenant. Mirrors the API's `0`/`1` integer-as-string flag.
- `force_remove_type` (String) — Force-remove behavior code. One of `0` (Restrict), `8`, `9`, `10`, `11`, `12`, `13`, `14`, `15`, `16`. Higher numeric codes correspond to progressively stronger removal policies; see the ZCC portal "Device Cleanup" UI for the labeled mapping.
- `device_exceed_limit` (Number) — Threshold at which the tenant is considered over its per-user enrolled-device limit. The cleanup engine consults this value when enforcing `force_remove_type`.
- `auto_removal_days` (Number) — How many days an inactive device may remain enrolled before being automatically removed. Allowed values: `0` (Never), `30`, `60`, `90`, `120`, `150`, `180`.
- `auto_purge_days` (Number) — How many days a removed device record is retained before being purged. Allowed values: `0`, `30`, `60`, `90`, `120`, `150`, `180`.

### Read-Only

- `id` (String) — Synthetic singleton identifier returned by the API.

## Import

```shell
terraform import zcc_device_cleanup.this <id>
```

The ID can be any value — the import handler always re-reads the singleton record and stores the API-reported `id`.
