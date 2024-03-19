package dynamicaccess

import (
	"crypto/ecdsa"
)

type Grantee interface {
	AddGrantees(topic string, addList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error)
	RemoveGrantees(topic string, removeList []*ecdsa.PublicKey) error
	GetGrantees(topic string) []*ecdsa.PublicKey
}

type defaultGrantee struct {
	grantees map[string][]*ecdsa.PublicKey
}


func (g *defaultGrantee) GetGrantees(topic string) []*ecdsa.PublicKey {
	return g.grantees[topic]
}

func (g *defaultGrantee) AddGrantees(topic string, addList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error) {
	for i, _ := range addList {
        g.grantees[topic] = append(g.grantees[topic], &addList[i])
    }
    return addList, nil
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
