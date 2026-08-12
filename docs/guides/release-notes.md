---
layout: "zscaler"
page_title: "Release Notes"
description: |-
  The Zscaler Client Connector (ZCC) provider Release Notes
---

# ZCC Provider: Release Notes

Track ZCC Terraform provider releases: resources, data sources, and fixes.

---
``Last updated: v0.1.1``

---

## 0.1.1 (July 30, 2026)

### Notes

- Release date: **(July 30, 2026)**
- Supported Terraform version: **v1.x**

### Deprecations

- [PR #9](https://github.com/zscaler/terraform-provider-zcc/pull/9) - Deprecated the `parallelism` provider attribute. The attribute has no effect and will be removed in a future major release; remove it from the provider block. Rate limiting requires no configuration: when a limit is exceeded, the API returns the interval to wait and the provider retries the request automatically.

### Bug Fixes

- [PR #26](https://github.com/zscaler/terraform-provider-zcc/pull/26) - `zcc_trusted_network`: added support for tenants where the v2 `/trusted-networks` API is not yet available. The provider now detects the API generation automatically — no configuration needed — and transparently falls back to the legacy v1 `/webTrustedNetwork` endpoints using the same HCL. Import and data-source lookups by name now also resolve an unambiguous partial name on both API versions.

### Documentation

- [PR #9](https://github.com/zscaler/terraform-provider-zcc/pull/9) - Removed the `parallelism` attribute from the provider argument reference.

## 0.1.0 (May, 20 2026) - Initial Beta Release

### Notes

- Release date: **(May, 20 2026)**
- Supported Terraform version: **v1.x**

### Initial Release

- [PR #8](https://github.com/zscaler/terraform-provider-zcc/pull/8) - Initial Terraform Plugin Framework provider for Zscaler Client Connector (ZCC)
    - Resources:
        - `zcc_device_cleanup`: Manages the configuration for device cleanup.
        - `zcc_failopen_policy`: Manages FailOpen policy for the company.
        - `zcc_forwarding_profile`: Manages forwarding profiles.
        - `zcc_notification_template`: Manages notification templates
        - `zcc_trusted_network`: Manages trusted networks.
        - `zcc_web_privacy`: Adds or updates the configuration information for end user and device-related PII.
        - `zcc_zia_posture`: Manages ZIA Posture configuration

    - Data sources:
        - `zcc_admin_roles`: List of admin roles in your organization.
        - `zcc_admin_user`: List of admin users in your organization.
        - `zcc_company_info`: Retrieves information about your organization such as the name of the business, domains, etc.
        - `zcc_custom_ip_apps`: Retrieves the list of custom IP-based applications.
        - `zcc_device_cleanup`: Retrieves the configuration for device cleanup.
        - `zcc_devices`:
        - `zcc_failopen_policy`: Retrieves a specific FailOpen policy for the company.
        - `zcc_forwarding_profile`: Retrieves forwarding profiles
        - `zcc_notification_template`: Retrieves the list of notification templates
        - `zcc_predefined_ip_apps`: Retrieves the list of predefined IP-based applications.
        - `zcc_process_based_apps`: Retrieves the list of process-based applications.
        - `zcc_trusted_network`: Gets the list of trusted networks by company.
        - `zcc_web_app_service`: Gets the list of applications to bypass Zscaler such as Zoom, Microsoft Teams, etc.
        - `zcc_web_privacy`: Gets the configuration information for end user and device-related PII.
        - `zcc_zia_posture`: Gets the list of ZIA Posture configuration