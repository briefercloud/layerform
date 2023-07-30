terraform {
  required_providers {
    layerform = {
      source  = "ergomake/layerform"
      version = "~> 0.1"
    }
  }
}

provider "layerform" {}

resource "layerform_layer" "eks" {
  name   = "eks"
  path = "./eks"
}

resource "layerform_layer" "kibana" {
  name   = "kibana"
  source = "./kibana"
  dependencies = [
    layerform_layer.eks.id
  ]
}

resource "layerform_layer" "metric" {
  name   = "metric"
  path = "./metric"
  dependencies = [
    layerform_layer.kibana.id
  ]
}

resource "layerform_layer" "filebeat" {
  name   = "filebeat"
  path = "./filebeat"
  dependencies = [
    layerform_layer.kibana.id
  ]
}
