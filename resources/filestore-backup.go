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

const ResourceTypeFilestoreBackup = "FilestoreBackup"

type FilestoreBackup struct {
	name     string
	location string
	project  string
}

func init() {
	register(ResourceTypeFilestoreBackup, GetFilestoreBackupClient, ListFilestoreBackup)
}

func GetFilestoreBackupClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeFilestoreBackup); ok {
		return client, nil
	}

	client, err := filestore.NewCloudFilestoreManagerClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create filestore client: %v", err)
	}
	project.AddClient(ResourceTypeFilestoreBackup, client)
	return client, nil
}

func ListFilestoreBackup(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	filestoreClient := client.(*filestore.CloudFilestoreManagerClient)

	resources := make([]Resource, 0)

	req := &filestorepb.ListBackupsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, "-"),
	}
	it := filestoreClient.ListBackups(project.GetContext(), req)
	for {
		backup, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list filestore backups: %v", err)
		}
		_, loc := path.Split(path.Dir(path.Dir(backup.Name)))

		resources = append(resources, &FilestoreBackup{
			name:     path.Base(backup.Name),
			location: loc,
			project:  project.Name,
		})
	}
	return resources, nil
}

func (x *FilestoreBackup) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	filestoreClient := client.(*filestore.CloudFilestoreManagerClient)

	req := &filestorepb.DeleteBackupRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/backups/%s", x.project, x.location, x.name),
	}
	_, err := filestoreClient.DeleteBackup(project.GetContext(), req)
	if err != nil {
		return err
	}
	return nil
}

func (x *FilestoreBackup) GetOperationError(_ context.Context) error {
	return nil
}

func (x *FilestoreBackup) String() string {
	return x.name
}

func (x *FilestoreBackup) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
