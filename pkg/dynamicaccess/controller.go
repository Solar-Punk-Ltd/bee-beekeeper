package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"io"
	"time"

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
	//PATCH /grantees
	//{publisher, addList, removeList}
	HandleGrantees(ctx context.Context, granteeref swarm.Address, historyref swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) (swarm.Address, swarm.Address, error)

	//GET /grantees/{history root hash}
	GetGrantees(ctx context.Context, rootHash swarm.Address) ([]*ecdsa.PublicKey, error)
}

// TODO: add granteeList ref to history metadata to solve inconsistency
type Controller interface {
	GranteeManager
	// DownloadHandler decrypts the encryptedRef using the lookupkey based on the history and timestamp.
	DownloadHandler(ctx context.Context, encryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address, timestamp int64) (swarm.Address, error)
	// TODO: history encryption
	// UploadHandler encrypts the reference and stores it in the history as the latest update.
	UploadHandler(ctx context.Context, reference swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, swarm.Address, swarm.Address, error)
	io.Closer
}

type controller struct {
	accessLogic ActLogic
	getter      storage.Getter
	putter      storage.Putter
}

var _ Controller = (*controller)(nil)

func (c *controller) DownloadHandler(
	ctx context.Context,
	encryptedRef swarm.Address,
	publisher *ecdsa.PublicKey,
	historyRootHash swarm.Address,
	timestamp int64,
) (swarm.Address, error) {
	ls := loadsave.New(c.getter, c.putter, requestPipelineFactory(ctx, c.putter, false, redundancy.NONE))
	history, err := NewHistoryReference(ls, historyRootHash)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	entry, err := history.Lookup(ctx, timestamp)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	// TODO: hanlde granteelist ref in mtdt
	kvs, err := kvs.NewReference(ls, entry.Reference())
	if err != nil {
		return swarm.ZeroAddress, err
	}

	return c.accessLogic.DecryptRef(ctx, kvs, encryptedRef, publisher)
}

func (c *controller) UploadHandler(
	ctx context.Context,
	refrefence swarm.Address,
	publisher *ecdsa.PublicKey,
	historyRootHash swarm.Address,
) (swarm.Address, swarm.Address, swarm.Address, error) {
	ls := loadsave.New(c.getter, c.putter, requestPipelineFactory(ctx, c.putter, false, redundancy.NONE))
	historyRef := historyRootHash
	var (
		storage kvs.KeyValueStore
		actRef  swarm.Address
	)
	now := time.Now().Unix()
	if historyRef.IsZero() {
		history, err := NewHistory(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		storage, err = kvs.New(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		err = c.accessLogic.AddPublisher(ctx, storage, publisher)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		actRef, err = storage.Save(ctx)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		// TODO: pass granteelist ref as mtdt
		err = history.Add(ctx, actRef, &now, nil)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		historyRef, err = history.Store(ctx)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	} else {
		history, err := NewHistoryReference(ls, historyRef)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		// TODO: hanlde granteelist ref in mtdt
		entry, err := history.Lookup(ctx, now)
		actRef = entry.Reference()
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		storage, err = kvs.NewReference(ls, actRef)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	}

	encryptedRef, err := c.accessLogic.EncryptRef(ctx, storage, publisher, refrefence)
	return actRef, historyRef, encryptedRef, err
}

func NewController(ctx context.Context, accessLogic ActLogic, getter storage.Getter, putter storage.Putter) Controller {
	return &controller{
		accessLogic: accessLogic,
		getter:      getter,
		putter:      putter,
	}
}

func (c *controller) HandleGrantees(ctx context.Context, granteeref swarm.Address, historyref swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) (swarm.Address, swarm.Address, error) {
	var (
		err error
		h   History
		act kvs.KeyValueStore
		ls  = loadsave.New(c.getter, c.putter, requestPipelineFactory(ctx, c.putter, false, redundancy.NONE))
		gls = loadsave.New(c.getter, c.putter, requestPipelineFactory(ctx, c.putter, granteeListEncrypt, redundancy.NONE))
	)
	if !historyref.IsZero() {
		h, err = NewHistoryReference(ls, historyref)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		entry, err := h.Lookup(ctx, time.Now().Unix())

		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		actref := entry.Reference()
		act, err = kvs.NewReference(ls, actref)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	} else {
		h, err = NewHistory(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		act, err = kvs.New(ls)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	}

	var gl GranteeList
	if granteeref.IsZero() {
		gl = NewGranteeList(gls)
	} else {
		gl = NewGranteeListReference(gls, granteeref)
	}
	gl.Add(addList)
	gl.Remove(removeList)

	granteesToAdd := addList

	// generate new access key and new act
	if len(removeList) != 0 || granteeref.IsZero() {
		c.accessLogic.AddPublisher(ctx, act, publisher)
		granteesToAdd = gl.Get()
	}

	for _, grantee := range granteesToAdd {
		c.accessLogic.AddGrantee(ctx, act, publisher, grantee, nil)
	}

	actref, err := act.Save(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}

	h.Add(ctx, actref, nil, nil)
	href, err := h.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}

	glref, err := gl.Save(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	return glref, href, nil
}

func (c *controller) GetGrantees(ctx context.Context, granteeRef swarm.Address) ([]*ecdsa.PublicKey, error) {
	ls := loadsave.New(c.getter, c.putter, requestPipelineFactory(ctx, c.putter, granteeListEncrypt, redundancy.NONE))
	gl := NewGranteeListReference(ls, granteeRef)
	return gl.Get(), nil
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
