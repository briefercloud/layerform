resource "local_file" "baz" {
  content  = "baz content"
  filename = "${local.dir}/baz-${random_string.baz_suffix.result}.txt"
}

resource "random_string" "baz_suffix" {
  length  = 4
  upper   = false
  special = false
}
