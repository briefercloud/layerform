variable "prefix" {
  type    = string
  default = ""
}

locals {
  dir = pathexpand("~/.layerform/examples/local/${var.prefix}foo-${var.lf_names.foo}")
}

resource "local_file" "foo" {
  content  = ""
  filename = "${local.dir}/.keep"
}

output "foo_file" {
  value = local_file.foo.filename
}
