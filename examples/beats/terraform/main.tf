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
  name = "eks"
  files = [
    "layers/eks.tf",
    "layers/eks/main.tf",
    "layers/eks/output.tf",
  ]
}

resource "layerform_layer" "kibana" {
  name   = "kibana"
  files = [
    "layers/kibana.tf",
    "layers/kibana/main.tf",
    "layers/kibana/output.tf",
    "layers/kibana/variables.tf",
  ]
  dependencies = [
    layerform_layer.eks.id
  ]
}

resource "layerform_layer" "metricbeat" {
  name = "metric"
  files = [
    "layers/metricbeat.tf",
    "layers/metricbeat/main.tf",
    "layers/metricbeat/variables.tf",
  ]
  dependencies = [
    layerform_layer.kibana.id
  ]
}

resource "layerform_layer" "filebeat" {
  name = "filebeat"
  files = [
    "layers/filebeat.tf",
    "layers/filebeat/main.tf",
    "layers/filebeat/variables.tf",
  ]
  dependencies = [
    layerform_layer.kibana.id
  ]
}
