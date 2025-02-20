package spaces_test

import (
	"fmt"
	"testing"

	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/acceptance"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDigitalOceanSpacesKey_basic(t *testing.T) {
	expectedName := acceptance.RandomTestName()
	bucketName := acceptance.RandomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanSpacesKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDigitalOceanSpacesKeyConfig(expectedName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "name", expectedName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.bucket", bucketName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.permission", "read"),
				),
			},
		},
	})
}

func TestAccDigitalOceanSpacesKey_updateGrant(t *testing.T) {
	expectedName := acceptance.RandomTestName()
	expectedNewName := acceptance.RandomTestName()
	bucketName := acceptance.RandomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanSpacesKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDigitalOceanSpacesKeyConfig(expectedName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "name", expectedName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.bucket", bucketName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.permission", "read"),
				),
			},
			{
				Config: testAccDigitalOceanSpacesKeyConfigWithGrantUpdate(expectedNewName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "name", expectedNewName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.bucket", bucketName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.permission", "read"),
				),
			},
		},
	})
}

func TestAccDigitalOceanSpacesKey_multipleGrants(t *testing.T) {
	expectedName := acceptance.RandomTestName()
	bucketName := acceptance.RandomTestName()
	bucketName2 := acceptance.RandomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanSpacesKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDigitalOceanSpacesKeyConfigMultipleGrants(expectedName, bucketName, bucketName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "name", expectedName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.bucket", bucketName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.permission", "read"),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.1.bucket", bucketName2),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.1.permission", "readwrite"),
				),
			},
			{
				Config: testAccDigitalOceanSpacesKeyConfig(expectedName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "name", expectedName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.bucket", bucketName),
					resource.TestCheckResourceAttr(
						"digitalocean_spaces_key.key", "grant.0.permission", "read"),
					resource.TestCheckNoResourceAttr(
						"digitalocean_spaces_key.key", "grant.1.bucket"),
					resource.TestCheckNoResourceAttr(
						"digitalocean_spaces_key.key", "grant.1.permission"),
				),
			},
		},
	})
}

func testAccDigitalOceanSpacesKeyConfig(name, bucket string) string {
	return fmt.Sprintf(`
resource "digitalocean_spaces_bucket" "bucket" {
  name   = "%s"
  acl    = "public-read"
  region = "sfo3"
}
resource "digitalocean_spaces_key" "key" {
  name = "%s"
  grant {
    bucket     = digitalocean_spaces_bucket.bucket.name
    permission = "read"
  }
}
`, bucket, name)
}

func testAccDigitalOceanSpacesKeyConfigWithGrantUpdate(name, bucket string) string {
	return fmt.Sprintf(`
resource "digitalocean_spaces_key" "key" {
  name = "%s"
  grant {
    bucket     = "%s"
    permission = "read"
  }
}
`, name, bucket)
}

func testAccDigitalOceanSpacesKeyConfigMultipleGrants(name, bucket, bucket2 string) string {
	return fmt.Sprintf(`
resource "digitalocean_spaces_bucket" "bucket" {
  name   = "%s"
  acl    = "public-read"
  region = "sfo3"
}
resource "digitalocean_spaces_bucket" "bucket2" {
  name   = "%s"
  acl    = "public-read"
  region = "sfo3"
}
resource "digitalocean_spaces_key" "key" {
  name = "%s"
  grant {
    bucket     = digitalocean_spaces_bucket.bucket.name
    permission = "read"
  }
  grant {
    bucket     = digitalocean_spaces_bucket.bucket2.name
    permission = "readwrite"
  }
}
`, bucket, bucket2, name)
}
