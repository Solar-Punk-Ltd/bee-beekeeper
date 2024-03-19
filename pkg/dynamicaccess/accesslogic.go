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
	Get(act Act, encryped_ref swarm.Address, publisher ecdsa.PublicKey, tag string) (swarm.Address, error)
	EncryptRef(act Act, publisherPubKey ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error)
	//Add(act *Act, ref string, publisher ecdsa.PublicKey, tag string) (string, error)
	getKeys(publicKey ecdsa.PublicKey) ([][]byte, error)
	getEncryptedAccessKey(act Act, lookup_key []byte) ([]byte, error)
	//createEncryptedAccessKey(ref string)
	Add_New_Grantee_To_Content(act Act, publisherPubKey, granteePubKey ecdsa.PublicKey) (Act, error)
	Get(act Act, encryped_ref swarm.Address, publisher ecdsa.PublicKey) (swarm.Address, error)
	EncryptRef(act Act, publisherPubKey ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error)
}

type DefaultAccessLogic struct {
	session Session
	//encryption    encryption.Interface
}

// Adds a new publisher to an empty act
func (al *DefaultAccessLogic) AddPublisher(act Act, publisher ecdsa.PublicKey) (Act, error) {
	accessKey := encryption.GenerateRandomKey(encryption.KeyLength)

	keys, err := al.getKeys(publisher)
	if err != nil {
		return nil, err
	}
	lookup_key := keys[0]
	access_key_encryption_key := keys[1]

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

	// 2 Diffie-Hellman for the Grantee
	keys, err := al.getKeys(granteePubKey)
	if err != nil {
		return nil, err
	}
	lookup_key := keys[0]
	access_key_encryption_key := keys[1]

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
	keys, err := al.getKeys(publisherPubKey)
	if err != nil {
		return nil
	}
	publisher_lookup_key := keys[0]
	publisher_ak_decryption_key := keys[1]

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

//
// act[lookupKey] := valamilyen_cipher.Encrypt(access_key)

// end of pseudo code like code

// func (al *DefaultAccessLogic) CreateAccessKey(reference string) {
// }

func (al *DefaultAccessLogic) getKeys(publicKey ecdsa.PublicKey) ([][]byte, error) {
	// Generate lookup key and access key decryption
	oneByteArray := []byte{1}
	zeroByteArray := []byte{0}

	keys, err := al.session.Key(&publicKey, [][]byte{zeroByteArray, oneByteArray})
	if err != nil {
		return [][]byte{}, err
	}
	return keys, nil
}

// Gets the encrypted access key for a given grantee
func (al *DefaultAccessLogic) getEncryptedAccessKey(act Act, lookup_key []byte) ([]byte, error) {
	val, err := act.Lookup(lookup_key)
	if err != nil {
		return []byte{}, err
	}
	return val, nil
}

// Get will return a decrypted reference, for given encrypted reference and grantee
func (al *DefaultAccessLogic) Get(act Act, encryped_ref swarm.Address, grantee ecdsa.PublicKey) (swarm.Address, error) {
	if encryped_ref.Compare(swarm.EmptyAddress) == 0 {
		return swarm.EmptyAddress, nil
	}
	if grantee == (ecdsa.PublicKey{}) {
		return swarm.EmptyAddress, fmt.Errorf("grantee not provided")
	}

	keys, err := al.getKeys(publisher)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	lookup_key := keys[0]
	access_key_decryption_key := keys[1]

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

func NewAccessLogic(s Session) AccessLogic {
	return &DefaultAccessLogic{
		session: s,
	}
}
