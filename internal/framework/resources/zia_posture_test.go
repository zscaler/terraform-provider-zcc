package resources_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/errorx"
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/zia_posture"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccZIAPosture_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-posture-%s", acctest.RandString(8))
	rNameUpdated := rName + "-upd"
	resourceName := "zcc_zia_posture.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckZIAPostureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZIAPostureConfig(rName, "macos"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", "macos"),
				),
			},
			{
				Config: testAccZIAPostureConfig(rNameUpdated, "windows"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "platform", "windows"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccZIAPosture_platformValidator(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccZIAPostureConfig(
					fmt.Sprintf("tf-acc-test-posture-%s", acctest.RandString(8)),
					"bsd",
				),

				ExpectError: regexp.MustCompile(`Attribute platform value must be one of`),
			},
		},
	})
}

func testAccZIAPostureConfig(name, platform string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_zia_posture" "this" {
  name     = %[1]q
  platform = %[2]q
  high_trust_criteria = {
    cs = [
      {
        cn = [
          { id = "9911", name = "CrowdStrike_ZPA_ZTA_40" },
          { id = "criterion-id-2", name = "Firewall Enabled" },
        ]
      }
    ]
  }

  medium_trust_criteria = {
    cs = [
      {
        cn = [
          { id = "9913", name = "CrowdStrike_ZPA_ZTA_80" },
        ]
      }
    ]
  }

  low_trust_criteria = {
    cs = [
      { id = "9913", name = "CrowdStrike_ZPA_ZTA_80" },
    ]
  }
}
`, name, platform)
}

func testAccCheckZIAPostureDestroy(s *terraform.State) error {
	c, err := client.NewClient(zccacctest.ClientConfigFromEnv())
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zcc_zia_posture" {
			continue
		}
		id, convErr := strconv.Atoi(rs.Primary.ID)
		if convErr != nil {
			return fmt.Errorf("invalid zia posture id %q in state: %w", rs.Primary.ID, convErr)
		}
		_, err := zia_posture.Get(ctx, c.Service, id)
		if err == nil {
			return fmt.Errorf("zia posture %d still exists", id)
		}
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			continue
		}
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}
		return fmt.Errorf("unexpected error verifying zia posture destroy: %w", err)
	}
	return nil
}
