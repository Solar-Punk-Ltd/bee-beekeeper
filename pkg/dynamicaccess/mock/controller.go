package mock

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/ethersphere/bee/pkg/dynamicaccess"
	"github.com/ethersphere/bee/pkg/file/loadsave"
	"github.com/ethersphere/bee/pkg/file/pipeline"
	"github.com/ethersphere/bee/pkg/file/pipeline/builder"
	"github.com/ethersphere/bee/pkg/file/redundancy"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/storage"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/ethersphere/bee/pkg/swarm"
)

type controllerMock struct {
	history        dynamicaccess.History
	granteeManager dynamicaccess.GranteeManager
	accessLogic    dynamicaccess.AccessLogic
}

func NewControllerMock(
	h dynamicaccess.History,
	gm dynamicaccess.GranteeManager,
	al dynamicaccess.AccessLogic,
) *controllerMock {
	return &controllerMock{
		history:        h,
		granteeManager: gm,
		accessLogic:    al,
	}
}

func (c *controllerMock) DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, tag string) (swarm.Address, error) {
	act, err := c.history.Lookup(timestamp)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	addr, err := c.accessLogic.Get(act, enryptedRef, *publisher, tag)
	return addr, err
}

func (c *controllerMock) UploadHandler(ref swarm.Address, publisher *ecdsa.PublicKey, topic string) (swarm.Address, error) {
	act, _ := c.history.Lookup(0)
	if act == nil {
		// new feed
		act = dynamicaccess.NewInMemoryAct()
		act = c.granteeManager.Publish(act, *publisher, topic)
		err := c.history.Add(time.Now().Unix(), act)
		if err != nil {
			return swarm.ZeroAddress, err
		}
	}
	//FIXME: check if ACT is consistent with the grantee list
	return c.accessLogic.EncryptRef(act, *publisher, ref)
}

// TODO: storer should be member and passed as a dep., e.g. /bzz
func (c *controllerMock) Store(act dynamicaccess.Act) (swarm.Address, error) {
	ctx := context.Background()
	storer := mockstorer.New()
	ls := loadsave.New(storer.ChunkStore(), storer.Cache(), pipelineFactory(storer.Cache(), false, 0))
	rootManifest, err := manifest.NewDefaultManifest(ls, false)
	if err != nil {
		return swarm.ZeroAddress, err
	}

	//FIXME
	//actManifEntry := act.Load(swarm.EmptyAddress)
	err = rootManifest.Add(ctx, manifest.RootPath, manifest.NewEntry(swarm.EmptyAddress, map[string]string{}))
	if err != nil {
		return swarm.ZeroAddress, err
	}
	manifRef, err := rootManifest.Store(ctx)
	if err != nil {
		return swarm.ZeroAddress, err
	}

	return manifRef, nil
}

func pipelineFactory(s storage.Putter, encrypt bool, rLevel redundancy.Level) func() pipeline.Interface {
	return func() pipeline.Interface {
		return builder.NewPipelineBuilder(context.Background(), s, encrypt, rLevel)
	}
}
