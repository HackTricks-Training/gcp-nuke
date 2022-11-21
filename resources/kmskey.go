package resources

import (
	"context"
	"fmt"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
)

const ResourceTypeKmsKey = "KMSKey"

type KmsKey struct {
	name         string
	keyRing      string
	labels       map[string]string
	creationDate string
}

func init() {
	register(ResourceTypeKmsKey, GetKMSClient, ListKmsKeys)
}

func GetKMSClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeKmsKey); ok {
		return client, nil
	}

	client, err := kms.NewKeyManagementClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %v", err)
	}
	project.AddClient(ResourceTypeKmsKey, client)
	return client, nil
}

func ListKmsKeys(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	kmsClient := client.(*kms.KeyManagementClient)

	resources := make([]Resource, 0)

	var reqKeyRing *kmspb.ListKeyRingsRequest
	for _, location := range project.Locations {
		reqKeyRing = &kmspb.ListKeyRingsRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}
		itKeyRing := kmsClient.ListKeyRings(project.GetContext(), reqKeyRing)
		for {
			keyRing, err := itKeyRing.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to list key rings for location %s: %v", location, err)
			}
			reqKey := &kmspb.ListCryptoKeysRequest{
				Parent: keyRing.Name,
			}
			itKey := kmsClient.ListCryptoKeys(project.GetContext(), reqKey)
			for {
				key, err := itKey.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("failed to list key for keyring %s: %v", keyRing.Name, err)
				}
				if isDestroyed(key.Primary) {
					continue
				}
				resources = append(resources, &KmsKey{
					name:         key.Name,
					keyRing:      keyRing.Name,
					creationDate: key.CreateTime.AsTime().Format(time.RFC3339),
					labels:       key.GetLabels(),
				})
			}
		}
	}

	return resources, nil
}

func (x *KmsKey) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	kmsClient := client.(*kms.KeyManagementClient)

	reqKeyVersions := &kmspb.ListCryptoKeyVersionsRequest{
		Parent: x.name,
	}
	itKeyVersions := kmsClient.ListCryptoKeyVersions(project.GetContext(), reqKeyVersions)
	for {
		keyVersion, err := itKeyVersions.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list key versions for key %s: %v", x.name, err)
		}

		if !isDestroyed(keyVersion) {
			reqKeyVersionDestroy := &kmspb.DestroyCryptoKeyVersionRequest{
				Name: keyVersion.Name,
			}
			_, err = kmsClient.DestroyCryptoKeyVersion(project.GetContext(), reqKeyVersionDestroy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (x *KmsKey) GetOperationError(_ context.Context) error {
	return nil
}

func isDestroyed(keyVersion *kmspb.CryptoKeyVersion) bool {
	if keyVersion == nil {
		return true
	}
	return keyVersion.State == kmspb.CryptoKeyVersion_DESTROYED || keyVersion.State == kmspb.CryptoKeyVersion_DESTROY_SCHEDULED
}

func (x *KmsKey) String() string {
	return x.name
}

func (x *KmsKey) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("KeyRing", x.keyRing)
	properties.Set("CreationDate", x.creationDate)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
