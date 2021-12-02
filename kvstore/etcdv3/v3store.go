package etcdv3

import (
	"context"
	"strings"

	"time"

	"bytes"

	"github.com/aggronmagi/walle/kvstore"
	clientv3 "go.etcd.io/etcd/client/v3"
	namespace "go.etcd.io/etcd/client/v3/namespace"

	// "github.com/coreos/etcd/clientv3"
	// "github.com/coreos/etcd/clientv3/namespace"
	"go.uber.org/zap"
)

// Option config etcdv3 store
//go:generate gogen option -n Option -o option.go
func walleOptions() interface{} {
	return map[string]interface{}{
		// etcd endpoints
		"Endpoints": []string{"127.0.0.1:2379"},
		"UserName":  "",
		"Password":  "",
		// context
		"Context": context.Context(context.Background()),
		// dial etcd timeout config
		"DialTimeout": time.Duration(0),
		// CustomSet custom set etcd options
		"CustomSet": func(cfg *clientv3.Config) {},
		// namespace
		"Namespace": "",
		// Lease Second and keepalive
		"Lease":  int64(5),
		"Logger": (*zap.Logger)(zap.NewNop()),
	}
}

type etcdStore struct {
	cli     *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	leaseID clientv3.LeaseID
	opts    *Options
	stop    chan struct{}
}

func New(opt ...Option) (_ kvstore.Store, err error) {
	// new options
	opts := NewOptions(opt...)
	opts.Logger = opts.Logger.Named("kvstore-etcd")
	log := opts.Logger
	cfg := clientv3.Config{
		Endpoints:   opts.Endpoints,
		DialTimeout: opts.DialTimeout,
		Username:    opts.UserName,
		Password:    opts.Password,
		Context:     opts.Context,
		// PermitWithoutStream:  false,
		// RejectOldCluster:     false,
		// AutoSyncInterval:     0,
		// DialKeepAliveTime:    0,
		// DialKeepAliveTimeout: 0,
		// MaxCallSendMsgSize:   0,
		// MaxCallRecvMsgSize:   0,
		// TLS:                  &tls.Config{},
		// DialOptions:          []grpc.DialOption{},
		// LogConfig:            &zap.Config{},
	}
	opts.CustomSet(&cfg)
	cli, err := clientv3.New(cfg)
	if err != nil {
		log.Error("new etcd client failed",
			zap.Error(err), zap.Any("config", cfg),
		)
		return
	}

	var lease clientv3.Lease
	leaseID := clientv3.LeaseID(0)

	if opts.Lease > 0 {
		lease = clientv3.NewLease(cli)
		if len(opts.Namespace) > 0 {
			lease = namespace.NewLease(lease, opts.Namespace)
		}
		rsp, err2 := lease.Grant(context.Background(), opts.Lease)
		if err2 != nil {
			log.Error("etcd client grant lease failed",
				zap.Error(err2), zap.Any("config", cfg),
				zap.Int64("lease", opts.Lease),
			)
			lease.Close()
			cli.Close()
			err = err2
			return
		}
		leaseID = rsp.ID
		krsp, err2 := lease.KeepAlive(context.Background(), leaseID)
		if err2 != nil {
			log.Error("etcd client keep alive lease failed",
				zap.Error(err2), zap.Any("config", cfg),
				zap.Int64("lease", opts.Lease),
			)
			lease.Revoke(context.Background(), leaseID)
			lease.Close()
			cli.Close()
			err = err2
			return
		}
		_ = <-krsp
	}

	kv := cli.KV
	if len(opts.Namespace) > 0 {
		kv = namespace.NewKV(kv, opts.Namespace)
	}

	log.Debug("etcd create success")

	return &etcdStore{
		cli:     cli,
		leaseID: leaseID,
		opts:    opts,
		kv:      kv,
		lease:   lease,
		stop:    make(chan struct{}),
	}, nil
}

