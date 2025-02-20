# Developing the Provider

Testing Your Changes Locally
----------------------------

### 1. Rebuilding the Provider
1.1 If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ ls -la $GOPATH/bin/terraform-provider-digitalocean
...
```

Remember to rebuild the provider with `make build` to apply any new changes.

1.2 In order to check changes you made locally to the provider, you can use the binary you just compiled by adding the following
to your `~/.terraformrc` file. This is valid for Terraform 0.14+. Please see
[Terraform's documentation](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers)
for more details.

```
provider_installation {

  # Use /home/developer/go/bin as an overridden package directory
  # for the digitalocean/digitalocean provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # digitalocean provider plugin in the given directory.
  dev_overrides {
    "digitalocean/digitalocean" = "/home/developer/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

### 2. Creating a Sample Terraform Configuration
2.1 From the root of the project, create a directory for your Terraform configuration:

```console
mkdir -p examples/my-tf
```

2.2 Create a new Terraform configuration file:

``` console
touch examples/my-tf/main.tf
```

2.3 Populate the main.tf file with the following example configuration.  
* [Available versions for the DigitalOcean Terraform provider](https://registry.terraform.io/providers/digitalocean/digitalocean/latest)  
* Make sure to update the token value with your own [DigitalOcean token](https://docs.digitalocean.com/reference/api/create-personal-access-token).

```console
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = ">= 2.44.1"
    }
  }
}

provider "digitalocean" {
  token = "dop_v1_12345d7ce104413a59023656565"
}

resource "digitalocean_droplet" "foobar-my-tf" {
  image  = "ubuntu-22-04-x64"
  name   = "tf-acc-test-my-tf"
  region = "nyc3"
  size   = "s-1vcpu-1gb"
}
``` 
2.4 Before using the Terraform configuration, you need to initialize your working directory.
```console
cd examples/my-tf
terraform init
```

### 3. Running Terraform Commands
3.1 Navigate to the working directory:

```console
cd examples/my-tf
```

3.2 To see the changes that will be applied run:

```console
terraform plan
```

3.3 To apply the changes run:

```console
terraform apply
```
This will interact with your DigitalOcean account, using the token you provided in the `main.tf` configuration file.  
Once you've finished testing, it's a good practice to clean up any resources you created to avoid incurring charges.

### 4. Debugging and Logging
You can add logs to your code with `log.Printf()`. Remember to run `make build` to apply changes.

If you'd like to see more detailed logs for debugging, you can set the `TF_LOG` environment variable to `DEBUG` or `TRACE`.

``` console
export TF_LOG=DEBUG
export TF_LOG=TRACE
```

After setting the log level, you can run `terraform plan` or `terraform apply` again to see more detailed output.

Provider Automation Tests
--------------------
### Running unit tests
In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

### Running Acceptance Tests

Rebuild the provider before running acceptance tests.

Please be aware that **running ALL acceptance tests will take a significant amount of time and could be expensive**, as they interact with your **real DigitalOcean account**. For this reason, it is highly recommended to run only one acceptance test at a time to minimize both time and cost.

- It is preferable to run one acceptance test at a time.
In order to run a specific acceptance test, use the `TESTARGS` environment variable. For example, the following command will run `TestAccDigitalOceanDomain_Basic` acceptance test only:

```sh
$ make testacc TESTARGS='-run=TestAccDigitalOceanDomain_Basic'
```

- All acceptance tests for a specific package can be run by setting the `PKG_NAME` environment variable. For example:

```sh
$ make testacc PKG_NAME=digitalocean/account
```

- In order to run the full suite of acceptance tests, run `make testacc`.

**Note:** Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

For information about writing acceptance tests, see the main Terraform [contributing guide](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#writing-acceptance-tests).

Releasing the Provider
----------------------
The dedicated DigitalOcean team is responsible for releasing the provider.

#### To release the provider:

1. Pull the latest changes from the `main` branch

   ```bash
   git checkout main; git pull
   ```

2. Ensure that each PR has the correct label.
    - **bug**: For fixes related to issues or bugs.
    - **enhancement**: For changes or improvements to existing resources.
    - **feature**: For the addition of new resources or functionality.
    - **misc**: For non-user-impacting changes, such as updates to tests or documentation. 

&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; These labels categorize PRs in the [GitHub Release Notes](https://github.com/digitalocean/terraform-provider-digitalocean/releases). 

&nbsp; &nbsp; &nbsp; &nbsp; &nbsp; You can get a list of PRs to release with [github-changelog-generator](https://github.com/digitalocean/github-changelog-generator). It shows the changes since the last release.

3. Determine the new release version.  
   PR labels determine the release version type: *patch*, *minor*, or *major*.  
   terraform-provider-digitalocean follows [semver](https://www.semver.org/) versioning
   semantics.  
   For example:
    - bug, misc, doc: A bug fix results in a *patch* version increment (e.g., 1.2.3 → 1.2.4).
    - feature, enhancement: A new feature results in a *minor* version increment (e.g., 1.2.3 → 1.3.0).
    - breaking-change: A breaking change results in a *major* version increment (e.g., 1.2.3 → 2.0.0).
4. Once all PRs to release have labels and the version increment is decided, create a new tag with the new version.  

   ```bash
   export new_version=<new-version>; git tag -m "release $new_version" -a "$new_version"
   ```

5. Push the tag:

   ```bash
   git push "$origin" tag "$new_version"
   ```

&nbsp; &nbsp; &nbsp; &nbsp; &nbsp;This repository contains a GitHub Action configured to automatically build and
publish assets for release when a tag is pushed that matches the pattern `v*`
(ie. `v0.1.0`).

&nbsp; &nbsp; &nbsp; &nbsp; &nbsp;A [Goreleaser](https://goreleaser.com/) configuration is provided that produces
build artifacts matching the [layout required](https://www.terraform.io/docs/registry/providers/publishing.html#manually-preparing-a-release)
to publish the provider in the Terraform Registry.

6. Publish the release.  
A new release will appear as drafts.  
Hit "Generate Release Notes" when publish the release.  
Mark the release as Latest.  
Once marked as published on the GitHub Releases page,
they will become available via the Terraform Registry.

Reviewing Pull Requests
-----------------------

Acceptance tests use the production API and create resources that incur costs.
Running the full suite of acceptance tests can also take quite a long time to run.
In order to prevent misuse, the acceptance tests for a pull request must be manually
triggered by a reviewer with write access to the repository.

To trigger a run of the acceptance tests for a PR, you may use the `/testacc` in a
comment. The `pkg` and `sha` arguments are required. This allows us to limit the
packages being tested for quick feedback and protect against timing attacks.
For example, to run the acceptance tests for the `droplet` package, the command
may look like:

    /testacc pkg=digitalocean/droplet sha=d358bd2418b4e30d7bdf2b98b4c151e357814c63

To run the entire suite, use `pkg=digitalocean`.

If multiple packages are to be tested, each command must be posted as a separate
comment. Only the first line of the comment is evaluated.

We leverage the [`peter-evans/slash-command-dispatch`](https://github.com/peter-evans/slash-command-dispatch)
GitHub Action for the command processing.
