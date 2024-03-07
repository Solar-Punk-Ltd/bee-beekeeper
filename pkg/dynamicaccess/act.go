package dynamicaccess

import (
	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Act interface {
	Add(ref swarm.Address, value string) error
	Get(index swarm.Address) (string, error)
}

var _ Act = (*defaultAct)(nil)

type defaultAct struct {
	container *mock.ActMock
}

func (act *defaultAct) Add(ref swarm.Address, value string) error {
	return act.container.Add(ref, value)
}

func (act *defaultAct) Get(index swarm.Address) (string, error) {
	return act.container.Get(index)
}

func NewDefaultAct() Act {
	return &defaultAct{
		container: mock.NewActMock(),
	}
}
