output "aws_region" {
  description = "AWS Region"
  value       = local.region
}

output "bucket_name" {
  description = "AWS Region"
  value       = aws_s3_object.kibana.bucket
}

output "kibana_folder" {
  description = "AWS Region"
  value       = local.kibana_folder
}
