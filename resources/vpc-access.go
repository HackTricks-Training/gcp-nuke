package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	vpcaccess "cloud.google.com/go/vpcaccess/apiv1"
	"cloud.google.com/go/vpcaccess/apiv1/vpcaccesspb"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
)

const ResourceTypeVpcAccess = "VPCAccess"

type VpcAccess struct {
	name      string
	region    string
	network   string
	operation *vpcaccess.DeleteConnectorOperation
}

func init() {
	register(ResourceTypeVpcAccess, GetVpcAccessClient, ListVpcAccess)
}

func GetVpcAccessClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeVpcAccess); ok {
		return client, nil
	}

	client, err := vpcaccess.NewClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud run client: %v", err)
	}
	project.AddClient(ResourceTypeVpcAccess, client)
	return client, nil
}

func ListVpcAccess(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	vpcAccessClient := client.(*vpcaccess.Client)

	resources := make([]Resource, 0)
	var req *vpcaccesspb.ListConnectorsRequest
	for _, location := range project.Locations {
		// global isn't a valid location for VPC access
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &vpcaccesspb.ListConnectorsRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}

		it := vpcAccessClient.ListConnectors(project.GetContext(), req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list subnetworks: %v", err)
			}
			resources = append(resources, &VpcAccess{
				name:    path.Base(resp.GetName()),
				region:  location,
				network: resp.GetNetwork(),
			})
		}
	}
	return resources, nil
}

func (x *VpcAccess) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	vpcAccessClient := client.(*vpcaccess.Client)

	req := &vpcaccesspb.DeleteConnectorRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/connectors/%s", project.Name, x.region, x.name),
	}

	x.operation, err = vpcAccessClient.DeleteConnector(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *VpcAccess) GetOperationError(ctx context.Context) (err error) {
	if x.operation != nil {
		err = x.operation.Poll(ctx)
	}
	return err
}

func (x *VpcAccess) String() string {
	return x.name
}

func (x *VpcAccess) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Region", x.region)
	properties.Set("Network", x.network)

	return properties
}
