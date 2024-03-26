package kvs

import (
	"context"
	"encoding/hex"

	"github.com/ethersphere/bee/pkg/file"
	"github.com/ethersphere/bee/pkg/manifest"
	"github.com/ethersphere/bee/pkg/storer"
	"github.com/ethersphere/bee/pkg/swarm"
)

type KeyValueStore interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Load() manifest.Interface
	Save(putter storer.PutterSession) (swarm.Address, error)
}

type keyValueStore struct {
	m manifest.Interface
}

var _ KeyValueStore = (*keyValueStore)(nil)

// TODO: pass context as dep.
func (s *keyValueStore) Get(key []byte) ([]byte, error) {
	entry, err := s.m.Lookup(context.Background(), hex.EncodeToString(key))
	if err != nil {
		return nil, err
	}
	ref := entry.Reference()
	return ref.Bytes(), nil
}

func (s *keyValueStore) Put(key []byte, value []byte) error {
	return s.m.Add(context.Background(), hex.EncodeToString(key), manifest.NewEntry(swarm.NewAddress(value), map[string]string{}))
}

func (s *keyValueStore) Load() manifest.Interface {
	return s.m
}

func (s *keyValueStore) Save(putter storer.PutterSession) (swarm.Address, error) {
	ref, err := s.m.Store(context.Background())
	if err != nil {
		return swarm.ZeroAddress, err
	}
	err = putter.Done(ref)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	return ref, nil
}

func New(ls file.LoadSaver, rootHash swarm.Address) KeyValueStore {
	var (
		manif manifest.Interface
		err   error
	)
	if swarm.ZeroAddress.Equal(rootHash) || swarm.EmptyAddress.Equal(rootHash) {
		manif, err = manifest.NewSimpleManifest(ls)
	} else {
		manif, err = manifest.NewSimpleManifestReference(rootHash, ls)
	}
	if err != nil {
		return nil
	}

	return &keyValueStore{
		m: manif,
	}
}
