variable "cluster_name" {
  default = "test-cluster"
}

variable "cluster_region" {
  default = "nyc3"
}

variable "cluster_version" {
  default = "1.19"
}

variable "worker_count" {
  default = 3
}

variable "worker_size" {
  default = "s-2vcpu-4gb"
}

variable "write_kubeconfig" {
  type        = bool
  default     = false
}