package tyk

import (
	"context"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceEnv() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvCreate,
		ReadContext:   resourceEnvRead,
		UpdateContext: resourceEnvUpdate,
		DeleteContext: resourceEnvDelete,
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
		},
	}
}

func resourceEnvDelete(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	uid := data.Id()
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	_, _, err := client.LoadoutsApi.DeleteLoadout(ctx, orgId, teamUid, uid, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return diags
}

func resourceEnvUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*cloud.APIClient)
	uid := data.Id()
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	if data.HasChange("name") {
		name := data.Get("name").(string)
		load := cloud.Loadout{Name: name}
		_, _, err := client.LoadoutsApi.UpdateLoadout(ctx, load, orgId, teamUid, uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := data.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceEnvRead(ctx, data, m)
}

func resourceEnvRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	uid := data.Id()
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)
	loadout, _, err := client.LoadoutsApi.GetLoadout(ctx, orgId, teamUid, uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("name", loadout.Payload.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("uid", loadout.Payload.UID); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("org_id", loadout.Payload.OID); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("team_uid", loadout.Payload.TID); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceEnvCreate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	name := data.Get("name").(string)
	orgId := data.Get("org_id").(string)
	teamUid := data.Get("team_uid").(string)

	env := cloud.Loadout{
		Name: name,
		OID:  orgId,
		TID:  teamUid,
	}
	loadout, _, err := client.LoadoutsApi.CreateLoadout(ctx, env, orgId, teamUid)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(loadout.Payload.UID)
	resourceEnvRead(ctx, data, m)
	return diags
}
