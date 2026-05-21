resource "zcc_device_cleanup" "this" {
  active = true
  force_remove_type = "0"
  device_exceed_limit = 14
  auto_removal_days = 150
  auto_purge_days = 150
}