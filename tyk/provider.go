package tyk

import (
	"context"
	"errors"
	"fmt"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

var (
	DashboardUrl         = "https://dashboard.cloud-ara.tyk.io"
	ErrLoginFailed       = errors.New("login failed")
	cookieAuthorisation  = "cookieAuthorisation"
	signature            = "signature"
	ErrTokenNotFound     = errors.New("no token found")
	ErrSignatureNotFound = errors.New("signature not found")
	baseUrl              = "https://controller-aws-euw2.cloud-ara.tyk.io:37001"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TYK_EMAIL", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TYK_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tyk_team": resourceTeam(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"tyk_teams": dataSourceTeams(),
			"tyk_org":   dataSourceOrg(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	email := d.Get("email").(string)
	password := d.Get("password").(string)
	var diags diag.Diagnostics
	if email == "" {
		return nil, diag.FromErr(errors.New("email is required"))
	}
	if password == "" {
		return nil, diag.FromErr(errors.New("password is required"))
	}
	client := resty.New().SetBaseURL(DashboardUrl)
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetBody(map[string]string{"email": email, "password": password}).
		Post("api/login")
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, diag.FromErr(ErrLoginFailed)
	}
	var token string
	var cookieSignature string
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case cookieAuthorisation:
			token = cookie.Value

		case signature:

			cookieSignature = cookie.Value
		}

	}
	if len(token) == 0 {
		return nil, diag.FromErr(ErrTokenNotFound)

	}
	if cookieSignature == "" {
		return nil, diag.FromErr(ErrSignatureNotFound)

	}
	tn := fmt.Sprintf("%s.%s", token, cookieSignature)
	conf := cloud.Configuration{
		DefaultHeader: map[string]string{},
		BasePath:      baseUrl,
	}
	t := fmt.Sprintf("Bearer %s", tn)
	c := cloud.NewAPIClient(&conf)
	conf.AddDefaultHeader("Authorization", t)
	c.ChangeBasePath(baseUrl)
	return c, diags
}
