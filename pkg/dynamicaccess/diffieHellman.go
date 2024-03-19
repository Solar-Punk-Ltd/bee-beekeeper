package dynamicaccess

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/ethersphere/bee/pkg/crypto"
)

type DiffieHellman interface {
	SharedSecret(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) // tag- topic?
}

var _ DiffieHellman = (*defaultDiffieHellman)(nil)

type defaultDiffieHellman struct {
	key *ecdsa.PrivateKey
}

func (dh *defaultDiffieHellman) SharedSecret(publicKey *ecdsa.PublicKey, tag string, salt []byte) ([]byte, error) {
	fmt.Println("PUBLIC KEY INSIDE SharedSecret: ", publicKey)
	fmt.Println("x: ", publicKey.X)
	fmt.Println("y: ", publicKey.Y)
	fmt.Println("dh.key.D: ", dh.key.D)
	fmt.Println("Salt: ", salt)
	x, _ := publicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, dh.key.D.Bytes())
	fmt.Println("x: ", x)
	fmt.Println(crypto.LegacyKeccak256(append(x.Bytes(), salt...)))
	if x == nil {
		return nil, errors.New("shared secret is point at infinity")
	}
	return crypto.LegacyKeccak256(append(x.Bytes(), salt...))
}

func NewDiffieHellman(key *ecdsa.PrivateKey) DiffieHellman {
	return &defaultDiffieHellman{key: key}

}
