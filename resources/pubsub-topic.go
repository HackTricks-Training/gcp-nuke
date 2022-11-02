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

const ResourceTypePubSubTopic = "PubSubTopic"

type PubSubTopic struct {
	name   string
	labels map[string]string
}

func init() {
	register(ResourceTypePubSubTopic, GetPubSubClient, ListPubSubTopics)
}

func GetPubSubClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypePubSubTopic); ok {
		return client, nil
	}

	client, err := pubsub.NewClient(project.GetContext(), project.Name, project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %v", err)
	}
	project.AddClient(ResourceTypePubSubTopic, client)
	return client, nil
}

func ListPubSubTopics(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	pubsubClient := client.(*pubsub.Client)

	resources := make([]Resource, 0)

	it := pubsubClient.Topics(project.GetContext())
	for {
		topic, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list pubsub topics: %v", err)
		}
		// get the TopicConfig
		tc, err := topic.Config(project.GetContext())
		if err != nil {
			return nil, fmt.Errorf("failed to get pubsub topic config for '%s': %v", topic.String(), err)
		}
		resources = append(resources, &PubSubTopic{
			name:   path.Base(topic.String()),
			labels: tc.Labels,
		})
	}
	return resources, nil
}

func (x *PubSubTopic) Remove(project *gcputil.Project, client gcputil.GCPClient) error {
	pubsubClient := client.(*pubsub.Client)

	topic := pubsubClient.Topic(x.name)

	return topic.Delete(project.GetContext())
}

func (x *PubSubTopic) GetOperationError(_ context.Context) error {
	return nil
}

func (x *PubSubTopic) String() string {
	return x.name
}

func (x *PubSubTopic) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
