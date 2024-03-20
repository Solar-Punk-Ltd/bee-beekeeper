package dynamicaccess

import (
	"crypto/ecdsa"
	"fmt"
)

type Grantee interface {
	AddGrantees(topic string, addList []*ecdsa.PublicKey) (error)
	RemoveGrantees(topic string, removeList []*ecdsa.PublicKey) error
	GetGrantees(topic string) []*ecdsa.PublicKey
}

type defaultGrantee struct {
	grantees map[string][]*ecdsa.PublicKey
}

func (g *defaultGrantee) GetGrantees(topic string) []*ecdsa.PublicKey {
	grantees := g.grantees[topic]
	keys := make([]*ecdsa.PublicKey, len(grantees))
	for i, key := range grantees {
		keys[i] = key
	}
	return keys
}

func (g *defaultGrantee) AddGrantees(topic string, addList []*ecdsa.PublicKey) (error) {
	for i, _ := range addList {
		g.grantees[topic] = append(g.grantees[topic], addList[i])
	}
	return nil
}

func (g *defaultGrantee) RemoveGrantees(topic string, removeList []*ecdsa.PublicKey) error {
	for _, remove := range removeList {
		for i, grantee := range g.grantees[topic] {
			if *grantee == *remove {
				fmt.Println("REMOVE")
				g.grantees[topic] = append(g.grantees[topic][:i], g.grantees[topic][i+1:]...)
			}
		}
	}
	return nil
}

func NewGrantee() Grantee {
	return &defaultGrantee{grantees: make(map[string][]*ecdsa.PublicKey)}
}