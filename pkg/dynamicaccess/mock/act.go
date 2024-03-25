// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

import (
	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/kvs"
	kvsmock "github.com/ethersphere/bee/pkg/kvs/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

type ActMock struct {
	storage kvs.KeyValueStore
}

var _ dynamicaccess.Act = (*ActMock)(nil)

func (act *ActMock) Add(root swarm.Address, key []byte, val []byte) (swarm.Address, error) {
	return act.storage.Put(root, key, val)
}

func (act *ActMock) Lookup(root swarm.Address, key []byte) ([]byte, error) {
	return act.storage.Get(root, key)
}

func NewActMock() *ActMock {
	return &ActMock{
		storage: kvsmock.New(),
	}
}
