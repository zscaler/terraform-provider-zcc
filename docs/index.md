---
layout: "zscaler"
page_title: "Provider: Zscaler Client Connector (ZCC)"
description: |-
    The Zscaler Client Connector provider is used to interact with Zscaler Client Connector (ZCC) API
---

# Zscaler Client Connector (ZCC) Provider

The Zscaler Client Connector provider is used to interact with the ZCC API, to automate the provisioning of trusted networks, forwarding profiles, notification templates, ZIA posture profiles, fail open policy, device cleanup, web privacy and web app service (bypass) entries. The provider is intended to save time and reduce configuration errors. With this ZCC provider, DevOps teams can automate Client Connector administration and transform it into DevSecOps workflows. To use this provider, you must create ZCC API credentials in [Zidentity](https://help.zscaler.com/zidentity/what-zidentity).

Use the navigation on the left to read about the available resources.

## Support Disclaimer

-> **Disclaimer:** Please refer to our [General Support Statement](guides/support.md) before proceeding with the use of this provider. You can also refer to our [troubleshooting guide](guides/troubleshooting.md) for guidance on typical problems.

## Feature Availability and API Parity

-> **Important:** The ZCC Terraform provider maintains parity with publicly available API endpoints. In some instances, certain features or attributes available via the Zscaler UI may not be immediately available through the API, and therefore cannot be included in the Terraform provider. This does not indicate that the provider is lagging behind; rather, it reflects that we implement only the features that are currently exposed by the public API.

If there is a feature or attribute you would like to see included in the provider, you are welcome to:

- Submit a feature request via [GitHub Issues](https://github.com/zscaler/terraform-provider-zcc/issues)
- Contact Zscaler Global Support by opening a support ticket

Our team continuously works with product teams to expand API coverage and will incorporate new features into the provider as they become publicly available through the API.

## Zscaler OneAPI Framework

The ZCC Terraform Provider authenticates exclusively via [OneAPI](https://help.zscaler.com/oneapi/understanding-oneapi) OAuth2 authentication through [Zidentity](https://help.zscaler.com/zidentity/what-zidentity).

**NOTE** This provider does **not** offer a legacy authentication path — the legacy ZCC V2 API client (`zcc_client_id` / `zcc_client_secret` / `zcc_cloud`) is not supported. Tenants must be migrated to [Zidentity](https://help.zscaler.com/zidentity/what-zidentity) before they can be managed through this provider.

**NOTE** Notice that OneAPI and Zidentity are not currently supported for the following clouds: `zscalergov` and `zscalerten`.

## Examples Usage - Client Secret Authentication

```hcl
# Configure the Zscaler Client Connector Provider
terraform {
    required_providers {
        zcc = {
            version = "~> 0.1.0"
            source  = "zscaler/zcc"
        }
    }
}

# Configure the ZCC Provider (OneAPI Authentication)
#
# NOTE: Change place holder values denoted by brackets to real values, including
# the brackets.
#
# NOTE: If environment variables are utilized for provider settings the
# corresponding variable name does not need to be set in the provider config
# block.
provider "zcc" {
  client_id     = "[ZSCALER_CLIENT_ID]"
  client_secret = "[ZSCALER_CLIENT_SECRET]"
  vanity_domain = "[ZSCALER_VANITY_DOMAIN]"
  zscaler_cloud = "[ZSCALER_CLOUD]"
}
```

## Examples Usage - Private Key Authentication

```hcl
# Configure the Zscaler Client Connector Provider
terraform {
    required_providers {
        zcc = {
            version = "~> 0.1.0"
            source  = "zscaler/zcc"
        }
    }
}

# Configure the ZCC Provider (OneAPI Authentication) - Private Key
#
# NOTE: Change place holder values denoted by brackets to real values, including
# the brackets.
#
# NOTE: If environment variables are utilized for provider settings the
# corresponding variable name does not need to be set in the provider config
# block.
provider "zcc" {
  client_id     = "[ZSCALER_CLIENT_ID]"
  private_key   = "[ZSCALER_PRIVATE_KEY]"
  vanity_domain = "[ZSCALER_VANITY_DOMAIN]"
  zscaler_cloud = "[ZSCALER_CLOUD]"
}
```

**NOTE**: The `zscaler_cloud` is optional and only required when authenticating to other environments i.e `beta`

⚠️ **WARNING:** Hard-coding credentials into any Terraform configuration is not recommended, and risks secret leakage should this file be committed to public version control

For the resources and data sources examples, please check the [examples](https://github.com/zscaler/terraform-provider-zcc/tree/master/examples) directory.

## Authentication - OneAPI Framework

This provider supports authentication via the Zscaler API framework [OneAPI](https://help.zscaler.com/oneapi/understanding-oneapi).

Zscaler OneAPI uses the OAuth 2.0 authorization framework to provide secure access to Zscaler Client Connector (ZCC) APIs. OAuth 2.0 allows third-party applications to obtain controlled access to protected resources using access tokens. OneAPI uses the Client Credentials OAuth flow, in which client applications can exchange their credentials with the authorization server for an access token and obtain access to the API resources, without any user authentication involved in the process.

- [ZCC API](https://help.zscaler.com/oneapi/understanding-oneapi#:~:text=ZCC%20API)

### Default Environment variables

You can provide credentials via the `ZSCALER_CLIENT_ID`, `ZSCALER_CLIENT_SECRET`, `ZSCALER_VANITY_DOMAIN`, `ZSCALER_CLOUD` environment variables, representing your Zidentity OneAPI credentials `clientId`, `clientSecret`, `vanityDomain` and `zscaler_cloud` respectively.

| Argument        | Description                                                                                         | Environment Variable     |
|-----------------|-----------------------------------------------------------------------------------------------------|--------------------------|
| `client_id`     | _(String)_ Zscaler API Client ID, used with `clientSecret` or `PrivateKey` OAuth auth mode.         | `ZSCALER_CLIENT_ID`      |
| `client_secret` | _(String)_ Secret key associated with the API Client ID for authentication.                         | `ZSCALER_CLIENT_SECRET`  |
| `private_key`   | _(String)_ A string Private key value.                                                              | `ZSCALER_PRIVATE_KEY`    |
| `vanity_domain` | _(String)_ Refers to the domain name used by your organization.                                     | `ZSCALER_VANITY_DOMAIN`  |
| `zscaler_cloud` | _(String)_ The name of the Zidentity cloud, e.g., beta.                                             | `ZSCALER_CLOUD`          |

### Alternative OneAPI Cloud Environments

OneAPI supports authentication and can interact with alternative Zscaler environments i.e `beta`. To authenticate to these environments you must provide the following values:

| Argument        | Description                                                         | Environment Variable     |
|-----------------|---------------------------------------------------------------------|--------------------------|
| `vanity_domain` | _(String)_ Refers to the domain name used by your organization     | `ZSCALER_VANITY_DOMAIN`  |
| `zscaler_cloud` | _(String)_ The name of the Zidentity cloud i.e beta                | `ZSCALER_CLOUD`          |

For example: Authenticating to Zscaler Beta environment:

```sh
export ZSCALER_VANITY_DOMAIN="acme"
export ZSCALER_CLOUD="beta"
```

### OneAPI (API Client Scope)

OneAPI Resources are automatically created within the ZIdentity Admin UI based on the RBAC Roles applicable to APIs within the various products. For example, in ZCC, navigate to `Administration -> Role Management` and select `Add API Role`.

Once this role has been saved, return to the ZIdentity Admin UI and from the Integration menu select API Resources. Click the `View` icon to the right of Zscaler APIs and under the ZCC dropdown you will see the newly created Role. In the event a newly created role is not seen in the ZIdentity Admin UI a `Sync Now` button is provided in the API Resources menu which will initiate an on-demand sync of newly created roles.

## Argument Reference - OneAPI

Before starting with this Terraform provider you must create an API Client in the Zscaler Identity Service portal [Zidentity](https://help.zscaler.com/zidentity/what-zidentity).

- `client_id` - (Required) This is the client ID for obtaining the API token. It can also be sourced from the `ZSCALER_CLIENT_ID` environment variable.

- `client_secret` - (Optional) This is the client secret for obtaining the API token. It can also be sourced from the `ZSCALER_CLIENT_SECRET` environment variable. `client_secret` conflicts with `private_key`.

- `private_key` - (Optional) This is the private key for obtaining the API token (can be represented by a filepath, or the key itself). It can also be sourced from the `ZSCALER_PRIVATE_KEY` environment variable. `private_key` conflicts with `client_secret`. The format of the PK is PKCS#1 unencrypted (header starts with `-----BEGIN RSA PRIVATE KEY-----`) or PKCS#8 unencrypted (header starts with `-----BEGIN PRIVATE KEY-----`).

- `vanity_domain` - (Optional) This refers to the domain name used by your organization. It can also be sourced from the `ZSCALER_VANITY_DOMAIN` environment variable.

- `zscaler_cloud` - (Optional) This refers to the Zscaler cloud name where API calls will be directed to i.e `beta`. It can also be sourced from the `ZSCALER_CLOUD` environment variable.

- `http_proxy` - (Optional) This is a custom URL endpoint that can be used for unit testing or local caching proxies. Can also be sourced from the `ZSCALER_HTTP_PROXY` environment variable.

- `parallelism` - (Optional) Number of concurrent requests to make within a resource where bulk operations are not possible. The provider creates a worker pool of this size to serialize API calls. The default is `1`. Note that this is separate from Terraform's CLI `-parallelism` flag, which controls how many resources are processed concurrently (default `10`). [Learn More](https://help.zscaler.com/oneapi/understanding-rate-limiting)

- `max_retries` - (Optional) Maximum number of retries to attempt before returning an error, the default is `5`.

- `request_timeout` - (Optional) Timeout for single request (in seconds) which is made to Zscaler, the default is `0` (means no limit is set). The maximum value can be `300`.
