package mock

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethersphere/bee/pkg/crypto"
)

type DiffieHellmanMock struct {
	key *ecdsa.PrivateKey
	//keyStoreService  keystore.Service
	SharedSecretFunc func(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error)
}

func (dhm *DiffieHellmanMock) SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) {
	if dhm.SharedSecretFunc == nil {
		//_, _, _ = dhm.keyStoreService.Key("test", "test", crypto.EDGSecp256_K1)
		x, _ := publicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, dhm.key.D.Bytes())
		if x == nil {
			return nil, errors.New("shared secret is point at infinity")
		}
		return crypto.LegacyKeccak256(append(x.Bytes(), moment...))
	}
	return dhm.SharedSecretFunc(publicKey, tag, moment)

}

func NewDiffieHellmanMock(key *ecdsa.PrivateKey) *DiffieHellmanMock {
	// return &DiffieHellmanMock{keyStoreService: KeyStoreMem.New()}
	return &DiffieHellmanMock{key: key}
}
