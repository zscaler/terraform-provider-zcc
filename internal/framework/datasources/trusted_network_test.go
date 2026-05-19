package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccDataSourceTrustedNetwork_byID(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-tn-ds-%s", acctest.RandString(8))
	resourceName := "zcc_trusted_network.this"
	dataSourceName := "data.zcc_trusted_network.by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTrustedNetworkByIDConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_name", resourceName, "network_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "trusted_subnets", resourceName, "trusted_subnets"),
				),
			},
		},
	})
}

func testAccDataSourceTrustedNetworkByIDConfig(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  network_name    = %[1]q
  active          = true
  trusted_subnets = "192.0.2.0/24"
}

data "zcc_trusted_network" "by_id" {
  id = zcc_trusted_network.this.id
}
`, name)
}
