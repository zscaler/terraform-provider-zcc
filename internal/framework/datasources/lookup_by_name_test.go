package datasources_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

func TestAccCustomIPApps_byName(t *testing.T) {
	name := os.Getenv("TF_ACC_ZCC_CUSTOM_IP_APP_NAME")
	if name == "" {
		t.Skip("set TF_ACC_ZCC_CUSTOM_IP_APP_NAME to an existing custom IP app name in the test tenant")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamedDataSource("zcc_custom_ip_apps", "name", name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.zcc_custom_ip_apps.this", "app_name"),
					resource.TestCheckResourceAttrSet("data.zcc_custom_ip_apps.this", "id"),
				),
			},
		},
	})
}

func TestAccPredefinedIPApps_byName(t *testing.T) {
	name := os.Getenv("TF_ACC_ZCC_PREDEFINED_IP_APP_NAME")
	if name == "" {
		t.Skip("set TF_ACC_ZCC_PREDEFINED_IP_APP_NAME to an existing predefined IP app name in the test tenant")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamedDataSource("zcc_predefined_ip_apps", "name", name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.zcc_predefined_ip_apps.this", "app_name"),
				),
			},
		},
	})
}

func TestAccProcessBasedApps_byName(t *testing.T) {
	name := os.Getenv("TF_ACC_ZCC_PROCESS_BASED_APP_NAME")
	if name == "" {
		t.Skip("set TF_ACC_ZCC_PROCESS_BASED_APP_NAME to an existing process-based app name in the test tenant")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamedDataSource("zcc_process_based_apps", "name", name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.zcc_process_based_apps.this", "app_name"),
				),
			},
		},
	})
}

func TestAccApplicationProfiles_byName(t *testing.T) {
	name := os.Getenv("TF_ACC_ZCC_APPLICATION_PROFILE_NAME")
	if name == "" {
		t.Skip("set TF_ACC_ZCC_APPLICATION_PROFILE_NAME to an existing application profile name in the test tenant")
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamedDataSource("zcc_application_profiles", "name", name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.zcc_application_profiles.this", "name", name),
				),
			},
		},
	})
}

func testAccNamedDataSource(dataType, attr, value string) string {
	return fmt.Sprintf(`
provider "zcc" {}

data "%[1]s" "this" {
  %[2]s = %[3]q
}
`, dataType, attr, value)
}
