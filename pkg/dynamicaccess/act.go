// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamicaccess

import (
	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/kvs"
	"github.com/ethersphere/bee/pkg/swarm"
)

// Act represents an interface for accessing and manipulating data.
type Act interface {
	// Add adds a key-value pair to the data store.
	Add(rootHash swarm.Address, key []byte, val []byte) (swarm.Address, error)

	// Lookup retrieves the value associated with the given key from the data store.
	Lookup(rootHash swarm.Address, key []byte) ([]byte, error)
}

// act is an implementation of the Act interface that uses kvs storage.
type act struct {
	storage kvs.KeyValueStore
}

// Add adds a key-value pair to the in-memory data store.
func (a *act) Add(rootHash swarm.Address, key []byte, val []byte) (swarm.Address, error) {
	return a.storage.Put(rootHash, key, val)
}

// Lookup retrieves the value associated with the given key from the in-memory data store.
func (a *act) Lookup(rootHash swarm.Address, key []byte) ([]byte, error) {
	return a.storage.Get(rootHash, key)
}

func NewAct(ls file.LoadSaver) Act {
	return &act{
		storage: kvs.New(ls),
	}
}
