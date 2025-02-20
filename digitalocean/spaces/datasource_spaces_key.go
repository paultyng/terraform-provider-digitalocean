package spaces

import (
	"context"
	"log"

	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceDigitalOceanSpacesKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSpacesKeyRead,

		Schema: spacesKeySchema(),
	}
}

func dataSourceSpacesKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	opts := &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	var key *godo.SpacesKey

	for {
		keys, resp, err := client.SpacesKeys.List(ctx, opts)
		if err != nil {
			return diag.Errorf("Error reading Spaces key: %s", err)
		}

		for _, k := range keys {
			if k.Name == d.Get("name").(string) {
				key = k
				break
			}
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return diag.Errorf("Error reading Spaces key: %s", err)
		}

		opts.Page = page + 1
	}

	if key == nil {
		log.Printf("[WARN] Key not found: %s", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(key.AccessKey)
	d.Set("name", key.Name)
	d.Set("access_key", key.AccessKey)
	d.Set("grant", flattenGrants(key.Grants))
	d.Set("created_at", key.CreatedAt)
	return nil
}
