---
page_title: "zcc_forwarding_profile Resource - terraform-provider-zcc"
subcategory: "Forwarding"
description: |-
  Official documentation: https://help.zscaler.com/zscaler-client-connector/about-forwarding-profile
  Manages a ZCC web forwarding profile (per-network forwarding actions, trusted-network references, unified-tunnel options).
---

# zcc_forwarding_profile (Resource)

[![General Availability](https://img.shields.io/badge/Lifecycle%20Stage-General%20Availability-%2345c6e8)](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)

* [Official documentation](https://help.zscaler.com/zscaler-client-connector/about-forwarding-profile)
* [Automation Hub API reference](https://automate.zscaler.com/docs/api-reference-and-guides/api-reference/zcc/public-api-controller)
* [Legacy API reference](https://help.zscaler.com/legacy-apis/public-api-controller)

Manages a ZCC **forwarding profile**. Forwarding profiles bundle:

- the **trusted-network match** (DNS servers / search domains, hostname, trusted subnets / gateways / DHCP servers / egress IPs, and references to named `zcc_trusted_network` records);
- the **per-network forwarding actions** the Client Connector applies — typically distinct entries for "on trusted network", "off trusted network", and (when unified tunnel is enabled) the merged tunnel configuration; and
- a parallel set of **ZPA forwarding actions** and a single **unified-tunnel** action.

The forwarding-profile API contract is very wide — the resource exposes every field the SDK models. The block diagrams below cover the most-used attributes; less common knobs (timeouts, probe intervals, fallback codes) are listed under each nested block.

## Example Usage

```terraform
resource "zcc_forwarding_profile" "this" {
  name                           = "RoadWarriorProfile"
  active                         = true
  condition_type                 = 1
  dns_servers                    = ""
  dns_search_domains             = ""
  enable_lwf_driver              = true
  hostname                       = ""
  resolved_ips_for_hostname      = ""
  predefined_trusted_networks    = false
  predefined_tn_all              = false
  evaluate_trusted_network       = true
  enable_split_vpn_tn            = true
  skip_trusted_criteria_match    = true
  enable_unified_tunnel          = false
  enable_all_default_adapters_tn = true
  trusted_dhcp_servers           = "10.0.0.1, 10.0.0.2"
  trusted_egress_ips             = "10.0.0.3, 10.0.0.4"
  trusted_gateways               = "10.0.0.5, 10.0.0.6"
  trusted_subnets                = "10.0.0.0/8"

  # ── ZIA Forwarding Profile Actions ──────────────────────────────────────
  forwarding_profile_actions {
    network_type                            = 0
    action_type                             = 2
    system_proxy                            = true
    custom_pac                              = ""
    enable_packet_tunnel                    = false
    primary_transport                       = 1
    dtls_timeout                            = 9
    udp_timeout                             = 9
    tls_timeout                             = 5
    mtu_for_zadapter                        = 0
    block_unreachable_domains_traffic       = true
    allow_tls_fallback                      = true
    tunnel2_fallback_type                   = 0
    send_all_dns_to_trusted_server          = false
    drop_ipv6_traffic                       = false
    redirect_web_traffic                    = false
    drop_ipv6_include_traffic_in_t2         = false
    use_tunnel2_for_proxied_web_traffic     = false
    use_tunnel2_for_unencrypted_web_traffic = false
    path_mtu_discovery                      = true
    latency_based_zen_enablement            = true
    zen_probe_interval                      = 60
    zen_probe_sample_size                   = 5
    zen_threshold_limit                     = 2
    drop_ipv6_traffic_in_ipv6_network       = false
    optimise_for_unstable_connections       = false

    system_proxy_data {
      proxy_action                = 1
      enable_auto_detect          = false
      enable_pac                  = true
      pac_url                     = ""
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = ""
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }

  forwarding_profile_actions {
    network_type                            = 1
    action_type                             = 2
    system_proxy                            = true
    custom_pac                              = ""
    enable_packet_tunnel                    = false
    primary_transport                       = 1
    dtls_timeout                            = 9
    udp_timeout                             = 9
    tls_timeout                             = 5
    mtu_for_zadapter                        = 0
    block_unreachable_domains_traffic       = true
    allow_tls_fallback                      = true
    tunnel2_fallback_type                   = 0
    send_all_dns_to_trusted_server          = false
    drop_ipv6_traffic                       = false
    redirect_web_traffic                    = false
    drop_ipv6_include_traffic_in_t2         = false
    use_tunnel2_for_proxied_web_traffic     = false
    use_tunnel2_for_unencrypted_web_traffic = false
    path_mtu_discovery                      = true
    latency_based_zen_enablement            = true
    zen_probe_interval                      = 60
    zen_probe_sample_size                   = 5
    zen_threshold_limit                     = 2
    drop_ipv6_traffic_in_ipv6_network       = false
    optimise_for_unstable_connections       = false
    is_same_as_on_trusted_network           = true

    system_proxy_data {
      proxy_action                = 1
      enable_auto_detect          = false
      enable_pac                  = true
      pac_url                     = ""
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = ""
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }

  forwarding_profile_actions {
    network_type                            = 2
    action_type                             = 2
    system_proxy                            = true
    custom_pac                              = "https://pac.zscalerbeta.net/acme.com/Test_Pac_File_01"
    enable_packet_tunnel                    = false
    primary_transport                       = 1
    dtls_timeout                            = 9
    udp_timeout                             = 9
    tls_timeout                             = 5
    mtu_for_zadapter                        = 0
    block_unreachable_domains_traffic       = true
    allow_tls_fallback                      = true
    tunnel2_fallback_type                   = 0
    send_all_dns_to_trusted_server          = false
    drop_ipv6_traffic                       = false
    redirect_web_traffic                    = false
    drop_ipv6_include_traffic_in_t2         = false
    use_tunnel2_for_proxied_web_traffic     = false
    use_tunnel2_for_unencrypted_web_traffic = false
    path_mtu_discovery                      = true
    latency_based_zen_enablement            = true
    zen_probe_interval                      = 60
    zen_probe_sample_size                   = 5
    zen_threshold_limit                     = 2
    drop_ipv6_traffic_in_ipv6_network       = false
    optimise_for_unstable_connections       = false
    is_same_as_on_trusted_network           = true

    system_proxy_data {
      proxy_action                = 1
      enable_auto_detect          = false
      enable_pac                  = true
      pac_url                     = "https://pac.zscalerbeta.net/acme.com/Test_Pac_File_01"
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = ""
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = "https://pac.zscalerbeta.net/acme.com/Test_Pac_File_01"
    }
  }

  forwarding_profile_actions {
    network_type                            = 3
    action_type                             = 2
    system_proxy                            = true
    custom_pac                              = ""
    enable_packet_tunnel                    = false
    primary_transport                       = 1
    dtls_timeout                            = 9
    udp_timeout                             = 9
    tls_timeout                             = 5
    mtu_for_zadapter                        = 0
    block_unreachable_domains_traffic       = true
    allow_tls_fallback                      = true
    tunnel2_fallback_type                   = 0
    send_all_dns_to_trusted_server          = false
    drop_ipv6_traffic                       = false
    redirect_web_traffic                    = false
    drop_ipv6_include_traffic_in_t2         = false
    use_tunnel2_for_proxied_web_traffic     = false
    use_tunnel2_for_unencrypted_web_traffic = false
    path_mtu_discovery                      = true
    latency_based_zen_enablement            = true
    zen_probe_interval                      = 60
    zen_probe_sample_size                   = 5
    zen_threshold_limit                     = 2
    drop_ipv6_traffic_in_ipv6_network       = false
    optimise_for_unstable_connections       = false
    is_same_as_on_trusted_network           = true

    system_proxy_data {
      proxy_action                = 1
      enable_auto_detect          = false
      enable_pac                  = true
      pac_url                     = ""
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = ""
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }

  # ── ZPA Forwarding Profile Actions ──────────────────────────────────────
  forwarding_profile_zpa_actions {
    network_type                       = 1
    action_type                        = 1
    primary_transport                  = 0
    dtls_timeout                       = 9
    tls_timeout                        = 5
    mtu_for_zadapter                   = 0
    send_trusted_network_result_to_zpa = false
    latency_based_server_enablement    = false
    lbs_probe_interval                 = 30
    lbs_probe_sample_size              = 5
    lbs_threshold_limit                = 1
    latency_based_server_mt_enablement = false
    is_same_as_on_trusted_network      = true

    partner_info {
      primary_transport  = 1
      allow_tls_fallback = true
      mtu_for_zadapter   = 0
    }
  }

  forwarding_profile_zpa_actions {
    network_type                       = 2
    action_type                        = 1
    primary_transport                  = 0
    dtls_timeout                       = 9
    tls_timeout                        = 5
    mtu_for_zadapter                   = 0
    send_trusted_network_result_to_zpa = false
    latency_based_server_enablement    = false
    lbs_probe_interval                 = 30
    lbs_probe_sample_size              = 5
    lbs_threshold_limit                = 1
    latency_based_server_mt_enablement = false
    is_same_as_on_trusted_network      = true

    partner_info {
      primary_transport  = 1
      allow_tls_fallback = true
      mtu_for_zadapter   = 0
    }
  }

  forwarding_profile_zpa_actions {
    network_type                       = 0
    action_type                        = 1
    primary_transport                  = 0
    dtls_timeout                       = 9
    tls_timeout                        = 5
    mtu_for_zadapter                   = 0
    send_trusted_network_result_to_zpa = false
    latency_based_server_enablement    = false
    lbs_probe_interval                 = 30
    lbs_probe_sample_size              = 5
    lbs_threshold_limit                = 1
    latency_based_server_mt_enablement = false

    partner_info {
      primary_transport  = 1
      allow_tls_fallback = true
      mtu_for_zadapter   = 0
    }
  }

  forwarding_profile_zpa_actions {
    network_type                       = 3
    action_type                        = 1
    primary_transport                  = 0
    dtls_timeout                       = 9
    tls_timeout                        = 5
    mtu_for_zadapter                   = 0
    send_trusted_network_result_to_zpa = false
    latency_based_server_enablement    = false
    lbs_probe_interval                 = 30
    lbs_probe_sample_size              = 5
    lbs_threshold_limit                = 1
    latency_based_server_mt_enablement = false
    is_same_as_on_trusted_network      = true

    partner_info {
      primary_transport  = 1
      allow_tls_fallback = true
      mtu_for_zadapter   = 0
    }
  }

  # ── Unified Tunnel ──────────────────────────────────────────────────────
  unified_tunnel {
    network_type                      = 0
    action_type_zia                   = 1
    action_type_zpa                   = 1
    primary_transport                 = 1
    dtls_timeout                      = 9
    tls_timeout                       = 5
    mtu_for_zadapter                  = 0
    allow_tls_fallback                = true
    path_mtu_discovery                = true
    optimise_for_unstable_connections = false
    tunnel2_fallback_type             = 0
    redirect_web_traffic              = false
    drop_ipv6_traffic                 = false
    drop_ipv6_traffic_in_ipv6_network = false
    block_unreachable_domains_traffic = false
    drop_ipv6_include_traffic_in_t2   = false
    send_all_dns_to_trusted_server    = false
    same_as_on_trusted                = false

    system_proxy_data {
      proxy_action                = 0
      enable_auto_detect          = false
      enable_pac                  = false
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = "0"
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }

  unified_tunnel {
    network_type                      = 1
    action_type_zia                   = 0
    action_type_zpa                   = 0
    primary_transport                 = 1
    dtls_timeout                      = 9
    tls_timeout                       = 5
    mtu_for_zadapter                  = 0
    allow_tls_fallback                = true
    path_mtu_discovery                = true
    optimise_for_unstable_connections = false
    tunnel2_fallback_type             = 0
    redirect_web_traffic              = false
    drop_ipv6_traffic                 = false
    drop_ipv6_traffic_in_ipv6_network = false
    block_unreachable_domains_traffic = false
    drop_ipv6_include_traffic_in_t2   = false
    send_all_dns_to_trusted_server    = false
    same_as_on_trusted                = false

    system_proxy_data {
      proxy_action                = 1
      enable_auto_detect          = false
      enable_pac                  = false
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = "0"
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }

  unified_tunnel {
    network_type                      = 2
    action_type_zia                   = 1
    action_type_zpa                   = 1
    primary_transport                 = 1
    dtls_timeout                      = 9
    tls_timeout                       = 5
    mtu_for_zadapter                  = 0
    allow_tls_fallback                = true
    path_mtu_discovery                = true
    optimise_for_unstable_connections = false
    tunnel2_fallback_type             = 0
    redirect_web_traffic              = false
    drop_ipv6_traffic                 = false
    drop_ipv6_traffic_in_ipv6_network = false
    block_unreachable_domains_traffic = false
    drop_ipv6_include_traffic_in_t2   = false
    send_all_dns_to_trusted_server    = false
    same_as_on_trusted                = true

    system_proxy_data {
      proxy_action                = 0
      enable_auto_detect          = false
      enable_pac                  = false
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = "0"
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }

  unified_tunnel {
    network_type                      = 3
    action_type_zia                   = 1
    action_type_zpa                   = 1
    primary_transport                 = 1
    dtls_timeout                      = 9
    tls_timeout                       = 5
    mtu_for_zadapter                  = 0
    allow_tls_fallback                = true
    path_mtu_discovery                = true
    optimise_for_unstable_connections = false
    tunnel2_fallback_type             = 0
    redirect_web_traffic              = false
    drop_ipv6_traffic                 = false
    drop_ipv6_traffic_in_ipv6_network = false
    block_unreachable_domains_traffic = false
    drop_ipv6_include_traffic_in_t2   = false
    send_all_dns_to_trusted_server    = false
    same_as_on_trusted                = true

    system_proxy_data {
      proxy_action                = 0
      enable_auto_detect          = false
      enable_pac                  = false
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = "0"
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = ""
    }
  }
}
```

## Schema

### Required

- `name` (String) — Display name of the forwarding profile. API field: `name`.

### Optional

#### Top-level

- `active` (Boolean) — Whether the profile is active. API field: `active`.
- `condition_type` (Number) — Match policy applied across the trusted-network criteria — `0` for OR semantics (`ANY`) or `1` for AND semantics (`ALL`). API field: `conditionType`.
- `dns_servers` (String) — Comma-separated DNS server IPs. API field: `dnsServers`.
- `dns_search_domains` (String) — Comma-separated DNS search domains. API field: `dnsSearchDomains`.
- `enable_lwf_driver` (Boolean) — Enable the Windows Lightweight Filter driver. API field: `enableLWFDriver`.
- `hostname` (String) — Hostname used to identify the network. API field: `hostname`.
- `resolved_ips_for_hostname` (String) — Comma-separated list of IPs the configured `hostname` resolves to. API field: `resolvedIpsForHostname`.
- `trusted_subnets` (String) — Comma-separated list of trusted CIDR ranges. API field: `trustedSubnets`.
- `trusted_gateways` (String) — Comma-separated list of trusted default-gateway IPs. API field: `trustedGateways`.
- `trusted_dhcp_servers` (String) — Comma-separated list of trusted DHCP server IPs. API field: `trustedDhcpServers`.
- `trusted_egress_ips` (String) — Comma-separated list of trusted egress (NAT/public) IPs. API field: `trustedEgressIps`.
- `predefined_trusted_networks` (Boolean) — Use predefined trusted networks instead of inline criteria. API field: `predefinedTrustedNetworks`.
- `predefined_tn_all` (Boolean) — When `predefined_trusted_networks` is set, include all predefined networks. API field: `predefinedTnAll`.
- `enable_split_vpn_tn` (Boolean) — Enable split-VPN behavior for trusted networks. API field: `enableSplitVpnTN`.
- `enable_unified_tunnel` (Boolean) — Enable unified tunnel mode (single tunnel for ZIA + ZPA). API field: `enableUnifiedTunnel`.
- `evaluate_trusted_network` (Boolean) — Whether to evaluate trusted-network criteria. API field: `evaluateTrustedNetwork`.
- `skip_trusted_criteria_match` (Boolean) — Skip trusted-criteria match (treat the device as untrusted). API field: `skipTrustedCriteriaMatch`.
- `enable_all_default_adapters_tn` (Boolean) — Enable trusted-network evaluation across all default network adapters. API field: `enableAllDefaultAdaptersTN`.
- `trusted_network_ids` (List of Number) — Numeric IDs of `zcc_trusted_network` records this profile references. API field: `trustedNetworkIds`.
- `trusted_networks` (List of String) — Names of referenced trusted networks (kept in sync with `trusted_network_ids` by the API). API field: `trustedNetworks`.
- `trusted_network_ids_selected` (List of Number) — Subset of `trusted_network_ids` currently selected in the policy editor. API field: `trustedNetworkIdsSelected`.

#### Blocks

- `forwarding_profile_actions` ([Block, repeatable](#forwarding_profile_actions)) — Per-network forwarding actions applied to ZIA traffic.
- `forwarding_profile_zpa_actions` ([Block, repeatable](#forwarding_profile_zpa_actions)) — Per-network forwarding actions applied to ZPA traffic.
- `unified_tunnel` ([Block, repeatable](#unified_tunnel)) — Unified-tunnel action used when `enable_unified_tunnel` is `true`.

### Read-Only

- `id` (String) — Numeric identifier of the forwarding profile (carried as a string).

<a id="forwarding_profile_actions"></a>
### Nested block: `forwarding_profile_actions`

Repeatable block describing the ZIA forwarding action for a given **network type** (on-trusted vs. off-trusted, etc.).

| Attribute | Type | Description |
|-----------|------|-------------|
| `network_type` | Number | Network classification this action applies to (`0` = off-trusted, `1` = on-trusted, etc.). |
| `action_type` | Number | Forwarding mode (`0` = direct, `1` = tunnel, `2` = none, etc.). |
| `primary_transport` | Number | Primary transport (`0` = TLS/ZTunnel 2.0, `1` = DTLS, etc.). |
| `dtls_timeout`, `tls_timeout`, `udp_timeout` | Number | Per-transport handshake / idle timeouts in seconds. |
| `mtu_for_zadapter` | Number | MTU applied to the Zscaler virtual adapter. |
| `tunnel2_fallback_type` | Number | Behavior when the Tunnel 2.0 transport fails. |
| `zen_probe_interval`, `zen_probe_sample_size`, `zen_threshold_limit` | Number | ZEN service-edge health probe tuning. |
| `lbs_probe_interval`, `lbs_probe_sample_size`, `lbs_threshold_limit` | Number | Latency-based ZEN selection probe tuning. |
| `custom_pac` | String | Custom PAC URL applied when `action_type` selects PAC mode. |
| `system_proxy` | Boolean | Use the system proxy. |
| `enable_packet_tunnel` | Boolean | Enable the packet-tunnel data path. |
| `block_unreachable_domains_traffic` | Boolean | Block traffic to domains the configured DNS cannot resolve. |
| `allow_tls_fallback` | Boolean | Allow TLS fallback when DTLS is unavailable. |
| `send_all_dns_to_trusted_server` | Boolean | Send all DNS queries to the trusted DNS server. |
| `drop_ipv6_traffic`, `drop_ipv6_traffic_in_ipv6_network`, `drop_ipv6_include_traffic_in_t2` | Boolean | IPv6 traffic-handling toggles. |
| `redirect_web_traffic` | Boolean | Redirect web traffic to ZCC's listening proxy. |
| `use_tunnel2_for_proxied_web_traffic` / `use_tunnel2_for_unencrypted_web_traffic` | Boolean | Tunnel 2.0 routing selectors. |
| `path_mtu_discovery` | Boolean | Enable Path MTU Discovery. |
| `latency_based_zen_enablement` / `latency_based_server_enablement` / `latency_based_server_mt_enablement` | Boolean | Enable latency-based service-edge / server selection. |
| `optimise_for_unstable_connections` | Boolean | Enable optimizations for unstable links. |
| `is_same_as_on_trusted_network` | Boolean | Inherit the configuration of the matching on-trusted action. |
| `system_proxy_data` | Nested block | System-proxy configuration (see [system_proxy_data](#system_proxy_data) below). |

<a id="forwarding_profile_zpa_actions"></a>
### Nested block: `forwarding_profile_zpa_actions`

Repeatable block describing the ZPA forwarding action.

| Attribute | Type | Description |
|-----------|------|-------------|
| `network_type` | Number | Network classification. |
| `action_type` | Number | Forwarding mode. |
| `primary_transport` | Number | Primary transport. |
| `dtls_timeout` / `tls_timeout` | Number | Handshake / idle timeouts. |
| `mtu_for_zadapter` | Number | Zadapter MTU. |
| `lbs_probe_interval` / `lbs_probe_sample_size` / `lbs_threshold_limit` | Number | Latency-based server probe tuning. |
| `send_trusted_network_result_to_zpa` | Boolean | Report trusted-network detection to the ZPA service. |
| `latency_based_server_enablement` / `latency_based_server_mt_enablement` | Boolean | Latency-based server selection toggles. |
| `is_same_as_on_trusted_network` | Boolean | Inherit the on-trusted equivalent. |
| `partner_info` | Nested block | Partner-tunnel parameters: `primary_transport`, `mtu_for_zadapter`, `allow_tls_fallback`. |

<a id="unified_tunnel"></a>
### Nested block: `unified_tunnel`

Used when `enable_unified_tunnel = true`.

| Attribute | Type | Description |
|-----------|------|-------------|
| `network_type` | Number | Network classification. |
| `action_type_zia` / `action_type_zpa` | Number | Separate ZIA / ZPA actions inside the unified tunnel. |
| `primary_transport` / `dtls_timeout` / `tls_timeout` / `mtu_for_zadapter` / `tunnel2_fallback_type` | Number | Same semantics as the ZIA action block. |
| `allow_tls_fallback`, `path_mtu_discovery`, `optimise_for_unstable_connections`, `redirect_web_traffic` | Boolean | Tunnel-wide toggles. |
| `drop_ipv6_traffic`, `drop_ipv6_traffic_in_ipv6_network`, `drop_ipv6_include_traffic_in_t2` | Boolean | IPv6 toggles. |
| `block_unreachable_domains_traffic`, `send_all_dns_to_trusted_server`, `same_as_on_trusted` | Boolean | Same semantics as the ZIA action block. |
| `system_proxy_data` | Nested block | Same as below. |

<a id="system_proxy_data"></a>
### Nested block: `system_proxy_data`

Used inside `forwarding_profile_actions` and `unified_tunnel`.

| Attribute | Type | Description |
|-----------|------|-------------|
| `proxy_action` | Number | Proxy action code. |
| `enable_auto_detect` | Boolean | Enable WPAD-style proxy auto-detection. |
| `enable_pac` | Boolean | Use a PAC file. |
| `pac_url` | String | PAC file URL when `enable_pac = true`. |
| `pac_data_path` | String | Local PAC data path. |
| `enable_proxy_server` | Boolean | Use an explicit proxy server. |
| `proxy_server_address` | String | Proxy server address. |
| `proxy_server_port` | String | Proxy server port (string per API). |
| `bypass_proxy_for_private_ip` | Boolean | Bypass the proxy for RFC1918 destinations. |
| `perform_gp_update` | Boolean | Trigger a Group Policy update after applying the proxy change. |

## Import

```shell
terraform import zcc_forwarding_profile.road_warrior <id>
```
