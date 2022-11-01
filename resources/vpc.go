package resources

import (
	"context"
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
	operation    *compute.Operation
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

func (x *Vpc) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	networksClient := client.(*compute.NetworksClient)

	req := &computepb.DeleteNetworkRequest{
		Network: x.name,
		Project: project.Name,
	}

	x.operation, err = networksClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *Vpc) GetOperationError(ctx context.Context) error {
	return getComputeOperationError(ctx, x.operation)
}

func getComputeOperationError(ctx context.Context, op *compute.Operation) error {
	if op != nil {
		if err := op.Poll(ctx); err == nil {
			if op.Done() {
				if op.Proto().GetHttpErrorStatusCode() != http.StatusOK {
					return fmt.Errorf("Delete error on '%s': %s", op.Proto().GetTargetLink(), op.Proto().GetHttpErrorMessage())
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func (x *Vpc) String() string {
	return x.name
}

func (x *Vpc) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
