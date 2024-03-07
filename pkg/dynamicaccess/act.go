package dynamicaccess

import (
	"context"

	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Act interface {
	Add(ctx context.Context, rootHash string, lookupKey []byte, encryptedAccessKey string) (swarm.Address, error)
	Get(ctx context.Context, rootHash []byte, lookupKey0 []byte) (string, error)
}

var _ Act = (*defaultAct)(nil)

type defaultAct struct {
	container *mock.ActMock
}

func (act *defaultAct) Add(ctx context.Context, rootHash string, lookupKey []byte, encryptedAccessKey string) (swarm.Address, error) {
	return act.container.Add(ctx, rootHash, lookupKey, encryptedAccessKey)
}

func (act *defaultAct) Get(ctx context.Context, rootHash []byte, lookupKey0 []byte) (string, error) {
	return act.container.Get(ctx, rootHash, lookupKey0)
}

func NewDefaultAct(m manifest.Interface) Act {
	return &defaultAct{
		container: mock.NewActMock(m),
	}
}
