package azurerm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/satori/uuid"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"

	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/2017-10-01-preview/sql"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmMsSqlDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmMsSqlDatabaseCreateUpdate,
		Read:   resourceArmMsSqlDatabaseRead,
		Update: resourceArmMsSqlDatabaseCreateUpdate,
		Delete: resourceArmMsSqlDatabaseDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateMsSqlDatabaseName,
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"server_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateMsSqlServerName,
			},

			"create_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(sql.Default),
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(sql.CreateModeCopy),
					string(sql.CreateModeDefault),
					string(sql.CreateModeOnlineSecondary),
					string(sql.CreateModePointInTimeRestore),
					string(sql.CreateModeRecovery),
					string(sql.CreateModeRestore),
					string(sql.CreateModeRestoreExternalBackup),
					string(sql.CreateModeRestoreExternalBackupSecondary),
					string(sql.CreateModeRestoreLongTermRetentionBackup),
					string(sql.CreateModeSecondary),
				}, true),
			},

			"elasticpool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"long_term_retention_backup_resource_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.RFC3339Time,
			},

			"max_size_gb": {
				Type:     schema.Int,
				Optional: true,
				Computed: true,
			},

			"read_scale": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.UUID,
			},

			"recoverable_database_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validate.NoEmptyStrings,
			},

			"recovery_services_recovery_point_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.RFC3339Time,
			},

			"restorable_dropped_database_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"restore_point_in_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sample_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_database_deletion_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_database_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_redundant": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sku": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"BasicPool",
								"StandardPool",
								"PremiumPool",
								"GP_Gen4",
								"GP_Gen5",
								"BC_Gen4",
								"BC_Gen5",
							}, true),
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"capacity": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"tier": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Basic",
								"Standard",
								"Premium",
								"GeneralPurpose",
								"BusinessCritical",
							}, true),
							DiffSuppressFunc: suppress.CaseDifference,
						},

						"family": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"Gen4",
								"Gen5",
							}, true),
							DiffSuppressFunc: suppress.CaseDifference,
						},
					},
				},
			},

			"tags": tagsSchema(),
		},

		CustomizeDiff: func(diff *schema.ResourceDiff, v interface{}) error {

			threatDetection, hasThreatDetection := diff.GetOk("threat_detection_policy")
			if hasThreatDetection {
				if tl := threatDetection.([]interface{}); len(tl) > 0 {
					t := tl[0].(map[string]interface{})

					state := strings.ToLower(t["state"].(string))
					_, hasStorageEndpoint := t["storage_endpoint"]
					_, hasStorageAccountAccessKey := t["storage_account_access_key"]
					if state == "enabled" && !hasStorageEndpoint && !hasStorageAccountAccessKey {
						return fmt.Errorf("`storage_endpoint` and `storage_account_access_key` are required when `state` is `Enabled`")
					}
				}
			}

			return nil
		},
	}
}

func resourceArmMsSqlDatabaseCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).msSqlDatabasesClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	serverName := d.Get("server_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	location := azureRMNormalizeLocation(d.Get("location").(string))
	createMode := d.Get("create_mode").(string)
	tags := d.Get("tags").(map[string]interface{})

	if requireResourcesToBeImported && d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup, serverName, name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing SQL Database %q (Resource Group %q, Server %q): %+v", name, resourceGroup, serverName, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_sql_database", *existing.ID)
		}
	}

	threatDetection, err := expandArmMsSqlServerThreatDetectionPolicy(d, location)
	if err != nil {
		return fmt.Errorf("Error parsing the database threat detection policy: %+v", err)
	}

	properties := sql.Database{
		Location: utils.String(location),
		DatabaseProperties: &sql.DatabaseProperties{
			CreateMode: sql.CreateMode(createMode),
		},
		Tags: expandTags(tags),
	}

	if v, ok := d.GetOk("source_database_id"); ok {
		sourceDatabaseID := v.(string)
		properties.DatabaseProperties.SourceDatabaseID = utils.String(sourceDatabaseID)
	}

	if v, ok := d.GetOk("edition"); ok {
		edition := v.(string)
		properties.DatabaseProperties.Edition = sql.DatabaseEdition(edition)
	}

	if v, ok := d.GetOk("collation"); ok {
		collation := v.(string)
		properties.DatabaseProperties.Collation = utils.String(collation)
	}

	if v, ok := d.GetOk("max_size_bytes"); ok {
		maxSizeBytes := v.(string)
		properties.DatabaseProperties.MaxSizeBytes = utils.String(maxSizeBytes)
	}

	if v, ok := d.GetOk("source_database_deletion_date"); ok {
		sourceDatabaseDeletionString := v.(string)
		sourceDatabaseDeletionDate, err2 := date.ParseTime(time.RFC3339, sourceDatabaseDeletionString)
		if err2 != nil {
			return fmt.Errorf("`source_database_deletion_date` wasn't a valid RFC3339 date %q: %+v", sourceDatabaseDeletionString, err2)
		}

		properties.DatabaseProperties.SourceDatabaseDeletionDate = &date.Time{
			Time: sourceDatabaseDeletionDate,
		}
	}

	if v, ok := d.GetOk("requested_service_objective_id"); ok {
		requestedServiceObjectiveID := v.(string)
		id, err2 := uuid.FromString(requestedServiceObjectiveID)
		if err2 != nil {
			return fmt.Errorf("`requested_service_objective_id` wasn't a valid UUID %q: %+v", requestedServiceObjectiveID, err2)
		}
		properties.DatabaseProperties.RequestedServiceObjectiveID = &id
	}

	if v, ok := d.GetOk("elastic_pool_name"); ok {
		elasticPoolName := v.(string)
		properties.DatabaseProperties.ElasticPoolName = utils.String(elasticPoolName)
	}

	if v, ok := d.GetOk("requested_service_objective_name"); ok {
		requestedServiceObjectiveName := v.(string)
		properties.DatabaseProperties.RequestedServiceObjectiveName = sql.ServiceObjectiveName(requestedServiceObjectiveName)
	}

	if v, ok := d.GetOk("restore_point_in_time"); ok {
		restorePointInTime := v.(string)
		restorePointInTimeDate, err2 := date.ParseTime(time.RFC3339, restorePointInTime)
		if err2 != nil {
			return fmt.Errorf("`restore_point_in_time` wasn't a valid RFC3339 date %q: %+v", restorePointInTime, err2)
		}

		properties.DatabaseProperties.RestorePointInTime = &date.Time{
			Time: restorePointInTimeDate,
		}
	}

	// The requested Service Objective Name does not match the requested Service Objective Id.
	if d.HasChange("requested_service_objective_name") && !d.HasChange("requested_service_objective_id") {
		properties.DatabaseProperties.RequestedServiceObjectiveID = nil
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, serverName, name, properties)
	if err != nil {
		return fmt.Errorf("Error issuing create/update request for SQL Database %q (Resource Group %q, Server %q): %+v", name, resourceGroup, serverName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting on create/update future for SQL Database %q (Resource Group %q, Server %q): %+v", name, resourceGroup, serverName, err)
	}

	if _, ok := d.GetOk("import"); ok {
		if !strings.EqualFold(createMode, "default") {
			return fmt.Errorf("import can only be used when create_mode is Default")
		}
		importParameters := expandAzureRmMsSqlDatabaseImport(d)
		importFuture, err2 := client.CreateImportOperation(ctx, resourceGroup, serverName, name, importParameters)
		if err2 != nil {
			return err2
		}

		// this is set in config.go, but something sets
		// it back to 15 minutes, which isn't long enough
		// for most imports
		client.Client.PollingDuration = 60 * time.Minute

		if err = importFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
			return err
		}
	}

	resp, err := client.Get(ctx, resourceGroup, serverName, name, "")
	if err != nil {
		return fmt.Errorf("Error issuing get request for SQL Database %q (Resource Group %q, Server %q): %+v", name, resourceGroup, serverName, err)
	}

	d.SetId(*resp.ID)

	threatDetectionClient := meta.(*ArmClient).sqlDatabaseThreatDetectionPoliciesClient
	if _, err = threatDetectionClient.CreateOrUpdate(ctx, resourceGroup, serverName, name, *threatDetection); err != nil {
		return fmt.Errorf("Error setting database threat detection policy: %+v", err)
	}

	return resourceArmMsSqlDatabaseRead(d, meta)
}

func resourceArmMsSqlDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).msSqlDatabasesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	serverName := id.Path["servers"]
	name := id.Path["databases"]

	resp, err := client.Get(ctx, resourceGroup, serverName, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading SQL Database %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Sql Database %s: %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}

	d.Set("server_name", serverName)

	if props := resp.DatabaseProperties; props != nil {
		// TODO: set `create_mode` & `source_database_id` once this issue is fixed:
		// https://github.com/Azure/azure-rest-api-specs/issues/1604

		d.Set("collation", props.Collation)
		d.Set("default_secondary_location", props.DefaultSecondaryLocation)
		d.Set("edition", string(props.Edition))
		d.Set("elastic_pool_name", props.ElasticPoolName)
		d.Set("max_size_bytes", props.MaxSizeBytes)
		d.Set("requested_service_objective_name", string(props.RequestedServiceObjectiveName))

		if cd := props.CreationDate; cd != nil {
			d.Set("creation_date", cd.String())
		}

		if rsoid := props.RequestedServiceObjectiveID; rsoid != nil {
			d.Set("requested_service_objective_id", rsoid.String())
		}

		if rpit := props.RestorePointInTime; rpit != nil {
			d.Set("restore_point_in_time", rpit.String())
		}

		if sddd := props.SourceDatabaseDeletionDate; sddd != nil {
			d.Set("source_database_deletion_date", sddd.String())
		}

		d.Set("encryption", flattenEncryptionStatus(props.TransparentDataEncryption))
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmMsSqlDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).msSqlDatabasesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	serverName := id.Path["servers"]
	name := id.Path["databases"]

	resp, err := client.Delete(ctx, resourceGroup, serverName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error making Read request on Sql Database %s: %+v", name, err)
	}

	if err != nil {
		return fmt.Errorf("Error deleting SQL Database: %+v", err)
	}

	return nil
}
