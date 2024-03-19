// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

import (
	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/manifest"
)

type ActMock struct {
	AddFunc    func(key []byte, value []byte) dynamicaccess.Act
	LookupFunc func(key []byte) []byte
	LoadFunc   func(key []byte) manifest.Entry
	StoreFunc  func(me manifest.Entry)
}

var _ dynamicaccess.Act = (*ActMock)(nil)

func (act *ActMock) Add(key []byte, value []byte) dynamicaccess.Act {
	if act.AddFunc == nil {
		return act
	}
	return act.AddFunc(key, value)
}

func (act *ActMock) Lookup(key []byte) []byte {
	if act.LookupFunc == nil {
		return make([]byte, 0)
	}
	return act.LookupFunc(key)
}

func (act *ActMock) Load(key []byte) manifest.Entry {
	if act.LoadFunc == nil {
		return nil
	}
	return act.LoadFunc(key)
}

func (act *ActMock) Store(me manifest.Entry) {
	if act.StoreFunc == nil {
		return
	}
	act.StoreFunc(me)
}

func NewActMock(addFunc func(key []byte, value []byte) dynamicaccess.Act, getFunc func(key []byte) []byte) dynamicaccess.Act {
	return &ActMock{
		AddFunc:    addFunc,
		LookupFunc: getFunc,
	}
}
