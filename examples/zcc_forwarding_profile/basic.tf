resource "zcc_forwarding_profile" "beta_fw_profile" {
  name                           = "Beta_FW_Profile02"
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
    custom_pac                              = "https://pac.zscalerbeta.net/securitygeek.io/Test_Pac_File_01"
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
      pac_url                     = "https://pac.zscalerbeta.net/securitygeek.io/Test_Pac_File_01"
      enable_proxy_server         = false
      proxy_server_address        = ""
      proxy_server_port           = ""
      bypass_proxy_for_private_ip = false
      perform_gp_update           = false
      pac_data_path               = "https://pac.zscalerbeta.net/securitygeek.io/Test_Pac_File_01"
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
