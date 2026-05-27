package datasources_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

// TestAccDataSourceWebAppService_byName exercises the `zcc_web_app_service`
// data source as an open-query lookup (the resource side of this entity
// is *not* registered in the provider — bypass apps are managed in the
// tenant UI and consumed read-only from Terraform).
//
// The data source matches by `name` against `/getWebAppService`. Two
// well-known stock bypass apps that ship with every ZCC tenant are
// pinned in the test config — `ZOOMMEETING` and `Microsoft Teams` —
// so the test does not need any env var or pre-existing tenant-specific
// state to run.
func TestAccDataSourceWebAppService_byName(t *testing.T) {
	zoomDS := "data.zcc_web_app_service.zoom"
	teamsDS := "data.zcc_web_app_service.teams"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWebAppServiceByNameConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// ZOOMMEETING — stock bypass app, name is all
					// upper-case and single-token.
					resource.TestCheckResourceAttrSet(zoomDS, "id"),
					resource.TestCheckResourceAttr(zoomDS, "name", "ZOOMMEETING"),
					resource.TestCheckResourceAttr(zoomDS, "app_name", "ZOOMMEETING"),
					resource.TestCheckResourceAttrSet(zoomDS, "active"),

					// "Microsoft Teams" — stock bypass app, name
					// contains a space and mixed case. Pinning both
					// shapes catches regressions in name-encoding /
					// query-param escaping on the GetByName SDK path.
					resource.TestCheckResourceAttrSet(teamsDS, "id"),
					resource.TestCheckResourceAttr(teamsDS, "name", "Microsoft Teams"),
					resource.TestCheckResourceAttr(teamsDS, "app_name", "Microsoft Teams"),
					resource.TestCheckResourceAttrSet(teamsDS, "active"),
				),
			},
		},
	})
}

func testAccDataSourceWebAppServiceByNameConfig() string {
	return `
provider "zcc" {}

data "zcc_web_app_service" "zoom" {
  name = "ZOOMMEETING"
}

data "zcc_web_app_service" "teams" {
  name = "Microsoft Teams"
}
`
}
