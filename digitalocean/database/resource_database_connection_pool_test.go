package database_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/acceptance"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDigitalOceanDatabaseConnectionPool_Basic(t *testing.T) {
	var databaseConnectionPool godo.DatabasePool
	databaseName := acceptance.RandomTestName()
	databaseConnectionPoolName := acceptance.RandomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanDatabaseConnectionPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseConnectionPoolConfigBasic, databaseName, databaseConnectionPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseConnectionPoolExists("digitalocean_database_connection_pool.pool-01", &databaseConnectionPool),
					testAccCheckDigitalOceanDatabaseConnectionPoolAttributes(&databaseConnectionPool, databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "name", databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "size", "10"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "mode", "transaction"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "db_name", "defaultdb"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "user", "doadmin"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "private_host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "port"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "uri"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "private_uri"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "password"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseConnectionPoolConfigUpdated, databaseName, databaseConnectionPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseConnectionPoolExists("digitalocean_database_connection_pool.pool-01", &databaseConnectionPool),
					testAccCheckDigitalOceanDatabaseConnectionPoolAttributes(&databaseConnectionPool, databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "name", databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "mode", "session"),
				),
			},
		},
	})
}

func TestAccDigitalOceanDatabaseConnectionPool_InboundUser(t *testing.T) {

	var databaseConnectionPool godo.DatabasePool
	databaseName := acceptance.RandomTestName()
	databaseConnectionPoolName := acceptance.RandomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanDatabaseConnectionPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseConnectionPoolConfigInboundUser, databaseName, databaseConnectionPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseConnectionPoolExists("digitalocean_database_connection_pool.pool-01", &databaseConnectionPool),
					testAccCheckDigitalOceanDatabaseConnectionPoolAttributes(&databaseConnectionPool, databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "name", databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "size", "10"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "mode", "transaction"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "db_name", "defaultdb"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "user", ""),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "private_host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "port"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "uri"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "private_uri"),
				),
			},
		},
	})
}

func TestAccDigitalOceanDatabaseConnectionPool_BadModeName(t *testing.T) {
	databaseName := acceptance.RandomTestName()
	databaseConnectionPoolName := acceptance.RandomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanDatabaseConnectionPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testAccCheckDigitalOceanDatabaseConnectionPoolConfigBad, databaseName, databaseConnectionPoolName),
				ExpectError: regexp.MustCompile(`expected mode to be one of`),
			},
		},
	})
}

func testAccCheckDigitalOceanDatabaseConnectionPoolDestroy(s *terraform.State) error {
	client := acceptance.TestAccProvider.Meta().(*config.CombinedConfig).GodoClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "digitalocean_database_connection_pool" {
			continue
		}
		clusterId := rs.Primary.Attributes["cluster_id"]
		name := rs.Primary.Attributes["name"]
		// Try to find the database connection_pool
		_, _, err := client.Databases.GetPool(context.Background(), clusterId, name)

		if err == nil {
			return fmt.Errorf("DatabaseConnectionPool still exists")
		}
	}

	return nil
}

func TestAccDigitalOceanDatabaseConnectionPool_SkipIfExistIsPassed(t *testing.T) {
	databaseName := acceptance.RandomTestName()
	databaseConnectionPoolName := acceptance.RandomTestName()
	var databaseConnectionPool godo.DatabasePool

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      testAccCheckDigitalOceanDatabaseConnectionPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckDigitalOceanDatabaseConnectionPoolConfigBasicSkipIfExists, databaseName, databaseConnectionPoolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDigitalOceanDatabaseConnectionPoolExists("digitalocean_database_connection_pool.pool-01", &databaseConnectionPool),
					testAccCheckDigitalOceanDatabaseConnectionPoolAttributes(&databaseConnectionPool, databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "name", databaseConnectionPoolName),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "size", "10"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "mode", "transaction"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "db_name", "defaultdb"),
					resource.TestCheckResourceAttr(
						"digitalocean_database_connection_pool.pool-01", "user", "doadmin"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "private_host"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "port"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "uri"),
					resource.TestCheckResourceAttrSet(
						"digitalocean_database_connection_pool.pool-01", "private_uri"),
				),
			},
		},
	})
}

