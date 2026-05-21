package resources_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttr(resourceName, "network_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "trusted_subnet_ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trusted_subnet_ips.0", "192.0.2.0/24"),
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
					resource.TestCheckResourceAttrPair("data.zcc_trusted_network.this", "network_name", resourceName, "network_name"),
				),
			},
		},
	})
}

func testAccTrustedNetworkConfig(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  network_name       = %[1]q
  active             = true
  condition_type     = "ALL"
  trusted_subnet_ips = ["192.0.2.0/24"]
}
`, name)
}

func testAccTrustedNetworkWithDataSource(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  network_name       = %[1]q
  active             = true
  condition_type     = "ALL"
  trusted_subnet_ips = ["192.0.2.0/24"]
}

data "zcc_trusted_network" "this" {
  network_name = zcc_trusted_network.this.network_name
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
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}
		return fmt.Errorf("unexpected error verifying trusted network destroy: %w", err)
	}
	return nil
}
