package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccWebPrivacy_basic(t *testing.T) {
	resourceName := "zcc_web_privacy.this"
	dataSourceName := "data.zcc_web_privacy.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccWebPrivacyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "collect_user_info", "true"),
					resource.TestCheckResourceAttr(resourceName, "collect_machine_hostname", "true"),
					resource.TestCheckResourceAttr(resourceName, "collect_zdx_location", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_packet_capture", "true"),
					resource.TestCheckResourceAttr(resourceName, "disable_crashlytics", "true"),
					resource.TestCheckResourceAttr(resourceName, "override_t2_protocol_setting", "true"),
					resource.TestCheckResourceAttr(resourceName, "restrict_remote_packet_capture", "true"),
					resource.TestCheckResourceAttr(resourceName, "grant_access_to_zscaler_log_folder", "true"),
					resource.TestCheckResourceAttr(resourceName, "export_logs_for_non_admin", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_log_snippet", "true"),
					resource.TestCheckResourceAttr(resourceName, "enforce_secure_pac_urls", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_fqdn_match_for_vpn_bypasses", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"enforce_secure_pac_urls",
					"enable_fqdn_match_for_vpn_bypasses",
				},
			},
			{
				Config: testAccWebPrivacyWithDataSource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "active", resourceName, "active"),
					resource.TestCheckResourceAttrPair(dataSourceName, "collect_user_info", resourceName, "collect_user_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "collect_machine_hostname", resourceName, "collect_machine_hostname"),
					resource.TestCheckResourceAttrPair(dataSourceName, "collect_zdx_location", resourceName, "collect_zdx_location"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_packet_capture", resourceName, "enable_packet_capture"),
					resource.TestCheckResourceAttrPair(dataSourceName, "disable_crashlytics", resourceName, "disable_crashlytics"),
					resource.TestCheckResourceAttrPair(dataSourceName, "override_t2_protocol_setting", resourceName, "override_t2_protocol_setting"),
					resource.TestCheckResourceAttrPair(dataSourceName, "restrict_remote_packet_capture", resourceName, "restrict_remote_packet_capture"),
					resource.TestCheckResourceAttrPair(dataSourceName, "grant_access_to_zscaler_log_folder", resourceName, "grant_access_to_zscaler_log_folder"),
					resource.TestCheckResourceAttrPair(dataSourceName, "export_logs_for_non_admin", resourceName, "export_logs_for_non_admin"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_auto_log_snippet", resourceName, "enable_auto_log_snippet"),
				),
			},
		},
	})
}

func testAccWebPrivacyConfig() string {
	return `
provider "zcc" {}

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
`
}

func testAccWebPrivacyWithDataSource() string {
	return testAccWebPrivacyConfig() + `
data "zcc_web_privacy" "this" {}
`
}
