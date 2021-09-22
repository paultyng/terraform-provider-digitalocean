output "primary_cluster" {
  value = {
    id                     = digitalocean_kubernetes_cluster.primary.id
    name                   = digitalocean_kubernetes_cluster.primary.name
    endpoint               = digitalocean_kubernetes_cluster.primary.endpoint
    token                  = digitalocean_kubernetes_cluster.primary.kube_config[0].token
    cluster_ca_certificate = digitalocean_kubernetes_cluster.primary.kube_config[0].cluster_ca_certificate
    raw_config             = digitalocean_kubernetes_cluster.primary.kube_config[0].raw_config
  }
}
