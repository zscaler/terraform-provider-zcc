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

// TestAccZIAPosture_basic exercises the full lifecycle of zcc_zia_posture:
//
//  1. Create — name, platform=macos, no trust criteria (the three
//     *_trust_criteria blocks are Optional+Computed so this is a valid
//     minimal config).
//  2. Update — rename + flip platform to windows. Both fields are
//     Required + non-RequiresReplace, so the same resource is updated
//     in place via PUT.
//  3. ImportState — round-trip the resource through `terraform import`
//     and verify the imported state matches the apply-side state.
//
// The trust-criteria blocks are intentionally NOT exercised by this
// acceptance test: their child `id` field must reference a real
// criterion in the tenant's ZIA criteria catalog, and that catalog is
// not deterministic across test runs. A second test can be layered on
// top once we have a way to fetch a valid criterion id (e.g. via a
// data source).
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

// TestAccZIAPosture_platformValidator verifies that the
// stringvalidator.OneOfCaseInsensitive guard rejects unknown platform
// names at plan time, before the SDK boundary translation would map
// them silently to 0.
//
// This step uses ExpectError instead of an apply round-trip, so it does
// not need a real tenant (PreCheck still runs to make the test honest
// about the provider being configurable).
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
				// The stringvalidator.OneOfCaseInsensitive guard wraps the
				// acceptable values; anchoring on this prefix is the most
				// stable substring across plugin-testing releases.
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
}
`, name, platform)
}

// testAccCheckZIAPostureDestroy verifies that every zcc_zia_posture in
// state was deleted upstream. The SDK Get call returns either an
// errorx.ErrorResponse with IsObjectNotFound() == true (preferred) or a
// generic error whose message contains "not found"; both are accepted.
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
