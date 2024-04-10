package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/file/loadsave"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/kvs"
	kvsmock "github.com/ethersphere/bee/v2/pkg/kvs/mock"
	"github.com/ethersphere/bee/v2/pkg/storage"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type GranteeManager interface {
	//PUT /grantees/{grantee}
	//body: {publisher?, grantee root hash ,grantee}
	Grant(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error
	//DELETE /grantees/{grantee}
	//body: {publisher?, grantee root hash , grantee}
	Revoke(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error
	//[ ]
	//POST /grantees
	//body: {publisher, historyRootHash}
	Commit(ctx context.Context, granteesAddress swarm.Address, actRootHash swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, swarm.Address, error)

	//Post /grantees
	//{publisher, addList, removeList}
	HandleGrantees(ctx context.Context, rootHash swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) error

	//GET /grantees/{history root hash}
	GetGrantees(ctx context.Context, rootHash swarm.Address) ([]*ecdsa.PublicKey, error)
}

// TODO: ądd granteeList ref to history metadata to solve inconsistency
type Controller interface {
	GranteeManager
	DownloadHandler(ctx context.Context, timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error)
	UploadHandler(ctx context.Context, ref swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, swarm.Address, error)
}

type controller struct {
	accessLogic ActLogic
	granteeList GranteeList
	//[ ]: do we need to protect this with a mutex?
	revokeFlag []swarm.Address
	loadsaver  file.LoadSaver
}

var _ Controller = (*controller)(nil)

func (c *controller) DownloadHandler(ctx context.Context, timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error) {
	history, err := NewHistory(c.loadsaver, &historyRootHash)
	if err != nil {
		return swarm.ZeroAddress, err
	}

	kvsRef, err := history.Lookup(ctx, timestamp)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	kvs := kvs.New(c.loadsaver, kvsRef)
	return c.accessLogic.DecryptRef(ctx, kvs, enryptedRef, publisher)
}

// TODO: review return params: how to get back history ref ?
func (c *controller) UploadHandler(ctx context.Context, ref swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, swarm.Address, error) {
	var hptr *swarm.Address
	if !historyRootHash.Equal(swarm.ZeroAddress) {
		hptr = &historyRootHash
	}
	history, err := NewHistory(c.loadsaver, hptr)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	now := time.Now().Unix()
	kvsRef, err := history.Lookup(ctx, now)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	kvs := kvs.New(c.loadsaver, kvsRef)
	if kvsRef.Equal(swarm.ZeroAddress) {
		err = c.accessLogic.AddPublisher(ctx, kvs, publisher)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
		kvsRef, err = kvs.Save(ctx)
		if err != nil {
			return swarm.ZeroAddress, swarm.ZeroAddress, err
		}
	}
	err = history.Add(ctx, kvsRef, &now)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	hRef, err := history.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, swarm.ZeroAddress, err
	}
	enryptedRef, err := c.accessLogic.EncryptRef(ctx, kvs, publisher, ref)
	return hRef, enryptedRef, err
}

func requestPipelineFactory(ctx context.Context, s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(ctx, s, encrypt, rLevel)
	}
}

func NewController(ctx context.Context, accessLogic ActLogic, getter storage.Getter, putter storage.Putter) Controller {
	return &controller{
		granteeList: nil, //NewGranteeList(ls, ps, swarm.EmptyAddress),
		accessLogic: accessLogic,
		// TODO: set redundancy level and encryption flag
		loadsaver: loadsave.New(getter, putter, requestPipelineFactory(ctx, putter, false, redundancy.NONE)),
	}
}

func (c *controller) Grant(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	return c.granteeList.Add([]*ecdsa.PublicKey{grantee})
}

func (c *controller) Revoke(ctx context.Context, granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	if !c.isRevokeFlagged(granteesAddress) {
		c.setRevokeFlag(granteesAddress, true)
	}
	return c.granteeList.Remove([]*ecdsa.PublicKey{grantee})
}

func (c *controller) Commit(ctx context.Context, granteesAddress swarm.Address, actRootHash swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, swarm.Address, error) {
	//HACKreplace mock with real kvs
	var act kvs.KeyValueStore
	if c.isRevokeFlagged(granteesAddress) {
		act = kvsmock.New()
		c.accessLogic.AddPublisher(ctx, act, publisher)
	} else {
		act = kvsmock.NewReference(actRootHash)
	}

	grantees := c.granteeList.Get()
	for _, grantee := range grantees {
		c.accessLogic.AddGrantee(ctx, act, publisher, grantee, nil)
	}

	//HACK: Store not implemented
	granteeref, err := c.granteeList.Save(ctx)
	if err != nil {
		return swarm.EmptyAddress, swarm.EmptyAddress, err
	}

	actref, err := act.Save(ctx)
	if err != nil {
		return swarm.EmptyAddress, swarm.EmptyAddress, err
	}

	c.setRevokeFlag(granteesAddress, false)
	return granteeref, actref, err
}

func (c *controller) HandleGrantees(ctx context.Context, granteesAddress swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) error {
	act := kvsmock.New()

	c.accessLogic.AddPublisher(ctx, act, publisher)
	for _, grantee := range addList {
		c.accessLogic.AddGrantee(ctx, act, publisher, grantee, nil)
	}
	// granteeList.Store()
	return nil
}

func (c *controller) GetGrantees(ctx context.Context, granteeRootHash swarm.Address) ([]*ecdsa.PublicKey, error) {
	//[ ]: grantee list address deterministic or stored?
	// grateeListAddress, _ := hash(append([]byte(topic), []byte("grantee")...))
	return c.granteeList.Get(), nil
}

func (c *controller) isRevokeFlagged(granteeRootHash swarm.Address) bool {
	for _, revoke := range c.revokeFlag {
		if revoke.Equal(granteeRootHash) {
			return true
		}
	}
	return false
}

func (c *controller) setRevokeFlag(granteeRootHash swarm.Address, set bool) {
	if set {
		c.revokeFlag = append(c.revokeFlag, granteeRootHash)
	} else {
		for i, revoke := range c.revokeFlag {
			if revoke.Equal(granteeRootHash) {
				c.revokeFlag = append(c.revokeFlag[:i], c.revokeFlag[i+1:]...)
			}
		}
	}
}