// Put a value at the specified key
func (store *etcdStore) Put(ctx context.Context, key string, value []byte, opts ...kvstore.WriteOption) error {
	log := store.opts.Logger
	key = Normalize(key)
	_, err := store.kv.Put(ctx, key, string(value))
	if err != nil {
		log.Error("put failed",
			zap.String("key", key),
			zap.Binary("value", value),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Get a value given its key
func (store *etcdStore) Get(ctx context.Context, key string) (*kvstore.KVPair, error) {
	log := store.opts.Logger
	key = Normalize(key)
	rsp, err := store.kv.Get(ctx, key)
	if err != nil {
		log.Error("get failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}
	if rsp.Count == 0 {
		return nil, kvstore.ErrKeyNotFound
	}
	return &kvstore.KVPair{
		Key:       key,
		Value:     rsp.Kvs[0].Value,
		LastIndex: uint64(rsp.Kvs[0].ModRevision),
	}, nil
}

// Delete the value at the specified key
func (store *etcdStore) Delete(ctx context.Context, key string) error {
	log := store.opts.Logger
	key = Normalize(key)
	rsp, err := store.kv.Delete(ctx, key)
	if err != nil {
		log.Error("delete failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}
	if rsp.Deleted == 0 {
		return kvstore.ErrKeyNotFound
	}
	return nil
}

// Verify if a Key exists in the store
func (store *etcdStore) Exists(ctx context.Context, key string) (bool, error) {
	// log := store.opts.Logger
	_, err := store.Get(ctx, key)
	if err != nil {
		if err == kvstore.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// List the content of a given prefix
func (store *etcdStore) List(ctx context.Context, directory string) (res []*kvstore.KVPair, err error) {
	log := store.opts.Logger
	directory = Normalize(directory)
	rsp, err := store.kv.Get(ctx, directory,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	)
	if err != nil {
		log.Error("list failed",
			zap.String("directory", directory),
			zap.Error(err),
		)
		return
	}
	if rsp.Count == 0 {
		err = kvstore.ErrKeyNotFound
		return
	}
	res = make([]*kvstore.KVPair, 0, len(rsp.Kvs))
	key := []byte(directory) //util.StringToBytes(directory)
	for _, v := range rsp.Kvs {
		if bytes.Compare(v.Key, key) == 0 {
			continue
		}
		res = append(res, &kvstore.KVPair{
			Key:       string(v.Key),
			Value:     v.Value,
			LastIndex: uint64(v.ModRevision),
		})
	}
	return
}

// DeleteTree deletes a range of keys under a given directory
func (store *etcdStore) DeleteTree(ctx context.Context, directory string) error {
	log := store.opts.Logger
	directory = Normalize(directory)
	rsp, err := store.kv.Delete(ctx, directory, clientv3.WithPrefix())
	if err != nil {
		log.Error("delete-tree failed",
			zap.String("key", directory),
			zap.Error(err),
		)
		return err
	}
	if rsp.Deleted == 0 {
		return kvstore.ErrKeyNotFound
	}
	return nil
}

// Watch for changes on a key
func (store *etcdStore) Watch(ctx context.Context, key string, stopCh <-chan struct{}) (_ <-chan *kvstore.KVPair, err error) {
	log := store.opts.Logger
	// before watch,get its value first
	fv, err := store.Get(ctx, key)
	if err != nil {
		log.Error("watch key failed by get value first",
			zap.String("key", key), zap.Error(err),
		)
		return
	}
	ntf := make(chan *kvstore.KVPair)

	//
	go func() {
		watcher := clientv3.NewWatcher(store.cli)
		if len(store.opts.Namespace) > 0 {
			watcher = namespace.NewWatcher(watcher, store.opts.Namespace)
		}
		defer watcher.Close()
		defer close(ntf)
		// send first value
		ntf <- fv
		key = Normalize(key)
		ch := watcher.Watch(context.Background(), key)
		for {
			select {
			case <-stopCh:
				// stop watch
				log.Debug("stop watch normal")
				return
			case <-store.stop:
				// store close
				log.Debug("stop watch by close store")
				return
			case ev := <-ch:
				if ev.Err() != nil && ev.Canceled {
					log.Error("watch chan stop",
						zap.String("key", key),
						zap.Error(ev.Err()),
						zap.Any("ev", ev),
					)
					return
				}
				for _, v := range ev.Events {
					// NOTE: when delete key, len(v.Kv.Value) equal zero.
					ntf <- &kvstore.KVPair{
						Key:       key,
						Value:     v.Kv.Value,
						LastIndex: uint64(v.Kv.ModRevision),
					}
				}
			}
		}
	}()
	return ntf, nil
}

// WatchTree watches for changes on child nodes under
// a given directory
func (store *etcdStore) WatchTree(ctx context.Context, directory string, stopCh <-chan struct{}) (_ <-chan []*kvstore.KVPair, err error) {
	log := store.opts.Logger
	// before watch,get its value first
	fv, err := store.List(ctx, directory)
	if err != nil {
		log.Error("watch key failed by get value first",
			zap.String("key", directory), zap.Error(err),
		)
		return
	}
	ntf := make(chan []*kvstore.KVPair)

	//
	go func() {
		watcher := clientv3.NewWatcher(store.cli)
		if len(store.opts.Namespace) > 0 {
			watcher = namespace.NewWatcher(watcher, store.opts.Namespace)
		}
		defer watcher.Close()
		defer close(ntf)
		// send first value
		ntf <- fv
		key := Normalize(directory)
		ctx := context.Background()
		ch := watcher.Watch(ctx, key, clientv3.WithPrefix())
		for {
			select {
			case <-stopCh:
				// stop watch
				log.Debug("stop watch-tree normal")
				return
			case <-store.stop:
				// store close
				log.Debug("stop watch-tree by close store")
				return
			case ev := <-ch:
				if ev.Err() != nil && ev.Canceled {
					log.Error("watch-tree chan stop",
						zap.String("key", key),
						zap.Error(ev.Err()),
						zap.Any("ev", ev),
					)
					return
				}
				fv, err = store.List(ctx, key)
				if err != nil {
					log.Error("watch-tree list failed",
						zap.String("key", key),
						zap.Error(err),
					)
					return
				}
				ntf <- fv
			}
		}
	}()
	return ntf, nil
}

// NewLock creates a lock for a given key.
// The returned Locker is not held and must be acquired
// with `.Lock`. The Value is optional.
func (store *etcdStore) NewLock(ctx context.Context, key string, opts ...kvstore.LockOption) (kvstore.Locker, error) {
	// TODO: impletement
	return nil, nil
}

// Atomic CAS operation on a single value.
// Pass previous = nil to create a new key.
func (store *etcdStore) AtomicPut(ctx context.Context, key string, value []byte, previous *kvstore.KVPair, opts ...kvstore.WriteOption) (bool, *kvstore.KVPair, error) {
	// TODO: impletement
	return false, nil, nil
}

// Atomic delete of a single value
func (store *etcdStore) AtomicDelete(ctx context.Context, key string, previous *kvstore.KVPair) (bool, error) {
	// TODO: impletement
	return false, nil
}

// Close the store connection
func (store *etcdStore) Close(ctx context.Context) {
	log := store.opts.Logger
	if store.cli == nil {
		log.Debug("etcd store not init")
		return
	}
	// clean lease
	var err error
	if store.lease != nil && store.leaseID > 0 {
		_, err = store.lease.Revoke(ctx, store.leaseID)
		if err != nil {
			log.Error("etcd revoke lease failed",
				zap.Int64("leaseID", int64(store.leaseID)),
				zap.Error(err),
			)
		}
		store.lease.Close()
	}
	// close etcd client
	err = store.cli.Close()
	if err != nil {
		log.Error("etcd client close failed",
			zap.Error(err),
		)
	}
	close(store.stop)
	return
}

// Normalize the key for usage in Etcd
var Normalize = func(key string) string {
	key = kvstore.Normalize(key)
	return strings.TrimPrefix(key, "/")
}
