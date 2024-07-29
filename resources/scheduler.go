package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	scheduler "cloud.google.com/go/scheduler/apiv1"
	"cloud.google.com/go/scheduler/apiv1/schedulerpb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeSchedulerJob = "SchedulerJob"

type SchedulerJob struct {
	name     string
	location string
	project  string
}

func init() {
	register(ResourceTypeSchedulerJob, GetSchedulerClient, ListSchedulerJobs)
}

func GetSchedulerClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeSchedulerJob); ok {
		return client, nil
	}

	client, err := scheduler.NewCloudSchedulerClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler client: %v", err)
	}
	project.AddClient(ResourceTypeSchedulerJob, client)
	return client, nil
}

func ListSchedulerJobs(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	schedulerClient := client.(*scheduler.CloudSchedulerClient)

	resources := make([]Resource, 0)

	for _, location := range project.Locations {
		// global isn't a valid location for Scheduler
		if strings.ToLower(location) == "global" {
			continue
		}
		req := &schedulerpb.ListJobsRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}
		it := schedulerClient.ListJobs(project.GetContext(), req)
		for {
			job, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list scheduler jobs: %v", err)
			}
			resources = append(resources, &SchedulerJob{
				name:     path.Base(job.Name),
				location: location,
				project:  project.Name,
			})
		}

	}
	return resources, nil
}

func (x *SchedulerJob) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	schedulerClient := client.(*scheduler.CloudSchedulerClient)

	req := &schedulerpb.DeleteJobRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", x.project, x.location, x.name),
	}
	err := schedulerClient.DeleteJob(project.GetContext(), req)
	if err != nil {
		return err
	}
	return nil
}

func (x *SchedulerJob) GetOperationError(_ context.Context) error {
	return nil
}

func (x *SchedulerJob) String() string {
	return x.name
}

func (x *SchedulerJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
