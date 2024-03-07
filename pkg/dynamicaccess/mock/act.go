package mock

import (
	"context"
	"encoding/hex"

	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
)

const (
	ContentTypeHeader = "Content-Type"
)

type ActMock struct {
	AddFunc  func(ctx context.Context, rootHash string, lookupKey0 []byte, encryptedAccessKey string) (swarm.Address, error)
	GetFunc  func(ctx context.Context, rootHash []byte, key []byte) (string, error)
	manifest manifest.Interface
	// TODO putter
}

// TODO: check length of keys, publisher etc.
func (act *ActMock) Add(ctx context.Context, rootHash string, lookupKey0 []byte, encryptedAccessKey string) (swarm.Address, error) {
	if act.AddFunc == nil {
		metadata := make(map[string]string)
		metadata[ContentTypeHeader] = "text/plain"
		metadata[hex.EncodeToString(lookupKey0)] = encryptedAccessKey
		err := act.manifest.Add(ctx, manifest.RootPath, manifest.NewEntry(swarm.ZeroAddress, metadata))
		if err != nil {
			return swarm.ZeroAddress, err
		}
		manifestReference, err := act.manifest.Store(ctx)
		return manifestReference, err
	}
	return act.AddFunc(ctx, rootHash, lookupKey0, encryptedAccessKey)
}

func (act *ActMock) Get(ctx context.Context, rootHash []byte, lookupKey0 []byte) (string, error) {
	if act.GetFunc == nil {
		me, err := act.manifest.Lookup(ctx, manifest.RootPath)
		if err != nil {
			// "not enough bytes for node fork: 64 (96) on byte '61'"
			return swarm.ZeroAddress.String(), err
		}
		encryptedAccessKey := me.Metadata()[hex.EncodeToString(lookupKey0)]
		return encryptedAccessKey, err
	}
	return act.GetFunc(ctx, rootHash, lookupKey0)
}

func NewActMock(m manifest.Interface) *ActMock {
	return &ActMock{
		manifest: m,
	}
}
