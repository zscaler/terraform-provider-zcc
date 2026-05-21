---
layout: "zscaler"
page_title: "Release Notes"
description: |-
  The Zscaler Client Connector (ZCC) provider Release Notes
---

# ZCC Provider: Release Notes

Track ZCC Terraform provider releases: resources, data sources, and fixes.

---
``Last updated: v0.1.0``

---


## 0.1.0 (May, 20 2026) - Initial Beta Release

### Notes

- Release date: **(May, 20 2026)**
- Supported Terraform version: **v1.x**

### Initial Release

- [PR #1](https://github.com/zscaler/terraform-provider-zcc/pull/1) - Initial Terraform Plugin Framework provider for Zscaler Client Connector (ZCC)
    - Resources:
        - `zcc_device_cleanup`:
        - `zcc_failopen_policy`:
        - `zcc_forwarding_profile`:
        - `zcc_notification_template`:
        - `zcc_trusted_network`:
        - `zcc_web_app_service`:
        - `zcc_web_privacy`:
        - `zcc_zia_posture`:

    - Data sources:
        - `zcc_admin_roles`:
        - `zcc_admin_user`:
        - `zcc_company_info`:
        - `zcc_custom_ip_apps`:
        - `zcc_device_cleanup`:
        - `zcc_devices`:
        - `zcc_failopen_policy`:
        - `zcc_forwarding_profile`:
        - `zcc_notification_template`:
        - `zcc_predefined_ip_apps`:
        - `zcc_process_based_apps`:
        - `zcc_trusted_network`:
        - `zcc_web_app_service`:
        - `zcc_web_privacy`:
        - `zcc_zia_posture`: