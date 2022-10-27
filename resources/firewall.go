package resources

import (
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

const ResourceTypeFirewall = "Firewall"

type Firewall struct {
	name         string
	network      string
	creationDate string
}

func init() {
	register(ResourceTypeFirewall, GetFirewallClient, ListFirewalls)
}

func GetFirewallClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeFirewall); ok {
		return client, nil
	}

	client, err := compute.NewFirewallsRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall client: %v", err)
	}
	project.AddClient(ResourceTypeFirewall, client)
	return client, nil
}

func ListFirewalls(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	firewallsClient := client.(*compute.FirewallsClient)

	resources := make([]Resource, 0)
	req := &computepb.ListFirewallsRequest{
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
			return nil, fmt.Errorf("failed to list firewalls: %v", err)
		}
		resources = append(resources, &Firewall{
			name:         *resp.Name,
			network:      *resp.Network,
			creationDate: *resp.CreationTimestamp,
		})
	}
	return resources, nil
}

func (v *Firewall) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	firewallsClient := client.(*compute.FirewallsClient)

	req := &computepb.DeleteFirewallRequest{
		Firewall: v.name,
		Project:  project.Name,
	}

	_, err := firewallsClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (v *Firewall) String() string {
	return v.name
}

func (v *Firewall) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", v.name)
	properties.Set("CreationDate", v.creationDate)

	return properties
}
