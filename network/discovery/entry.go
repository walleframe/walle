package discovery

import (
	"github.com/aggronmagi/walle/network"
)

type Client = network.Client

//go:generate mockgen -source entry.go -destination ../../testpkg/mock_discovery/entry.go

// EntryCodec Use for encode/decode entry from store
// EntryCodec 用于序列化和反序列化从store读取的信息
type EntryCodec interface {
	// Marshal marshal key/value to store.
	Mashal(e Entry) (key string, value []byte, err error)
	// Unmarshal unmarshal key/value from store.
	Unmarshal(e Entry, key string, value []byte) (err error)
}

// EntryState entry state
type EntryState int8

const (
	EntryStateOnline EntryState = iota
	EntryStateOffline
)

// An Entry represents a node.
type Entry interface {
	// Equals returns true if cmp contains the same data.
	Equals(Entry) bool
	// String returns the string form of an entry.
	String() string
	// Metadata get entry metadate
	Metadata(key string) string
	// Address
	Address() (network, addr string)
	// BalanceName get balance name
	BalanceName() string
	// State return node state
	State() EntryState
	// ModifyState use for modify entry state
	ModifyState(state EntryState)
	// Client
	Client() network.Client
	SetClient(network.Client)
}

// Entries is a list of *Entry with some helpers.
type Entries []Entry

// Equals returns true if cmp contains the same data.
func (e Entries) Equals(cmp Entries) bool {
	// Check if the file has really changed.
	if len(e) != len(cmp) {
		return false
	}
	for i := range e {
		if !e[i].Equals(cmp[i]) {
			return false
		}
	}
	return true
}

// Contains returns true if the Entries contain a given Entry.
func (e Entries) Contains(entry Entry) bool {
	for _, curr := range e {
		if curr.Equals(entry) {
			return true
		}
	}
	return false
}

// Diff compares two entries and returns the added and removed entries.
func (e Entries) Diff(cmp Entries) (Entries, Entries) {
	if len(e) < 1 {
		return cmp, nil
	}
	added := Entries{}
	for _, entry := range cmp {
		if !e.Contains(entry) {
			added = append(added, entry)
		}
	}

	removed := Entries{}
	for _, entry := range e {
		if !cmp.Contains(entry) {
			removed = append(removed, entry)
		}
	}

	return added, removed
}
