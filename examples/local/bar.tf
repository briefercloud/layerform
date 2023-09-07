resource "local_file" "bar" {
  content  = "bar content"
  filename = "${local.dir}/bar-${random_string.bar_suffix.result}.txt"
}

resource "random_string" "bar_suffix" {
  length  = 4
  upper   = false
  special = false
}
