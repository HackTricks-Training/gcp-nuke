package resources

// import (
// 	"fmt"

// 	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
// 	"github.com/dshelley66/gcp-nuke/pkg/types"
// 	iam "google.golang.org/api/iam/v1"
// 	"google.golang.org/api/iterator"
// 	iampb "google.golang.org/genproto/googleapis/cloud/iam/v1"
// )

// const ResourceTypeIAMServiceAccount = "IAMServiceAccount"

// type IAMServiceAccount struct {
// 	name         string
// 	network      string
// 	creationDate string
// 	operation    *iam.Operation
// }

// func init() {
// 	register(ResourceTypeIAMServiceAccount, GetIAMClient, ListIAMServiceAccounts)
// }

// func GetIAMClient(project *gcputil.Project) (gcputil.GCPClient, error) {
// 	if client, ok := project.GetClient(ResourceTypeIAMServiceAccount); ok {
// 		return client, nil
// 	}

// 	client, err := iam.NewService(project.GetContext(), project.Creds.GetNewClientOptions()...)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create IAM client: %v", err)
// 	}
// 	project.AddClient(ResourceTypeIAMServiceAccount, client)
// 	return client, nil
// }

// func ListIAMServiceAccounts(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
// 	IAMServiceAccountsClient := client.(*compute.IAMServiceAccountsClient)

// 	resources := make([]Resource, 0)
// 	req := &computepb.ListIAMServiceAccountsRequest{
// 		Project: project.Name,
// 		Filter:  &noDefaultNetworkFilter,
// 	}

// 	it := IAMServiceAccountsClient.List(project.GetContext(), req)
// 	for {
// 		resp, err := it.Next()
// 		if err == iterator.Done {
// 			break
// 		}

// 		if err != nil {
// 			return nil, fmt.Errorf("failed to list IAMServiceAccounts: %v", err)
// 		}
// 		resources = append(resources, &IAMServiceAccount{
// 			name:         *resp.Name,
// 			network:      *resp.Network,
// 			creationDate: *resp.CreationTimestamp,
// 		})
// 	}
// 	return resources, nil
// }

// func (x *IAMServiceAccount) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
// 	IAMServiceAccountsClient := client.(*compute.IAMServiceAccountsClient)

// 	req := &computepb.DeleteIAMServiceAccountRequest{
// 		IAMServiceAccount: x.name,
// 		Project:           project.Name,
// 	}

// 	x.operation, err = IAMServiceAccountsClient.Delete(project.GetContext(), req)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (x *IAMServiceAccount) GetOperationError(ctx context.Context) error {
// 	return getComputeOperationError(ctx, x.operation)
// }

// func (x *IAMServiceAccount) String() string {
// 	return x.name
// }

// func (x *IAMServiceAccount) Properties() types.Properties {
// 	properties := types.NewProperties()
// 	properties.Set("Name", x.name)
// 	properties.Set("CreationDate", x.creationDate)

// 	return properties
// }
