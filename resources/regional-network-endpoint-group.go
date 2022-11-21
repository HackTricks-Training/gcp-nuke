package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeRegionalNetworkEndpointGroup = "RegionalNetworkEndpointGroup"

type RegionalNetworkEndpointGroup struct {
	name         string
	negType      string
	region       string
	creationDate string
	operation    *compute.Operation
}

func init() {
	register(ResourceTypeRegionalNetworkEndpointGroup, GetRegionalNetworkEndpointGroupClient, ListRegionalNetworkEndpointGroups)
}

func GetRegionalNetworkEndpointGroupClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeRegionalNetworkEndpointGroup); ok {
		return client, nil
	}

	client, err := compute.NewRegionNetworkEndpointGroupsRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create regional network endpoint group client: %v", err)
	}
	project.AddClient(ResourceTypeRegionalNetworkEndpointGroup, client)
	return client, nil
}

func ListRegionalNetworkEndpointGroups(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	negClient := client.(*compute.RegionNetworkEndpointGroupsClient)

	resources := make([]Resource, 0)
	var req *computepb.ListRegionNetworkEndpointGroupsRequest
	for _, location := range project.Locations {
		// global isn't a valid location for regional NEGs
		if strings.ToLower(location) == "global" {
			continue
		}

		req = &computepb.ListRegionNetworkEndpointGroupsRequest{
			Project: project.Name,
			Region:  location,
		}

		it := negClient.List(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list Regional network endpiont groups: %v", err)
			}

			resources = append(resources, &RegionalNetworkEndpointGroup{
				name:         resp.GetName(),
				negType:      resp.GetNetworkEndpointType(),
				region:       path.Base(resp.GetRegion()),
				creationDate: resp.GetCreationTimestamp(),
			})
		}
	}
	return resources, nil
}

func (x *RegionalNetworkEndpointGroup) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	negClient := client.(*compute.RegionNetworkEndpointGroupsClient)

	req := &computepb.DeleteRegionNetworkEndpointGroupRequest{
		Project:              project.Name,
		Region:               x.region,
		NetworkEndpointGroup: x.name,
	}

	x.operation, err = negClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *RegionalNetworkEndpointGroup) GetOperationError(ctx context.Context) error {
	return getComputeOperationError(ctx, x.operation)
}

func (x *RegionalNetworkEndpointGroup) String() string {
	return x.name
}

func (x *RegionalNetworkEndpointGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Region", x.region)
	properties.Set("EndpointType", x.negType)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
