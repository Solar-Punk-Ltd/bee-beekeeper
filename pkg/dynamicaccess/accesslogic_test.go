package dynamicaccess_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethersphere/bee/pkg/crypto"
	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/swarm"
)

func setupAccessLogic() dynamicaccess.AccessLogic {
	privateKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		errors.New("error creating private key")
		fmt.Println(err)
	}
	diffieHellman := dynamicaccess.NewDiffieHellman(privateKey)
	al := dynamicaccess.NewAccessLogic(diffieHellman)

	return al
}

func generateFixPrivateKey(input int64) ecdsa.PrivateKey {
	fixedD := big.NewInt(input) // Replace 42 with your desired fixed value
	privateKey := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(), // Use the desired elliptic curve
			X:     big.NewInt(0),   // These values can be anything since they're not used for testing GetEncryptedAccessKey
			Y:     big.NewInt(0),
		},
		D: fixedD, // Set the fixed value for the private key's D field
	}

	return privateKey
}

// Can't be called anymore, because getLookUpKey is not exported
// func TestGetLookupKey_Success(t *testing.T) {
// 	al := setupAccessLogic()

// 	id0, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	// ! this will be random, we can not know the lookup key for a randomly generated key
// 	act, encryptedRef, _ := al.AddPublisher(ACT, swarm.NewAddress([]byte("42")), id0.PublicKey, "")
// 	fmt.Println(act, encryptedRef)

// 	tag := "exampleTag"

// 	lookupKey, err := al.getLookUpKey(id0.PublicKey, tag)
// 	if err != nil {
// 		t.Errorf("Could not fetch lookup key from publisher and tag")
// 	}

// 	expectedLookupKey := "expectedLookupKey"
// 	if lookupKey != expectedLookupKey {
// 		fmt.Println(string(lookupKey))
// 		t.Errorf("The lookup key that was returned is not correct!")
// 	}
// }

// Can't be called anymore, because getLookUpKey is not exported
// func TestGetLookUpKey_Error(t *testing.T) {
// 	al := setupAccessLogic()

// 	invalidPublisher := ecdsa.PublicKey{}
// 	tag := "exampleTag"

// 	lookupKey, err := al.getLookUpKey(invalidPublisher, tag)

// 	if err != nil {
// 		t.Errorf("There was an error while fetching lookup key")
// 	}

// 	if lookupKey != "" {
// 		t.Errorf("Expected lookup key to be empty for invalid input")
// 	}
// }

// Can't be called anymore, because getAccessKeyDecriptionKey is not exported
// func TestGetAccessKeyDecriptionKey_Success(t *testing.T) {
// 	al := setupAccessLogic()

// 	id0, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	tag := "exampleTag"

// 	access_key_decryption_key, err := al.getAccessKeyDecriptionKey(id0.PublicKey, tag)
// 	if err != nil {
// 		t.Errorf("GetAccessKeyDecriptionKey gave back error")
// 	}

// 	expectedResult := "we-dont-know"
// 	if access_key_decryption_key != expectedResult {
// 		t.Errorf("The access key decryption key is not correct!")
// 	}
// }

// Can't be called anymore, because getAccessKeyDecriptionKey is not exported
// func TestGetAccessKeyDecriptionKey_Error(t *testing.T) {
// 	al := setupAccessLogic()

// 	invalidPublisher := ecdsa.PublicKey{}
// 	tag := "exampleTag"

// 	access_key_decryption_key, err := al.getAccessKeyDecriptionKey(invalidPublisher, tag)
// 	if err != nil {
// 		t.Errorf("GetAccessKeyDecriptionKey gave back error")
// 	}

// 	if access_key_decryption_key != "" {
// 		t.Errorf("GetAccessKeyDecriptionKey should give back empty string for invalid input!")
// 	}
// }

// Can't be called anymore, because getEncryptedAccessKey is not exported
// func TestGetEncryptedAccessKey_Success(t *testing.T) {
// 	al := setupAccessLogic()

// 	lookupKey := "exampleLookupKey"
// 	id0, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

// 	act, _, _ := al.ActInit(swarm.NewAddress([]byte("42")), id0.PublicKey, "")

