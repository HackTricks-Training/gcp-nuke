package gcputil

// This wrapper exists because the IAM client from Google doesn't have a
// Close() method and therefore doesn't conform to the GCPClient interface that
// is used by gcp-nuke.

import (
	"context"
	"fmt"

	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type IAMClient struct {
	client *iam.Service
}

func (c *IAMClient) Close() error {
	return nil
}

func NewIAMClient(ctx context.Context, opts ...option.ClientOption) (c *IAMClient, err error) {
	c = &IAMClient{}
	c.client, err = iam.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("Error creating IAMClient: %v", err)
	}
	return
}

func (c IAMClient) GetIAMService() *iam.Service {
	return c.client
}
