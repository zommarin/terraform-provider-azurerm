package azurerm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
)

func TestAccAzureRMHDInsightApplication_basic(t *testing.T) {
	resourceName := "azurerm_hdinsight_application.test"
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(11))
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMHDInsightApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHDInsightApplication_basic(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHDInsightApplicationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// whilst 'returned' from the API it can be malformed, since it's a ForceNew anyway we don't set it
					"vm_size",
				},
			},
		},
	})
}

func TestAccAzureRMHDInsightApplication_requiresImport(t *testing.T) {
	if !requireResourcesToBeImported {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	resourceName := "azurerm_hdinsight_application.test"
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(11))
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMHDInsightApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHDInsightApplication_basic(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHDInsightApplicationExists(resourceName),
				),
			},
			{
				Config:      testAccAzureRMHDInsightApplication_requiresImport(ri, rs, location),
				ExpectError: testRequiresImportError("azurerm_hdinsight_application"),
			},
		},
	})
}

func TestAccAzureRMHDInsightApplication_ports(t *testing.T) {
	resourceName := "azurerm_hdinsight_application.test"
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(11))
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMHDInsightApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHDInsightApplication_ports(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHDInsightApplicationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// whilst 'returned' from the API it can be malformed, since it's a ForceNew anyway we don't set it
					"vm_size",
				},
			},
		},
	})
}

func TestAccAzureRMHDInsightApplication_complete(t *testing.T) {
	resourceName := "azurerm_hdinsight_application.test"
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(11))
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMHDInsightApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHDInsightApplication_complete(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHDInsightApplicationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// whilst 'returned' from the API it can be malformed, since it's a ForceNew anyway we don't set it
					"vm_size",
				},
			},
		},
	})
}

func TestAccAzureRMHDInsightApplication_uninstallAction(t *testing.T) {
	resourceName := "azurerm_hdinsight_application.test"
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(11))
	location := testLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMHDInsightApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHDInsightApplication_uninstallAction(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHDInsightApplicationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// whilst 'returned' from the API it can be malformed, since it's a ForceNew anyway we don't set it
					"vm_size",
				},
			},
		},
	})
}

func testCheckAzureRMHDInsightApplicationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_hdinsight_application" {
			continue
		}

		client := testAccProvider.Meta().(*ArmClient).hdinsightApplicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		applicationName := rs.Primary.Attributes["name"]
		clusterId := rs.Primary.Attributes["cluster_id"]

		id, err := parseAzureResourceID(clusterId)
		if err != nil {
			return err
		}

		clusterName := id.Path["clusters"]
		resourceGroup := id.ResourceGroup
		resp, err := client.Get(ctx, resourceGroup, clusterName, applicationName)

		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return err
			}
		}
	}

	return nil
}

func testCheckAzureRMHDInsightApplicationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		applicationName := rs.Primary.Attributes["name"]
		clusterId := rs.Primary.Attributes["cluster_id"]

		id, err := parseAzureResourceID(clusterId)
		if err != nil {
			return err
		}

		clusterName := id.Path["clusters"]
		resourceGroup := id.ResourceGroup

		client := testAccProvider.Meta().(*ArmClient).hdinsightApplicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, resourceGroup, clusterName, applicationName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: HDInsight Application %q (Cluster %q / Resource Group: %q) does not exist", applicationName, clusterName, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on hdinsightClustersClient: %+v", err)
		}

		return nil
	}
}

func testAccAzureRMHDInsightApplication_basic(rInt int, rString string, location string) string {
	template := testAccAzureRMHDInsightApplication_template(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_hdinsight_application" "test" {
  name                   = "acctest-%d"
  cluster_id             = "${azurerm_hdinsight_hadoop_cluster.test.id}"
  marketplace_identifier = "CustomApplication"
  vm_size                = "Standard_D4_V2"

  install_script_action {
    name  = "say-hello"
    uri   = "https://gist.githubusercontent.com/tombuildsstuff/74ff75620a83cf2a737843920185dbc2/raw/8217fbbcf9728e23807c19a35f65136351e6da7a/hello.sh"
    roles = [ "edgenode" ]
  }
}
`, template, rInt)
}

func testAccAzureRMHDInsightApplication_requiresImport(rInt int, rString string, location string) string {
	template := testAccAzureRMHDInsightApplication_basic(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_hdinsight_application" "import" {
  name                   = "${azurerm_hdinsight_application.test.name}"
  cluster_id             = "${azurerm_hdinsight_application.test.cluster_id}"
  marketplace_identifier = "${azurerm_hdinsight_application.test.marketplace_identifier}"
  vm_size                = "${azurerm_hdinsight_application.test.vm_size}"
  install_script_action  = "${azurerm_hdinsight_application.test.install_script_action}"
}
`, template)
}

