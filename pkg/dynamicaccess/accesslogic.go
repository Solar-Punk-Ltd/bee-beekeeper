package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"errors"

	encryption "github.com/ethersphere/bee/pkg/encryption"
	file "github.com/ethersphere/bee/pkg/file"
	manifest "github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
	"golang.org/x/crypto/sha3"
)

var hashFunc = sha3.NewLegacyKeccak256

type AccessLogic interface {
	Get(act *Act, encryped_ref string, publisher string, tag string) (string, error)
	Add(act *Act, ref string, publisher string, tag string) (string, error)
	getLookUpKey(publisher string, tag string) (string, error)
	getAccessKeyDecriptionKey(publisher string, tag string) (string, error)
	getEncryptedAccessKey(act_root_hash string, lookup_key string) (manifest.Entry, error)
	createEncryptedAccessKey(ref string)
	Add_New_Grantee_To_Content(act *Act,ref string, publisher ecdsa.PublicKey) (*Act, error)
	// CreateAccessKey()
}

type DefaultAccessLogic struct {
	diffieHellman DiffieHellman
	//encryption    encryption.Interface
	act defaultAct
}

// Will give back Swarm reference with symmertic encryption key (128 byte)
// @publisher: public key


func actInit(ref string, publisher string, tag string) (*Act, encryptedRef string, error) {
	//act := NewAct()
}

// publisher is public key
func (al *DefaultAccessLogic) Add_New_Grantee_To_Content(act *Act, ref swarm.Address, publisher ecdsa.PublicKey) (*act , error){
	lookup_key := al.getLookUpKey(publisher_public_key, tag)
	akdk := al.getAccessKeyDecriptionKey(publisher_public_key, tag)
	//pseudo code like code
	if self_public_key == publisher {
		ak := encryption.GenerateRandomKey(encryption.KeyLength)
	} else {
		lookup_key := al.getLookUpKey(publisher_public_key, tag)
		akdk := al.getAccessKeyDecriptionKey(publisher_public_key, tag)
		encrypted_ak := al.getEncryptedAccessKey(act*Act, lookup_key)
		cipher := encryption.New(akdk, 0, uint32(0), hashFunc)
		ak := cipher.Decrypt(encrypted_ak)
	}

	access_key_cipher := encryption.New(ak, 0, uint32(0), hashFunc)
	encrypted_access_key := access_key_cipher.Encrypt([]byte(ak))
	ref_cipher := encryption.New(ak, 0, uint32(0), hashFunc).Encrypt([]byte(ref))
	encrypted_ref := ref_cipher.Encrypt([]byte(ref))

}


//
// act[lookupKey] := valamilyen_cipher.Encrypt(access_key)

// end of pseudo code like code

// func (al *DefaultAccessLogic) CreateAccessKey(reference string) {
// }

func (al *DefaultAccessLogic) getLookUpKey(publisher string, tag string) (string, error) {
	zeroByteArray := []byte{0}
	// Generate lookup key using Diffie Hellman
	lookup_key, err := al.diffieHellman.SharedSecret(publisher, tag, zeroByteArray)
	if err != nil {
		return "", err
	}
	return lookup_key, nil

}

func (al *DefaultAccessLogic) getAccessKeyDecriptionKey(publisher string, tag string) (string, error) {
	oneByteArray := []byte{1}
	// Generate access key decryption key using Diffie Hellman
	access_key_decryption_key, err := al.diffieHellman.SharedSecret(publisher, tag, oneByteArray)
	if err != nil {
		return "", err
	}
	return access_key_decryption_key, nil
}

func (al *DefaultAccessLogic) getEncryptedAccessKey(act *Act, lookup_key string) (manifest.Entry, error) {
	if act == nil {
		return nil, errors.New("no ACT root hash was provided")
	}
	if lookup_key == "" {
		return nil, errors.New("no lookup key")
	}

	manifest_raw, err := act.Get(lookup_key)
	if err != nil {
		return nil, err
	}
	//al.act.Get(act_root_hash)

	// Lookup encrypted access key from the ACT manifest
	var loadSaver file.LoadSaver
	var ctx context.Context
	loadSaver.Load(ctx, []byte(manifest_raw)) // Load the manifest file into loadSaver
	//y, err := x.Load(ctx, []byte(manifest_obj))
	manifestObj, err := manifest.NewDefaultManifest(loadSaver, false)
	if err != nil {
		return nil, err
	}
	encrypted_access_key, err := manifestObj.Lookup(ctx, lookup_key)
	if err != nil {
		return nil, err
	}

	return encrypted_access_key, nil
}

func (al *DefaultAccessLogic) Get(act *Act, encryped_ref string, publisher string, tag string) (string, error) {

	lookup_key, err := al.getLookUpKey(publisher, tag)
	if err != nil {
		return "", err
	}
	access_key_decryption_key, err := al.getAccessKeyDecriptionKey(publisher, tag)
	if err != nil {
		return "", err
	}

	// Lookup encrypted access key from the ACT manifest

	encrypted_access_key, err := al.getEncryptedAccessKey(act*Act, lookup_key)
	if err != nil {
		return "", err
	}

	// Decrypt access key
	access_key_cipher := encryption.New(encryption.Key(access_key_decryption_key), 4096, uint32(0), hashFunc)
	access_key, err := access_key_cipher.Decrypt(encrypted_access_key.Reference().Bytes())
	if err != nil {
		return "", err
	}

	// Decrypt reference
	ref_cipher := encryption.New(access_key, 4096, uint32(0), hashFunc)
	ref, err := ref_cipher.Decrypt([]byte(encryped_ref))
	if err != nil {
		return "", err
	}

	return string(ref), nil
}

func (al *DefaultAccessLogic) Add(act *Act, encryped_ref string, publisher string, tag string) (string, error) {
	//generrate access key
	access_key := encryption.GenerateRandomKey(10)
	lookup_key, err := al.getLookUpKey(publisher, tag)
	if err != nil {
		return "", err
	}
	access_key_decryption_key, err := al.getAccessKeyDecriptionKey(publisher, tag)
	if err != nil {
		return "", err
	}

	x, err := al.act.Add(oldItemKey, act_root_hash)
}

func NewAccessLogic(diffieHellman DiffieHellman) AccessLogic {
	return &DefaultAccessLogic{
		diffieHellman: diffieHellman,
		act:           defaultAct{},
	}
}