// 	encrypted_access_key, err := al.getEncryptedAccessKey(*act, lookupKey)
// 	if err != nil {
// 		t.Errorf("There was an error while executing GetEncryptedAccessKey")
// 	}

// 	expectedEncryptedKey := "abc013encryptedkey"
// 	if encrypted_access_key.Reference().String() != expectedEncryptedKey {
// 		t.Errorf("GetEncryptedAccessKey didn't give back the expected value!")
// 	}
// }

// Can't be called anymore, because getEncryptedAccessKey is not exported
// func TestGetEncryptedAccessKey_Error(t *testing.T) {
// 	al := setupAccessLogic()

// 	lookupKey := "exampleLookupKey"
// 	id0, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

// 	act, _, _ := al.ActInit(swarm.NewAddress([]byte("42")), id0.PublicKey, "")
// 	empty_act_result, _ := al.getEncryptedAccessKey(*act, lookupKey)
// 	if empty_act_result != nil {
// 		t.Errorf("GetEncryptedAccessKey should give back nil for empty act root hash!")
// 	}

// 	empty_lookup_result, _ := al.getEncryptedAccessKey(*act, "")

// 	if empty_lookup_result != nil {
// 		t.Errorf("GetEncryptedAccessKey should give back nil for empty lookup key!")
// 	}
// }

func TestGet_Success(t *testing.T) {
	al := setupAccessLogic()
	id0 := generateFixPrivateKey(0)

	act := dynamicaccess.NewDefaultAct()
	act, _ = al.AddPublisher(act, id0.PublicKey, "")
	expectedRef := "39a5ea87b141fe44aa609c3327ecd896c0e2122897f5f4bbacf74db1033c5559"
	
	encryptedRef, _ := al.EncryptRef(act, id0.PublicKey, swarm.NewAddress([]byte(expectedRef)))
	tag := "exampleTag"
	
	ref, err := al.Get(act, encryptedRef, id0.PublicKey, tag)
	if err != nil {
		t.Errorf("There was an error while calling Get: ")
		t.Error(err)
	}
	
	if ref != expectedRef {
		t.Errorf("Get gave back wrong Swarm reference!")
	}
}

func TestGet_Error(t *testing.T) {
	al := setupAccessLogic()
	
	id0, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	act := dynamicaccess.NewDefaultAct()
	
	act, err = al.AddPublisher(act, id0.PublicKey, "")
	encryptedRef, _ := al.EncryptRef(act, id0.PublicKey, swarm.NewAddress([]byte("42")))
	
	if err != nil {
		t.Errorf("Error initializing Act")
		t.Errorf(err.Error())
	}
	//encryptedRef := "bzzabcasab"
	tag := "exampleTag"
	
	refOne, err := al.Get(act, encryptedRef, id0.PublicKey, tag)
	if err != nil {
		t.Errorf(err.Error())
	}
	if refOne != "" {
		t.Errorf("Get should give back empty string if ACT root hash not provided!")
	}
	
	refTwo, _ := al.Get(act, swarm.EmptyAddress, id0.PublicKey, tag)
	if refTwo != "" {
		t.Errorf("Get should give back empty string if encrypted ref not provided!")
	}
	
	refThree, _ := al.Get(act, encryptedRef, ecdsa.PublicKey{}, tag)
	if refThree != "" {
		t.Errorf("Get should give back empty string if publisher not provided!")
	}
	
	refFour, _ := al.Get(act, encryptedRef, id0.PublicKey, "")
	if refFour != "" {
		t.Errorf("Get should give back empty string if tag was not provided!")
	}
}

func TestNewAccessLogic(t *testing.T) {
	logic := setupAccessLogic()
	
	_, ok := logic.(*dynamicaccess.DefaultAccessLogic)
	if !ok {
		t.Errorf("NewAccessLogic: expected type *DefaultAccessLogic, got %T", logic)
	}
}

