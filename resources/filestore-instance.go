package resources

import (
	"context"
	"fmt"
	"path"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeFilestoreInstance = "FilestoreInstance"

type FilestoreInstance struct {
	name     string
	location string
	project  string
}

func init() {
	register(ResourceTypeFilestoreInstance, GetFilestoreInstanceClient, ListFilestoreInstance)
}

func GetFilestoreInstanceClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeFilestoreInstance); ok {
		return client, nil
	}

	client, err := filestore.NewCloudFilestoreManagerClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create filestore client: %v", err)
	}
	project.AddClient(ResourceTypeFilestoreInstance, client)
	return client, nil
}

func ListFilestoreInstance(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	filestoreClient := client.(*filestore.CloudFilestoreManagerClient)

	resources := make([]Resource, 0)

	req := &filestorepb.ListInstancesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, "-"),
	}
	it := filestoreClient.ListInstances(project.GetContext(), req)
	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list filestore instances: %v", err)
		}
		_, loc := path.Split(path.Dir(path.Dir(instance.Name)))

		resources = append(resources, &FilestoreInstance{
			name:     path.Base(instance.Name),
			location: loc,
			project:  project.Name,
		})
	}

	return resources, nil
}

func (x *FilestoreInstance) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	filestoreClient := client.(*filestore.CloudFilestoreManagerClient)

	req := &filestorepb.DeleteInstanceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/instances/%s", x.project, x.location, x.name),
	}
	_, err := filestoreClient.DeleteInstance(project.GetContext(), req)
	if err != nil {
		return err
	}
	return nil
}

func (x *FilestoreInstance) GetOperationError(_ context.Context) error {
	return nil
}

func (x *FilestoreInstance) String() string {
	return x.name
}

func (x *FilestoreInstance) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
