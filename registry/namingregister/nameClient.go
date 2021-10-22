package namingregister

import (
	"context"

	"github.com/gmsec/micro/naming"
)

type NamingClient interface {
	// Put puts a key-value pair
	Put(ctx context.Context, serviceName string, val naming.Update) error
	// Delete deletes a key, or optionally using WithRange(end), [key, end).
	Delete(ctx context.Context, serviceName string, val naming.Update) error

	// Get retrieves keys.
	Get(ctx context.Context, serviceName string) ([]*naming.Update, error)

	// Watchering Watcher is init
	Watchering() bool

	// Watch start watch
	Watch(ctx context.Context, serviceName string) error

	// WatcherNext watching next
	WatcherNext() ([]*naming.Update, error)

	// New new watching client
	New(serviceName string) NamingClient

	// Close close
	Close() error
}