func testAccAzureRMHDInsightApplication_uninstallAction(rInt int, rString string, location string) string {
	template := testAccAzureRMHDInsightApplication_template(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_hdinsight_application" "test" {
  name                   = "acctest-%d"
  cluster_id             = "${azurerm_hdinsight_hadoop_cluster.test.id}"
  marketplace_identifier = "CustomApplication"
  vm_size                = "Standard_D4_V2"

  install_script_action {
    name  = "say-hello"
    uri   = "https://gist.githubusercontent.com/tombuildsstuff/74ff75620a83cf2a737843920185dbc2/raw/8217fbbcf9728e23807c19a35f65136351e6da7a/hello.sh"
    roles = [ "edgenode" ]
  }

  uninstall_script_action {
    name  = "say-goodbye"
    uri   = "TODO" # TODO: link me
    roles = [ "edgenode" ]
  }
}
`, template, rInt)
}

func testAccAzureRMHDInsightApplication_ports(rInt int, rString string, location string) string {
	template := testAccAzureRMHDInsightApplication_template(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_hdinsight_application" "test" {
  name                   = "acctest-%d"
  cluster_id             = "${azurerm_hdinsight_hadoop_cluster.test.id}"
  marketplace_identifier = "CustomApplication"
  vm_size                = "Standard_D4_V2"

  install_script_action {
    name  = "say-hello"
    uri   = "https://gist.githubusercontent.com/tombuildsstuff/74ff75620a83cf2a737843920185dbc2/raw/8217fbbcf9728e23807c19a35f65136351e6da7a/hello.sh"
    roles = [ "edgenode" ]
  }

  https_endpoint {
    destination_port = 8080
  }
}
`, template, rInt)
}

func testAccAzureRMHDInsightApplication_complete(rInt int, rString string, location string) string {
	template := testAccAzureRMHDInsightApplication_template(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_hdinsight_application" "test" {
  name                   = "acctest-%d"
  cluster_id             = "${azurerm_hdinsight_hadoop_cluster.test.id}"
  marketplace_identifier = "CustomApplication"
  vm_size                = "Standard_D4_V2"

  install_script_action {
    name  = "say-hello"
    uri   = "https://gist.githubusercontent.com/tombuildsstuff/74ff75620a83cf2a737843920185dbc2/raw/8217fbbcf9728e23807c19a35f65136351e6da7a/hello.sh"
    roles = [ "edgenode" ]
  }

  install_script_action {
    name  = "say-hola"
    uri   = "https://gist.githubusercontent.com/tombuildsstuff/74ff75620a83cf2a737843920185dbc2/raw/8217fbbcf9728e23807c19a35f65136351e6da7a/hello.sh"
    roles = [ "edgenode" ]
  }

  uninstall_script_action {
    name  = "say-goodbye"
    uri   = "TODO"
    roles = [ "edgenode" ]
  }

  uninstall_script_action {
    name  = "say-adios"
    uri   = "TODO"
    roles = [ "edgenode" ]
  }

  https_endpoint {
    destination_port = 8080
    public_port      = 32617
  }

  https_endpoint {
    destination_port = 8088
    public_port      = 32618
    access_modes = [ "webpage" ]
  }
}
`, template, rInt)
}

func testAccAzureRMHDInsightApplication_template(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestsa%s"
  resource_group_name      = "${azurerm_resource_group.test.name}"
  location                 = "${azurerm_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "test" {
  name                  = "acctest"
  resource_group_name   = "${azurerm_resource_group.test.name}"
  storage_account_name  = "${azurerm_storage_account.test.name}"
  container_access_type = "private"
}

resource "azurerm_hdinsight_hadoop_cluster" "test" {
  name                = "acctesthdi-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  cluster_version     = "3.6"
  tier                = "Standard"

  component_version {
    hadoop = "2.7"
  }

  gateway {
    enabled  = true
    username = "acctestusrgw"
    password = "TerrAform123!"
  }

  storage_account {
    storage_container_id = "${azurerm_storage_container.test.id}"
    storage_account_key  = "${azurerm_storage_account.test.primary_access_key}"
    is_default           = true
  }

  roles {
    head_node {
      vm_size  = "Standard_D3_v2"
      username = "acctestusrvm"
      password = "AccTestvdSC4daf986!"
    }

    worker_node {
      vm_size               = "Standard_D4_V2"
      username              = "acctestusrvm"
      password              = "AccTestvdSC4daf986!"
      target_instance_count = 2
    }

    zookeeper_node {
      vm_size  = "Standard_D3_v2"
      username = "acctestusrvm"
      password = "AccTestvdSC4daf986!"
    }
  }
}
`, rInt, location, rString, rInt)
}
