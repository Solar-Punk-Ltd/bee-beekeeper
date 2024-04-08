package dynamicaccess

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethersphere/bee/v2/pkg/file"
	"github.com/ethersphere/bee/v2/pkg/file/loadsave"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline"
	"github.com/ethersphere/bee/v2/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/kvs"
	kvsmock "github.com/ethersphere/bee/v2/pkg/kvs/mock"
	"github.com/ethersphere/bee/v2/pkg/storage"
	mockstorer "github.com/ethersphere/bee/v2/pkg/storer/mock"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

var mockStorer = mockstorer.New()

// TODO: refactor grantees to not use topic
// const topic = "grantees"

type GranteeManager interface {
	//PUT /grantees/{grantee}
	//body: {publisher?, grantee root hash ,grantee}
	Grant(granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error
	//DELETE /grantees/{grantee}
	//body: {publisher?, grantee root hash , grantee}
	Revoke(granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error
	//[ ]
	//POST /grantees
	//body: {publisher, historyRootHash}
	Commit(granteesAddress swarm.Address, actRootHash swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, swarm.Address, error)

	//Post /grantees
	//{publisher, addList, removeList}
	HandleGrantees(rootHash swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) error

	//GET /grantees/{history root hash}
	GetGrantees(rootHash swarm.Address) ([]*ecdsa.PublicKey, error)
}

type Controller interface {
	GranteeManager
	DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error)
	UploadHandler(ref swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error)
}

type controller struct {
	//history     History
	accessLogic ActLogic
	granteeList GranteeList
	//[ ]: do we need to protect this with a mutex?
	revokeFlag []swarm.Address
}

var _ Controller = (*controller)(nil)

func (c *controller) DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error) {
	//FIXME: newHistoryReference
	var mockStorer = mockstorer.New()
	ls := loadsave.New(mockStorer.ChunkStore(), mockStorer.Cache(), requestPipelineFactory(context.Background(), mockStorer.Cache(), false, redundancy.NONE))
	history, err := NewHistory(ls, &enryptedRef)
	if err != nil {
		return swarm.EmptyAddress, err
	}

	kvsRef, err := history.Lookup(context.Background(), timestamp)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	kvs := kvs.New(ls, mockStorer.DirectUpload(), kvsRef)
	addr, err := c.accessLogic.DecryptRef(kvs, enryptedRef, publisher)
	return addr, err
}

func (c *controller) UploadHandler(ref swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error) {
	var mockStorer = mockstorer.New()
	ls := loadsave.New(mockStorer.ChunkStore(), mockStorer.Cache(), requestPipelineFactory(context.Background(), mockStorer.Cache(), false, redundancy.NONE))
	history, err := NewHistory(ls, &ref)
	kvsRef, _ := history.Lookup(context.Background(), 0)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	// if actRootHash.Equal(swarm.EmptyAddress) {
	// 	actRootHash = c.granteeManager.Publish(actRootHash, publisher)
	// }
	// TODO: add to history
	kvs := kvs.New(ls, mockStorer.DirectUpload(), kvsRef)
	return c.accessLogic.EncryptRef(kvs, publisher, ref)
}

func requestPipelineFactory(ctx context.Context, s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(ctx, s, encrypt, rLevel)
	}
}

func createLs() file.LoadSaver {
	return loadsave.New(mockStorer.ChunkStore(), mockStorer.Cache(), requestPipelineFactory(context.Background(), mockStorer.Cache(), false, redundancy.NONE))
}

func NewController(accessLogic ActLogic) Controller {
	return &controller{
		granteeList: NewGranteeList(createLs(), mockStorer.DirectUpload(), swarm.EmptyAddress),
		//history:     NewHistory([]byte(""), common.HexToAddress("")),
		accessLogic: accessLogic,
	}
}

func (c *controller) Grant(granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	return c.granteeList.Add([]*ecdsa.PublicKey{grantee})
}

func (c *controller) Revoke(granteesAddress swarm.Address, grantee *ecdsa.PublicKey) error {
	if !c.isRevokeFlagged(granteesAddress) {
		c.setRevokeFlag(granteesAddress, true)
	}
	return c.granteeList.Remove([]*ecdsa.PublicKey{grantee})
}

func (c *controller) Commit(granteesAddress swarm.Address, actRootHash swarm.Address, publisher *ecdsa.PublicKey) (swarm.Address, swarm.Address, error) {
	//HACKreplace mock with real kvs
	var act kvs.KeyValueStore
	if c.isRevokeFlagged(granteesAddress) {
		act = kvsmock.New()
		c.accessLogic.AddPublisher(act, publisher)
	} else {
		act = kvsmock.NewReference(actRootHash)
	}

	grantees := c.granteeList.Get()
	for _, grantee := range grantees {
		c.accessLogic.AddGrantee(act, publisher, grantee, nil)
	}

	//HACK: Store not implemented
	granteeref, err := c.granteeList.Save()
	if err != nil {
		return swarm.EmptyAddress, swarm.EmptyAddress, err
	}

	actref, err := act.Save()
	if err != nil {
		return swarm.EmptyAddress, swarm.EmptyAddress, err
	}

	c.setRevokeFlag(granteesAddress, false)
	return granteeref, actref, err
}

func (c *controller) HandleGrantees(granteesAddress swarm.Address, publisher *ecdsa.PublicKey, addList, removeList []*ecdsa.PublicKey) error {
	act := kvsmock.New()

	c.accessLogic.AddPublisher(act, publisher)
	for _, grantee := range addList {
		c.accessLogic.AddGrantee(act, publisher, grantee, nil)
	}
	// granteeList.Store()
	return nil
}

func (c *controller) GetGrantees(granteeRootHash swarm.Address) ([]*ecdsa.PublicKey, error) {
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
