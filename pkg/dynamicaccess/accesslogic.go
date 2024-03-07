package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"hash"

	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	encryption "github.com/ethersphere/bee/pkg/encryption"
)

type AccessLogic interface {
	Get(ctx context.Context, act_root_hash string, encryped_ref string, publisher string, tag string) (string, error)
}

type DefaultAccessLogic struct {
	diffieHellman DiffieHellman
	encryption    encryption.Interface
	act           Act
}

// Will give back Swarm reference with symmertic encryption key (128 byte)
// @publisher: public key
func (al *DefaultAccessLogic) Get(ctx context.Context, act_root_hash string, encryped_ref string, publisher string, tag string) (string, error) {

	// Create byte arrays
	zeroByteArray := []byte{0}
	oneByteArray := []byte{1}
	// Generate lookup key using Diffie Hellman
	lookup_key, err := al.diffieHellman.SharedSecret(publisher, tag, zeroByteArray)
	if err != nil {
		return "", err
	}
	// Generate access key decryption key using Diffie Hellman
	access_key_decryption_key, err := al.diffieHellman.SharedSecret(publisher, tag, oneByteArray)
	if err != nil {
		return "", err
	}
	// Retrive MANIFEST from ACT
	// Lookup encrypted access key from the ACT manifest
	ENCRYPTED_ACCESS_KEY, err := al.act.Get(ctx, []byte(act_root_hash), lookup_key)

	// Decrypt access key
	ACCESS_KEY := DECRYPT(ENCRYPTED_ACCESS_KEY, access_key_decryption_key)
	// Decrypt reference
	return DECRYPT(encryped_ref, ACCESS_KEY)
}

// use DiffieHellmanMock
func NewAccessLogic(key encryption.Key, padding int, initCtr uint32, hashFunc func() hash.Hash) AccessLogic {
	return &DefaultAccessLogic{
		diffieHellman: &mock.DiffieHellmanMock{
			SharedSecretFunc: func(publicKey *ecdsa.PublicKey, tag string, moment []byte) ([]byte, error) {
				return []byte{}, nil
			},
		},
		encryption: encryption.New(key, padding, initCtr, hashFunc),
		act:        NewDefaultAct(),
	}
}

// -------
// act: &mock.ContainerMock{
// 	AddFunc: func(ref string, publisher string, tag string) error {
// 		return nil
// 	},
// 	GetFunc: func(ref string, publisher string, tag string) (string, error) {
// 		return "", nil
// 	},
// },
