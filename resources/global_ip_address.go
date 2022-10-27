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

const ResourceTypeGlobalIPAddress = "GlobalIPAddress"

type GlobalIPAddress struct {
	name         string
	network      string
	creationDate string
}

func init() {
	register(ResourceTypeGlobalIPAddress, GetGlobalIPAddressClient, ListGlobalIPAddresss)
}

func GetGlobalIPAddressClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeGlobalIPAddress); ok {
		return client, nil
	}

	client, err := compute.NewGlobalAddressesRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IP Global Addresses client: %v", err)
	}
	project.AddClient(ResourceTypeGlobalIPAddress, client)
	return client, nil
}

func ListGlobalIPAddresss(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	addresssesClient := client.(*compute.GlobalAddressesClient)

	resources := make([]Resource, 0)
	req := &computepb.ListGlobalAddressesRequest{
		Project: project.Name,
	}
	it := addresssesClient.List(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list global IP Addresses: %v", err)
		}

		resources = append(resources, &GlobalIPAddress{
			name:         *resp.Name,
			network:      path.Base(UnPtrString(resp.Network, "")),
			creationDate: *resp.CreationTimestamp,
		})
	}
	return resources, nil
}

func (x *GlobalIPAddress) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	addressesClient := client.(*compute.GlobalAddressesClient)

	req := &computepb.DeleteGlobalAddressRequest{
		Project: project.Name,
		Address: x.name,
	}

	_, err := addressesClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *GlobalIPAddress) String() string {
	return x.name
}

func (x *GlobalIPAddress) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Network", x.network)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
