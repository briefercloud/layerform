module "kibana" {
  source = "./kibana"

  aws_region  = module.eks.aws_region
  bucket_name = module.eks.bucket.bucket
}
