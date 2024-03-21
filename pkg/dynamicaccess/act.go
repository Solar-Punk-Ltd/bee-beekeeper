// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamicaccess

import (
	"github.com/ethersphere/bee/pkg/api"
	"github.com/ethersphere/bee/pkg/kvs"
	"github.com/ethersphere/bee/pkg/swarm"
)

/*
var lock = &sync.Mutex{}

type single struct {
	memoryMock map[string]manifest.Entry
}

var singleInMemorySwarm *single

func getInMemorySwarm() *single {
	if singleInMemorySwarm == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInMemorySwarm == nil {
			singleInMemorySwarm = &single{
				memoryMock: make(map[string]manifest.Entry)}
		}
	}
	return singleInMemorySwarm
}

func getMemory() map[string]manifest.Entry {
	ch := make(chan *single)
	go func() {
		ch <- getInMemorySwarm()
	}()
	mem := <-ch
	return mem.memoryMock
}
*/

// Act represents an interface for accessing and manipulating data.
type Act interface {
	// Add adds a key-value pair to the data store.
	Add(rootHash swarm.Address, key []byte, val []byte) (swarm.Address, error)

	// Lookup retrieves the value associated with the given key from the data store.
	Lookup(rootHash swarm.Address, key []byte) ([]byte, error)

	// Load loads the data store from the given address.
	//Load(addr swarm.Address) error

	// Store stores the current state of the data store and returns the address of the ACT.
	//Store() (swarm.Address, error)
}

var _ Act = (*inMemoryAct)(nil)

// inMemoryAct is a simple implementation of the Act interface, with in memory storage.
type inMemoryAct struct {
	storage kvs.KeyValueStore
}

func (act *inMemoryAct) Add(rootHash swarm.Address, key []byte, val []byte) (swarm.Address, error) {
	return act.storage.Put(rootHash, key, val)
}

func (act *inMemoryAct) Lookup(rootHash swarm.Address, key []byte) ([]byte, error) {
	return act.storage.Get(rootHash, key)
}

func NewInMemoryAct() Act {
	return &inMemoryAct{
		storage: kvs.NewmemoryKeyValueStore(swarm.EmptyAddress),
	}
}

func NewManifestAct(storer api.Storer) Act {
	return &inMemoryAct{
		storage: kvs.NewManifestKeyValueStore(storer),
	}
}
