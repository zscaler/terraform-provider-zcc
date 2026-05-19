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

## 0.1.0 (April, xx 2026) - Initial Release

### Notes

- Release date: **(April, xx 2026)**
- Supported Terraform version: **v1.x**

### Initial Release

- [PR #1](https://github.com/zscaler/terraform-provider-zcc/pull/1) - Initial Terraform Plugin Framework provider for Zscaler Client Connector (ZCC), built on `zscaler-sdk-go/v3` v3.8.30. Resources: `zcc_trusted_network`, `zcc_forwarding_profile`, `zcc_failopen_policy`, `zcc_web_app_service`. Data sources: `zcc_trusted_network`, `zcc_forwarding_profile`, `zcc_failopen_policy`, `zcc_web_app_service`, `zcc_admin_user`, `zcc_admin_roles`, `zcc_devices`, `zcc_custom_ip_apps`, `zcc_predefined_ip_apps`, `zcc_process_based_apps`, `zcc_application_profiles`.

### Build & tooling

- Aligned `terraform-plugin-framework` to `v1.19.0` and `terraform-plugin-testing` to `v1.15.0` so the provider builds cleanly against `terraform-plugin-go v0.31.0` (Terraform 1.14+ resource configuration generation).
- Aligned `.golangci.yml` / `.golangci.toml`, `codecov.yml`, and `.goreleaser.yml` with the patterns used by `terraform-provider-zia`, scoped to `internal/` paths.

