locals {
  load_balancer_ip = one(data.kubernetes_service.nginx-ingress-controller.status.0.load_balancer.0.ingress[*].ip)
}

output "test_url" {
  value = local.load_balancer_ip != null ? format("http://%s/test", local.load_balancer_ip) : "[PENDING]"
}
