package kvstore

import (
	"context"
	"errors"
	"time"
)

// // Backend represents a KV Store Backend
// type Backend string

// const (
// 	// CONSUL backend
// 	CONSUL Backend = "consul"
// 	// ETCD backend
// 	ETCD Backend = "etcd"
// 	// ZK backend
// 	ZK Backend = "zk"
// 	// BOLTDB backend
// 	BOLTDB Backend = "boltdb"
// )

var (
	// ErrBackendNotSupported is thrown when the backend k/v store is not supported by libkv
	ErrBackendNotSupported = errors.New("Backend storage not supported yet, please choose one of")
	// ErrCallNotSupported is thrown when a method is not implemented/supported by the current backend
	ErrCallNotSupported = errors.New("The current call is not supported with this backend")
	// ErrNotReachable is thrown when the API cannot be reached for issuing common store operations
	ErrNotReachable = errors.New("Api not reachable")
	// ErrCannotLock is thrown when there is an error acquiring a lock on a key
	ErrCannotLock = errors.New("Error acquiring the lock")
	// ErrKeyModified is thrown during an atomic operation if the index does not match the one in the store
	ErrKeyModified = errors.New("Unable to complete atomic operation, key modified")
	// ErrKeyNotFound is thrown when the key is not found in the store during a Get operation
	ErrKeyNotFound = errors.New("Key not found in store")
	// ErrPreviousNotSpecified is thrown when the previous value is not specified for an atomic operation
	ErrPreviousNotSpecified = errors.New("Previous K/V pair should be provided for the Atomic operation")
	// ErrKeyExists is thrown when the previous value exists in the case of an AtomicPut
	ErrKeyExists = errors.New("Previous K/V pair exists, cannot complete Atomic operation")
)

// // Config contains the options for a storage client
// type Config struct {
// 	ClientTLS         *ClientTLSConfig
// 	TLS               *tls.Config
// 	ConnectionTimeout time.Duration
// 	Bucket            string
// 	PersistConnection bool
// 	Username          string
// 	Password          string
// }

// // ClientTLSConfig contains data for a Client TLS configuration in the form
// // the etcd client wants it.  Eventually we'll adapt it for ZK and Consul.
// type ClientTLSConfig struct {
// 	CertFile   string
// 	KeyFile    string
// 	CACertFile string
// }

// Store represents the backend K/V storage
// Each store should support every call listed
// here. Or it couldn't be implemented as a K/V
// backend for libkv
type Store interface {
	// Put a value at the specified key
	Put(ctx context.Context, key string, value []byte, opts ...WriteOption) error

	// Get a value given its key
	Get(ctx context.Context, key string) (*KVPair, error)

	// Delete the value at the specified key
	Delete(ctx context.Context, key string) error

	// Verify if a Key exists in the store
	Exists(ctx context.Context, key string) (bool, error)

	// Watch for changes on a key
	Watch(ctx context.Context, key string, stopCh <-chan struct{}) (<-chan *KVPair, error)

	// WatchTree watches for changes on child nodes under
	// a given directory
	WatchTree(ctx context.Context, directory string, stopCh <-chan struct{}) (<-chan []*KVPair, error)

	// List the content of a given prefix
	List(ctx context.Context, directory string) ([]*KVPair, error)

	// DeleteTree deletes a range of keys under a given directory
	DeleteTree(ctx context.Context, directory string) error

	// TODO: etcd实现
	// NewLock creates a lock for a given key.
	// The returned Locker is not held and must be acquired
	// with `.Lock`. The Value is optional.
	NewLock(ctx context.Context, key string, opts ...LockOption) (Locker, error) //

	// TODO: etcd实现
	// Atomic CAS operation on a single value.
	// Pass previous = nil to create a new key.
	AtomicPut(ctx context.Context, key string, value []byte, previous *KVPair, opts ...WriteOption) (bool, *KVPair, error)

	// TODO: etcd实现
	// Atomic delete of a single value
	AtomicDelete(ctx context.Context, key string, previous *KVPair) (bool, error)

	// Close the store connection
	Close(ctx context.Context)
}

// KVPair represents {Key, Value, Lastindex} tuple
type KVPair struct {
	Key       string
	Value     []byte
	LastIndex uint64
}

// WriteOptions contains optional request parameters
// TODO: Write选项合理化配置以及生效
//
//go:generate gogen option -n WriteOption -f Write -o option.write.go
func walleStoreWrite() interface{} {
	return map[string]interface{}{
		"IsDir": bool(false),
		"TTL":   time.Duration(0),
	}
}

// LockOptions contains optional request parameters
// TODO: Lock选项合理化配置以及生效
//
//go:generate gogen option -n LockOption -f Lock -o option.lock.go
func walleStoreLock() interface{} {
	return map[string]interface{}{
		// Value  Optional, value to associate with the lock
		"Value": []byte(nil),
		// TTL Optional, expiration ttl associated with the lock
		"TTL": time.Duration(0),
		// RenewLock Optional, chan used to control and stop the session ttl renewal for the lock
		"RenewLock": chan struct{}(nil),
	}
}

// TODO: 添加Read选项，合理化配置以及生效
// //go:generate gogen option -n ReadOption -f Read -o option.read.go
func walleStoreRead() interface{} {
	return map[string]interface{}{}
}

// Locker provides locking mechanism on top of the store.
// Similar to `sync.Lock` except it may return errors.
// TODO: 分布式锁实现
type Locker interface {
	Lock(stopChan chan struct{}) (<-chan struct{}, error)
	Unlock() error
}
