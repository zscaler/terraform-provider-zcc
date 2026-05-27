package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	zccacctest "github.com/zscaler/terraform-provider-zcc/internal/framework/acctest"
)

// TestAccDataSourceZIAPosture_byID seeds a `zcc_zia_posture` resource and
// then reads it back through the data source by its numeric id, which is
// currently the only supported lookup key (the `/zia-posture-profiles`
// list endpoint mishandles pagination, so name / platform branches are
// disabled in the data source until the upstream API is fixed — see the
// schema description in datasources/zia_posture.go).
//
// Pair-checking against the managing resource is preferred over hard-
// coded value assertions because it makes the test resilient to
// arbitrary tenant state changes and to API field-ordering quirks.
func TestAccDataSourceZIAPosture_byID(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-posture-ds-%s", acctest.RandString(8))
	resourceName := "zcc_zia_posture.this"
	dataSourceName := "data.zcc_zia_posture.by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { zccacctest.PreCheck(t) },
		ProtoV6ProviderFactories: zccacctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceZIAPostureByIDConfig(rName, "macos"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					// Pair-check the nested trust-criteria blocks at the
					// shape level: `cs.#` (number of criteria sets) and
					// `cs.0.cn.#` (number of criteria in the first set)
					// must match between resource and data source.
					resource.TestCheckResourceAttrPair(dataSourceName,
						"high_trust_criteria.cs.#",
						resourceName, "high_trust_criteria.cs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"high_trust_criteria.cs.0.cn.#",
						resourceName, "high_trust_criteria.cs.0.cn.#"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"medium_trust_criteria.cs.#",
						resourceName, "medium_trust_criteria.cs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName,
						"low_trust_criteria.cs.#",
						resourceName, "low_trust_criteria.cs.#"),
				),
			},
		},
	})
}

func testAccDataSourceZIAPostureByIDConfig(name, platform string) string {
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

data "zcc_zia_posture" "by_id" {
  id = zcc_zia_posture.this.id
}
`, name, platform)
}
