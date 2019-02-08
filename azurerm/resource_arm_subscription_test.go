package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
)

func TestAccAzureRMSubscription_basic(t *testing.T) {
	var sS storage.Share

	ri := tf.AccRandTimeInt()
	config := testAccAzureRMSubscription_basic(ri)
	resourceName := "azurerm_storage_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(resourceName, &sS),
				),
			},
		},
	})
}

func testCheckAzureRMSubscriptionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["enrollment_account"]

		client := testAccProvider.Meta().(*ArmClient).subscriptionsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on subscriptonsClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Subscription %q does not exist", name)
		}

		return nil
	}
}

func testAccAzureRMSubscription_basic(rInt int) string {
	return fmt.Sprintf(`
resource "azurerm_subscription" "test" {
  enrollment_account                 = "testsubscription-%d"
  offer_type = "MS-AZR-0148P"
}
`, rInt)
}
