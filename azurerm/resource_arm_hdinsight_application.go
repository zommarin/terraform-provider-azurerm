package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/preview/hdinsight/mgmt/2018-06-01-preview/hdinsight"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmHDInsightApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmHDInsightApplicationCreate,
		Read:   resourceArmHDInsightApplicationRead,
		Delete: resourceArmHDInsightApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO: validation
			},

			"cluster_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"marketplace_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vm_size": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO: validation for the SKU
			},

			"install_script_action": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"roles": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"edgenode",
									"headnode",
									"workernode",
									"zookeepernode",
								}, false),
							},
							Set: schema.HashString,
						},
					},
				},
			},

			"uninstall_script_action": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"roles": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"edgenode",
									"headnode",
									"workernode",
									"zookeepernode",
								}, false),
							},
							Set: schema.HashString,
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},

			"https_endpoint": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_port": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"public_port": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"access_modes": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceArmHDInsightApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).hdinsightApplicationsClient
	clustersClient := meta.(*ArmClient).hdinsightClustersClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for HDInsight Application creation.")

	name := d.Get("name").(string)
	clusterIdStr := d.Get("cluster_id").(string)
	clusterId, err := parseAzureResourceID(clusterIdStr)
	if err != nil {
		return err
	}

	clusterName := clusterId.Path["clusters"]
	resourceGroup := clusterId.ResourceGroup

	if cluster, err := clustersClient.Get(ctx, resourceGroup, clusterName); err != nil {
		if utils.ResponseWasNotFound(cluster.Response) {
			return fmt.Errorf("Error: HDInsight Cluster %q was not found in Resource Group %q!", clusterName, resourceGroup)
		}

		return fmt.Errorf("Error retrieving HDInsight Cluster %q (Resource Group %q): %+v", clusterName, resourceGroup, err)
	}

	if requireResourcesToBeImported {
		existing, err := client.Get(ctx, resourceGroup, clusterName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing HDInsight Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_hdinsight_application", *existing.ID)
		}
	}

	marketplaceIdentifier := d.Get("marketplace_identifier").(string)
	vmSize := d.Get("vm_size").(string)

	httpsEndpointsRaw := d.Get("https_endpoint").([]interface{})
	httpsEndpoints := expandHDInsightApplicationHttpsEndpoints(httpsEndpointsRaw)
	installScriptActionsRaw := d.Get("install_script_action").([]interface{})
	installScriptActions := expandHDInsightApplicationScriptActions(installScriptActionsRaw)
	uninstallScriptActionsRaw := d.Get("uninstall_script_action").([]interface{})
	uninstallScriptActions := expandHDInsightApplicationScriptActions(uninstallScriptActionsRaw)

	application := hdinsight.Application{
		Properties: &hdinsight.ApplicationProperties{
			ApplicationType:       utils.String("CustomApplication"),
			MarketplaceIdentifier: utils.String(marketplaceIdentifier),
			ComputeProfile: &hdinsight.ComputeProfile{
				Roles: &[]hdinsight.Role{
					{
						// these have to be hard-coded
						Name:                utils.String("edgenode"),
						TargetInstanceCount: utils.Int32(int32(1)),
						HardwareProfile: &hdinsight.HardwareProfile{
							VMSize: utils.String(vmSize),
						},
					},
				},
			},
			HTTPSEndpoints:         httpsEndpoints,
			InstallScriptActions:   installScriptActions,
			UninstallScriptActions: uninstallScriptActions,
		},
	}

	future, err := client.Create(ctx, resourceGroup, clusterName, name, application)
	if err != nil {
		return fmt.Errorf("Error creating HDInsight Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for creation of HDInsight Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, clusterName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving HDInsights Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("[ERROR] Cannot read ID for HDInsight Application %q (Cluster %q / Resource Group %q)", name, clusterName, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmHDInsightApplicationRead(d, meta)
}

func resourceArmHDInsightApplicationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).hdinsightApplicationsClient
	clustersClient := meta.(*ArmClient).hdinsightClustersClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	clusterName := id.Path["clusters"]
	name := id.Path["applications"]

	resp, err := client.Get(ctx, resourceGroup, clusterName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] HDInsight Application %q (Cluster %q / Resource Group %q) was not found - removing from state!", name, clusterName, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving HDInsight Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
	}

	cluster, err := clustersClient.Get(ctx, resourceGroup, clusterName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] HDInsight Cluster %q (Resource Group %q) was not found - removing from state!", clusterName, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving HDInsight Cluster %q (Resource Group %q): %+v", clusterName, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("cluster_id", cluster.ID)

	if props := resp.Properties; props != nil {
		// NOTE: whilst the vm_size is returned via the props.ComputeProfile.HardwareProfile - it can be transformed, as such we ignore it
		d.Set("marketplace_identifier", props.MarketplaceIdentifier)

		httpsEndpoints := flattenHDInsightApplicationHttpsEndpoints(props.HTTPSEndpoints)
		if err := d.Set("https_endpoint", httpsEndpoints); err != nil {
			return fmt.Errorf("Error setting `https_endpoints`: %+v", err)
		}

		installActions := flattenHDInsightApplicationScriptActions(props.InstallScriptActions)
		if err := d.Set("install_script_action", installActions); err != nil {
			return fmt.Errorf("Error setting `install_script_action`: %+v", err)
		}

		uninstallActions := flattenHDInsightApplicationScriptActions(props.UninstallScriptActions)
		if err := d.Set("uninstall_script_action", uninstallActions); err != nil {
			return fmt.Errorf("Error setting `uninstall_script_action`: %+v", err)
		}
	}

	return nil
}

func resourceArmHDInsightApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).hdinsightApplicationsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	clusterName := id.Path["clusters"]
	name := id.Path["applications"]

	future, err := client.Delete(ctx, resourceGroup, clusterName, name)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error deleting HDInsight Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
		}
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("Error waiting for deletion of HDInsight Application %q (Cluster %q / Resource Group %q): %+v", name, clusterName, resourceGroup, err)
		}
	}

	return nil
}

func expandHDInsightApplicationScriptActions(input []interface{}) *[]hdinsight.RuntimeScriptAction {
	actions := make([]hdinsight.RuntimeScriptAction, 0)

	for _, v := range input {
		val := v.(map[string]interface{})

		name := val["name"].(string)
		uri := val["uri"].(string)

		rolesRaw := val["roles"].(*schema.Set).List()
		roles := make([]string, 0)
		for _, v := range rolesRaw {
			role := v.(string)
			roles = append(roles, role)
		}

		action := hdinsight.RuntimeScriptAction{
			Name:  utils.String(name),
			URI:   utils.String(uri),
			Roles: &roles,
		}

		actions = append(actions, action)
	}

	return &actions
}

func flattenHDInsightApplicationScriptActions(input *[]hdinsight.RuntimeScriptAction) []interface{} {
	outputs := make([]interface{}, 0)
	if input == nil {
		return outputs
	}

	for _, action := range *input {
		output := make(map[string]interface{}, 0)

		if name := action.Name; name != nil {
			output["name"] = *name
		}

		if uri := action.URI; uri != nil {
			output["uri"] = *uri
		}

		roles := make([]interface{}, 0)
		if action.Roles != nil {
			for _, r := range *action.Roles {
				roles = append(roles, r)
			}
		}
		output["roles"] = schema.NewSet(schema.HashString, roles)
		outputs = append(outputs, output)
	}

	return outputs
}

func expandHDInsightApplicationHttpsEndpoints(input []interface{}) *[]hdinsight.ApplicationGetHTTPSEndpoint {
	results := make([]hdinsight.ApplicationGetHTTPSEndpoint, 0)

	for _, raw := range input {
		v := raw.(map[string]interface{})

		accessModesRaw := v["access_modes"].(*schema.Set).List()
		accessModes := make([]string, 0)
		for _, v := range accessModesRaw {
			accessModes = append(accessModes, v.(string))
		}

		destinationPort := v["destination_port"].(int)
		result := hdinsight.ApplicationGetHTTPSEndpoint{
			DestinationPort: utils.Int32(int32(destinationPort)),
			AccessModes:     &accessModes,
		}

		publicPort := v["public_port"].(int)
		if publicPort > 0 {
			result.PublicPort = utils.Int32(int32(publicPort))
		}

		results = append(results, result)
	}

	return &results
}

func flattenHDInsightApplicationHttpsEndpoints(input *[]hdinsight.ApplicationGetHTTPSEndpoint) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	outputs := make([]interface{}, 0)

	for _, v := range *input {
		output := map[string]interface{}{
			"access_modes":     []interface{}{},
			"destination_port": 0,
			"public_port":      0,
		}

		if v.DestinationPort != nil {
			output["destination_port"] = int(*v.DestinationPort)
		}

		if v.PublicPort != nil {
			output["public_port"] = int(*v.PublicPort)
		}

		accessModes := make([]interface{}, 0)
		if v.AccessModes != nil {
			for _, v := range *v.AccessModes {
				accessModes = append(accessModes, v)
			}
		}
		output["access_modes"] = accessModes

		outputs = append(outputs, output)
	}

	return outputs
}
