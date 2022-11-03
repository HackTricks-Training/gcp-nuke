package resources

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	redis "cloud.google.com/go/redis/apiv1"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"google.golang.org/api/iterator"
	redispb "google.golang.org/genproto/googleapis/cloud/redis/v1"
)

const ResourceTypeRedis = "Redis"

type Redis struct {
	name         string
	region       string
	zone         string
	labels       map[string]string
	creationDate string
	state        string
	operation    *redis.DeleteInstanceOperation
}

func init() {
	register(ResourceTypeRedis, GetRedisClient, ListRedis)
}

func GetRedisClient(project *gcputil.Project) (gcputil.GCPClient, error) {
	if client, ok := project.GetClient(ResourceTypeRedis); ok {
		return client, nil
	}

	client, err := redis.NewCloudRedisClient(project.GetContext(), project.Creds.GetNewClientOptions()...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %v", err)
	}
	project.AddClient(ResourceTypeRedis, client)
	return client, nil
}

func ListRedis(project *gcputil.Project, client gcputil.GCPClient) ([]Resource, error) {
	redisClient := client.(*redis.CloudRedisClient)

	resources := make([]Resource, 0)

	var reqRedis *redispb.ListInstancesRequest
	for _, location := range project.Locations {
		// global isn't a valid location for Redis
		if strings.ToLower(location) == "global" {
			continue
		}

		reqRedis = &redispb.ListInstancesRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", project.Name, location),
		}
		it := redisClient.ListInstances(project.GetContext(), reqRedis)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to list Redis instances for location %s: %v", location, err)
			}
			resources = append(resources, &Redis{
				name:         path.Base(resp.GetName()),
				region:       location,
				zone:         resp.GetLocationId(),
				creationDate: resp.GetCreateTime().AsTime().Format(time.RFC3339),
				labels:       resp.GetLabels(),
				state:        resp.GetState().Enum().String(),
			})
		}
	}

	return resources, nil
}

func (x *Redis) Remove(project *gcputil.Project, client gcputil.GCPClient) (err error) {
	redisClient := client.(*redis.CloudRedisClient)

	req := &redispb.DeleteInstanceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/instances/%s", project.Name, x.region, x.name),
	}
	x.operation, err = redisClient.DeleteInstance(project.GetContext(), req)
	if err != nil {
		return err
	}

	return nil
}

func (x *Redis) GetOperationError(ctx context.Context) (err error) {
	if x.operation != nil {
		err = x.operation.Poll(ctx)
	}
	return err
}

func (x *Redis) String() string {
	return x.name
}

func (x *Redis) Properties() types.Properties {
	properties := types.NewProperties()
	properties.Set("Name", x.name)
	properties.Set("Region", x.region)
	properties.Set("Zone", x.zone)
	properties.Set("CreationDate", x.creationDate)
	properties.Set("State", x.state)

	for labelKey, label := range x.labels {
		properties.SetTag(labelKey, label)
	}

	return properties
}
