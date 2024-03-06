package dynamicaccess

import (
	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Act interface{}

type defaultAct struct {
	//root       swarm.Address
	//sessionKey *ecdsa.PublicKey
	mockAct *mock.ActMock
}

func (a *defaultAct) Add(ref swarm.Address, value string) error {
	return a.mockAct.AddFunc(ref, value)
}

func (a *defaultAct) Get(index swarm.Address) (string, error) {
	return a.mockAct.GetFunc(index)
}

func NewAct() Act {
	return &defaultAct{
		mockAct: &mock.ActMock{},
	}
}
