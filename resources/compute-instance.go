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

const ResourceTypeComputeInstance = "ComputeInstance"

type ComputeInstance struct {
	name         string
	zone         string
	creationDate string
	status       string
	machineType  string
	labels       map[string]string
	operation    *compute.Operation
}

func init() {
	register(ResourceTypeComputeInstance, GetComputeInstanceClient, ListComputeInstances)
}

func GetComputeInstanceClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeComputeInstance); ok {
		return client, nil
	}

	client, err := compute.NewInstancesRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %v", err)
	}
	project.AddClient(ResourceTypeComputeInstance, client)
	return client, nil
}

func ListComputeInstances(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	instancesClient := client.(*compute.InstancesClient)

	resources := make([]Resource, 0)
	req := &computepb.AggregatedListInstancesRequest{
		Project: project.Name,
	}

	it := instancesClient.AggregatedList(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %v", err)
		}
		if ZoneInRegionList(resp.Key, project.Locations) {
			for _, instance := range resp.Value.GetInstances() {
				resources = append(resources, &ComputeInstance{
					name:         instance.GetName(),
					zone:         path.Base(instance.GetZone()),
					status:       instance.GetStatus(),
					machineType:  path.Base(instance.GetMachineType()),
					creationDate: instance.GetCreationTimestamp(),
					labels:       instance.GetLabels(),
				})
			}
		}
	}
	return resources, nil
}

func (x *ComputeInstance) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	instancesClient := client.(*compute.InstancesClient)

	req := &computepb.DeleteInstanceRequest{
		Project:  project.Name,
		Zone:     x.zone,
		Instance: x.name,
	}

	x.operation, err = instancesClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *ComputeInstance) GetOperationError(ctx context.Context) error {
	return getComputeOperationError(ctx, x.operation)
}

func (x *ComputeInstance) String() string {
	return x.name
}

func (x *ComputeInstance) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Zone", x.zone)
	properties.Set("Status", x.status)
	properties.Set("MachineType", x.machineType)
	properties.Set("CreationDate", x.creationDate)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
