package dynamicaccess_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/dynamicaccess"
	encryption "github.com/ethersphere/bee/v2/pkg/encryption"
	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/kvs"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/sha3"
)

func getHistoryFixture(ctx context.Context, ls file.LoadSaver, al dynamicaccess.ActLogic, publisher *ecdsa.PublicKey) (swarm.Address, error) {
	h, err := dynamicaccess.NewHistory(ls)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	pk1 := getPrivKey(1)
	pk2 := getPrivKey(2)

	kvs0, _ := kvs.New(ls)
	al.AddPublisher(ctx, kvs0, publisher)
	kvs0Ref, _ := kvs0.Save(ctx)
	kvs1, _ := kvs.New(ls)
	al.AddPublisher(ctx, kvs1, publisher)
	al.AddGrantee(ctx, kvs1, publisher, &pk1.PublicKey, nil)
	kvs1Ref, _ := kvs1.Save(ctx)
	kvs2, _ := kvs.New(ls)
	al.AddPublisher(ctx, kvs2, publisher)
	al.AddGrantee(ctx, kvs2, publisher, &pk2.PublicKey, nil)
	kvs2Ref, _ := kvs2.Save(ctx)
	firstTime := time.Date(1994, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	secondTime := time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	thirdTime := time.Date(2015, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()

	fmt.Printf("bagoy kvs0Ref: %v\n", kvs0Ref)
	fmt.Printf("bagoy kvs1Ref: %v\n", kvs1Ref)
	fmt.Printf("bagoy kvs2Ref: %v\n", kvs2Ref)

	h.Add(ctx, kvs0Ref, &thirdTime, nil)
	h.Add(ctx, kvs1Ref, &firstTime, nil)
	h.Add(ctx, kvs2Ref, &secondTime, nil)
	return h.Store(ctx)
}

func TestController_NewUpload(t *testing.T) {
	ctx := context.Background()
	publisher := getPrivKey(0)
	diffieHellman := dynamicaccess.NewDefaultSession(publisher)
	al := dynamicaccess.NewLogic(diffieHellman)
	c := dynamicaccess.NewController(al)
	ref := swarm.RandAddress(t)
	actRef, hRef, encRef, err := c.UploadHandler(ctx, mockStorer.ChunkStore(), mockStorer.Cache(), ref, &publisher.PublicKey, swarm.ZeroAddress)
	assert.NoError(t, err)

	ls := createLs()
	h, err := dynamicaccess.NewHistoryReference(ls, hRef)
	assert.NoError(t, err)
	entry, err := h.Lookup(ctx, time.Now().Unix())
	assert.NoError(t, err)
	assert.True(t, actRef.Equal(swarm.NewAddress(entry.Entry())))
	act, err := kvs.NewReference(ls, swarm.NewAddress(entry.Entry()))
	assert.NoError(t, err)
	expRef, err := al.EncryptRef(ctx, act, &publisher.PublicKey, ref)

	assert.NoError(t, err)
	assert.Equal(t, expRef, encRef)
	assert.NotEqual(t, hRef, swarm.ZeroAddress)
}

// TODO: same test suite with different test runs
func TestController_ConsecutiveUploads(t *testing.T) {
	ctx := context.Background()
	ls := createLs()
	publisher := getPrivKey(0)
	diffieHellman := dynamicaccess.NewDefaultSession(publisher)
	al := dynamicaccess.NewLogic(diffieHellman)
	c := dynamicaccess.NewController(al)
	ref1 := swarm.RandAddress(t)

	actRef1, hRef1, encRef1, err := c.UploadHandler(ctx, mockStorer.ChunkStore(), mockStorer.Cache(), ref1, &publisher.PublicKey, swarm.ZeroAddress)
	assert.NoError(t, err)
	h1, err := dynamicaccess.NewHistoryReference(ls, hRef1)
	assert.NoError(t, err)
	entry1, err := h1.Lookup(ctx, time.Now().Unix())
	assert.NoError(t, err)
	assert.True(t, actRef1.Equal(swarm.NewAddress(entry1.Entry())))

	act1, err := kvs.NewReference(ls, swarm.NewAddress(entry1.Entry()))
	assert.NoError(t, err)
	expRef1, err := al.EncryptRef(ctx, act1, &publisher.PublicKey, ref1)
	assert.NoError(t, err)
	assert.Equal(t, expRef1, encRef1)
	hRef1Stored, err := h1.Store(ctx)
	assert.NoError(t, err)
	assert.True(t, hRef1.Equal((hRef1Stored)))
	actRef1Stored, err := act1.Save(ctx)
	assert.NoError(t, err)
	assert.True(t, actRef1.Equal((actRef1Stored)))

	// second upload to the same act and history but with different reference
	ref2 := swarm.RandAddress(t)
	actRef2, hRef2, encRef2, err := c.UploadHandler(ctx, mockStorer.ChunkStore(), mockStorer.Cache(), ref2, &publisher.PublicKey, hRef1)
	assert.NoError(t, err)
	h2, err := dynamicaccess.NewHistoryReference(ls, hRef2)
	assert.NoError(t, err)
	entry2, err := h2.Lookup(ctx, time.Now().Unix())
	assert.NoError(t, err)
	act2, err := kvs.NewReference(ls, swarm.NewAddress(entry2.Entry()))
	assert.NoError(t, err)
	assert.True(t, actRef2.Equal(swarm.NewAddress(entry2.Entry())))

	hRef2Stored, err := h2.Store(ctx)
	assert.NoError(t, err)
	assert.True(t, hRef2.Equal((hRef2Stored)))
	actRef2Stored, err := act2.Save(ctx)
	assert.NoError(t, err)
	t.Logf("actRef2Stored: %v\n", actRef2Stored)
	assert.True(t, swarm.NewAddress(entry2.Entry()).Equal(actRef2Stored))

	expRef2, err := al.EncryptRef(ctx, act2, &publisher.PublicKey, ref2)
	assert.NoError(t, err)
	assert.Equal(t, expRef2, encRef2)
}

func TestController_PublisherDownload(t *testing.T) {
	ctx := context.Background()
	publisher := getPrivKey(0)
	diffieHellman := dynamicaccess.NewDefaultSession(publisher)
	al := dynamicaccess.NewLogic(diffieHellman)
	c := dynamicaccess.NewController(al)
	ls := createLs()
	ref := swarm.RandAddress(t)
	href, err := getHistoryFixture(ctx, ls, al, &publisher.PublicKey)
	assert.NoError(t, err)
	h, err := dynamicaccess.NewHistoryReference(ls, href)
	assert.NoError(t, err)
	entry, err := h.Lookup(ctx, time.Now().Unix())
	assert.NoError(t, err)
	actRef := entry.Reference()
	t.Logf("actRef: %v\n", swarm.NewAddress(actRef))
	// act, err := kvs.NewReference(ls, swarm.NewAddress(actRef))
	actRefe := entry.Entry()
	t.Logf("actRefe: %v\n", swarm.NewAddress(actRefe))
	act, err := kvs.NewReference(ls, swarm.NewAddress(actRefe))
	assert.NoError(t, err)
	encRef, err := al.EncryptRef(ctx, act, &publisher.PublicKey, ref)
	assert.NoError(t, err)
	dref, err := c.DownloadHandler(ctx, mockStorer.ChunkStore(), encRef, &publisher.PublicKey, href, time.Now().Unix())
	assert.NoError(t, err)
	assert.Equal(t, ref, dref)
}

func TestController_GranteeDownload(t *testing.T) {
	ctx := context.Background()
	publisher := getPrivKey(0)
	grantee := getPrivKey(2)
	publisherDH := dynamicaccess.NewDefaultSession(publisher)
	publisherAL := dynamicaccess.NewLogic(publisherDH)

	diffieHellman := dynamicaccess.NewDefaultSession(grantee)
	al := dynamicaccess.NewLogic(diffieHellman)
	ls := createLs()
	c := dynamicaccess.NewController(al)
	ref := swarm.RandAddress(t)
	href, err := getHistoryFixture(ctx, ls, publisherAL, &publisher.PublicKey)
	h, err := dynamicaccess.NewHistoryReference(ls, href)
	ts := time.Date(2001, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err := h.Lookup(ctx, ts)
	// actRef := entry.Reference()
	actRefe := entry.Entry()
	act, err := kvs.NewReference(ls, swarm.NewAddress(actRefe))
	encRef, err := publisherAL.EncryptRef(ctx, act, &publisher.PublicKey, ref)

	assert.NoError(t, err)
	dref, err := c.DownloadHandler(ctx, mockStorer.ChunkStore(), encRef, &publisher.PublicKey, href, ts)
	assert.NoError(t, err)
	assert.Equal(t, ref, dref)
}

func TestController_HandleGrantees(t *testing.T) {
	ctx := context.Background()
	publisher := getPrivKey(1)
	diffieHellman := dynamicaccess.NewDefaultSession(publisher)
	al := dynamicaccess.NewLogic(diffieHellman)
	keys, _ := al.Session.Key(&publisher.PublicKey, [][]byte{{1}})
	refCipher := encryption.New(keys[0], 0, uint32(0), sha3.NewLegacyKeccak256)
	ls := createLs()
	getter := mockStorer.ChunkStore()
	putter := mockStorer.Cache()
	c := dynamicaccess.NewController(al)
	href, _ := getHistoryFixture(ctx, ls, al, &publisher.PublicKey)

	grantee1 := getPrivKey(0)
	grantee := getPrivKey(2)

	t.Run("add to new list", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, _, _, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, swarm.ZeroAddress, &publisher.PublicKey, addList, nil)

		gl, err := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)
	})
	t.Run("add to existing list", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, eglref, _, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, href, &publisher.PublicKey, addList, nil)

		gl, err := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)

		addList = []*ecdsa.PublicKey{&getPrivKey(0).PublicKey}
		granteeRef, _, _, _, err = c.HandleGrantees(ctx, getter, putter, eglref, href, &publisher.PublicKey, addList, nil)
		gl, err = dynamicaccess.NewGranteeListReference(createLs(), granteeRef)
		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 2)
	})
	t.Run("add and revoke", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		revokeList := []*ecdsa.PublicKey{&grantee1.PublicKey}
		gl, _ := dynamicaccess.NewGranteeList(createLs())
		gl.Add([]*ecdsa.PublicKey{&publisher.PublicKey, &grantee1.PublicKey})
		granteeRef, err := gl.Save(ctx)
		eglref, _ := refCipher.Encrypt(granteeRef.Bytes())

		granteeRef, _, _, _, err = c.HandleGrantees(ctx, getter, putter, swarm.NewAddress(eglref), href, &publisher.PublicKey, addList, revokeList)
		gl, err = dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 2)
	})

	t.Run("add twice", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey, &grantee.PublicKey}
		granteeRef, eglref, _, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, href, &publisher.PublicKey, addList, nil)
		granteeRef, _, _, _, err = c.HandleGrantees(ctx, getter, putter, eglref, href, &publisher.PublicKey, addList, nil)
		gl, err := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)
	})
	t.Run("revoke non-existing", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, _, _, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, href, &publisher.PublicKey, addList, nil)
		gl, err := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)
	})
}

