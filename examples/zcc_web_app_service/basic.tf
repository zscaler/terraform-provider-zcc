resource "zcc_web_app_service" "office365" {
  app_name = "Office 365"
  active   = true

  app_data_blob = [
    {
      proto  = "TCP"
      port   = "443"
      ipaddr = "13.107.6.152/31"
      fqdn   = "*.office.com"
    },
  ]
}