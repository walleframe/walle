package discovery

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/walleframe/walle/network"
)

var (
	ErrEntryTypeNotMatch = errors.New("discovery entry node type not match")
	ErrEntryUnmarshalKey = errors.New("discovery entry unmashal key failed")
)

type jsonEntryCodec struct{}

var NodeJsonEntryCodec EntryCodec = jsonEntryCodec{}

// Marshal marshal key/value to store.
func (ec jsonEntryCodec) Mashal(entry Entry) (key string, value []byte, err error) {
	node, ok := entry.(*Node)
	if !ok {
		err = ErrEntryTypeNotMatch
		return
	}
	key = node.Identifier
	value, err = json.Marshal(node)
	return
}

// Unmarshal unmarshal key/value from store.
func (ec jsonEntryCodec) Unmarshal(entry Entry, key string, value []byte) (err error) {
	node, ok := entry.(*Node)
	if !ok {
		err = ErrEntryTypeNotMatch
		return
	}
	err = json.Unmarshal(value, node)
	if err != nil {
		return err
	}
	node.Identifier = key
	return
}

// Node implement Entry interface.
type Node struct {
	// Identifier use for register path
	Identifier string            `json:"-"`
	Network    string            `json:"net,omitemtpy"`
	Addr       string            `json:"addr"`
	Balance    string            `json:"bn,omitemtpy"`
	Status     int               `json:"state,omitempty"`
	MD         map[string]string `json:"md,omitempty"`
	cli        network.Client
}

// Equals returns true if cmp contains the same data.
func (n *Node) Equals(e Entry) bool {
	if v, ok := e.(*Node); ok {
		return n.Network == v.Network &&
			n.Addr == v.Addr &&
			n.Identifier == v.Identifier
	}
	return false
}

// String returns the string form of an entry.
func (n *Node) String() string {
	return fmt.Sprintf("%s:{addr:%s state:%d balance:%s meta:%v}",
		n.Identifier, n.Addr, n.Status, n.Balance, n.MD,
	)
}

// Metadata get entry metadate
func (n *Node) Metadata(key string) (v string) {
	if n.MD == nil {
		return
	}
	v, _ = n.MD[key]
	return
}

// Address
func (n *Node) Address() (network, addr string) {
	return n.Network, n.Addr
}

// BalanceName get balance name
func (n *Node) BalanceName() string {
	return n.Balance
}

// State return node state
func (n *Node) State() EntryState {
	return EntryState(n.Status)
}

// ModifyState use for modify entry state
func (n *Node) ModifyState(state EntryState) {
	n.Status = int(state)
}

func (n *Node) Client() network.Client {
	return n.cli
}

func (n *Node) SetClient(cli network.Client) {
	n.cli = cli
}
