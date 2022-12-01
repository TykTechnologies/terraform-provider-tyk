output "deployment" {
  value = tyk_deployment.home
}
output "orgs" {
  value = data.tyk_org.first
}
output "team" {
  value = tyk_team.team
}