package resources

import (
	"fmt"
	"net/http"
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
	operation    *compute.Operation
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

func (x *Subnet) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	subnetworksClient := client.(*compute.SubnetworksClient)

	req := &computepb.DeleteSubnetworkRequest{
		Project:    project.Name,
		Region:     x.region,
		Subnetwork: x.name,
	}

	x.operation, err = subnetworksClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *Subnet) GetOperationError() error {
	if x.operation != nil && x.operation.Done() {
		if x.operation.Proto().GetHttpErrorStatusCode() != http.StatusOK {
			return fmt.Errorf("IPAddress Delete error: %s", *x.operation.Proto().HttpErrorMessage)
		}
	}
	return nil
}

func (x *Subnet) String() string {
	return x.name
}

func (x *Subnet) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Network", x.network)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
