package resources

import (
	"fmt"

	"github.com/dshelley66/gcp-nuke/pkg/config"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
)

type ResourceMethods map[string]ResourceMethod

type ResourceMethod struct {
	Lister       ResourceLister
	ClientGetter ResourceClientGetter
}

type ResourceLister func(*gcputil.Project, gcputil.GCPClient) ([]Resource, error)
type ResourceClientGetter func(*gcputil.Project) (gcputil.GCPClient, error)

type Resource interface {
	Remove(*gcputil.Project, gcputil.GCPClient) error
}

type Filter interface {
	Resource
	Filter() error
}

type LegacyStringer interface {
	Resource
	String() string
}

type ResourcePropertyGetter interface {
	Resource
	Properties() types.Properties
}

type FeatureFlagGetter interface {
	Resource
	FeatureFlags(config.FeatureFlags)
}

var resourceMethods = make(ResourceMethods)

func register(name string, clientGetter ResourceClientGetter, lister ResourceLister) {
	_, exists := resourceMethods[name]
	if exists {
		panic(fmt.Sprintf("a resource with the name %s already exists", name))
	}

	resourceMethods[name] = ResourceMethod{
		ClientGetter: clientGetter,
		Lister:       lister,
	}

}

func GetLister(name string) ResourceLister {
	return resourceMethods[name].Lister
}

func GetClient(name string) ResourceClientGetter {
	return resourceMethods[name].ClientGetter
}

func GetListerNames() []string {
	names := []string{}
	for resourceType := range resourceMethods {
		names = append(names, resourceType)
	}

	return names
}
