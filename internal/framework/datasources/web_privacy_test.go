package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

// TestAccDataSourceWebPrivacy_basic exercises the singleton
// `zcc_web_privacy` data source. The upstream `/getWebPrivacyInfo` GET
// always returns a record on a live tenant, so this test is a pure
// read-only smoke test: no resource is created, no env vars are
// required. It verifies the GET succeeds and the schema-shaped state is
// populated.
//
// The companion resource test (resources/web_privacy_test.go) does the
// fuller resource+data source pair check after toggling every flag to
// `true`; this test is intentionally narrower so it can run against any
// tenant without disturbing its configuration.
func TestAccDataSourceWebPrivacy_basic(t *testing.T) {
	dataSourceName := "data.zcc_web_privacy.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWebPrivacyConfig(),
				// Every attribute on this singleton is a Computed
				// types.Bool / types.String. We don't assert specific
				// values (the tenant state is whatever it is) — only
				// that the GET succeeded and every attribute was
				// populated, which is the contract this data source
				// exposes to consumers.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "active"),
					resource.TestCheckResourceAttrSet(dataSourceName, "collect_user_info"),
					resource.TestCheckResourceAttrSet(dataSourceName, "collect_machine_hostname"),
					resource.TestCheckResourceAttrSet(dataSourceName, "collect_zdx_location"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enable_packet_capture"),
					resource.TestCheckResourceAttrSet(dataSourceName, "disable_crashlytics"),
					resource.TestCheckResourceAttrSet(dataSourceName, "override_t2_protocol_setting"),
					resource.TestCheckResourceAttrSet(dataSourceName, "restrict_remote_packet_capture"),
					resource.TestCheckResourceAttrSet(dataSourceName, "grant_access_to_zscaler_log_folder"),
					resource.TestCheckResourceAttrSet(dataSourceName, "export_logs_for_non_admin"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enable_auto_log_snippet"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enforce_secure_pac_urls"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enable_fqdn_match_for_vpn_bypasses"),
				),
			},
		},
	})
}

func testAccDataSourceWebPrivacyConfig() string {
	return `
provider "zcc" {}

data "zcc_web_privacy" "this" {}
`
}
