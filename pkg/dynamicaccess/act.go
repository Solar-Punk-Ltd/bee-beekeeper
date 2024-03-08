package dynamicaccess

import (
	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Act interface {
	Add(rootHash string, lookupKey []byte, encryptedAccessKey []byte) (swarm.Address, error)
	Get(rootHash string, lookupKey []byte) (string, error)
}

var _ Act = (*defaultAct)(nil)

type defaultAct struct {
	container *mock.ActMock
}

func (act *defaultAct) Add(rootHash string, lookupKey []byte, encryptedAccessKey []byte) (swarm.Address, error) {
	return act.container.Add(rootHash, lookupKey, encryptedAccessKey)
}

func (act *defaultAct) Get(rootHash string, lookupKey0 []byte) (string, error) {
	return act.container.Get(rootHash, lookupKey0)
}

func NewDefaultAct() Act {
	return &defaultAct{
		container: mock.NewActMock(),
	}
}
