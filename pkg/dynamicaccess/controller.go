package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/ethersphere/bee/v2/pkg/crypto"
	encryption "github.com/ethersphere/bee/v2/pkg/encryption"
	"github.com/ethersphere/bee/v2/pkg/file/loadsave"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/kvs"
	"github.com/ethersphere/bee/v2/pkg/storage"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

const granteeListEncrypt = true

type GranteeManager interface {
	// TODO: doc
	HandleGrantees(ctx context.Context, getter storage.Getter, putter storage.Putter, granteeref swarm.Address, historyref swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) (swarm.Address, swarm.Address, swarm.Address, swarm.Address, error)
	// GetGrantees returns the list of grantees for the given publisher.
	// The list is accessible only by the publisher.
	GetGrantees(ctx context.Context, getter storage.Getter, publisher *ecdsa.PublicKey, encryptedglref swarm.Address) ([]*ecdsa.PublicKey, error)
}

type Controller interface {
	GranteeManager
	// DownloadHandler decrypts the encryptedRef using the lookupkey based on the history and timestamp.
	DownloadHandler(ctx context.Context, getter storage.Getter, encryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address, timestamp int64) (swarm.Address, error)
	// UploadHandler encrypts the reference and stores it in the history as the latest update.
	UploadHandler(ctx context.Context, getter storage.Getter, putter storage.Putter, reference swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, swarm.Address, swarm.Address, error)
	io.Closer
}

type controller struct {
	accessLogic ActLogic
}

var _ Controller = (*controller)(nil)

