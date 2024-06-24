// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvs

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/manifest"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type manifestStore struct {
	manifest manifest.Interface
	putCnt   int
}

var _ KeyValueStore = (*manifestStore)(nil)

// Get retrieves the value associated with the given key.
func (s *manifestStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	entry, err := s.manifest.Lookup(ctx, hex.EncodeToString(key))
	if err != nil {
		switch {
		case errors.Is(err, manifest.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get value from manifest %w", err)
		}
	}
	ref := entry.Reference()
	return ref.Bytes(), nil
}

// Put stores the given key-value pair in the store.
func (s *manifestStore) Put(ctx context.Context, key []byte, value []byte) error {
	err := s.manifest.Add(ctx, hex.EncodeToString(key), manifest.NewEntry(swarm.NewAddress(value), map[string]string{}))
	if err != nil {
		return fmt.Errorf("failed to put value to manifest %w", err)
	}
	s.putCnt++
	return nil
}

// Save saves key-value pair to the underlying storage and returns the reference.
func (s *manifestStore) Save(ctx context.Context) (swarm.Address, error) {
	if s.putCnt == 0 {
		return swarm.ZeroAddress, ErrNothingToSave
	}
	ref, err := s.manifest.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("failed to store manifest %w", err)
	}
	s.putCnt = 0
	return ref, nil
}

// NewManifest creates a new key-value store with simple manifest as the underlying storage.
func NewManifest(ls file.LoadSaver) (KeyValueStore, error) {
	m, err := manifest.NewSimpleManifest(ls)
	if err != nil {
		return nil, fmt.Errorf("failed to create simple manifest: %w", err)
	}

	return &manifestStore{
		manifest: m,
	}, nil
}

// NewManifestReference loads a key-value store from the given root hash with simple manifest as the underlying storage.
func NewManifestReference(ls file.LoadSaver, ref swarm.Address) (KeyValueStore, error) {
	m, err := manifest.NewSimpleManifestReference(ref, ls)
	if err != nil {
		return nil, fmt.Errorf("failed to create simple manifest reference: %w", err)
	}

	return &manifestStore{
		manifest: m,
	}, nil
}
