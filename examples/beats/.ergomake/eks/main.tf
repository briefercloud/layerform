terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}

locals {
  aws_region = "us-west-1"
}

provider "aws" {
  region = local.aws_region
}

provider "random" {}

resource "aws_s3_bucket" "eks_bucket" {
  bucket = "eks-bucket-${random_string.bucket_suffix.result}"
}

resource "random_string" "bucket_suffix" {
  length  = 8
  upper   = false
  special = false
}

output "aws_region" {
  value       = local.aws_region
  description = "AWS Region"
}

output "bucket" {
  value       = aws_s3_bucket.eks_bucket
  description = "AWS Region"
}
