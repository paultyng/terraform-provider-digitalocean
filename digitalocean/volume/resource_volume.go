package volume

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/tag"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceDigitalOceanVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDigitalOceanVolumeCreate,
		ReadContext:   resourceDigitalOceanVolumeRead,
		UpdateContext: resourceDigitalOceanVolumeUpdate,
		DeleteContext: resourceDigitalOceanVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					// DO API V2 region slug is always lowercase
					return strings.ToLower(val.(string))
				},
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"urn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the uniform resource name for the volume.",
			},
			"size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // Update-ability Coming Soon ™
				ValidateFunc: validation.NoZeroValues,
			},

			"snapshot_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"initial_filesystem_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ext4",
					"xfs",
				}, false),
			},

			"initial_filesystem_label": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"droplet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Computed: true,
			},

			"filesystem_type": {
				Type:     schema.TypeString,
				Optional: true, // Backward compatibility for existing resources.
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ext4",
					"xfs",
				}, false),
				ConflictsWith: []string{"initial_filesystem_type"},
				Deprecated:    "This fields functionality has been replaced by `initial_filesystem_type`. The property will still remain as a computed attribute representing the current volumes filesystem type.",
			},

			"filesystem_label": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tag.TagsSchema(),
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {

			// if the new size of the volume is smaller than the old one return an error since
			// only expanding the volume is allowed
			oldSize, newSize := diff.GetChange("size")
			if newSize.(int) < oldSize.(int) {
				return fmt.Errorf("volumes `size` can only be expanded and not shrunk")
			}

			return nil
		},
	}
}

func resourceDigitalOceanVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	opts := &godo.VolumeCreateRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tags:        tag.ExpandTags(d.Get("tags").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("region"); ok {
		opts.Region = strings.ToLower(v.(string))
	}
	if v, ok := d.GetOk("size"); ok {
		opts.SizeGigaBytes = int64(v.(int))
	}
	if v, ok := d.GetOk("snapshot_id"); ok {
		opts.SnapshotID = v.(string)
	}
	if v, ok := d.GetOk("initial_filesystem_type"); ok {
		opts.FilesystemType = v.(string)
	} else if v, ok := d.GetOk("filesystem_type"); ok {
		// backward compatibility
		opts.FilesystemType = v.(string)
	}
	if v, ok := d.GetOk("initial_filesystem_label"); ok {
		opts.FilesystemLabel = v.(string)
	}

	log.Printf("[DEBUG] Volume create configuration: %#v", opts)
	volume, _, err := client.Storage.CreateVolume(context.Background(), opts)
	if err != nil {
		return diag.Errorf("Error creating Volume: %s", err)
	}

	d.SetId(volume.ID)
	log.Printf("[INFO] Volume name: %s", volume.Name)

	return resourceDigitalOceanVolumeRead(ctx, d, meta)
}

func resourceDigitalOceanVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	id := d.Id()
	region := strings.ToLower(d.Get("region").(string))

	if d.HasChange("size") {
		size := d.Get("size").(int)

		log.Printf("[DEBUG] Volume resize configuration: %v", size)
		action, _, err := client.StorageActions.Resize(context.Background(), id, size, region)
		if err != nil {
			return diag.Errorf("Error resizing volume (%s): %s", id, err)
		}

		log.Printf("[DEBUG] Volume resize action id: %d", action.ID)
		if err = util.WaitForAction(client, action); err != nil {
			return diag.Errorf(
				"Error waiting for resize volume (%s) to finish: %s", id, err)
		}
	}

	if d.HasChange("tags") {
		err := tag.SetTags(client, d, godo.VolumeResourceType)
		if err != nil {
			return diag.Errorf("Error updating tags: %s", err)
		}
	}

	return resourceDigitalOceanVolumeRead(ctx, d, meta)
}

func resourceDigitalOceanVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	volume, resp, err := client.Storage.GetVolume(context.Background(), d.Id())
	if err != nil {
		// If the volume is somehow already destroyed, mark as
		// successfully gone
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving volume: %s", err)
	}

	d.Set("name", volume.Name)
	d.Set("region", volume.Region.Slug)
	d.Set("size", int(volume.SizeGigaBytes))
	d.Set("urn", volume.URN())
	d.Set("tags", tag.FlattenTags(volume.Tags))

	if v := volume.Description; v != "" {
		d.Set("description", v)
	}
	if v := volume.FilesystemType; v != "" {
		d.Set("filesystem_type", v)
		d.Set("initial_filesystem_type", v)
	}
	if v := volume.FilesystemLabel; v != "" {
		d.Set("filesystem_label", v)
		d.Set("initial_filesystem_label", v)
	}

	if err = d.Set("droplet_ids", flattenDigitalOceanVolumeDropletIds(volume.DropletIDs)); err != nil {
		return diag.Errorf("[DEBUG] Error setting droplet_ids: %#v", err)
	}

	return nil
}

func resourceDigitalOceanVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	log.Printf("[INFO] Deleting volume: %s", d.Id())
	_, err := client.Storage.DeleteVolume(context.Background(), d.Id())
	if err != nil {
		return diag.Errorf("Error deleting volume: %s", err)
	}

	d.SetId("")
	return nil
}

func flattenDigitalOceanVolumeDropletIds(droplets []int) *schema.Set {
	flattenedDroplets := schema.NewSet(schema.HashInt, []interface{}{})
	for _, v := range droplets {
		flattenedDroplets.Add(v)
	}

	return flattenedDroplets
}
