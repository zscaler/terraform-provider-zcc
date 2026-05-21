---
page_title: "zcc_trusted_network Resource - terraform-provider-zcc"
subcategory: "Trusted Network"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/configuring-trusted-network-rule
  Manages a ZCC trusted network via the v2 trusted-networks endpoint (list-of-string IP and domain criteria).
---

# zcc_trusted_network (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/configuring-trusted-network-rule)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages a ZCC **trusted network** through `/zcc/papi/public/v2/trusted-networks`. A trusted network defines a set of criteria (DNS servers, search domains, hostname, SSID, DHCP / gateway / egress IPs, subnets) that the Client Connector evaluates to determine whether the endpoint is on a known corporate network. Matching networks are typically referenced by forwarding profiles (`zcc_forwarding_profile.trusted_network_ids`) to apply different forwarding behavior.

> The v2 endpoint exchanges IP / domain criteria as **lists of strings** on the wire — the v1 comma-separated string surface is gone. Pass an empty list (`[]`) for any criterion you do not want to set.

## Example Usage

```terraform
resource "zcc_trusted_network" "corp_office" {
  name           = "corp-office"
  active         = true
  condition_type = "ALL" # match every criterion below; use "ANY" for OR

  hostname           = "corp.local"
  ssid               = "CorpWiFi"
  dns_search_domains = ["corp.local", "corp.internal"]
  dns_server_ips     = ["10.0.0.10", "10.0.0.11"]

  trusted_subnet_ips        = ["10.0.0.0/8", "192.168.1.0/24"]
  trusted_gateway_ips       = ["10.0.0.1"]
  trusted_dhcp_servers_ips  = ["10.0.0.5"]
  trusted_egress_ips        = ["203.0.113.10"]
  resolved_ips_for_hostname = ["10.0.0.100"]
}
```

## Schema

### Required

- `active` (Boolean) — Whether the trusted network is active. API field: `active`.
- `condition_type` (String) — Match policy across the criteria below. The API accepts `ALL` / `ANY` (or the numeric forms `"0"` / `"1"`). API field: `conditionType`.

### Optional

- `name` (String) — Display name of the trusted network. API field: `name`.
- `hostname` (String) — Hostname used to identify the network. API field: `hostname`.
- `ssid` (String) — Wi-Fi SSID the network is identified by. API field: `ssid`.
- `dns_search_domains` (List of String) — DNS search-domain suffixes to match. API field: `dnsSearchDomains`.
- `dns_server_ips` (List of String) — DNS server IPs to match. API field: `dnsServerIps`.
- `resolved_ips_for_hostname` (List of String) — IPs that the configured `hostname` resolves to (for DNS-pinning checks). API field: `resolvedIpsForHostname`.
- `trusted_dhcp_servers_ips` (List of String) — Trusted DHCP server IPs. API field: `trustedDhcpServersIps`.
- `trusted_egress_ips` (List of String) — Trusted egress IPs (NAT / public addresses observed from the network). API field: `trustedEgressIps`.
- `trusted_gateway_ips` (List of String) — Trusted default-gateway IPs. API field: `trustedGatewayIps`.
- `trusted_subnet_ips` (List of String) — Trusted CIDR ranges (for example `"192.0.2.0/24"`). API field: `trustedSubnetIps`.

### Read-Only

- `id` (String) — Numeric identifier of the trusted network (carried as a string per Terraform convention). API field: `id`.
- `zpa_id` (String) — Linked ZPA tenant identifier the API returns alongside the trusted network. API field: `zpaId`.

## Import

```shell
terraform import zcc_trusted_network.corp_office <id>
```

The handler accepts either a numeric ID or a string identifier that resolves through the trusted-network list API.
