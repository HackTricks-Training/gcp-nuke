package gcputil

import (
	"context"
	"sync"
)

type Project struct {
	Name string

	Creds     *Credentials
	Locations []string
	clients   sync.Map
	ctx       context.Context
}

func (p *Project) GetClient(resourceType string) (GCPClient, bool) {
	client, ok := p.clients.Load(resourceType)
	if !ok {
		return nil, ok
	}
	return client.(GCPClient), ok
}

func (p *Project) AddClient(resourceType string, client GCPClient) {
	p.clients.Store(resourceType, client)
}

func (p *Project) GetContext() context.Context {
	return p.ctx
}

func (p *Project) CloseClients() {
	p.clients.Range(func(k, client interface{}) bool {
		client.(GCPClient).Close()
		return true
	})
}
func NewProject(creds *Credentials) *Project {
	return &Project{
		Name:    creds.Project,
		Creds:   creds,
		clients: sync.Map{},
		ctx:     context.Background(),
	}
}
