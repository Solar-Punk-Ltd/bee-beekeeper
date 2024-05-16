// Copyright 2024 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pot

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

type KeyValueStore interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
	Put(ctx context.Context, key, value []byte) error
	Save(ctx context.Context) (swarm.Address, error)
}

var _ KeyValueStore = (*potStore)(nil)

type potStore struct {
	idx    *potter.Index
	putCnt int
	ls     file.LoadSaver
}

func (ps *potStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	entry, err := ps.idx.Find(ctx, key)
	if err != nil {
		return nil, err
	}

	return entry.MarshalBinary()
}

func (ps *potStore) Put(ctx context.Context, key []byte, value []byte) error {
	entry, err := potter.NewPotEntry(key, value)
	if err != nil {
		return err
	}
	err = ps.idx.Add(ctx, entry)
	if err != nil {
		return err
	}
	ps.putCnt++
	return nil
}

func (ps *potStore) Save(ctx context.Context) (swarm.Address, error) {
	if ps.putCnt == 0 {
		return swarm.ZeroAddress, errors.New("nothing to save")
	}
	ref, err := ps.idx.Save(ctx)
	if err != nil {
		return swarm.ZeroAddress, fmt.Errorf("pot save error %w", err)
	}
	ps.putCnt = 0
	return swarm.NewAddress(ref), nil
}

func New(ls file.LoadSaver) (KeyValueStore, error) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("kvs_potter", log.WithSink(buf), log.WithVerbosity(log.VerbosityDebug))
	basePotMode := pot.NewSingleOrder(baseDepth)
	mode := pot.NewPersistedPot(basePotMode, ls, func() pot.Entry { return &potter.Entry{} })
	idx, err := potter.New(mode, logger)
	if err != nil {
		return nil, err
	}

	return &potStore{
		idx: idx,
		ls:  ls,
	}, nil
}

func NewReference(ls file.LoadSaver, ref swarm.Address) (KeyValueStore, error) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger("kvs_potter", log.WithSink(buf), log.WithVerbosity(log.VerbosityDebug))
	basePotMode := pot.NewSingleOrder(baseDepth)
	mode := pot.NewPersistedPotReference(basePotMode, ls, ref.Bytes(), func() pot.Entry { return &potter.Entry{} })
	idx, err := potter.NewReference(mode, ref.Bytes(), logger)
	if err != nil {
		return nil, err
	}

	return &potStore{
		idx: idx,
		ls:  ls,
	}, nil
}
