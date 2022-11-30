package tyk

import (
	"context"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamsRead,
		Schema: map[string]*schema.Schema{
			"oid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"teams": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"oid": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTeamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*cloud.APIClient)
	oid := d.Get("oid").(string)
	teams, _, err := client.TeamsApi.GetTeams(ctx, oid)
	if err != nil {
		return diag.FromErr(err)
	}
	fetchedTeams := flattenTeamData(teams.Payload.Teams)
	if err := d.Set("teams", fetchedTeams); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenTeamData(teams []cloud.Team) []interface{} {
	if teams != nil {
		ois := make([]interface{}, len(teams), len(teams))
		for i, team := range teams {
			oi := make(map[string]interface{})
			oi["uid"] = team.UID
			oi["name"] = team.Name
			oi["oid"] = team.OID

			ois[i] = oi

		}
		return ois
	}
	return make([]interface{}, 0)
}
