output "aws_region" {
  value       = local.aws_region
  description = "AWS Region"
}

output "bucket" {
  value       = aws_s3_bucket.eks_bucket
  description = "AWS Region"
}
