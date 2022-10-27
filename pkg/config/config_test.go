package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/dshelley66/gcp-nuke/pkg/types"
)

func TestConfigBlocklist(t *testing.T) {
	config := new(Nuke)

	if config.HasRestrictedList() {
		t.Errorf("HasBlocklist() returned true on a nil backlist.")
	}

	if config.InBlocklist("blubber") {
		t.Errorf("InBlocklist() returned true on a nil backlist.")
	}

	config.ProjectRestrictedList = []string{}

	if config.HasRestrictedList() {
		t.Errorf("HasBlocklist() returned true on a empty backlist.")
	}

	if config.InBlocklist("foobar") {
		t.Errorf("InBlocklist() returned true on a empty backlist.")
	}

	config.ProjectRestrictedList = append(config.ProjectRestrictedList, "bim")

	if !config.HasRestrictedList() {
		t.Errorf("HasBlocklist() returned false on a backlist with one element.")
	}

	if !config.InBlocklist("bim") {
		t.Errorf("InBlocklist() returned false on looking up an existing value.")
	}

	if config.InBlocklist("baz") {
		t.Errorf("InBlocklist() returned true on looking up an non existing value.")
	}
}

func TestLoadExampleConfig(t *testing.T) {
	config, err := Load("test-fixtures/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	expect := Nuke{
		ProjectRestrictedList: []string{"production-project"},
		Projects: map[string]Project{
			"gcp-test-project": {
				Locations: []string{"eu-west-1"},
				Presets:   []string{"terraform"},
				Filters: Filters{
					"IAMRole": {
						NewExactFilter("uber.admin"),
					},
					"IAMRolePolicyAttachment": {
						NewExactFilter("uber.admin -> AdministratorAccess"),
					},
				},
				ResourceTypes: ResourceTypes{
					Targets: types.Collection{"S3Bucket"},
				},
			},
		},
		ResourceTypes: ResourceTypes{
			Targets:  types.Collection{"DynamoDBTable", "S3Bucket", "S3Object"},
			Excludes: types.Collection{"IAMRole"},
		},
		Presets: map[string]PresetDefinitions{
			"terraform": {
				Filters: Filters{
					"S3Bucket": {
						Filter{
							Type:  FilterTypeGlob,
							Value: "my-statebucket-*",
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(*config, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", *config)
		t.Errorf("  Expected: %#v", expect)
	}
}

func TestConfigValidation(t *testing.T) {
	config, err := Load("test-fixtures/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		ID         string
		Aliases    []string
		ShouldFail bool
	}{
		{ID: "gcp-test-project", Aliases: []string{"staging"}, ShouldFail: false},
		{ID: "1234567890", Aliases: []string{"staging"}, ShouldFail: true},
		{ID: "1111111111", Aliases: []string{"staging"}, ShouldFail: true},
		{ID: "555133742", Aliases: []string{"production"}, ShouldFail: true},
		{ID: "555133742", Aliases: []string{}, ShouldFail: true},
		{ID: "555133742", Aliases: []string{"staging", "prod"}, ShouldFail: true},
	}

	for i, tc := range cases {
		name := fmt.Sprintf("%d_%s/%v/%t", i, tc.ID, tc.Aliases, tc.ShouldFail)
		t.Run(name, func(t *testing.T) {
			err := config.ValidateProject(tc.ID)
			if tc.ShouldFail && err == nil {
				t.Fatal("Expected an error but didn't get one.")
			}
			if !tc.ShouldFail && err != nil {
				t.Fatalf("Didn't excpect an error, but got one: %v", err)
			}
		})
	}
}

func TestFilterMerge(t *testing.T) {
	config, err := Load("test-fixtures/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	filters, err := config.Filters("gcp-test-project")
	if err != nil {
		t.Fatal(err)
	}

	expect := Filters{
		"S3Bucket": []Filter{
			{
				Type: "glob", Value: "my-statebucket-*",
			},
		},
		"IAMRole": []Filter{
			{
				Type:  "exact",
				Value: "uber.admin",
			},
		},
		"IAMRolePolicyAttachment": []Filter{
			{
				Type:  "exact",
				Value: "uber.admin -> AdministratorAccess",
			},
		},
	}

	if !reflect.DeepEqual(filters, expect) {
		t.Errorf("Read struct mismatches:")
		t.Errorf("  Got:      %#v", filters)
		t.Errorf("  Expected: %#v", expect)
	}
}
