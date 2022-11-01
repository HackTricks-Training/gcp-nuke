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

const ResourceTypeIPAddress = "IPAddress"

type IPAddress struct {
	name         string
	network      string
	creationDate string
	region       string
	operation    *compute.Operation
}

func init() {
	register(ResourceTypeIPAddress, GetIPAddressClient, ListIPAddresss)
}

func GetIPAddressClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeIPAddress); ok {
		return client, nil
	}

	client, err := compute.NewAddressesRESTClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IP Addresses client: %v", err)
	}
	project.AddClient(ResourceTypeIPAddress, client)
	return client, nil
}

func ListIPAddresss(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	addresssesClient := client.(*compute.AddressesClient)

	resources := make([]Resource, 0)
	var req *computepb.ListAddressesRequest
	for _, location := range project.Locations {
		// global isn't a valid location for IPAddress
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &computepb.ListAddressesRequest{
			Project: project.Name,
			Region:  location,
		}
		it := addresssesClient.List(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list IP Addresses: %v", err)
			}

			resources = append(resources, &IPAddress{
				name:         *resp.Name,
				network:      path.Base(UnPtrString(resp.Network, "")),
				creationDate: *resp.CreationTimestamp,
				region:       path.Base(UnPtrString(resp.Region, "")),
			})
		}
	}
	return resources, nil
}

func (x *IPAddress) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	addressesClient := client.(*compute.AddressesClient)

	req := &computepb.DeleteAddressRequest{
		Project: project.Name,
		Address: x.name,
		Region:  x.region,
	}

	x.operation, err = addressesClient.Delete(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *IPAddress) GetOperationError() error {
	if x.operation != nil && x.operation.Done() {
		if x.operation.Proto().GetHttpErrorStatusCode() != http.StatusOK {
			return fmt.Errorf("IPAddress Delete error: %s", *x.operation.Proto().HttpErrorMessage)
		}
	}
	return nil
}

func (x *IPAddress) String() string {
	return x.name
}

func (x *IPAddress) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Network", x.network)
	properties.Set("CreationDate", x.creationDate)

	return properties
}
