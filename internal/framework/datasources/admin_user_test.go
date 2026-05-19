package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccAdminUser_basic(t *testing.T) {
	userName := os.Getenv("TF_ACC_ZCC_ADMIN_USER_NAME")
	if userName == "" {
		t.Skip("set TF_ACC_ZCC_ADMIN_USER_NAME to an admin login that exists in the test tenant")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAdminUserDataSourceConfig(userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.zcc_admin_user.this", "user_name", userName),
				),
			},
		},
	})
}

func testAccAdminUserDataSourceConfig(userName string) string {
	return fmt.Sprintf(`
provider "zcc" {}

data "zcc_admin_user" "this" {
  user_name = %[1]q
}
`, userName)
}
