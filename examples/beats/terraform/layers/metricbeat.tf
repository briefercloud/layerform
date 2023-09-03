module "metricbeat" {
  source = "./metricbeat"

  aws_region           = module.kibana.aws_region
  bucket_name          = module.kibana.bucket_name
  kibana_folder        = module.kibana.kibana_folder
  elasticsearch_folder = module.elasticsearch.elasticsearch_folder
}
