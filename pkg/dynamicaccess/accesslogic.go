package dynamicaccess

import (
	"crypto/ecdsa"
	"fmt"

	encryption "github.com/ethersphere/bee/pkg/encryption"
	"github.com/ethersphere/bee/pkg/swarm"
	"golang.org/x/crypto/sha3"
)

var hashFunc = sha3.NewLegacyKeccak256

type AccessLogic interface {
	AddPublisher(act Act, publisher ecdsa.PublicKey) (Act, error)
	Add_New_Grantee_To_Content(act Act, publisherPubKey, granteePubKey ecdsa.PublicKey) (Act, error)
	Get(act Act, encryped_ref swarm.Address, publisher ecdsa.PublicKey) (string, error)
	EncryptRef(act Act, publisherPubKey ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error)
}

type DefaultAccessLogic struct {
	diffieHellman DiffieHellman
}

// Adds a new publisher to an empty act
func (al *DefaultAccessLogic) AddPublisher(act Act, publisher ecdsa.PublicKey) (Act, error) {
	accessKey := encryption.GenerateRandomKey(encryption.KeyLength)

	lookupKey, err := al.getLookUpKey(publisher)
	if err != nil {
		return nil, err
	}

	accessKeyEncryptionKey, err := al.getAccessKeyDecriptionKey(publisher)
	if err != nil {
		return nil, err
	}

	accessKeyCipher := encryption.New(encryption.Key(accessKeyEncryptionKey), 0, uint32(0), hashFunc)
	encryptedAccessKey, err := accessKeyCipher.Encrypt([]byte(accessKey))
	if err != nil {
		return nil, err
	}

	act.Add([]byte(lookupKey), encryptedAccessKey)

	return act, nil
}

// Encrypts a SWARM reference for a publisher
func (al *DefaultAccessLogic) EncryptRef(act Act, publisherPubKey ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error) {
	accessKey := al.getAccessKey(act, publisherPubKey)
	refCipher := encryption.New(accessKey, 0, uint32(0), hashFunc)
	encryptedRef, _ := refCipher.Encrypt(ref.Bytes())

	return swarm.NewAddress(encryptedRef), nil
}

// Adds a new grantee to the ACT
func (al *DefaultAccessLogic) Add_New_Grantee_To_Content(act Act, publisherPubKey, granteePubKey ecdsa.PublicKey) (Act, error) {
	// Get previously generated access key
	accessKey := al.getAccessKey(act, publisherPubKey)

	// Encrypt the access key for the new Grantee (currently two Diffie-Hellman but propably this will change)
	lookupKey, err := al.getLookUpKey(granteePubKey)
	if err != nil {
		return nil, err
	}

	accessKeyEncryptionKey, err := al.getAccessKeyDecriptionKey(granteePubKey)
	if err != nil {
		return nil, err
	}

	// Encrypt the access key for the new Grantee
	cipher := encryption.New(encryption.Key(accessKeyEncryptionKey), 0, uint32(0), hashFunc)
	granteeEncryptedAccessKey, err := cipher.Encrypt(accessKey)
	if err != nil {
		return nil, err
	}

	// Add the new encrypted access key for the Act
	act.Add([]byte(lookupKey), granteeEncryptedAccessKey)

	return act, nil

}

// Will return the access key for a publisher (public key)
func (al *DefaultAccessLogic) getAccessKey(act Act, publisherPubKey ecdsa.PublicKey) []byte {
	publisherLookupKey, err := al.getLookUpKey(publisherPubKey)
	if err != nil {
		return nil
	}

	publisherAKDecryptionKey, err := al.getAccessKeyDecriptionKey(publisherPubKey)
	if err != nil {
		return nil
	}

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

// Gets the lookup key for a given grantee
func (al *DefaultAccessLogic) getLookUpKey(grantee ecdsa.PublicKey) (string, error) {
	zeroByteArray := []byte{0}
	// Generate lookup key using Diffie Hellman
	lookupKey, err := al.diffieHellman.SharedSecret(&grantee, "", zeroByteArray)
	if err != nil {
		return "", err
	}
	if len(lookupKey) != 32 {
		return "", fmt.Errorf("lookup key length is not 32 (not found)")
	}
	return string(lookupKey), nil

}

// Gets the access key decryption key for the publisher
func (al *DefaultAccessLogic) getAccessKeyDecriptionKey(publisher ecdsa.PublicKey) (string, error) {
	oneByteArray := []byte{1}

	// Generate access key decryption key using Diffie Hellman
	accessKeyDecryptionKey, err := al.diffieHellman.SharedSecret(&publisher, "", oneByteArray)
	if err != nil {
		return "", err
	}

	return string(accessKeyDecryptionKey), nil
}

// Gets the encrypted access key for a given grantee
func (al *DefaultAccessLogic) getEncryptedAccessKey(act Act, lookup_key string) ([]byte, error) {
	byteResult := act.Get([]byte(lookup_key))
	if len(byteResult) == 0 {
		return nil, fmt.Errorf("encrypted access key not found")
	}
	return byteResult, nil
}

// Get will return a decrypted reference, for given encrypted reference and grantee
func (al *DefaultAccessLogic) Get(act Act, encryped_ref swarm.Address, grantee ecdsa.PublicKey) (string, error) {
	if encryped_ref.Compare(swarm.EmptyAddress) == 0 {
		return "", nil
	}
	if grantee == (ecdsa.PublicKey{}) {
		return "", fmt.Errorf("grantee not provided")
	}

	lookupKey, err := al.getLookUpKey(grantee)
	if err != nil {
		return "", err
	}
	accessKeyDecryptionKey, err := al.getAccessKeyDecriptionKey(grantee)
	if err != nil {
		return "", err
	}

	// Lookup encrypted access key from the ACT manifest

	encryptedAccessKey, err := al.getEncryptedAccessKey(act, lookupKey)
	if err != nil {
		return "", err
	}

	// Decrypt access key
	accessKeyCipher := encryption.New(encryption.Key(accessKeyDecryptionKey), 0, uint32(0), hashFunc)
	accessKey, err := accessKeyCipher.Decrypt(encryptedAccessKey)
	if err != nil {
		return "", err
	}

	// Decrypt reference
	refCipher := encryption.New(accessKey, 0, uint32(0), hashFunc)
	ref, err := refCipher.Decrypt(encryped_ref.Bytes())
	if err != nil {
		return "", err
	}

	return string(ref), nil
}

func NewAccessLogic(diffieHellman DiffieHellman) AccessLogic {
	return &DefaultAccessLogic{
		diffieHellman: diffieHellman,
	}
}
