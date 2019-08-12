terraform {
  backend "local" {}
}

variable "region" {}

resource "null_resource" "bar" {}
