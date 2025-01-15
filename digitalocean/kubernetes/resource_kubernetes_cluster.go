package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/tag"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	yaml "gopkg.in/yaml.v2"
)

var (
	MultipleNodePoolImportError = fmt.Errorf("Cluster contains multiple node pools. Manually add the `%s` tag to the pool that should be used as the default. Additional pools must be imported separately as 'digitalocean_kubernetes_node_pool' resources.", DigitaloceanKubernetesDefaultNodePoolTag)
)

func ResourceDigitalOceanKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDigitalOceanKubernetesClusterCreate,
		ReadContext:   resourceDigitalOceanKubernetesClusterRead,
		UpdateContext: resourceDigitalOceanKubernetesClusterUpdate,
		DeleteContext: resourceDigitalOceanKubernetesClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDigitalOceanKubernetesClusterImportState,
		},
		SchemaVersion: 3,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					// DO API V2 region slug is always lowercase
					return strings.ToLower(val.(string))
				},
				ValidateFunc: validation.NoZeroValues,
			},

			"surge_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"ha": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"registry_integration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: false,
			},

			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"vpc_uuid": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"cluster_subnet": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
			},

			"service_subnet": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
			},

			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tag.TagsSchema(),

			"maintenance_policy": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"any",
								"monday",
								"tuesday",
								"wednesday",
								"thursday",
								"friday",
								"saturday",
								"sunday"},
								true,
							),
						},
						"start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"duration": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"node_pool": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: nodePoolSchema(false),
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"kube_config": kubernetesConfigSchema(),

			"auto_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"urn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"destroy_all_associated_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"kubeconfig_expire_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("version", func(ctx context.Context, old, new, meta interface{}) bool {
				// "version" can only be upgraded to newer versions, so we must create a new resource
				// if it is decreased.
				newVer, err := version.NewVersion(new.(string))
				if err != nil {
					return false
				}

				oldVer, err := version.NewVersion(old.(string))
				if err != nil {
					return false
				}

				if newVer.LessThan(oldVer) {
					return true
				}
				return false
			}),
		),
	}
}

func kubernetesConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeList,
		Computed:  true,
		Sensitive: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"raw_config": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"host": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"cluster_ca_certificate": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"client_key": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"client_certificate": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"token": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"expires_at": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func resourceDigitalOceanKubernetesClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	pools := expandNodePools(d.Get("node_pool").([]interface{}))
	poolCreateRequests := make([]*godo.KubernetesNodePoolCreateRequest, len(pools))
	for i, pool := range pools {
		tags := append(pool.Tags, DigitaloceanKubernetesDefaultNodePoolTag)
		poolCreateRequests[i] = &godo.KubernetesNodePoolCreateRequest{
			Name:      pool.Name,
			Size:      pool.Size,
			Tags:      tags,
			Labels:    pool.Labels,
			Count:     pool.Count,
			AutoScale: pool.AutoScale,
			MinNodes:  pool.MinNodes,
			MaxNodes:  pool.MaxNodes,
			Taints:    pool.Taints,
		}
	}

	opts := &godo.KubernetesClusterCreateRequest{
		Name:         d.Get("name").(string),
		RegionSlug:   d.Get("region").(string),
		VersionSlug:  d.Get("version").(string),
		SurgeUpgrade: d.Get("surge_upgrade").(bool),
		HA:           d.Get("ha").(bool),
		Tags:         tag.ExpandTags(d.Get("tags").(*schema.Set).List()),
		NodePools:    poolCreateRequests,
	}

	if maint, ok := d.GetOk("maintenance_policy"); ok {
		maintPolicy, err := expandMaintPolicyOpts(maint.([]interface{}))
		if err != nil {
			return diag.Errorf("Error setting Kubernetes maintenance policy : %s", err)
		}
		opts.MaintenancePolicy = maintPolicy
	}

	if vpc, ok := d.GetOk("vpc_uuid"); ok {
		opts.VPCUUID = vpc.(string)
	}

	if clusterSubnet, ok := d.GetOk("cluster_subnet"); ok {
		opts.ClusterSubnet = clusterSubnet.(string)
	}

	if serviceSubnet, ok := d.GetOk("service_subnet"); ok {
		opts.ServiceSubnet = serviceSubnet.(string)
	}

	if autoUpgrade, ok := d.GetOk("auto_upgrade"); ok {
		opts.AutoUpgrade = autoUpgrade.(bool)
	}

	cluster, _, err := client.Kubernetes.Create(context.Background(), opts)
	if err != nil {
		return diag.Errorf("Error creating Kubernetes cluster: %s", err)
	}

	// set the cluster id
	d.SetId(cluster.ID)

	// wait for completion
	_, err = waitForKubernetesClusterCreate(client, d)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error creating Kubernetes cluster: %s", err)
	}

	if d.Get("registry_integration") == true {
		err = enableRegistryIntegration(client, cluster.ID)
		if err != nil {
			return diag.Errorf("Error enabling registry integration: %s", err)
		}
	}

	return resourceDigitalOceanKubernetesClusterRead(ctx, d, meta)
}

func resourceDigitalOceanKubernetesClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	cluster, resp, err := client.Kubernetes.Get(context.Background(), d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving Kubernetes cluster: %s", err)
	}

	return digitaloceanKubernetesClusterRead(client, cluster, d)
}

func digitaloceanKubernetesClusterRead(
	client *godo.Client,
	cluster *godo.KubernetesCluster,
	d *schema.ResourceData,
) diag.Diagnostics {
	d.Set("name", cluster.Name)
	d.Set("region", cluster.RegionSlug)
	d.Set("version", cluster.VersionSlug)
	d.Set("surge_upgrade", cluster.SurgeUpgrade)
	d.Set("ha", cluster.HA)
	d.Set("cluster_subnet", cluster.ClusterSubnet)
	d.Set("service_subnet", cluster.ServiceSubnet)
	d.Set("ipv4_address", cluster.IPv4)
	d.Set("endpoint", cluster.Endpoint)
	d.Set("tags", tag.FlattenTags(FilterTags(cluster.Tags)))
	d.Set("status", cluster.Status.State)
	d.Set("created_at", cluster.CreatedAt.UTC().String())
	d.Set("updated_at", cluster.UpdatedAt.UTC().String())
	d.Set("vpc_uuid", cluster.VPCUUID)
	d.Set("auto_upgrade", cluster.AutoUpgrade)
	d.Set("urn", cluster.URN())

	if err := d.Set("maintenance_policy", flattenMaintPolicyOpts(cluster.MaintenancePolicy)); err != nil {
		return diag.Errorf("[DEBUG] Error setting maintenance_policy - error: %#v", err)
	}

	// find the default node pool from all the pools in the cluster
	// the default node pool has a custom tag terraform:default-node-pool
	foundDefaultNodePool := false
	for i, p := range cluster.NodePools {
		for _, t := range p.Tags {
			if t == DigitaloceanKubernetesDefaultNodePoolTag {
				if foundDefaultNodePool {
					log.Printf("[WARN] Multiple node pools are marked as the default; only one node pool may have the `%s` tag", DigitaloceanKubernetesDefaultNodePoolTag)
				}

				keyPrefix := fmt.Sprintf("node_pool.%d.", i)
				if err := d.Set("node_pool", flattenNodePool(d, keyPrefix, p, cluster.Tags...)); err != nil {
					log.Printf("[DEBUG] Error setting node pool attributes: %s %#v", err, cluster.NodePools)
				}

				foundDefaultNodePool = true
			}
		}
	}
	if !foundDefaultNodePool {
		log.Printf("[WARN] No default node pool was found. The default node pool must have the `%s` tag if created with Terraform.", DigitaloceanKubernetesDefaultNodePoolTag)
	}

	// fetch cluster credentials and update the resource if the credentials are expired.
	var creds map[string]interface{}
	if d.Get("kube_config") != nil && len(d.Get("kube_config").([]interface{})) > 0 {
		creds = d.Get("kube_config").([]interface{})[0].(map[string]interface{})
	}
	var expiresAt time.Time
	if creds["expires_at"] != nil && creds["expires_at"].(string) != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, creds["expires_at"].(string))
		if err != nil {
			return diag.Errorf("Unable to parse Kubernetes credentials expiry: %s", err)
		}
	}
	if expiresAt.IsZero() || expiresAt.Before(time.Now()) {
		expireSeconds := d.Get("kubeconfig_expire_seconds").(int)
		creds, _, err := client.Kubernetes.GetCredentials(context.Background(), cluster.ID, &godo.KubernetesClusterCredentialsGetRequest{
			ExpirySeconds: &expireSeconds,
		})
		if err != nil {
			return diag.Errorf("Unable to fetch Kubernetes credentials: %s", err)
		}
		d.Set("kube_config", flattenCredentials(cluster.Name, cluster.RegionSlug, creds))
	}

	return nil
}

func resourceDigitalOceanKubernetesClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	// Figure out the changes and then call the appropriate API methods
	if d.HasChanges("name", "tags", "auto_upgrade", "surge_upgrade", "maintenance_policy", "ha") {

		opts := &godo.KubernetesClusterUpdateRequest{
			Name:         d.Get("name").(string),
			Tags:         tag.ExpandTags(d.Get("tags").(*schema.Set).List()),
			AutoUpgrade:  godo.PtrTo(d.Get("auto_upgrade").(bool)),
			SurgeUpgrade: d.Get("surge_upgrade").(bool),
			HA:           godo.PtrTo(d.Get("ha").(bool)),
		}

		if maint, ok := d.GetOk("maintenance_policy"); ok {
			maintPolicy, err := expandMaintPolicyOpts(maint.([]interface{}))
			if err != nil {
				return diag.Errorf("Error setting Kubernetes maintenance policy : %s", err)
			}
			opts.MaintenancePolicy = maintPolicy
		}

		_, resp, err := client.Kubernetes.Update(context.Background(), d.Id(), opts)
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				d.SetId("")
				return nil
			}

			return diag.Errorf("Unable to update cluster: %s", err)
		}
	}

	// Update the node pool if necessary
	if d.HasChange("node_pool") {
		old, new := d.GetChange("node_pool")
		oldPool := old.([]interface{})[0].(map[string]interface{})
		newPool := new.([]interface{})[0].(map[string]interface{})

		// If the node_count is unset, then remove it from the update map.
		if _, ok := d.GetOk("node_pool.0.node_count"); !ok {
			delete(newPool, "node_count")
		}

		// update the existing default pool
		timeout := d.Timeout(schema.TimeoutCreate)
		_, err := digitaloceanKubernetesNodePoolUpdate(client, timeout, newPool, d.Id(), oldPool["id"].(string), DigitaloceanKubernetesDefaultNodePoolTag)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("version") {
		opts := &godo.KubernetesClusterUpgradeRequest{
			VersionSlug: d.Get("version").(string),
		}

		_, err := client.Kubernetes.Upgrade(context.Background(), d.Id(), opts)
		if err != nil {
			return diag.Errorf("Unable to upgrade cluster version: %s", err)
		}
	}

	if d.HasChanges("registry_integration") {
		if d.Get("registry_integration") == true {
			err := enableRegistryIntegration(client, d.Id())
			if err != nil {
				return diag.Errorf("Error enabling registry integration: %s", err)
			}
		} else {
			err := disableRegistryIntegration(client, d.Id())
			if err != nil {
				return diag.Errorf("Error disabling registry integration: %s", err)
			}
		}
	}

	return resourceDigitalOceanKubernetesClusterRead(ctx, d, meta)
}

func resourceDigitalOceanKubernetesClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()
	destroyAssociatedResources := d.Get("destroy_all_associated_resources").(bool)

	if destroyAssociatedResources {
		log.Printf("[WARN] destroy_all_associated_resources set to true. All resources (load balancers, volumes, and volume snapshots) associated with the cluster will be destroyed.")

		list, _, err := client.Kubernetes.ListAssociatedResourcesForDeletion(ctx, d.Id())
		if err != nil {
			return diag.Errorf("Failed to list associated resources: %s", err)
		}

		log.Printf("[WARN] The following resources will be destroyed: %s", godo.Stringify(list))

		resp, err := client.Kubernetes.DeleteDangerous(ctx, d.Id())
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				d.SetId("")
				return nil
			}

			return diag.Errorf("Unable to delete cluster: %s", err)
		}
	} else {
		resp, err := client.Kubernetes.Delete(ctx, d.Id())
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				d.SetId("")
				return nil
			}

			return diag.Errorf("Unable to delete cluster: %s", err)
		}
	}

	d.SetId("")

	return nil
}

// Import a Kubernetes cluster and its node pools into the Terraform state.
//
// Note: This resource cannot make use of the pass-through importer because special handling is
// required to ensure the default node pool has the `terraform:default-node-pool` tag.
func resourceDigitalOceanKubernetesClusterImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*config.CombinedConfig).GodoClient()

	cluster, _, err := client.Kubernetes.Get(ctx, d.Id())
	if err != nil {
		return nil, err
	}

	// Check how many node pools have the required tag. The goal is ensure that one and only one node pool
	// has the tag (i.e., the default node pool).
	countOfNodePoolsWithTag := 0
	for _, nodePool := range cluster.NodePools {
		for _, tag := range nodePool.Tags {
			if tag == DigitaloceanKubernetesDefaultNodePoolTag {
				countOfNodePoolsWithTag += 1
			}
		}
	}
	if countOfNodePoolsWithTag > 1 {
		// Multiple node pools have the tag. Stop the import and notify the user they need to manually ensure
		// only one node pool has the tag.
		return nil, fmt.Errorf("Multiple node pools are tagged as the default node pool; only one node pool may have the `%s` tag", DigitaloceanKubernetesDefaultNodePoolTag)
	} else if countOfNodePoolsWithTag == 0 {
		// None of the node pools have the tag. If there is only one node pool, then it must be the default
		// node pool and thus add the tag. Adding the tag is non-destructive, and thus should be fine.
		if len(cluster.NodePools) == 1 {
			nodePool := cluster.NodePools[0]
			tags := append(nodePool.Tags, DigitaloceanKubernetesDefaultNodePoolTag)

			nodePoolUpdateRequest := &godo.KubernetesNodePoolUpdateRequest{
				Tags: tags,
			}

			log.Printf("[INFO] Adding %s tag to node pool %s in cluster %s", DigitaloceanKubernetesDefaultNodePoolTag,
				nodePool.ID, cluster.ID)

			_, _, err := client.Kubernetes.UpdateNodePool(ctx, cluster.ID, nodePool.ID, nodePoolUpdateRequest)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, MultipleNodePoolImportError
		}
	}

	if len(cluster.NodePools) > 1 {
		log.Printf("[WARN] Cluster contains multiple node pools. Importing pool tagged with '%s' as default node pool. Additional pools must be imported separately as 'digitalocean_kubernetes_node_pool' resources.",
			DigitaloceanKubernetesDefaultNodePoolTag,
		)
	}

	return []*schema.ResourceData{d}, nil
}

func enableRegistryIntegration(client *godo.Client, clusterUUID string) error {
	_, err := client.Kubernetes.AddRegistry(context.Background(), &godo.KubernetesClusterRegistryRequest{ClusterUUIDs: []string{clusterUUID}})
	return err
}

func disableRegistryIntegration(client *godo.Client, clusterUUID string) error {
	_, err := client.Kubernetes.RemoveRegistry(context.Background(), &godo.KubernetesClusterRegistryRequest{ClusterUUIDs: []string{clusterUUID}})
	return err
}

