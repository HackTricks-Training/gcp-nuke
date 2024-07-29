package resources

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeBucketObject = "BucketObject"

type BucketObject struct {
	name         string
	generation   int64
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

		query := &storage.Query{
			Versions: bucket.VersioningEnabled,
		}
		itObj := storageClient.Bucket(bucket.Name).Objects(project.GetContext(), query)
		for {
			objAttrs, err := itObj.Next()
			if err == iterator.Done {
				break
			}
			fmt.Printf("Object %s | %s | %d", objAttrs.Name, objAttrs.Deleted, objAttrs.Generation)

			if err != nil {
				return nil, fmt.Errorf("failed to list bucket objects for %s: %v", bucket.Name, err)
			}
			if !bucket.VersioningEnabled && !objAttrs.Deleted.IsZero() {
				continue
			}
			resources = append(resources, &BucketObject{
				name:         objAttrs.Name,
				generation:   objAttrs.Generation,
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

	err := bucketObject.Generation(b.generation).Delete(project.GetContext())
	if err != nil {
		return err
	}

	return nil
}

func (b *BucketObject) GetOperationError(_ context.Context) error {
	return nil
}

func (b *BucketObject) String() string {
	return b.name
}

func (b *BucketObject) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", b.name)
	properties.Set("Generation", b.generation)
	properties.Set("CreationDate", b.creationDate)
	properties.Set("Bucket", b.bucket)

	for labelKey, label := range b.bucketLabels {
		properties.SetTagWithPrefix("bucket", labelKey, label)
	}

	return properties
}
