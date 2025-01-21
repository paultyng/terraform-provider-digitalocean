package spaces

import (
	"fmt"

	"github.com/digitalocean/terraform-provider-digitalocean/digitalocean/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	// SpacesRegions is a list of DigitalOcean regions that support Spaces.
	SpacesRegions = []string{"ams3", "blr1", "fra1", "lon1", "nyc3", "sfo2", "sfo3", "sgp1", "syd1", "tor1"}
)

type bucketMetadataStruct struct {
	name   string
	region string
}

func spacesBucketSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Bucket name",
		},
		"urn": {
			Type:        schema.TypeString,
			Description: "the uniform resource name for the bucket",
		},
		"region": {
			Type:        schema.TypeString,
			Description: "Bucket region",
		},
		"bucket_domain_name": {
			Type:        schema.TypeString,
			Description: "The FQDN of the bucket",
		},
		"endpoint": {
			Type:        schema.TypeString,
			Description: "The FQDN of the bucket without the bucket name",
		},
	}
}

func spacesKeySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A name for the key. This is used to identify the key in the DigitalOcean control panel.",
		},
		"grant": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of grants to apply to the key. Can be left empty to apply no grants.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"bucket": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The name of the bucket to grant the key access to.",
					},
					"permission": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The permission to grant the key. Valid values are `read`, `readwrite`, or `fullaccess`.",
					},
				},
			},
		},
		"access_key": {
			Type:        schema.TypeString,
			Description: "The access key for the Spaces key",
			Computed:    true,
		},
		"secret_key": {
			Type:        schema.TypeString,
			Description: "The secret key for the Spaces key",
			Computed:    true,
			Sensitive:   true,
		},
		"created_at": {
			Type:        schema.TypeString,
			Description: "The date and time the key was created",
			Computed:    true,
		},
	}
}

func getSpacesBucketsInRegion(meta interface{}, region string) ([]*s3.Bucket, error) {
	client, err := meta.(*config.CombinedConfig).SpacesClient(region)
	if err != nil {
		return nil, err
	}

	svc := s3.New(client)

	input := s3.ListBucketsInput{}
	output, err := svc.ListBuckets(&input)
	if err != nil {
		return nil, err
	}

	return output.Buckets, nil
}

func getDigitalOceanBuckets(meta interface{}, extra map[string]interface{}) ([]interface{}, error) {
	// The DigitalOcean API does not currently return what regions have Spaces available. Thus, this
	// function hard-codes the regions in which Spaces operates.
	var buckets []interface{}

	for _, region := range SpacesRegions {
		bucketsInRegion, err := getSpacesBucketsInRegion(meta, region)
		if err != nil {
			return nil, err
		}

		for _, bucketInRegion := range bucketsInRegion {
			metadata := &bucketMetadataStruct{
				name:   *bucketInRegion.Name,
				region: region,
			}
			buckets = append(buckets, metadata)
		}
	}

	return buckets, nil
}

func flattenSpacesBucket(rawBucketMetadata, meta interface{}, extra map[string]interface{}) (map[string]interface{}, error) {
	bucketMetadata := rawBucketMetadata.(*bucketMetadataStruct)

	name := bucketMetadata.name
	region := bucketMetadata.region

	flattenedBucket := map[string]interface{}{}
	flattenedBucket["name"] = name
	flattenedBucket["region"] = region
	flattenedBucket["bucket_domain_name"] = BucketDomainName(name, region)
	flattenedBucket["urn"] = fmt.Sprintf("do:space:%s", name)
	flattenedBucket["endpoint"] = BucketEndpoint(region)

	return flattenedBucket, nil
}

// CompareSpacesBucketPolicy determines the equivalence of two S3 policies
// using github.com/hashicorp/awspolicyequivalence.PoliciesAreEquivalent
func CompareSpacesBucketPolicy(policy1, policy2 string) bool {
	equivalent, err := awspolicy.PoliciesAreEquivalent(policy1, policy2)
	if err != nil {
		return false
	}
	return equivalent
}

// spacesBucketForceDelete deletes all objects in a Spaces bucket.
func spacesBucketForceDelete(svc *s3.S3, bucket string) error {
	listParams := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	}
	resp, err := svc.ListObjectVersions(listParams)
	if err != nil {
		return err
	}

	objectsToDelete := make([]*s3.ObjectIdentifier, 0)
	if len(resp.DeleteMarkers) != 0 {
		for _, v := range resp.DeleteMarkers {
			objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
				Key:       v.Key,
				VersionId: v.VersionId,
			})
		}
	}

	if len(resp.Versions) != 0 {
		for _, v := range resp.Versions {
			objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
				Key:       v.Key,
				VersionId: v.VersionId,
			})
		}
	}

	params := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: objectsToDelete,
		},
	}

	_, err = svc.DeleteObjects(params)
	if err != nil {
		return err
	}

	return nil
}
