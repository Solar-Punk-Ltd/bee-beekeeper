package dynamicaccess

import (
	"crypto/ecdsa"
)

type Grantee interface {
	AddGrantees(addList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error)
	RemoveGrantees(removeList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error)
	GetGrantees() []ecdsa.PublicKey
}

type defaultGrantee struct {
	// topic    string            //lint:ignore U1000 Ignore unused struct field
	grantees []ecdsa.PublicKey
}

func (g *defaultGrantee) GetGrantees(topic string) []*ecdsa.PublicKey {
	return g.grantees[topic]
}

func (g *defaultGrantee) AddGrantees(addList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error) {
	g.grantees = append(g.grantees, addList...)
	return g.grantees, nil
}

func (g *defaultGrantee) RemoveGrantees(topic string, removeList []*ecdsa.PublicKey) error {
	for _, remove := range removeList {
		for i, grantee := range g.grantees[topic] {
			if grantee == remove {
				g.grantees[topic] = append(g.grantees[topic][:i], g.grantees[topic][i+1:]...)
			}
		}
	}
	return nil
}

func NewGrantee() Grantee {
	return &defaultGrantee{grantees: make(map[string][]*ecdsa.PublicKey)}
}
