package dynamicaccess

import (
	"context"
	"crypto/ecdsa"
	"io"

	"github.com/ethersphere/bee/v2/pkg/file/redundancy"
	"github.com/ethersphere/bee/v2/pkg/swarm"
)

type Service interface {
	DownloadHandler(ctx context.Context, timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address, encrypt bool, rLevel redundancy.Level) (swarm.Address, error)
	UploadHandler(ctx context.Context, reference swarm.Address, publisher *ecdsa.PublicKey, historyRootHash *swarm.Address, encrypt bool, rLevel redundancy.Level) (swarm.Address, swarm.Address, error)
	io.Closer
}

type service struct {
	controller Controller
}

func (s *service) DownloadHandler(ctx context.Context, timestamp int64, enryptedRef swarm.Address, publisher *ecdsa.PublicKey, historyRootHash swarm.Address, encrypt bool, rLevel redundancy.Level) (swarm.Address, error) {
	return s.controller.DownloadHandler(ctx, timestamp, enryptedRef, publisher, historyRootHash, encrypt, rLevel)
}

func (s *service) UploadHandler(ctx context.Context, reference swarm.Address, publisher *ecdsa.PublicKey, historyRootHash *swarm.Address, encrypt bool, rLevel redundancy.Level) (swarm.Address, swarm.Address, error) {
	return s.controller.UploadHandler(ctx, reference, publisher, historyRootHash, encrypt, rLevel)
}

func (s *service) Close() error {
	return nil
}

func NewService(controller Controller) (Service, error) {
	return &service{
		controller: controller,
	}, nil
}