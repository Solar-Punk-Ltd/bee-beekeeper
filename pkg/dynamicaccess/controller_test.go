package dynamicaccess_test

import (
	"context"
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/ethersphere/bee/v2/pkg/dynamicaccess"
	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/kvs"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/stretchr/testify/assert"
)

func getHistoryFixture(ctx context.Context, ls file.LoadSaver, al dynamicaccess.ActLogic, publisher *ecdsa.PublicKey) (swarm.Address, error) {
	h, err := dynamicaccess.NewHistory(ls)
	if err != nil {
		return swarm.ZeroAddress, nil
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
	c := dynamicaccess.NewController(ctx, al)
	ref := swarm.RandAddress(t)
	_, hRef, encRef, err := c.UploadHandler(ctx, mockStorer.ChunkStore(), mockStorer.Cache(), ref, &publisher.PublicKey, swarm.ZeroAddress)

	ls := createLs()
	h, err := dynamicaccess.NewHistoryReference(ls, hRef)
	entry, err := h.Lookup(ctx, time.Now().Unix())
	actRef := entry.Reference()
	act, err := kvs.NewReference(ls, actRef)
	expRef, err := al.EncryptRef(ctx, act, &publisher.PublicKey, ref)

	assert.NoError(t, err)
	assert.Equal(t, encRef, expRef)
	assert.NotEqual(t, hRef, swarm.ZeroAddress)
}

func TestController_PublisherDownload(t *testing.T) {
	ctx := context.Background()
	publisher := getPrivKey(0)
	diffieHellman := dynamicaccess.NewDefaultSession(publisher)
	al := dynamicaccess.NewLogic(diffieHellman)
	c := dynamicaccess.NewController(ctx, al)
	ls := createLs()
	ref := swarm.RandAddress(t)
	href, err := getHistoryFixture(ctx, ls, al, &publisher.PublicKey)
	h, err := dynamicaccess.NewHistoryReference(ls, href)
	entry, err := h.Lookup(ctx, time.Now().Unix())
	actRef := entry.Reference()
	act, err := kvs.NewReference(ls, actRef)
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
	c := dynamicaccess.NewController(ctx, al)
	ref := swarm.RandAddress(t)
	href, err := getHistoryFixture(ctx, ls, publisherAL, &publisher.PublicKey)
	h, err := dynamicaccess.NewHistoryReference(ls, href)
	ts := time.Date(2001, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err := h.Lookup(ctx, ts)
	actRef := entry.Reference()
	act, err := kvs.NewReference(ls, actRef)
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
	ls := createLs()
	getter := mockStorer.ChunkStore()
	putter := mockStorer.Cache()
	c := dynamicaccess.NewController(ctx, al)
	href, _ := getHistoryFixture(ctx, ls, al, &publisher.PublicKey)

	grantee1 := getPrivKey(0)
	grantee := getPrivKey(2)

	t.Run("add to new list", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, swarm.ZeroAddress, &publisher.PublicKey, addList, nil)

		gl := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)
	})
	t.Run("add to existing list", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, href, &publisher.PublicKey, addList, nil)

		gl := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)

		addList = []*ecdsa.PublicKey{&getPrivKey(0).PublicKey}
		granteeRef, _, err = c.HandleGrantees(ctx, getter, putter, granteeRef, href, &publisher.PublicKey, addList, nil)
		gl = dynamicaccess.NewGranteeListReference(createLs(), granteeRef)
		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 2)
	})
	t.Run("add and revoke", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		revokeList := []*ecdsa.PublicKey{&grantee1.PublicKey}
		gl := dynamicaccess.NewGranteeList(createLs())
		gl.Add([]*ecdsa.PublicKey{&publisher.PublicKey, &grantee1.PublicKey})
		granteeRef, err := gl.Save(ctx)

		granteeRef, _, err = c.HandleGrantees(ctx, getter, putter, granteeRef, href, &publisher.PublicKey, addList, revokeList)
		gl = dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 2)
	})

	t.Run("add twice", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey, &grantee.PublicKey}
		granteeRef, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, href, &publisher.PublicKey, addList, nil)
		granteeRef, _, err = c.HandleGrantees(ctx, getter, putter, granteeRef, href, &publisher.PublicKey, addList, nil)
		gl := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)
	})
	t.Run("revoke non-existing", func(t *testing.T) {
		addList := []*ecdsa.PublicKey{&grantee.PublicKey}
		granteeRef, _, err := c.HandleGrantees(ctx, getter, putter, swarm.ZeroAddress, href, &publisher.PublicKey, addList, nil)
		gl := dynamicaccess.NewGranteeListReference(createLs(), granteeRef)

		assert.NoError(t, err)
		assert.Len(t, gl.Get(), 1)
	})
}
