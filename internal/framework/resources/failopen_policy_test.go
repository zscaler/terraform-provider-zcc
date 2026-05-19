package resources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccFailOpenPolicy_basic(t *testing.T) {
	resourceName := "zcc_failopen_policy.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		// Singleton: API object always remains after Terraform destroy.
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccFailOpenPolicyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "company_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFailOpenPolicyWithDataSource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.zcc_failopen_policy.this", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.zcc_failopen_policy.this", "company_id", resourceName, "company_id"),
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
