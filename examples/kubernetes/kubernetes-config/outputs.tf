output "load_balancer_ip" {
  value = one(data.kubernetes_service.nginx-ingress-controller.status.0.load_balancer.0.ingress[*].ip)
}
