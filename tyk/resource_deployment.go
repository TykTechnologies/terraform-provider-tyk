package tyk

import (
	"context"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/antihax/optional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

const (
	StateInitialise           = "initialise"
	StatePreDeploy            = "pre-deploy"
	StateDeploying            = "deploying"
	StatePostDeploy           = "post-deploy"
	StateDeployed             = "deployed"
	StateDeployingDescendants = "deploying-descendants"
	StateUpdatingParent       = "updating-parent"
	StateFailing              = "failing"
	StateFailed               = "failed"
	StateDestroying           = "destroying"
	StateDestroyed            = "destroyed"
	StateUpdating             = "updating"
	// tykDbEdgeEndpoints is env variable name.
	tykDbEdgeEndpoints      = "TYK_DB_EDGEENDPOINTS"
	DeploymentStatusTimeout = 10 * time.Minute
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
			"linked_control_plane": {
				Type:     schema.TypeString,
				Optional: true,
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
	//de := data.Get("delete").(bool)
	purge := data.Get("purge").(bool)
	deployStateConf := &resource.StateChangeConf{
		Delay:   3 * time.Second,
		Pending: []string{StatePreDeploy, StateDeploying, StateFailing, StateUpdating},
		Refresh: checkDeploymentStatusChange(ctx, client, orgId, teamUid, envUid, uid),
		Target:  []string{StateDeployed, StateFailed, StateDestroyed},
		Timeout: DeploymentStatusTimeout,
	}
	_, err := deployStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	_, _, err = client.DeploymentsApi.DestroyDeployment(ctx, orgId, teamUid, envUid, uid, &cloud.DeploymentsApiDestroyDeploymentOpts{
		///Delete: optional.NewBool(false),
		Purge: optional.NewBool(purge),
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
	///kind := data.Get("kind").(string)
	///zone := data.Get("zone_code").(string)
	if data.HasChanges("name") {
		name := data.Get("name").(string)
		deployment := cloud.Deployment{
			Name: name,
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
	deployment, resp, err := client.DeploymentsApi.GetDeployment(ctx, orgId, teamUid, envUid, uid, nil)
	if resp.StatusCode == 404 {
		data.SetId("")
		return diags
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("name", deployment.Payload.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("zone_code", deployment.Payload.ZoneCode); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("uid", deployment.Payload.UID); err != nil {
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
	linkedControlPlane := data.Get("linked_control_plane").(string)
	deployment := cloud.Deployment{
		Kind:              kind,
		LID:               envUid,
		LinkedDeployments: map[string]string{},
		LastUpdate:        time.Now().UTC(),
		Created:           time.Now().UTC(),
		Driver:            "K8s_sp",
		DriverMetaData: &cloud.Status{
			CurrentState: "starting",
			Timestamp:    time.Now().UTC(),
		},
		ExtraContext: &cloud.MetaDataStore{
			Data: map[string]map[string]interface{}{},
		},
		Name:     name,
		TID:      teamUid,
		ZoneCode: zone,
	}
	if linkedControlPlane != "" && deployment.Kind == "Gateway" {
		deployment.LinkedDeployments["LinkedMDCBID"] = linkedControlPlane
	}
	createDeployment, _, err := client.DeploymentsApi.CreateDeployment(ctx, deployment, orgId, teamUid, envUid)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(createDeployment.Payload.UID)
	if deploy {
		_, _, err := client.DeploymentsApi.StartDeployment(ctx, orgId, teamUid, envUid, createDeployment.Payload.UID)
		if err != nil {
			return diag.FromErr(err)
		}
		deployStateConf := &resource.StateChangeConf{
			Delay:   10 * time.Second,
			Pending: []string{StateInitialise, StatePreDeploy, StateDeploying},
			Refresh: checkDeploymentStatusChange(ctx, client, orgId, teamUid, envUid, createDeployment.Payload.UID),
			Target:  []string{StateDeployed},
			Timeout: DeploymentStatusTimeout,
		}
		_, err = deployStateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceDeploymentRead(ctx, data, m)
	return diags
}

func checkDeploymentStatusChange(ctx context.Context, client *cloud.APIClient, orgId, teamId, envId, deploymentId string) resource.StateRefreshFunc {
	return func() (result interface{}, state string, err error) {
		deployment, _, err := client.DeploymentsApi.GetDeployment(ctx, orgId, teamId, envId, deploymentId, nil)
		if err != nil {
			return nil, StateFailed, err
		}

		return deployment, deployment.Payload.State, nil
	}
}
