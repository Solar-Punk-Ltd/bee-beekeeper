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

type MockDacService struct {
	ctrl dynamicaccess.Controller
}

// TODO: is a mockservice even needed ?
func NewService(ctrl dynamicaccess.Controller) *MockDacService {
	return &MockDacService{ctrl: ctrl}
}

func (m *MockDacService) DownloadHandler(ctx context.Context, timestamp int64, enryptedRef swarm.Address, pubkey *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error) {
	return m.ctrl.DownloadHandler(ctx, timestamp, enryptedRef, pubkey, historyRootHash)
}

func (m *MockDacService) UploadHandler(ctx context.Context, ref swarm.Address, pubkey *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, swarm.Address, error) {
	return m.ctrl.UploadHandler(ctx, ref, pubkey, historyRootHash)
}

func (m *MockDacService) Close() error {
	return nil
}

func (m *MockDacService) Grant(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	return nil
}
func (m *MockDacService) Revoke(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	return nil
}
func (m *MockDacService) Commit(ctx context.Context, granteesAddress swarm.Address, actRootHash swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, swarm.Address, error) {
	return swarm.ZeroAddress, swarm.ZeroAddress, nil
}
func (m *MockDacService) HandleGrantees(ctx context.Context, rootHash swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) error {
	return nil
}
func (m *MockDacService) GetGrantees(ctx context.Context, rootHash swarm.Address) ([]*ecdsa.PublicKey, error) {
	return nil, nil
}

var _ dynamicaccess.Controller = (*MockDacService)(nil)
