package resources_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/forwarding_profile"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccForwardingProfile_basic(t *testing.T) {
	name := fmt.Sprintf("tf-acc-test-fp-%s", acctest.RandString(8))
	resourceName := "zcc_forwarding_profile.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckForwardingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccForwardingProfileConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccForwardingProfileConfigUpdated(name + "-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name+"-updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccForwardingProfileWithDataSource(name + "-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.zcc_forwarding_profile.this", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.zcc_forwarding_profile.this", "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccForwardingProfileConfig(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_forwarding_profile" "this" {
  name = %[1]q
}
`, name)
}

func testAccForwardingProfileConfigUpdated(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_forwarding_profile" "this" {
  name = %[1]q
}
`, name)
}

func testAccForwardingProfileWithDataSource(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_forwarding_profile" "this" {
  name = %[1]q
}

data "zcc_forwarding_profile" "this" {
  name = zcc_forwarding_profile.this.name
}
`, name)
}

func testAccCheckForwardingProfileDestroy(s *terraform.State) error {
	c, err := client.NewClient(zccacctest.ClientConfigFromEnv())
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zcc_forwarding_profile" {
			continue
		}
		idInt, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid forwarding profile id: %w", err)
		}
		profiles, err := forwarding_profile.GetForwardingProfileByCompanyID(ctx, c.Service, "", nil, nil)
		if err != nil {
			return err
		}
		for i := range profiles {
			if int(profiles[i].ID) == idInt {
				return fmt.Errorf("forwarding profile %s still exists", rs.Primary.ID)
			}
		}
	}
	return nil
}
