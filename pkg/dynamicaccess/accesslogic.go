package dynamicaccess

import (
	"hash"

	"github.com/ethersphere/bee/pkg/dynamicaccess/mock"
	encryption "github.com/ethersphere/bee/pkg/encryption"
)

type AccessLogic interface {
	Get(act_root_hash string, encryped_ref string, publisher string, tag string) (string, error)
}

type DefaultAccessLogic struct {
	diffieHellman DiffieHellman
	encryption    encryption.Interface
	act           defaultAct
}

// Will give back Swarm reference with symmertic encryption key (128 byte)
// @publisher: public key
func (al *DefaultAccessLogic) Get(act_root_hash string, encryped_ref string, publisher string, tag string) (string, error) {
	
	// Create byte arrays
	zeroByteArray := []byte{0}
	oneByteArray := []byte{1}
	// Generate lookup key using Diffie Hellman
	lookup_key, err := al.diffieHellman.SharedSecret(publisher, tag, zeroByteArray)
	if err!=nil {
		return "", err
	}
	// Generate access key decryption key using Diffie Hellman
	access_key_decryption_key, err := al.diffieHellman.SharedSecret(publisher, tag, oneByteArray)
	if err!=nil {
		return "", err
	}
	// Retrive MANIFEST from ACT
	manifest, err := al.act.Get(act_root_hash)
	al.act.Get(act_root_hash)
	// Lookup encrypted access key from the ACT manifest
	ENCRYPTED_ACCESS_KEY := SOME_LOOKUP_FUNCTION(manifest, lookup_key)

	// Decrypt access key
	ACCESS_KEY := DECRYPT(ENCRYPTED_ACCESS_KEY, access_key_decryption_key)
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
		act: defaultAct{},

		// {
		// 	AddFunc: func(ref string, publisher string, tag string) error {
		// 		return nil
		// 	},
		// 	GetFunc: func(ref string, publisher string, tag string) (string, error) {
		// 		return "", nil
		// 	},
		// },
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
