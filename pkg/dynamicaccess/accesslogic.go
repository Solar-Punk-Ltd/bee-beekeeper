package dynamicaccess

import (
	"hash"

	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	encryption "github.com/ethersphere/bee/pkg/encryption"
)

type AccessLogic interface {
	Get(encryped_ref string, publisher string, tag string) (string, error)
}

type DefaultAccessLogic struct {
	diffieHellman DiffieHellman
	encryption    encryption.Interface
	act           Act
}

// Will give back Swarm reference with symmertic encryption key (128 byte)
// @publisher: public key
func (al *DefaultAccessLogic) Get(act_root_hash string, encryped_ref string, publisher string, tag string) (string, error) {
	// Create byte arrays
	ZERO := CREATE_ZERO_BYTE_ARRAY()
	ONE := CREATE_ONE_BYTE_ARRAY()

	// Generate lookup key using Diffie Hellman
	LOOKUP_KEY, err := al.diffieHellman.SharedSecret(publisher, tag, ZERO)
	// Generate access key decryption key using Diffie Hellman
	ACCESS_KEY_DECRYPTION_KEY, err := al.diffieHellman.SharedSecret(publisher, tag, ONE)

	// Retrive MANIFEST from ACT
	MANIFEST, err := al.act.GetFunc(act_root_hash)
	// Lookup encrypted access key from the ACT manifest
	ENCRYPTED_ACCESS_KEY := MANIFEST.SOME_LOOKUP_FUNCTION(LOOKUP_KEY)

	// Decrypt access key
	ACCESS_KEY := DECRYPT(ENCRYPTED_ACCESS_KEY, ACCESS_KEY_DECRYPTION_KEY)
	// Decrypt reference
	return DECRYPT(encryped_ref, ACCESS_KEY)
}

func NewAccessLogic(key encryption.Key, padding int, initCtr uint32, hashFunc func() hash.Hash) AccessLogic {
	return &DefaultAccessLogic{
		diffieHellman: &mock.DiffieHellmanMock{
			SharedSecretFunc: func(publicKey string, tag string, moment []byte) (string, error) {
				return publicKey, nil
			},
		},
		encryption: encryption.New(key, padding, initCtr, hashFunc),
		act: &mock.ContainerMock{
			AddFunc: func(ref string, publisher string, tag string) error {
				return nil
			},
			GetFunc: func(ref string, publisher string, tag string) (string, error) {
				return "", nil
			},
		},
	}
}
