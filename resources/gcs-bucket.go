package resources

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
)

const ResourceTypeBucket = "Bucket"

type Bucket struct {
	name         string
	labels       map[string]string
	creationDate string
	location     string
}

func init() {
	register(ResourceTypeBucket, GetGCSClient, ListBuckets)
}

func GetGCSClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeBucket); ok {
		return client, nil
	}

	client, err := storage.NewClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}
	project.AddClient(ResourceTypeBucket, client)
	return client, nil
}

func ListBuckets(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	storageClient := client.(*storage.Client)

	resources := make([]Resource, 0)

	it := storageClient.Buckets(project.GetContext(), project.Name)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list buckets: %v", err)
		}
		resources = append(resources, &Bucket{
			name:         resp.Name,
			creationDate: resp.Created.Format(time.RFC3339),
			labels:       resp.Labels,
			location:     resp.Location,
		})
	}
	return resources, nil
}

func (b *Bucket) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	storageClient := client.(*storage.Client)

	bucket := storageClient.Bucket(b.name)
	err := bucket.Delete(project.GetContext())
	if err != nil {
		return err
	}

	return nil
}

func (b *Bucket) GetOperationError(_ context.Context) error {
	return nil
}

func (b *Bucket) String() string {
	return b.name
}

func (b *Bucket) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", b.name)
	properties.Set("CreationDate", b.creationDate)
	properties.Set("Location", b.location)

	for labelKey, label := range b.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
