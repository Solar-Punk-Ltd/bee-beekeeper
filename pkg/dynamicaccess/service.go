package dynamicaccess

import (
	"crypto/ecdsa"
	"io"

	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type Service interface {
	DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error)
	io.Closer
}

type service struct {
	controller Controller
}

func (s *service) DownloadHandler(timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address) (swarm.Address, error) {
	return s.controller.DownloadHandler(timestamp, enryptedRef, publisher, historyRootHash)
}

func (s *service) Close() error {
	return nil
}

func NewService(controller Controller) (Service, error) {
	return &service{
		controller: controller,
	}, nil
}
