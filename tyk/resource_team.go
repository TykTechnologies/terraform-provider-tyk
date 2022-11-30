package tyk

import (
	"context"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Schema: map[string]*schema.Schema{
			"uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"oid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceTeamDelete(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	uid := data.Id()
	oid := data.Get("oid").(string)
	_, _, err := client.TeamsApi.DeleteTeam(ctx, oid, uid, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return diags
}

func resourceTeamUpdate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {

	client := m.(*cloud.APIClient)
	uid := data.Id()
	if data.HasChange("name") {
		oid := data.Get("oid").(string)
		name := data.Get("name").(string)
		team := cloud.Team{
			Name: name,
			OID:  oid,
		}
		_, _, err := client.TeamsApi.UpdateTeam(ctx, team, oid, uid)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := data.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceTeamRead(ctx, data, m)
}

func resourceTeamRead(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	uid := data.Id()
	oid := data.Get("oid").(string)
	team, _, err := client.TeamsApi.GetTeam(ctx, oid, uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("name", team.Payload.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("oid", team.Payload.OID); err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set("uid", team.Payload.UID); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceTeamCreate(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	oid := data.Get("oid").(string)
	name := data.Get("name").(string)
	team := cloud.Team{
		Name: name,
		OID:  oid,
	}
	teamPayload, _, err := client.TeamsApi.CreateTeam(ctx, team, oid)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(teamPayload.Payload.UID)
	resourceTeamRead(ctx, data, m)
	return diags
}
