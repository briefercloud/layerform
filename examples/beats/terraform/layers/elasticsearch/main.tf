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

provider "aws" {
  region = var.aws_region
}

locals {
  region               = var.aws_region
  elasticsearch_folder = "elasticsearch-${random_string.suffix.id}"
}

resource "random_string" "suffix" {
  length  = 8
  upper   = false
  special = false
}

resource "aws_s3_object" "elasticsearch" {
  bucket  = var.bucket_name
  key     = "${local.elasticsearch_folder}/.keep"
  content = "elasticsearch"
}

