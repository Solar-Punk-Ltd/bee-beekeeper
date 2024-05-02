// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvs

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/manifest"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type KeyValueStore interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
	Put(ctx context.Context, key, value []byte) error
	Save(ctx context.Context) (swarm.Address, error)
}

type keyValueStore struct {
	manifest manifest.Interface
	putCnt   int
}

var _ KeyValueStore = (*keyValueStore)(nil)

func (s *keyValueStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	fmt.Printf("bagoy get key: %v\n", hex.EncodeToString(key))
	entry, err := s.manifest.Lookup(ctx, hex.EncodeToString(key))
	if err != nil {
		fmt.Printf("bagoy kvs manif lookup err: %v\n", err)
		return nil, err
	}
	ref := entry.Reference()
	fmt.Printf("bagoy kvs found ref: %v\n", ref)
	return ref.Bytes(), nil
}

func (s *keyValueStore) Put(ctx context.Context, key []byte, value []byte) error {
	err := s.manifest.Add(ctx, hex.EncodeToString(key), manifest.NewEntry(swarm.NewAddress(value), map[string]string{}))
	if err != nil {
		return err
	}
	fmt.Printf("bagoy put key: %v\n", hex.EncodeToString(key))

	s.putCnt++
	return nil
}

func (s *keyValueStore) Save(ctx context.Context) (swarm.Address, error) {
	// if s.putCnt == 0 {
	// 	return swarm.ZeroAddress, errors.New("nothing to save")
	// }
	ref, err := s.manifest.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	s.putCnt = 0
	return ref, nil
}

func New(ls file.LoadSaver) (KeyValueStore, error) {
	manif, err := manifest.NewMantarayManifest(ls, false)
	if err != nil {
		return nil, err
	}

	return &keyValueStore{
		manifest: manif,
	}, nil
}

func NewReference(ls file.LoadSaver, rootHash swarm.Address) (KeyValueStore, error) {
	manif, err := manifest.NewMantarayManifestReference(rootHash, ls)
	if err != nil {
		return nil, err
	}

	return &keyValueStore{
		manifest: manif,
	}, nil
}
