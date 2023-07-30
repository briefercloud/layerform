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

resource "aws_s3_bucket" "eks_bucket" {
  bucket = "eks-bucket-${random_string.bucket_suffix.result}"
}

resource "random_string" "bucket_suffix" {
  length  = 8
  upper   = false
  special = false
}

