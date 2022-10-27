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

const ResourceTypeSubnet = "Subnet"

type Subnet struct {
	name         string
	network      string
	creationDate string
	region       string
}

func init() {
	register(ResourceTypeSubnet, GetSubnetworkClient, ListSubnets)
}

func GetSubnetworkClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeSubnet); ok {
		return client, nil
	}

	client, err := compute.NewSubnetworksRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnetwork client: %v", err)
	}
	project.AddClient(ResourceTypeSubnet, client)
	return client, nil
}

func ListSubnets(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	subnetworksClient := client.(*compute.SubnetworksClient)

	resources := make([]Resource, 0)
	var req *computepb.ListSubnetworksRequest
	for _, location := range project.Locations {
		// global isn't a valid location for subnets
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &computepb.ListSubnetworksRequest{
			Project: project.Name,
			Filter:  &noDefaultNetworkFilter,
			Region:  location,
		}

		it := subnetworksClient.List(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list subnetworks: %v", err)
			}
			resources = append(resources, &Subnet{
				name:         *resp.Name,
				network:      path.Base(*resp.Network),
				creationDate: *resp.CreationTimestamp,
				region:       path.Base(*resp.Region),
			})
		}
	}
	return resources, nil
}

func (s *Subnet) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	subnetworksClient := client.(*compute.SubnetworksClient)

	req := &computepb.DeleteSubnetworkRequest{
		Project:    project.Name,
		Region:     s.region,
		Subnetwork: s.name,
	}

	_, err := subnetworksClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (s *Subnet) String() string {
	return s.name
}

func (s *Subnet) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", s.name)
	properties.Set("Network", s.network)
	properties.Set("CreationDate", s.creationDate)

	return properties
}
