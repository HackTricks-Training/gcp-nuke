package cmd

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/dshelley66/gcp-nuke/pkg/gcputil"
	"github.com/dshelley66/gcp-nuke/pkg/util"
	"github.com/dshelley66/gcp-nuke/resources"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

const ScannerParallelQueries = 16

func Scan(project *gcputil.Project, resourceTypes []string) <-chan *Item {
	s := &scanner{
		items:     make(chan *Item, 100),
		semaphore: semaphore.NewWeighted(ScannerParallelQueries),
	}
	go s.run(project, resourceTypes)

	return s.items
}

type scanner struct {
	items     chan *Item
	semaphore *semaphore.Weighted
}

func (s *scanner) run(project *gcputil.Project, resourceTypes []string) {
	ctx := context.Background()

	for _, resourceType := range resourceTypes {
		s.semaphore.Acquire(ctx, 1)
		go s.list(project, resourceType)
	}

	// Wait for all routines to finish.
	s.semaphore.Acquire(ctx, ScannerParallelQueries)

	close(s.items)
}

func (s *scanner) list(project *gcputil.Project, resourceType string) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("%v\n\n%s", r.(error), string(debug.Stack()))
			dump := util.Indent(fmt.Sprintf("%v", err), "    ")
			log.Errorf("Listing %s failed:\n%s", resourceType, dump)
		}
	}()
	defer s.semaphore.Release(1)

	clientGetter := resources.GetClient(resourceType)
	gcpClient, err := clientGetter(project)
	if err != nil {
		dump := util.Indent(fmt.Sprintf("%v", err), "    ")
		log.Errorf("Listing %s failed:\n%s", resourceType, dump)
	}
	lister := resources.GetLister(resourceType)
	var rs []resources.Resource
	rs, err = lister(project, gcpClient)
	if err != nil {
		_, ok := err.(gcputil.ErrSkipRequest)
		if ok {
			log.Debugf("skipping request: %v", err)
			return
		}

		_, ok = err.(gcputil.ErrUnknownEndpoint)
		if ok {
			log.Warnf("skipping request: %v", err)
			return
		}

		dump := util.Indent(fmt.Sprintf("%v", err), "    ")
		log.Errorf("Listing %s failed:\n%s", resourceType, dump)
		return
	}

	for _, r := range rs {
		s.items <- &Item{
			Project:  project,
			Resource: r,
			State:    ItemStateNew,
			Type:     resourceType,
		}
	}
}
