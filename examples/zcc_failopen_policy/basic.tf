resource "zcc_failopen_policy" "this" {
  active                                  = true
  captive_portal_web_sec_disable_minutes  = 5 
  enable_captive_portal_detection         = true
  enable_fail_open                        = true
  enable_strict_enforcement_prompt        = true
  enable_web_sec_on_proxy_unreachable     = true
  enable_web_sec_on_tunnel_failure        = true
  strict_enforcement_prompt_delay_minutes = 5 
  strict_enforcement_prompt_message       = "This is a test"
  tunnel_failure_retry_count              = 10
}