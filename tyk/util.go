package tyk

import (
	"context"
	"errors"
	"fmt"
	"github.com/TykTechnologies/cloud-sdk/cloud"
	"github.com/go-resty/resty/v2"
	"golang.org/x/exp/slices"
	"strings"
	"time"
)

// /getUserRole returns the user role.
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

func getUserInfo(ctx context.Context, client *resty.Client) (*UserInfo, error) {
	var userInfo UserInfo
	request := client.R().SetHeader(contentType, applicationJson)
	request.SetContext(ctx).SetResult(&userInfo)
	resp, err := request.Get(userInfoPath)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, NewGenericHttpError(resp.String())
	}
	return &userInfo, nil
}

func GetOrgInfo(ctx context.Context, client *resty.Client, orgId string) (*OrgInfo, error) {
	var orgInfo OrgInfo
	request := client.R().SetHeader(contentType, applicationJson)
	request.SetContext(ctx).SetResult(&orgInfo)
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

func GenerateUrlFromZone(region string, useStaging bool) (string, error) {
	regionPart := strings.Split(region, "-")
	if len(regionPart) != 4 {
		return "", errors.New("the format of this region is wrong")
	}
	suffix := "cloud-ara.tyk.io:37001"
	if useStaging {
		suffix = "ara-staging.tyk.technology:37001"
	}
	url := fmt.Sprintf("https://controller-aws-%s%s%s.%s", regionPart[1], AbbreviateDirection(regionPart[2]), regionPart[3], suffix)
	return url, nil
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

func AbbreviateDirection(direction string) string {
	switch direction {
	case "east":
		return "e"
	case "north":
		return "n"
	case "south":
		return "s"
	case "northeast":
		return "ne"
	case "northwest":
		return "nw"
	case "west":
		return "w"
	case "southwest":
		return "sw"
	case "southeast":
		return "se"
	case "central":
		return "c"
	}

	return ""
}
