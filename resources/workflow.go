package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	workflows "cloud.google.com/go/workflows/apiv1"
	workflowspb "cloud.google.com/go/workflows/apiv1/workflowspb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeWorkflow = "Workflow"

type Workflow struct {
	name     string
	location string
	project  string
}

func init() {
	register(ResourceTypeWorkflow, GetWorkflowsClient, ListWorkflows)
}

func GetWorkflowsClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeWorkflow); ok {
		return client, nil
	}

	client, err := workflows.NewClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflows client: %v", err)
	}
	project.AddClient(ResourceTypeWorkflow, client)
	return client, nil
}

func ListWorkflows(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	workflowsClient := client.(*workflows.Client)

	resources := make([]Resource, 0)

	for _, location := range project.Locations {
		// global isn't a valid location for Workflows
		if strings.ToLower(location) == "global" {
			continue
		}
		req := &workflowspb.ListWorkflowsRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}
		it := workflowsClient.ListWorkflows(project.GetContext(), req)
		for {
			workflows, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list workflows: %v", err)
			}
			resources = append(resources, &Workflow{
				name:     path.Base(workflows.Name),
				location: location,
				project:  project.Name,
			})
		}

	}
	return resources, nil
}

func (x *Workflow) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	workflowsClient := client.(*workflows.Client)

	req := &workflowspb.DeleteWorkflowRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/workflows/%s", x.project, x.location, x.name),
	}
	_, err := workflowsClient.DeleteWorkflow(project.GetContext(), req)
	if err != nil {
		return err
	}
	return nil
}

func (x *Workflow) GetOperationError(_ context.Context) error {
	return nil
}

func (x *Workflow) String() string {
	return x.name
}

func (x *Workflow) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
