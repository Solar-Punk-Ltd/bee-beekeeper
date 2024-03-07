package mock

import (
	"crypto/ecdsa"

	"github.com/ethersphere/bee/pkg/crypto"

	"github.com/ethersphere/bee/pkg/keystore"
	KeyStoreMem "github.com/ethersphere/bee/pkg/keystore/mem"
)

type DiffieHellmanMock struct {
	key              *ecdsa.PrivateKey
	keyStoreService  keystore.Service
	SharedSecretFunc func(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error)
}

func (dhm *DiffieHellmanMock) SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) {
	if dhm.SharedSecretFunc == nil {
		dhm.key, _, _ = dhm.keyStoreService.Key("test", "test", crypto.EDGSecp256_K1)
		b := make([]byte, 32)
		for i := range b {
			b[i] = 0xff
		}
		return b, nil
	}
	return dhm.SharedSecretFunc(publicKey, tag, moment)

}

func NewDiffieHellmanMock() *DiffieHellmanMock {
	return &DiffieHellmanMock{keyStoreService: KeyStoreMem.New()}
}
