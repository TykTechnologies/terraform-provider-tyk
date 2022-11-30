terraform {
  required_providers {
    tyk = {
      version = "0.2"
      source  = "tyk.io/tyk/tyk"
    }
  }
}

provider "tyk" {}

data "tyk_org" "first" {

}

resource "tyk_team" "team" {
  name = "terraform testing changed me"
  oid  = data.tyk_org.first.uid
}


resource "tyk_env" "env" {
  name     = "Terraform i changed name"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.first.uid
}

