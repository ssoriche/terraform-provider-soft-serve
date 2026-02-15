resource "softserve_repository_collaborator" "example" {
  repository   = softserve_repository.example.name
  username     = softserve_user.example.username
  access_level = "read-write"
}
