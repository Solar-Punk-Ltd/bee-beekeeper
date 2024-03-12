package mock

import (
	"encoding/hex"

	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
)

const (
	ContentTypeHeader = "Content-Type"
)

type ActMock struct {
	AddFunc      func(rootHash string, lookupKey []byte, encryptedAccessKey []byte) (swarm.Address, error)
	GetFunc      func(rootHash string, lookpupkey []byte) (string, error)
	manifestMock map[string]map[string]string
}

// TODO: check length of keys, publisher etc.
func (act *ActMock) Add(rootHash string, lookupKey []byte, encryptedAccessKey []byte) (swarm.Address, error) {
	if act.AddFunc == nil {
		metadata := make(map[string]string)
		metadata[ContentTypeHeader] = "text/plain"
		metadata[hex.EncodeToString(lookupKey)] = hex.EncodeToString(encryptedAccessKey)
		act.manifestMock[rootHash] = metadata
		return swarm.ZeroAddress, nil
	}
	return act.AddFunc(rootHash, lookupKey, encryptedAccessKey)
}

func (act *ActMock) Get(rootHash string, lookupKey []byte) (string, error) {
	if act.GetFunc == nil {
		metadata := act.manifestMock[rootHash]
		if metadata == nil {
			return swarm.ZeroAddress.String(), manifest.ErrNotFound
		}
		encryptedAccessKey := metadata[hex.EncodeToString(lookupKey)]
		return encryptedAccessKey, nil
	}
	return act.GetFunc(rootHash, lookupKey)
}

func NewActMock() *ActMock {
	return &ActMock{
		manifestMock: make(map[string]map[string]string),
	}
}