func testAccCheckDigitalOceanDatabaseConnectionPoolExists(n string, database *godo.DatabasePool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DatabaseConnectionPool ID is set")
		}

		client := acceptance.TestAccProvider.Meta().(*config.CombinedConfig).GodoClient()
		clusterId := rs.Primary.Attributes["cluster_id"]
		name := rs.Primary.Attributes["name"]

		foundDatabaseConnectionPool, _, err := client.Databases.GetPool(context.Background(), clusterId, name)

		if err != nil {
			return err
		}

		if foundDatabaseConnectionPool.Name != name {
			return fmt.Errorf("DatabaseConnectionPool not found")
		}

		*database = *foundDatabaseConnectionPool

		return nil
	}
}

func testAccCheckDigitalOceanDatabaseConnectionPoolAttributes(databaseConnectionPool *godo.DatabasePool, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if databaseConnectionPool.Name != name {
			return fmt.Errorf("Bad name: %s", databaseConnectionPool.Name)
		}

		return nil
	}
}

const testAccCheckDigitalOceanDatabaseConnectionPoolConfigBasic = `
resource "digitalocean_database_cluster" "foobar" {
  name       = "%s"
  engine     = "pg"
  version    = "15"
  size       = "db-s-1vcpu-1gb"
  region     = "nyc1"
  node_count = 1
}

resource "digitalocean_database_connection_pool" "pool-01" {
  cluster_id = digitalocean_database_cluster.foobar.id
  name       = "%s"
  mode       = "transaction"
  size       = 10
  db_name    = "defaultdb"
  user       = "doadmin"
}`

const testAccCheckDigitalOceanDatabaseConnectionPoolConfigBasicSkipIfExists = `
resource "digitalocean_database_cluster" "foobar" {
  name       = "%s"
  engine     = "pg"
  version    = "14"
  size       = "db-s-1vcpu-1gb"
  region     = "nyc1"
  node_count = 1
}

resource "digitalocean_database_connection_pool" "pool-01" {
  cluster_id     = digitalocean_database_cluster.foobar.id
  name           = "%s"
  mode           = "transaction"
  size           = 10
  db_name        = "defaultdb"
  user           = "doadmin"
  skip_if_exists = true
}`

const testAccCheckDigitalOceanDatabaseConnectionPoolConfigUpdated = `
resource "digitalocean_database_cluster" "foobar" {
  name       = "%s"
  engine     = "pg"
  version    = "15"
  size       = "db-s-1vcpu-1gb"
  region     = "nyc1"
  node_count = 1
}

resource "digitalocean_database_connection_pool" "pool-01" {
  cluster_id = digitalocean_database_cluster.foobar.id
  name       = "%s"
  mode       = "session"
  size       = 10
  db_name    = "defaultdb"
}`

const testAccCheckDigitalOceanDatabaseConnectionPoolConfigBad = `
resource "digitalocean_database_cluster" "foobar" {
  name       = "%s"
  engine     = "pg"
  version    = "15"
  size       = "db-s-1vcpu-1gb"
  region     = "nyc1"
  node_count = 1
}

resource "digitalocean_database_connection_pool" "pool-01" {
  cluster_id = digitalocean_database_cluster.foobar.id
  name       = "%s"
  mode       = "transactional"
  size       = 10
  db_name    = "defaultdb"
  user       = "doadmin"
}`

const testAccCheckDigitalOceanDatabaseConnectionPoolConfigInboundUser = `
resource "digitalocean_database_cluster" "foobar" {
  name       = "%s"
  engine     = "pg"
  version    = "15"
  size       = "db-s-1vcpu-1gb"
  region     = "nyc1"
  node_count = 1
}

resource "digitalocean_database_connection_pool" "pool-01" {
  cluster_id = digitalocean_database_cluster.foobar.id
  name       = "%s"
  mode       = "transaction"
  size       = 10
  db_name    = "defaultdb"
}`
