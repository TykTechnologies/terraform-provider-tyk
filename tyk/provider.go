package tyk

import (
	"context"
	"errors"
	"fmt"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"time"
)

var (
	DashboardUrl         = "https://dashboard.cloud-ara.tyk.io"
	StagingURl           = "https://dash.ara-staging.tyk.technology"
	ErrLoginFailed       = errors.New("login failed")
	cookieAuthorisation  = "cookieAuthorisation"
	signature            = "signature"
	ErrTokenNotFound     = errors.New("no token found")
	ErrSignatureNotFound = errors.New("signature not found")
	Env                  = "DEV"
	userInfoPath         = "/api/users/whoami"
	applicationJson      = "application/json"
	contentType          = "Content-Type"
	orgInfoPath          = "api/organisations/"
	ErrorNoRoleFound     = errors.New("role not found")
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"basic_user": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TYK_BASIC_USER", nil),
			},
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

	var diags diag.Diagnostics
	if email == "" {
		return nil, diag.FromErr(errors.New("email is required"))
	}
	if password == "" {
		return nil, diag.FromErr(errors.New("password is required"))
	}
	client := resty.New()
	staging := false
	if Env == "DEV" {
		staging = true
		client.SetBaseURL(StagingURl)
	} else {
		client.SetBaseURL(DashboardUrl)
	}
	req := client.R()
	if staging {
		req = req.SetBasicAuth(basicUserName, basicPassword)
	}
	resp, err :=
		req.SetHeader("Accept", "application/json").
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
	var tk string
	var sig string
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case cookieAuthorisation:
			token = cookie.Value
			tk = cookie.Value
		case signature:

			cookieSignature = cookie.Value
			sig = cookie.Value
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
	}
	///client.SetAuthToken(tn)
	if staging {
		client.SetBasicAuth(basicUserName, basicPassword)
		client.SetCookie(&http.Cookie{
			Name:  "cookieAuthorisation",
			Value: tk,
		})
		client.SetCookie(&http.Cookie{
			Name:  "signature",
			Value: sig,
		})
	} else {
		client.SetAuthToken(tn)
	}
	info, err := getUserInfo(ctx, client)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	role, err := getUserRole(info.Roles)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	log.Println("the organization id is", role.OrgID)
	orgInfo, err := GetOrgInfo(ctx, client, role.OrgID)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	controllerUrl, err := GenerateUrlFromZone(orgInfo.Organisation.Zone, staging)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	t := fmt.Sprintf("Bearer %s", tn)
	c := cloud.NewAPIClient(&conf)
	conf.AddDefaultHeader("Authorization", t)
	c.ChangeBasePath(controllerUrl)
	return c, diags
}

func GetOrgInfo(ctx context.Context, client *resty.Client, orgId string) (*OrgInfo, error) {
	var orgInfo OrgInfo
	request := client.R().SetHeader(contentType, applicationJson)
	request.SetContext(ctx).SetResult(&orgInfo)
	log.Println("below organization id is", orgId)
	path := fmt.Sprintf("%s%s", orgInfoPath, orgId)
	response, err := request.Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, NewGenericHttpError(response.String())
	}

	return &orgInfo, nil
}
func getUserInfo(ctx context.Context, client *resty.Client) (*UserInfo, error) {
	var userInfo UserInfo
	request := client.R().SetHeader(contentType, applicationJson)
	request.SetContext(ctx).SetResult(&userInfo)
	resp, err := request.Get(userInfoPath)
	if err != nil {
		log.Println("i failed here check me", err)
		return nil, err
	}
	if resp.StatusCode() != 200 {
		log.Println("am failing here check it", resp.Request.Token)
		return nil, NewGenericHttpError(resp.String())
	}
	return &userInfo, nil
}

type UserInfo struct {
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
	PasswordUpdated time.Time `json:"password_updated"`
	Email           string    `json:"email"`
	LastName        string    `json:"lastName"`
	AccountID       string    `json:"account_id"`
	FirstName       string    `json:"firstName"`
	ID              string    `json:"id"`
	Roles           []Role    `json:"roles"`
	HubspotID       int       `json:"hubspot_id"`
	IsActive        bool      `json:"is_active"`
	IsEmailVerified bool      `json:"is_email_verified"`
}

type Role struct {
	Role      string `json:"role"`
	OrgID     string `json:"org_id"`
	TeamID    string `json:"team_id"`
	OrgName   string `json:"org_name"`
	TeamName  string `json:"team_name"`
	AccountID string `json:"account_id"`
}

type OrgInfo struct {
	Organisation cloud.Organisation `json:"Organisation"`
}

// / getUserRole returns the user role.
func getUserRole(roles []Role) (*Role, error) {
	roleList := []string{"org_admin", "team_admin", "team_member"}
	for _, role := range roles {
		contain := slices.Contains(roleList, role.Role)
		if contain {
			return &role, nil
		}
	}
	return nil, ErrorNoRoleFound
}
