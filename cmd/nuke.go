package cmd

import (
	"fmt"
	"time"

	"github.com/dshelley66/gcp-nuke/pkg/config"
	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/types"
	"github.com/dshelley66/gcp-nuke/pkg/util"
	"github.com/dshelley66/gcp-nuke/resources"
	log "github.com/sirupsen/logrus"
)

type Nuke struct {
	Parameters NukeParameters
	Creds      *gcputil.Credentials
	Project    *gcputil.Project
	Config     *config.Nuke

	ResourceTypes types.Collection

	items Queue
}

func NewNuke(params NukeParameters, creds *gcputil.Credentials) *Nuke {
	n := Nuke{
		Parameters: params,
		Creds:      creds,
	}

	return &n
}

func (n *Nuke) Run() error {
	var err error

	defer func() {
		if n.Project != nil {
			n.Project.CloseClients()
		}
	}()

	if n.Parameters.ForceSleep < 3 && n.Parameters.NoDryRun {
		return fmt.Errorf("Value for --force-sleep cannot be less than 3 seconds if --no-dry-run is set. This is for your own protection.")
	}
	forceSleep := time.Duration(n.Parameters.ForceSleep) * time.Second

	fmt.Printf("gcp-nuke version %s - %s - %s\n\n", BuildVersion, BuildDate, BuildHash)

	err = n.Config.ValidateProject(n.Creds.Project)
	if err != nil {
		return err
	}

	fmt.Printf("Do you really want to nuke the project with the ID %s?\n", n.Creds.Project)
	if n.Parameters.Force {
		fmt.Printf("Waiting %v before continuing.\n", forceSleep)
		time.Sleep(forceSleep)
	} else {
		fmt.Printf("Do you want to continue? Enter project ID to continue.\n")
		err = Prompt(n.Creds.Project)
		if err != nil {
			return err
		}
	}

	err = n.Scan()
	if err != nil {
		return err
	}

	if n.items.Count(ItemStateNew) == 0 {
		fmt.Println("No resource to delete.")
		return nil
	}

	if !n.Parameters.NoDryRun {
		fmt.Println("The above resources would be deleted with the supplied configuration. Provide --no-dry-run to actually destroy resources.")
		return nil
	}

	fmt.Printf("Do you really want to nuke these resources on the project with the ID %s?\n", n.Creds.Project)
	if n.Parameters.Force {
		fmt.Printf("Waiting %v before continuing.\n", forceSleep)
		time.Sleep(forceSleep)
	} else {
		fmt.Printf("Do you want to continue? Enter project ID to continue.\n")
		err = Prompt(n.Creds.Project)
		if err != nil {
			return err
		}
	}

	failCount := 0
	waitingCount := 0

	for {
		n.HandleQueue()

		if n.items.Count(ItemStatePending, ItemStateWaiting, ItemStateNew) == 0 && n.items.Count(ItemStateFailed) > 0 {
			if failCount >= 2 {
				log.Errorf("There are resources in failed state, but none are ready for deletion, anymore.")
				fmt.Println()

				for _, item := range n.items {
					if item.State != ItemStateFailed {
						continue
					}

					item.Print()
					log.Error(item.Reason)
				}

				return fmt.Errorf("failed")
			}

			failCount = failCount + 1
		} else {
			failCount = 0
		}
		if n.Parameters.MaxWaitRetries != 0 && n.items.Count(ItemStateWaiting, ItemStatePending) > 0 && n.items.Count(ItemStateNew) == 0 {
			if waitingCount >= n.Parameters.MaxWaitRetries {
				return fmt.Errorf("Max wait retries of %d exceeded.\n\n", n.Parameters.MaxWaitRetries)
			}
			waitingCount = waitingCount + 1
		} else {
			waitingCount = 0
		}
		if n.items.Count(ItemStateNew, ItemStatePending, ItemStateFailed, ItemStateWaiting) == 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Printf("Nuke complete: %d failed, %d skipped, %d finished.\n\n",
		n.items.Count(ItemStateFailed), n.items.Count(ItemStateFiltered), n.items.Count(ItemStateFinished))

	return nil
}

func (n *Nuke) Scan() error {
	accountConfig := n.Config.Projects[n.Creds.Project]

	resourceTypes := ResolveResourceTypes(
		resources.GetListerNames(),
		map[string]string{},
		[]types.Collection{
			n.Parameters.Targets,
			n.Config.ResourceTypes.Targets,
			accountConfig.ResourceTypes.Targets,
		},
		[]types.Collection{
			n.Parameters.Excludes,
			n.Config.ResourceTypes.Excludes,
			accountConfig.ResourceTypes.Excludes,
		},
		[]types.Collection{},
	)

	n.Project.Locations = accountConfig.Locations
	queue := make(Queue, 0)

	items := Scan(n.Project, resourceTypes)
	for item := range items {
		ffGetter, ok := item.Resource.(resources.FeatureFlagGetter)
		if ok {
			ffGetter.FeatureFlags(n.Config.FeatureFlags)
		}

		queue = append(queue, item)
		err := n.Filter(item)
		if err != nil {
			return err
		}

		if item.State != ItemStateFiltered || !n.Parameters.Quiet {
			item.Print()
		}
	}

	fmt.Printf("Scan complete: %d total, %d nukeable, %d filtered.\n\n",
		queue.CountTotal(), queue.Count(ItemStateNew), queue.Count(ItemStateFiltered))

	n.items = queue

	return nil
}

func (n *Nuke) Filter(item *Item) error {

	checker, ok := item.Resource.(resources.Filter)
	if ok {
		err := checker.Filter()
		if err != nil {
			item.State = ItemStateFiltered
			item.Reason = err.Error()

			// Not returning the error, since it could be because of a failed
			// request to the API. We do not want to block the whole nuking,
			// because of an issue on GCP side.
			return nil
		}
	}

	accountFilters, err := n.Config.Filters(n.Creds.Project)
	if err != nil {
		return err
	}

	itemFilters, ok := accountFilters[item.Type]
	if !ok {
		return nil
	}

	for _, filter := range itemFilters {
		prop, err := item.GetProperty(filter.Property)
		if err != nil {
			log.Warnf(err.Error())
			continue
		}
		match, err := filter.Match(prop)
		if err != nil {
			return err
		}

		if IsTrue(filter.Invert) {
			match = !match
		}

		if match {
			item.State = ItemStateFiltered
			item.Reason = "filtered by config"
			return nil
		}
	}

	return nil
}

func (n *Nuke) HandleQueue() {
	listCache := make(map[string]map[string][]resources.Resource)

	for _, item := range n.items {
		switch item.State {
		case ItemStateNew:
			n.HandleRemove(item)
			item.Print()
		case ItemStateFailed:
			n.HandleRemove(item)
			n.HandleWait(item, listCache)
			item.Print()
		case ItemStatePending:
			n.HandleWait(item, listCache)
			item.State = ItemStateWaiting
			item.Print()
		case ItemStateWaiting:
			n.HandleWait(item, listCache)
			item.Print()
		}

	}

	fmt.Println()
	fmt.Printf("Removal requested: %d waiting, %d failed, %d skipped, %d finished\n\n",
		n.items.Count(ItemStateWaiting, ItemStatePending), n.items.Count(ItemStateFailed),
		n.items.Count(ItemStateFiltered), n.items.Count(ItemStateFinished))
}

func (n *Nuke) HandleRemove(item *Item) {
	clientGetter := resources.GetClient(item.Type)
	gcpClient, err := clientGetter(item.Project)
	if err != nil {
		dump := util.Indent(fmt.Sprintf("%v", err), "    ")
		log.Errorf("Listing %s failed:\n%s", item.Type, dump)
	}

	err = item.Resource.Remove(item.Project, gcpClient)
	if err != nil {
		item.State = ItemStateFailed
		item.Reason = err.Error()
		return
	}

	item.State = ItemStatePending
	item.Reason = ""
}

func (n *Nuke) HandleWait(item *Item, cache map[string]map[string][]resources.Resource) {
	var err error
	project := item.Project.Name
	if err := item.Resource.GetOperationError(item.Project.GetContext()); err != nil {
		item.State = ItemStateFailed
		item.Reason = err.Error()
		return
	}
	_, ok := cache[project]
	if !ok {
		cache[project] = map[string][]resources.Resource{}
	}
	left, ok := cache[project][item.Type]
	if !ok {
		left, err = item.List()
		if err != nil {
			item.State = ItemStateFailed
			item.Reason = err.Error()
			return
		}
		cache[project][item.Type] = left
	}

	for _, r := range left {
		if item.Equals(r) {
			checker, ok := r.(resources.Filter)
			if ok {
				err := checker.Filter()
				if err != nil {
					break
				}
			}

			return
		}
	}

	item.State = ItemStateFinished
	item.Reason = ""
}
