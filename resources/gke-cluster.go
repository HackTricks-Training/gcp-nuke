package resources

import (
	"fmt"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeGKECluster = "GKECluster"

type GKECluster struct {
	name         string
	labels       map[string]string
	creationDate string
	location     string
}

func init() {
	register(ResourceTypeGKECluster, GetGKEClient, ListGKEClusters)
}

func GetGKEClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeGKECluster); ok {
		return client, nil
	}

	client, err := container.NewClusterManagerClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create container client: %v", err)
	}
	project.AddClient(ResourceTypeGKECluster, client)
	return client, nil
}

func ListGKEClusters(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	gkeClient := client.(*container.ClusterManagerClient)

	resources := make([]Resource, 0)

	var req *containerpb.ListClustersRequest
	for _, location := range project.Locations {
		// global isn't a valid location for GKE
		if strings.ToLower(location) == "global" {
			continue
		}
		req = &containerpb.ListClustersRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}

		resp, err := gkeClient.ListClusters(project.GetContext(), req)
		if err != nil {
			return nil, fmt.Errorf("failed to list GKE clusters: %v", err)
		}

		for _, cluster := range resp.Clusters {
			resources = append(resources, &GKECluster{
				name:         cluster.Name,
				creationDate: cluster.CreateTime,
				labels:       cluster.ResourceLabels,
				location:     cluster.Location,
			})
		}
	}
	return resources, nil
}

func (c *GKECluster) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	gkeClient := client.(*container.ClusterManagerClient)

	delReq := &containerpb.DeleteClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project.Name, c.location, c.name),
	}
	_, err := gkeClient.DeleteCluster(project.GetContext(), delReq)
	if err != nil {
		return err
	}

	return nil
}

func (c *GKECluster) String() string {
	return c.name
}

func (c *GKECluster) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", c.name)
	properties.Set("CreationDate", c.creationDate)
	properties.Set("Location", c.location)

	for labelKey, label := range c.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
