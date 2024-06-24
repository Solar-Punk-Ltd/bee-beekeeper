// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kvs provides functionalities needed
// for storing key-value pairs on Swarm.
//
//nolint:ireturn
package kvs

import (
	"context"
	"errors"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

var (
	// ErrNothingToSave indicates that no new key-value pair was added to the store.
	ErrNothingToSave = errors.New("nothing to save")
	// ErrNotFound is returned when an Entry is not found in the storage.
	ErrNotFound = errors.New("entry not found")
)

// KeyValueStore represents a key-value store.
type KeyValueStore interface {
	// Get retrieves the value associated with the given key.
	Get(ctx context.Context, key []byte) ([]byte, error)
	// Put stores the given key-value pair in the store.
	Put(ctx context.Context, key, value []byte) error
	// Save saves key-value pair to the underlying storage and returns the reference.
	Save(ctx context.Context) (swarm.Address, error)
}

// NewDefault creates a new default key-value store with pot.
func NewDefault(ls file.LoadSaver) (KeyValueStore, error) {
	return NewPot(ls)
}

// NewDefaultReference loads the default key-value store with pot from the given root hash.
func NewDefaultReference(ls file.LoadSaver, rootHash swarm.Address) (KeyValueStore, error) {
	return NewPotReference(ls, rootHash)
}

// NewManifestKvs creates a new key-value store with a simple manifest.
func NewManifestKvs(ls file.LoadSaver) (KeyValueStore, error) {
	return NewManifest(ls)
}

// NewManifestKvsReference loads a key-value store with a simple manifest from the given root hash.
func NewManifestKvsReference(ls file.LoadSaver, rootHash swarm.Address) (KeyValueStore, error) {
	return NewManifestReference(ls, rootHash)
}
