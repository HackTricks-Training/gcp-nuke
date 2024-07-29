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

const ResourceTypeCloudRunService = "CloudRunService"

type CloudRunService struct {
	name         string
	creationDate string
	creator      string
	region       string
	labels       map[string]string
	operation    *run.DeleteServiceOperation
}

func init() {
	register(ResourceTypeCloudRunService, GetCloudRunServiceClient, ListCloudRunServices)
}

func GetCloudRunServiceClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeCloudRunService); ok {
		return client, nil
	}

	client, err := run.NewServicesClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run service client: %v", err)
	}
	project.AddClient(ResourceTypeCloudRunService, client)
	return client, nil
}

func ListCloudRunServices(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	cloudRunClient := client.(*run.ServicesClient)

	resources := make([]Resource, 0)
	var req *runpb.ListServicesRequest
	for _, location := range project.Locations {
		// global isn't a valid location for cloud run
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &runpb.ListServicesRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}

		it := cloudRunClient.ListServices(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list cloud run services: %v", err)
			}
			resources = append(resources, &CloudRunService{
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

func (x *CloudRunService) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	cloudRunClient := client.(*run.ServicesClient)

	req := &runpb.DeleteServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", project.Name, x.region, x.name),
	}

	x.operation, err = cloudRunClient.DeleteService(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *CloudRunService) GetOperationError(ctx context.Context) error {
	return getRunServiceOperationError(ctx, x.operation)
}

func getRunServiceOperationError(ctx context.Context, op *run.DeleteServiceOperation) (err error) {
	if op != nil {
		_, err = op.Poll(ctx)
	}
	return err
}

func (x *CloudRunService) String() string {
	return x.name
}

func (x *CloudRunService) Properties() types.Properties {
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
