package resources

import (
	"context"
	"fmt"
	"path"

	"cloud.google.com/go/pubsub"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
)

const ResourceTypePubSubSubscription = "PubSubSubscription"

type PubSubSubscription struct {
	name   string
	labels map[string]string
}

func init() {
	register(ResourceTypePubSubSubscription, GetPubSubClient, ListPubSubSubscriptions)
}

func ListPubSubSubscriptions(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	pubsubClient := client.(*pubsub.Client)

	resources := make([]Resource, 0)

	it := pubsubClient.Subscriptions(project.GetContext())
	for {
		subscription, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list pubsub subscriptions: %v", err)
		}
		// get the SubscriptionConfig
		sc, err := subscription.Config(project.GetContext())
		if err != nil {
			return nil, fmt.Errorf("failed to get pubsub subscription config for '%s': %v", subscription.String(), err)
		}
		resources = append(resources, &PubSubSubscription{
			name:   path.Base(subscription.String()),
			labels: sc.Labels,
		})
	}
	return resources, nil
}

func (x *PubSubSubscription) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	pubsubClient := client.(*pubsub.Client)

	subscription := pubsubClient.Subscription(x.name)

	return subscription.Delete(project.GetContext())
}

func (x *PubSubSubscription) GetOperationError(_ context.Context) error {
	return nil
}

func (x *PubSubSubscription) String() string {
	return x.name
}

func (x *PubSubSubscription) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
