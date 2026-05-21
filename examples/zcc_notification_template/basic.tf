resource "zcc_notification_template" "this" {
  name = "BDNotificationTemplate_Dev100"
  enable_client = true
  enable_zia = true
  enable_app_updates = true
  enable_service_status = true
  duration_in_seconds = 5
  enable_persistent = true
  enable_do_not_disturb = true

  zia_notification_template = {
    enable_zia_firewall       = true
    enable_zia_firewall_popup = true
    enable_zia_dns            = true
    enable_zia_dns_popup      = true
    enable_zia_ips            = true
    enable_zia_ips_popup      = true
    enable_zia_persistent     = true
  }
  zpa_notification_template = {
    enable_device_posture_failure  = true
    enable_zpa_reauth              = true
    zpa_reauth_interval_in_minutes = 5
    delay_posture_failure_seconds  = 0
  }
}
