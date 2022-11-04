package gcputil

// This wrapper exists because the Cloud SQL client from Google doesn't have a
// Close() method and therefore doesn't conform to the GCPClient interface that
// is used by gcp-nuke.

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	cloudsql "google.golang.org/api/sqladmin/v1beta4"
)

type CloudSQLClient struct {
	client *cloudsql.Service
}

func (c *CloudSQLClient) Close() error {
	return nil
}

func NewSQLClient(ctx context.Context, opts ...option.ClientOption) (c *CloudSQLClient, err error) {
	c = &CloudSQLClient{}
	c.client, err = cloudsql.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("Error creating CloudSQLClient: %v", err)
	}
	return
}

func (c *CloudSQLClient) List(ctx context.Context, project string) (*cloudsql.InstancesListResponse, error) {
	return c.client.Instances.List(project).Context(ctx).Do()
}

func (c *CloudSQLClient) Remove(ctx context.Context, project, dbInstance string) (*cloudsql.Operation, error) {
	return c.client.Instances.Delete(project, dbInstance).Context(ctx).Do()
}

func (c *CloudSQLClient) GetOperation(ctx context.Context, project, operation string) (*cloudsql.Operation, error) {
	return c.client.Operations.Get(project, operation).Context(ctx).Do()
}
