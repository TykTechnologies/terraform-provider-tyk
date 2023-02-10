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
data "tyk_org" "org" {

}
resource "tyk_team" "team" {
  name = "Terraform team"
  oid  = data.tyk_org.org.uid
}


resource "tyk_env" "env" {
  name     = "Terraform env"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.org.uid
}

resource "tyk_deployment" "home" {
  name ="Terraform home deployment"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.org.uid
  zone_code=data.tyk_org.org.zone
  env_uid=tyk_env.env.uid
  deploy= true
  delete=true
  purge=true
}

resource "tyk_deployment" "edge" {
  name ="Terraform edge gateway"
  team_uid = tyk_team.team.uid
  org_id   = data.tyk_org.org.uid
  zone_code=data.tyk_org.org.zone
  env_uid=tyk_env.env.uid
  kind="Gateway"
  linked_control_plane=tyk_deployment.home.uid
  deploy= true
  delete=true
  purge=true
}





