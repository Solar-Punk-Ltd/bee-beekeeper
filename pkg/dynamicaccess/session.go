package dynamicaccess

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/keystore"
)

// Session represents an interface for a Diffie-Helmann key derivation
type Session interface {
	// Key returns a derived key for each nonce
	Key(publicKey *ecdsa.PublicKey, nonces [][]byte) ([][]byte, error)
}

var _ Session = (*session)(nil)

type session struct {
	key *ecdsa.PrivateKey
}

func (s *session) Key(publicKey *ecdsa.PublicKey, nonces [][]byte) ([][]byte, error) {
	x, _ := publicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, s.key.D.Bytes())
	if x == nil {
		return nil, errors.New("shared secret is point at infinity")
	}

	keys := make([][]byte, len(nonces))
	for i, nonce := range nonces {
		key, err := crypto.LegacyKeccak256(append(x.Bytes(), nonce...))
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}

	return keys, nil
}

func NewDefaultSession(key *ecdsa.PrivateKey) Session {
	return &session{
		key: key,
	}
}

func NewFromKeystore(ks keystore.Service, tag, password string) Session {
	return nil
}
