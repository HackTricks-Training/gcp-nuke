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

const ResourceTypeComputeDisk = "ComputeDisk"

type ComputeDisk struct {
	name         string
	zone         string
	creationDate string
	status       string
	labels       map[string]string
	operation    *compute.Operation
}

func init() {
	register(ResourceTypeComputeDisk, GetComputeDiskClient, ListComputeDisks)
}

func GetComputeDiskClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeComputeDisk); ok {
		return client, nil
	}

	client, err := compute.NewDisksRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute disks client: %v", err)
	}
	project.AddClient(ResourceTypeComputeDisk, client)
	return client, nil
}

func ListComputeDisks(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	disksClient := client.(*compute.DisksClient)

	resources := make([]Resource, 0)
	req := &computepb.AggregatedListDisksRequest{
		Project: project.Name,
	}

	it := disksClient.AggregatedList(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list compute disks: %v", err)
		}
		if ZoneInRegionList(resp.Key, project.Locations) {
			for _, instance := range resp.Value.GetDisks() {
				resources = append(resources, &ComputeDisk{
					name:         instance.GetName(),
					zone:         path.Base(instance.GetZone()),
					status:       instance.GetStatus(),
					creationDate: instance.GetCreationTimestamp(),
					labels:       instance.GetLabels(),
				})
			}
		}
	}
	return resources, nil
}

func (x *ComputeDisk) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	disksClient := client.(*compute.DisksClient)

	req := &computepb.DeleteDiskRequest{
		Project: project.Name,
		Zone:    x.zone,
		Disk:    x.name,
	}

	x.operation, err = disksClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *ComputeDisk) GetOperationError(ctx context.Context) error {
	return getComputeOperationError(ctx, x.operation)
}

func (x *ComputeDisk) String() string {
	return x.name
}

func (x *ComputeDisk) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Zone", x.zone)
	properties.Set("Status", x.status)
	properties.Set("CreationDate", x.creationDate)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
