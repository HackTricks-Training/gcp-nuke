package resources

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeCloudRunJob = "CloudRunJob"

type CloudRunJob struct {
	name         string
	creationDate string
	creator      string
	region       string
	labels       map[string]string
	operation    *run.DeleteJobOperation
}

func init() {
	register(ResourceTypeCloudRunJob, GetCloudRunJobClient, ListCloudRunJobs)
}

func GetCloudRunJobClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeCloudRunJob); ok {
		return client, nil
	}

	client, err := run.NewJobsClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run job client: %v", err)
	}
	project.AddClient(ResourceTypeCloudRunJob, client)
	return client, nil
}

func ListCloudRunJobs(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	cloudRunClient := client.(*run.JobsClient)

	resources := make([]Resource, 0)
	var req *runpb.ListJobsRequest
	for _, location := range project.Locations {
		// global isn't a valid location for cloud run
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &runpb.ListJobsRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}

		it := cloudRunClient.ListJobs(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list cloud run jobs: %v", err)
			}
			resources = append(resources, &CloudRunJob{
				name:         path.Base(resp.GetName()),
				creationDate: resp.GetCreateTime().AsTime().Format(time.RFC3339),
				creator:      resp.GetCreator(),
				region:       location,
				labels:       resp.GetLabels(),
			})
		}
	}
	return resources, nil
}

func (x *CloudRunJob) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	cloudRunClient := client.(*run.JobsClient)

	req := &runpb.DeleteJobRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project.Name, x.region, x.name),
	}

	x.operation, err = cloudRunClient.DeleteJob(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *CloudRunJob) GetOperationError(ctx context.Context) error {
	return getRunJobOperationError(ctx, x.operation)
}

func getRunJobOperationError(ctx context.Context, op *run.DeleteJobOperation) (err error) {
	if op != nil {
		_, err = op.Poll(ctx)
	}
	return err
}

func (x *CloudRunJob) String() string {
	return x.name
}

func (x *CloudRunJob) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Region", x.region)
	properties.Set("CreationDate", x.creationDate)
	properties.Set("Creator", x.creator)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
