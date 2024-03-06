package mock

import (
	"github.com/ethersphere/bee/pkg/swarm"
)

type ActMock struct {
	AddFunc func(ref swarm.Address, oldRootHash string) error
	GetFunc func(index swarm.Address) (string, error)
	data    map[string]string
}

func (act *ActMock) Add(ref swarm.Address, value string) error {
	if act.AddFunc == nil {
		act.data[ref.String()] = value
		return nil
	}
	return act.AddFunc(ref, value)
}

func (act *ActMock) Get(index swarm.Address) (string, error) {
	if act.GetFunc == nil {
		return act.data[index.String()], nil
	}
	return act.GetFunc(index)
}

func NewActMock() *ActMock {
	return &ActMock{data: make(map[string]string)}
}
