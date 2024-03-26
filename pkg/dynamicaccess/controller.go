package dynamicaccess

import (
	"crypto/ecdsa"

	"github.com/ethersphere/bee/pkg/kvs"
	"github.com/ethersphere/bee/pkg/swarm"
)

type Controller interface {
	DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, tag string) (swarm.Address, error)
	UploadHandler(ref swarm.Address, publisher *ecdsa.PublicKey, topic string) (swarm.Address, error)
}

type defaultController struct {
	history        History
	granteeManager GranteeManager
	accessLogic    ActLogic
}

func (c *defaultController) DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, tag string) (swarm.Address, error) {
	kvs, err := c.history.Lookup(timestamp)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	addr, err := c.accessLogic.Get(kvs, enryptedRef, publisher)
	return addr, err
}

func (c *defaultController) UploadHandler(ref swarm.Address, publisher *ecdsa.PublicKey, topic string) (swarm.Address, error) {
	act, err := c.history.Lookup(0)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	var s kvs.KeyValueStore
	if act == nil {
		// new feed
		s = kvs.New(nil, swarm.ZeroAddress)
		_, err = c.granteeManager.Publish(s, publisher, topic)
		if err != nil {
			return swarm.EmptyAddress, err
		}
	}
	//FIXME: check if ACT is consistent with the grantee list
	return c.accessLogic.EncryptRef(s, publisher, ref)
}

func NewController(history History, granteeManager GranteeManager, accessLogic ActLogic) Controller {
	return &defaultController{
		history:        history,
		granteeManager: granteeManager,
		accessLogic:    accessLogic,
	}
}
