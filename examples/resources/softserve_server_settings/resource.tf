resource "softserve_server_settings" "this" {
  allow_keyless = false
  anon_access   = "read-only"
}
