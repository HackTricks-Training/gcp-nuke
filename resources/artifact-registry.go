package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	artifactregistrypb "cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"google.golang.org/api/iterator"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

const ResourceTypeArtifactRegistry = "ArtifactRegistry"

type ArtifactRegistry struct {
	name     string
	location string
	project  string
}

func init() {
	register(ResourceTypeArtifactRegistry, GetArtifactRegistryClient, ListArtifactRegistry)
}

func GetArtifactRegistryClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeArtifactRegistry); ok {
		return client, nil
	}

	client, err := artifactregistry.NewClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact registry client: %v", err)
	}
	project.AddClient(ResourceTypeArtifactRegistry, client)
	return client, nil
}

func ListArtifactRegistry(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	artifactRegistryClient := client.(*artifactregistry.Client)

	resources := make([]Resource, 0)

	for _, location := range project.Locations {
		// global isn't a valid location for Artifact Registries
		if strings.ToLower(location) == "global" {
			continue
		}
		req := &artifactregistrypb.ListRepositoriesRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}
		it := artifactRegistryClient.ListRepositories(project.GetContext(), req)
		for {
			repository, err := it.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				return nil, fmt.Errorf("failed to list artifact repositories: %v", err)
			}
			resources = append(resources, &ArtifactRegistry{
				name:     path.Base(repository.Name),
				location: location,
				project:  project.Name,
			})
		}

	}
	return resources, nil
}

func (x *ArtifactRegistry) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	schedulerClient := client.(*artifactregistry.Client)

	req := &artifactregistrypb.DeleteRepositoryRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/repositories/%s", x.project, x.location, x.name),
	}
	_, err := schedulerClient.DeleteRepository(project.GetContext(), req)
	if err != nil {
		return err
	}
	return nil
}

func (x *ArtifactRegistry) GetOperationError(_ context.Context) error {
	return nil
}

func (x *ArtifactRegistry) String() string {
	return x.name
}

func (x *ArtifactRegistry) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Location", x.location)
	properties.Set("Project", x.project)

	return properties
}
