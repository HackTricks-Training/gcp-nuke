package resources

import (
	"fmt"
	"path"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeRoute = "Route"

type Route struct {
	name         string
	network      string
	creationDate string
}

func init() {
	register(ResourceTypeRoute, GetRouteClient, ListRoutes)
}

func GetRouteClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeRoute); ok {
		return client, nil
	}

	client, err := compute.NewRoutesRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create routes client: %v", err)
	}
	project.AddClient(ResourceTypeRoute, client)
	return client, nil
}

func ListRoutes(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	firewallsClient := client.(*compute.RoutesClient)

	resources := make([]Resource, 0)
	req := &computepb.ListRoutesRequest{
		Project: project.Name,
		Filter:  &noDefaultNetworkFilter,
	}

	it := firewallsClient.List(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list routes: %v", err)
		}
		resources = append(resources, &Route{
			name:         *resp.Name,
			network:      *resp.Network,
			creationDate: *resp.CreationTimestamp,
		})
	}
	return resources, nil
}

func (v *Route) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	firewallsClient := client.(*compute.RoutesClient)

	req := &computepb.DeleteRouteRequest{
		Route:   v.name,
		Project: project.Name,
	}

	_, err := firewallsClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (v *Route) String() string {
	return v.name
}

func (v *Route) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", v.name)
	properties.Set("Network", path.Base(v.network))
	properties.Set("CreationDate", v.creationDate)

	return properties
}
