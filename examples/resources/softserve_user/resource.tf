resource "softserve_user" "example" {
  username = "alice"
  admin    = false
  public_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... alice@laptop",
  ]
}
