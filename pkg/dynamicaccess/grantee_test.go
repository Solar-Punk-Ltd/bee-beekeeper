package dynamicaccess_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/file/loadsave"
	"github.com/ethersphere/bee/pkg/file/pipeline"
	"github.com/ethersphere/bee/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/pkg/file/redundancy"
	"github.com/ethersphere/bee/pkg/storage"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

var mockStorer = mockstorer.New()

func requestPipelineFactory(ctx context.Context, s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(ctx, s, encrypt, rLevel)
	}
}

func createLs() file.LoadSaver {
	return loadsave.New(mockStorer.ChunkStore(), mockStorer.Cache(), requestPipelineFactory(context.Background(), mockStorer.Cache(), false, redundancy.NONE))
}

// func TestGranteeKeySerialization(t *testing.T) {
// 	putter := mockStorer.DirectUpload()
// 	grantee := dynamicaccess.NewGranteeList(createLs(), putter, swarm.ZeroAddress)

// 	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	serializedPub := grantee.SerializePublicKey(&key1.PublicKey)
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	deSerializedPub := grantee.DeserializeBytes(serializedPub)
// 	if !key1.PublicKey.Equal(deSerializedPub) {
// 		t.Errorf("Expected key1.PublicKey.X %v, got %v", &key1.PublicKey.X, &deSerializedPub.X)
// 	}
// }

// func TestGranteeListSerialization(t *testing.T) {
// 	putter := mockStorer.DirectUpload()
// 	grantee := dynamicaccess.NewGranteeList(createLs(), putter, swarm.ZeroAddress)

// 	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}

// 	gl := []*ecdsa.PublicKey{&key1.PublicKey, &key2.PublicKey}

// 	serializedPubs := grantee.Serialize(gl)

// 	deSerializedPubs := grantee.Deserialize(serializedPubs)
// 	if !key1.PublicKey.Equal(deSerializedPubs[0]) {
// 		t.Errorf("Expected key1.PublicKey.X %v, got %v", &key1.PublicKey.X, &deSerializedPubs[0].X)
// 	}
// 	if !key2.PublicKey.Equal(deSerializedPubs[1]) {
// 		t.Errorf("Expected key1.PublicKey.X %v, got %v", &key1.PublicKey.X, &deSerializedPubs[1].X)
// 	}
// }

func TestGranteeAddGrantees(t *testing.T) {
	putter := mockStorer.DirectUpload()
	grantee := dynamicaccess.NewGranteeList(createLs(), putter, swarm.ZeroAddress)

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []*ecdsa.PublicKey{&key1.PublicKey, &key2.PublicKey}
	err = grantee.Add(addList)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	grantees := grantee.Get()
	if !reflect.DeepEqual(grantees, addList) {
		t.Errorf("Expected grantees %v, got %v", addList, grantees)
	}
}

// TODO: fix index out of range error
func TestRemoveGrantees(t *testing.T) {
	putter := mockStorer.DirectUpload()
	grantee := dynamicaccess.NewGranteeList(createLs(), putter, swarm.ZeroAddress)

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []*ecdsa.PublicKey{&key1.PublicKey, &key2.PublicKey}
	err = grantee.Add(addList)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	removeList := []*ecdsa.PublicKey{&key1.PublicKey}
	err = grantee.Remove(removeList)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	grantees := grantee.Get()
	expectedGrantees := []*ecdsa.PublicKey{&key2.PublicKey}

	for i, grantee := range grantees {
		if !grantee.Equal(expectedGrantees[i]) {
			t.Errorf("Expected grantee %v, got %v", &expectedGrantees[i].X, &grantee.X)
		}
	}
}

func TestGetGrantees(t *testing.T) {
	putter := mockStorer.DirectUpload()
	grantee := dynamicaccess.NewGranteeList(createLs(), putter, swarm.ZeroAddress)

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []*ecdsa.PublicKey{&key1.PublicKey, &key2.PublicKey}
	err = grantee.Add(addList)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	grantees := grantee.Get()
	for i, grantee := range grantees {
		if !grantee.Equal(addList[i]) {
			t.Errorf("Expected grantee %v, got %v", &addList[i].X, &grantee.X)
		}
	}
}

func TestGranteeSave(t *testing.T) {
	putter := mockStorer.DirectUpload()
	grantee1 := dynamicaccess.NewGranteeList(createLs(), putter, swarm.ZeroAddress)

	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addList := []*ecdsa.PublicKey{&key1.PublicKey, &key2.PublicKey}
	err = grantee1.Add(addList)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addr, err := grantee1.Save()
	if err != nil {
		t.Errorf("Save expected no error, got %v", err)
	}

	grantee2 := dynamicaccess.NewGranteeList(createLs(), putter, addr)

	grantees2 := grantee2.Get()
	if !reflect.DeepEqual(grantees2, addList) {
		t.Errorf("Expected grantees %v, got %v", addList, grantees2)
	}
}
