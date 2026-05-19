package resources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccWebAppService_basic(t *testing.T) {
	appName := os.Getenv("TF_ACC_ZCC_WEB_APP_NAME")
	if appName == "" {
		t.Skip("set TF_ACC_ZCC_WEB_APP_NAME to an existing ZCC web app service (bypass app) name in the test tenant")
	}

	resourceName := "zcc_web_app_service.this"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppServiceConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "app_name", appName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppServiceWithDataSource(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.zcc_web_app_service.this", "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair("data.zcc_web_app_service.this", "app_name", resourceName, "app_name"),
				),
			},
		},
	})
}

func testAccWebAppServiceConfig(appName string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_web_app_service" "this" {
  app_name = %[1]q
}
`, appName)
}

func testAccWebAppServiceWithDataSource(appName string) string {
	return fmt.Sprintf(`
provider "zcc" {}

resource "zcc_web_app_service" "this" {
  app_name = %[1]q
}

data "zcc_web_app_service" "this" {
  name = zcc_web_app_service.this.app_name
}
`, appName)
}
