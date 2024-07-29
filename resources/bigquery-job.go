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

const ResourceTypeBigqueryJob = "BigqueryJob"

type BigqueryJob struct {
	id       string
	location string
	project  string
}

func init() {
	register(ResourceTypeBigqueryJob, GetBigqueryJobClient, ListBigqueryJob)
}

func GetBigqueryJobClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeBigqueryJob); ok {
		return client, nil
	}

	client, err := bigquery.NewClient(project.GetContext(), project.Name, project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery datasets client: %v", err)
	}
	project.AddClient(ResourceTypeBigqueryJob, client)
	return client, nil
}

func ListBigqueryJob(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	bigqueryClient := client.(*bigquery.Client)

	resources := make([]Resource, 0)

	it := bigqueryClient.Jobs(project.GetContext())
	for {
		job, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list bigquery jobs: %v", err)
		}
		resources = append(resources, &BigqueryJob{
			id:       path.Base(job.ID()),
			location: job.Location(),
			project:  project.Name,
		})
	}

	return resources, nil
}

func (x *BigqueryJob) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	bigqueryClient := client.(*bigquery.Client)

	job, err := bigqueryClient.JobFromIDLocation(project.GetContext(), x.id, x.location)
	if err != nil {
		return err
	}
	err = job.Delete(project.GetContext())
	if err != nil {
		return err
	}
	return nil
}

func (x *BigqueryJob) GetOperationError(_ context.Context) error {
	return nil
}

func (x *BigqueryJob) String() string {
	return x.id
}

func (x *BigqueryJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("ID", x.id)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
