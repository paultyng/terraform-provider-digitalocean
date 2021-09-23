variable "cluster_name" {
  type = string
}

variable "cluster_id" {
  type = string
}

variable "write_kubeconfig" {
  type        = bool
  default     = false
}

# Helm chart deployment can sometimes take longer than the default 5 minutes
variable "nginx_ingress_helm_timeout_seconds" {
  type        = number
  default     = 600
}
