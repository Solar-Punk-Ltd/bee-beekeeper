package dynamicaccess

import (
	"crypto/ecdsa"

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
	_, err := c.history.Lookup(timestamp)
	if err != nil {
		return swarm.EmptyAddress, err
	}
<<<<<<< HEAD
	addr, err := c.accessLogic.Get(swarm.EmptyAddress, enryptedRef, publisher)
=======
	addr, err := c.accessLogic.Get(act, enryptedRef, publisher)
>>>>>>> origin/act
	return addr, err
}

func (c *defaultController) UploadHandler(ref swarm.Address, publisher *ecdsa.PublicKey, topic string) (swarm.Address, error) {
	act, err := c.history.Lookup(0)
	if err != nil {
		return swarm.EmptyAddress, err
	}
	var actRef swarm.Address
	if act == nil {
		// new feed
<<<<<<< HEAD
		// act = NewInMemoryAct()
		actRef, err = c.granteeManager.Publish(swarm.EmptyAddress, publisher, topic)
		if err != nil {
			return swarm.EmptyAddress, err
		}
	}
	//FIXME: check if ACT is consistent with the grantee list
	return c.accessLogic.EncryptRef(actRef, publisher, ref)
=======
		act = NewInMemoryAct()
		act = c.granteeManager.Publish(act, publisher, topic)
	}
	//FIXME: check if ACT is consistent with the grantee list
	return c.accessLogic.EncryptRef(act, publisher, ref)
>>>>>>> origin/act
}

func NewController(history History, granteeManager GranteeManager, accessLogic ActLogic) Controller {
	return &defaultController{
		history:        history,
		granteeManager: granteeManager,
		accessLogic:    accessLogic,
	}
}
