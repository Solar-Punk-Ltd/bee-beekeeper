package kvs

import (
	"context"
	"encoding/hex"

	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/swarm"
)

type KeyValueStore interface {
	Get(rootHash swarm.Address, key []byte) ([]byte, error)
	Put(rootHash swarm.Address, key, value []byte) (swarm.Address, error)
}

type keyValueStore struct {
	ls file.LoadSaver
}

// TODO: pass context as dep.
func (s *keyValueStore) Get(rootHash swarm.Address, key []byte) ([]byte, error) {
	// existing manif
	manif, err := manifest.NewSimpleManifestReference(rootHash, s.ls)
	if err != nil {
		return nil, err
	}
	entry, err := manif.Lookup(context.Background(), hex.EncodeToString(key))
	if err != nil {
		return nil, err
	}
	ref := entry.Reference()
	return ref.Bytes(), nil
}

func (s *keyValueStore) Put(rootHash swarm.Address, key []byte, value []byte) (swarm.Address, error) {
	// existing manif
	manif, err := manifest.NewSimpleManifestReference(rootHash, s.ls)
	if err != nil {
		// new manif
		manif, err = manifest.NewSimpleManifest(s.ls)
		if err != nil {
			return swarm.EmptyAddress, err
		}
	}
	err = manif.Add(context.Background(), hex.EncodeToString(key), manifest.NewEntry(swarm.NewAddress(value), map[string]string{}))
	if err != nil {
		return swarm.EmptyAddress, err
	}
	manifRef, err := manif.Store(context.Background())
	if err != nil {
		return swarm.EmptyAddress, err
	}
	return manifRef, nil
}

func New(ls file.LoadSaver) KeyValueStore {
	return &keyValueStore{
		ls: ls,
	}
}
