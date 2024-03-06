package mock

import "crypto/ecdsa"

type DiffieHellmanMock struct {
	SharedSecretFunc func(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error)
}

func (ma *DiffieHellmanMock) SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) {
	if ma.SharedSecretFunc == nil {
		return []byte{}, nil
	}
	return ma.SharedSecretFunc(publicKey, tag, moment)
}
