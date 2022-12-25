package resources

import (
	"context"
	"fmt"
	"path"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	iam "google.golang.org/api/iam/v1"
)

const ResourceTypeIAMRole = "IAMRole"

type IAMRole struct {
	id    string
	name  string
	stage string
}

func init() {
	register(ResourceTypeIAMRole, GetIAMClient, ListIAMRoles)
}

func ListIAMRoles(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	iamClient := client.(*gcputil.IAMClient)

	resources := make([]Resource, 0)
	pageToken := ""
	for {
		resp, err := iam.NewProjectsRolesService(iamClient.GetIAMService()).
			List(fmt.Sprintf("projects/%s", project.Name)).
			PageToken(pageToken).
			Do()
		if err != nil {
			return resources, err
		}

		for _, role := range resp.Roles {
			if !role.Deleted {
				resources = append(resources, &IAMRole{
					id:    role.Name,
					name:  path.Base(role.Name),
					stage: role.Stage,
				})
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return resources, nil
}

func (x *IAMRole) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	iamClient := client.(*gcputil.IAMClient)

	_, err = iam.NewProjectsRolesService(iamClient.GetIAMService()).Delete(x.id).Do()
	if err != nil {
		return err
	}

	return nil
}

func (x *IAMRole) GetOperationError(_ context.Context) error {
	return nil
}

func (x *IAMRole) String() string {
	return x.name
}

func (x *IAMRole) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Stage", x.stage)

	return properties
}
