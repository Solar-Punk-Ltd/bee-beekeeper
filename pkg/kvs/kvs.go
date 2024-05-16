// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvs

import (
	"context"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/kvs/manifest"
	"github.com/ethersphere/bee/v2/pkg/kvs/pot"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type KeyValueStore interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
	Put(ctx context.Context, key, value []byte) error
	Save(ctx context.Context) (swarm.Address, error)
}

func NewDefault(ls file.LoadSaver) (KeyValueStore, error) {
	return pot.New(ls)
}

func NewDefaultReference(ls file.LoadSaver, rootHash swarm.Address) (KeyValueStore, error) {
	return pot.NewReference(ls, rootHash)
}

func NewManifest(ls file.LoadSaver) (KeyValueStore, error) {
	return manifest.New(ls)
}

func NewManifestReference(ls file.LoadSaver, rootHash swarm.Address) (KeyValueStore, error) {
	return manifest.NewReference(ls, rootHash)
}
