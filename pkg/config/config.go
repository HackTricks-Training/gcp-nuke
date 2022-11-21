package config

import (
	"fmt"
	"os"

	"github.com/dshelley66/gcp-nuke/pkg/types"
	"gopkg.in/yaml.v2"
)

type ResourceTypes struct {
	Targets      types.Collection `yaml:"targets"`
	Excludes     types.Collection `yaml:"excludes"`
	CloudControl types.Collection `yaml:"cloud-control"`
}

type Project struct {
	Locations     []string      `yaml:"locations"`
	Filters       Filters       `yaml:"filters"`
	ResourceTypes ResourceTypes `yaml:"resource-types"`
	Presets       []string      `yaml:"presets"`
}

type Nuke struct {
	ProjectRestrictedList []string                     `yaml:"project-restricted-list"`
	Projects              map[string]Project           `yaml:"projects"`
	ResourceTypes         ResourceTypes                `yaml:"resource-types"`
	Presets               map[string]PresetDefinitions `yaml:"presets"`
	FeatureFlags          FeatureFlags                 `yaml:"feature-flags"`
}

type FeatureFlags struct {
	DisableDeletionProtection DisableDeletionProtection `yaml:"disable-deletion-protection"`
}

type DisableDeletionProtection struct {
	ComputeInstance bool `yaml:"ComputeInstance"`
}

type PresetDefinitions struct {
	Filters Filters `yaml:"filters"`
}

type CustomService struct {
	Service               string `yaml:"service"`
	URL                   string `yaml:"url"`
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify"`
}

type CustomServices []*CustomService

type CustomRegion struct {
	Region                string         `yaml:"region"`
	Services              CustomServices `yaml:"services"`
	TLSInsecureSkipVerify bool           `yaml:"tls_insecure_skip_verify"`
}

type CustomEndpoints []*CustomRegion

func Load(path string) (*Nuke, error) {
	var err error

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(Nuke)
	err = yaml.UnmarshalStrict(raw, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Nuke) HasRestrictedList() bool {
	return c.ProjectRestrictedList != nil && len(c.ProjectRestrictedList) > 0
}

func (c *Nuke) InBlocklist(searchID string) bool {
	for _, restrictedProject := range c.ProjectRestrictedList {
		if restrictedProject == searchID {
			return true
		}
	}

	return false
}

func (c *Nuke) ValidateProject(projectID string) error {
	if !c.HasRestrictedList() {
		return fmt.Errorf("The config file contains an empty restricted list. " +
			"For safety reasons you need to specify at least one project ID. " +
			"This should be your production account.")
	}

	if c.InBlocklist(projectID) {
		return fmt.Errorf("You are trying to nuke the project with the ID %s, "+
			"but it is restricted. Aborting.", projectID)
	}

	if _, ok := c.Projects[projectID]; !ok {
		return fmt.Errorf("Your project '%s' isn't listed in the config. "+
			"Aborting.", projectID)
	}

	return nil
}

func (c *Nuke) Filters(accountID string) (Filters, error) {
	account := c.Projects[accountID]
	filters := account.Filters

	if filters == nil {
		filters = Filters{}
	}

	if account.Presets == nil {
		return filters, nil
	}

	for _, presetName := range account.Presets {
		notFound := fmt.Errorf("Could not find filter preset '%s'", presetName)
		if c.Presets == nil {
			return nil, notFound
		}

		preset, ok := c.Presets[presetName]
		if !ok {
			return nil, notFound
		}

		filters.Merge(preset.Filters)
	}

	return filters, nil
}
