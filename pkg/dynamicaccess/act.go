// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamicaccess

import (
	"encoding/hex"

	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
)

// Act represents an interface for accessing and manipulating data.
type Act interface {
	// Add adds a key-value pair to the data store.
	Add(key []byte, val []byte) Act

	// Lookup retrieves the value associated with the given key from the data store.
	Lookup(key []byte) []byte

	// Load retrieves the manifest entry associated with the given key from the data store.
	Load(key []byte) manifest.Entry

	// Store stores the given manifest entry in the data store.
	Store(me manifest.Entry)
}

var _ Act = (*defaultAct)(nil)

type defaultAct struct {
	container map[string]string
}

func (act *defaultAct) Add(key []byte, val []byte) Act {
	act.container[hex.EncodeToString(key)] = hex.EncodeToString(val)
	return act
}

func (act *defaultAct) Lookup(key []byte) []byte {
	if key, ok := act.container[hex.EncodeToString(key)]; ok {
		bytes, err := hex.DecodeString(key)
		if err == nil {
			return bytes
		}
	}
	return make([]byte, 0)
}

// to manifestEntry
func (act *defaultAct) Load(key []byte) manifest.Entry {
	return manifest.NewEntry(swarm.NewAddress(key), act.container)
}

// from manifestEntry
func (act *defaultAct) Store(me manifest.Entry) {
	if act.container == nil {
		act.container = make(map[string]string)
	}
	act.container = me.Metadata()
}

func NewDefaultAct() Act {
	return &defaultAct{
		container: make(map[string]string),
	}
}
