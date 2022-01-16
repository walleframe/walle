package discovery

import (
	"context"
	net "net"
)

type NoOpRegistry struct{}

// NewEntry create entry,wait for register
func (NoOpRegistry) NewEntry(ctx context.Context, addr net.Addr) (err error) {
	return
}

// Online set entry online status and update to store
func (NoOpRegistry) Online(ctx context.Context) (err error) {
	return
}

// Offline set entry offline status and update to store
func (NoOpRegistry) Offline(ctx context.Context) (err error) {
	return
}

// Close clean entry info
func (NoOpRegistry) Clean(ctx context.Context) (err error) {
	return
}

type NoOpDiscovery struct{}

func (NoOpDiscovery) Watch(ctx context.Context) (changes <-chan Entries, err error) {
	return
}
func (NoOpDiscovery) GetAll(ctx context.Context) (all Entries, err error) {
	return
}
func (NoOpDiscovery) Close(ctx context.Context) {
	return
}
func (NoOpDiscovery) WatchEventNotify(ctx context.Context, eventNotify func(Entries)) (err error) {
	return
}
