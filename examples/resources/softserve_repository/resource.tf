resource "softserve_repository" "example" {
  name        = "my-project"
  description = "An example repository"
  private     = true
  hidden      = false
}