func (c *controller) DownloadHandler(
	ctx context.Context,
	getter storage.Getter,
	encryptedRef swarm.Address,
	publisher *ecdsa.PublicKey,
	historyRootHash swarm.Address,
	timestamp int64,
) (swarm.Address, error) {
	ls := loadsave.NewReadonly(getter)
	fmt.Printf("bagoy DownloadHandler\n")
	history, err := NewHistoryReference(ls, historyRootHash)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy historyRootHash: %v\n", historyRootHash)

	// ts := time.Now().Unix()
	// ts := time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC).Unix()
	entry, err := history.Lookup(ctx, timestamp)
	if err != nil {
		fmt.Printf("bagoy Lookup err: %v\n", err)
		return swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy act ref: %v\n", swarm.NewAddress(entry.Reference()))
	fmt.Printf("bagoy act entry: %v\n", swarm.NewAddress(entry.Entry()))
	// mockRefS := "497b4f4a07dd11ba1903b1fb1f975e8dab5c8e84b5570c2da807d073ae8580be"
	// mockRefB, _ := hex.DecodeString(mockRefS)
	// mockRef := swarm.NewAddress(mockRefB)
	// fmt.Printf("bagoy mockRef: %v\n", mockRef)
	act, err := kvs.NewReference(ls, swarm.NewAddress(entry.Entry()))
	if err != nil {
		fmt.Printf("bagoy kvs.newref err: %v\n", err)
		return swarm.ZeroAddress, err
	}

	dref, err := c.accessLogic.DecryptRef(ctx, act, encryptedRef, publisher)
	if err != nil {
		fmt.Printf("bagoy DecryptRef err: %v\n", err)
	} else {
		fmt.Printf("bagoy dref: %v\n", dref)
	}
	return dref, err
}

func (c *controller) UploadHandler(
	ctx context.Context,
	getter storage.Getter,
	putter storage.Putter,
	refrefence swarm.Address,
	publisher *ecdsa.PublicKey,
	historyRootHash swarm.Address,
) (swarm.Address, swarm.Address, swarm.Address, error) {
	ls := loadsave.New(getter, putter, requestPipelineFactory(ctx, putter, false, redundancy.NONE))
	historyRef := historyRootHash
	var (
		act    kvs.KeyValueStore
		actRef swarm.Address
	)
	now := time.Now().Unix()
	if historyRef.IsZero() {
		history, err := NewHistory(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		act, err = kvs.New(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		err = c.accessLogic.AddPublisher(ctx, act, publisher)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		actRef, err = act.Save(ctx)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy new actRef: %v\n", actRef)
		err = history.Add(ctx, actRef, &now, nil)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		historyRef, err = history.Store(ctx)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy new historyRef: %v\n", historyRef)
	} else {
		history, err := NewHistoryReference(ls, historyRef)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy historyRef: %v\n", historyRef)
		entry, err := history.Lookup(ctx, now)
		actRef = swarm.NewAddress(entry.Entry())
		fmt.Printf("bagoy actRef: %v\n", actRef)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		act, err = kvs.NewReference(ls, actRef)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		// actRefe := swarm.NewAddress(entry.Entry())
		// fmt.Printf("bagoy actRefe: %v\n", actRefe)
		// _, err = kvs.NewReference(ls, actRefe)
		// if err != nil {
		// 	return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		// }
	}

	encryptedRef, err := c.accessLogic.EncryptRef(ctx, act, publisher, refrefence)
	return actRef, historyRef, encryptedRef, err
}

func NewController(accessLogic ActLogic) Controller {
	return &controller{
		accessLogic: accessLogic,
	}
}

func (c *controller) HandleGrantees(
	ctx context.Context,
	getter storage.Getter,
	putter storage.Putter,
	encryptedglref swarm.Address,
	historyref swarm.Address,
	publisher *ecdsa.PublicKey,
	addList []*ecdsa.PublicKey,
	removeList []*ecdsa.PublicKey,
) (swarm.Address, swarm.Address, swarm.Address, swarm.Address, error) {
	var (
		err        error
		h          History
		act        kvs.KeyValueStore
		granteeref swarm.Address
		ls         = loadsave.New(getter, putter, requestPipelineFactory(ctx, putter, false, redundancy.NONE))
		gls        = loadsave.New(getter, putter, requestPipelineFactory(ctx, putter, granteeListEncrypt, redundancy.NONE))
	)
	if !historyref.IsZero() {
		h, err = NewHistoryReference(ls, historyref)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy href: %v\n", historyref)
		entry, err := h.Lookup(ctx, time.Now().Unix())
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		// actref := entry.Reference()
		actref := entry.Entry()
		// fmt.Printf("bagoy lookup actref: %v\n", actref)
		fmt.Printf("bagoy lookup actref: %s\n", swarm.NewAddress(actref))
		act, err = kvs.NewReference(ls, swarm.NewAddress(actref))
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy after kvs.NewReference\n")
	} else {
		h, err = NewHistory(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy after NewHistory\n")
		act, err = kvs.New(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy after kvs.New\n")
	}

	var gl GranteeList
	if encryptedglref.IsZero() {
		gl, err = NewGranteeList(gls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	} else {
		fmt.Printf("bagoy before decryptRefForPublisher\n")
		granteeref, err = c.decryptRefForPublisher(publisher, encryptedglref)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy granteeref: %v\n", granteeref)

		gl, err = NewGranteeListReference(gls, granteeref)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy before NewGranteeListReference\n")
	}
	err = gl.Add(addList)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy after gl.Add\n")
	if len(removeList) != 0 {
		err = gl.Remove(removeList)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	}

	var granteesToAdd []*ecdsa.PublicKey
	// generate new access key and new act
	if len(removeList) != 0 || encryptedglref.IsZero() {
		if historyref.IsZero() {
			err = c.accessLogic.AddPublisher(ctx, act, publisher)
			if err != nil {
				return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
			}
			fmt.Printf("bagoy after AddPublisher\n")
		}
		granteesToAdd = gl.Get()
	} else {
		granteesToAdd = addList
	}

	for i, grantee := range granteesToAdd {
		fmt.Printf("bagoy before AddGrantee\n")
		err := c.accessLogic.AddGrantee(ctx, act, publisher, grantee, nil)
		fmt.Printf("bagoy granteeref: %s\n", hex.EncodeToString(crypto.EncodeSecp256k1PublicKey(grantee)))
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		fmt.Printf("bagoy after AddGrantee ix: %d\n", i)
	}

	actref, err := act.Save(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy saved actref: %v\n", actref)

	glref, err := gl.Save(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy saved glref: %v\n", glref)

	eglref, err := c.encryptRefForPublisher(publisher, glref)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy saved eglref: %v\n", eglref)

	// mtdt := map[string]string{"encryptedglref": eglref.String()}
	// err = h.Add(ctx, actref, nil, &mtdt)
	err = h.Add(ctx, actref, nil, nil)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	href, err := h.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	fmt.Printf("bagoy saved href: %v\n", href)

	return glref, eglref, href, actref, nil
}

func (c *controller) GetGrantees(ctx context.Context, getter storage.Getter, publisher *ecdsa.PublicKey, encryptedglref swarm.Address) ([]*ecdsa.PublicKey, error) {
	ls := loadsave.NewReadonly(getter)
	granteeRef, err := c.decryptRefForPublisher(publisher, encryptedglref)
	if err != nil {
		return nil, err
	}
	gl, err := NewGranteeListReference(ls, granteeRef)
	if err != nil {
		return nil, err
	}
	return gl.Get(), nil
}

func (c *controller) encryptRefForPublisher(publisherPubKey *ecdsa.PublicKey, ref swarm.Address) (swarm.Address, error) {
	keys, err := c.accessLogic.Session.Key(publisherPubKey, [][]byte{oneByteArray})
	if err != nil {
		return swarm.ZeroAddress, err
	}
	refCipher := encryption.New(keys[0], 0, uint32(0), hashFunc)
	encryptedRef, err := refCipher.Encrypt(ref.Bytes())
	if err != nil {
		return swarm.ZeroAddress, err
	}

	return swarm.NewAddress(encryptedRef), nil
}

func (c *controller) decryptRefForPublisher(publisherPubKey *ecdsa.PublicKey, encryptedRef swarm.Address) (swarm.Address, error) {
	keys, err := c.accessLogic.Session.Key(publisherPubKey, [][]byte{oneByteArray})
	if err != nil {
		return swarm.ZeroAddress, err
	}
	refCipher := encryption.New(keys[0], 0, uint32(0), hashFunc)
	ref, err := refCipher.Decrypt(encryptedRef.Bytes())
	if err != nil {
		return swarm.ZeroAddress, err
	}

	return swarm.NewAddress(ref), nil
}

func requestPipelineFactory(ctx context.Context, s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(ctx, s, encrypt, rLevel)
	}
}

// TODO: what to do in close ?
func (s *controller) Close() error {
	return nil
}
