package resources

import (
	"context"
	"fmt"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	cloudsql "google.golang.org/api/sqladmin/v1beta4"
)

const ResourceTypeCloudSQL = "CloudSQL"

type CloudSQL struct {
	name         string
	labels       map[string]string
	creationDate string
	location     string
	dbVersion    string
	project      string
	operation    *cloudsql.Operation
	sqlClient    *gcputil.CloudSQLClient
}

func init() {
	register(ResourceTypeCloudSQL, GetCloudSQLClient, ListCloudSQLs)
}

func GetCloudSQLClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeCloudSQL); ok {
		return client, nil
	}

	client, err := gcputil.NewSQLClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql client: %v", err)
	}
	project.AddClient(ResourceTypeCloudSQL, client)
	return client, nil
}

func ListCloudSQLs(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	cloudSQLClient := client.(*gcputil.CloudSQLClient)

	resources := make([]Resource, 0)

	resp, err := cloudSQLClient.List(project.GetContext(), project.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to list SQL instances: %v", err)
	}
	for _, instance := range resp.Items {
		resources = append(resources, &CloudSQL{
			name:         instance.Name,
			creationDate: instance.CreateTime,
			labels:       instance.Settings.UserLabels,
			location:     instance.Region,
			dbVersion:    instance.DatabaseVersion,
			project:      instance.Project,
		})
	}
	return resources, nil
}

func (x *CloudSQL) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	x.sqlClient = client.(*gcputil.CloudSQLClient)

	x.operation, err = x.sqlClient.Remove(project.GetContext(), project.Name, x.name)
	if err != nil {
		return err
	}

	return nil
}

func (x *CloudSQL) GetOperationError(ctx context.Context) error {
	if x.operation != nil {
		if op, err := x.sqlClient.GetOperation(ctx, x.project, x.operation.Name); err == nil {
			if op.Status == "DONE" {
				if op.Error != nil {
					return fmt.Errorf("Delete error on '%s': %s", op.TargetLink, op.Error.Errors[0].Message)
				}
			}
		} else {
			return err
		}
	}
	return nil
}

func (x *CloudSQL) String() string {
	return x.name
}

func (x *CloudSQL) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("CreationDate", x.creationDate)
	properties.Set("Location", x.location)
	properties.Set("DBVersion", x.dbVersion)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
