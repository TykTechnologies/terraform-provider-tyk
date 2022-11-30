package tyk

import (
	"context"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceOrg() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrgRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceOrgRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*cloud.APIClient)
	orgs, _, err := client.OrganisationsApi.GetOrgs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(orgs.Payload.Organisations) > 0 {
		firstOrg := orgs.Payload.Organisations[0]
		if err := d.Set("name", firstOrg.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("zone", firstOrg.Zone); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("uid", firstOrg.UID); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(orgs.Payload.Organisations[0].UID)
	} else {
		d.SetId(strconv.FormatInt(time.Now().Unix(), 1))
	}
	return diags
}
