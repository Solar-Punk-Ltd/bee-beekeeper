package dynamicaccess_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/swarm"
)

func setupAccessLogic() dynamicaccess.AccessLogic {
	//privateKey, err := crypto.GenerateSecp256k1Key()
	privateKey := generateFixPrivateKey(1000)
	// if err != nil {
	// 	errors.New("error creating private key")
	// 	fmt.Println(err)
	// }
	diffieHellman := dynamicaccess.NewDiffieHellman(&privateKey)
	al := dynamicaccess.NewAccessLogic(diffieHellman)

	return al
}

func generateFixPrivateKey(input int64) ecdsa.PrivateKey {
	fixedD := big.NewInt(input)
	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(fixedD.Bytes())

	privateKey := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: fixedD,
	}

	return privateKey
}

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
	id0 := generateFixPrivateKey(0)

	act := dynamicaccess.NewDefaultAct()
	act, _ = al.AddPublisher(act, id0.PublicKey, "")
	expectedRef := "39a5ea87b141fe44aa609c3327ecd896c0e2122897f5f4bbacf74db1033c5559"

	encryptedRef, _ := al.EncryptRef(act, id0.PublicKey, swarm.NewAddress([]byte(expectedRef)))
	tag := "exampleTag"

	_, err := al.Get(dynamicaccess.NewDefaultAct(), encryptedRef, id0.PublicKey, tag)
	if err == nil {
		t.Errorf("Get should give back encrypted access key not found error!")
	}

	refTwo, _ := al.Get(act, swarm.EmptyAddress, id0.PublicKey, tag)
	if refTwo != "" {
		t.Errorf("Get should give back empty string if encrypted ref not provided!")
	}

	_, err = al.Get(act, encryptedRef, ecdsa.PublicKey{}, tag)
	if err == nil {
		t.Errorf("Get should give back error if grantee not provided!")
	}

	// refFour, _ := al.Get(act, encryptedRef, id0.PublicKey, "")
	// if refFour != "" {
	// 	t.Errorf("Get should give back empty string if tag was not provided!")
	// }
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
	id2 := generateFixPrivateKey(2)
	publisherLookupKey := "bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
	firstAddedGranteeLookupKey := "55cb0b2ee9a2028df92dcb61e152e407f37cca141c8d9d25feacbc6691290895"
	secondAddedGranteeLookupKey := "6ca9afe0f4d0a4b05806a013fa6d98939d1578b5dc2c93a6d559996c2d38c1dc"

	act := dynamicaccess.NewDefaultAct()
	act, _ = al.AddPublisher(act, id0.PublicKey, "")

	act, _ = al.Add_New_Grantee_To_Content(act, id0.PublicKey, id1.PublicKey)
	act, _ = al.Add_New_Grantee_To_Content(act, id0.PublicKey, id2.PublicKey)

	// lookup key is changing, altough it should change at all, input public key does not change
	lookupKeyAsByte, _ := hex.DecodeString(publisherLookupKey)
	actFirstResult := act.Get(lookupKeyAsByte)
	fmt.Println("encrypted access key for publisher: ", actFirstResult)

	lookupKeyAsByte, _ = hex.DecodeString(firstAddedGranteeLookupKey)
	actSecondResult := act.Get(lookupKeyAsByte)
	fmt.Println("encrypted access key for publisher: ", actSecondResult)

	lookupKeyAsByte, _ = hex.DecodeString(secondAddedGranteeLookupKey)
	actThirdResult := act.Get(lookupKeyAsByte)
	fmt.Println("encrypted access key for publisher: ", actThirdResult)

	if 2 == 2 {
		t.Errorf("act result %s", hex.EncodeToString(actFirstResult))
	}
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
