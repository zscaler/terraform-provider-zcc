package resources_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/trusted_network"

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
  network_name    = %[1]q
  active          = true
  trusted_subnets = "192.0.2.0/24"
}
`, name)
}

func testAccTrustedNetworkWithDataSource(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_trusted_network" "this" {
  network_name    = %[1]q
  active          = true
  trusted_subnets = "192.0.2.0/24"
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
		_, _, err := trusted_network.GetTrustedNetworkByID(ctx, c.Service, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("trusted network %s still exists", rs.Primary.ID)
		}
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}
		return fmt.Errorf("unexpected error verifying trusted network destroy: %w", err)
	}
	return nil
}
