package mock

import (
	"crypto/ecdsa"

	"github.com/ethersphere/bee/pkg/swarm"
)

type GranteeListMock interface {
	Add(topic string, addList []*ecdsa.PublicKey) error
	Remove(topic string, removeList []*ecdsa.PublicKey) error
	Get(topic string) []*ecdsa.PublicKey
	Save() (swarm.Address, error)
}

type GranteeListStructMock struct {
	grantees map[string][]*ecdsa.PublicKey
}

func (g *GranteeListStructMock) Get(topic string) []*ecdsa.PublicKey {
	grantees := g.grantees[topic]
	keys := make([]*ecdsa.PublicKey, len(grantees))
	copy(keys, grantees)
	return keys
}

func (g *GranteeListStructMock) Add(topic string, addList []*ecdsa.PublicKey) error {
	g.grantees[topic] = append(g.grantees[topic], addList...)
	return nil
}

func (g *GranteeListStructMock) Remove(topic string, removeList []*ecdsa.PublicKey) error {
	for _, remove := range removeList {
		for i, grantee := range g.grantees[topic] {
			if *grantee == *remove {
				g.grantees[topic][i] = g.grantees[topic][len(g.grantees[topic])-1]
				g.grantees[topic] = g.grantees[topic][:len(g.grantees[topic])-1]
			}
		}
	}

	return nil
}

func (g *GranteeListStructMock) Save() (swarm.Address, error) {
	return swarm.EmptyAddress, nil
}

func NewGrantee() *GranteeListStructMock {
	return &GranteeListStructMock{grantees: make(map[string][]*ecdsa.PublicKey)}
}
