package dynamicaccess_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"testing"

	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/dynamicaccess"
	"github.com/ethersphere/bee/v2/pkg/dynamicaccess/mock"
	"github.com/ethersphere/bee/v2/pkg/encryption"
	kvsmock "github.com/ethersphere/bee/v2/pkg/kvs/mock"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"golang.org/x/crypto/sha3"
)

var hashFunc = sha3.NewLegacyKeccak256

func testHistory(t *testing.T) dynamicaccess.History {
	const layout = "2006-Jan-02"
	h := mock.NewHistory()
	h.Insert(10, kvsmock.NewReference(swarm.RandAddress(t)))
	h.Insert(1000, kvsmock.NewReference(swarm.RandAddress(t)))
	kvs := kvsmock.NewReference(swarm.RandAddress(t))

	kvs.Put([]byte("key1"), []byte("value1"))
	h.Insert(100000, kvs)

	return h
}

func TestControllerUploadHandler(t *testing.T) {
	// pk := getPrivateKey()
	// ak := encryption.Key([]byte("cica"))

	// si := dynamicaccess.NewDefaultSession(pk)
	// aek, _ := si.Key(&pk.PublicKey, [][]byte{{1}})
	// e2 := encryption.New(aek[0], 0, uint32(0), hashFunc)
	// _, err := e2.Encrypt(ak)

	// h := NewHistoryReference()
	// h := mock.NewHistory()
	// al := setupAccessLogic2()
	// c := dynamicaccess.NewController(h, al)
	// eref, ref := prepareEncryptedChunkReference(ak)

	// key1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	// err := c.Grant([]*ecdsa.PublicKey{&key1.PublicKey})
	// if err != nil {
	// 	t.Fatalf("gm.Add() returned an error: %v", err)
	// }

	// addr, _ := c.UploadHandler(ref, &pk.PublicKey, swarm.EmptyAddress)
	// if !addr.Equal(eref) {
	// 	t.Fatalf("Encrypted chunk address: %s is not the expected: %s", addr, eref)
	// }
	// _, err := h.Lookup(0)
	// if err != nil {
	// 	t.Fatalf("h.Lookup() returned an error: %v", err)
	// }

	//FIXME
	// storer := mockstorer.New()
	// ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false, 0))
	// m, err := manifest.NewDefaultManifestReference(manifRef, ls)
	// if err != nil {
	// 	t.Fatalf("NewDefaultManifestReference() returned an error: %v", err)
	// }
	// me, err := m.Lookup(context.Background(), hex.EncodeToString(aek))
	// if err != nil {
	// 	t.Fatalf("m.Lookup() returned an error: %v", err)
	// }
	// if me.Reference().String() != addr.String() {
	// 	t.Fatalf("me.Reference(): %s is not the expected: %s", me.Reference(), addr)

	// }
}

func TestControllerGrant(t *testing.T) {
	al := setupAccessLogic2()
	//h := mock.NewHistory()
	c := dynamicaccess.NewController(al)

	gladdr := swarm.RandAddress(t)
	gl := kvsmock.NewReference(gladdr)
	c.Grant(gladdr, getPrivateKey().Public().(*ecdsa.PublicKey))
	gl.Get([]byte("key1"))
}

func TestControllerRevoke(t *testing.T) {

}

func TestControllerCommit(t *testing.T) {

}

func prepareEncryptedChunkReference(ak []byte) (swarm.Address, swarm.Address) {
	addr, _ := hex.DecodeString("f7b1a45b70ee91d3dbfd98a2a692387f24db7279a9c96c447409e9205cf265baef29bf6aa294264762e33f6a18318562c86383dd8bfea2cec14fae08a8039bf3")
	e1 := encryption.New(ak, 0, uint32(0), hashFunc)
	ech, err := e1.Encrypt(addr)
	if err != nil {
		return swarm.EmptyAddress, swarm.EmptyAddress
	}
	return swarm.NewAddress(ech), swarm.NewAddress(addr)
}

func getPrivateKey() *ecdsa.PrivateKey {
	data, _ := hex.DecodeString("c786dd84b61485de12146fd9c4c02d87e8fd95f0542765cb7fc3d2e428c0bcfa")

	privKey, _ := crypto.DecodeSecp256k1PrivateKey(data)
	return privKey
}