func waitForKubernetesClusterCreate(client *godo.Client, d *schema.ResourceData) (*godo.KubernetesCluster, error) {
	var (
		tickerInterval = 10 * time.Second
		timeoutSeconds = d.Timeout(schema.TimeoutCreate).Seconds()
		timeout        = int(timeoutSeconds / tickerInterval.Seconds())
		n              = 0
		ticker         = time.NewTicker(tickerInterval)
	)

	for range ticker.C {
		cluster, _, err := client.Kubernetes.Get(context.Background(), d.Id())
		if err != nil {
			ticker.Stop()
			return nil, fmt.Errorf("Error trying to read cluster state: %s", err)
		}

		if cluster.Status.State == "running" {
			ticker.Stop()
			return cluster, nil
		}

		if cluster.Status.State == "error" {
			ticker.Stop()
			return nil, fmt.Errorf(cluster.Status.Message)
		}

		if n > timeout {
			ticker.Stop()
			break
		}

		n++
	}

	return nil, fmt.Errorf("Timeout waiting to create cluster")
}

type kubernetesConfig struct {
	APIVersion     string                    `yaml:"apiVersion"`
	Kind           string                    `yaml:"kind"`
	Clusters       []kubernetesConfigCluster `yaml:"clusters"`
	Contexts       []kubernetesConfigContext `yaml:"contexts"`
	CurrentContext string                    `yaml:"current-context"`
	Users          []kubernetesConfigUser    `yaml:"users"`
}

type kubernetesConfigCluster struct {
	Cluster kubernetesConfigClusterData `yaml:"cluster"`
	Name    string                      `yaml:"name"`
}
type kubernetesConfigClusterData struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type kubernetesConfigContext struct {
	Context kubernetesConfigContextData `yaml:"context"`
	Name    string                      `yaml:"name"`
}
type kubernetesConfigContextData struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type kubernetesConfigUser struct {
	Name string                   `yaml:"name"`
	User kubernetesConfigUserData `yaml:"user"`
}

type kubernetesConfigUserData struct {
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	Token                 string `yaml:"token"`
}

func flattenCredentials(name string, region string, creds *godo.KubernetesClusterCredentials) []interface{} {
	if creds == nil {
		return nil
	}

	raw := map[string]interface{}{
		"cluster_ca_certificate": base64.StdEncoding.EncodeToString(creds.CertificateAuthorityData),
		"host":                   creds.Server,
		"token":                  creds.Token,
		"expires_at":             creds.ExpiresAt.Format(time.RFC3339),
	}

	if creds.ClientKeyData != nil {
		raw["client_key"] = string(creds.ClientKeyData)
	}

	if creds.ClientCertificateData != nil {
		raw["client_certificate"] = string(creds.ClientCertificateData)
	}

	kubeconfigYAML, err := RenderKubeconfig(name, region, creds)
	if err != nil {
		log.Printf("[DEBUG] error marshalling config: %s", err)
		return nil
	}
	raw["raw_config"] = string(kubeconfigYAML)

	return []interface{}{raw}
}

func RenderKubeconfig(name string, region string, creds *godo.KubernetesClusterCredentials) ([]byte, error) {
	clusterName := fmt.Sprintf("do-%s-%s", region, name)
	userName := fmt.Sprintf("do-%s-%s-admin", region, name)
	config := kubernetesConfig{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []kubernetesConfigCluster{{
			Name: clusterName,
			Cluster: kubernetesConfigClusterData{
				CertificateAuthorityData: base64.StdEncoding.EncodeToString(creds.CertificateAuthorityData),
				Server:                   creds.Server,
			},
		}},
		Contexts: []kubernetesConfigContext{{
			Context: kubernetesConfigContextData{
				Cluster: clusterName,
				User:    userName,
			},
			Name: clusterName,
		}},
		CurrentContext: clusterName,
		Users: []kubernetesConfigUser{{
			Name: userName,
			User: kubernetesConfigUserData{
				Token: creds.Token,
			},
		}},
	}
	if creds.ClientKeyData != nil {
		config.Users[0].User.ClientKeyData = base64.StdEncoding.EncodeToString(creds.ClientKeyData)
	}
	if creds.ClientCertificateData != nil {
		config.Users[0].User.ClientCertificateData = base64.StdEncoding.EncodeToString(creds.ClientCertificateData)
	}
	return yaml.Marshal(config)
}
