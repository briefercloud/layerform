resource "local_file" "bar" {
  content  = "bar content"
  filename = "${local.dir}/bar-${var.lf_names.bar}.txt"
}

output "bar_file" {
  value = local_file.bar.filename
}
