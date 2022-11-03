package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	cloudsql "google.golang.org/api/sqladmin/v1beta4"
)

const ResourceTypeCloudSQL = "CloudSQL"

type CloudSQL struct {
	name         string
	labels       map[string]string
	creationDate string
	location     string
}

func init() {
	register(ResourceTypeCloudSQL, GetCloudSQLClient, ListCloudSQLs)
}

func GetCloudSQLClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeCloudSQL); ok {
		return client, nil
	}

	client, err := cloudsql.NewService(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql client: %v", err)
	}
	project.AddClient(ResourceTypeCloudSQL, client)
	return client, nil
}

func ListCloudSQLs(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	cloudSqlClient := client.(*cloudsql.Client)

	resources := make([]Resource, 0)

	it := storageClient.CloudSQLs(project.GetContext(), project.Name)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list buckets: %v", err)
		}
		resources = append(resources, &CloudSQL{
			name:         resp.Name,
			creationDate: resp.Created.Format(time.RFC3339),
			labels:       resp.Labels,
			location:     resp.Location,
		})
	}
	return resources, nil
}

func (b *CloudSQL) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	storageClient := client.(*storage.Client)

	bucket := storageClient.CloudSQL(b.name)
	err := bucket.Delete(project.GetContext())
	if err != nil {
		return err
	}

	return nil
}

func (b *CloudSQL) GetOperationError(_ context.Context) error {
	return nil
}

func (b *CloudSQL) String() string {
	return b.name
}

func (b *CloudSQL) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", b.name)
	properties.Set("CreationDate", b.creationDate)
	properties.Set("Location", b.location)

	for labelKey, label := range b.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
