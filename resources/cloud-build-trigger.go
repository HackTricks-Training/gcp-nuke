package resources

import (
	"context"
	"fmt"
	"path"
	"time"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	cloudbuildpb "cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeCloudBuildTrigger = "CloudBuildTrigger"

type CloudBuildTrigger struct {
	name         string
	creationDate string
	region       string
}

func init() {
	register(ResourceTypeCloudBuildTrigger, GetCloudBuildTriggerClient, ListCloudBuildTriggers)
}

func GetCloudBuildTriggerClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeCloudBuildTrigger); ok {
		return client, nil
	}

	client, err := cloudbuild.NewClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud build trigger client: %v", err)
	}
	project.AddClient(ResourceTypeCloudBuildTrigger, client)
	return client, nil
}

func ListCloudBuildTriggers(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	cloudBuildClient := client.(*cloudbuild.Client)

	resources := make([]Resource, 0)
	var req *cloudbuildpb.ListBuildTriggersRequest
	for _, location := range project.Locations {
		// global isn't a valid location for cloud build
		// if strings.ToLower(location) == "global" {
		// 	continue
		// }
		req = &cloudbuildpb.ListBuildTriggersRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}

		it := cloudBuildClient.ListBuildTriggers(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list cloud build triggers: %v", err)
			}
			resources = append(resources, &CloudBuildTrigger{
				name:         path.Base(resp.GetName()),
				creationDate: resp.GetCreateTime().AsTime().Format(time.RFC3339),
				region:       location,
			})
		}
	}
	return resources, nil
}

func (x *CloudBuildTrigger) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	cloudRunClient := client.(*cloudbuild.Client)

	req := &cloudbuildpb.DeleteBuildTriggerRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/triggers/%s", project.Name, x.region, x.name),
	}

	err = cloudRunClient.DeleteBuildTrigger(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *CloudBuildTrigger) GetOperationError(ctx context.Context) error {
	return nil
}

func (x *CloudBuildTrigger) String() string {
	return x.name
}

func (x *CloudBuildTrigger) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Region", x.region)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
