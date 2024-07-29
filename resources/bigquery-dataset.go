package resources

import (
	"context"
	"fmt"
	"path"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeBigqueryDataset = "BigqueryDataset"

type BigqueryDataset struct {
	id      string
	project string
}

func init() {
	register(ResourceTypeBigqueryDataset, GetBigqueryDatasetClient, ListBigqueryDataset)
}

func GetBigqueryDatasetClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeBigqueryDataset); ok {
		return client, nil
	}

	client, err := bigquery.NewClient(project.GetContext(), project.Name, project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery datasets client: %v", err)
	}
	project.AddClient(ResourceTypeBigqueryDataset, client)
	return client, nil
}

func ListBigqueryDataset(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	bigqueryClient := client.(*bigquery.Client)

	resources := make([]Resource, 0)

	it := bigqueryClient.Datasets(project.GetContext())
	for {
		dataset, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list bigquery datasets: %v", err)
		}
		resources = append(resources, &BigqueryDataset{
			id:      path.Base(dataset.DatasetID),
			project: project.Name,
		})
	}

	return resources, nil
}

func (x *BigqueryDataset) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	bigqueryClient := client.(*bigquery.Client)

	err := bigqueryClient.Dataset(x.id).DeleteWithContents(project.GetContext())
	if err != nil {
		return err
	}
	return nil
}

func (x *BigqueryDataset) GetOperationError(_ context.Context) error {
	return nil
}

func (x *BigqueryDataset) String() string {
	return x.id
}

func (x *BigqueryDataset) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", x.id)
	properties.Set("Project", x.project)

	return properties
}
