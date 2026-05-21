package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccFailOpenPolicy_basic(t *testing.T) {
	resourceName := "zcc_failopen_policy.this"
	dataSourceName := "data.zcc_failopen_policy.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccFailOpenPolicyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel_failure_retry_count"),
					resource.TestCheckResourceAttrSet(resourceName, "captive_portal_web_sec_disable_minutes"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFailOpenPolicyWithDataSource(),
				// Only pair attributes that share the same tfsdk type
				// between the resource (Bool toggles + Int64 counters)
				// and the data source (String toggles + Int64 counters).
				// `active` / `enable_*` are Bool on the resource and
				// String on the data source, so they cannot be paired
				// directly without normalisation.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "captive_portal_web_sec_disable_minutes", resourceName, "captive_portal_web_sec_disable_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tunnel_failure_retry_count", resourceName, "tunnel_failure_retry_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "strict_enforcement_prompt_delay_minutes", resourceName, "strict_enforcement_prompt_delay_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "strict_enforcement_prompt_message", resourceName, "strict_enforcement_prompt_message"),
				),
			},
		},
	})
}

func testAccFailOpenPolicyConfig() string {
	return `
provider "zcc" {}

resource "zcc_failopen_policy" "this" {}
`
}

func testAccFailOpenPolicyWithDataSource() string {
	return `
provider "zcc" {}

resource "zcc_failopen_policy" "this" {}

data "zcc_failopen_policy" "this" {}
`
}
