locals {
  dir = pathexpand("~/.layerform/examples/local/foo-${random_string.foo_suffix.result}")
}

resource "local_file" "foo" {
  content  = ""
  filename = "${local.dir}/.keep"
}

resource "random_string" "foo_suffix" {
  length  = 4
  upper   = false
  special = false
}

output "foo_file" {
  value = local_file.foo.filename
}
