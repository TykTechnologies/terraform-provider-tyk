terraform {
  required_providers {
    tyk = {
      version = "0.2"
      source  = "tyk.io/tyk/tyk"
    }
  }
}

provider "tyk" {

}
data "tyk_org" "first" {

}
resource "tyk_team" "team" {
  name = "Testing tretararform team "
  oid  = data.tyk_org.first.uid
}


resource "tyk_env" "env" {
  name     = "change this  Terraform i changed name"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.first.uid
}

resource "tyk_deployment" "home" {
  name ="test home deployment deployment"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.first.uid
  zone_code=data.tyk_org.first.zone
  env_uid=tyk_env.env.uid
  deploy= true
  delete=true
  purge=true
}
resource "tyk_deployment" "gateway" {
  name ="terraform gateway deployment"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.first.uid
  zone_code=data.tyk_org.first.zone
  env_uid=tyk_env.env.uid
  kind="Gateway"
  deploy= true
  delete=true
  purge=true
  linked_control_plane=tyk_deployment.home.uid
}




output "deployment" {
  value = tyk_deployment.home
}
output "orgs" {
  value = data.tyk_org.first
}
output "team" {
  value = tyk_team.team
}
