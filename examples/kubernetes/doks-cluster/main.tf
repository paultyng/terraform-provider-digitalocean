terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = ">= 2.4.0"
    }
  }
}

provider "digitalocean" {
  # Provider is configured using environment variables:
  # DIGITALOCEAN_TOKEN, DIGITALOCEAN_ACCESS_TOKEN
}

data "digitalocean_kubernetes_versions" "current" {
  version_prefix = var.cluster_version
}

resource "digitalocean_kubernetes_cluster" "primary" {
  name    = var.cluster_name
  region  = var.cluster_region
  version = data.digitalocean_kubernetes_versions.current.latest_version

  node_pool {
    name       = "default"
    size       = var.worker_size
    node_count = var.worker_count
  }

  # Ignore the tags/labels added by DigitalOcean
  lifecycle {
    ignore_changes = [
      tags, node_pool["labels"], node_pool["tags"]
    ]
  }
}
