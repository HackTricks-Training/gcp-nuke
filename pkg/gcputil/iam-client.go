package gcputil

// // This wrapper exists because the IAM client from Google doesn't have a
// // Close() method and therefore doesn't conform to the GCPClient interface that
// // is used by gcp-nuke.

// import (
// 	"context"
// 	"fmt"

// 	"google.golang.org/api/option"
// 	iam "google.golang.org/api/iam/v1"
// )

// type IAMClient struct {
// 	client *iam.Service
// }

// func (c *IAMClient) Close() error {
// 	return nil
// }

// func NewIAMClient(ctx context.Context, opts ...option.ClientOption) (c *IAMClient, err error) {
// 	c = &IAMClient{}
// 	c.client, err = iam.NewService(ctx, opts...)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error creating IAMClient: %v", err)
// 	}
// 	return
// }

// func (c *IAMClient) List(ctx context.Context, project string) (*iam.service, error) {
// 	return c.client.Projects.ServiceAccounts.
// 	Instances.List(project).Context(ctx).Do()
// }

// func (c *IAMClient) Remove(ctx context.Context, project, dbInstance string) (*cloudsql.Operation, error) {
// 	return c.client.Instances.Delete(project, dbInstance).Context(ctx).Do()
// }

// func (c *IAMClient) GetOperation(ctx context.Context, project, operation string) (*cloudsql.Operation, error) {
// 	return c.client.Operations.Get(project, operation).Context(ctx).Do()
// }
