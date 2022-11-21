package resources

import (
	"context"
	"fmt"
	"path"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeZonalNetworkEndpointGroup = "ZonalNetworkEndpointGroup"

var boolFalse = false

type ZonalNetworkEndpointGroup struct {
	name         string
	zone         string
	negType      string
	creationDate string
	operation    *compute.Operation
}

func init() {
	register(ResourceTypeZonalNetworkEndpointGroup, GetZonalNetworkEndpointGroupClient, ListZonalNetworkEndpointGroups)
}

func GetZonalNetworkEndpointGroupClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeZonalNetworkEndpointGroup); ok {
		return client, nil
	}

	client, err := compute.NewNetworkEndpointGroupsRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create network endpoint group client: %v", err)
	}
	project.AddClient(ResourceTypeZonalNetworkEndpointGroup, client)
	return client, nil
}

func ListZonalNetworkEndpointGroups(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	negClient := client.(*compute.NetworkEndpointGroupsClient)

	resources := make([]Resource, 0)

	req := &computepb.AggregatedListNetworkEndpointGroupsRequest{
		Project:          project.Name,
		IncludeAllScopes: &boolFalse,
	}

	it := negClient.AggregatedList(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list network endpiont groups: %v", err)
		}

		if ZoneInRegionList(resp.Key, project.Locations) {
			for _, neg := range resp.Value.GetNetworkEndpointGroups() {
				resources = append(resources, &ZonalNetworkEndpointGroup{
					name:         neg.GetName(),
					zone:         path.Base(neg.GetZone()),
					negType:      neg.GetNetworkEndpointType(),
					creationDate: neg.GetCreationTimestamp(),
				})
			}
		}
	}
	return resources, nil
}

func (x *ZonalNetworkEndpointGroup) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	negClient := client.(*compute.NetworkEndpointGroupsClient)

	req := &computepb.DeleteNetworkEndpointGroupRequest{
		Project:              project.Name,
		Zone:                 x.zone,
		NetworkEndpointGroup: x.name,
	}

	x.operation, err = negClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *ZonalNetworkEndpointGroup) GetOperationError(ctx context.Context) error {
	return getComputeOperationError(ctx, x.operation)
}

func (x *ZonalNetworkEndpointGroup) String() string {
	return x.name
}

func (x *ZonalNetworkEndpointGroup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Zone", x.zone)
	properties.Set("EndpointType", x.negType)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
