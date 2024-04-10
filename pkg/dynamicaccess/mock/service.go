// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethersphere/bee/v2/pkg/dynamicaccess"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type optionFunc func(*mockDac)

// Option is an option passed to a mock dynamicaccess Service.
type Option interface {
	apply(*mockDac)
}

func (f optionFunc) apply(r *mockDac) { f(r) }

// New creates a new mock dynamicaccess service.
// func New(o ...Option) dynamicaccess.Service {
// 	m := &mockDac{}
// 	for _, v := range o {
// 		v.apply(m)
// 	}

// 	return m
// }

func New(ctrl dynamicaccess.Controller, pk *ecdsa.PrivateKey) dynamicaccess.Service {
	m := &mockDac{key: pk, ctrl: ctrl}
	return m
}

type mockDac struct {
	key  *ecdsa.PrivateKey
	ctrl dynamicaccess.Controller
}

func (m *mockDac) DownloadHandler(ctx context.Context, timestamp int64, enryptedRef swarm.Address, _ *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error) {
	return m.ctrl.DownloadHandler(ctx, timestamp, enryptedRef, &m.key.PublicKey, historyRootHash)
}

func (m *mockDac) UploadHandler(ctx context.Context, ref swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, swarm.Address, error) {
	return m.ctrl.UploadHandler(ctx, ref, &m.key.PublicKey, historyRootHash)
}

func (m *mockDac) Close() error {
	return nil
}

func (m *mockDac) Grant(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	return nil
}
func (m *mockDac) Revoke(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	return nil
}
func (m *mockDac) Commit(ctx context.Context, granteesAddress swarm.Address, actRootHash swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, swarm.Address, error) {
	return swarm.ZeroAddress, swarm.ZeroAddress, nil
}
func (m *mockDac) HandleGrantees(ctx context.Context, rootHash swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) error {
	return nil
}
func (m *mockDac) GetGrantees(ctx context.Context, rootHash swarm.Address) ([]*ecdsa.PublicKey, error) {
	return nil, nil
}

var _ dynamicaccess.Controller = (*mockDac)(nil)
