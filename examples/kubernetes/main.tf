terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = ">= 2.4.0"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
      version = ">= 2.0.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.0.1"
    }
  }
}

locals {
  cluster_name = "${var.cluster_name}-${var.cluster_region}"
}

module "doks-cluster" {
  source             = "./doks-cluster"
  cluster_name       = local.cluster_name
  cluster_region     = var.cluster_region
  cluster_version    = var.cluster_version

  worker_size        = var.worker_size
  worker_count       = var.worker_count
}

module "kubernetes-config" {
  source           = "./kubernetes-config"
  primary_cluster  = module.doks-cluster.primary_cluster

  write_kubeconfig = var.write_kubeconfig
}
