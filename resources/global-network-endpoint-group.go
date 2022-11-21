package resources

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeGlobalNetworkEndpointGroup = "GlobalNetworkEndpointGroup"

type GlobalNetworkEndpointGroup struct {
	name         string
	negType      string
	creationDate string
	operation    *compute.Operation
}

func init() {
	register(ResourceTypeGlobalNetworkEndpointGroup, GetGlobalNetworkEndpointGroupClient, ListGlobalNetworkEndpointGroups)
}

func GetGlobalNetworkEndpointGroupClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeGlobalNetworkEndpointGroup); ok {
		return client, nil
	}

	client, err := compute.NewGlobalNetworkEndpointGroupsRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create global network endpoint group client: %v", err)
	}
	project.AddClient(ResourceTypeGlobalNetworkEndpointGroup, client)
	return client, nil
}

func ListGlobalNetworkEndpointGroups(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	negClient := client.(*compute.GlobalNetworkEndpointGroupsClient)

	resources := make([]Resource, 0)

	req := &computepb.ListGlobalNetworkEndpointGroupsRequest{
		Project: project.Name,
	}

	it := negClient.List(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list global network endpiont groups: %v", err)
		}

		resources = append(resources, &GlobalNetworkEndpointGroup{
			name:         resp.GetName(),
			negType:      resp.GetNetworkEndpointType(),
			creationDate: resp.GetCreationTimestamp(),
		})
	}
	return resources, nil
}

func (x *GlobalNetworkEndpointGroup) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	negClient := client.(*compute.GlobalNetworkEndpointGroupsClient)

	req := &computepb.DeleteGlobalNetworkEndpointGroupRequest{
		Project:              project.Name,
		NetworkEndpointGroup: x.name,
	}

	x.operation, err = negClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *GlobalNetworkEndpointGroup) GetOperationError(ctx context.Context) error {
	return getComputeOperationError(ctx, x.operation)
}

func (x *GlobalNetworkEndpointGroup) String() string {
	return x.name
}

func (x *GlobalNetworkEndpointGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("EndpointType", x.negType)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
