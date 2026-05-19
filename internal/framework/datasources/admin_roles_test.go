package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccAdminRoles_basic(t *testing.T) {
	roleName := os.Getenv("TF_ACC_ZCC_ADMIN_ROLE_NAME")
	if roleName == "" {
		roleName = "Super Admin"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAdminRolesDataSourceConfig(roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.zcc_admin_roles.this", "role_name", roleName),
					resource.TestCheckResourceAttrSet("data.zcc_admin_roles.this", "id"),
				),
			},
		},
	})
}

func testAccAdminRolesDataSourceConfig(roleName string) string {
	return fmt.Sprintf(`
provider "zcc" {}

data "zcc_admin_roles" "this" {
  role_name = %[1]q
}
`, roleName)
}
