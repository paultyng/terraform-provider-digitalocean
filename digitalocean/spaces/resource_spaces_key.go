package spaces

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"
	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/util"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDigitalOceanSpacesKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDigitalOceanSpacesKeyCreate,
		ReadContext:   resourceDigitalOceanSpacesKeyRead,
		UpdateContext: resourceDigitalOceanSpacesKeyUpdate,
		DeleteContext: resourceDigitalOceanSpacesKeyDelete,

		Schema: spacesKeySchema(),
	}
}

func resourceDigitalOceanSpacesKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	name := d.Get("name").(string)
	rawGrants := d.Get("grant").([]interface{})

	req := &godo.SpacesKeyCreateRequest{
		Name:   name,
		Grants: parseRawGrants(rawGrants),
	}

	var key *godo.SpacesKey
	var err error
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		log.Printf("[DEBUG] Creating new Spaces key")
		key, _, err = client.SpacesKeys.Create(ctx, req)
		if err != nil {
			if util.IsDigitalOceanError(err, http.StatusInternalServerError, "") || util.IsDigitalOceanError(err, http.StatusPreconditionFailed, "") {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(
				fmt.Errorf("[WARN] Error creating Spaces key: %s, retrying: %s", name, err))
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("Error creating Spaces key: %s", err)
	}

	log.Println("Spaces Key created")
	d.SetId(key.AccessKey)
	d.Set("name", key.Name)
	d.Set("access_key", key.AccessKey)
	d.Set("secret_key", key.SecretKey)
	d.Set("grant", flattenGrants(key.Grants))
	d.Set("created_at", key.CreatedAt)
	return nil
}

func resourceDigitalOceanSpacesKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
			if k.AccessKey == d.Id() {
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

	d.Set("name", key.Name)
	d.Set("access_key", key.AccessKey)
	d.Set("grant", flattenGrants(key.Grants))
	d.Set("created_at", key.CreatedAt)
	return nil
}

func resourceDigitalOceanSpacesKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	name := d.Get("name").(string)
	rawGrants := d.Get("grant").([]interface{})

	req := &godo.SpacesKeyUpdateRequest{
		Name:   name,
		Grants: parseRawGrants(rawGrants),
	}

	var key *godo.SpacesKey
	var err error
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		log.Printf("[DEBUG] Updating Spaces key: %s", name)
		key, _, err = client.SpacesKeys.Update(ctx, d.Id(), req)
		if err != nil {
			if util.IsDigitalOceanError(err, http.StatusInternalServerError, "") || util.IsDigitalOceanError(err, http.StatusPreconditionFailed, "") {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(
				fmt.Errorf("[WARN] Error updating Spaces key: %s, retrying: %s", name, err))
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("Error updating Spaces key: %s", err)
	}

	log.Println("Spaces Key updated")
	d.Set("name", key.Name)
	d.Set("access_key", key.AccessKey)
	d.Set("grant", flattenGrants(key.Grants))
	d.Set("created_at", key.CreatedAt)
	return nil
}

func resourceDigitalOceanSpacesKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*config.CombinedConfig).GodoClient()

	err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		log.Printf("[DEBUG] Deleting Spaces key: %s", d.Id())
		_, err := client.SpacesKeys.Delete(ctx, d.Id())
		if err != nil {
			if util.IsDigitalOceanError(err, http.StatusInternalServerError, "") || util.IsDigitalOceanError(err, http.StatusPreconditionFailed, "") {
				return retry.NonRetryableError(err)
			}
			return retry.RetryableError(
				fmt.Errorf("[WARN] Error deleting Spaces key: %s, retrying: %s", d.Id(), err))
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("Error deleting Spaces key: %s", err)
	}

	log.Println("Spaces Key deleted")
	d.SetId("")
	return nil
}

func parseRawGrants(rawGrants []interface{}) []*godo.Grant {
	grants := make([]*godo.Grant, 0, len(rawGrants))
	for _, rawGrant := range rawGrants {
		grant := rawGrant.(map[string]interface{})
		g := &godo.Grant{}
		for k, v := range grant {
			if k == "bucket" {
				g.Bucket = v.(string)
			} else {
				switch v.(string) {
				case "read":
					g.Permission = godo.SpacesKeyRead
				case "readwrite":
					g.Permission = godo.SpacesKeyReadWrite
				case "fullaccess":
					g.Permission = godo.SpacesKeyFullAccess
				}
			}
		}
		grants = append(grants, g)
	}
	return grants
}

func flattenGrants(grants []*godo.Grant) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(grants))
	for _, grant := range grants {
		g := make(map[string]interface{})
		g["bucket"] = grant.Bucket
		switch grant.Permission {
		case godo.SpacesKeyRead:
			g["permission"] = "read"
		case godo.SpacesKeyReadWrite:
			g["permission"] = "readwrite"
		case godo.SpacesKeyFullAccess:
			g["permission"] = "fullaccess"
		}
		results = append(results, g)
	}
	return results
}
