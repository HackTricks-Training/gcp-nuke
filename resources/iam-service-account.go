package resources

import (
	"context"
	"fmt"
	"path"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	iam "google.golang.org/api/iam/v1"
)

const ResourceTypeIAMServiceAccount = "IAMServiceAccount"

type IAMServiceAccount struct {
	id          string
	name        string
	displayName string
	disabled    bool
}

func init() {
	register(ResourceTypeIAMServiceAccount, GetIAMClient, ListIAMServiceAccounts)
}

func GetIAMClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeIAMServiceAccount); ok {
		return client, nil
	}

	client, err := gcputil.NewIAMClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IAM client: %v", err)
	}
	project.AddClient(ResourceTypeIAMServiceAccount, client)
	return client, nil
}

func ListIAMServiceAccounts(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	iamClient := client.(*gcputil.IAMClient)

	resources := make([]Resource, 0)
	pageToken := ""
	for {
		resp, err := iam.NewProjectsServiceAccountsService(iamClient.GetIAMService()).
			List(fmt.Sprintf("projects/%s", project.Name)).
			PageToken(pageToken).
			Do()
		if err != nil {
			return resources, err
		}

		for _, servAcct := range resp.Accounts {
			resources = append(resources, &IAMServiceAccount{
				id:          servAcct.Name,
				name:        path.Base(servAcct.Name),
				displayName: servAcct.DisplayName,
				disabled:    servAcct.Disabled,
			})
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return resources, nil
}

func (x *IAMServiceAccount) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	iamClient := client.(*gcputil.IAMClient)

	_, err = iam.NewProjectsServiceAccountsService(iamClient.GetIAMService()).Delete(x.id).Do()
	if err != nil {
		return err
	}

	return nil
}

func (x *IAMServiceAccount) GetOperationError(_ context.Context) error {
	return nil
}

func (x *IAMServiceAccount) String() string {
	return x.name
}

func (x *IAMServiceAccount) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("DisplayName", x.displayName)
	properties.Set("Disabled", x.disabled)

	return properties
}
