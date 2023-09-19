resource "local_file" "baz" {
  content  = "baz content"
  filename = "${local.dir}/baz-${var.lf_names.baz}.txt"
}

output "baz_file" {
  value = local_file.baz.filename
}
