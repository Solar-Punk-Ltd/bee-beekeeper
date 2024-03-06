package dynamicaccess

import (
	"crypto/ecdsa"

	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/pkg/keystore"
	KeyStoreMem "github.com/ethersphere/bee/pkg/keystore/mem"
)

type DiffieHellman interface {
	SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) // tag- topic?
}

type defaultDiffieHellman struct {
	//swarmPrivateKey, _, err := keystore.Key("swarm", password, crypto.EDGSecp256_K1)
	key             *ecdsa.PrivateKey
	keyStoreService keystore.Service
	//keyStoreEdg     keystore.EDG
}

func (d *defaultDiffieHellman) SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) {
	// Use mock.DiffieHellmanMock
	mock := &mock.DiffieHellmanMock{
		SharedSecretFunc: func(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) {
			//d.keyStoreService.Key("test", "test", Crypto.EDGSecp256_K1)
			b := make([]byte, 32)
			for i := range b {
				b[i] = 0xff
			}
			return b, nil
		},
	}
	return mock.SharedSecretFunc(publicKey, tag, moment)
}

func NewDiffieHellman(key *ecdsa.PrivateKey) DiffieHellman {
	return &defaultDiffieHellman{
		key:             key,
		keyStoreService: KeyStoreMem.New(),
		//keyStoreEdg:     Crypto.EDGSecp256_K1,
	}
}
