package dynamicaccess

import (
	"testing"

	"github.com/ethersphere/bee/pkg/encryption"
)

func TestGetLookupKey_Success(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	publisher := "examplePublisher"
	tag := "exampleTag"

	lookupKey, err := al.GetLookUpKey(publisher, tag)
	if err != nil {
		t.Errorf("Could not fetch lookup key from publisher and tag")
	}

	expectedLookupKey := "expectedLookupKey"
	if lookupKey != expectedLookupKey {
		t.Errorf("The lookup key that was returned is not correct!")
	}
}

func TestGetLookupKey_Error(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	invalidPublisher := ""
	tag := "exampleTag"

	lookupKey, err := al.GetLookUpKey(invalidPublisher, tag)
	if err != nil {
		t.Errorf("There was an error while fetching lookup key")
	}

	if lookupKey != "" {
		t.Errorf("Expected lookup key to be empty for invalid input")
	}
}

func TestGetAccessKeyDecriptionKey_Success(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	publisher := "examplePublisher"
	tag := "exampleTag"

	access_key_decryption_key, err := al.GetAccessKeyDecriptionKey(publisher, tag)
	if err != nil {
		t.Errorf("GetAccessKeyDecriptionKey gave back error")
	}

	expectedResult := "we-dont-know"
	if access_key_decryption_key != expectedResult {
		t.Errorf("The access key decryption key is not correct!")
	}
}

func TestGetAccessKeyDecriptionKey_Error(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	invalidPublisher := ""
	tag := "exampleTag"

	access_key_decryption_key, err := al.GetAccessKeyDecriptionKey(invalidPublisher, tag)
	if err != nil {
		t.Errorf("GetAccessKeyDecriptionKey gave back error")
	}

	if access_key_decryption_key != "" {
		t.Errorf("GetAccessKeyDecriptionKey should give back empty string for invalid input!")
	}
}

func TestGetEncryptedAccessKey_Success(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	actRootHash := "0xabcexample"
	lookupKey := "exampleLookupKey"

	encrypted_access_key, err := al.GetEncryptedAccessKey(actRootHash, lookupKey)
	if err != nil {
		t.Errorf("There was an error while executing GetEncryptedAccessKey")
	}

	expectedEncryptedKey := "abc013encryptedkey"
	if encrypted_access_key != expectedEncryptedKey {
		t.Errorf("GetEncryptedAccessKey didn't give back the expected value!")
	}
}

func TestGetEncryptedAccessKey_Error(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	actRootHash := "0xabcexample"
	lookupKey := "exampleLookupKey"

	empty_act_result, _ := al.GetEncryptedAccessKey("", lookupKey)
	if empty_act_result != nil {
		t.Errorf("GetEncryptedAccessKey should give back nil for empty act root hash!")
	}

	empty_lookup_result, _ := al.GetEncryptedAccessKey(actRootHash, "")

	if empty_lookup_result != nil {
		t.Errorf("GetEncryptedAccessKey should give back nil for empty lookup key!")
	}
}

func TestGet_Success(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	actRootHash := "0xabcexample"
	encryptedRef := "bzzabcasab"
	publisher := "examplePublisher"
	tag := "exampleTag"

	ref, err := al.Get(actRootHash, encryptedRef, publisher, tag)
	if err != nil {
		t.Errorf("There was an error while calling Get")
	}

	expectedRef := "bzzNotEncrypted128long"
	if ref != expectedRef {
		t.Errorf("Get gave back wrong Swarm reference!")
	}
}

func TestGet_Error(t *testing.T) {
	al := NewAccessLogic(encryption.Key{0}, 4096, uint32(0), hashFunc)

	actRootHash := "0xabcexample"
	encryptedRef := "bzzabcasab"
	publisher := "examplePublisher"
	tag := "exampleTag"

	refOne, _ := al.Get("", encryptedRef, publisher, tag)
	if refOne != "" {
		t.Errorf("Get should give back empty string if ACT root hash not provided!")
	}

	refTwo, _ := al.Get(actRootHash, "", publisher, tag)
	if refTwo != "" {
		t.Errorf("Get should give back empty string if encrypted ref not provided!")
	}

	refThree, _ := al.Get(actRootHash, encryptedRef, "", tag)
	if refThree != "" {
		t.Errorf("Get should give back empty string if publisher not provided!")
	}

	refFour, _ := al.Get(actRootHash, encryptedRef, publisher, "")
	if refFour != "" {
		t.Errorf("Get should give back empty string if tag was not provided!")
	}
}

func TestXxx(t *testing.T) {
	/*var loadSaver file.LoadSaver
	var ctx context.Context
	loadSaver.Load(ctx, []byte())
	testManifest, err := manifest.NewDefaultManifest(loadSaver, false)
	testManifest.Add(ctx, "x", manifest.NewEntry())
	if err != nil {

	}*/
	/*
		//key encryption.Key, padding int, initCtr uint32, hashFunc func() hash.Hash
		al := NewAccessLogic(nil, 0, 0, nil)
		if al == nil {
			t.Errorf("Error creating access logic")
		}
		newObj, err := al.Get("rootKey", "encryped_ref", "publisher", "tag")
		if err != nil {
			println(newObj)
		}*/
}

/*
type simpleLoadSaver struct{}

func (s *simpleLoadSaver) Load(ctx context.Context, address swarm.Address) ([]byte, error) {
	// Implement Load method as needed for testing
	return nil, nil
}

func (s *simpleLoadSaver) Save(ctx context.Context, data []byte) (swarm.Address, error) {
	// Implement Save method as needed for testing
	return swarm.NewAddress([]byte("8cff4f491ad41765012d07290ba08d2f7aa0b2e1314b4ad552319adc9be8f024")), nil
}

type basicManifest struct {
	entries map[string]swarm.Address
}

func NewBasicManifest() *basicManifest {
	return &basicManifest{
		entries: make(map[string]swarm.Address),
	}
}

func TestAccessLogic(t *testing.T) {
	var ctx context.Context
	var loadSaver file.LoadSaver
	testManifest, err := manifest.NewDefaultManifest(loadSaver, false)
	if err != nil {
		t.Errorf("Error creating default manifest: %v", err)
	}

	myMap := make(map[string]string)

	// Add some key-value pairs to the map
	myMap["name"] = "John"
	myMap["age"] = "30"
	myMap["city"] = "New York"

	testManifest.Add(context.Background(), "example/path1", manifest.NewEntry(swarm.NewAddress([]byte{4, 5, 6}), myMap))
	testManifest.Add(context.Background(), "example/path2", manifest.NewEntry(swarm.EmptyAddress, myMap))

	entry := manifest.NewEntry(swarm.NewAddress([]byte{1, 2, 3}), map[string]string{"filename": "example.txt"})
	err = testManifest.Add(ctx, "example/path", entry)
	if err != nil {
		t.Errorf("Error adding entry to manifest: %v", err)
	}

	// Now you can test your AccessLogic with this manifest
	// For example:
	// al := NewAccessLogic(...)
	// newObj, err := al.Get("rootKey", "encrypted_ref", "publisher", "tag")
	// Perform assertions on newObj and err as needed
}

func SomeTest() {
	// Create a new basic manifest
	manifestt := NewBasicManifest()

	// Add some entries
	manifestt.Add(context.Background(), "example/path1", swarm.NewAddress([]byte{1, 2, 3}))
	manifestt.Add(context.Background(), "example/path2", swarm.NewAddress([]byte{4, 5, 6}))

	// Lookup an entry
	addr, err := manifest.Lookup(context.Background(), "example/path1")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Address:", addr)
	}
}
*/
