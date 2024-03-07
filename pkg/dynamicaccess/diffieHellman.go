package dynamicaccess

import (
	"crypto/ecdsa"

	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
)

type DiffieHellman interface {
	SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) // tag- topic?
}

var _ DiffieHellman = (*defaultDiffieHellman)(nil)

type defaultDiffieHellman struct {
	mock *mock.DiffieHellmanMock
}

func (dhm *defaultDiffieHellman) SharedSecret(publicKey *ecdsa.PublicKey, tag string, salt []byte) ([]byte, error) {
	return dhm.mock.SharedSecret(publicKey, tag, salt)
}

func NewDiffieHellman(key *ecdsa.PrivateKey) DiffieHellman {
	return &defaultDiffieHellman{
		mock: mock.NewDiffieHellmanMock(key),
	}

}
