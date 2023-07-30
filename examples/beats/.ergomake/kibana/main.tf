terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
    layerform = {
      source  = "ergomake/layerform"
      version = "~> 0.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}

provider "layerform" {}

provider "aws" {
  region = local.region
}

provider "random" {}

locals {
  region        = data.layerform_layer.eks.output.aws_region
  kibana_folder = "kibana-${random_string.suffix.id}"
}

data "layerform_layer" "eks" {
  name = "eks"
}

resource "random_string" "suffix" {
  length  = 8
  upper   = false
  special = false
}

resource "aws_s3_object" "kibana" {
  bucket  = data.layerform_layer.eks.output.bucket.bucket
  key     = "${local.s3_path}/.keep"
  content = "kibana"
}

output "aws_region" {
  description = "AWS Region"
  value       = local.region
}

output "bucket" {
  description = "AWS Region"
  value       = aws_s3_object.kibana.bucket
}

output "kibana_folder" {
  description = "AWS Region"
  value       = local.kibana_folder
}
