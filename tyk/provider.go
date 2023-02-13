package tyk

import (
	"context"
	"errors"
	"fmt"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
)

var (
	DashboardUrl        = "https://dashboard.cloud-ara.tyk.io"
	StagingURl          = "https://dash.ara-staging.tyk.technology"
	ErrLoginFailed      = errors.New("login failed")
	cookieAuthorisation = "cookieAuthorisation"
	signature           = "signature"
	userInfoPath        = "/api/users/whoami"
	applicationJson     = "application/json"
	contentType         = "Content-Type"
	orgInfoPath         = "api/organisations/"
	ErrorNoRoleFound    = errors.New("role not found")
	ENV                 = "DEV"
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
			"basic_pass": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TYK_BASIC_PASS", nil),
			},
			"basic_user": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TYK_BASIC_USER", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tyk_team":       resourceTeam(),
			"tyk_env":        resourceEnv(),
			"tyk_deployment": resourceDeployment(),
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
	basicUserName := d.Get("basic_user").(string)
	basicPassword := d.Get("basic_pass").(string)
	log.Println("The environment is", ENV)
	var diags diag.Diagnostics
	cookies, err := login(email, password, basicUserName, basicPassword, ENV == "DEV")
	if err != nil {
		return nil, diag.FromErr(err)
	}
	client := resty.New()
	staging := false
	if ENV == "DEV" {
		staging = true
		client.SetBaseURL(StagingURl)
		client.SetBasicAuth(basicUserName, basicPassword)
	} else {
		client.SetBaseURL(DashboardUrl)
		client.SetAuthToken(createTokenFromCookies(cookies))
	}
	client.SetCookies(cookies)
	conf := cloud.Configuration{
		DefaultHeader: map[string]string{},
	}
	controllerUrl, err := createUserController(ctx, client, staging)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	t := fmt.Sprintf("Bearer %s", createTokenFromCookies(cookies))
	c := cloud.NewAPIClient(&conf)
	conf.AddDefaultHeader("Authorization", t)
	c.ChangeBasePath(controllerUrl)
	return c, diags
}

func login(email, password, basicUserName, basicPassword string, staging bool) ([]*http.Cookie, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	client := resty.New()
	if staging {
		staging = true
		client.SetBaseURL(StagingURl).SetBasicAuth(basicUserName, basicPassword)
	} else {
		client.SetBaseURL(DashboardUrl)
	}
	req := client.R()
	resp, err :=
		req.SetHeader("Accept", "application/json").
			SetBody(map[string]string{"email": email, "password": password}).
			Post("api/login")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK && resp.Body() != nil {
		return nil, NewGenericHttpError(resp.String())
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, ErrLoginFailed
	}
	return resp.Cookies(), nil
}

func createTokenFromCookies(cookies []*http.Cookie) string {
	var cookieAuth, cookieSignature string
	for _, cookie := range cookies {
		switch cookie.Name {
		case cookieAuthorisation:
			cookieAuth = cookie.Value
		case signature:
			cookieSignature = cookie.Value
		}

	}
	return fmt.Sprintf("%s.%s", cookieAuth, cookieSignature)
}
func createUserController(ctx context.Context, client *resty.Client, staging bool) (string, error) {
	info, err := getUserInfo(ctx, client)
	if err != nil {
		return "", err
	}
	role, err := getUserRole(info.Roles)
	if err != nil {
		return "", err
	}
	orgInfo, err := GetOrgInfo(ctx, client, role.OrgID)
	if err != nil {
		return "", err
	}
	return GenerateUrlFromZone(orgInfo.Organisation.Zone, staging)

}
