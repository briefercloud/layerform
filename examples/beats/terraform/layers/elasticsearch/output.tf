output "aws_region" {
  description = "AWS Region"
  value       = local.region
}

output "bucket_name" {
  description = "AWS Region"
  value       = aws_s3_object.elasticsearch.bucket
}

output "elasticsearch_folder" {
  description = "AWS Region"
  value       = local.elasticsearch_folder
}
