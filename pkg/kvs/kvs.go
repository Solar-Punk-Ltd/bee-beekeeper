package kvs

import (
	"github.com/ethersphere/bee/pkg/api"
	"github.com/ethersphere/bee/pkg/kvs/manifest"
	"github.com/ethersphere/bee/pkg/kvs/memory"
	"github.com/ethersphere/bee/pkg/swarm"
)

type KeyValueStore interface {
	Get(rootHash swarm.Address, key []byte) ([]byte, error)
	Put(rootHash swarm.Address, key, value []byte) (swarm.Address, error)
}

func NewManifestKeyValueStore(storer api.Storer) KeyValueStore {
	return &manifest.ManifestKeyValueStore{
		Storer: storer,
	}
}

func NewMemoryKeyValueStore(rootHash swarm.Address) KeyValueStore {
	return &memory.MemoryKeyValueStore{
		RootHash: rootHash,
	}
}
