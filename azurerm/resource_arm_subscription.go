package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/preview/subscription/mgmt/2018-03-01-preview/subscription"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSubscriptionCreate,
		Read:   resourceArmSubscriptionRead,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 24),
			},

			"enrollment_account": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"owners": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"offer_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"additional_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceArmSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {

	eaClient := meta.(*ArmClient).subscriptionsEAClient
	client := meta.(*ArmClient).subscriptionsClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for Azure ARM Subscription creation.")

	// name := d.Get("name").(string)
	enrollmentAccount := d.Get("enrollment_account").(string)

	/*
		if requireResourcesToBeImported && d.IsNewResource() {
			existing, err := client.Get(ctx, resGroup, name)
			if err != nil {
				if !utils.ResponseWasNotFound(existing.Response) {
					return fmt.Errorf("Error checking for presence of existing User Assigned Identity %q (Resource Group %q): %+v", name, resGroup, err)
				}
			}

			if existing.ID != nil && *existing.ID != "" {
				return tf.ImportAsExistsError("azurerm_user_assigned_identity", *existing.ID)
			}
		}*/

	creationParameters := subscription.CreationParameters{
		OfferType: subscription.OfferType(d.Get("offer_type").(string)),
		// DisplayName: &name,
	}

	future, err := eaClient.CreateSubscriptionInEnrollmentAccount(ctx, enrollmentAccount, creationParameters)
	if err != nil {
		return fmt.Errorf("Error Creating Subscription %q: %+v", enrollmentAccount, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for the Subscription %q to finish creating: %+v", enrollmentAccount, err)
	}

	read, err := client.Get(ctx, enrollmentAccount)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Subscription %q", enrollmentAccount)
	}

	d.SetId(*read.SubscriptionID)

	return resourceArmSubscriptionRead(d, meta)
}

func resourceArmSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient)
	groupClient := client.subscriptionsClient
	ctx := client.StopContext

	resp, err := groupClient.Get(ctx, d.Id())
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading subscription: %+v", err)
	}

	d.Set("subscription_id", resp.SubscriptionID)
	d.Set("display_name", resp.DisplayName)
	d.Set("state", resp.State)
	if resp.SubscriptionPolicies != nil {
		d.Set("location_placement_id", resp.SubscriptionPolicies.LocationPlacementID)
		d.Set("quota_id", resp.SubscriptionPolicies.QuotaID)
		d.Set("spending_limit", resp.SubscriptionPolicies.SpendingLimit)
	}

	return nil
}
