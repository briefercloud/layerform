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
  region  = local.region
}

data "layerform_layer" "kibana" {
  name = "kibana"
}

locals {
  kibana_folder = data.layerform_layer.kibana.output.kibana_folder
}

provider "aws" {
  region  = data.layerform_layer.kibana.output.aws_region
}

resource "aws_s3_object" "filebeat" {
  bucket = data.layerform_layer.kibana.output.bucket.bucket
  key    = ${kibana_folder}/filebeat-${random_string.suffix}/.keep"
  content = "filebeat"
}

resource "random_string" "suffix" {
  length  = 8
  upper   = false
  special = false
}
