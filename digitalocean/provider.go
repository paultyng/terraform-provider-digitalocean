package digitalocean

import (
	"context"

	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/account"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/app"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/cdn"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/certificate"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/database"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/domain"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/droplet"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/dropletautoscale"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/firewall"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/image"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/kubernetes"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/loadbalancer"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/monitoring"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/project"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/region"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/registry"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/reservedip"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/reservedipv6"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/size"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/snapshot"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/spaces"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/sshkey"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/tag"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/uptime"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/volume"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/vpc"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/vpcpeering"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a schema.Provider for DigitalOcean.
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"DIGITALOCEAN_TOKEN",
					"DIGITALOCEAN_ACCESS_TOKEN",
				}, nil),
				Description: "The token key for API operations.",
			},
			"api_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DIGITALOCEAN_API_URL", "https://api.digitalocean.com"),
				Description: "The URL to use for the DigitalOcean API.",
			},
			"spaces_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SPACES_ENDPOINT_URL", "https://{{.Region}}.digitaloceanspaces.com"),
				Description: "The URL to use for the DigitalOcean Spaces API.",
			},
			"spaces_access_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SPACES_ACCESS_KEY_ID", nil),
				Description: "The access key ID for Spaces API operations.",
			},
			"spaces_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SPACES_SECRET_ACCESS_KEY", nil),
				Description: "The secret access key for Spaces API operations.",
			},
			"requests_per_second": {
				Type:        schema.TypeFloat,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DIGITALOCEAN_REQUESTS_PER_SECOND", 0.0),
				Description: "The rate of requests per second to limit the HTTP client.",
			},
			"http_retry_max": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DIGITALOCEAN_HTTP_RETRY_MAX", 4),
				Description: "The maximum number of retries on a failed API request.",
			},
			"http_retry_wait_min": {
				Type:        schema.TypeFloat,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DIGITALOCEAN_HTTP_RETRY_WAIT_MIN", 1.0),
				Description: "The minimum wait time (in seconds) between failed API requests.",
			},
			"http_retry_wait_max": {
				Type:        schema.TypeFloat,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DIGITALOCEAN_HTTP_RETRY_WAIT_MAX", 30.0),
				Description: "The maximum wait time (in seconds) between failed API requests.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"digitalocean_account":                  account.DataSourceDigitalOceanAccount(),
			"digitalocean_app":                      app.DataSourceDigitalOceanApp(),
			"digitalocean_certificate":              certificate.DataSourceDigitalOceanCertificate(),
			"digitalocean_container_registry":       registry.DataSourceDigitalOceanContainerRegistry(),
			"digitalocean_database_cluster":         database.DataSourceDigitalOceanDatabaseCluster(),
			"digitalocean_database_connection_pool": database.DataSourceDigitalOceanDatabaseConnectionPool(),
			"digitalocean_database_ca":              database.DataSourceDigitalOceanDatabaseCA(),
			"digitalocean_database_replica":         database.DataSourceDigitalOceanDatabaseReplica(),
			"digitalocean_database_user":            database.DataSourceDigitalOceanDatabaseUser(),
			"digitalocean_domain":                   domain.DataSourceDigitalOceanDomain(),
			"digitalocean_domains":                  domain.DataSourceDigitalOceanDomains(),
			"digitalocean_droplet":                  droplet.DataSourceDigitalOceanDroplet(),
			"digitalocean_droplet_autoscale":        dropletautoscale.DataSourceDigitalOceanDropletAutoscale(),
			"digitalocean_droplets":                 droplet.DataSourceDigitalOceanDroplets(),
			"digitalocean_droplet_snapshot":         snapshot.DataSourceDigitalOceanDropletSnapshot(),
			"digitalocean_firewall":                 firewall.DataSourceDigitalOceanFirewall(),
			"digitalocean_floating_ip":              reservedip.DataSourceDigitalOceanFloatingIP(),
			"digitalocean_image":                    image.DataSourceDigitalOceanImage(),
			"digitalocean_images":                   image.DataSourceDigitalOceanImages(),
			"digitalocean_kubernetes_cluster":       kubernetes.DataSourceDigitalOceanKubernetesCluster(),
			"digitalocean_kubernetes_versions":      kubernetes.DataSourceDigitalOceanKubernetesVersions(),
			"digitalocean_loadbalancer":             loadbalancer.DataSourceDigitalOceanLoadbalancer(),
			"digitalocean_project":                  project.DataSourceDigitalOceanProject(),
			"digitalocean_projects":                 project.DataSourceDigitalOceanProjects(),
			"digitalocean_record":                   domain.DataSourceDigitalOceanRecord(),
			"digitalocean_records":                  domain.DataSourceDigitalOceanRecords(),
			"digitalocean_region":                   region.DataSourceDigitalOceanRegion(),
			"digitalocean_regions":                  region.DataSourceDigitalOceanRegions(),
			"digitalocean_reserved_ip":              reservedip.DataSourceDigitalOceanReservedIP(),
			"digitalocean_reserved_ipv6":            reservedipv6.DataSourceDigitalOceanReservedIPV6(),
			"digitalocean_sizes":                    size.DataSourceDigitalOceanSizes(),
			"digitalocean_spaces_bucket":            spaces.DataSourceDigitalOceanSpacesBucket(),
			"digitalocean_spaces_buckets":           spaces.DataSourceDigitalOceanSpacesBuckets(),
			"digitalocean_spaces_bucket_object":     spaces.DataSourceDigitalOceanSpacesBucketObject(),
			"digitalocean_spaces_bucket_objects":    spaces.DataSourceDigitalOceanSpacesBucketObjects(),
			"digitalocean_spaces_key":               spaces.DataSourceDigitalOceanSpacesKey(),
			"digitalocean_ssh_key":                  sshkey.DataSourceDigitalOceanSSHKey(),
			"digitalocean_ssh_keys":                 sshkey.DataSourceDigitalOceanSSHKeys(),
			"digitalocean_tag":                      tag.DataSourceDigitalOceanTag(),
			"digitalocean_tags":                     tag.DataSourceDigitalOceanTags(),
			"digitalocean_volume_snapshot":          snapshot.DataSourceDigitalOceanVolumeSnapshot(),
			"digitalocean_volume":                   volume.DataSourceDigitalOceanVolume(),
			"digitalocean_vpc":                      vpc.DataSourceDigitalOceanVPC(),
			"digitalocean_vpc_peering":              vpcpeering.DataSourceDigitalOceanVPCPeering(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"digitalocean_app":                                   app.ResourceDigitalOceanApp(),
			"digitalocean_certificate":                           certificate.ResourceDigitalOceanCertificate(),
			"digitalocean_container_registry":                    registry.ResourceDigitalOceanContainerRegistry(),
			"digitalocean_container_registry_docker_credentials": registry.ResourceDigitalOceanContainerRegistryDockerCredentials(),
			"digitalocean_cdn":                                   cdn.ResourceDigitalOceanCDN(),
			"digitalocean_database_cluster":                      database.ResourceDigitalOceanDatabaseCluster(),
			"digitalocean_database_connection_pool":              database.ResourceDigitalOceanDatabaseConnectionPool(),
			"digitalocean_database_db":                           database.ResourceDigitalOceanDatabaseDB(),
			"digitalocean_database_firewall":                     database.ResourceDigitalOceanDatabaseFirewall(),
			"digitalocean_database_replica":                      database.ResourceDigitalOceanDatabaseReplica(),
			"digitalocean_database_user":                         database.ResourceDigitalOceanDatabaseUser(),
			"digitalocean_database_redis_config":                 database.ResourceDigitalOceanDatabaseRedisConfig(),
			"digitalocean_database_postgresql_config":            database.ResourceDigitalOceanDatabasePostgreSQLConfig(),
			"digitalocean_database_mysql_config":                 database.ResourceDigitalOceanDatabaseMySQLConfig(),
			"digitalocean_database_mongodb_config":               database.ResourceDigitalOceanDatabaseMongoDBConfig(),
			"digitalocean_database_kafka_config":                 database.ResourceDigitalOceanDatabaseKafkaConfig(),
			"digitalocean_database_opensearch_config":            database.ResourceDigitalOceanDatabaseOpensearchConfig(),
			"digitalocean_database_kafka_topic":                  database.ResourceDigitalOceanDatabaseKafkaTopic(),
			"digitalocean_domain":                                domain.ResourceDigitalOceanDomain(),
			"digitalocean_droplet":                               droplet.ResourceDigitalOceanDroplet(),
			"digitalocean_droplet_autoscale":                     dropletautoscale.ResourceDigitalOceanDropletAutoscale(),
			"digitalocean_droplet_snapshot":                      snapshot.ResourceDigitalOceanDropletSnapshot(),
			"digitalocean_firewall":                              firewall.ResourceDigitalOceanFirewall(),
			"digitalocean_floating_ip":                           reservedip.ResourceDigitalOceanFloatingIP(),
			"digitalocean_floating_ip_assignment":                reservedip.ResourceDigitalOceanFloatingIPAssignment(),
			"digitalocean_kubernetes_cluster":                    kubernetes.ResourceDigitalOceanKubernetesCluster(),
			"digitalocean_kubernetes_node_pool":                  kubernetes.ResourceDigitalOceanKubernetesNodePool(),
			"digitalocean_loadbalancer":                          loadbalancer.ResourceDigitalOceanLoadbalancer(),
			"digitalocean_monitor_alert":                         monitoring.ResourceDigitalOceanMonitorAlert(),
			"digitalocean_project":                               project.ResourceDigitalOceanProject(),
			"digitalocean_project_resources":                     project.ResourceDigitalOceanProjectResources(),
			"digitalocean_record":                                domain.ResourceDigitalOceanRecord(),
			"digitalocean_reserved_ip":                           reservedip.ResourceDigitalOceanReservedIP(),
			"digitalocean_reserved_ip_assignment":                reservedip.ResourceDigitalOceanReservedIPAssignment(),
			"digitalocean_reserved_ipv6":                         reservedipv6.ResourceDigitalOceanReservedIPV6(),
			"digitalocean_reserved_ipv6_assignment":              reservedipv6.ResourceDigitalOceanReservedIPV6Assignment(),
			"digitalocean_spaces_bucket":                         spaces.ResourceDigitalOceanBucket(),
			"digitalocean_spaces_bucket_cors_configuration":      spaces.ResourceDigitalOceanBucketCorsConfiguration(),
			"digitalocean_spaces_bucket_object":                  spaces.ResourceDigitalOceanSpacesBucketObject(),
			"digitalocean_spaces_bucket_policy":                  spaces.ResourceDigitalOceanSpacesBucketPolicy(),
			"digitalocean_spaces_key":                            spaces.ResourceDigitalOceanSpacesKey(),
			"digitalocean_ssh_key":                               sshkey.ResourceDigitalOceanSSHKey(),
			"digitalocean_tag":                                   tag.ResourceDigitalOceanTag(),
			"digitalocean_uptime_check":                          uptime.ResourceDigitalOceanUptimeCheck(),
			"digitalocean_uptime_alert":                          uptime.ResourceDigitalOceanUptimeAlert(),
			"digitalocean_volume":                                volume.ResourceDigitalOceanVolume(),
			"digitalocean_volume_attachment":                     volume.ResourceDigitalOceanVolumeAttachment(),
			"digitalocean_volume_snapshot":                       snapshot.ResourceDigitalOceanVolumeSnapshot(),
			"digitalocean_vpc":                                   vpc.ResourceDigitalOceanVPC(),
			"digitalocean_vpc_peering":                           vpcpeering.ResourceDigitalOceanVPCPeering(),
			"digitalocean_custom_image":                          image.ResourceDigitalOceanCustomImage(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		client, err := providerConfigure(d, terraformVersion)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return client, nil
	}

	return p
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	conf := config.Config{
		Token:             d.Get("token").(string),
		APIEndpoint:       d.Get("api_endpoint").(string),
		AccessID:          d.Get("spaces_access_id").(string),
		SecretKey:         d.Get("spaces_secret_key").(string),
		RequestsPerSecond: d.Get("requests_per_second").(float64),
		HTTPRetryMax:      d.Get("http_retry_max").(int),
		HTTPRetryWaitMin:  d.Get("http_retry_wait_min").(float64),
		HTTPRetryWaitMax:  d.Get("http_retry_wait_max").(float64),
		TerraformVersion:  terraformVersion,
	}

	if endpoint, ok := d.GetOk("spaces_endpoint"); ok {
		conf.SpacesAPIEndpoint = endpoint.(string)
	}

	return conf.Client()
}
