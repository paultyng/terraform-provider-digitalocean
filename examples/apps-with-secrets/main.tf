// Problem: When passing a secret environment variable to the apps resource,
// the Apps resource will encrypt the value and store that to state. This results 
// in `terraform plan` comparing the encrypted and the plain variable.
// As a result, always thinking there are changes to the environment variables.

// Work around: Manually copy the encrypted values returned by the API into a tfvars file:
// - On the first run, use `terraform apply -var "secret_token=thisisverysecureandwillbeencrypted"`
// - Then, retrieve the encrypted value using: `doctl apps get <app-uuid>  -o json` or from passing 
//   TF_ACC=debug in the terraform apply command in the first step. 
// - Next, put the encrypted value into app-platform-encrypted.tvars
// - Now running terraform apply -var-file="app-platform-encrypted.tvars" does not detect any changes.
// - If I want to replace that value, I can run terraform apply -var "secret_token=a-new-secret" and the change is detected.


terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = ">= 2.8.0"
    }
  }
}

provider "digitalocean" {
  # You need to set this in your .bashrc
  # export DIGITALOCEAN_TOKEN="Your API TOKEN"
  #
}

resource "digitalocean_app" "golang-sample" {
  spec {
    name   = "golang-sample"
    region = "ams"

    service {
      name               = "go-service"
      environment_slug   = "go"
      instance_count     = 1
      instance_size_slug = "professional-xs"

        env {
            key   = "APP_ENV"
            value = "production"
            scope = "RUN_AND_BUILD_TIME"
            type  = "GENERAL"
        }

        env {
            key   = "SECRET_TOKEN"
            value = (var.encrypted_secret_token != "" ? var.encrypted_secret_token : var.secret_token)
            scope = "RUN_TIME"
            type  = "SECRET"
        }

        git {
            repo_clone_url = "https://github.com/digitalocean/sample-golang.git"
            branch         = "main"
        }
    }

    database {
      name       = "db"
      engine     = "PG"
      production = false
    }
  }
}