// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package manifest

import (
	"context"
	"encoding/hex"
	"errors"
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

type mainfestStore struct {
	manifest manifest.Interface
	putCnt   int
	ls       file.LoadSaver
}

var _ KeyValueStore = (*mainfestStore)(nil)

func (ms *mainfestStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	entry, err := ms.manifest.Lookup(ctx, hex.EncodeToString(key))
	if err != nil {
		return nil, err
	}
	ref := entry.Reference()
	return ref.Bytes(), nil
}

func (ms *mainfestStore) Put(ctx context.Context, key []byte, value []byte) error {
	err := ms.manifest.Add(ctx, hex.EncodeToString(key), manifest.NewEntry(swarm.NewAddress(value), map[string]string{}))
	if err != nil {
		return err
	}
	ms.putCnt++
	return nil
}

func (ms *mainfestStore) Save(ctx context.Context) (swarm.Address, error) {
	if ms.putCnt == 0 {
		return swarm.ZeroAddress, errors.New("nothing to save")
	}
	ref, err := ms.manifest.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("manifest save error %w", err)
	}
	ms.putCnt = 0
	return ref, nil
}

func New(ls file.LoadSaver) (KeyValueStore, error) {
	manif, err := manifest.NewSimpleManifest(ls)
	if err != nil {
		return nil, err
	}

	return &mainfestStore{
		manifest: manif,
		ls:       ls,
	}, nil
}

func NewReference(ls file.LoadSaver, rootHash swarm.Address) (KeyValueStore, error) {
	manif, err := manifest.NewSimpleManifestReference(rootHash, ls)
	if err != nil {
		return nil, err
	}

	return &mainfestStore{
		manifest: manif,
		ls:       ls,
	}, nil
}
