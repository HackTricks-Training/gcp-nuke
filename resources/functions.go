package resources

import (
	"context"
	"fmt"
	"path"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeFunction = "Function"

type Function struct {
	name     string
	location string
	project  string
}

func init() {
	register(ResourceTypeFunction, GetFunctionClient, ListFunction)
}

func GetFunctionClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeFunction); ok {
		return client, nil
	}

	client, err := functions.NewFunctionClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create functions client: %v", err)
	}
	project.AddClient(ResourceTypeFunction, client)
	return client, nil
}

func ListFunction(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	functionsClient := client.(*functions.FunctionClient)

	resources := make([]Resource, 0)

	req := &functionspb.ListFunctionsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, "-"),
	}
	it := functionsClient.ListFunctions(project.GetContext(), req)
	for {
		function, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list functions: %v", err)
		}
		_, loc := path.Split(path.Dir(path.Dir(function.Name)))

		resources = append(resources, &Function{
			name:     path.Base(function.Name),
			location: loc,
			project:  project.Name,
		})
	}

	return resources, nil
}

func (x *Function) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	functionsClient := client.(*functions.FunctionClient)

	req := &functionspb.DeleteFunctionRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/functions/%s", x.project, x.location, x.name),
	}
	_, err := functionsClient.DeleteFunction(project.GetContext(), req)
	if err != nil {
		return err
	}
	return nil
}

func (x *Function) GetOperationError(_ context.Context) error {
	return nil
}

func (x *Function) String() string {
	return x.name
}

func (x *Function) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
