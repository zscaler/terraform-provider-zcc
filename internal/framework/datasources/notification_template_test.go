package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccDataSourceNotificationTemplate_byName(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-nt-ds-%s", acctest.RandString(8))
	resourceName := "zcc_notification_template.this"
	dataSourceName := "data.zcc_notification_template.by_name"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNotificationTemplateConfigByName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_client", resourceName, "enable_client"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_zia", resourceName, "enable_zia"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_app_updates", resourceName, "enable_app_updates"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_service_status", resourceName, "enable_service_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "duration_in_seconds", resourceName, "duration_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_persistent", resourceName, "enable_persistent"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_do_not_disturb", resourceName, "enable_do_not_disturb"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"zia_notification_template.enable_zia_firewall",
						resourceName, "zia_notification_template.enable_zia_firewall"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"zia_notification_template.enable_zia_persistent",
						resourceName, "zia_notification_template.enable_zia_persistent"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"zpa_notification_template.zpa_reauth_interval_in_minutes",
						resourceName, "zpa_notification_template.zpa_reauth_interval_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"zpa_notification_template.delay_posture_failure_seconds",
						resourceName, "zpa_notification_template.delay_posture_failure_seconds"),
				),
			},
		},
	})
}

// TestAccDataSourceNotificationTemplate_byID is the same exercise as
// _byName but goes through the numeric id lookup branch — important
// because the data source supports both shapes and they hit different
// SDK code paths (Get vs GetByName).
func TestAccDataSourceNotificationTemplate_byID(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-nt-ds-id-%s", acctest.RandString(8))
	resourceName := "zcc_notification_template.this"
	dataSourceName := "data.zcc_notification_template.by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNotificationTemplateConfigByID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "duration_in_seconds", resourceName, "duration_in_seconds"),
				),
			},
		},
	})
}

func testAccDataSourceNotificationTemplateConfigByName(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_notification_template" "this" {
  name                  = %[1]q
  enable_client         = true
  enable_zia            = true
  enable_app_updates    = true
  enable_service_status = true
  duration_in_seconds   = 5
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
    zpa_reauth_interval_in_minutes = 5
    delay_posture_failure_seconds  = 0
  }
}

data "zcc_notification_template" "by_name" {
  name = zcc_notification_template.this.name
}
`, name)
}

func testAccDataSourceNotificationTemplateConfigByID(name string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_notification_template" "this" {
  name                  = %[1]q
  enable_client         = true
  enable_zia            = true
  enable_app_updates    = true
  enable_service_status = true
  duration_in_seconds   = 5
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
    zpa_reauth_interval_in_minutes = 5
    delay_posture_failure_seconds  = 0
  }
}

data "zcc_notification_template" "by_id" {
  id = zcc_notification_template.this.id
}
`, name)
}
