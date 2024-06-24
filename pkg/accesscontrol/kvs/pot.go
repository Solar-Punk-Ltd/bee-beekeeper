// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kvs

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/log"
	"github.com/ethersphere/bee/v2/pkg/potter"
	"github.com/ethersphere/bee/v2/pkg/potter/pot"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

const baseDepth = 256

var _ KeyValueStore = (*potStore)(nil)

type potStore struct {
	idx    *potter.Index
	putCnt int
}

// Get retrieves the value associated with the given key.
func (ps *potStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	entry, err := ps.idx.Find(ctx, key)
	if err != nil {
		switch {
		case errors.Is(err, pot.ErrNotFound):
			return nil, ErrNotFound
		default:
			return nil, fmt.Errorf("failed to get value from pot %w", err)
		}
	}

	return entry.MarshalBinary()
}

// Put stores the given key-value pair in the store.
func (ps *potStore) Put(ctx context.Context, key []byte, value []byte) error {
	entry, err := potter.NewPotEntry(key, value)
	if err != nil {
		return err
	}
	err = ps.idx.Add(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to put value to pot %w", err)
	}
	ps.putCnt++
	return nil
}

// Save saves key-value pair to the underlying storage and returns the reference.
func (ps *potStore) Save(ctx context.Context) (swarm.Address, error) {
	if ps.putCnt == 0 {
		return swarm.ZeroAddress, ErrNothingToSave
	}
	ref, err := ps.idx.Save(ctx)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("failed to store pot %w", err)
	}
	ps.putCnt = 0
	return swarm.NewAddress(ref), nil
}

// NewPot creates a new key-value store with pot as the underlying storage.
func NewPot(ls file.LoadSaver) (KeyValueStore, error) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("kvs_potter", log.WithSink(buf), log.WithVerbosity(log.VerbosityDebug))
	basePotMode := pot.NewSingleOrder(baseDepth)
	mode := pot.NewPersistedPot(basePotMode, ls, func() pot.Entry { return &potter.Entry{} })
	idx, err := potter.New(mode, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create pot: %w", err)
	}

	return &potStore{
		idx: idx,
	}, nil
}

// NewReference loads a key-value store from the given root hash with pot as the underlying storage.
func NewPotReference(ls file.LoadSaver, ref swarm.Address) (KeyValueStore, error) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("kvs_potter", log.WithSink(buf), log.WithVerbosity(log.VerbosityDebug))
	basePotMode := pot.NewSingleOrder(baseDepth)
	mode := pot.NewPersistedPotReference(basePotMode, ls, ref.Bytes(), func() pot.Entry { return &potter.Entry{} })
	idx, err := potter.NewReference(mode, ref.Bytes(), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create pot reference: %w", err)
	}

	return &potStore{
		idx: idx,
	}, nil
}
