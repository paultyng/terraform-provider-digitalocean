---
page_title: "DigitalOcean: digitalocean_database_connection_pool"
subcategory: "Databases"
---

# digitalocean\_database\_connection\_pool

Provides a DigitalOcean database connection pool resource.

## Example Usage

### Create a new PostgreSQL database connection pool
```hcl
resource "digitalocean_database_connection_pool" "pool-01" {
  cluster_id = digitalocean_database_cluster.postgres-example.id
  name       = "pool-01"
  mode       = "transaction"
  size       = 20
  db_name    = "defaultdb"
  user       = "doadmin"
}

resource "digitalocean_database_cluster" "postgres-example" {
  name       = "example-postgres-cluster"
  engine     = "pg"
  version    = "15"
  size       = "db-s-1vcpu-1gb"
  region     = "nyc1"
  node_count = 1
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The ID of the source database cluster. Note: This must be a PostgreSQL cluster.
* `name` - (Required) The name for the database connection pool.
* `mode` - (Required) The PGBouncer transaction mode for the connection pool. The allowed values are session, transaction, and statement.
* `size` - (Required) The desired size of the PGBouncer connection pool.
* `db_name` - (Required) The database for use with the connection pool.
* `user` - (Optional) The name of the database user for use with the connection pool. When excluded, all sessions connect to the database as the inbound user.
* `skip_if_exists` - (Optional) Skips creating a new connection pool if the connection pool already exists.

## Attributes Reference

In addition to the above arguments, the following attributes are exported:

* `id` - The ID of the database connection pool.
* `host` - The hostname used to connect to the database connection pool.
* `private_host` - Same as `host`, but only accessible from resources within the account and in the same region.
* `port` - Network port that the database connection pool is listening on.
* `uri` - The full URI for connecting to the database connection pool.
* `private_uri` - Same as `uri`, but only accessible from resources within the account and in the same region.
* `password` - Password for the connection pool's user.

## Import

Database connection pools can be imported using the `id` of the source database cluster
and the `name` of the connection pool joined with a comma. For example:

```
terraform import digitalocean_database_connection_pool.pool-01 245bcfd0-7f31-4ce6-a2bc-475a116cca97,pool-01
```
