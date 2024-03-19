package dynamicaccess

import "crypto/ecdsa"

type Grantee interface {
	AddGrantees(addList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error)
	RemoveGrantees(removeList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error)
	GetGrantees() []ecdsa.PublicKey
}

type defaultGrantee struct {
	// topic    string            //lint:ignore U1000 Ignore unused struct field
	grantees []ecdsa.PublicKey
}

func (g *defaultGrantee) GetGrantees() []ecdsa.PublicKey {
	return g.grantees
}

func (g *defaultGrantee) AddGrantees(addList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error) {
	g.grantees = append(g.grantees, addList...)
	return g.grantees, nil
}

func (g *defaultGrantee) RemoveGrantees(removeList []ecdsa.PublicKey) ([]ecdsa.PublicKey, error) {
	for _, remove := range removeList {
		for i, grantee := range g.grantees {
			if grantee == remove {
				g.grantees = append(g.grantees[:i], g.grantees[i+1:]...)
			}
		}
	}
	return g.grantees, nil
}

func NewGrantee() Grantee {
	return &defaultGrantee{}
}
