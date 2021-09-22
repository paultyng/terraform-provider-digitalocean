variable "primary_cluster" {
  type = object({
    id                     = string
    name                   = string
    endpoint               = string
    token                  = string
    cluster_ca_certificate = string
    raw_config             = string
  })
}

# Helm chart deployment can sometimes take longer than the default 5 minutes
variable "nginx_ingress_helm_timeout_seconds" {
  default     = 600
}

variable "write_kubeconfig" {
  type        = bool
  default     = false
}