package resources

import (
	"fmt"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const ResourceTypeSecret = "Secret"

type Secret struct {
	name         string
	labels       map[string]string
	creationDate string
}

func init() {
	register(ResourceTypeSecret, GetSecretClient, ListSecret)
}

func GetSecretClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeSecret); ok {
		return client, nil
	}

	client, err := secretmanager.NewClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	project.AddClient(ResourceTypeSecret, client)
	return client, nil
}

func ListSecret(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	secretClient := client.(*secretmanager.Client)

	resources := make([]Resource, 0)
	req := &secretmanagerpb.ListSecretsRequest{
		Parent: fmt.Sprintf("projects/%s", project.Name),
	}

	it := secretClient.ListSecrets(project.GetContext(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %v", err)
		}
		resources = append(resources, &Secret{
			name:         resp.Name,
			creationDate: resp.CreateTime.AsTime().Format(time.RFC3339),
			labels:       resp.GetLabels(),
		})
	}
	return resources, nil
}

func (e *Secret) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	secretClient := client.(*secretmanager.Client)

	req := &secretmanagerpb.DeleteSecretRequest{
		Name: e.name,
	}

	err := secretClient.DeleteSecret(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (e *Secret) String() string {
	return e.name
}

func (e *Secret) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", e.name)
	properties.Set("CreationDate", e.creationDate)

	for labelKey, label := range e.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
