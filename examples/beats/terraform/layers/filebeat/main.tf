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

resource "aws_s3_object" "filebeat" {
  bucket  = var.bucket_name
  key     = "${var.kibana_folder}/filebeat-${random_string.suffix.result}/.keep"
  content = "filebeat"
}

resource "random_string" "suffix" {
  length  = 8
  upper   = false
  special = false
}
