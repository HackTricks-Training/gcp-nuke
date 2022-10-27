package resources

import (
	"fmt"
	"net/http"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeVPC = "VPC"

type Vpc struct {
	name         string
	creationDate string
}

var noDefaultNetworkFilter = "name != default"

func init() {
	register(ResourceTypeVPC, GetNetworkClient, ListVpcs)
}

func GetNetworkClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeVPC); ok {
		return client, nil
	}

	client, err := compute.NewNetworksRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create network client: %v", err)
	}
	project.AddClient(ResourceTypeVPC, client)
	return client, nil
}

func ListVpcs(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	networksClient := client.(*compute.NetworksClient)

	resources := make([]Resource, 0)
	req := &computepb.ListNetworksRequest{
		Project: project.Name,
		Filter:  &noDefaultNetworkFilter,
	}

	it := networksClient.List(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list networks: %v", err)
		}
		resources = append(resources, &Vpc{
			name:         UnPtrString(resp.Name, ""),
			creationDate: *resp.CreationTimestamp,
		})
	}
	return resources, nil
}

func (v *Vpc) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	networksClient := client.(*compute.NetworksClient)

	req := &computepb.DeleteNetworkRequest{
		Network: v.name,
		Project: project.Name,
	}

	op, err := networksClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	err = op.Wait(project.GetContext())
	if err != nil {
		return err
	}
	if *op.Proto().HttpErrorStatusCode != http.StatusOK {
		return fmt.Errorf("VPC Delete error: %s", *op.Proto().HttpErrorMessage)
	}
	return nil
}

func (v *Vpc) String() string {
	return v.name
}

func (v *Vpc) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", v.name)
	properties.Set("CreationDate", v.creationDate)

	return properties
}
