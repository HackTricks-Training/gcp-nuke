package resources

import (
	"fmt"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeRouter = "Router"

type Router struct {
	name         string
	network      string
	creationDate string
	region       string
}

func init() {
	register(ResourceTypeRouter, GetRouterClient, ListRouters)
}

func GetRouterClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeRouter); ok {
		return client, nil
	}

	client, err := compute.NewRoutersRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create routers client: %v", err)
	}
	project.AddClient(ResourceTypeRouter, client)
	return client, nil
}

func ListRouters(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	routersClient := client.(*compute.RoutersClient)

	resources := make([]Resource, 0)
	var req *computepb.ListRoutersRequest
	for _, location := range project.Locations {
		// global isn't a valid location for routers
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &computepb.ListRoutersRequest{
			Project: project.Name,
			Region:  location,
		}

		it := routersClient.List(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list routers: %v", err)
			}
			resources = append(resources, &Router{
				name:         *resp.Name,
				network:      path.Base(*resp.Network),
				creationDate: *resp.CreationTimestamp,
				region:       path.Base(*resp.Region),
			})
		}
	}
	return resources, nil
}

func (r *Router) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	routersClient := client.(*compute.RoutersClient)

	req := &computepb.DeleteRouterRequest{
		Project: project.Name,
		Region:  r.region,
		Router:  r.name,
	}

	_, err := routersClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (r *Router) String() string {
	return r.name
}

func (r *Router) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", r.name)
	properties.Set("Network", r.network)
	properties.Set("CreationDate", r.creationDate)

	return properties
}
