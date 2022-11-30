package tyk

import (
	"context"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

func resourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeploymentCreate,
		ReadContext:   resourceDeploymentRead,
		UpdateContext: resourceDeploymentUpdate,
		DeleteContext: resourceDeploymentDelete,
		Schema: map[string]*schema.Schema{
			"uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"team_uid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"env_uid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Home",
			},
			"zone_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"deploy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"purge": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceDeploymentDelete(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	uid := data.Id()
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	envUid := data.Get("env_uid").(string)
	de := data.Get("delete").(bool)
	purge := data.Get("purge").(bool)
	_, _, err := client.DeploymentsApi.DestroyDeployment(ctx, orgId, teamUid, envUid, uid, &cloud.DeploymentsApiDestroyDeploymentOpts{
		Delete: optional.NewBool(de),
		Purge:  optional.NewBool(purge),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return diags
}

func resourceDeploymentUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*cloud.APIClient)
	uid := data.Id()
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	envUid := data.Get("env_uid").(string)
	kind := data.Get("kind").(string)
	zone := data.Get("zone_code").(string)
	if data.HasChanges("name") {
		name := data.Get("name").(string)
		deployment := cloud.Deployment{
			BundleChannel:     "",
			BundleVersion:     "",
			Kind:              kind,
			LID:               envUid,
			LinkedDeployments: nil,
			Name:              name,
			TID:               teamUid,
			ZoneCode:          zone,
		}
		_, _, err := client.DeploymentsApi.UpdateDeployment(ctx, deployment, orgId, teamUid, envUid, uid, nil)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := data.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceDeploymentRead(ctx, data, m)
}

func resourceDeploymentRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	uid := data.Id()
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	envUid := data.Get("env_uid").(string)
	deployment, _, err := client.DeploymentsApi.GetDeployment(ctx, orgId, teamUid, envUid, uid, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("name", deployment.Payload.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("zone_code", deployment.Payload.ZoneCode); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDeploymentCreate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	name := data.Get("name").(string)
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	envUid := data.Get("env_uid").(string)
	kind := data.Get("kind").(string)
	zone := data.Get("zone_code").(string)
	deploy := data.Get("deploy").(bool)
	deployment := cloud.Deployment{
		Kind:              kind,
		LID:               envUid,
		LinkedDeployments: nil,
		Name:              name,
		TID:               teamUid,
		ZoneCode:          zone,
	}
	createDeployment, _, err := client.DeploymentsApi.CreateDeployment(ctx, deployment, orgId, teamUid, envUid)
	if err != nil {
		log.Println(err)
		return diag.FromErr(err)
	}
	data.SetId(createDeployment.Payload.UID)
	if deploy {
		_, _, err := client.DeploymentsApi.StartDeployment(ctx, orgId, teamUid, envUid, createDeployment.Payload.UID)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	resourceDeploymentRead(ctx, data, m)
	return diags
}
