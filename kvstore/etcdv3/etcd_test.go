package etcdv3_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/walleframe/walle/kvstore"
	"github.com/walleframe/walle/kvstore/etcdv3"
	"go.uber.org/zap"
)

func newStore(t testing.TB) kvstore.Store {
	log := zap.NewNop()
	var err error
	if false {
		log, err = zap.NewDevelopment()
		if err != nil {
			t.Fatal(err)
		}
	}
	store, err := etcdv3.New(
		etcdv3.WithDialTimeout(time.Second),
		etcdv3.WithEndpoints("127.0.0.1:2379"),
		etcdv3.WithLease(5),
		etcdv3.WithLogger(log),
	)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestEtcdStoreOpration(t *testing.T) {
	dats := []struct {
		key   string
		value []byte
	}{
		{"k1", []byte("v1")},
		// reset k1
		{"k1", []byte("v2")},
		{"k/3", []byte("v3")},
		{"k4/", []byte("v4")},
		{"/k5", []byte("vxx1")},
		{"k6", []byte("vxxx1")},
	}

	store := newStore(t)

	for _, v := range dats {
		t.Run(fmt.Sprintf("key: [%s]", v.key), func(t *testing.T) {
			ctx := context.Background()
			err := store.Put(ctx, v.key, v.value)
			assert.Nil(t, err, "put key")
			exists, err := store.Exists(ctx, v.key)
			assert.Nil(t, err, "exists key")
			assert.Equal(t, true, exists, "exists key value")
			kv, err := store.Get(ctx, v.key)
			assert.Nil(t, err, "get key")
			assert.EqualValues(t, v.value, kv.Value, "values compare")
			err = store.Delete(ctx, v.key)
			assert.Nil(t, err, "del key")
			_, err = store.Get(ctx, v.key)
			assert.NotNil(t, err, "get after del")
			assert.Equal(t, err, kvstore.ErrKeyNotFound, "get after del")
			exists, err = store.Exists(ctx, v.key)
			assert.Nil(t, err, "exists after del")
			assert.Equal(t, false, exists, "exists key value")
		})
	}
	store.Close(context.Background())
}

func TestEtcdStoreWatch(t *testing.T) {
	ctx := context.Background()
	key := etcdv3.Normalize("/test/key")
	store := newStore(t)

	err := store.Put(ctx, key, []byte("v1"))
	assert.Nil(t, err, "put key")

	ch := make(chan struct{})
	ret, err := store.Watch(ctx, key, ch)
	assert.Nil(t, err, "watch key")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		index := 1
		for {
			select {
			case kv, ok := <-ret:
				if !ok {
					return
				}
				t.Log(index, kv)
				if index > 2 {
					assert.EqualValues(t, int(0), len(kv.Value), "delete key")
				} else {
					assert.EqualValues(t, fmt.Sprintf("v%d", index),
						string(kv.Value), "watch value")
				}

				index++
			}
		}
	}()
	runtime.Gosched()
	time.Sleep(time.Millisecond * 200)

	err = store.Put(ctx, key, []byte("v2"))
	assert.Nil(t, err, "put key v2")

	time.Sleep(time.Millisecond * 200)

	err = store.Delete(ctx, key)
	assert.Nil(t, err, "del key")
	t.Log("delete key")

	time.Sleep(time.Millisecond * 200)

	close(ch)
	wg.Wait()

	store.Close(ctx)
}

func TestEtcdStoreWatchTree(t *testing.T) {
	ctx := context.Background()

	dats := []struct {
		key   string
		value string
	}{
		{"k1", "v1"},
		// reset k1
		{"k2", "v2"},
		{"k3", "v3"},
		{"k4", "v4"},
		{"k5", "vxx1"},
		{"k6", "vxxx1"},
	}

	keyPrefix := etcdv3.Normalize("/test/key")
	store := newStore(t)

	for _, v := range dats {
		err := store.Put(ctx, keyPrefix+"/"+v.key, []byte(v.value))
		assert.Nil(t, err, "put key")
	}

	ch := make(chan struct{})
	ret, err := store.WatchTree(ctx, keyPrefix+"/", ch)
	assert.Nil(t, err, "watch key")
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		index := 1
		for {
			select {
			case kv, ok := <-ret:
				if !ok {
					return
				}
				t.Log(index, kv)
				if index == 1 {
					assert.Equal(t, len(dats), len(kv), "watch tree 1 len")
					for k, v := range kv {
						c := dats[k]
						assert.Equal(t, keyPrefix+"/"+c.key, v.Key, "%d key", k)
						assert.Equal(t, c.value, string(v.Value), "%d key", k)
					}
				} else {
					assert.Equal(t, int(0), len(kv), "watch tree 2 len")
					t.Log("touch delete tree")
				}

				index++
			}
		}
	}()
	runtime.Gosched()
	time.Sleep(time.Millisecond * 200)

	// err = store.Put(ctx, keyPrefix+"/a1", []byte("a1"), nil)
	// assert.Nil(t, err, "put key v2")

	time.Sleep(time.Millisecond * 200)

	err = store.DeleteTree(ctx, keyPrefix+"/")
	assert.Nil(t, err, "del key")
	t.Log("delete key")

	time.Sleep(time.Millisecond * 200)

	// close(ch)
	wg.Wait()

	store.Close(ctx)
}
