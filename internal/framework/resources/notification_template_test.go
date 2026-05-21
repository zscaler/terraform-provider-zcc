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
	"github.com/zscaler/zscaler-sdk-go/v3/zscaler/zcc/services/notification_template"

	"github.com/zscaler/terraform-provider-zcc/internal/client"
	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

// TestAccNotificationTemplate_basic exercises the full lifecycle of
// zcc_notification_template:
//
//  1. Create — full config mirrored from local_dev/notification_template/main.tf
//     (all top-level toggles + both nested zia/zpa blocks).
//  2. Update — rename and flip a representative pair of bool toggles +
//     bump the ZPA reauth interval, all without RequiresReplace, so the
//     same resource is updated in place via PUT.
//  3. ImportState — round-trip via `terraform import` and verify the
//     imported state matches the apply-side state.
//
// CheckDestroy uses the SDK Get; both an errorx.ErrorResponse with
// IsObjectNotFound() == true and a generic "not found" error message
// are accepted, mirroring the resource's own Read handling.
func TestAccNotificationTemplate_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-nt-%s", acctest.RandString(8))
	rNameUpdated := rName + "-upd"
	resourceName := "zcc_notification_template.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNotificationTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationTemplateConfig(rName, 5, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_client", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_zia", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_app_updates", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_service_status", "true"),
					resource.TestCheckResourceAttr(resourceName, "duration_in_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "enable_persistent", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_do_not_disturb", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_firewall", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_firewall_popup", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_dns", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_dns_popup", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_ips", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_ips_popup", "true"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_persistent", "true"),
					resource.TestCheckResourceAttr(resourceName, "zpa_notification_template.enable_device_posture_failure", "true"),
					resource.TestCheckResourceAttr(resourceName, "zpa_notification_template.enable_zpa_reauth", "true"),
					resource.TestCheckResourceAttr(resourceName, "zpa_notification_template.zpa_reauth_interval_in_minutes", "5"),
					resource.TestCheckResourceAttr(resourceName, "zpa_notification_template.delay_posture_failure_seconds", "0"),
				),
			},
			{
				Config: testAccNotificationTemplateConfigUpdated(rNameUpdated, 10, 15),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "duration_in_seconds", "10"),
					// Flipped toggles
					resource.TestCheckResourceAttr(resourceName, "enable_app_updates", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_do_not_disturb", "false"),
					resource.TestCheckResourceAttr(resourceName, "zia_notification_template.enable_zia_ips_popup", "false"),
					// Bumped interval
					resource.TestCheckResourceAttr(resourceName, "zpa_notification_template.zpa_reauth_interval_in_minutes", "15"),
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

// testAccNotificationTemplateConfig mirrors local_dev/notification_template/main.tf
// — every top-level toggle on, both nested blocks fully populated.
// Parameters are kept explicit so the update step can vary the
// transient-display duration and the ZPA reauth interval without
// touching the rest of the body.
func testAccNotificationTemplateConfig(name string, durationSeconds, reauthMinutes int) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_notification_template" "this" {
  name                  = %[1]q
  enable_client         = true
  enable_zia            = true
  enable_app_updates    = true
  enable_service_status = true
  duration_in_seconds   = %[2]d
  enable_persistent     = true
  enable_do_not_disturb = true

  zia_notification_template = {
    enable_zia_firewall       = true
    enable_zia_firewall_popup = true
    enable_zia_dns            = true
    enable_zia_dns_popup      = true
    enable_zia_ips            = true
    enable_zia_ips_popup      = true
    enable_zia_persistent     = true
  }

  zpa_notification_template = {
    enable_device_posture_failure  = true
    enable_zpa_reauth              = true
    zpa_reauth_interval_in_minutes = %[3]d
    delay_posture_failure_seconds  = 0
  }
}
`, name, durationSeconds, reauthMinutes)
}

// testAccNotificationTemplateConfigUpdated flips a representative subset
// of booleans (enable_app_updates, enable_do_not_disturb, and the
// nested enable_zia_ips_popup) so the Update step in the test verifies
// both top-level and nested PUTs round-trip cleanly through the API.
func testAccNotificationTemplateConfigUpdated(name string, durationSeconds, reauthMinutes int) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_notification_template" "this" {
  name                  = %[1]q
  enable_client         = true
  enable_zia            = true
  enable_app_updates    = false
  enable_service_status = true
  duration_in_seconds   = %[2]d
  enable_persistent     = true
  enable_do_not_disturb = false

  zia_notification_template = {
    enable_zia_firewall       = true
    enable_zia_firewall_popup = true
    enable_zia_dns            = true
    enable_zia_dns_popup      = true
    enable_zia_ips            = true
    enable_zia_ips_popup      = false
    enable_zia_persistent     = true
  }

  zpa_notification_template = {
    enable_device_posture_failure  = true
    enable_zpa_reauth              = true
    zpa_reauth_interval_in_minutes = %[3]d
    delay_posture_failure_seconds  = 0
  }
}
`, name, durationSeconds, reauthMinutes)
}

// testAccCheckNotificationTemplateDestroy verifies that every
// zcc_notification_template in state has been deleted upstream. The SDK
// Get call returns either an errorx.ErrorResponse with
// IsObjectNotFound() == true (preferred) or a generic error whose
// message contains "not found"; both are accepted, mirroring the
// resource's own Read handling.
func testAccCheckNotificationTemplateDestroy(s *terraform.State) error {
	c, err := client.NewClient(zccacctest.ClientConfigFromEnv())
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zcc_notification_template" {
			continue
		}
		id, convErr := strconv.Atoi(rs.Primary.ID)
		if convErr != nil {
			return fmt.Errorf("invalid notification template id %q in state: %w", rs.Primary.ID, convErr)
		}
		_, err := notification_template.Get(ctx, c.Service, id)
		if err == nil {
			return fmt.Errorf("notification template %d still exists", id)
		}
		var respErr *errorx.ErrorResponse
		if errors.As(err, &respErr) && respErr.IsObjectNotFound() {
			continue
		}
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			continue
		}
		return fmt.Errorf("unexpected error verifying notification template destroy: %w", err)
	}
	return nil
}
