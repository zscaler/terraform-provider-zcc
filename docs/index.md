---
layout: "zscaler"
page_title: "Provider: Zscaler Client Connector (ZCC)"
description: |-
  The Zscaler Client Connector (ZCC) Terraform provider manages ZCC configuration through the ZCC public API (via zscaler-sdk-go). Use OneAPI (Zidentity) or the legacy ZCC V2 client.
---

# Zscaler Client Connector (ZCC) Provider

The **zcc** provider interacts with [Zscaler Client Connector](https://www.zscaler.com/products/zscaler-client-connector) administration APIs. Resources and data sources map to ZCC policy objects such as trusted networks, forwarding profiles, fail-open policy, and web app services.

Use the navigation for resource and data source reference pages. See [examples](https://github.com/zscaler/terraform-provider-zcc/tree/master/examples) in this repository for runnable HCL.

## Support

Refer to the [support statement](guides/support.md) and [troubleshooting guide](guides/troubleshooting.md).

## Authentication

### OneAPI (recommended)

Configure OAuth2 client credentials or private key authentication. Typical environment variables:

| Argument | Environment variable |
|----------|----------------------|
| `client_id` | `ZSCALER_CLIENT_ID` |
| `client_secret` | `ZSCALER_CLIENT_SECRET` |
| `private_key` | `ZSCALER_PRIVATE_KEY` |
| `vanity_domain` | `ZSCALER_VANITY_DOMAIN` |
| `zscaler_cloud` | `ZSCALER_CLOUD` |

Either `client_secret` **or** `private_key` is required together with `client_id` and `vanity_domain`.

### Legacy ZCC API

Set `use_legacy_client = true` (or `ZSCALER_USE_LEGACY_CLIENT=true`) and provide:

| Argument | Environment variable |
|----------|----------------------|
| `zcc_client_id` | `ZCC_CLIENT_ID` |
| `zcc_client_secret` | `ZCC_CLIENT_SECRET` |
| `zcc_cloud` | `ZCC_CLOUD` |

## Example usage

```hcl
terraform {
  required_providers {
    zcc = {
      source  = "zscaler/zcc"
      version = "~> 0.1.0"
    }
  }
}

provider "zcc" {
  # Prefer environment variables for secrets; explicit attributes are optional.
  # client_id     = var.zscaler_client_id
  # client_secret = var.zscaler_client_secret
  # vanity_domain = var.zscaler_vanity_domain
  # zscaler_cloud = var.zscaler_cloud
}
```

Hard-coding credentials in Terraform configuration is discouraged.

## Optional provider arguments

- `http_proxy` — HTTP(S) proxy (`ZSCALER_HTTP_PROXY`)
- `max_retries` — SDK retry count
- `parallelism` — reserved for bulk operations
- `request_timeout` — per-request timeout (seconds)
- `min_wait_seconds` / `max_wait_seconds` — retry backoff bounds

## Acceptance tests

Integration tests require `TF_ACC=1` and the same credentials as above. Optional variables select tenant-specific fixtures:

| Variable | Purpose |
|----------|---------|
| `TF_ACC_ZCC_WEB_APP_NAME` | Existing web app service (bypass app) name |
| `TF_ACC_ZCC_ADMIN_USER_NAME` | Admin user login for `zcc_admin_user` data source |
| `TF_ACC_ZCC_ADMIN_ROLE_NAME` | Admin role name (defaults to `Super Admin`) |
| `TF_ACC_ZCC_CUSTOM_IP_APP_NAME` | Custom IP app name |
| `TF_ACC_ZCC_PREDEFINED_IP_APP_NAME` | Predefined IP app name |
| `TF_ACC_ZCC_PROCESS_BASED_APP_NAME` | Process-based app name |
| `TF_ACC_ZCC_APPLICATION_PROFILE_NAME` | Application profile name |

Run:

```shell
TF_ACC=1 go test ./internal/framework/... -run TestAcc -timeout 120m
```
