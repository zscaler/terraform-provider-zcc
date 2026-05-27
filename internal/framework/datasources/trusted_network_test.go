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
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "trusted_subnet_ips.#", resourceName, "trusted_subnet_ips.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "trusted_subnet_ips.0", resourceName, "trusted_subnet_ips.0"),
				),
			},
		},
	})
}

func testAccDataSourceTrustedNetworkByIDConfig(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  name       		 		= %[1]q
  condition_type 			= "ALL"
  dns_server_ips    		= ["10.11.12.13"]
  dns_search_domains   		= ["acme.com"]
  hostname            		= "www.acme.com"
  trusted_subnet_ips      	= ["10.0.0.0/8", "20.0.0.0/8"]
  trusted_gateway_ips 		= ["10.0.0.1"]
  trusted_dhcp_servers_ips 	= ["10.0.0.2"]
  resolved_ips_for_hostname = [ "20.20.20.20"]
  trusted_egress_ips 		= ["10.0.0.3", "10.0.0.4"]
  active 					= true
}

data "zcc_trusted_network" "by_id" {
  id = zcc_trusted_network.this.id
}
`, name)
}
