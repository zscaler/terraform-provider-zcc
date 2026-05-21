resource "zcc_trusted_network" "this" {
  name = "BDTrustedNetwork_Dev100"
  # Matches GET listByCompany for this object (95359): conditionType is 0
  condition_type = "ALL"
  dns_server_ips    = ["10.11.12.13"]
  dns_search_domains   = ["acme.com"]
  hostname            = "www.acme.com"
  trusted_subnet_ips      = ["10.0.0.0/8", "20.0.0.0/8"]
  trusted_gateway_ips = ["10.0.0.1"]
  trusted_dhcp_servers_ips = ["10.0.0.2"]
  resolved_ips_for_hostname = [ "20.20.20.20"]
  trusted_egress_ips = ["10.0.0.3", "10.0.0.4"]
  active = true
}
