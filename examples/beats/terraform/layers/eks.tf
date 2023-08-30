module "eks" {
  source = "./eks"
}

output "bucket_name" {
  value       = module.eks.bucket.id
  description = "AWS S3 Bucket"
}
