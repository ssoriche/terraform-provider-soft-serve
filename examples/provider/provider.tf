terraform {
  required_providers {
    softserve = {
      source = "registry.terraform.io/ssoriche/soft-serve"
    }
  }
}

provider "softserve" {
  host     = "localhost"
  port     = 23231
  username = "admin"
  # private_key_path = "~/.ssh/id_ed25519"
  # identity_file    = "~/.ssh/id_ed25519.pub"
  use_agent = true
}