func TestController_GetGrantees(t *testing.T) {
	ctx := context.Background()
	publisher := getPrivKey(1)
	caller := getPrivKey(0)
	grantee := getPrivKey(2)
	diffieHellman1 := dynamicaccess.NewDefaultSession(publisher)
	diffieHellman2 := dynamicaccess.NewDefaultSession(caller)
	al1 := dynamicaccess.NewLogic(diffieHellman1)
	al2 := dynamicaccess.NewLogic(diffieHellman2)
	ls := createLs()
	getter := mockStorer.ChunkStore()
	putter := mockStorer.Cache()
	c1 := dynamicaccess.NewController(al1)
	c2 := dynamicaccess.NewController(al2)

	t.Run("get by publisher", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, eglRef, _, _, err := c1.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, swarm.ZeroAddress, &publisher.PublicKey, addList, nil)

		grantees, err := c1.GetGrantees(ctx, getter, &publisher.PublicKey, eglRef)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(grantees, addList))

		gl, _ := dynamicaccess.NewGranteeListReference(ls, granteeRef)
		assert.True(t, reflect.DeepEqual(gl.Get(), addList))
	})
	t.Run("get by non-publisher", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		_, eglRef, _, _, err := c1.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, swarm.ZeroAddress, &publisher.PublicKey, addList, nil)
		grantees, err := c2.GetGrantees(ctx, getter, &publisher.PublicKey, eglRef)
		assert.Error(t, err)
		assert.Nil(t, grantees)
	})
}
