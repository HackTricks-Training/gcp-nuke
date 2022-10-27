package resources

import (
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
)

const ResourceTypeBucketObject = "BucketObject"

type BucketObject struct {
	name         string
	bucket       string
	creationDate string
	bucketLabels map[string]string
}

func init() {
	register(ResourceTypeBucketObject, GetGCSClient, ListBucketObjects)
}

func ListBucketObjects(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	storageClient := client.(*storage.Client)

	resources := make([]Resource, 0)

	it := storageClient.Buckets(project.GetContext(), project.Name)
	for {
		bucket, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list buckets: %v", err)
		}
		itObj := storageClient.Bucket(bucket.Name).Objects(project.GetContext(), nil)
		for {
			objAttrs, err := itObj.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list bucket objects for %s: %v", bucket.Name, err)
			}
			if !objAttrs.Deleted.IsZero() {
				continue
			}
			resources = append(resources, &BucketObject{
				name:         objAttrs.Name,
				bucket:       objAttrs.Bucket,
				creationDate: objAttrs.Created.Format(time.RFC3339),
				bucketLabels: bucket.Labels,
			})
		}
	}
	return resources, nil
}

func (b *BucketObject) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	storageClient := client.(*storage.Client)

	bucketObject := storageClient.Bucket(b.bucket).Object(b.name)
	err := bucketObject.Delete(project.GetContext())
	if err != nil {
		return err
	}

	return nil
}

func (b *BucketObject) String() string {
	return b.name
}

func (b *BucketObject) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", b.name)
	properties.Set("CreationDate", b.creationDate)
	properties.Set("Bucket", b.bucket)

	for labelKey, label := range b.bucketLabels {
		properties.SetTagWithPrefix("bucket", labelKey, label)
	}

	return properties
}
