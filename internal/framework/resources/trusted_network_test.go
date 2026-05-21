package resources_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network_v2"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccTrustedNetwork_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-tn-%s", acctest.RandString(8))
	resourceName := "zcc_trusted_network.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTrustedNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrustedNetworkConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "condition_type", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "hostname", "www.acme.com"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "dns_server_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_server_ips.0", "10.11.12.13"),
					resource.TestCheckResourceAttr(resourceName, "dns_search_domains.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_search_domains.0", "acme.com"),
					resource.TestCheckResourceAttr(resourceName, "resolved_ips_for_hostname.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resolved_ips_for_hostname.0", "20.20.20.20"),
					resource.TestCheckResourceAttr(resourceName, "trusted_subnet_ips.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "trusted_subnet_ips.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "trusted_subnet_ips.1", "20.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "trusted_gateway_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_gateway_ips.0", "10.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_dhcp_servers_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_dhcp_servers_ips.0", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "trusted_egress_ips.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "trusted_egress_ips.0", "10.0.0.3"),
					resource.TestCheckResourceAttr(resourceName, "trusted_egress_ips.1", "10.0.0.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrustedNetworkWithDataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "condition_type", resourceName, "condition_type"),
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "hostname", resourceName, "hostname"),
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "trusted_subnet_ips.#", resourceName, "trusted_subnet_ips.#"),
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "trusted_egress_ips.#", resourceName, "trusted_egress_ips.#"),
				),
			},
		},
	})
}

func testAccTrustedNetworkConfig(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  name       = %[1]q
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
`, name)
}

func testAccTrustedNetworkWithDataSource(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  name       	            = %[1]q
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

data "zcc_trusted_network" "this" {
  name = zcc_trusted_network.this.name
}
`, name)
}

func testAccCheckTrustedNetworkDestroy(s *terraform.State) error {
	c, err := client.NewClient(zccacctest.ClientConfigFromEnv())
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zcc_trusted_network" {
			continue
		}
		id, convErr := strconv.Atoi(rs.Primary.ID)
		if convErr != nil {
			return fmt.Errorf("invalid trusted network id %q in state: %w", rs.Primary.ID, convErr)
		}
		_, err := trusted_network_v2.Get(ctx, c.Service, id)
		if err == nil {
			return fmt.Errorf("trusted network %d still exists", id)
		}
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			continue
		}
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "not found") || strings.Contains(lower, "record not available") {
			continue
		}
		return fmt.Errorf("unexpected error verifying trusted network destroy: %w", err)
	}
	return nil
}
