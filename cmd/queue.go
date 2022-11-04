package cmd

import (
	"fmt"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/util"
	"github.com/dshelley66/gcp-nuke/resources"
	log "github.com/sirupsen/logrus"
)

type ItemState int

// States of Items based on the latest request to GCP.
const (
	ItemStateNew ItemState = iota
	ItemStatePending
	ItemStateWaiting
	ItemStateFailed
	ItemStateFiltered
	ItemStateFinished
)

// An Item describes an actual GCP resource entity with the current state and
// some metadata.
type Item struct {
	Resource resources.Resource

	State  ItemState
	Reason string

	Project *gcputil.Project
	Type    string
}

func (i *Item) Print() {
	switch i.State {
	case ItemStateNew:
		Log(i.Project, i.Type, i.Resource, ReasonWaitPending, "would remove")
	case ItemStatePending:
		Log(i.Project, i.Type, i.Resource, ReasonWaitPending, "triggered remove")
	case ItemStateWaiting:
		Log(i.Project, i.Type, i.Resource, ReasonWaitPending, "waiting")
	case ItemStateFailed:
		Log(i.Project, i.Type, i.Resource, ReasonError, "failed")
		ReasonError.Printf("ERROR: %v\n", i.Reason)
	case ItemStateFiltered:
		Log(i.Project, i.Type, i.Resource, ReasonSkip, i.Reason)
	case ItemStateFinished:
		Log(i.Project, i.Type, i.Resource, ReasonSuccess, "removed")
	}
}

// List gets all resource items of the same resource type like the Item.
func (i *Item) List() ([]resources.Resource, error) {
	clientGetter := resources.GetClient(i.Type)
	gcpClient, err := clientGetter(i.Project)
	if err != nil {
		dump := util.Indent(fmt.Sprintf("%v", err), "    ")
		log.Errorf("Listing %s failed:\n%s", i.Type, dump)
		return nil, err
	}

	lister := resources.GetLister(i.Type)
	return lister(i.Project, gcpClient)
}

func (i *Item) GetProperty(key string) (string, error) {
	if key == "" {
		stringer, ok := i.Resource.(resources.LegacyStringer)
		if !ok {
			return "", fmt.Errorf("%T does not support legacy IDs", i.Resource)
		}
		return stringer.String(), nil
	}

	getter, ok := i.Resource.(resources.ResourcePropertyGetter)
	if !ok {
		return "", fmt.Errorf("%T does not support custom properties", i.Resource)
	}

	return getter.Properties().Get(key), nil
}

func (i *Item) Equals(o resources.Resource) bool {
	iType := fmt.Sprintf("%T", i.Resource)
	oType := fmt.Sprintf("%T", o)
	if iType != oType {
		return false
	}

	iStringer, iOK := i.Resource.(resources.LegacyStringer)
	oStringer, oOK := o.(resources.LegacyStringer)
	if iOK != oOK {
		return false
	}
	if iOK && oOK {
		return iStringer.String() == oStringer.String()
	}

	iGetter, iOK := i.Resource.(resources.ResourcePropertyGetter)
	oGetter, oOK := o.(resources.ResourcePropertyGetter)
	if iOK != oOK {
		return false
	}
	if iOK && oOK {
		return iGetter.Properties().Equals(oGetter.Properties())
	}

	return false
}

type Queue []*Item

func (q Queue) CountTotal() int {
	return len(q)
}

func (q Queue) Count(states ...ItemState) int {
	count := 0
	for _, item := range q {
		for _, state := range states {
			if item.State == state {
				count = count + 1
				break
			}
		}
	}
	return count
}
