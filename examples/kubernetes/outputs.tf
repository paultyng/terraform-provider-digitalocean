output "cluster_name" {
  value = module.doks-cluster.primary_cluster.name
}

output "kubeconfig_path" {
  value = var.write_kubeconfig ? abspath("${path.root}/kubeconfig") : "none"
}

output "test_url_status" {
  value = module.kubernetes-config.load_balancer_ip != null ? "available" : "pending load balancer IP address assignment"
}

output "test_url" {
  value = module.kubernetes-config.load_balancer_ip != null ? "http://${module.kubernetes-config.load_balancer_ip}/test" : null
}
