// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamicaccess

import (
	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/kvs"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Act interface {
	Add(rootHash swarm.Address, key []byte, val []byte) (swarm.Address, error)

	Lookup(rootHash swarm.Address, key []byte) ([]byte, error)
}

var _ Act = (*act)(nil)

type act struct {
	storage kvs.KeyValueStore
}

func (a *act) Add(rootHash swarm.Address, key []byte, val []byte) (swarm.Address, error) {
	return a.storage.Put(rootHash, key, val)
}

func (a *act) Lookup(rootHash swarm.Address, key []byte) ([]byte, error) {
	return a.storage.Get(rootHash, key)
}

func NewAct(ls file.LoadSaver) Act {
	return &act{
		storage: kvs.New(ls),
	}
}
