package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

// TestAccAdminUser_basic exercises the zcc_admin_user data source against
// the live tenant. The upstream `/getAdminUsers` API requires a `userType`
// query parameter (1=ZIA, 2=ZPA, 3=ZID, 4=ZDX); the data source accepts the
// case-insensitive aliases `ZIA`/`ZPA`/`ZID`/`ZDX` and translates them to
// the numeric form before calling the SDK. We hard-code `user_type = "ZIA"`
// because that is the scope under which the default super-admin login lives
// in the integration tenant.
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
					resource.TestCheckResourceAttr("data.zcc_admin_user.this", "user_type", "ZIA"),
					resource.TestCheckResourceAttrSet("data.zcc_admin_user.this", "id"),
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
  user_type = "ZIA"
}
`, userName)
}
