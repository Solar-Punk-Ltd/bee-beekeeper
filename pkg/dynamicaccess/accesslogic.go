package dynamicaccess

import (
	"crypto/ecdsa"
	"fmt"

	encryption "github.com/ethersphere/bee/pkg/encryption"
	"github.com/ethersphere/bee/pkg/swarm"
	"golang.org/x/crypto/sha3"
)

var hashFunc = sha3.NewLegacyKeccak256

// Logic has the responsibility to return a ref for a given grantee and create new encrypted reference for a grantee
type Logic interface {
	// Adds a new publisher to an empty act
	AddPublisher(act Act, publisher *ecdsa.PublicKey) (Act, error) 
	// Adds a new grantee to the ACT
	AddNewGranteeToContent(act Act, publisherPubKey, granteePubKey *ecdsa.PublicKey) (Act, error)
	// Will return the access key for a publisher (public key)
	Get(act Act, encryped_ref swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, error)
	// Encrypts the provided SWARM reference
	EncryptRef(act Act, publisherPubKey *ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error)
}

type actLogic struct {
	session Session
}

var _ Logic = (*actLogic)(nil)

// Adds a new publisher to an empty act
func (al *actLogic) AddPublisher(act Act, publisher *ecdsa.PublicKey) (Act, error) {
	accessKey := encryption.GenerateRandomKey(encryption.KeyLength)

	keys, err := al.getKeys(publisher)
	if err != nil {
		return nil, err
	}
	lookupKey := keys[0]
	accessKeyEncryptionKey := keys[1]

	accessKeyCipher := encryption.New(encryption.Key(accessKeyEncryptionKey), 0, uint32(0), hashFunc)
	encryptedAccessKey, err := accessKeyCipher.Encrypt([]byte(accessKey))
	if err != nil {
		return nil, err
	}

	act.Add(lookupKey, encryptedAccessKey)

	return act, nil
}

// Encrypts a SWARM reference for a publisher
func (al *actLogic) EncryptRef(act Act, publisherPubKey *ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error) {
	accessKey := al.getAccessKey(act, publisherPubKey)
	refCipher := encryption.New(accessKey, 0, uint32(0), hashFunc)
	encryptedRef, _ := refCipher.Encrypt(ref.Bytes())

	return swarm.NewAddress(encryptedRef), nil
}

// Adds a new grantee to the ACT
func (al *actLogic) AddNewGranteeToContent(act Act, publisherPubKey, granteePubKey *ecdsa.PublicKey) (Act, error) {
	// Get previously generated access key
	accessKey := al.getAccessKey(act, publisherPubKey)

	// Encrypt the access key for the new Grantee
	keys, err := al.getKeys(granteePubKey)
	if err != nil {
		return nil, err
	}
	lookupKey := keys[0]
	accessKeyEncryptionKey := keys[1]

	// Encrypt the access key for the new Grantee
	cipher := encryption.New(encryption.Key(accessKeyEncryptionKey), 0, uint32(0), hashFunc)
	granteeEncryptedAccessKey, err := cipher.Encrypt(accessKey)
	if err != nil {
		return nil, err
	}

	// Add the new encrypted access key for the Act
	act.Add(lookupKey, granteeEncryptedAccessKey)

	return act, nil

}

// Will return the access key for a publisher (public key)
func (al *actLogic) getAccessKey(act Act, publisherPubKey *ecdsa.PublicKey) []byte {
	keys, err := al.getKeys(publisherPubKey)
	if err != nil {
		return nil
	}
	publisherLookupKey := keys[0]
	publisherAKDecryptionKey := keys[1]

	accessKeyDecryptionCipher := encryption.New(encryption.Key(publisherAKDecryptionKey), 0, uint32(0), hashFunc)
	encryptedAK, err := al.getEncryptedAccessKey(act, publisherLookupKey)
	if err != nil {
		return nil
	}

	accessKey, err := accessKeyDecryptionCipher.Decrypt(encryptedAK)
	if err != nil {
		return nil
	}

	return accessKey
}

func (al *actLogic) getKeys(publicKey *ecdsa.PublicKey) ([][]byte, error) {
	// Generate lookup key and access key decryption
	oneByteArray := []byte{1}
	zeroByteArray := []byte{0}

	keys, err := al.session.Key(publicKey, [][]byte{zeroByteArray, oneByteArray})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Gets the encrypted access key for a given grantee
func (al *actLogic) getEncryptedAccessKey(act Act, lookup_key []byte) ([]byte, error) {
	val, err := act.Lookup(lookup_key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Get will return a decrypted reference, for given encrypted reference and grantee
func (al *actLogic) Get(act Act, encryped_ref swarm.Address, grantee *ecdsa.PublicKey) (swarm.Address, error) {
	if encryped_ref.Compare(swarm.EmptyAddress) == 0 {
		return swarm.EmptyAddress, fmt.Errorf("encrypted ref not provided")
	}
	if grantee == nil {
		return swarm.EmptyAddress, fmt.Errorf("grantee not provided")
	}

	keys, err := al.getKeys(grantee)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	lookupKey := keys[0]
	accessKeyDecryptionKey := keys[1]

	// Lookup encrypted access key from the ACT manifest
	encryptedAccessKey, err := al.getEncryptedAccessKey(act, lookupKey)
	if err != nil {
		return swarm.EmptyAddress, err
	}

	// Decrypt access key
	accessKeyCipher := encryption.New(encryption.Key(accessKeyDecryptionKey), 0, uint32(0), hashFunc)
	accessKey, err := accessKeyCipher.Decrypt(encryptedAccessKey)
	if err != nil {
		return swarm.EmptyAddress, err
	}

	// Decrypt reference
	refCipher := encryption.New(accessKey, 0, uint32(0), hashFunc)
	ref, err := refCipher.Decrypt(encryped_ref.Bytes())
	if err != nil {
		return swarm.EmptyAddress, err
	}

	return swarm.NewAddress(ref), nil
}

func NewLogic(s Session) Logic {
	return &actLogic{
		session: s,
	}
}

// TODO
// interface -nél nem ACT típusokat kell használni hanem swarm assreess-t, és csak encrypt és decryptáló metódusok legyen
// OK defaultLogic helyett ACTLogic nevezékú
// OK lookup- key-ek byte-tá castolásást ellenőrizni
//  sof link grantee.go 46 sorához kicserélni