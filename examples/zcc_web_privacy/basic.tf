resource "zcc_web_privacy" "this" {
  active                             = true
  collect_user_info                  = true
  collect_machine_hostname           = true
  collect_zdx_location               = true
  enable_packet_capture              = true
  disable_crashlytics                = true
  override_t2_protocol_setting       = true
  restrict_remote_packet_capture     = true
  grant_access_to_zscaler_log_folder = true
  export_logs_for_non_admin          = true
  enable_auto_log_snippet            = true
  enforce_secure_pac_urls            = true
  enable_fqdn_match_for_vpn_bypasses = true
}