func TestAddPublisher(t *testing.T) {
	al := setupAccessLogic()
	id0 := generateFixPrivateKey(0)
	fmt.Println("id0", id0)
	savedLookupKey := "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
	act := dynamicaccess.NewDefaultAct()
	act, _ = al.AddPublisher(act, id0.PublicKey, "")
	decodedSavedLookupKey, _ := hex.DecodeString(savedLookupKey)
	encryptedAccessKey := act.Get(decodedSavedLookupKey)
	fmt.Println("encryptedAccessKey", encryptedAccessKey)
	decodedEncryptedAccessKey := hex.EncodeToString(encryptedAccessKey)
	// A random value is returned so it is only possibly to check the length of the returned value
	// We know the lookup key because the generated private key is fixed
	if len(decodedEncryptedAccessKey) != 64 {
		t.Errorf("AddPublisher: expected encrypted access key length 64, got %d", len(decodedEncryptedAccessKey))
		
	}
	if act == nil {
		t.Errorf("AddPublisher: expected act, got nil")
	}
}

func TestAdd_New_Grantee_To_Content(t *testing.T) {
	al := setupAccessLogic()
	id0 := generateFixPrivateKey(0)
	id1 := generateFixPrivateKey(1)
	
	act := dynamicaccess.NewDefaultAct()
	act, _ = al.AddPublisher(act, id0.PublicKey, "")
	
	act, _ = al.Add_New_Grantee_To_Content(act, id0.PublicKey, id1.PublicKey)
}

func TestEncryptRef(t *testing.T) {
	
	ref := "39a5ea87b141fe44aa609c3327ecd896c0e2122897f5f4bbacf74db1033c5559"
	savedEncryptedRef := "230cdcfb2e67adddb2822b38f70105213ab3e4f97d03560bfbfbb218f487c5303e9aa9a97e62aa1a8003f162679e7c65e1c8e3aacaec2043fd5d2a4a7d69285e"
	
	fmt.Println("ref", ref)
	al := setupAccessLogic()
	id0 := generateFixPrivateKey(0)
	act := dynamicaccess.NewDefaultAct()
	decodedLookupKey, _ := hex.DecodeString("bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a")
	act.Add(decodedLookupKey, []byte("42"))
	fmt.Println("act", act.Get([]byte("42")))
	fmt.Println("act", act)
	encryptedRefValue, _ := al.EncryptRef(act, id0.PublicKey, swarm.NewAddress([]byte(ref)))
	act, _ = al.AddPublisher(act, id0.PublicKey, "")
	if encryptedRefValue.String() != savedEncryptedRef {
		t.Errorf("EncryptRef: expected encrypted ref, got empty address")
	}
}

// func TestAddGrantee(t *testing.T) {
// 	al := setupAccessLogic()
// 	//	ref := "example_ref"
// 	id0, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	testGranteeList := dynamicaccess.NewGrantee()

// 	// Add grantee keys to the testGranteeList
// 	id1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	id2, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	id3, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	testGranteeList.AddGrantees([]ecdsa.PublicKey{id1.PublicKey, id2.PublicKey, id3.PublicKey})

// 	// Initialize empty ACT
// 	actMock := MockAct.NewActMock()
// 	actMockRootHash := "exampleRootHash"

// 	// Add each grantee to content using ActMock and validate the resulting ACT
// 	for i := 0; i < len(testGranteeList.GetGrantees()); i++ {
// 		lookupKey, _ := al.getLookUpKey(testGranteeList.GetGrantees()[i], "")
// 		encryptedAccessKey := "exampleEncryptedAccessKey"
// 		_, err := actMock.Add(actMockRootHash, []byte(lookupKey), []byte(encryptedAccessKey))
// 		if err != nil {
// 			t.Fatalf("Failed to add grantee to content using ActMock: %v", err)
// 		}

// 		// Validate the resulting ACT
// 		encryptedAccessKeyFromMock, err := actMock.Get(actMockRootHash, []byte(lookupKey))
// 		if err != nil {
// 			t.Fatalf("Failed to retrieve encrypted access key from ActMock: %v", err)
// 		}
// 		encryptedAccessKeyFromMockBytes, _ := hex.DecodeString(encryptedAccessKeyFromMock)
// 		if string(encryptedAccessKeyFromMockBytes) != encryptedAccessKey {
// 			t.Errorf("Encrypted access key retrieved from ActMock doesn't match expected value")
// 		}
// 	}

// 	al.Add_New_Grantee_To_Content(actMock, encryptedRef, id0.PublicKey, testGranteeList.GetGrantees()[i])
// }
